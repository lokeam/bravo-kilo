package handlers

import (
	"bravo-kilo/config"
	"bravo-kilo/internal/data"
	"bravo-kilo/internal/utils"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var jwtKey = []byte("extra-super-secret-256-bit-key")

type Claims struct {
	UserID int `json:"userId"`
	jwt.RegisteredClaims
}

// Handlers struct to hold the logger
type Handlers struct {
	logger *slog.Logger
	models data.Models
}

type jsonResponse struct {
	Error    bool        `json:"error"`
	Message  string      `json:"message"`
	Data     interface{} `json:"data,omitempty"`
}


// NewHandlers creates a new Handlers instance
func NewHandlers(logger *slog.Logger, models data.Models) *Handlers {
	return &Handlers{
		logger: logger,
		models: models,
	}
}

// Generate random state for CSRF protection
func generateState() string {
	byteSlice := make([]byte, 16)
	_, err := rand.Read(byteSlice)
	if err != nil {
		log.Fatalf("Error generating random state: %s", err)
	}

	return base64.URLEncoding.EncodeToString(byteSlice)
}

// Init OAuth with Google
func (h *Handlers) GoogleSignIn(response http.ResponseWriter, request *http.Request) {
	h.logger.Info("Handling Google OAuth callback")
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	var GoogleLoginConfig = oauth2.Config{
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URI"),
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	randomState := generateState()

	// Set the state as a cookie
	http.SetCookie(response, &http.Cookie{
		Name:     "oauthstate",
		Value:    randomState,
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
		Secure:   true,
	})

	url := GoogleLoginConfig.AuthCodeURL(randomState)

	http.Redirect(response, request, url, http.StatusSeeOther)
}

// Process Google OAuth callback
func (h *Handlers) GoogleCallback(response http.ResponseWriter, request *http.Request) {
	h.logger.Info("Handling Google OAuth callback")

	state := request.URL.Query().Get("state")

	// Retrieve the state from the cookie
	cookie, err := request.Cookie("oauthstate")
	if err != nil {
		h.logger.Error("Error: State cookie not found", "error", err)
		http.Error(response, "Error: State cookie not found", http.StatusBadRequest)
		return
	}

	if state != cookie.Value {
		h.logger.Error("Error: URL State and Cookie state don't match", "error", err)
		http.Error(response, "Error: States don't match", http.StatusBadRequest)
		return
	}

	code := request.URL.Query().Get("code")

	googleCfg := config.AppConfig.GoogleLoginConfig

	token, err := googleCfg.Exchange(context.Background(), code)
	if err != nil {
		h.logger.Error("Error exchanging code for token", "error", err)
		http.Error(response, "Error exchaning code for token", http.StatusInternalServerError)
		return
	}

	oauthResponse, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		h.logger.Error("Error getting user info", "error", err)
		http.Error(response, "Error getting user info", http.StatusInternalServerError)
		return
	}

	defer oauthResponse.Body.Close()

	var userInfo struct {
		Id     string `json:"od"`
		Email  string `json:"email"`
		Name   string `json:"name"`
	}

	// Decode response
	if err := json.NewDecoder(oauthResponse.Body).Decode(&userInfo); err != nil {
		h.logger.Error("Error decoding user info response:", "error", err)
		http.Error(response, "Error decoding user info response", http.StatusInternalServerError)
		return
	}

	h.logger.Info("User info recieved!", "user", userInfo)

	firstName, lastName := utils.SplitFullName(userInfo.Name)

	// Save userInfo + access tokens
	user := data.User{
		Email:      userInfo.Email,
		FirstName:  firstName,
		LastName:   lastName,
	}

	userId, err := h.models.User.Insert(user)
	if err != nil {
		h.logger.Error("Error adding user to db", "error", err)
		http.Error(response, "Error adding user to db", http.StatusInternalServerError)
		return
	}

	tokenRecord := data.Token{
		UserID:        userId,
		RefreshToken:  token.RefreshToken,
		TokenExpiry:   token.Expiry,
	}

	if err := h.models.Token.Insert(tokenRecord); err != nil {
		h.logger.Error("Error adding token to db", "error", err)
		http.Error(response, "Error adding token to db", http.StatusInternalServerError)
		return
	}

	h.logger.Info("User and tokens stored successfully")

	// Create JWT
	expirationTime := time.Now().Add(15 * time.Minute)
	claims := &Claims{
		UserID: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtString, err := jwtToken.SignedString(jwtKey)
	if err != nil {
		h.logger.Error("Error generating JWT", "error", err)
		http.Error(response, "Error generating JWT", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Generated JWT")

	// JWT as a cookie
	http.SetCookie(response, &http.Cookie{
		Name:     "token",
		Value:    jwtString,
		Expires:  expirationTime,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})

	// Redirect to FE dashboard
	dashboardURL := "http://localhost:5173/library"
	http.Redirect(response, request, dashboardURL, http.StatusSeeOther)

	h.logger.Info("JWT successfully sent to FE with status code: ", "info", http.StatusSeeOther)
}
