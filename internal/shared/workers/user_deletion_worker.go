package workers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/lokeam/bravo-kilo/internal/shared/redis"
	goredis "github.com/redis/go-redis/v9"
)

type UserDeletionService interface {
	SetUserDeletionMarker(ctx context.Context, userID string, expiration time.Duration) error
	GetUserDeletionMarker(ctx context.Context, userID string) (bool, error)
	AddToDeletionQueue(ctx context.Context, userID string) error
	GetDeletionQueue(ctx context.Context) ([]string, error)
	RemoveFromDeletionQueue(ctx context.Context, userID string) error
}

type UserDeletionServiceImpl struct {
	redisClient *redis.RedisClient
	logger      *slog.Logger
}

func NewUserDeletionService(
    redisClient *redis.RedisClient,
    logger *slog.Logger,
) UserDeletionService {
    if redisClient == nil {
        panic("redisClient is nil")
    }
    if logger == nil {
        panic("logger is nil")
    }

    return &UserDeletionServiceImpl{
			redisClient: redisClient,
			logger:      logger,
	}
}

func (s *UserDeletionServiceImpl) SetUserDeletionMarker(ctx context.Context, userID string, expiration time.Duration) error {
	key := fmt.Sprintf("user:deletion:%s", userID)
	return s.redisClient.Set(ctx, key, time.Now().Unix(), expiration)
}

func (s *UserDeletionServiceImpl) GetUserDeletionMarker(ctx context.Context, userID string) (bool, error) {
    s.logger.Info("Getting user deletion marker", "userID", userID)
    key := s.buildKey("userDelete", userID)

    _, err := s.redisClient.Get(ctx, key)
    if err == goredis.Nil {
        s.logger.Info("User deletion marker not found", "userID", userID)
        return false, nil
    }
    if err != nil {
        s.logger.Error("Error getting deletion marker", "userID", userID, "error", err)
        return false, fmt.Errorf("failed to get deletion marker: %w", err)
    }

    s.logger.Info("User deletion marker found", "userID", userID)
    return true, nil
}

func (s *UserDeletionServiceImpl) AddToDeletionQueue(ctx context.Context, userID string) error {
    return s.redisClient.AddToDeletionQueue(ctx, redis.PrefixDeletionQueue, userID)
}

func (s *UserDeletionServiceImpl) GetDeletionQueue(ctx context.Context) ([]string, error) {
    return s.redisClient.GetDeletionQueueItems(ctx, redis.PrefixDeletionQueue, 0, -1)
}

func (s *UserDeletionServiceImpl) RemoveFromDeletionQueue(ctx context.Context, userID string) error {
    return s.redisClient.RemoveFromDeletionQueue(ctx, redis.PrefixDeletionQueue, 0, userID)
}

func (s *UserDeletionServiceImpl) buildKey(operation string, params ...interface{}) string {
    switch operation {
    case "userDelete":
        if len(params) > 0 {
            return fmt.Sprintf("%s%v", redis.PrefixUserDelete, params[0])
        }
    }
    return ""
}