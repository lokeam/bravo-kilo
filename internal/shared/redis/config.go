package redis

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type RedisConfig struct {
	// Metadata
	Name         string  // Instance name for logging
	Environment  string  // dev/stg/prd

	// Cache durations for different types of data
	CacheConfig struct {
		BookList                       time.Duration      // 30min
		BookDetail                     time.Duration
		BooksByAuthor                  time.Duration
		BooksByFormat                  time.Duration
		BooksByGenre                   time.Duration
		BooksByTag                     time.Duration
		BookHomepage                   time.Duration
		UserData                       time.Duration
		DefaultTTL                     time.Duration
		OperationTimeout               time.Duration
		DefaultBookCacheExpiration     time.Duration
		AuthTokenExpiration            time.Duration
		UserDeletionMarkerExpiration   time.Duration
		GeminiResponse                 time.Duration
	}

	// Connection
	Host       string
	Port       int
	Password   string
	DB         int
	URL        string

	// Pool Settings
	PoolConfig struct {
		MinIdleConns      int
		MaxIdleConns      int
		MaxActiveConns    int
		IdleTimeout       time.Duration
		MaxConnLifetime   time.Duration
		WaitTimeout       time.Duration
		PoolTimeout       time.Duration
	}

	// Retry Config
	RetryConfig struct {
		MaxRetries        int
		BackoffInitial    time.Duration
		BackoffMax        time.Duration
		BackoffFactor     float64
	}

	// Timeouts
	TimeoutConfig struct {
		Dial   time.Duration
		Read   time.Duration
		Write  time.Duration
	}

	// Health Check
	HealthConfig struct {
		Enabled        bool
		Interval       time.Duration
		Timeout        time.Duration
		MaxRetries     int
		RetryInterval  time.Duration
	}

	// Circuit Breaker config for Redis
	CircuitBreaker struct {
		Enabled          bool          `env:"REDIS_CIRCUIT_BREAKER_ENABLED" envDefault:"true"`
		MaxFailures      int           `env:"REDIS_CIRCUIT_BREAKER_MAX_FAILURES" envDefault:"15"`
		ResetTimeout     time.Duration `env:"REDIS_CIRCUIT_BREAKER_RESET_TIMEOUT" envDefault:"30s"`
		HalfOpenRequests int           `env:"REDIS_CIRCUIT_BREAKER_HALF_OPEN_REQUESTS" envDefault:"5"`
	}
}

// Configuration validation and defaults
func NewRedisConfig() *RedisConfig {
	config := &RedisConfig{
		Host: "localhost",
		Port: 6379,
		DB: 0,
	}

	// Pool defaults
	config.PoolConfig.MinIdleConns = 25                   // Keeps a small set of connections ready, reducing latency for sudden traffic spikes
	config.PoolConfig.MaxIdleConns = 50                  // Balances resource usage with performance (10% of MaxActiveConns is a common ratio)
	config.PoolConfig.MaxActiveConns = 500               // Maximum number of connections total (active + idle)
	config.PoolConfig.IdleTimeout = 5 * time.Minute      // Average cleanup interval for unused connections
	config.PoolConfig.MaxConnLifetime = 1 * time.Hour    // Prevents connection staleness while not being too aggressive
	config.PoolConfig.WaitTimeout = 5 * time.Second      // Aligns w/ typical web request timeout expectations
	config.PoolConfig.PoolTimeout = 5 * time.Second      // Slightly higher than WaitTimeout to allow for connection creation

	// Timeout defaults
	config.TimeoutConfig.Dial = 10 * time.Second          // Allow for network latency + DNS resolution
	config.TimeoutConfig.Read = 3 * time.Second          // Provide buffer for network issues while failing fast enough to avoid pile up
	config.TimeoutConfig.Write = 3 * time.Second         // Provide buffer for network issues while failing fast enough to avoid pile up

	// Health check defaults
	config.HealthConfig.Enabled = true                    // Toggle health checks
	config.HealthConfig.Interval = 15 * time.Second       // Frequent enough to detect issues but not too often to impact performance
	config.HealthConfig.Timeout = 3 * time.Second         // Match operation timeout
	config.HealthConfig.MaxRetries = 5                    // Standard 3 retry pattern
	config.HealthConfig.RetryInterval = 2 * time.Second   // Allow temporary network issues to resolve

	// Cache duration defaults
	config.CacheConfig.BookList = 1 * time.Hour                         // List of books changes infrequently
	config.CacheConfig.BookDetail = 1 * time.Hour                         // Individual book details are very stable
	config.CacheConfig.BooksByAuthor = 3 * time.Hour                      // Author collections change rarely
	config.CacheConfig.BooksByFormat = 3 * time.Hour                      // Format groupings are very stable
	config.CacheConfig.BooksByGenre = 2 * time.Hour                       // Genre groupings are very stable
	config.CacheConfig.BooksByTag = 2 * time.Hour                         // Tags might change more frequently
	config.CacheConfig.BookHomepage = 30 * time.Minute                    // Homepage needs fresher data
	config.CacheConfig.UserData = 30 * time.Minute                        // User preferences/settings
	config.CacheConfig.DefaultBookCacheExpiration = 1 * time.Hour
	config.CacheConfig.AuthTokenExpiration = 24 * time.Hour
	config.CacheConfig.UserDeletionMarkerExpiration = 48 * time.Hour
	config.CacheConfig.DefaultTTL = 48 * time.Minute                      // Conservative default
  config.CacheConfig.GeminiResponse = 15 * time.Minute

	// Circuit Breaker defaults
	config.CircuitBreaker.Enabled = true
	config.CircuitBreaker.MaxFailures = 25
	config.CircuitBreaker.ResetTimeout = 45 * time.Second
	config.CircuitBreaker.HalfOpenRequests = 10

	return config
}

