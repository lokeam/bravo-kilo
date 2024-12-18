package redis

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type ClientStatus int // Current state of Redis client

const (
	StatusInitializing ClientStatus = iota
	StatusReady
	StatusRecovering
	StatusError
	StatusClosed
)

// Track operational stats for Redis client
type ClientStats struct {
	Connections    int64
	Operations     int64
	Errors         int64
	LastOperation  time.Time
	Uptime         time.Duration
	StartTime      time.Time
}

// Wrap Redis client w/ additional functionality
type RedisClient struct {
	client           *redis.Client
	config           *RedisConfig
	logger           *slog.Logger
	health           *HealthChecker
	metrics          *Metrics
	retrier          *Retrier

	// Internal state
	mu         sync.RWMutex
	isReady    bool
	status     ClientStatus
	lastError  error


	// Channels for control
	shutdown chan struct{}
	done     chan struct{}

	// Statistics
	stats      *ClientStats

	// Circuit breaker
	breaker *CircuitBreaker
}

// Functional options
type ClientOption func(*RedisClient)

func (s ClientStatus) String() string {
	switch s {
	case StatusInitializing:
			return "INITIALIZING"
	case StatusReady:
			return "READY"
	case StatusError:
			return "ERROR"
	case StatusRecovering:
			return "RECOVERING"
	case StatusClosed:
			return "CLOSED"
	default:
			return "UNKNOWN"
	}
}

// Add Metrics collection
func WithMetrics(metrics *Metrics) ClientOption {
	return func(c *RedisClient) {
		c.metrics = metrics
	}
}
// Adds Retry logic
func WithRetrier(retrier *Retrier) ClientOption {
	return func (c *RedisClient) {
		c.retrier = retrier
	}
}

// Main client constructor
func NewRedisClient(cfg *RedisConfig, logger *slog.Logger, opts ...ClientOption) (*RedisClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("redis configuration cannot be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	logger.Info("Initializing Redis client",
		"host", cfg.Host,
		"port", cfg.Port,
		"circuitBreakerEnabled", cfg.CircuitBreaker.Enabled)

	// Only create circuit breaker if enabled
	var breaker *CircuitBreaker
	if cfg.CircuitBreaker.Enabled {
		logger.Info("Configuring circuit breaker",
			"maxFailures", cfg.CircuitBreaker.MaxFailures,
			"resetTimeout", cfg.CircuitBreaker.ResetTimeout,
			"halfOpenRequests", cfg.CircuitBreaker.HalfOpenRequests)

		cbConfig := &CircuitBreakerConfig{
			MaxFailures:      cfg.CircuitBreaker.MaxFailures,
			ResetTimeout:     cfg.CircuitBreaker.ResetTimeout,
			HalfOpenRequests: cfg.CircuitBreaker.HalfOpenRequests,
		}

		var err error
		breaker, err = NewCircuitBreaker(cbConfig)
		if err != nil {
			logger.Error("Circuit breaker initialization failed",
					"error", err,
					"config", cbConfig)
			return nil, fmt.Errorf("failed to create circuit breaker: %w", err)
		}

		logger.Info("Circuit breaker initialized successfully",
			"state", breaker.GetState())
	} else {
		logger.Info("Circuit breaker disabled by configuration")
	}

	client := &RedisClient{
		breaker: breaker,
		config: cfg,
		logger: logger,
		status: StatusInitializing,
		shutdown: make(chan struct{}),
		done: make(chan struct{}),
		stats: &ClientStats{
			StartTime: time.Now(),
		},
		metrics: NewMetrics(),
	}

	if client.metrics == nil {
		client.metrics = NewMetrics()
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}

	// Create Redis client
	redisOpts := &redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  cfg.TimeoutConfig.Dial,
		ReadTimeout:  cfg.TimeoutConfig.Read,
		WriteTimeout: cfg.TimeoutConfig.Write,
		PoolSize:     cfg.PoolConfig.MaxActiveConns,
		MinIdleConns: cfg.PoolConfig.MinIdleConns,
		MaxIdleConns: cfg.PoolConfig.MaxIdleConns,
		ConnMaxIdleTime:  cfg.PoolConfig.IdleTimeout,
		ConnMaxLifetime:   cfg.PoolConfig.MaxConnLifetime,
		PoolTimeout:  cfg.PoolConfig.PoolTimeout,
	}

	client.client = redis.NewClient(redisOpts)

	// TEST THE CONNECTION IMMEDIATELY
	ctx, cancel := context.WithTimeout(context.Background(), cfg.TimeoutConfig.Dial)
	defer cancel()

	if err := client.client.Ping(ctx).Err(); err != nil {
			logger.Error("Redis connection test failed",
					"error", err,
					"host", cfg.Host,
					"port", cfg.Port)

			// Update status before returning error
			client.status = StatusError
			client.lastError = err

			return nil, fmt.Errorf("redis connection test failed: %w", err)
	}

	logger.Info("Redis connection test successful",
		"host", cfg.Host,
		"port", cfg.Port)

	// Initialize health checker
	health, err := NewHealthChecker(client, cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create health checker: %w", err)
	}
	client.health = health

	// Set status to ready only after successful connection test
	client.status = StatusReady

	logger.Info("Redis client fully initialized",
			"status", client.status,
			"circuitBreakerEnabled", breaker != nil,
			"healthCheckerEnabled", health != nil)

	return client, nil
}

