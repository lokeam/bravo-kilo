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
		Id       string `json:"od"`
		Email    string `json:"email"`
		Name     string `json:"name"`
		Picture  string `json:"picture"`
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
		Picture:    userInfo.Picture,
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
	expirationTime := time.Now().Add(60 * time.Minute)
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
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	// Redirect to FE dashboard
	dashboardURL := "http://localhost:5173/library"
	http.Redirect(response, request, dashboardURL, http.StatusSeeOther)

	h.logger.Info("JWT successfully sent to FE with status code: ", "info", http.StatusSeeOther)
}

// Verify JWT Token
func (h *Handlers) VerifyToken(response http.ResponseWriter, request *http.Request) {
	cookie, err := request.Cookie("token")
	if err != nil {
			h.logger.Error("Error: No token cookie", "error", err)
			http.Error(response, "Error: No token cookie", http.StatusUnauthorized)
			return
	}

	tokenStr := cookie.Value
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
	})
	if err != nil || !token.Valid {
			h.logger.Error("Error: Invalid token", "error", err)
			http.Error(response, "Invalid token", http.StatusUnauthorized)
			return
	}

	user, err := h.models.User.GetByID(claims.UserID)
	if err != nil {
			h.logger.Error("Error: User not found", "error", err)
			http.Error(response, "User not found", http.StatusInternalServerError)
			return
	}

	h.logger.Info("User info retrieved!", "user", user)

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(map[string]interface{}{
			"user": user,
	})
}

// Process user sign out
func (h *Handlers) SignOut(response http.ResponseWriter, request *http.Request) {
	// Clear token cookie
	http.SetCookie(response, &http.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	h.logger.Info("User logged out, token cookie cleared")
	response.WriteHeader(http.StatusOK)
	response.Write([]byte("Logged out successfully"))
}

// Process Google Books API Search
func (h *Handlers) SearchBooks(response http.ResponseWriter, request *http.Request) {
	query := request.URL.Query().Get("query")
	if query == "" {
		h.logger.Error("Search query is missing", "error", "missing query")
		http.Error(response, "Query parameter is required", http.StatusBadRequest)
		return
	}

	googleBooksAPI := "https://www.googleapis.com/books/v1/volumes"
	req, err := http.NewRequest("GET", googleBooksAPI, nil)
	if err != nil {
		h.logger.Error("Error creating request", "error", err)
		http.Error(response, "Error creating request", http.StatusInternalServerError)
		return
	}

	q := req.URL.Query()
	q.Add("q", query)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error("Error making request to Google Books API", "error", err)
		http.Error(response, "Error making request to Google Books API", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.logger.Error("Error response from Google Books API", "status", resp.Status)
		http.Error(response, "Error response from Google Books API", resp.StatusCode)
		return
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		h.logger.Error("Error decoding Google Books API response", "error", err)
		http.Error(response, "Error decoding Google Books API response", http.StatusInternalServerError)
		return
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(result)
}

