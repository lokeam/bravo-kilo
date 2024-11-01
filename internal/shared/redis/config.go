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
}

// Configuration validation and defaults
func NewRedisConfig() *RedisConfig {
	config := &RedisConfig{
		Host: "localhost",
		Port: 6379,
		DB: 0,
	}

	// Pool defaults
	config.PoolConfig.MinIdleConns = 3                   // Keeps a small set of connections ready, reducing latency for sudden traffic spikes
	config.PoolConfig.MaxIdleConns = 10                  // Balances resource usage with performance (10% of MaxActiveConns is a common ratio)
	config.PoolConfig.MaxActiveConns = 100               // Maximum number of connections total (active + idle)
	config.PoolConfig.IdleTimeout = 5 * time.Minute      // Average cleanup interval for unused connections
	config.PoolConfig.MaxConnLifetime = 1 * time.Hour    // Prevents connection staleness while not being too aggressive
	config.PoolConfig.WaitTimeout = 3 * time.Second      // Aligns w/ typical web request timeout expectations
	config.PoolConfig.PoolTimeout = 4 * time.Second      // Slightly higher than WaitTimeout to allow for connection creation

	// Timeout defaults
	config.TimeoutConfig.Dial = 5 * time.Second          // Allow for network latency + DNS resolution
	config.TimeoutConfig.Read = 2 * time.Second          // Provide buffer for network issues while failing fast enough to avoid pile up
	config.TimeoutConfig.Write = 2 * time.Second         // Provide buffer for network issues while failing fast enough to avoid pile up

	// Health check defaults
	config.HealthConfig.Enabled = true                    // Toggle health checks
	config.HealthConfig.Interval = 30 * time.Second       // Frequent enough to detect issues but not too often to impact performance
	config.HealthConfig.Timeout = 2 * time.Second         // Match operation timeout
	config.HealthConfig.MaxRetries = 3                    // Standard 3 retry pattern
	config.HealthConfig.RetryInterval = 1 * time.Second   // Allow temporary network issues to resolve

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

	return nil
}
