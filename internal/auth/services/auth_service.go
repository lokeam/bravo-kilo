package authservices

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/lokeam/bravo-kilo/internal/shared/models"
	"github.com/lokeam/bravo-kilo/internal/shared/transaction"
	"github.com/lokeam/bravo-kilo/internal/shared/utils"
)

type AuthService interface {
    ProcessGoogleAuth(ctx context.Context, code string) (*AuthResponse, error)
    VerifyAndGetUserInfo(r *http.Request) (*models.User, error)
    SignOut(ctx context.Context, userID int) error
    ProcessAccountDeletion(ctx context.Context, userID int) error
}

type AuthResponse struct {
    Token     string         `json:"token"`
    ExpiresAt time.Time      `json:"expiresAt"`
    User      *models.User   `json:"user"`
}

type UserSession struct {
    User      *models.User
    ExpiresAt time.Time
}

type AuthServiceImpl struct {
    logger              *slog.Logger
    dbManager           transaction.DBManager
    userRepo            models.UserRepository
    tokenRepo           models.TokenModel
    oauthService        OAuthService
    tokenService        TokenService
    userDeletionService UserDeletionService
}

var (
    ErrInvalidToken = errors.New("invalid token")
    ErrUserNotFound = errors.New("user not found")
)

func NewAuthService(
    logger *slog.Logger,
    dbManager transaction.DBManager,
    userRepo models.UserRepository,
    tokenRepo models.TokenModel,
    oauthService OAuthService,
    tokenService TokenService,
    userDeletionService UserDeletionService,
) AuthService {
    if logger == nil {
        panic("logger is nil")
    }
    if dbManager == nil {
        panic("dbManager is nil")
    }
    if userRepo == nil {
        panic("userRepo is nil")
    }
    if oauthService == nil {
        panic("oauthService is nil")
    }
    if tokenService == nil {
        panic("tokenService is nil")
    }
    if userDeletionService == nil {
        panic("userDeletionService is nil")
    }

    return &AuthServiceImpl{
        logger:              logger.With("component", "auth_service"),
        dbManager:           dbManager,
        userRepo:            userRepo,
        tokenRepo:           tokenRepo,
        oauthService:        oauthService,
        tokenService:        tokenService,
        userDeletionService: userDeletionService,
    }
}


func (as *AuthServiceImpl) ProcessGoogleAuth(ctx context.Context, code string) (*AuthResponse, error) {
    // 1. Exchange code for token
    token, err := as.oauthService.ExchangeCode(ctx, code)
    if err != nil {
        return nil, fmt.Errorf("code exchange failed: %w", err)
    }

    // 2. Get user info from token
    userInfo, err := as.oauthService.GetUserInfo(ctx, token)
    if err != nil {
        return nil, fmt.Errorf("failed to get user info: %w", err)
    }

    // Verify email is present and verified
    if !userInfo.EmailVerified {
        return nil, fmt.Errorf("email not verified")
    }

    // 3. Get or create user
    firstName, lastName := utils.SplitFullName(userInfo.Name)
    user, err := as.userRepo.GetByEmail(userInfo.Email)
    if err == sql.ErrNoRows {
        // Create new user
        user = &models.User{
            Email:     userInfo.Email,
            FirstName: firstName,
            LastName:  lastName,
            Picture:   userInfo.Picture,
        }
        userID, err := as.userRepo.Insert(*user)
        if err != nil {
            return nil, fmt.Errorf("failed to create user: %w", err)
        }
        user.ID = userID
    } else if err != nil {
        return nil, fmt.Errorf("failed to check existing user: %w", err)
    }

    // 4. Store refresh token
    if token.RefreshToken != "" {
        tokenRecord := models.Token{
            UserID:       user.ID,
            RefreshToken: token.RefreshToken,
            TokenExpiry:  token.Expiry,
        }
        if err := as.tokenRepo.Insert(tokenRecord); err != nil {
            return nil, fmt.Errorf("failed to store refresh token: %w", err)
        }
    }

    // 5. Create JWT
    expirationTime := time.Now().Add(60 * time.Minute)
    jwtString, err := as.tokenService.CreateJWT(user.ID, expirationTime)
    if err != nil {
        return nil, fmt.Errorf("failed to create JWT: %w", err)
    }

    return &AuthResponse{
        Token:     jwtString,
        ExpiresAt: expirationTime,
        User:      user,
    }, nil
}

func (as *AuthServiceImpl) VerifyAndGetUserInfo(r *http.Request) (*models.User, error) {
    // 1. Get and validate token from cookie
    userID, err := as.tokenService.GetUserIDFromToken(r)
    if err != nil {
        as.logger.Error("Failed to get user ID from token", "error", err)
        return nil, ErrInvalidToken
    }

    // 2. Get user from database
    user, err := as.userRepo.GetByID(userID)
    if err != nil {
        if err == sql.ErrNoRows {
            as.logger.Error("User not found", "userID", userID)
            return nil, ErrUserNotFound
        }
        as.logger.Error("Database error while fetching user",
            "error", err,
            "userID", userID,
        )
        return nil, fmt.Errorf("failed to fetch user: %w", err)
    }

    as.logger.Info("User verified and retrieved",
        "userID", user.ID,
        "email", user.Email,
    )

    return user, nil
}

func (as *AuthServiceImpl) SignOut(ctx context.Context, userID int) error {
    // Delete refresh token from database
    if err := as.tokenRepo.DeleteByUserID(userID); err != nil {
        as.logger.Error("Failed to delete refresh token",
            "error", err,
            "userID", userID,
        )
        return fmt.Errorf("failed to delete refresh token: %w", err)
    }

    as.logger.Info("User signed out successfully",
        "userID", userID,
    )

    return nil
}

func (as *AuthServiceImpl) ProcessAccountDeletion(ctx context.Context, userID int) error {
    // Start transaction with context
    tx, err := as.dbManager.BeginTransaction(ctx)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer as.dbManager.RollbackTransaction(tx)

    // Mark user for deletion with context and deletion time
    deletionTime := time.Now().Add(24 * time.Hour) // Mark for deletion in 24 hours
    if err := as.userRepo.MarkForDeletion(ctx, tx, userID, deletionTime); err != nil {
        as.logger.Error("Failed to mark user for deletion",
            "error", err,
            "userID", userID,
            "deletionTime", deletionTime,
        )
        return fmt.Errorf("failed to mark user for deletion: %w", err)
    }

    // Rest of the method remains the same...
    if err := as.tokenRepo.DeleteByUserID(userID); err != nil {
        as.logger.Error("Failed to delete refresh token",
            "error", err,
            "userID", userID,
        )
        return fmt.Errorf("failed to delete refresh token: %w", err)
    }

    userIDStr := fmt.Sprintf("%d", userID)
    if err := as.userDeletionService.SetUserDeletionMarker(ctx, userIDStr, 24*time.Hour); err != nil {
        as.logger.Error("Failed to set user deletion marker",
            "error", err,
            "userID", userID,
        )
        // Continue execution as this is not critical
    }

    if err := as.userDeletionService.AddToDeletionQueue(ctx, userIDStr); err != nil {
        as.logger.Error("Failed to add user to deletion queue",
            "error", err,
            "userID", userID,
        )
        // Continue execution as this is not critical
    }

    // Commit transaction
    if err := as.dbManager.CommitTransaction(tx); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    as.logger.Info("Account marked for deletion successfully",
        "userID", userID,
        "deletionTime", deletionTime,
    )

    return nil
}