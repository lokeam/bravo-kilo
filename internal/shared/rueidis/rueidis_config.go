package rueidis

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/redis/rueidis"
)

type Config struct {
	// Connection settings (required, loaded from environment)
	Host     string // Redis server hostname
	Port     int    // Redis server port
	Password string // Redis password (optional)
	DB       int    // Redis database number

	// Timeout
	ConnWriteTimeout time.Duration   // How long to wait on writes
	ConnReadTimeout  time.Duration   // How long to wait on reads
	BlockingPoolSize int             // Size of the connection pool for blocking operations
	TimeoutConfig    TimeoutConfig

	// Migration specific
	EnableMetrics    bool
	EnableTracing    bool

	// Application-specific cache settings
	CacheConfig CacheConfig
}

type CacheConfig struct {
	// User-related caches
	UserData          time.Duration

	// AI-related caches
	GeminiResponse    time.Duration

	// Default TTL
	DefaultTTL        time.Duration
}

type TimeoutConfig struct {
	Read    time.Duration
	Write   time.Duration
	Default time.Duration
}


func NewConfig() *Config {
	return &Config{
		ConnWriteTimeout: 5 * time.Second,
		ConnReadTimeout:  5 * time.Second,
		BlockingPoolSize: 10,

		TimeoutConfig: TimeoutConfig{
			Read:    2 * time.Second,
			Write:   2 * time.Second,
			Default: 3 * time.Second,
		},

		EnableMetrics:    true,
		EnableTracing:    true,

		CacheConfig: CacheConfig{
			UserData:          30 * time.Minute,
			GeminiResponse:    15 * time.Minute,
			DefaultTTL:        15 * time.Minute,
	},
	}
}

func (c *Config) LoadFromEnv() error {
	// Required: Host
	host := os.Getenv("REDIS_HOST")
	if host == "" {
			return fmt.Errorf("REDIS_HOST environment variable is required")
	}
	c.Host = host

	// Required: Port
	portStr := os.Getenv("REDIS_PORT")
	if portStr == "" {
			return fmt.Errorf("REDIS_PORT environment variable is required")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
			return fmt.Errorf("invalid REDIS_PORT value: %w", err)
	}
	if port < 1 || port > 65535 {
			return fmt.Errorf("REDIS_PORT must be between 1 and 65535")
	}
	c.Port = port

	// Optional: Password
	if pass := os.Getenv("REDIS_PASSWORD"); pass != "" {
			c.Password = pass
	}

	// Optional: Database number
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
			db, err := strconv.Atoi(dbStr)
			if err != nil {
					return fmt.Errorf("invalid REDIS_DB value: %w", err)
			}
			if db < 0 {
					return fmt.Errorf("REDIS_DB must be non-negative")
			}
			c.DB = db
	}

	// Load timeout configurations
	if readTimeout := os.Getenv("REDIS_TIMEOUT_READ"); readTimeout != "" {
		duration, err := time.ParseDuration(readTimeout)
		if err != nil {
				return fmt.Errorf("invalid REDIS_TIMEOUT_READ value: %w", err)
		}
		c.TimeoutConfig.Read = duration
	}

	if writeTimeout := os.Getenv("REDIS_TIMEOUT_WRITE"); writeTimeout != "" {
			duration, err := time.ParseDuration(writeTimeout)
			if err != nil {
					return fmt.Errorf("invalid REDIS_TIMEOUT_WRITE value: %w", err)
			}
			c.TimeoutConfig.Write = duration
	}

	if defaultTimeout := os.Getenv("REDIS_TIMEOUT_DEFAULT"); defaultTimeout != "" {
			duration, err := time.ParseDuration(defaultTimeout)
			if err != nil {
					return fmt.Errorf("invalid REDIS_TIMEOUT_DEFAULT value: %w", err)
			}
			c.TimeoutConfig.Default = duration
	}

	// Load Gemini cache duration
	if ttl := os.Getenv("REDIS_GEMINI_CACHE_TTL"); ttl != "" {
		duration, err := time.ParseDuration(ttl)
		if err != nil {
				return fmt.Errorf("invalid REDIS_GEMINI_CACHE_TTL value: %w", err)
		}
		c.CacheConfig.GeminiResponse = duration
	}

	return nil
}

func (c *Config) Validate() error {
	if c.Host == "" {
			return fmt.Errorf("redis host cannot be empty")
	}
	if c.Port < 1 || c.Port > 65535 {
			return fmt.Errorf("invalid Redis port: %d", c.Port)
	}

	// Validate timeout values
	if c.TimeoutConfig.Read <= 0 {
		return fmt.Errorf("TimeoutConfig.Read must be positive")
	}
	if c.TimeoutConfig.Write <= 0 {
			return fmt.Errorf("TimeoutConfig.Write must be positive")
	}
	if c.TimeoutConfig.Default <= 0 {
			return fmt.Errorf("TimeoutConfig.Default must be positive")
	}


	if c.ConnWriteTimeout <= 0 {
			return fmt.Errorf("ConnWriteTimeout must be positive")
	}
	if c.ConnReadTimeout <= 0 {
			return fmt.Errorf("ConnReadTimeout must be positive")
	}
	if c.BlockingPoolSize < 1 {
			return fmt.Errorf("BlockingPoolSize must be at least 1")
	}
	return nil
}

func (c *Config) ToRueidisOptions() rueidis.ClientOption {
	return rueidis.ClientOption{
			InitAddress:      []string{fmt.Sprintf("%s:%d", c.Host, c.Port)},
			Password:         c.Password,
			SelectDB:        c.DB,
			ConnWriteTimeout: c.ConnWriteTimeout,  // Direct field in ClientOption
			Dialer: net.Dialer{
				Timeout: c.ConnReadTimeout,
		},
			BlockingPoolSize:  c.BlockingPoolSize,
	}
}
