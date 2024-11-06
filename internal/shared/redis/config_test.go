package redis

import (
	"os"
	"testing"
	"time"
)

func TestRedisConfig(t *testing.T) {
	config := NewRedisConfig()

	if config.Host != "localhost" {
		t.Errorf("expected default host to be localhost, got %s", config.Host)
	}
	if config.Port != 6379 {
		t.Errorf("expected default port to be 6379, got %d", config.Port)
	}
}

func TestLoadFromEnv(t *testing.T) {
	tests := []struct {
			name        string
			envVars    map[string]string
			wantErr    bool
			checkFunc  func(*RedisConfig) bool
	}{
			{
					name: "valid URL",
					envVars: map[string]string{
							"REDIS_URL": "redis://localhost:6379",
					},
					wantErr: false,
					checkFunc: func(c *RedisConfig) bool {
							return c.URL == "redis://localhost:6379"
					},
			},
			{
					name: "valid host and port",
					envVars: map[string]string{
							"REDIS_HOST": "redis.example.com",
							"REDIS_PORT": "6380",
					},
					wantErr: false,
					checkFunc: func(c *RedisConfig) bool {
							return c.Host == "redis.example.com" && c.Port == 6380
					},
			},
			{
					name: "invalid port",
					envVars: map[string]string{
							"REDIS_PORT": "invalid",
					},
					wantErr: true,
					checkFunc: func(c *RedisConfig) bool {
							return true
					},
			},
	}

	for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
					// Clear environment
					os.Clearenv()

					// Set environment variables
					for k, v := range tt.envVars {
							os.Setenv(k, v)
					}

					config := NewRedisConfig()
					err := config.LoadFromEnv()

					if (err != nil) != tt.wantErr {
							t.Errorf("LoadFromEnv() error = %v, wantErr %v", err, tt.wantErr)
							return
					}

					if !tt.wantErr && !tt.checkFunc(config) {
							t.Errorf("LoadFromEnv() failed validation check")
					}
			})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
			name    string
			config  func() *RedisConfig
			wantErr bool
	}{
			{
					name: "valid config",
					config: func() *RedisConfig {
							return NewRedisConfig()
					},
					wantErr: false,
			},
			{
					name: "empty host without URL",
					config: func() *RedisConfig {
							c := NewRedisConfig()
							c.Host = ""
							return c
					},
					wantErr: true,
			},
			{
					name: "invalid port",
					config: func() *RedisConfig {
							c := NewRedisConfig()
							c.Port = 70000
							return c
					},
					wantErr: true,
			},
			{
					name: "invalid pool config",
					config: func() *RedisConfig {
							c := NewRedisConfig()
							c.PoolConfig.MaxActiveConns = 5
							c.PoolConfig.MaxIdleConns = 10
							return c
					},
					wantErr: true,
			},
			{
					name: "invalid timeout",
					config: func() *RedisConfig {
							c := NewRedisConfig()
							c.TimeoutConfig.Read = -1 * time.Second
							return c
					},
					wantErr: true,
			},
	}

	for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
					config := tt.config()
					err := config.Validate()
					if (err != nil) != tt.wantErr {
							t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
					}
			})
	}
}
