package rueidis

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/redis/rueidis"
)

type Client struct {
	client  rueidis.Client
	logger  *slog.Logger
	config  *Config
	stats   *ClientStats
	status  atomic.Int32 // Using atomic for thread-safe status
}

type ClientStats struct {
	Operations     atomic.Int64
	Errors        atomic.Int64
	LastOperation atomic.Value // time.Time
	StartTime     time.Time
}

type Stats struct {
	Operations    int64
	Errors       int64
	LastOperation time.Time
	StartTime    time.Time
	Uptime       time.Duration
}

type ClientStatus int32

const (
	StatusReady ClientStatus = iota
	StatusError
	StatusClosed
)

func NewRedisClient(
	cfg *Config,
	logger *slog.Logger,
	) (*Client, error) {
	// Guard clauses
	if cfg == nil {
			return nil, fmt.Errorf("config cannot be nil")
	}
	if logger == nil {
			return nil, fmt.Errorf("logger cannot be nil")
	}

	// Convert config to rueidis options
	rueidisOpts := cfg.ToRueidisOptions()

	// Create client
	rueidisClient, err := rueidis.NewClient(rueidisOpts)
	if err != nil {
			return nil, fmt.Errorf("failed to create redis client: %w", err)
	}

	// Init stats
	stats := &ClientStats{
			StartTime: time.Now(),
	}
	stats.LastOperation.Store(time.Now())

	// Init client wrapper
	client := &Client{
			client:  rueidisClient,
			logger:  logger,
			config:  cfg,
			stats:   stats,
	}

	// Set initial status to ready
	client.status.Store(int32(StatusReady))

	return client, nil
}

// Core operations with built-in metrics and logging
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	start := time.Now()
	c.logger.Debug("attempting cache get",
		"key", key,
		"operation", "GET",
		"timestamp", start)


	defer func() {
			c.stats.Operations.Add(1)
			c.stats.LastOperation.Store(time.Now())
	}()

	if !c.IsReady() {
		return "", ErrClientNotReady
	}

	// Build and execute command
	cmd := c.client.B().Get().Key(key).Build()
	result, err := c.client.Do(ctx, cmd).ToString()

	if err != nil {
			c.stats.Errors.Add(1)
			c.logger.Error("redis get failed",
					"key", key,
					"error", err,
					"duration", time.Since(start))
			return "", err
	}

	c.logger.Debug("cache operation complete",
		"key", key,
		"operation", "GET",
		"duration", time.Since(start),
		"hit", result != "",
		"size", len(result))

	return result, nil
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	// Convert value to string based on type
	start := time.Now()
	var strValue string
	switch v := value.(type) {
	case string:
			strValue = v
	case []byte:
			strValue = string(v)
	default:
			// For other types, use JSON marshaling
			jsonBytes, err := json.Marshal(value)
			if err != nil {
					c.logger.Error("failed to marshal value",
							"key", key,
							"valueType", fmt.Sprintf("%T", value),
							"error", err)
					return fmt.Errorf("failed to marshal value: %w", err)
			}
			strValue = string(jsonBytes)
	}

	c.logger.Debug("attempting redis SET",
			"key", key,
			"valueType", fmt.Sprintf("%T", value),
			"valueSize", len(strValue),
			"ttl", expiration,
	)

	// Build and execute command with properly serialized value
	cmd := c.client.B().Set().Key(key).Value(strValue).Ex(expiration).Build()
	err := c.client.Do(ctx, cmd).Error()

	if err != nil {
			c.stats.Errors.Add(1)
			c.logger.Error("redis set failed",
					"key", key,
					"error", err,
					"duration", time.Since(start))
			return fmt.Errorf("redis set failed: %w", err)
	}

	c.logger.Debug("redis set successful",
			"key", key,
			"duration", time.Since(start))

	return nil
}

func (c *Client) Delete(ctx context.Context, keys ...string) error {
	start := time.Now()
	defer func() {
			c.stats.Operations.Add(1)
			c.stats.LastOperation.Store(time.Now())
	}()

	// Build and execute command
	cmd := c.client.B().Del().Key(keys...).Build()
	err := c.client.Do(ctx, cmd).Error()

	if err != nil {
			c.stats.Errors.Add(1)
			c.logger.Error("redis delete failed",
					"keys", keys,
					"error", err,
					"duration", time.Since(start))
			return fmt.Errorf("redis delete failed: %w", err)
	}

	c.logger.Debug("redis delete successful",
			"keys", keys,
			"duration", time.Since(start))

	return nil
}

