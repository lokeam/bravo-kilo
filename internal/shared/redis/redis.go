// Replace entire file with:
package redis

import (
	"context"
	"log/slog"
	"sync"
)

var (
    defaultClient *RedisClient
    once         sync.Once
)

// GetClient returns the default Redis client instance
func GetClient() *RedisClient {
    return defaultClient
}

// For backward compatibility
func InitRedis(ctx context.Context, logger *slog.Logger) (*RedisClient, error) {
    var err error
    once.Do(func() {
        cfg := NewRedisConfig()
        if err = cfg.LoadFromEnv(); err != nil {
            return
        }

        defaultClient, err = NewRedisClient(cfg, logger)
        if err != nil {
            return
        }

        err = defaultClient.Connect(ctx)
    })

    if err != nil {
        return nil, err
    }

    return defaultClient, nil
}

// For backward compatibility
func Close(logger *slog.Logger) {
    if defaultClient != nil {
        if err := defaultClient.Close(); err != nil {
            logger.Error("Error closing Redis connection", "error", err)
        }
    }
}