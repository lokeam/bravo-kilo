package redis

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type HealthChecker struct {
	client      *RedisClient
	config      *RedisConfig
	logger      *slog.Logger
	status      HealthStatus
	thresholds  HealthThresholds
	mu          sync.RWMutex
	stopChan    chan struct{}
	doneChan    chan struct{}

	// Metric windows for tracking
	errorWindow   *TimeWindow
	latencyWindow *TimeWindow
}

// Track metrics over time
type TimeWindow struct {
	duration    time.Duration
	samples     []Sample
	mu          sync.RWMutex
}

type Sample struct {
	timestamp   time.Time
	value       float64
}


func NewHealthChecker(client *RedisClient, config *RedisConfig, logger *slog.Logger) (*HealthChecker, error) {
	hc := &HealthChecker{
		client:           client,
		config:           config,
		logger:           logger,
		stopChan:         make(chan struct{}),
		doneChan:         make(chan struct{}),
		thresholds:       getDefaultThresholds(),
		errorWindow:      NewTimeWindow(5 * time.Minute),
		latencyWindow:    NewTimeWindow(5 * time.Minute),
	}

	return hc, nil
}

func (h *HealthChecker) Start(ctx context.Context) error {
	go h.healthCheckLoop(ctx)
	return nil
}

func (h *HealthChecker) Stop() error {
	close(h.stopChan)
	<-h.doneChan
	return nil
}

func (h *HealthChecker) GetStatus() HealthStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.status
}

func (h *HealthChecker) performHealthCheck(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	metrics := HealthMetrics{
		LastPingTime: time.Now(),
	}

	// 1. Check basic connection health
	start := time.Now()
	if err := h.checkConnectivity(ctx); err != nil {
		return h.handleHealthCheckError(err, metrics)
	}
	metrics.ResponseTime = time.Since(start)

	// 2. Check connection pool
	poolMetrics, err := h.checkConnectionPool()
	if err != nil {
		return h.handleHealthCheckError(err, metrics)
	}
	metrics.ConnectionPool = poolMetrics

	// 3. Check memory usage
	memoryUsage, err := h.checkMemoryUsage(ctx)
	if err != nil {
		return h.handleHealthCheckError(err, metrics)
	}
	metrics.MemoryUsage = memoryUsage

	// 4. Determine error rate
	metrics.ErrorRate, err = h.determineErrorRate()
	if err != nil {
		return h.handleHealthCheckError(err, metrics)
	}

	// 5. Update status based on metrics
	h.updateHealthStatus(metrics)
	return nil
}

func (h *HealthChecker) checkConnectivity(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, h.config.HealthConfig.Timeout)
	defer cancel()

	return h.client.Ping(ctx)
}

func (h *HealthChecker) checkConnectionPool() (PoolMetrics, error) {
	stats := h.client.GetClient().PoolStats()
	return PoolMetrics{
			TotalConnections:  stats.TotalConns,
			ActiveConnections: stats.Hits,
			IdleConnections:   stats.IdleConns,
			Hits:              stats.Hits,
			Misses:            stats.Misses,
			Timeouts:          stats.Timeouts,
			StaleConns:        stats.StaleConns,
	}, nil
}

func (h *HealthChecker) checkMemoryUsage(ctx context.Context) (int64, error) {
	info, err := h.client.GetClient().Info(ctx, "memory").Result()
	if err != nil {
			return 0, fmt.Errorf("failed to get memory info: %w", err)
	}
	// Parse memory usage from INFO command response
	return parseMemoryUsage(info), nil
}

func (h *HealthChecker) determineErrorRate() (float64, error) {
	return h.errorWindow.GetRate(), nil
}

func (h *HealthChecker) updateHealthStatus(metrics HealthMetrics) {
	h.status.Metrics = metrics
	h.status.LastCheck = time.Now()

	// Check against thresholds
	degraded := false
	messages := []string{}

	if metrics.ResponseTime > h.thresholds.MaxResponseTime {
			degraded = true
			messages = append(messages, fmt.Sprintf("high response time: %v", metrics.ResponseTime))
	}

	if metrics.ErrorRate > h.thresholds.MaxErrorRate {
			degraded = true
			messages = append(messages, fmt.Sprintf("high error rate: %.2f%%", metrics.ErrorRate*100))
	}

	// Convert uint32 to int for comparison
	if int(metrics.ConnectionPool.IdleConnections) < h.thresholds.MinIdleConnections {
			degraded = true
			messages = append(messages, "low idle connections")
	}

	if metrics.MemoryUsage > h.thresholds.MaxMemoryUsage {
			degraded = true
			messages = append(messages, "high memory usage")
	}

	h.status.Degraded = degraded
	if len(messages) > 0 {
			h.status.Message = fmt.Sprintf("Health check issues: %v", messages)
	} else {
			h.status.Message = "Health check passed without issues"
	}

	h.status.IsHealthy = !degraded
}

func (h *HealthChecker) healthCheckLoop(ctx context.Context) {
	ticker := time.NewTicker(h.config.HealthConfig.Interval)
	defer ticker.Stop()
	defer close(h.doneChan)

	for {
		select {
		case <-h.stopChan:
			return
		case <-ticker.C:
			h.performHealthCheck(ctx)
		}
	}
}

func (h *HealthChecker) handleHealthCheckError(err error, metrics HealthMetrics) error {
	h.status.LastError = err
	h.status.ConsecutiveFail++
	h.status.ConsecutiveOK = 0
	h.status.IsHealthy = false
	h.status.Metrics = metrics

	h.errorWindow.Add(1.0)

	if h.status.ConsecutiveFail >= h.config.HealthConfig.MaxRetries {
		h.logger.Error("Health check failures exceeded max retry threshold", "consecutive_failures", h.status.ConsecutiveFail, "error", err)
	}

	return err
}

func parseMemoryUsage(info string) int64 {
    // Parse the Redis INFO command response to extract used_memory
    // Format: used_memory:1234567
		var memoryUsed int64
		fmt.Sscanf(info, "used_memory:%d", &memoryUsed)

		return memoryUsed
}

func getDefaultThresholds() HealthThresholds {
	return HealthThresholds{
		MaxResponseTime: time.Second,            // 1 second max response time
		MaxErrorRate: 0.05,                      // 5% error rate
		MinIdleConnections: 2,                   // At least 2 idle connections
		MaxMemoryUsage: 1024 * 1024 * 100,       // 100MB max memory usage
		MaxWaitDuration: time.Second * 2,        // 2 seconds max wait time
	}
}