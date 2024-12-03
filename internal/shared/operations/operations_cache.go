package operations

import (
	"context"
	"encoding"
	"errors"
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
	co.logger.Debug("CACHE_OP: Starting cache retrieval",
		"component", "cache_operation",
		"function", "GetTyped",
		"userID", userID,
		"hasParams", params != nil,
		"hasClient", co.client != nil,
	)

	if co.client == nil {
		co.logger.Error("CACHE_OP: Redis client not initialized",
			"component", "cache_operation",
			"function", "GetTyped",
		)
		return zero, fmt.Errorf("redis client not initialized")
	}
	if params == nil {
		co.logger.Error("CACHE_OP: Params are nil",
			"component", "cache_operation",
			"function", "GetTyped",
		)
		return zero, fmt.Errorf("params cannot be nil")
	}

  return co.executor.Execute(ctx, func(ctx context.Context) (T, error) {
      cacheKey := fmt.Sprintf("%s:%d", params.Domain, userID)
			co.logger.Debug("CACHE_OP: Attempting cache fetch",
				"component", "cache_operation",
				"function", "GetTyped.Execute",
				"cacheKey", cacheKey,
			)

			stringData, err := co.client.Get(ctx, cacheKey)
			if err != nil {
				if co.metrics != nil {
					co.metrics.IncrementCacheMisses()
				}

				mappedErr := co.client.MapError(err, "GET")
				switch {
				case errors.Is(mappedErr, rueidis.ErrNotFound):  // Rueidis specific nil check
						if co.logger != nil {
								co.logger.Debug("CACHE_OP: Cache miss",
										"component", "cache_operation",
										"function", "GetTyped.Execute",
										"operation", "GET",
										"key", cacheKey)
						}
						return zero, nil

				case err == context.DeadlineExceeded:
						return zero, redis.NewOperationError("GET", cacheKey, redis.ErrTimeout)

				case errors.Is(err, context.DeadlineExceeded):
						return zero, redis.NewOperationError("GET", cacheKey, redis.ErrTimeout)
				case errors.Is(mappedErr, rueidis.ErrConnectionFailed):
						return zero, redis.NewOperationError("GET", cacheKey, err)
				case errors.Is(mappedErr, redis.ErrClientNotReady):
						return zero, redis.NewOperationError("GET", cacheKey, err)
				default:
						if co.logger != nil {
								co.logger.Error("CACHE_OP: Unexpected redis error",
										"component", "cache_operation",
										"function", "GetTyped.Execute",
										"operation", "GET",
										"key", cacheKey,
										"error", mappedErr)
						}
						return zero, rueidis.NewOperationError("GET", cacheKey, mappedErr)
				}
			}

			if stringData == "" {
				co.logger.Debug("CACHE_OP: Cache miss - empty data",
					"component", "cache_operation",
					"function", "GetTyped.Execute",
					"cacheKey", cacheKey,
				)
				if co.metrics != nil {
						co.metrics.IncrementCacheMisses()
				}
				return zero, nil
			}

			// Create appropriate type based on T
			var pageData T
			co.logger.Debug("CACHE_OP: Creating page data instance",
				"component", "cache_operation",
				"function", "GetTyped.Execute",
				"dataType", fmt.Sprintf("%T", pageData),
			)

			switch any(pageData).(type) {
			case *types.LibraryPageData:
					pageData = any(types.NewLibraryPageData(co.logger)).(T)
			case *types.HomePageData:
					pageData = any(types.NewHomePageData(co.logger)).(T)
			default:
				co.logger.Error("CACHE_OP: Unsupported page type",
					"component", "cache_operation",
					"function", "GetTyped.Execute",
					"type", fmt.Sprintf("%T", pageData),
				)
				return zero, fmt.Errorf("unsupported page type")
			}

			// Type assert to access UnmarshalBinary
			if unmarshaler, ok := any(pageData).(encoding.BinaryUnmarshaler); ok {
					if err := unmarshaler.UnmarshalBinary([]byte(stringData)); err != nil {
						co.logger.Error("CACHE_OP: Unmarshal failed",
							"component", "cache_operation",
							"function", "GetTyped.Execute",
							"error", err,
							"cacheKey", cacheKey,
						)

						if co.metrics != nil {
								co.metrics.IncrementCacheMisses()
						}
						return zero, redis.NewOperationError("GET", cacheKey, redis.ErrInvalidData)
					}
			}
			co.logger.Debug("CACHE_OP: Cache hit successful",
				"component", "cache_operation",
				"function", "GetTyped.Execute",
				"cacheKey", cacheKey,
				"dataType", fmt.Sprintf("%T", pageData),
			)

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
	co.logger.Debug("CACHE_OP: Starting cache set",
		"component", "cache_operation",
		"function", "SetTyped",
		"userID", userID,
		"hasParams", params != nil,
		"hasData", any(data) != nil,
	)


	// Safety check for nil dependencies
	if co.client == nil {
		co.logger.Error("CACHE_OP: Redis client not initialized",
			"component", "cache_operation",
			"function", "SetTyped",
		)
		return fmt.Errorf("redis client not initialized")
	}
	if params == nil {
		co.logger.Error("CACHE_OP: Params are nil",
			"component", "cache_operation",
			"function", "SetTyped",
		)

		return fmt.Errorf("params cannot be nil")
	}
    // Check for nil using type assertion
  if any(data) == nil {
		co.logger.Error("CACHE_OP: Data is nil",
			"component", "cache_operation",
			"function", "SetTyped",
		)
		return fmt.Errorf("cannot cache nil data")
	}

	_, err := co.executor.Execute(ctx, func(ctx context.Context) (T, error) {
			cacheKey := fmt.Sprintf("library:%s:%d", params.Domain,userID)
			co.logger.Debug("CACHE_OP: Validating data",
				"component", "cache_operation",
				"function", "SetTyped.Execute",
				"cacheKey", cacheKey,
			)

			// Validate data before caching if validator exists
			if co.validator != nil {
					if validationErrs := co.validator.ValidateStruct(ctx, data); len(validationErrs) > 0 {
							if co.metrics != nil {
									co.metrics.IncrementCacheMisses()
							}
							primaryErr := validationErrs[0]
							co.logger.Error("CACHE_OP: Validation failed",
								"component", "cache_operation",
								"function", "SetTyped.Execute",
								"errors", validationErrs,
								"cacheKey", cacheKey,
							)

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
				co.logger.Error("CACHE_OP: Marshal failed",
					"component", "cache_operation",
					"function", "SetTyped.Execute",
					"error", err,
					"cacheKey", cacheKey,
				)
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

			co.logger.Debug("CACHE_OP: Setting cache data",
        "component", "cache_operation",
        "function", "SetTyped.Execute",
        "cacheKey", cacheKey,
        "ttl", ttl,
      )

			// Set data in Redis
			if err := co.client.Set(ctx, cacheKey, marshaledData, ttl); err != nil {
					co.logger.Error("CACHE_OP: Cache set failed",
						"component", "cache_operation",
						"function", "SetTyped.Execute",
						"error", err,
						"cacheKey", cacheKey,
					)
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
				co.logger.Debug("CACHE_OP: Cache set successful",
					"component", "cache_operation",
					"function", "SetTyped.Execute",
					"cacheKey", cacheKey,
				)
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