// Load configuration from environment variables
func (c *RedisConfig) LoadFromEnv() error {
	// Connection settings
	if url := os.Getenv("REDIS_URL"); url != "" {
			c.URL = url
	} else {
			if host := os.Getenv("REDIS_HOST"); host != "" {
					c.Host = host
			}
			if port := os.Getenv("REDIS_PORT"); port != "" {
					portInt, err := strconv.Atoi(port)
					if err != nil {
							return fmt.Errorf("invalid REDIS_PORT: %w", err)
					}
					c.Port = portInt
			}
	}

	if pass := os.Getenv("REDIS_PASSWORD"); pass != "" {
			c.Password = pass
	}

	if db := os.Getenv("REDIS_DB"); db != "" {
			dbInt, err := strconv.Atoi(db)
			if err != nil {
					return fmt.Errorf("invalid REDIS_DB: %w", err)
			}
			c.DB = dbInt
	}

	// Cache durations
	if ttl := os.Getenv("REDIS_DEFAULT_BOOK_CACHE_TTL"); ttl != "" {
		duration, err := time.ParseDuration(ttl)
		if err != nil {
				return fmt.Errorf("invalid REDIS_DEFAULT_BOOK_CACHE_TTL: %w", err)
		}
		c.CacheConfig.DefaultBookCacheExpiration = duration
	}

	if ttl := os.Getenv("REDIS_AUTH_TOKEN_TTL"); ttl != "" {
			duration, err := time.ParseDuration(ttl)
			if err != nil {
					return fmt.Errorf("invalid REDIS_AUTH_TOKEN_TTL: %w", err)
			}
			c.CacheConfig.AuthTokenExpiration = duration
	}

	if ttl := os.Getenv("REDIS_USER_DELETION_MARKER_TTL"); ttl != "" {
			duration, err := time.ParseDuration(ttl)
			if err != nil {
					return fmt.Errorf("invalid REDIS_USER_DELETION_MARKER_TTL: %w", err)
			}
			c.CacheConfig.UserDeletionMarkerExpiration = duration
	}

	// Pool settings
	if maxConns := os.Getenv("REDIS_MAX_ACTIVE_CONNS"); maxConns != "" {
			val, err := strconv.Atoi(maxConns)
			if err != nil {
					return fmt.Errorf("invalid REDIS_MAX_ACTIVE_CONNS: %w", err)
			}
			c.PoolConfig.MaxActiveConns = val
	}

	if minConns := os.Getenv("REDIS_MIN_IDLE_CONNS"); minConns != "" {
		val, err := strconv.Atoi(minConns)
		if err != nil {
				return fmt.Errorf("invalid REDIS_MIN_IDLE_CONNS: %w", err)
		}
		c.PoolConfig.MaxActiveConns = val
	}

	// Cache durations
	if ttl := os.Getenv("REDIS_CACHE_BOOKS_TTL"); ttl != "" {
		duration, err := time.ParseDuration(ttl)
		if err != nil {
			return fmt.Errorf("invalid REDIS_CACHE_BOOKS_TTL: %w", err)
		}
		c.CacheConfig.DefaultTTL = duration
	} else {
		c.CacheConfig.DefaultTTL = 30 * time.Minute
	}

	if ttl := os.Getenv("REDIS_GEMINI_CACHE_TTL"); ttl != "" {
		duration, err := time.ParseDuration(ttl)
		if err != nil {
				return fmt.Errorf("invalid REDIS_GEMINI_CACHE_TTL: %w", err)
		}
		c.CacheConfig.GeminiResponse = duration
	}

	// Circuit breaker settings
	if enabled := os.Getenv("REDIS_CIRCUIT_BREAKER_ENABLED"); enabled != "" {
		val, err := strconv.ParseBool(enabled)
		if err != nil {
			return fmt.Errorf("invalid REDIS_CIRCUIT_BREAKER_ENABLED: %w", err)
		}
		c.CircuitBreaker.Enabled = val
	}

	if maxFailures := os.Getenv("REDIS_CIRCUIT_MAX_FAILURES"); maxFailures != "" {
		val, err := strconv.Atoi(maxFailures)
		if err != nil {
				return fmt.Errorf("invalid REDIS_CIRCUIT_MAX_FAILURES: %w", err)
		}
		c.CircuitBreaker.MaxFailures = val
	}

	if resetTimeout := os.Getenv("REDIS_CIRCUIT_RESET_TIMEOUT"); resetTimeout != "" {
		duration, err := time.ParseDuration(resetTimeout)
		if err != nil {
				return fmt.Errorf("invalid REDIS_CIRCUIT_RESET_TIMEOUT: %w", err)
		}
		c.CircuitBreaker.ResetTimeout = duration
	}

	if halfOpen := os.Getenv("REDIS_CIRCUIT_HALF_OPEN_REQUESTS"); halfOpen != "" {
		val, err := strconv.Atoi(halfOpen)
		if err != nil {
				return fmt.Errorf("invalid REDIS_CIRCUIT_HALF_OPEN_REQUESTS: %w", err)
		}
		c.CircuitBreaker.HalfOpenRequests = val
	}


	return nil
}


