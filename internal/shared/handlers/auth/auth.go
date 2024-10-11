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
	"strings"
	"time"

	"github.com/lokeam/bravo-kilo/config"
	"github.com/lokeam/bravo-kilo/internal/shared/crypto"
	"github.com/lokeam/bravo-kilo/internal/shared/models"
	"github.com/lokeam/bravo-kilo/internal/shared/transaction"
	"github.com/lokeam/bravo-kilo/internal/shared/types"
	"github.com/lokeam/bravo-kilo/internal/shared/utils"

	"errors"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

type AuthHandlers struct {
	logger     *slog.Logger
	models     models.Models
	dbManager  transaction.DBManager
}

var (
	oidcProvider *oidc.Provider
	oidcVerifier *oidc.IDTokenVerifier
	oauth2Config *oauth2.Config
	isProduction bool
)
var ErrNoRefreshToken = errors.New("no valid refresh token found")
func init() {
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}
	isProduction = env == "production"
}

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

	// Initialize OIDC Provider + Verifier
	ctx := context.Background()
	var err error
	oidcProvider, err = oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
			logger.Error("Failed to get OIDC provider", "error", err)
			return nil, fmt.Errorf("failed to get OIDC provider: %v", err)
	}

	oidcConfig := &oidc.Config{
			ClientID: os.Getenv("GOOGLE_CLIENT_ID"),
	}
	oidcVerifier = oidcProvider.Verifier(oidcConfig)

	// Initialize OAuth2 Config + OIDC scopes
	oauth2Config = &oauth2.Config{
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URI"),
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			Scopes: []string{
					oidc.ScopeOpenID,
					"profile",
					"email",
					"https://www.googleapis.com/auth/books",
			},
			Endpoint: oidcProvider.Endpoint(),
	}

	return &AuthHandlers{
			logger:    logger,
			models:    models,
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

	randomState := generateState()

	// Set the state as a cookie
	http.SetCookie(response, &http.Cookie{
		Name:     "oauthstate",
		Value:    randomState,
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   isProduction,
		Path:     "/",
	})

	url := oauth2Config.AuthCodeURL(
		randomState,
		oauth2.AccessTypeOffline,
	)

	http.Redirect(response, request, url, http.StatusSeeOther)
}

