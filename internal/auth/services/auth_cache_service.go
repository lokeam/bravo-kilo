package authservices

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/lokeam/bravo-kilo/config"
	"github.com/lokeam/bravo-kilo/internal/shared/models"
	"github.com/lokeam/bravo-kilo/internal/shared/redis"
	"github.com/lokeam/bravo-kilo/internal/shared/transaction"
	goredis "github.com/redis/go-redis/v9"
)

type AuthCacheService interface {
	SetAuthToken(ctx context.Context, userID string, token string) error
	GetAuthToken(ctx context.Context, userID string) (string, bool, error)
	InvalidateAuthToken(ctx context.Context, userID string) error
	ProcessAccountDeletion(ctx context.Context, userID int) error
}

type AuthCacheServiceImpl struct {
	logger           *slog.Logger
	redisClient      *redis.RedisClient
	config           *config.Config
	userRepo         models.UserRepository
	tokenRepo        models.TokenModel
	dbManager        transaction.DBManager
	models           models.Models
	userDeletionService  UserDeletionService
}

func NewAuthCacheService(
	redisClient *redis.RedisClient,
	logger *slog.Logger,
	config *config.Config,
	dbManager transaction.DBManager,
	userRepo models.UserRepository,
	tokenRepo models.TokenModel,
	models models.Models,
	userDeletionService UserDeletionService,
) AuthCacheService {
	if redisClient == nil {
			panic("redisClient is nil")
	}
	if logger == nil {
			panic("logger is nil")
	}

	return &AuthCacheServiceImpl{
		logger:           logger.With("component", "auth_cache_service"),
		redisClient:      redisClient,
		config:           config,
		dbManager:        dbManager,
		userRepo:         userRepo,
		tokenRepo:        tokenRepo,
		models:           models,
		userDeletionService:  userDeletionService,
	}
}

func (s *AuthCacheServiceImpl) SetAuthToken(ctx context.Context, userID string, token string) error {
	key := s.buildKey("auth", userID)
	return s.redisClient.Set(ctx, key, token, s.redisClient.GetConfig().CacheConfig.AuthTokenExpiration)
}

func (s *AuthCacheServiceImpl) GetAuthToken(ctx context.Context, userID string) (string, bool, error) {
	key := s.buildKey("auth", userID)
	token, err := s.redisClient.Get(ctx, key)
	if err != nil {
			if err == goredis.Nil {
					return "", false, nil
			}
			s.logger.Error("cache fetch error", "key", key, "error", err)
			return "", false, fmt.Errorf("cache fetch error: %w", err)
	}
	return token, true, nil
}

func (s *AuthCacheServiceImpl) InvalidateAuthToken(ctx context.Context, userID string) error {
	key := s.buildKey("auth", userID)
	return s.redisClient.Delete(ctx, key)
}

func (s *AuthCacheServiceImpl) buildKey(operation string, params ...interface{}) string {
	switch operation {
	case "auth":
			if len(params) > 0 {
					return fmt.Sprintf("%s:%v", redis.PrefixAuthToken, params[0])
			}
	}
	return ""
}

func (s *AuthCacheServiceImpl) ProcessAccountDeletion(ctx context.Context, userID int) error {
	s.logger.Info("Processing account deletion", "userID", userID)

	// Begin transaction
	tx, err := s.dbManager.BeginTransaction(ctx)
	if err != nil {
			return fmt.Errorf("error beginning transaction: %w", err)
	}
	defer s.dbManager.RollbackTransaction(tx)

	// 1. Mark user for deletion
	deletionTime := time.Now()
	if err := s.userRepo.MarkForDeletion(ctx, tx, userID, deletionTime); err != nil {
			return fmt.Errorf("error marking user for deletion: %w", err)
	}

	// 2. Delete refresh tokens
	if err := s.tokenRepo.DeleteByUserID(userID); err != nil {
			return fmt.Errorf("error deleting refresh tokens: %w", err)
	}

	// 3. Set deletion marker in Redis
	if err := s.userDeletionService.SetUserDeletionMarker(ctx, strconv.Itoa(userID), config.AppConfig.UserDeletionMarkerExpiration); err != nil {
			s.logger.Error("Error setting user deletion marker in Redis", "error", err)
			// Continue execution as this is not critical
	}

	// 4. Add to deletion queue
	if err := s.userDeletionService.AddToDeletionQueue(ctx, strconv.Itoa(userID)); err != nil {
			s.logger.Error("Error adding user to deletion queue in Redis", "error", err)
			// Continue execution as this is not critical
	}

	// Commit transaction
	if err := s.dbManager.CommitTransaction(tx); err != nil {
			return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}