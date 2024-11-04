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
	refreshToken, err := s.tokenRepo.GetRefreshTokenByUserID(userID)
	if err != nil {
			return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}
	if refreshToken == "" {
			return nil, ErrNoRefreshToken
	}

	// Use refresh token to get new access token
	return s.RefreshAccessToken(ctx, refreshToken)
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