func (c *RedisConfig) Validate() error {
	if c.URL == "" {
			if c.Host == "" {
					return fmt.Errorf("redis host cannot be empty")
			}
			if c.Port <= 0 || c.Port > 65535 {
					return fmt.Errorf("invalid redis port: %d", c.Port)
			}
	}

	// Validate pool configuration
	if c.PoolConfig.MaxActiveConns < c.PoolConfig.MaxIdleConns {
			return fmt.Errorf("maxActiveConns cannot be less than maxIdleConns")
	}
	if c.PoolConfig.MinIdleConns > c.PoolConfig.MaxIdleConns {
			return fmt.Errorf("minIdleConns cannot be greater than maxIdleConns")
	}
	if c.PoolConfig.IdleTimeout < 0 {
			return fmt.Errorf("idleTimeout cannot be negative")
	}
	if c.PoolConfig.MaxConnLifetime < 0 {
			return fmt.Errorf("maxConnLifetime cannot be negative")
	}

	// Validate timeout configuration
	if c.TimeoutConfig.Dial <= 0 {
			return fmt.Errorf("dial timeout must be positive")
	}
	if c.TimeoutConfig.Read <= 0 {
			return fmt.Errorf("read timeout must be positive")
	}
	if c.TimeoutConfig.Write <= 0 {
			return fmt.Errorf("write timeout must be positive")
	}

	// Validate health check configuration
	if c.HealthConfig.Enabled {
			if c.HealthConfig.Interval <= 0 {
					return fmt.Errorf("health check interval must be positive")
			}
			if c.HealthConfig.Timeout <= 0 {
					return fmt.Errorf("health check timeout must be positive")
			}
			if c.HealthConfig.MaxRetries < 0 {
					return fmt.Errorf("health check max retries cannot be negative")
			}
			if c.HealthConfig.RetryInterval <= 0 {
					return fmt.Errorf("health check retry interval must be positive")
			}
	}

	// Validate circuit breaker config
	if c.CircuitBreaker.MaxFailures <= 0 {
		return fmt.Errorf("circuit breaker max failures must be positive")
	}
	if c.CircuitBreaker.ResetTimeout <= 0 {
		return fmt.Errorf("circuit breaker reset timeout must be positive")
	}
	if c.CircuitBreaker.HalfOpenRequests <= 0 {
		return fmt.Errorf("circuit breaker half-open requests must be positive")
	}

	return nil
}
