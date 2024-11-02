package redis

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type ClientStatus int // Current state of Redis client

const (
	StatusInitializing ClientStatus = iota
	StatusReady
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

	client := &RedisClient{
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

	// Initialize health checker
	health, err := NewHealthChecker(client, cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create health checker: %w", err)
	}
	client.health = health

	return client, nil
}

// Return underlying Redis client for factory use
func (c *RedisClient) GetClient() *redis.Client {
	return c.client
}

// Establish connection to Redis
func (c *RedisClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isReady {
		return nil
	}

	if err := c.client.Ping(ctx).Err(); err != nil {
		c.status = StatusError
		c.lastError = err
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	c.status = StatusReady
	c.isReady = true

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

	if c.status == StatusClosed {
		return nil
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
	return nil
}


// CRUD operations
// Retrieve value from Redis
func (c *RedisClient) Get(ctx context.Context, key string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var val string
	err := c.executeWithRetry(ctx, "get", func() error {
			var err error
			val, err = c.client.Get(ctx, key).Result()
			return err
	})

	return val, err
}

// Store value in Redis
func (c *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.executeWithRetry(ctx, "set", func() error {
		return c.client.Set(ctx, key, value, expiration).Err()
	})
}

// Deletes key from Redis
func (c *RedisClient) Delete(ctx context.Context, key string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.executeWithRetry(ctx, "delete", func() error {
		return c.client.Del(ctx, key).Err()
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

// Wrap operations with retry logic, circuit breaker and metrics
func (c *RedisClient) executeWithRetry(ctx context.Context, op string, fn func() error) error {
	if !c.isReady {
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

	return err
}

func (c *RedisClient) withCircuitBreaker(ctx context.Context, op string, fn func() error) error {
	if c.breaker != nil {
		if err := c.breaker.AllowWithContext(ctx); err != nil {
			return fmt.Errorf("%s operation failed: circuit breaker rejected request: %w", op, err)
		}
	}

	err := fn()

	if c.breaker != nil {
		if err != nil {
			c.breaker.RecordFailure()
		} else {
			c.breaker.RecordSuccess()
		}
	}

	return err
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
	c.mu.Lock()  // Add mutex protection
	defer c.mu.Unlock()

	c.stats.Operations++
	c.stats.LastOperation = time.Now()

	if err != nil {
		c.stats.Errors++
	}
}