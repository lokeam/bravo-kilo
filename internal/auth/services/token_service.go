package authservices

import (
	"context"
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
	RefreshToken(ctx context.Context,r *http.Request) (*TokenResponse, error)
	GetUserIDFromToken(r *http.Request) (int, error)
	CreateSessionCookie(w http.ResponseWriter, token string, expiry time.Time)
	ClearSessionCookie(w http.ResponseWriter)
	GenerateState() string
	SetStateCookie(w http.ResponseWriter, state string)
	VerifyStateCookie(r *http.Request, state string) error
	Rotate(ctx context.Context, userID int, newToken, oldToken string, expiry time.Time) error
	ShouldRefreshToken(token *models.Token) bool
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

const (

	// MaxRefreshCount represents the maximum number of times a token can be refreshed
	// Based on: 7 days * 24 hours * 60 minutes / 55 minutes ≈ 183 legitimate refreshes
	// Adding ~30% buffer for network issues and retries
	MaxRefreshCount = 250

	// Token lifetime is 24 hours
	TokenLifetime = 24 * time.Hour

	// Start refresh attempts when token has 30% of lifetime remaining (17 hours into token lifetime)
	RefreshThreshold = TokenLifetime * 7 / 10
)

var (
	ErrTokenExpired        = fmt.Errorf("token expired")
	ErrTokenInvalid        = fmt.Errorf("token invalid")
	ErrTokenReused         = fmt.Errorf("token reuse detected")
	ErrRefreshRateExceeded = fmt.Errorf("refresh rate exceeded")
	ErrTokenFamilyRevoked  = fmt.Errorf("token family revoked")
)


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
			return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*types.Claims)
	if !ok {
			return nil, ErrTokenInvalid
	}

	return claims, nil
}

// GetUserIDFromToken extracts userID from the JWT in the request cookie
func (s *TokenServiceImpl) GetUserIDFromToken(r *http.Request) (int, error) {
	cookie, err := r.Cookie("token")
	if err != nil {
			return 0, ErrTokenInvalid
	}

	claims, err := s.VerifyJWT(cookie.Value)
	if err != nil {
			return 0, ErrTokenInvalid
	}

	return claims.UserID, nil
}

func (s *TokenServiceImpl) RefreshToken(ctx context.Context, r *http.Request) (*TokenResponse, error) {
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
			return nil, ErrNoRefreshToken
	}

	// 2. Verify JWT signature and claims
	claims, err := s.VerifyJWT(cookie.Value)
	if err != nil {
		s.logger.Error("Invalid token during refresh",
			"error", err,
			"component", "token_service",
		)
			return nil, ErrTokenInvalid
	}

	// Check context before proceeding
	if err := ctx.Err(); err != nil {
		s.logger.Error("Context cancelled during refresh",
			"error", err,
			"component", "token_service",
		)
		return nil, fmt.Errorf("operation cancelled: %w", err)
	}

	s.logger.Debug("Token verification successful",
		"userID", claims.UserID,
		"tokenExpiresAt", claims.ExpiresAt.Time,
		"component", "token_service",
  )

	// 3. Validate refresh token record
	currentToken, err := s.tokenModel.GetLatestActiveToken(ctx, claims.UserID)
	if err != nil {
		s.logger.Error("Failed to validate refresh token",
			"error", err,
			"userID", claims.UserID,
			"component", "token_service",
		)
		return nil, ErrTokenExpired
	}
	if currentToken == nil {
		return nil, fmt.Errorf("token not found or expired")
	}

	// Validate token family
	if err := s.validateTokenFamily(ctx, currentToken); err != nil {
		s.logger.Error("Token family validation failed",
			"error", err,
			"userID", claims.UserID,
			"component", "token_service",
		)
		return nil, err
	}

	// 4. Verify userID matches JWT and refresh token record
	if claims.UserID != currentToken.UserID {
		s.logger.Error("User ID mismatch between JWT and token record",
			"jwtUserID", claims.UserID,
			"tokenUserID", currentToken.UserID,
			"component", "token_service",
		)
		return nil, fmt.Errorf("user ID mismatch between JWT and token record")
	}

	// 5. Check if refresh is needed
	if !s.ShouldRefreshToken(currentToken) {
		s.logger.Debug("Token refresh not needed yet",
			"userID", claims.UserID,
			"tokenID", currentToken.ID,
			"expiresIn", time.Until(currentToken.TokenExpiry),
			"component", "token_service",
		)
		return &TokenResponse{
			Token:     cookie.Value,
			ExpiresAt: currentToken.TokenExpiry,
		}, nil
	}

	// 6. Generate new token
	expirationTime := time.Now().Add(TokenLifetime)
	newToken, err := s.CreateJWT(claims.UserID, expirationTime)
	if err != nil {
		s.logger.Error("Failed to create new JWT",
			"error", err,
			"userID", claims.UserID,
			"component", "token_service",
		)
		return nil, fmt.Errorf("failed to create new token: %w", err)
	}

	// 7. Rotate token in database, maintain family
	if err := s.tokenModel.Rotate(ctx, claims.UserID, newToken, cookie.Value, expirationTime); err != nil {
		s.logger.Error("Failed to rotate token",
			"error", err,
			"userID", claims.UserID,
			"component", "token_service",
		)
	}

	s.logger.Info("Successfully refreshed token",
		"userID", claims.UserID,
		"tokenID", currentToken.ID,
		"familyID", currentToken.FamilyID,
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
func (s *TokenServiceImpl) Rotate(ctx context.Context,userID int, newToken, oldToken string, expiry time.Time) error {
	return s.tokenModel.Rotate(ctx,userID, newToken, oldToken, expiry)
}

func (s *TokenServiceImpl) ShouldRefreshToken(token *models.Token) bool {
	if token == nil {
		return false
	}

	timeUntilExpiry := time.Until(token.TokenExpiry)

	s.logger.Debug("Checking token refresh status",
		"tokenID", token.ID,
		"timeUntilExpiry", timeUntilExpiry,
		"refreshThreshold", RefreshThreshold,
		"component", "token_service",
	)

	return timeUntilExpiry < RefreshThreshold
}

func (s *TokenServiceImpl) validateTokenFamily(ctx context.Context, token *models.Token) error {
	if token.FamilyID == "" {
			s.logger.Error("Invalid token family",
					"userID", token.UserID,
					"component", "token_service",
			)
			return fmt.Errorf("invalid token family")
	}

	isRevoked, err := s.tokenModel.IsFamilyRevoked(ctx, token.FamilyID)
	if err != nil {
			s.logger.Error("Failed to check family status",
					"error", err,
					"familyID", token.FamilyID,
					"component", "token_service",
			)
			return fmt.Errorf("check family status: %w", err)
	}
	if isRevoked {
			s.logger.Warn("Attempt to use token from revoked family",
					"familyID", token.FamilyID,
					"userID", token.UserID,
					"component", "token_service",
			)
			return ErrTokenFamilyRevoked
	}
	return nil
}