package authservices

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lokeam/bravo-kilo/internal/shared/crypto"
	"github.com/lokeam/bravo-kilo/internal/shared/models"
	"github.com/lokeam/bravo-kilo/internal/shared/types"
)
type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

type TokenService interface {
	CreateJWT(userID int, expirationTime time.Time) (string, error)
	VerifyJWT(tokenStr string) (*types.Claims, error)
	RefreshJWT(oldToken string) (string, error)
	RefreshToken(r *http.Request) (*TokenResponse, error)
	GetUserIDFromToken(r *http.Request) (int, error)
	CreateSessionCookie(w http.ResponseWriter, token string, expiry time.Time)
	ClearSessionCookie(w http.ResponseWriter)
	GenerateState() string
	SetStateCookie(w http.ResponseWriter, state string)
	VerifyStateCookie(r *http.Request, state string) error
	Rotate(userID int, newToken, oldToken string, expiry time.Time) error
}

type TokenServiceImpl struct {
	logger       *slog.Logger
	tokenModel   models.TokenModel
	publicKey    *rsa.PublicKey
	privateKey   *rsa.PrivateKey
	isProduction bool
}

type TokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// MaxRefreshCount represents the maximum number of times a token can be refreshed
// Based on: 7 days * 24 hours * 60 minutes / 55 minutes ≈ 183 legitimate refreshes
// Adding ~30% buffer for network issues and retries
const MaxRefreshCount = 250


func NewTokenService(
	logger *slog.Logger,
	tokenModel models.TokenModel,
	privateKey *rsa.PrivateKey,
	publicKey *rsa.PublicKey,
	isProduction bool,
) TokenService {
	if logger == nil {
		panic("logger cannot be nil")
	}
	if tokenModel == nil {
			panic("tokenModel cannot be nil")
	}
	if privateKey == nil {
			panic("privateKey cannot be nil")
	}
	if publicKey == nil {
			panic("publicKey cannot be nil")
	}

	return &TokenServiceImpl{
		logger:       logger,
		tokenModel:   tokenModel,
		privateKey:   privateKey,
		publicKey:    publicKey,
		isProduction: isProduction,
}
}

// Generate random state for CSRF protection
func (ts *TokenServiceImpl) GenerateState() string {
	byteSlice := make([]byte, 16)
	if _, err := rand.Read(byteSlice); err != nil {
			ts.logger.Error("Error generating random state", "error", err)
			return ""
	}
	return base64.URLEncoding.EncodeToString(byteSlice)
}


func (s *TokenServiceImpl) CreateJWT(userID int, expirationTime time.Time) (string, error) {
	claims := &types.Claims{
			UserID: userID,
			RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(expirationTime),
			},
	}

	return crypto.SignToken(claims, s.privateKey)
}