// Redis client getter for factory use
func (c *RedisClient) GetClient() *redis.Client {
	return c.client
}

// Redis config getter for handler use
func (c *RedisClient) GetConfig() *RedisConfig {
	return c.config
}

// Establish connection to Redis
func (c *RedisClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Set initial status before attempting connection
	c.updateStatus(StatusInitializing)

	c.logger.Info("Attempting Redis connection",
		"host", c.config.Host,
		"port", c.config.Port,
		"currentStatus", c.status)

	if c.isReady {
		return nil
	}

	if err := c.client.Ping(ctx).Err(); err != nil {
		c.status = StatusError
		c.lastError = err
		slog.Error("Redis connection failed",
				"error", err,
				"host", c.config.Host,
				"port", c.config.Port)
		return fmt.Errorf("failed to connect to Redis: %w", err)
}

	c.status = StatusReady
	c.isReady = true
	c.updateStatus(StatusReady)

	// Start health check
	if err := c.health.Start(ctx); err != nil {
		return fmt.Errorf("failed to start health check routine: %w", err)
	}

	// Start metrics collection
	if c.metrics != nil {
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					c.updatePoolMetrics()
				case <-c.shutdown:
					return
				}
			}
		}()
	}


	c.logger.Info("Successfully connected to Redis", "host", c.config.Host, "port", c.config.Port)

	return nil
}

// Gracefully close Redis connection
func (c *RedisClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.client.Close(); err != nil {
		c.lastError = err
		c.updateStatus(StatusError)
		return err
}

	// Signal shutdown
	close(c.shutdown)

	// Reset metrics
	if c.metrics != nil {
		c.metrics.Reset()
	}

	// Stop health check
	if err := c.health.Stop(); err != nil {
		c.logger.Error("Error stopping health checker", "error", err)
	}

	// Close Redis client
	if err := c.client.Close(); err != nil {
		return fmt.Errorf("failed to close Redis connection: %w", err)
	}

	c.status = StatusClosed
	close(c.done)

	c.logger.Info("Redis client closed successfully")
	c.isReady = false
	c.updateStatus(StatusClosed)
	return nil
}


