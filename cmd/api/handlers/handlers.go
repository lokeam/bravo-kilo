package handlers

import (
	"bravo-kilo/config"
	"context"
	"crypto/rand"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Handlers struct to hold the logger
type Handlers struct {
  infoLog *log.Logger
}

// NewHandlers creates a new Handlers instance
func NewHandlers(infoLog *log.Logger) *Handlers {
	return &Handlers{
		infoLog: infoLog,
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

func (h *Handlers) GoogleSignIn(response http.ResponseWriter, request *http.Request) {
	h.infoLog.Println("Handling Google OAuth callback")
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
	return
}

func (h *Handlers) GoogleCallback(response http.ResponseWriter, request *http.Request) {
	h.infoLog.Println("Handling Google OAuth callback")

	state := request.URL.Query().Get("state")

	// Retrieve the state from the cookie
	cookie, err := request.Cookie("oauthstate")
	if err != nil {
		h.infoLog.Println("Error: State cookie not found")
		http.Error(response, "Error: State cookie not found", http.StatusBadRequest)
		return
	}

	if state != cookie.Value {
		h.infoLog.Println("Error: URL State and Cookie state don't match")
		http.Error(response, "Error: States don't match", http.StatusBadRequest)
		return
	}

	code := request.URL.Query().Get("code")

	googleCfg := config.AppConfig.GoogleLoginConfig

	token, err := googleCfg.Exchange(context.Background(), code)
	if err != nil {
		h.infoLog.Println("Code/Token exchange failed", err)
		http.Error(response, "Code/Token exchange failed", http.StatusInternalServerError)
		return
	}

	oauthResponse, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		h.infoLog.Println("User data fetch failed", err)
		http.Error(response, "User data fetch failed", http.StatusInternalServerError)
		return
	}
	defer oauthResponse.Body.Close()

	userData, err := io.ReadAll(oauthResponse.Body)
	if err != nil {
		h.infoLog.Println("JSON parsing failed: ", err)
		http.Error(response, "JSON parsing failed", http.StatusInternalServerError)
		return
	}

	response.Header().Set("Content-Type", "application/json")
	response.Write(userData)
}
