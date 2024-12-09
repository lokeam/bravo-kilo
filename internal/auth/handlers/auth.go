package authhandlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"errors"

	authservice "github.com/lokeam/bravo-kilo/internal/auth/services"

	"golang.org/x/oauth2"
)

type AuthHandlers struct {
	logger       *slog.Logger
	authService  authservice.AuthService
	oauthService authservice.OAuthService
	tokenService authservice.TokenService
	isProduction bool
	config       *oauth2.Config
}

type DeleteAccountResponse struct {
	Message       string `json:"message"`
	RedirectURL   string `json:"redirectURL"`
}

var ErrNoRefreshToken = errors.New("no valid refresh token found")

var (
	ErrNoToken = errors.New("no token cookie found")
	ErrInvalidToken = errors.New("invalid token")
	ErrUserNotFound = errors.New("user not found")
)

func NewAuthHandlers(
	logger *slog.Logger,
	authService authservice.AuthService,
	oauthService authservice.OAuthService,
	tokenService authservice.TokenService,
	config *oauth2.Config,
) *AuthHandlers {
	if logger == nil {
			panic("logger cannot be nil")
	}
	if authService == nil {
			panic("authService cannot be nil")
	}
	if oauthService == nil {
			panic("oauthService cannot be nil")
	}
	if tokenService == nil {
			panic("tokenService cannot be nil")
	}

    // Get environment
    env := os.Getenv("ENV")
    if env == "" {
        env = "development"
    }
    isProduction := env == "production"

	return &AuthHandlers{
			logger:       logger,
			authService:  authService,
			oauthService: oauthService,
			tokenService: tokenService,
			isProduction: isProduction,
			config:       config,
	}
}

// Init OAuth with Google - complete
func (h *AuthHandlers) HandleGoogleSignIn(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling Google OAuth callback")

	// Generate and set state cookie
	state := h.tokenService.GenerateState()
	h.tokenService.SetStateCookie(w, state)

	// Get authorization URL with state
	authURL := h.oauthService.GetAuthURL(state)

	// Redirect to Google for authorization
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
	h.logger.Info("Redirecting to Google OAuth",
			"url", authURL,
			"statusCode", http.StatusTemporaryRedirect,
	)
}

// Process Google OAuth callback -
func (h *AuthHandlers) HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling Google OAuth callback")

	// 1. Verify state
	state := r.URL.Query().Get("state")
	if err := h.tokenService.VerifyStateCookie(r, state); err != nil {
			h.handleOAuthError(w, r, "state_mismatch", err)
			return
	}
	h.logger.Info("Starting OAuth callback processing",
		"state", state,
		"handler", "auth",
	)


	// 2. Process OAuth callback
	code := r.URL.Query().Get("code")
	authResponse, err := h.authService.ProcessGoogleAuth(r.Context(), code)
	if err != nil {
		h.logger.Error("Failed to exchange code for token",
			"error", err,
			"handler", "auth")
		h.handleOAuthError(w, r, "auth_failed", err)
		return
	}

    // Add check for refresh token
		if authResponse.RefreshToken == "" {
			h.logger.Warn("No refresh token received, redirecting for reauthorization",
					"handler", "auth")
			h.handleMissingRefreshToken(w, r)
			return
		}

	// 3. Set session cookie
	h.tokenService.CreateSessionCookie(w, authResponse.Token, authResponse.ExpiresAt)

	// 4. Redirect to dashboard
	dashboardURL := h.getDashboardURL()
	http.Redirect(w, r, dashboardURL, http.StatusSeeOther)
	h.logger.Info("Auth successful, redirecting to dashboard",
			"url", dashboardURL,
			"userID", authResponse.User.ID,
	)
}

// Refresh CSRF Token
func (h *AuthHandlers) HandleRefreshCSRFToken(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("HandleRefreshCSRFToken called")

	// CSRF Gorilla automatically sets new token in response header
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("CSRF token refreshed"))

	h.logger.Info("CSRF token refreshed successfully")
}

// Retrieve valid JWT access token for user
func (h *AuthHandlers) GetUserAccessToken(r *http.Request) (*oauth2.Token, error) {

	// 1. Get userID from JWT
	userID, err := h.tokenService.GetUserIDFromToken(r)
	if err != nil {
			h.logger.Error("Failed to get user ID from token", "error", err)
			return nil, fmt.Errorf("invalid token: %w", err)
	}

	h.logger.Debug("GetUserAccessToken called",
		"userID", userID)

	// 2. Get access token using userID
	token, err := h.oauthService.GetAccessToken(r.Context(), userID)
	if err != nil {
			if errors.Is(err, ErrNoRefreshToken) {
					h.logger.Error("No refresh token found", "userID", userID)
					return nil, ErrNoRefreshToken
			}
			h.logger.Error("Failed to get access token",
					"error", err,
					"userID", userID,
			)
			return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	return token, nil
}

// Verify JWT Token
func (h *AuthHandlers) HandleVerifyToken(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("HandleVerifyToken called")

	// 1. Verify token and get user info
	userInfo, err := h.authService.VerifyAndGetUserInfo(r)
	if err != nil {
			switch {
			case errors.Is(err, ErrNoToken):
					h.handleTokenVerificationError(w, "no_token", err)
			case errors.Is(err, ErrInvalidToken):
					h.handleTokenVerificationError(w, "invalid_token", err)
			case errors.Is(err, ErrUserNotFound):
					h.handleTokenVerificationError(w, "user_not_found", err)
			default:
					h.handleTokenVerificationError(w, "internal_error", err)
			}
			return
	}

	// 2. Send successful response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"user": userInfo}); err != nil {
			h.logger.Error("Error encoding response", "error", err)
			h.handleTokenVerificationError(w, "encoding_error", err)
			return
	}

	h.logger.Info("Token verification successful",
			"userID", userInfo.ID,
			"email", userInfo.Email,
	)
}

