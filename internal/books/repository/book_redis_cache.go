package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/lokeam/bravo-kilo/config"
	redisPrefix "github.com/lokeam/bravo-kilo/internal/shared/redis"
	"github.com/redis/go-redis/v9"
)


type BookRedisCache interface {
	GetBook(ctx context.Context, bookID string) ([]byte, error)
	SetBook(ctx context.Context, bookID string, value interface{}, expiration ...time.Duration) error
	DeleteBook(ctx context.Context, bookID string) error
	GetDeletionQueue(ctx context.Context) ([]string, error)
	SetAuthToken(ctx context.Context, userID string, token string, expiration time.Duration) error
	GetAuthToken(ctx context.Context, userID string) (string, error)
	SetUserDeletionMarker(ctx context.Context, userID string, expiration time.Duration) error
	GetUserDeletionMarker(ctx context.Context, userID string) (bool, error)
	RemoveFromDeletionQueue(ctx context.Context, userID string) error
}

type BookRedisCacheImpl struct {
	client *redis.Client
	logger *slog.Logger
}

func NewBookRedisCache(client *redis.Client, logger *slog.Logger) (BookRedisCache, error) {
	if client == nil {
		return nil, fmt.Errorf("redis client cannot be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger in redis client cannot be nil")
	}

	return &BookRedisCacheImpl{
		client: client,
		logger: logger,
	}, nil
}

// Book methods
func (brc *BookRedisCacheImpl) GetBook(ctx context.Context, bookID string) ([]byte, error) {
	key := fmt.Sprintf("book:%s", bookID)
	return brc.client.Get(ctx, key).Bytes()
}

func (brc *BookRedisCacheImpl) SetBook(ctx context.Context, bookID string, value interface{}, expiration ...time.Duration) error {
	key := fmt.Sprintf("book:%s", bookID)
	json, err := json.Marshal(value)
	if err != nil {
		return err
	}
	exp := config.AppConfig.DefaultBookCacheExpiration
	if len(expiration) > 0 {
		exp = expiration[0]
	}
	return brc.client.Set(ctx, key, json, exp).Err()
}

func (brc *BookRedisCacheImpl) DeleteBook(ctx context.Context, bookID string) error {
	key := fmt.Sprintf("%s%s", redisPrefix.PrefixBook, bookID)
	return brc.client.Del(ctx, key).Err()
}

// Auth token methods
func (brc *BookRedisCacheImpl) SetAuthToken(ctx context.Context, userID string, token string, expiration time.Duration) error {
	key := fmt.Sprintf("%s%s", redisPrefix.PrefixAuthToken, userID)
	return brc.client.Set(ctx, key, token, config.AppConfig.AuthTokenExpiration).Err()
}

func (brc *BookRedisCacheImpl) GetAuthToken(ctx context.Context, userID string) (string, error) {
	key := fmt.Sprintf("%s%s", redisPrefix.PrefixAuthToken, userID)
	return brc.client.Get(ctx, key).Result()
}

// User delete methods
func (brc *BookRedisCacheImpl) SetUserDeletionMarker(ctx context.Context, userID string, expiration time.Duration) error {
	key := fmt.Sprintf("%s%s", redisPrefix.PrefixUserDelete, userID)
	return brc.client.Set(ctx, key, time.Now().Unix(), config.AppConfig.UserDeletionMarkerExpiration).Err()
}

func (brc *BookRedisCacheImpl) GetUserDeletionMarker(ctx context.Context, userID string) (bool, error) {
	brc.logger.Info("Getting user deletion marker", "userID", userID)
	key := fmt.Sprintf("%s%s", redisPrefix.PrefixUserDelete, userID)
	_, err := brc.client.Get(ctx, key).Result()
	// Key not found
	if err == redis.Nil {
		brc.logger.Info("User deletion marker not found", "userID", userID)
		return false, nil
	}

	if err != nil {
		return false, err
	}
	// Key exists
	brc.logger.Info("User deletion marker found", "userID", userID)
	return true, nil
}

func (brc *BookRedisCacheImpl) AddToDeletionQueue(ctx context.Context, userID string) error {
	return brc.client.RPush(ctx, "deletion:queue", userID).Err()
}

func (brc *BookRedisCacheImpl) GetDeletionQueue(ctx context.Context) ([]string, error) {
	return brc.client.LRange(ctx, "deletion:queue", 0, -1).Result()
}

func (brc *BookRedisCacheImpl) RemoveFromDeletionQueue(ctx context.Context, userID string) error {
	return brc.client.LRem(ctx, "deletion:queue", 0, userID).Err()
}
