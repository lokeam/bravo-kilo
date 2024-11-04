package cache

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lokeam/bravo-kilo/internal/shared/redis"
)

type RedisInvalidator struct {
	redisClient    *redis.RedisClient
	logger         *slog.Logger
}

func NewRedisInvalidtor(client *redis.RedisClient, logger *slog.Logger) *RedisInvalidator {
	if client == nil {
		panic("redis client cannot be nil")
	}

	return &RedisInvalidator{
		redisClient: client,
		logger:      logger,
	}
}

func (r *RedisInvalidator) InvalidateL2Cache(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	// Delete keys from Redis
	if err := r.redisClient.Delete(ctx, keys...); err != nil {
		return fmt.Errorf("redis invalidator - failed to invalidate Redis cache: %w", err)
	}

	return nil
}

func (r *RedisInvalidator) GetCacheKeys(itemID, userID int) []string {
	// Typically overwritten by domain-specific invalidators
	return []string{
		fmt.Sprintf("default:user:%d:item:%d", userID, itemID),
	}
}