// CRUD operations
// Retrieve value from Redis
func (c *RedisClient) Get(ctx context.Context, key string) (string, error) {
	start := time.Now()
	operationID := uuid.New().String()[:8]

	// 1. Initial logging (unchanged)
	c.logger.Info("Redis operation initiated",
			"id", operationID,
			"operation", "GET",
			"key", key,
			"step", "start")

	// 2. Check client state (unchanged)
	status := c.GetStatus()
	c.logger.Info("Redis client state check",
			"id", operationID,
			"status", status,
			"isReady", c.IsReady(),
			"step", "state_check")

	// 3. Connection check (unchanged)
	if err := c.client.Ping(ctx).Err(); err != nil {
			c.logger.Error("Redis ping failed",
					"id", operationID,
					"error", err,
					"step", "ping_check",
					"duration", time.Since(start))
			return "", fmt.Errorf("redis connection check failed: %w", err)
	}

	// 4. Direct Redis GET (simplified)
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
			if err == redis.Nil {
					c.logger.Info("Cache miss",
							"id", operationID,
							"key", key,
							"step", "cache_miss")
					return "", nil
			}

			c.logger.Error("Redis GET failed",
					"id", operationID,
					"error", err,
					"duration", time.Since(start),
					"step", "get_error")
			return "", fmt.Errorf("redis GET failed: %w", err)
	}

	c.logger.Info("Redis GET successful",
			"id", operationID,
			"duration", time.Since(start),
			"step", "get_success")
	return val, nil
}

// Store value in Redis
func (c *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	start := time.Now()

	// Read lock for initial status check
	c.mu.RLock()
	initialStatus := c.status
	retrier := c.retrier // Get retrier reference under read lock
	c.mu.RUnlock()

	c.logger.Info("Redis operation starting",
			"operation", "SET",
			"key", key,
			"clientStatus", initialStatus,
			"circuitState", c.getCircuitBreakerStateString(),
	)

	// Operation to retry
	operation := func() error {
			// Add context timeout for each attempt
			opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			err := c.client.Set(opCtx, key, value, expiration).Err()
			if err != nil {
					c.logger.Warn("Redis set attempt failed",
							"operation", "SET",
							"key", key,
							"error", err,
					)
					return err // Return error to trigger retry
			}
			return nil
	}

	// Execute with retry logic if retrier exists
	var err error
	if retrier != nil {
			err = retrier.AttemptRetry(ctx, operation)
	} else {
			err = operation()
	}

	// Write lock for stats update
	c.mu.Lock()
	c.updateStats(err)
	c.mu.Unlock()

	if err != nil {
			// Write lock for error handling
			c.mu.Lock()
			c.handleOperationError(err)
			c.mu.Unlock()

			c.logger.Error("Redis set operation failed after retries",
					"operation", "SET",
					"key", key,
					"error", err,
					"duration", time.Since(start),
			)
			return fmt.Errorf("redis set failed after retries: %w", err)
	}

	// Write lock for success handling
	c.mu.Lock()
	c.handleOperationSuccess()
	c.mu.Unlock()

	c.logger.Info("Redis set operation completed",
			"operation", "SET",
			"key", key,
			"duration", time.Since(start),
			"status", "success",
	)

	return nil
}

// Deletes key from Redis
func (c *RedisClient) Delete(ctx context.Context, keys ...string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.executeWithRetry(ctx, "delete", func() error {
		return c.client.Del(ctx, keys...).Err()
	})
}

// Return client status
func (c *RedisClient) GetClientStatus() ClientStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status
}

// Return current client statistics
func (c *RedisClient) GetStats() ClientStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := *c.stats
	stats.Uptime = time.Since(stats.StartTime)

	// Add connection stats from metrics
	if c.metrics != nil {
			c.metrics.mu.RLock()
			stats.Connections = c.metrics.ActiveConnections
			stats.Operations = c.metrics.GetTotalOperations()

			// Sum all errors from metrics
			for _, count := range c.metrics.ErrorCount {
					stats.Errors += count
			}
			c.metrics.mu.RUnlock()
	}

	return stats
}

// Get connection pool statistics
func (c *RedisClient) PoolStats() *redis.PoolStats {
	return c.client.PoolStats()
}