// Process Google OAuth callback
func (h *AuthHandlers) HandleGoogleCallback(response http.ResponseWriter, request *http.Request) {
	h.logger.Info("Handling Google OAuth callback")

	// Verify state parameter
	state := request.URL.Query().Get("state")

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

	// Exchange the authorization code for tokens
	code := request.URL.Query().Get("code")
	ctx := request.Context()
	token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
			h.logger.Error("Error exchanging code for token", "error", err)
			http.Error(response, "Error exchanging code for token", http.StatusInternalServerError)
			return
	}

	// Extract the ID Token from OAuth2 token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
			h.logger.Error("Error: No id_token field in oauth2 token")
			http.Error(response, "Error: No id_token field in oauth2 token", http.StatusInternalServerError)
			return
	}

	// Verify ID Token
	idToken, err := oidcVerifier.Verify(ctx, rawIDToken)
	if err != nil {
			h.logger.Error("Error verifying ID Token", "error", err)
			http.Error(response, "Error verifying ID Token", http.StatusInternalServerError)
			return
	}

	// Extract user info from ID Token
	var userInfo struct {
			Email         string `json:"email"`
			EmailVerified bool   `json:"email_verified"`
			Name          string `json:"name"`
			Picture       string `json:"picture"`
			Sub           string `json:"sub"`
	}

	if err := idToken.Claims(&userInfo); err != nil {
			h.logger.Error("Error decoding ID Token claims", "error", err)
			http.Error(response, "Error decoding ID Token claims", http.StatusInternalServerError)
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
	claims := &types.Claims{
			UserID: userId,
			RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(expirationTime),
			},
	}

	if config.AppConfig.JWTPrivateKey == nil {
		h.logger.Error("JWT private key not found")
		http.Error(response, "JWT private key not found", http.StatusInternalServerError)
		return
	}

	jwtString, err := crypto.SignToken(claims, config.AppConfig.JWTPrivateKey)
	if err != nil {
			h.logger.Error("Error generating JWT", "error", err)
			http.Error(response, "Error generating JWT", http.StatusInternalServerError)
			return
	}

	h.logger.Info("Generated JWT")

	// Send the JWT as a cookie
	isProduction := os.Getenv("ENV") == "production"
	http.SetCookie(response, &http.Cookie{
			Name:     "token",
			Value:    jwtString,
			Expires:  expirationTime,
			HttpOnly: true,
			Secure:   isProduction,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
			Domain: "",
	})
	h.logger.Info("JWT cookie set",
		"name", "token",
		"value", jwtString[:10]+"...", // Log only the first 10 characters for security
		"expires", expirationTime,
		"secure", isProduction,
		"sameSite", http.SameSiteLaxMode,
		"path", "/",
	)
	// Set a non-HttpOnly cookie for testing
	http.SetCookie(response, &http.Cookie{
		Name:     "test_cookie",
		Value:    "test_value",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: false,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
	h.logger.Info("Test cookie set",
		"name", "test_cookie",
		"value", "test_value",
	)

	// Redirect to the frontend dashboard
	dashboardURL := "http://localhost:5173/library"
	http.Redirect(response, request, dashboardURL, http.StatusSeeOther)
	h.logger.Info("Redirecting to frontend",
		"url", dashboardURL,
		"statusCode", http.StatusSeeOther,
	)
	h.logger.Info("JWT successfully sent to FE with status code: ", "info", http.StatusSeeOther)
}

// Refresh CSRF Token
func (h *AuthHandlers) HandleRefreshCSRFToken(response http.ResponseWriter, request *http.Request) {
	h.logger.Info("HandleRefreshCSRFToken called")

	// CSRF Gorilla automatically sets new token in response header
	response.WriteHeader(http.StatusOK)
	response.Write([]byte("CSRF token refreshed"))

	h.logger.Info("CSRF token refreshed successfully")
}

// Retrieve JWT Token
func (h *AuthHandlers) GetUserAccessToken(request *http.Request) (*oauth2.Token, error) {
	// Get userID from JWT
	cookie, err := request.Cookie("token")
	if err != nil {
			h.logger.Error("Error: No token cookie", "error", err)
			return nil, fmt.Errorf("no token cookie")
	}

	tokenStr := cookie.Value
	token, err := crypto.VerifyToken(tokenStr, config.AppConfig.JWTPublicKey)
	if err != nil || !token.Valid {
			h.logger.Error("Error: Invalid token", "error", err)
			return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*types.Claims)
	if !ok {
		h.logger.Error("Error: Claims are not of type *types.Claims", "error", err)
		return nil, fmt.Errorf("invalid token claims")
	}

	// Get the refresh token for the user from the database
	h.logger.Info("Attempting to retrieve refresh token", "userID", claims.UserID)

	// Get refresh token from DB
	refreshToken, err := h.models.Token.GetRefreshTokenByUserID(claims.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "No refresh token found") {
			return nil, ErrNoRefreshToken
		}

			h.logger.Error("Error retrieving refresh token from DB", "error", err)
			return nil, fmt.Errorf("could not retrieve refresh token from DB")
	}
	if refreshToken == "" {
			h.logger.Error("No refresh token found for user", "userID", claims.UserID)
			return nil, fmt.Errorf("no refresh token found for user")
	}

	// Use the refresh token to obtain a new access token
	tokenSource := oauth2Config.TokenSource(context.Background(), &oauth2.Token{
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
	h.logger.Info("********************************")
	h.logger.Info("HandleVerifyToken called")

  // Log all cookies
  for _, cookie := range request.Cookies() {
		h.logger.Info("Received cookie", "name", cookie.Name, "value", cookie.Value)
	}

	cookie, err := request.Cookie("token")
	if err != nil {
			h.logger.Error("Error: No token cookie", "error", err)
			http.Error(response, "Error: No token cookie", http.StatusUnauthorized)
			return
	}
	h.logger.Info("Token cookie found", "cookieValue", cookie.Value[:10]+"...")

	tokenStr := cookie.Value

	token, err := crypto.VerifyToken(tokenStr, config.AppConfig.JWTPublicKey)
	if err != nil {
			h.logger.Error("Error verifying token", "error", err)
			http.Error(response, "Invalid token", http.StatusUnauthorized)
			return
	}

	h.logger.Info("Token verified successfully")

	claims, ok := token.Claims.(*types.Claims)
	if !ok {
			h.logger.Error("Error: Claims are not of type *types.Claims")
			http.Error(response, "Invalid token claims", http.StatusUnauthorized)
			return
	}

	// Log individual claims
	h.logger.Info("Parsed claims", "userID", claims.UserID, "expiresAt", claims.ExpiresAt)

	user, err := h.models.User.GetByID(claims.UserID)
	if err != nil {
			h.logger.Error("Error: User not found", "error", err, "userID", claims.UserID)
			http.Error(response, "User not found", http.StatusInternalServerError)
			return
	}

	h.logger.Info("User info retrieved", "userID", user.ID, "email", user.Email)

	response.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(response).Encode(map[string]interface{}{
			"user": user,
	})
	if err != nil {
			h.logger.Error("Error encoding user data to JSON", "error", err)
			http.Error(response, "Error encoding response", http.StatusInternalServerError)
			return
	}

	h.logger.Info("HandleVerifyToken completed successfully - text updated")
}

// Refresh JWT Token
func (h *AuthHandlers) HandleRefreshToken(response http.ResponseWriter, request *http.Request) {
	h.logger.Info("HandleRefreshToken called")

	// Grab refresh token from request
	cookie, err := request.Cookie("token")
	if err != nil {
		h.logger.Error("Error: No refresh token cookie", "error", err)
		http.Error(response, "Error: No refresh token cookie", http.StatusUnauthorized)
		return
	}

	oldTokenStr := cookie.Value

	// Parse old token
	claims := &types.Claims{}
	token, err := crypto.VerifyToken(oldTokenStr, config.AppConfig.JWTPublicKey)
	if err != nil || !token.Valid {
		h.logger.Error("Error: Invalid refresh token", "error", err)
		http.Error(response, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	// Extract claims from the token
	claims, ok := token.Claims.(*types.Claims)
	if !ok {
		h.logger.Error("Error: Claims are not of type *types.Claims")
		http.Error(response, "Invalid token claims", http.StatusUnauthorized)
		return
	}

    // Check if the token is about to expire
	if time.Until(claims.ExpiresAt.Time) > 5*time.Minute {
		h.logger.Info("Token is not close to expiration, no need to refresh")
		http.Error(response, "Token is not close to expiration", http.StatusBadRequest)
		return
	}

	// Generate new token (1 week)
	expirationTime := time.Now().Add(7 * 24 * time.Hour)
	newClaims := &types.Claims{
		UserID: claims.UserID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	h.logger.Info("New token generated", "newClaims", newClaims)

	newTokenStr, err := crypto.SignToken(newClaims, config.AppConfig.JWTPrivateKey)
	if err != nil {
		h.logger.Error("Error generating new JWT", "error", err)
		http.Error(response, "Error generating new JWT", http.StatusInternalServerError)
		return
	}

    // Log the values before calling Rotate
    h.logger.Info("Rotating token",
			"userID", claims.UserID,
			"newToken", newTokenStr,
			"oldToken", oldTokenStr,
			"expiry", expirationTime)

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
		Secure:   isProduction,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
	h.logger.Info("New token set as cookie", "newToken", newTokenStr[:10]+"...")
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
		Secure:   isProduction,
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
	claims := &types.Claims{}

	token, err := crypto.VerifyToken(tokenStr, config.AppConfig.JWTPublicKey)
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
	claims := &types.Claims{}

	token, err := crypto.VerifyToken(tokenStr, config.AppConfig.JWTPublicKey)
	if err != nil || !token.Valid {
		h.logger.Error("Error: Invalid token", "error", err)
		http.Error(response, "Error: Invalid token", http.StatusUnauthorized)
		return
	}

	userID := claims.UserID

	// Begin transaction
	ctx := request.Context()
	tx, err := h.dbManager.BeginTransaction(ctx)
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
		h.dbManager.RollbackTransaction(tx)
		http.Error(response, "Error marking user for deletion", http.StatusInternalServerError)
		return
	}

	// Delete refresh tokens
	err = h.models.Token.DeleteByUserID(userID)
	if err != nil {
		h.logger.Error("Error deleting refresh tokens", "error", err)
		h.dbManager.RollbackTransaction(tx)
		http.Error(response, "Error deleting refresh tokens", http.StatusInternalServerError)
		return
	}

	// Commit transaction
	err = h.dbManager.CommitTransaction(tx)
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
		Secure:   isProduction,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	h.logger.Info("User account marked for deletion and logged out", "userID", userID)
	response.WriteHeader(http.StatusOK)
}
