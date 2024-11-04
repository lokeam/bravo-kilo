package authcache

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lokeam/bravo-kilo/internal/shared/redis"
	goredis "github.com/redis/go-redis/v9"
)

type AuthCacheService interface {
    SetAuthToken(ctx context.Context, userID string, token string) error
    GetAuthToken(ctx context.Context, userID string) (string, bool, error)
    InvalidateAuthToken(ctx context.Context, userID string) error
}

type AuthCacheServiceImpl struct {
    redisClient *redis.RedisClient
    logger      *slog.Logger
    config      *redis.RedisConfig
}

func NewAuthCacheService(
    redisClient *redis.RedisClient,
    logger *slog.Logger,
) AuthCacheService {
    if redisClient == nil {
        panic("redisClient is nil")
    }
    if logger == nil {
        panic("logger is nil")
    }

    return &AuthCacheServiceImpl{
        redisClient: redisClient,
        logger:      logger.With("component", "auth_cache_service"),
        config:      redisClient.GetConfig(),
    }
}

func (s *AuthCacheServiceImpl) SetAuthToken(ctx context.Context, userID string, token string) error {
    key := s.buildKey("auth", userID)
    return s.redisClient.Set(ctx, key, token, s.config.CacheConfig.AuthTokenExpiration)
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