func (c *RedisClient) Ping(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.executeWithRetry(ctx, "ping", func() error {
			return c.client.Ping(ctx).Err()
	})
}

// User deletion queue operations
func (c *RedisClient) AddToDeletionQueue(ctx context.Context, key string, value interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.executeWithRetry(ctx, "add_to_deletion_queue", func() error{
		return c.client.RPush(ctx, key, value).Err()
	})
}

func (c *RedisClient) GetDeletionQueueItems(ctx context.Context, key string, start, stop int64) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var values []string
	err := c.executeWithRetry(ctx, "get_deletion_queue_items", func() error {
		var err error
		values, err = c.client.LRange(ctx, key, start, stop).Result()
		return err
	})

	return values, err
}

func (c *RedisClient) RemoveFromDeletionQueue(ctx context.Context, key string, count int64, value interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.executeWithRetry(ctx, "remove_from_deletion_queue", func() error {
		return c.client.LRem(ctx, key, count, value).Err()
	})
}

// Helper fns
// Wrap operations with retry logic, circuit breaker and metrics
func (c *RedisClient) executeWithRetry(ctx context.Context, op string, fn func() error) error {
	if !c.isReady {
		slog.Error("Redis client not ready for operation",
				"operation", op,
				"status", c.status,
				"isReady", c.isReady)
		return fmt.Errorf("%s operation failed: redis client is not ready", op)
}

	start := time.Now()
	var err error

	operation := func() error {
			return c.withCircuitBreaker(ctx, op, fn)
	}

	// Execute with retry if configured
	if c.retrier != nil {
			err = c.retrier.AttemptRetry(ctx, operation)
	} else {
			err = operation()
	}

	duration := time.Since(start)

    // Handle operation result
    if err != nil {
			c.handleOperationError(err)
		} else {
				c.handleOperationSuccess()
		}

	// Record metrics
	if c.metrics != nil {
			c.metrics.RecordOperationDuration(op, duration, err)
	}

	// Update health metrics
	if c.health != nil {
			c.health.latencyWindow.Add(float64(duration))
			if err != nil {
					c.health.errorWindow.Add(1.0)
			}
	}

	// Update client stats
	c.updateStats(err)


	c.logger.Info("Redis operation completed",
	"operation", op,
	"duration", duration,
	"error", err,
	"circuitBreakerState", c.getCircuitBreakerStateString())

	return err
}

// Handles operation execution safety with circuit breaker
func (c *RedisClient) withCircuitBreaker(ctx context.Context, op string, fn func() error) error {
	// First check if breaker allows operation
	err := c.safeCircuitBreakerOp("AllowOperation", func() error {
			if c.breaker != nil {
					return c.breaker.AllowWithContext(ctx)
			}
			return nil
	})
	if err != nil {
			c.logger.Error("Circuit breaker rejected request",
					"operation", op,
					"error", err)
			return fmt.Errorf("%s operation failed: circuit breaker rejected request: %w", op, err)
	}

	// Execute the actual operation
	opErr := fn()

	// Record the result
	if opErr != nil {
			_ = c.safeCircuitBreakerOp("RecordFailure", func() error {
					if c.breaker != nil {
							c.breaker.RecordFailure()
					}
					return nil
			})
	} else {
			_ = c.safeCircuitBreakerOp("RecordSuccess", func() error {
					if c.breaker != nil {
							c.breaker.RecordSuccess()
					}
					return nil
			})
	}

	return opErr
}

// Get safe state access for logging
func (c *RedisClient) getCircuitBreakerStateString() string {
	if c.breaker == nil {
			return "DISABLED"
	}
	return c.breaker.GetState().String()
}

func (c *RedisClient) updatePoolMetrics() {
	if c.metrics == nil {
		return
	}

	stats := c.client.PoolStats()
	c.metrics.mu.Lock()
	defer c.metrics.mu.Unlock()

	// Update all pool metrics at once
	c.metrics.ActiveConnections = int64(stats.TotalConns)
	c.metrics.IdleConnections = int64(stats.IdleConns)
	c.metrics.PoolHits = int64(stats.Hits)
	c.metrics.PoolMisses = int64(stats.Misses)
	c.metrics.PoolTimeout = int64(stats.Timeouts)
}