// Refresh JWT Token
func (h *AuthHandlers) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("HandleRefreshToken called")

    // 1. Attempt to refresh token
    newToken, err := h.tokenService.RefreshToken(r.Context(), r)
    if err != nil {
        switch {
					case errors.Is(err, ErrNoToken):
						h.handleTokenVerificationError(w, "no_token", err)
					case errors.Is(err, ErrInvalidToken):
							h.handleTokenVerificationError(w, "invalid_token", err)
					default:
							h.handleTokenVerificationError(w, "refresh_failed", err)
        }
        return
    }

    // 2. Set new token cookie
    h.tokenService.CreateSessionCookie(w, newToken.Token, newToken.ExpiresAt)

    h.logger.Info("Token refreshed successfully",
        "expiresAt", newToken.ExpiresAt,
    )

    w.WriteHeader(http.StatusOK)
}

// Process user sign out
func (h *AuthHandlers) HandleSignOut(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("HandleSignOut called")

	// 1. Get userID from token
	userID, err := h.tokenService.GetUserIDFromToken(r)
	if err != nil {
			switch {
			case errors.Is(err, ErrNoToken):
					h.handleTokenVerificationError(w, "no_token", err)
			case errors.Is(err, ErrInvalidToken):
					h.handleTokenVerificationError(w, "invalid_token", err)
			default:
					h.handleTokenVerificationError(w, "internal_error", err)
			}
			return
	}

	// 2. Process sign out in service layer
	if err := h.authService.SignOut(r.Context(), userID); err != nil {
			h.handleTokenVerificationError(w, "signout_failed", err)
			return
	}

	// 3. Clear session cookie
	h.tokenService.ClearSessionCookie(w)

	h.logger.Info("User signed out successfully",
			"userID", userID,
	)

	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandlers) HandleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("HandleDeleteAccount called")

	// 1. Get userID from token
	userID, err := h.tokenService.GetUserIDFromToken(r)
	if err != nil {
			switch {
			case errors.Is(err, ErrNoToken):
					h.handleTokenVerificationError(w, "no_token", err)
			case errors.Is(err, ErrInvalidToken):
					h.handleTokenVerificationError(w, "invalid_token", err)
			default:
					h.handleTokenVerificationError(w, "internal_error", err)
			}
			return
	}

	// 2. Process account deletion
	if err := h.authService.ProcessAccountDeletion(r.Context(), userID); err != nil {
			h.logger.Error("Failed to process account deletion",
					"error", err,
					"userID", userID,
			)
			h.handleTokenVerificationError(w, "deletion_failed", err)
			return
	}

	// 3. Clear session cookie
	h.tokenService.ClearSessionCookie(w)

	// 4. Send response
	response := DeleteAccountResponse{
			Message:     "Account marked for deletion",
			RedirectURL: h.getLoginRedirectURL(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
			h.logger.Error("Failed to encode response",
					"error", err,
					"userID", userID,
			)
			h.handleTokenVerificationError(w, "encoding_error", err)
			return
	}

	h.logger.Info("Account deletion processed successfully",
			"userID", userID,
	)
}

// URL Helper functions
// Get the frontend redirect dashboard URL
func (h *AuthHandlers) getDashboardURL() string {
	dashboardURL := os.Getenv("VITE_FRONTEND_DASHBOARD_URL")
	if dashboardURL == "" {
			dashboardURL = "http://localhost:5173/library"
	}
	return dashboardURL
}

func (h *AuthHandlers) getLoginRedirectURL() string {
	redirectURL := os.Getenv("VITE_FRONTEND_LOGIN_URL")
	if redirectURL == "" {
		redirectURL = "http://localhost:5173"
	}
	return redirectURL
}

// Error handling error helper functions
// handleOAuthError logs the error and redirects to login page with error type
func (h *AuthHandlers) handleOAuthError(w http.ResponseWriter, r *http.Request, errorType string, err error) {
	h.logger.Error("OAuth callback error",
			"errorType", errorType,
			"error", err,
	)

	frontendURL := os.Getenv("VITE_FRONTEND_URL")
	if frontendURL == "" {
			frontendURL = "http://localhost:5173"
	}

	redirectURL := fmt.Sprintf("%s/login?error=%s", frontendURL, errorType)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)

	h.logger.Info("Redirecting to login with error",
			"url", redirectURL,
			"errorType", errorType,
			"statusCode", http.StatusSeeOther,
	)
}

// handleTokenVerificationError logs the error and sends appropriate HTTP response
func (h *AuthHandlers) handleTokenVerificationError(w http.ResponseWriter, errorType string, err error) {
	h.logger.Error("Token verification error",
			"errorType", errorType,
			"error", err,
	)

	var statusCode int
	var message string

	switch errorType {
	case "no_token":
			statusCode = http.StatusUnauthorized
			message = "No token cookie found"
	case "invalid_token":
			statusCode = http.StatusUnauthorized
			message = "Invalid token"
	case "user_not_found":
			statusCode = http.StatusNotFound
			message = "User not found"
	default:
			statusCode = http.StatusInternalServerError
			message = "Internal server error"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (h *AuthHandlers) handleMissingRefreshToken(w http.ResponseWriter, r *http.Request) {
    // Generate state using the existing tokenService
    state := h.tokenService.GenerateState()

    // Set the state cookie
    h.tokenService.SetStateCookie(w, state)

    // Revoke current access and redirect to auth with force prompt
    redirectURL := h.config.AuthCodeURL(
        state,
        oauth2.AccessTypeOffline,
        oauth2.ApprovalForce,
    )
    http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}


