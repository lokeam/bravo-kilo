package operations

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/lokeam/bravo-kilo/internal/shared/redis"
	"github.com/lokeam/bravo-kilo/internal/shared/types"
	"github.com/lokeam/bravo-kilo/internal/shared/validator"
)


type CacheOperation struct {
	*OperationExecutor[*types.LibraryPageData]
	client       *redis.RedisClient
	validator    *validator.BaseValidator
	metrics      *redis.Metrics
	logger       *slog.Logger
}

func NewCacheOperation(
	client *redis.RedisClient,
	timeout time.Duration,
	logger *slog.Logger,
) *CacheOperation {
	return &CacheOperation{
		OperationExecutor: NewOperationExecutor[*types.LibraryPageData](
			"cache",
			timeout,
			logger,
		),
		client: client,
	}
}

// Get uses executor to wrap cache retrieval
func (co *CacheOperation) Get(
	ctx context.Context,
	userID int,
	params *types.LibraryQueryParams,
	) (*types.LibraryPageData, error) {
		// Use executor to wrap cache operation
		return co.Execute(ctx, func(ctx context.Context) (*types.LibraryPageData, error) {
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