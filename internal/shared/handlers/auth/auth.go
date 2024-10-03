package handlers

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/lokeam/bravo-kilo/config"
	"github.com/lokeam/bravo-kilo/internal/shared/models"
	"github.com/lokeam/bravo-kilo/internal/shared/transaction"
	"github.com/lokeam/bravo-kilo/internal/shared/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type AuthHandlers struct {
	logger     *slog.Logger
	models     models.Models
	dbManager  transaction.DBManager
}

var claims utils.Claims
var jwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))

func NewAuthHandlers(logger *slog.Logger, models models.Models, dbManager transaction.DBManager) (*AuthHandlers, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	if models.User == nil || models.Token == nil {
		return nil, fmt.Errorf("invalid models passed, user or token model is missing ")
	}

	if dbManager == nil {
		return nil, fmt.Errorf("DB Manager cannot be nil")
	}

	return &AuthHandlers{
		logger: logger,
		models: models,
		dbManager: dbManager,
	}, nil
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
func (h *AuthHandlers) HandleGoogleSignIn(response http.ResponseWriter, request *http.Request) {
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
			"https://www.googleapis.com/auth/books",
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

	url := GoogleLoginConfig.AuthCodeURL(
		randomState,
		oauth2.AccessTypeOffline,
	)

	http.Redirect(response, request, url, http.StatusSeeOther)
}

// Process Google OAuth callback
func (h *AuthHandlers) HandleGoogleCallback(response http.ResponseWriter, request *http.Request) {
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
		h.logger.Error("Error: URL State and Cookie state don't match")
		http.Error(response, "Error: States don't match", http.StatusBadRequest)
		return
	}

	code := request.URL.Query().Get("code")

	googleCfg := config.AppConfig.GoogleLoginConfig

	// Exchange the authorization code for an access token
	token, err := googleCfg.Exchange(context.Background(), code, oauth2.AccessTypeOffline)
	if err != nil {
		h.logger.Error("Error exchanging code for token", "error", err)
		http.Error(response, "Error exchanging code for token", http.StatusInternalServerError)
		return
	}

	// Retrieve user info using the access token
	oauthResponse, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		h.logger.Error("Error getting user info", "error", err)
		http.Error(response, "Error getting user info", http.StatusInternalServerError)
		return
	}
	defer oauthResponse.Body.Close()

	// Decode the user info response
	var userInfo struct {
		Id      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(oauthResponse.Body).Decode(&userInfo); err != nil {
		h.logger.Error("Error decoding user info response:", "error", err)
		http.Error(response, "Error decoding user info response", http.StatusInternalServerError)
		return
	}

	h.logger.Info("User info received!", "user", userInfo)

	firstName, lastName := utils.SplitFullName(userInfo.Name)

	// Check if the user already exists in the database
	existingUser, err := h.models.User.GetByEmail(userInfo.Email)
	if err != nil && err != sql.ErrNoRows {
		h.logger.Error("Error checking for existing user", "error", err)
		http.Error(response, "Error checking for existing user", http.StatusInternalServerError)
		return
	}

	var userId int
	if existingUser != nil {
		userId = existingUser.ID
	} else {
		// Save user info in the database
		user := models.User{
			Email:     userInfo.Email,
			FirstName: firstName,
			LastName:  lastName,
			Picture:   userInfo.Picture,
		}

		userId, err = h.models.User.Insert(user)
		if err != nil {
			h.logger.Error("Error adding user to db", "error", err)
			http.Error(response, "Error adding user to db", http.StatusInternalServerError)
			return
		}
	}

	// Store refresh token if available
	if token.RefreshToken != "" {
		tokenRecord := models.Token{
			UserID:       userId,
			RefreshToken: token.RefreshToken,
			TokenExpiry:  token.Expiry,
		}

  if err := h.models.Token.Insert(tokenRecord); err != nil {
    h.logger.Error("Error adding token to db", "error", err)
    http.Error(response, "Error adding token to db", http.StatusInternalServerError)
		return
	}
  } else {
	  h.logger.Warn("No refresh token received, re-authentication needed for offline access")
  }

	// Create a JWT for the session
	expirationTime := time.Now().Add(60 * time.Minute)
	claims := &utils.Claims{
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

	// Send the JWT as a cookie
	http.SetCookie(response, &http.Cookie{
		Name:     "token",
		Value:    jwtString,
		Expires:  expirationTime,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	// Redirect to the frontend dashboard with userID as a query parameter
	dashboardURL := fmt.Sprintf("http://localhost:5173/library?userID=%d", userId)
	http.Redirect(response, request, dashboardURL, http.StatusSeeOther)

	h.logger.Info("JWT successfully sent to FE with status code: ", "info", http.StatusSeeOther)
}

// Retrieve Token
func (h *AuthHandlers) GetUserAccessToken(request *http.Request) (*oauth2.Token, error) {
	// Get userID from JWT
	cookie, err := request.Cookie("token")
	if err != nil {
		h.logger.Error("Error: No token cookie", "error", err)
		return nil, fmt.Errorf("no token cookie")
	}

	tokenStr := cookie.Value
	claims := &utils.Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		h.logger.Error("Error: Invalid token", "error", err)
		return nil, fmt.Errorf("invalid token")
	}

	// Get the refresh token for the user from the database
	refreshToken, err := h.models.Token.GetRefreshTokenByUserID(claims.UserID)
	if err != nil {
		h.logger.Error("Error retrieving refresh token from DB", "error", err)
		return nil, fmt.Errorf("could not retrieve refresh token from DB")
	}
	if refreshToken == "" {
		h.logger.Error("No refresh token found for user", "userID", claims.UserID)
		return nil, fmt.Errorf("no refresh token found for user")
	}

	// Use the refresh token to obtain a new access token
	tokenSource := config.AppConfig.GoogleLoginConfig.TokenSource(context.Background(), &oauth2.Token{
		RefreshToken: refreshToken,
	})

	// Get a new access token
	newToken, err := tokenSource.Token()
	if err != nil {
		h.logger.Error("Error refreshing access token", "error", err)
		return nil, fmt.Errorf("could not refresh access token")
	}

	return newToken, nil
}

// Verify JWT Token
func (h *AuthHandlers) HandleVerifyToken(response http.ResponseWriter, request *http.Request) {
	cookie, err := request.Cookie("token")
	if err != nil {
    h.logger.Error("Error: No token cookie", "error", err)
    http.Error(response, "Error: No token cookie", http.StatusUnauthorized)
    return
	}

	tokenStr := cookie.Value
	claims := &utils.Claims{}

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

// Refresh Token
func (h *AuthHandlers) HandleRefreshToken(response http.ResponseWriter, request *http.Request) {
	// Grab refresh token from request
	cookie, err := request.Cookie("token")
	if err != nil {
		h.logger.Error("Error: No refresh token cookie", "error", err)
		http.Error(response, "Error: No refresh token cookie", http.StatusUnauthorized)
		return
	}

	oldTokenStr := cookie.Value

	// Parse old token
	claims := &utils.Claims{}
	token, err := jwt.ParseWithClaims(oldTokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		h.logger.Error("Error: Invalid refresh token", "error", err)
		http.Error(response, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	// Generate new token (1 week)
	expirationTime := time.Now().Add(7 * 24 * time.Hour)
	newClaims := &utils.Claims{
		UserID: claims.UserID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	newTokenStr, err := newToken.SignedString(jwtKey)
	if err != nil {
		h.logger.Error("Error generating new JWT", "error", err)
		http.Error(response, "Error generating new JWT", http.StatusInternalServerError)
		return
	}

	// Rotate the refresh one
	err = h.models.Token.Rotate(claims.UserID, newTokenStr, oldTokenStr, expirationTime)
	if err != nil {
		h.logger.Error("Error rotating refresh token", "error", err)
		http.Error(response, "Error rotating refresh token", http.StatusInternalServerError)
		return
	}

	// Set the new refresh as a cookie
	http.SetCookie(response, &http.Cookie{
		Name:     "token",
		Value:    newTokenStr,
		Expires:  expirationTime,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	response.WriteHeader(http.StatusOK)
	response.Write([]byte("Token refreshed successfully"))
}

// Process user sign out
func (h *AuthHandlers) HandleSignOut(response http.ResponseWriter, request *http.Request) {
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

	// Get the user ID from JWT token
	cookie, err := request.Cookie("token")
	if err != nil {
		h.logger.Error("Error: No token cookie", "error", err)
		http.Error(response, "Error: No token cookie", http.StatusUnauthorized)
		return
	}

	tokenStr := cookie.Value
	claims := &utils.Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		h.logger.Error("Error: Invalid token", "error", err)
		http.Error(response, "Error: Invalid token", http.StatusUnauthorized)
		return
	}

	// Delete the refresh token from db
	if err := h.models.Token.DeleteByUserID(claims.UserID); err != nil {
		h.logger.Error("Error deleting refresh token by user ID", "error", err)
		http.Error(response, "Error deleting refresh token", http.StatusInternalServerError)
		return
	}

	h.logger.Info("User logged out, token cookie cleared, token deleted")
	response.WriteHeader(http.StatusOK)
	response.Write([]byte("Logged out successfully"))
}

func (h *AuthHandlers) HandleDeleteAccount(response http.ResponseWriter, request *http.Request) {
	cookie, err := request.Cookie("token")
	if err != nil {
		h.logger.Error("Error: No token cookie", "error", err)
		http.Error(response, "Error: No token cookie", http.StatusUnauthorized)
		return
	}

	tokenStr := cookie.Value
	claims := &utils.Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		h.logger.Error("Error: Invalid token", "error", err)
		http.Error(response, "Error: Invalid token", http.StatusUnauthorized)
		return
	}

	userID := claims.UserID

	// Begin transaction
	ctx := request.Context()
	tx, err := h.models.DBManager.BeginTransaction(ctx)
	if err != nil {
		h.logger.Error("Error beginning transaction", "error", err)
		http.Error(response, "Error processing request", http.StatusInternalServerError)
		return
	}

	// Soft delete (mark for deletion)
	deletionTime := time.Now()
	err = h.models.User.MarkForDeletion(userID, deletionTime)
	if err != nil {
		h.logger.Error("Error marking user for deletion", "error", err)
		h.models.DBManager.RollbackTransaction(tx)
		http.Error(response, "Error marking user for deletion", http.StatusInternalServerError)
		return
	}

	// Delete refresh tokens
	err = h.models.Token.DeleteByUserID(userID)
	if err != nil {
		h.logger.Error("Error deleting refresh tokens", "error", err)
		h.models.DBManager.RollbackTransaction(tx)
		http.Error(response, "Error deleting refresh tokens", http.StatusInternalServerError)
		return
	}

	// Commit transaction
	err = h.models.DBManager.CommitTransaction(tx)
	if err != nil {
		h.logger.Error("Error committing transaction", "error", err)
		http.Error(response, "Error processing request", http.StatusInternalServerError)
		return
	}

	// Invalidate JWT session cookie
	http.SetCookie(response, &http.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	h.logger.Info("User account marked for deletion and logged out", "userID", userID)
	response.WriteHeader(http.StatusOK)
}

