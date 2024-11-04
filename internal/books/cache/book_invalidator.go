package cache

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/lokeam/bravo-kilo/internal/books/repository"
	"github.com/lokeam/bravo-kilo/internal/shared/cache"
	"github.com/lokeam/bravo-kilo/internal/shared/redis"
)

const (
    redisPrefix = "book:" // Base prefix for all book-related Redis keys
)

type BookCacheInvalidator interface {
	InvalidateL1Cache(itemID, userID int) error
	InvalidateL2Cache(ctx context.Context, keys []string) error
	GetType() string
	GetCacheKeys(itemID, userID int) []string
}

type BookCacheInvalidatorImpl struct {
    bookCache          repository.BookCache     // L1 cache
    redisInvalidator   *cache.RedisInvalidator  // L2 cache
    logger             *slog.Logger
    redisPrefix        string
}

func NewBookCacheInvalidator(
    bookCache repository.BookCache,
    redisClient *redis.RedisClient,
    logger *slog.Logger,
) *BookCacheInvalidatorImpl {
    if bookCache == nil {
        panic("bookCache cannot be nil")

    }
    if redisClient == nil {
        panic("redisClient cannot be nil")
    }
    if logger == nil {
        panic("logger cannot be nil")
    }

    return &BookCacheInvalidatorImpl{
        bookCache:          bookCache,
        redisInvalidator:   cache.NewRedisInvalidtor(redisClient, logger),
        logger:             logger,
        redisPrefix:        redisPrefix,
    }
}

// L1CacheInvalidator interface
func (b *BookCacheInvalidatorImpl) InvalidateL1Cache(itemID, userID int) error {
    start := time.Now()
    defer func() {
        b.logger.Debug("L1 cache invalidation completed",
            "duration", time.Since(start),
            "userID", userID,
            "bookID", itemID,
        )
    }()

    b.bookCache.InvalidateCaches(itemID, userID)
    return nil
}

func (b *BookCacheInvalidatorImpl) GetType() string {
    return "book"
}

// L2CacheInvalidator interface, actual deletion is handled by the CacheManager
func (b *BookCacheInvalidatorImpl) InvalidateL2Cache(ctx context.Context, keys []string) error {
    start := time.Now()
    defer func() {
        b.logger.Debug("L2 cache invalidation completed",
            "duration", time.Since(start),
            "keyCount", len(keys),
        )
    }()

    if err := b.validateKeys(keys); err != nil {
        return fmt.Errorf("invalid cache keys: %w", err)
    }

    return b.redisInvalidator.InvalidateL2Cache(ctx, keys)
}

func (b *BookCacheInvalidatorImpl) GetCacheKeys(itemID, userID int) []string {
    return []string{
        // User-specific book data
        fmt.Sprintf("%s%s%d", b.redisPrefix, redis.PrefixBook, userID),
        fmt.Sprintf("%s%s%d", b.redisPrefix, redis.PrefixBookAuthor, userID),
        fmt.Sprintf("%s%s%d", b.redisPrefix, redis.PrefixBookFormat, userID),
        fmt.Sprintf("%s%s%d", b.redisPrefix, redis.PrefixBookGenre, userID),
        fmt.Sprintf("%s%s%d", b.redisPrefix, redis.PrefixBookTag, userID),
        fmt.Sprintf("%s%s%d", b.redisPrefix, redis.PrefixBookHomepage, userID),

        // Book-specific data
        fmt.Sprintf("%s%s%d:%d", b.redisPrefix, redis.PrefixBookDetail, userID, itemID),
        fmt.Sprintf("%s%s%d", b.redisPrefix, redis.PrefixBookMetadata, itemID),

        // List/collection keys that might contain this book
        fmt.Sprintf("%s%s%d:recent", b.redisPrefix, redis.PrefixBookList, userID),
        fmt.Sprintf("%s%s%d:favorites", b.redisPrefix, redis.PrefixBookList, userID),
    }
}

// Helper method to validate cache keys
func (b *BookCacheInvalidatorImpl) validateKeys(keys []string) error {
    if len(keys) == 0 {
        return fmt.Errorf("no cache keys provided")
    }

    for _, key := range keys {
        if key == "" {
            return fmt.Errorf("empty cache key found")
        }
        // Reasonable max length for Redis keys
        if len(key) > 200 {
            return fmt.Errorf("key exceeds maximum length: %s", key)
        }
    }

    return nil
}