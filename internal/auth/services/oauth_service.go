package authservices

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/lokeam/bravo-kilo/internal/shared/models"
	"golang.org/x/oauth2"
)

type OAuthService interface {
	GetAuthURL(state string) string
	GetAccessToken(ctx context.Context, userID int) (*oauth2.Token, error)
	GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*oauth2.Token, error)
	ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error)
	VerifyIDToken(ctx context.Context, rawIDToken string) (*oidc.IDToken, error)
	GetConfig() *oauth2.Config
}

type OAuthServiceImpl struct {
    logger    *slog.Logger
    config    *oauth2.Config
    provider  *oidc.Provider
    verifier  *oidc.IDTokenVerifier
		tokenRepo models.TokenModel
}

type UserInfo struct {
    Email         string
    EmailVerified bool
    Name          string
    Picture       string
    Sub           string
}

var ErrNoRefreshToken = errors.New("no valid refresh token found")

func NewOAuthService(
	logger *slog.Logger,
	tokenRepo models.TokenModel,
) (OAuthService, error) {
    ctx := context.Background()

		// Initialize OIDC provider
    provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
    if err != nil {
        return nil, fmt.Errorf("failed to get OIDC provider: %w", err)
    }

		// Initialize OAuth2 config with OIDC scopes
    config := &oauth2.Config{
        RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URI"),
        ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
        ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
        Scopes: []string{
            oidc.ScopeOpenID,
            "profile",
            "email",
            "https://www.googleapis.com/auth/books",
        },
        Endpoint: provider.Endpoint(),
    }

    // Initialize OIDC Verifier
		verifier := provider.Verifier(&oidc.Config{
			ClientID: config.ClientID,
		})

    return &OAuthServiceImpl{
			logger:    logger,
			config:    config,
			provider:  provider,
			verifier:  verifier,
			tokenRepo: tokenRepo,
    }, nil
}


func (s *OAuthServiceImpl) GetAuthURL(state string) string {
	return s.config.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce,
	)
}

// InitializeOAuth initializes the OAuth2 configuration and OIDC provider
func (s *OAuthServiceImpl) InitializeOAuth() error {
	ctx := context.Background()

	// Initialize OIDC provider
	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
			return fmt.Errorf("failed to get OIDC provider: %w", err)
	}
	s.provider = provider

	// Initialize OAuth2 config
	s.config = &oauth2.Config{
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URI"),
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			Scopes: []string{
					oidc.ScopeOpenID,
					"profile",
					"email",
					"https://www.googleapis.com/auth/books",
			},
			Endpoint: provider.Endpoint(),
	}

	// Initialize verifier
	s.verifier = provider.Verifier(&oidc.Config{
			ClientID: s.config.ClientID,
	})

	s.logger.Info("OAuth service initialized successfully")
	return nil
}

func (s *OAuthServiceImpl) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := s.config.Exchange(ctx, code)
	if err != nil {
			return nil, fmt.Errorf("code exchange failed: %w", err)
	}
	return token, nil
}

func (s *OAuthServiceImpl) VerifyIDToken(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	return s.verifier.Verify(ctx, rawIDToken)
}

func (s *OAuthServiceImpl) GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
			return nil, fmt.Errorf("no id_token in token response")
	}

	idToken, err := s.VerifyIDToken(ctx, rawIDToken)
	if err != nil {
			return nil, fmt.Errorf("failed to verify ID token: %w", err)
	}

	var claims struct {
			Email         string `json:"email"`
			EmailVerified bool   `json:"email_verified"`
			Name          string `json:"name"`
			Picture       string `json:"picture"`
	}
	if err := idToken.Claims(&claims); err != nil {
			return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	return &UserInfo{
			Email:         claims.Email,
			EmailVerified: claims.EmailVerified,
			Name:          claims.Name,
			Picture:       claims.Picture,
			Sub:           idToken.Subject,
	}, nil
}