func (c *Client) Close() error {
	c.status.Store(int32(StatusClosed))
	c.client.Close()
	c.logger.Info("Redis client closed")
	return nil
}

func (c *Client) Ping(ctx context.Context) error {
	start := time.Now()
	defer func() {
			c.stats.Operations.Add(1)
			c.stats.LastOperation.Store(time.Now())
	}()

	err := c.client.Do(ctx, c.client.B().Ping().Build()).Error()
	if err != nil {
			c.stats.Errors.Add(1)
			c.logger.Error("redis ping failed",
					"error", err,
					"duration", time.Since(start))
			return fmt.Errorf("redis ping failed: %w", err)
	}

	return nil
}

func (c *Client) GetStatus() ClientStatus {
	return ClientStatus(c.status.Load())
}

// IsReady returns whether the client is ready for operations
func (c *Client) IsReady() bool {
	return c.GetStatus() == StatusReady
}

// GetStats returns current client statistics
func (c *Client) GetStats() Stats {
	now := time.Now()
	lastOp, _ := c.stats.LastOperation.Load().(time.Time)

	return Stats{
			Operations:    c.stats.Operations.Load(),
			Errors:       c.stats.Errors.Load(),
			LastOperation: lastOp,
			StartTime:    c.stats.StartTime,
			Uptime:       now.Sub(c.stats.StartTime),
	}
}

func (c *Client) GetConfig() *Config {
	return c.config
}

func (c *Client) updateStats(err error) {
	if err != nil {
			c.stats.Errors.Add(1)
	}
	c.stats.Operations.Add(1)
	c.stats.LastOperation.Store(time.Now())
}

// Deletion Queue Operations
// Queue operations
func (c *Client) AddToDeletionQueue(ctx context.Context, queueKey string, userID string) error {
	start := time.Now()
	defer c.updateStats(nil)

	// Current timestamp as score
	score := float64(time.Now().Unix())

	// Build and execute ZADD command with current timestamp as score
	cmd := c.client.B().Zadd().
			Key(queueKey).
			ScoreMember().
			ScoreMember(score, userID).
			Build()

	err := c.client.Do(ctx, cmd).Error()
	if err != nil {
			c.stats.Errors.Add(1)
			c.logger.Error("failed to add to deletion queue",
					"queueKey", queueKey,
					"userID", userID,
					"error", err,
					"duration", time.Since(start))
			return fmt.Errorf("failed to add to deletion queue: %w", err)
	}

	c.logger.Debug("added to deletion queue",
			"queueKey", queueKey,
			"userID", userID,
			"duration", time.Since(start))
	return nil
}

func (c *Client) GetDeletionQueueItems(ctx context.Context, queueKey string, start, stop int64) ([]string, error) {
	timeStart := time.Now()
	defer c.updateStats(nil)

	// Build and execute ZRANGEBYSCORE command
	cmd := c.client.B().Zrangebyscore().
			Key(queueKey).
			Min(fmt.Sprintf("%d", start)).
			Max(fmt.Sprintf("%d", stop)).
			Build()

	result, err := c.client.Do(ctx, cmd).AsStrSlice()
	if err != nil {
			c.stats.Errors.Add(1)
			c.logger.Error("failed to get deletion queue items",
					"queueKey", queueKey,
					"error", err,
					"duration", time.Since(timeStart))
			return nil, fmt.Errorf("failed to get deletion queue items: %w", err)
	}

	c.logger.Debug("got deletion queue items",
			"queueKey", queueKey,
			"count", len(result),
			"duration", time.Since(timeStart))
	return result, nil
}

func (c *Client) RemoveFromDeletionQueue(ctx context.Context, queueKey string, start int64, userID string) error {
	timeStart := time.Now()
	defer c.updateStats(nil)

	// Build and execute ZREM command
	cmd := c.client.B().Zrem().
			Key(queueKey).
			Member(userID).
			Build()

	err := c.client.Do(ctx, cmd).Error()
	if err != nil {
			c.stats.Errors.Add(1)
			c.logger.Error("failed to remove from deletion queue",
					"queueKey", queueKey,
					"userID", userID,
					"error", err,
					"duration", time.Since(timeStart))
			return fmt.Errorf("failed to remove from deletion queue: %w", err)
	}

	c.logger.Debug("removed from deletion queue",
			"queueKey", queueKey,
			"userID", userID,
			"duration", time.Since(timeStart))
	return nil
}

func (c *Client) ClearCorruptedEntry(ctx context.Context, key string) error {
	return c.client.Do(ctx, c.client.B().Del().Key(key).Build()).Error()
}