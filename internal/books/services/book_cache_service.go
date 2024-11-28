package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/lokeam/bravo-kilo/internal/shared/redis"
	"github.com/lokeam/bravo-kilo/internal/shared/rueidis"
)

type BookCacheService interface {
    // High-level cache operations
    SetCachedBook(ctx context.Context, userID int, bookID int, value interface{}, duration time.Duration) error
    SetCachedBookList(ctx context.Context, userID int, operation string, value interface{}, duration time.Duration) error
	SetCachedGeminiResponse(ctx context.Context, prompt string, response string, duration time.Duration) error

    InvalidateCache(ctx context.Context, userID int, bookID int) error
	GetCachedGeminiResponse(ctx context.Context, prompt string) (string, bool, error)
    GetCachedBook(ctx context.Context, userID int, bookID int, result interface{}) (bool, error)
    GetCachedBookList(ctx context.Context, userID int, operation string, result interface{}) (bool, error)
    GetCacheKeys(itemID, userID int) []string
}

type BookCacheServiceImpl struct {
    redisClient *rueidis.Client
    logger      *slog.Logger
    metrics     *BookCacheMetrics
}

func NewBookCacheService(
    redisClient *rueidis.Client,
    logger *slog.Logger,
) BookCacheService {
    if redisClient == nil {
        panic("redisClient is nil")
    }
    if logger == nil {
        panic("logger is nil")
    }

    return &BookCacheServiceImpl{
        redisClient: redisClient,
        logger:      logger.With("component", "book_cache_service"),
        metrics:     NewBookCacheMetrics(),
    }
}

// High-level methods for book operations
func (s *BookCacheServiceImpl) GetCachedBook(ctx context.Context, userID int, bookID int, result interface{}) (bool, error) {
    key := s.buildKey("bookDetail", userID, bookID)
    return s.getCachedData(ctx, key, result)
}

func (s *BookCacheServiceImpl) GetCachedBookList(ctx context.Context, userID int, operation string, result interface{}) (bool, error) {
    key := s.buildKey(operation, userID)
    return s.getCachedData(ctx, key, result)
}

func (s *BookCacheServiceImpl) SetCachedBook(
    ctx context.Context,
    userID int,
    bookID int,
    value interface{},
    duration time.Duration,
    ) error {
    key := s.buildKey("bookDetail", userID, bookID)
    return s.setCachedData(ctx, key, value, duration)
}

func (s *BookCacheServiceImpl) SetCachedBookList(
    ctx context.Context,
    userID int,
    operation string,
    value interface{},
    duration time.Duration,
    ) error {
    key := s.buildKey(operation, userID)
    return s.setCachedData(ctx, key, value, duration)
}

func (s *BookCacheServiceImpl) InvalidateCache(ctx context.Context, userID int, bookID int) error {
    keys := s.GetCacheKeys(bookID, userID)
    return s.redisClient.Delete(ctx, keys...)
}

// Gemini cache methods
func (s *BookCacheServiceImpl) GetCachedGeminiResponse(ctx context.Context, prompt string) (string, bool, error) {
	key := s.buildKey("gemini", 0, prompt)
	var response string
	found, err := s.getCachedData(ctx, key, &response)
	if err != nil {
			return "", false, err
	}
	return response, found, nil
}

// L2CacheInvalidator interface implementation
func (s *BookCacheServiceImpl) GetCacheKeys(itemID, userID int) []string {
    return []string{
        fmt.Sprintf("%s%d", redis.PrefixBook, userID),
        fmt.Sprintf("%s%d", redis.PrefixBookFormat, userID),
        fmt.Sprintf("%s%d", redis.PrefixBookGenre, userID),
        fmt.Sprintf("%s%d", redis.PrefixBookTag, userID),
        fmt.Sprintf("%s%d", redis.PrefixBookHomepage, userID),
        fmt.Sprintf("%s:%d:%d", redis.PrefixBookDetail, userID, itemID),
    }
}

func (s *BookCacheServiceImpl) SetCachedGeminiResponse(
    ctx context.Context,
    prompt string,
    response string,
    duration time.Duration,
    ) error {
	key := s.buildKey("gemini", 0, prompt)
	return s.setCachedData(ctx, key, response, duration)
}

// Internal helper methods
func (s *BookCacheServiceImpl) getCachedData(
    ctx context.Context,
    key string,
    result any,
    ) (bool, error) {
    data, err := s.redisClient.Get(ctx, key)
    if err != nil {
        if errors.Is(err, rueidis.ErrNotFound) {
            s.metrics.RecordCacheMiss(key)
            return false, nil
        }
        s.logger.Error("cache fetch error", "key", key, "error", err)
        return false, fmt.Errorf("cache fetch error: %w", err)
    }

    if err := json.Unmarshal([]byte(data), result); err != nil {
        s.logger.Error("cache unmarshal error", "key", key, "error", err)
        s.metrics.RecordUnmarshalError(key)
        return false, nil
    }

    s.metrics.RecordCacheHit(key)
    return true, nil
}

func (s *BookCacheServiceImpl) setCachedData(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return fmt.Errorf("failed to marshal cache data: %w", err)
    }

    if err := s.redisClient.Set(ctx, key, data, expiration); err != nil {
        s.logger.Error("failed to cache data", "key", key, "error", err)
        return err
    }

    return nil
}

func (s *BookCacheServiceImpl) buildKey(operation string, userID int, params ...interface{}) string {
    switch operation {
    case "bookDetail":
        if len(params) > 0 {
            return fmt.Sprintf("%s:%d:%v", redis.PrefixBookDetail, userID, params[0])
        }
    case "booksByFormat":
        return fmt.Sprintf("%s%d", redis.PrefixBookFormat, userID)
    case "booksByGenre":
        return fmt.Sprintf("%s%d", redis.PrefixBookGenre, userID)
    case "booksByTag":
        return fmt.Sprintf("%s%d", redis.PrefixBookTag, userID)
    case "homepage":
        return fmt.Sprintf("%s%d", redis.PrefixBookHomepage, userID)
		case "gemini":
			if len(params) > 0 {
				return fmt.Sprintf("%s:%v", redis.PrefixGemini, params[0])
			}
    }
    return ""
}
