package operations

import (
	"context"
	"encoding"
	"fmt"
	"log/slog"
	"time"

	"github.com/lokeam/bravo-kilo/internal/shared/binary"
	"github.com/lokeam/bravo-kilo/internal/shared/redis"
	"github.com/lokeam/bravo-kilo/internal/shared/rueidis"
	"github.com/lokeam/bravo-kilo/internal/shared/types"
)

type CacheOperation[T types.PageData] struct {
	executor     *OperationExecutor[T]
	validator    types.Validator
	client       *rueidis.Client
	metrics      *redis.Metrics
	logger       *slog.Logger
	config       *rueidis.Config
}

func NewCacheOperation[T types.PageData](
	client *rueidis.Client,
	timeout time.Duration,
	logger *slog.Logger,
	validator types.Validator,
) *CacheOperation[T] {
	metrics := redis.NewMetrics()

	// Get config safely
	var config *rueidis.Config
	if client != nil {
			config = client.GetConfig()
	}

	// Add initialization logging
	if logger != nil {
			logger.Debug("initializing cache operation",
					"clientNull", client == nil,
					"metricsNull", metrics == nil,
					"validatorNull", validator == nil,
					"configNull", config == nil)
	}

	return &CacheOperation[T]{
		executor: NewOperationExecutor[T](
				"cache",
				timeout,
				logger,
		),
		client: client,
		logger: logger,
		config: config,
		validator: validator,
		metrics: metrics,
}
}

// Get executor to wrap cache retrieval
func (co *CacheOperation[T]) GetTyped(
	ctx context.Context,
	userID int,
	params *types.PageQueryParams,
) (T, error) {
	var zero T
	if co.client == nil {
			return zero, fmt.Errorf("redis client not initialized")
	}
	if params == nil {
			return zero, fmt.Errorf("params cannot be nil")
	}

    return co.executor.Execute(ctx, func(ctx context.Context) (T, error) {
      cacheKey := fmt.Sprintf("%s:%d", params.Domain, userID)

			stringData, err := co.client.Get(ctx, cacheKey)
			if err != nil {
					if co.metrics != nil {
							co.metrics.IncrementCacheMisses()
					}

					return zero, err
			}

			if stringData == "" {
					if co.metrics != nil {
							co.metrics.IncrementCacheMisses()
					}
					return zero, nil
			}

			// Create appropriate type based on T
			var pageData T
			switch any(pageData).(type) {
			case *types.LibraryPageData:
					pageData = any(types.NewLibraryPageData(co.logger)).(T)
			case *types.HomePageData:
					pageData = any(types.NewHomePageData(co.logger)).(T)
			default:
					return zero, fmt.Errorf("unsupported page type")
			}

			// Type assert to access UnmarshalBinary
			if unmarshaler, ok := any(pageData).(encoding.BinaryUnmarshaler); ok {
					if err := unmarshaler.UnmarshalBinary([]byte(stringData)); err != nil {
							if co.metrics != nil {
									co.metrics.IncrementCacheMisses()
							}
							return zero, redis.NewOperationError("GET", cacheKey, redis.ErrInvalidData)
					}
			}

			// Success path
			if co.metrics != nil {
					co.metrics.IncrementCacheHits()
			}

			return pageData, nil
	})
}

// Set executor to wrap cache update
func (co *CacheOperation[T]) SetTyped(
	ctx context.Context,
	userID int,
	params *types.PageQueryParams,
	data T,
) error {
	var zero T  // Add zero value at the start

	// Safety check for nil dependencies
	if co.client == nil {
			return fmt.Errorf("redis client not initialized")
	}
	if params == nil {
		return fmt.Errorf("params cannot be nil")
	}
    // Check for nil using type assertion
  if any(data) == nil {
		return fmt.Errorf("cannot cache nil data")
	}

	_, err := co.executor.Execute(ctx, func(ctx context.Context) (T, error) {
			cacheKey := fmt.Sprintf("library:%s:%d", params.Domain,userID)

			// Validate data before caching if validator exists
			if co.validator != nil {
					if validationErrs := co.validator.ValidateStruct(ctx, data); len(validationErrs) > 0 {
							if co.metrics != nil {
									co.metrics.IncrementCacheMisses()
							}
							primaryErr := validationErrs[0]

							if co.logger != nil {
									co.logger.Error("validation failed during SET",
											"operation", "SET",
											"key", cacheKey,
											"errors", validationErrs)
							}
							return zero, redis.NewOperationError("SET", cacheKey, primaryErr)
					}
			}

			// Marshal data for storage
			marshaledData, err := binary.MarshalBinary(data)
			if err != nil {
					if co.metrics != nil {
							co.metrics.IncrementCacheMisses()
					}
					return zero, redis.NewOperationError("SET", cacheKey, fmt.Errorf("marshal failed: %w", err))
			}

			// Get TTL from config or use default
			ttl := time.Hour // Default TTL
			if co.config != nil && co.config.CacheConfig.DefaultTTL > 0 {
					ttl = co.config.CacheConfig.DefaultTTL
			}

			// Set data in Redis
			if err := co.client.Set(ctx, cacheKey, marshaledData, ttl); err != nil {
					if co.metrics != nil {
							co.metrics.IncrementCacheMisses()
					}

					switch err {
					case redis.ErrConnectionFailed:
							return zero, redis.NewOperationError("SET", cacheKey, err)
					case redis.ErrClientNotReady:
							return zero, redis.NewOperationError("SET", cacheKey, err)
					default:
							if co.logger != nil {
									co.logger.Error("unexpected redis error during SET",
											"operation", "SET",
											"key", cacheKey,
											"error", err)
							}
							return zero, redis.NewOperationError("SET", cacheKey, err)
					}
			}

			// Success path
			if co.metrics != nil {
					co.metrics.IncrementCacheHits()
			}
			if co.logger != nil {
					co.logger.Debug("cache set successful",
							"operation", "SET",
							"key", cacheKey)
			}

			return data, nil
	})

	return err
}

// Wrapper methods to satisfy CacheOperator interface by converting between specific types and any

// Get satisfies the CacheOperator interface
func (co *CacheOperation[T]) Get(ctx context.Context, userID int, params *types.PageQueryParams) (interface{}, error) {
	result, err := co.GetTyped(ctx, userID, params)
	return result, err
}

// Set satisfies the CacheOperator interface
func (co *CacheOperation[T]) Set(ctx context.Context, userID int, params *types.PageQueryParams, data interface{}) error {
	typedData, ok := data.(T)
	if !ok {
		return fmt.Errorf("invalid data type for cache operation")
	}
	return co.SetTyped(ctx, userID, params, typedData)
}