func (c *RedisClient) updateStats(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.stats.Operations++
	c.stats.LastOperation = time.Now()

	if err != nil && err != redis.Nil { // Don't count cache misses as errors
			c.stats.Errors++
			slog.Error("Redis operation error recorded",
					"totalErrors", c.stats.Errors,
					"error", err,
			)
	}
}

// Check if Redis client is ready
func (c *RedisClient) IsReady() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isReady && c.status == StatusReady
}

func (c *RedisClient) GetStatus() ClientStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status
}

func (c *RedisClient) updateStatus(newStatus ClientStatus) {
	prevStatus := c.status

	// Build log fields dynamically to handle nil circuit breaker
	logFields := []any{
		"prevStatus", prevStatus,
		"newStatus", newStatus,
	}

	// Add error info if exists
	if c.lastError != nil {
		logFields = append(logFields, "lastError", c.lastError)
	}

	// Only add circuit breaker state if it exists
	if c.breaker != nil {
		logFields = append(logFields, "circuitState", c.breaker.GetState())
	}

	// Add metrics if they exist
	if c.stats != nil {
		logFields = append(logFields,
				"totalOperations", c.stats.Operations,
				"totalErrors", c.stats.Errors,
				"uptime", time.Since(c.stats.StartTime),
		)
	}

	// Add connection status change logging
	slog.Info("Redis connection status changed",
		"prevStatus", prevStatus,
		"newStatus", newStatus,
		"circuitState", c.getCircuitBreakerStateString(),
		"lastError", c.lastError,
		"totalOperations", c.stats.Operations,
		"totalErrors", c.stats.Errors,
		"uptime", time.Since(c.stats.StartTime))


	c.status = newStatus
}

func (c *RedisClient) GetCircuitBreaker() *CircuitBreaker {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.breaker
}

// After successful recovery
func (c *RedisClient) handleOperationSuccess() {
	if c.status != StatusReady {
		c.updateStatus(StatusReady)
	}

	_ = c.safeCircuitBreakerOp("RecordSuccess", func() error {
			if c.breaker != nil {
					c.breaker.RecordSuccess()
			}
			return nil
	})
}

func (c *RedisClient) logCircuitBreakerState(operation string) {
	if c.breaker == nil {
			return
	}

	c.logger.Info("Circuit breaker status",
	"operation", operation,
	"state", c.breaker.GetState(),
	"failures", c.breaker.failures,           // Changed from Failures()
	"lastStateChange", c.breaker.lastStateChange,)  // Changed from GetLastStateChange()
}

// Wrapper function to safely execute circuit breaker operations
func (c *RedisClient) safeCircuitBreakerOp(operation string, fn func() error) (err error) {
	// If no circuit breaker, return immediately
	if c.breaker == nil {
			return nil
	}

	// Defer runs when the function exits
	defer func() {
			if r := recover(); r != nil {
					// Convert panic to error
					err = fmt.Errorf("circuit breaker panic in %s: %v", operation, r)

					// Log the panic
					c.logger.Error("Circuit breaker operation panicked",
							"operation", operation,
							"panic", r,
							"stack", string(debug.Stack()))

					// Update client status
					c.updateStatus(StatusError)
			}
	}()

	// Execute the wrapped function
	return fn()
}

// Update your existing methods to use this wrapper
func (c *RedisClient) handleOperationError(err error) {
	if err != nil && err != redis.Nil {
			c.lastError = err
			c.updateStatus(StatusError)

			_ = c.safeCircuitBreakerOp("RecordFailure", func() error {
					if c.breaker != nil {
							c.breaker.RecordFailure()
					}
					return nil
			})
	}
}
