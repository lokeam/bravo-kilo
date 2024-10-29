package redis

import (
	"context"
	"log/slog"
	"os"

	goredis "github.com/redis/go-redis/v9"
)

// Client wraps the redis client to avoid type conflicts
type Client = goredis.Client  // Use type alias instead of struct

// Global client instance
var redisClient *Client

func InitRedis(ctx context.Context,logger *slog.Logger) (*Client, error) {
	opt, err := goredis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
			logger.Error("Failed to parse Redis URL", "error", err)
			return nil, err
	}

	redisClient = goredis.NewClient(opt)

	// Ping Redis to check connection
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
			logger.Error("Failed to connect to Redis", "error", err)
			return nil, err
	}

	logger.Info("Successfully connected to Redis")
	return redisClient, nil
}

func Close(logger *slog.Logger) {
	if redisClient != nil {
			err := redisClient.Close()
			if err != nil {
					logger.Info("Error closing Redis client", "error", err)
			}
	}
}
