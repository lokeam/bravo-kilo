package redis

import (
	"context"
	"log/slog"
	"os"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func InitRedis(logger *slog.Logger) (*redis.Client, error) {
	opt, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		logger.Error("Failed to parse Redis URL", "error", err)
		return nil, err
	}

	Client = redis.NewClient(opt)

	// Ping Redis to check connection
	ctx := context.Background()
	_, err = Client.Ping(ctx).Result()
	if err != nil {
		logger.Error("Failed to connect to Redis", "error", err)
		return nil, err
	}

	logger.Info("Successfully connected to Redis")
	return Client, nil
}

func Close(logger *slog.Logger) {
	if Client != nil {
		err := Client.Close()
		if err != nil {
			logger.Info("Error closing Redis client", "error", err)
		}
	}
}