func (s *TokenServiceImpl) VerifyJWT(tokenStr string) (*types.Claims, error) {
	token, err := crypto.VerifyToken(tokenStr, s.publicKey)
	if err != nil || !token.Valid {
			return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*types.Claims)
	if !ok {
			return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// GetUserIDFromToken extracts userID from the JWT in the request cookie
func (s *TokenServiceImpl) GetUserIDFromToken(r *http.Request) (int, error) {
	cookie, err := r.Cookie("token")
	if err != nil {
			return 0, fmt.Errorf("no token cookie: %w", err)
	}

	claims, err := s.VerifyJWT(cookie.Value)
	if err != nil {
			return 0, fmt.Errorf("invalid token: %w", err)
	}

	return claims.UserID, nil
}

func (s *TokenServiceImpl) RefreshJWT(oldToken string) (string, error) {
	// Verify the old token
	claims, err := s.VerifyJWT(oldToken)
	if err != nil {
			return "", fmt.Errorf("invalid token: %w", err)
	}

	// Check if token needs refresh (within 5 minutes of expiration)
	if time.Until(claims.ExpiresAt.Time) > 5*time.Minute {
			return "", fmt.Errorf("token is not close to expiration")
	}

	// Generate new token with extended expiration
	expirationTime := time.Now().Add(7 * 24 * time.Hour)
	newToken, err := s.CreateJWT(claims.UserID, expirationTime)
	if err != nil {
			return "", fmt.Errorf("failed to create new token: %w", err)
	}

	// Rotate token in database
	if err := s.tokenModel.Rotate(claims.UserID, newToken, oldToken, expirationTime); err != nil {
			return "", fmt.Errorf("failed to rotate token: %w", err)
	}

	return newToken, nil
}

func (s *TokenServiceImpl) RefreshToken(r *http.Request) (*TokenResponse, error) {
	s.logger.Debug("Starting token refresh process",
		"component", "token_service",
	)

	// 1. Get and verify current token
	cookie, err := r.Cookie("token")
	if err != nil {
			s.logger.Error("No token cookie found",
				"error", err,
				"component", "token_service",
			)
			return nil, fmt.Errorf("no token cookie: %w", err)
	}

	claims, err := s.VerifyJWT(cookie.Value)
	if err != nil {
		s.logger.Error("Invalid token during refresh",
			"error", err,
			"component", "token_service",
		)
			return nil, fmt.Errorf("invalid token: %w", err)
	}

	s.logger.Debug("Token verification successful",
		"userID", claims.UserID,
		"tokenExpiresAt", claims.ExpiresAt.Time,
		"component", "token_service",
  )

	// 2. Check if token needs refresh
	timeUntilExpiry := time.Until(claims.ExpiresAt.Time)
	if timeUntilExpiry > 5*time.Minute {
			s.logger.Info("Token refresh rejected - not close to expiration",
					"userID", claims.UserID,
					"timeUntilExpiry", timeUntilExpiry,
					"component", "token_service",
			)
			return nil, fmt.Errorf("token is not close to expiration")
	}

	// UPDATED - Invalidate all previous tokens for user
	currentToken, err := s.tokenModel.GetLatestActiveToken(claims.UserID)
	if err != nil {
		s.logger.Error("Failed to get latest active token",
			"error", err,
			"userID", claims.UserID,
			"component", "token_service",
		)
    } else if currentToken != nil {
			s.logger.Info("Current token status",
					"userID", claims.UserID,
					"refreshCount", currentToken.RefreshCount,
					"maxAllowed", MaxRefreshCount,
					"component", "token_service",
			)

			if currentToken.RefreshCount > MaxRefreshCount {
					s.logger.Warn("Excessive token refreshes detected",
							"userID", claims.UserID,
							"refreshCount", currentToken.RefreshCount,
							"maxAllowed", MaxRefreshCount,
							"component", "token_service",
					)
			}
	}

  // Invalidate previous tokens
  if err := s.tokenModel.DeleteByUserID(claims.UserID); err != nil {
			s.logger.Error("Failed to invalidate previous tokens",
					"error", err,
					"userID", claims.UserID,
					"component", "token_service",
			)
	} else {
			s.logger.Debug("Successfully invalidated previous tokens",
					"userID", claims.UserID,
					"component", "token_service",
			)
	}

	// 3. Generate new token with extended expiration (1 week)
	expirationTime := time.Now().Add(7 * 24 * time.Hour)
	newToken, err := s.CreateJWT(claims.UserID, expirationTime)
	if err != nil {
		s.logger.Error("Failed to create new JWT",
			"error", err,
			"userID", claims.UserID,
			"component", "token_service",
		)
		return nil, fmt.Errorf("failed to create new token: %w", err)
	}

	// 4. Rotate token in database
	if err := s.tokenModel.Rotate(claims.UserID, newToken, cookie.Value, expirationTime); err != nil {
		s.logger.Error("Failed to rotate token",
			"error", err,
			"userID", claims.UserID,
			"component", "token_service",
    )
		return nil, fmt.Errorf("failed to rotate token: %w", err)
	}

	s.logger.Info("Successfully refreshed token",
		"userID", claims.UserID,
		"newExpiryTime", expirationTime,
		"component", "token_service",
	)

	return &TokenResponse{
			Token:     newToken,
			ExpiresAt: expirationTime,
	}, nil
}

func (s *TokenServiceImpl) CreateSessionCookie(w http.ResponseWriter, token string, expiry time.Time) {
	http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    token,
			Expires:  expiry,
			HttpOnly: true,
			Secure:   s.isProduction,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
	})
}

func (s *TokenServiceImpl) SetStateCookie(w http.ResponseWriter, state string) {
	http.SetCookie(w, &http.Cookie{
			Name:     "oauthstate",
			Value:    state,
			Expires:  time.Now().Add(10 * time.Minute),
			HttpOnly: true,
			Secure:   s.isProduction,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
	})
}

func (s *TokenServiceImpl) VerifyStateCookie(r *http.Request, state string) error {
	cookie, err := r.Cookie("oauthstate")
	if err != nil {
			return fmt.Errorf("state cookie not found: %w", err)
	}
	if cookie.Value != state {
			return fmt.Errorf("state mismatch")
	}
	return nil
}

func (s *TokenServiceImpl) ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    "",
			Expires:  time.Now().Add(-1 * time.Hour), // Past time to ensure deletion
			HttpOnly: true,
			Secure:   s.isProduction,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
	})
}


// Todo: refactor h.models.TokenModel.Rotate
func (s *TokenServiceImpl) Rotate(userID int, newToken, oldToken string, expiry time.Time) error {
	return s.tokenModel.Rotate(userID, newToken, oldToken, expiry)
}