// GetAccessToken retrieves a valid OAuth access token for the user
func (s *OAuthServiceImpl) GetAccessToken(ctx context.Context, userID int) (*oauth2.Token, error) {
	// Get refresh token from repository
	//refreshToken, err := s.tokenRepo.GetRefreshTokenByUserID(userID)
	s.logger.Info("Starting access token retrieval",
		"userID", userID,
		"component", "oauth_service",
	)

	tokenRecord, err := s.tokenRepo.GetLatestActiveToken(userID)
	if err != nil {
		s.logger.Error("Failed to get active token",
				"error", err,
				"userID", userID,
				"component", "oauth_service",
		)
		return nil, fmt.Errorf("token retrieval error: %w", err)
	}
	if tokenRecord == nil {
			return nil, ErrNoRefreshToken
	}
	s.logger.Debug("Active token found",
		"userID", userID,
		"tokenID", tokenRecord.ID,
		"tokenExpiry", tokenRecord.TokenExpiry,
		"lastUsed", tokenRecord.LastUsedAt,
		"component", "oauth_service",
	)

	// Track token usage
	if err := s.tokenRepo.UpdateLastUsed(tokenRecord.ID); err != nil {
		s.logger.Error("Failed to update token usage",
				"error", err,
				"tokenID", tokenRecord.ID,
				"component", "oauth_service",
		)
		// Non-critical error, continue
	} else {
		s.logger.Debug("Updated token last used timestamp",
			"tokenID", tokenRecord.ID,
			"userID", userID,
			"component", "oauth_service",
		)
	}

	// Use existing RefreshAccessToken method
	newToken, err := s.RefreshAccessToken(ctx, tokenRecord.RefreshToken)
	if err != nil {
			s.logger.Error("Token refresh failed",
				"error", err,
				"tokenID", tokenRecord.ID,
				"userID", userID,
				"component", "oauth_service",
      )

			// Record failed refresh attempt
			if recordErr := s.tokenRepo.RecordRefreshAttempt(tokenRecord.ID, false, err.Error()); recordErr != nil {
				s.logger.Error("Failed to record refresh attempt",
					"error", recordErr,
					"tokenID", tokenRecord.ID,
					"userID", userID,
					"component", "oauth_service",
				)
			}
			return nil, fmt.Errorf("token refresh failed: %w", err)
	}

	// Record successful refresh
	if err := s.tokenRepo.RecordRefreshAttempt(tokenRecord.ID, true, ""); err != nil {
		s.logger.Error("Failed to record successful refresh attempt",
				"error", err,
				"tokenID", tokenRecord.ID,
				"component", "oauth_service",
		)
	}
	s.logger.Info("Successfully refreshed access token",
		"userID", userID,
		"tokenID", tokenRecord.ID,
		"newTokenExpiry", newToken.Expiry,
		"hasNewRefreshToken", newToken.RefreshToken != "",
		"component", "oauth_service",
	)

	// Handle token rotation if we got a new refresh token
	if newToken.RefreshToken != "" && newToken.RefreshToken != tokenRecord.RefreshToken {
		s.logger.Info("New refresh token received, attempting rotation",
			"userID", userID,
			"tokenID", tokenRecord.ID,
			"component", "oauth_service",
		)

		err = s.tokenRepo.Rotate(
				userID,
				newToken.RefreshToken,
				tokenRecord.RefreshToken,
				newToken.Expiry,
		)
		if err != nil {
				s.logger.Error("Token rotation failed",
						"error", err,
						"userID", userID,
						"component", "oauth_service",
				)
				// Continue as we still have valid token
		} else {
			s.logger.Info("Token rotation successful",
				"userID", userID,
				"tokenID", tokenRecord.ID,
				"newExpiry", newToken.Expiry,
				"component", "oauth_service",
			)
		}
	}

	return newToken, nil
}

func (s *OAuthServiceImpl) RefreshAccessToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	tokenSource := s.config.TokenSource(ctx, &oauth2.Token{
			RefreshToken: refreshToken,
	})

	newToken, err := tokenSource.Token()
	if err != nil {
			return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return newToken, nil
}

func (s *OAuthServiceImpl) GetConfig() *oauth2.Config {
	return s.config
}