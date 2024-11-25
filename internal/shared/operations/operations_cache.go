package operations

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/lokeam/bravo-kilo/internal/shared/redis"
	"github.com/lokeam/bravo-kilo/internal/shared/types"
)


type CacheOperation struct {
	executor     types.OperationExecutor[*types.LibraryPageData]
	validator    types.Validator
	client       *redis.RedisClient
	metrics      *redis.Metrics
	logger       *slog.Logger
	config       *redis.RedisConfig
}

func NewCacheOperation(
	client *redis.RedisClient,
	timeout time.Duration,
	logger *slog.Logger,
) *CacheOperation {
	return &CacheOperation{
		executor: NewOperationExecutor[*types.LibraryPageData](
			"cache",
			timeout,
			logger,
		),
		client: client,
		logger: logger,
		config: client.GetConfig(),
	}
}

// Get executor to wrap cache retrieval
func (co *CacheOperation) Get(
	ctx context.Context,
	userID int,
	params *types.LibraryQueryParams,
	) (*types.LibraryPageData, error) {
		// Use executor to wrap cache operation
		return co.executor.Execute(ctx, func(ctx context.Context) (*types.LibraryPageData, error) {
			cacheKey := fmt.Sprintf("library:%d", userID)

			// Get string data from Redis
			stringData, err := co.client.Get(ctx, cacheKey)
			if err != nil {
				return nil, fmt.Errorf("redis cache get operation failed: %w", err)
			}

			// Create and unmarshal data
			pageData := types.NewLibraryPageData(co.logger)
			if err := pageData.UnmarshalBinary([]byte(stringData)); err != nil {
				co.metrics.IncrementCacheMisses()
				return nil, fmt.Errorf("unmarshal failed: %w", err)
			}

			// Validate data
			if err := co.validator.ValidateStruct(ctx, pageData); err != nil {
				return nil, fmt.Errorf("validation failed: %w", err)
			}

			// Increment metrics
			co.metrics.IncrementCacheHits()

			return pageData, nil
		})
}

// Set executor to wrap cache update
func (co *CacheOperation) Set(
	ctx context.Context,
	userID int,
	params *types.LibraryQueryParams,
	data *types.LibraryPageData,
) error {
	// Use executor to wrap cache operation
	_, err := co.executor.Execute(ctx, func(ctx context.Context) (*types.LibraryPageData, error) {
		// Validate data before caching
		if err := co.validator.ValidateStruct(ctx, data); err != nil {
			return nil, fmt.Errorf("validation failed: %w", err)
		}

		// Generate cache key
		cacheKey := fmt.Sprintf("library:%d", userID)

		// Marshal data for storage
		marshaledData, err := data.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("marshal failed: %w", err)
		}

		// Set data in Redis
		ttl := co.client.GetConfig().CacheConfig.DefaultTTL
		if ttl == 0 {
			ttl = time.Hour
		}

		// Set data in Redis with configured TTL
		if err := co.client.Set(ctx, cacheKey, string(marshaledData), ttl); err != nil {
			co.metrics.IncrementCacheMisses()
			return nil, fmt.Errorf("redis cache set operation failed: %w", err)
		}

		// Increment metrics
		co.metrics.IncrementCacheHits()

		// Return original data for executor type
		return data, nil
	})

	return err
}