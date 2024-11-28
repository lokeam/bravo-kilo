package operations

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/lokeam/bravo-kilo/internal/shared/redis"
	"github.com/lokeam/bravo-kilo/internal/shared/rueidis"
	"github.com/lokeam/bravo-kilo/internal/shared/types"
)


type CacheOperation struct {
	executor     types.OperationExecutor[*types.LibraryPageData]
	validator    types.Validator
	client       *rueidis.Client
	metrics      *redis.Metrics
	logger       *slog.Logger
	config       *rueidis.Config
}

func NewCacheOperation(
	client *rueidis.Client,
	timeout time.Duration,
	logger *slog.Logger,
	validator types.Validator,
) *CacheOperation {
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

	return &CacheOperation{
			executor: NewOperationExecutor[*types.LibraryPageData](
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
func (co *CacheOperation) Get(
	ctx context.Context,
	userID int,
	params *types.LibraryQueryParams,
) (*types.LibraryPageData, error) {
	// Safety check for nil dependencies
	if co.client == nil {
			return nil, fmt.Errorf("redis client not initialized")
	}

	return co.executor.Execute(ctx, func(ctx context.Context) (*types.LibraryPageData, error) {
			cacheKey := fmt.Sprintf("library:%d", userID)

			// Get string data from Redis
			stringData, err := co.client.Get(ctx, cacheKey)
			if err != nil {
					// Safely increment cache miss metric
					if co.metrics != nil {
							co.metrics.IncrementCacheMisses()
					}

					// Let rueidis client handle error mapping
					mappedErr := co.client.MapError(err, "GET")
					switch {
					case errors.Is(mappedErr, rueidis.ErrNotFound):  // Rueidis specific nil check
							if co.logger != nil {
									co.logger.Debug("cache miss",
											"operation", "GET",
											"key", cacheKey)
							}
							return nil, nil

					case err == context.DeadlineExceeded:
						return nil, redis.NewOperationError("GET", cacheKey, redis.ErrTimeout)

					case errors.Is(err, context.DeadlineExceeded):
							return nil, redis.NewOperationError("GET", cacheKey, redis.ErrTimeout)
					case errors.Is(mappedErr, rueidis.ErrConnectionFailed):
							return nil, redis.NewOperationError("GET", cacheKey, err)
					case errors.Is(mappedErr, redis.ErrClientNotReady):
							return nil, redis.NewOperationError("GET", cacheKey, err)
					default:
							if co.logger != nil {
									co.logger.Error("unexpected redis error",
											"operation", "GET",
											"key", cacheKey,
											"error", mappedErr)
							}
							return nil, rueidis.NewOperationError("GET", cacheKey, mappedErr)
					}
			}

			// Don't try to unmarshal empty data
			if stringData == "" {
				if co.metrics != nil {
						co.metrics.IncrementCacheMisses()
				}
				if co.logger != nil {
						co.logger.Debug("empty cache data",
								"operation", "GET",
								"key", cacheKey)
				}
				return nil, nil
		}


			// Create and unmarshal data
			pageData := types.NewLibraryPageData(co.logger)
			if err := pageData.UnmarshalBinary([]byte(stringData)); err != nil {
					if co.metrics != nil {
							co.metrics.IncrementCacheMisses()
					}
					return nil, redis.NewOperationError("GET", cacheKey, redis.ErrInvalidData)
			}

			// Validate data if validator is available
			if co.validator != nil {
					if validationErrs := co.validator.ValidateStruct(ctx, pageData); len(validationErrs) > 0 {
							if co.metrics != nil {
									co.metrics.IncrementCacheMisses()
							}
							primaryErr := validationErrs[0]

							if co.logger != nil {
									co.logger.Error("validation failed",
											"operation", "GET",
											"key", cacheKey,
											"errors", validationErrs)
							}
							return nil, redis.NewOperationError("GET", cacheKey, primaryErr)
					}
			}

			// Success path
			if co.metrics != nil {
					co.metrics.IncrementCacheHits()
			}
			if co.logger != nil {
					co.logger.Debug("cache hit",
							"operation", "GET",
							"key", cacheKey)
			}

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
	// Safety check for nil dependencies
	if co.client == nil {
			return fmt.Errorf("redis client not initialized")
	}
	if data == nil {
			return fmt.Errorf("cannot cache nil data")
	}

	_, err := co.executor.Execute(ctx, func(ctx context.Context) (*types.LibraryPageData, error) {
			cacheKey := fmt.Sprintf("library:%d", userID)

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
							return nil, redis.NewOperationError("SET", cacheKey, primaryErr)
					}
			}

			// Marshal data for storage
			marshaledData, err := data.MarshalBinary()
			if err != nil {
					if co.metrics != nil {
							co.metrics.IncrementCacheMisses()
					}
					return nil, redis.NewOperationError("SET", cacheKey, fmt.Errorf("marshal failed: %w", err))
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
							return nil, redis.NewOperationError("SET", cacheKey, err)
					case redis.ErrClientNotReady:
							return nil, redis.NewOperationError("SET", cacheKey, err)
					default:
							if co.logger != nil {
									co.logger.Error("unexpected redis error during SET",
											"operation", "SET",
											"key", cacheKey,
											"error", err)
							}
							return nil, redis.NewOperationError("SET", cacheKey, err)
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