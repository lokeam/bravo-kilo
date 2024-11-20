package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

type CompressionStats struct {
	OriginalSize      int64
	CompressedSize    int64
	RequestCount      int64
	Path              string
	CompressionRatio  float64
	AverageLatency    time.Duration
	LastUpdated       time.Time
	FailureCount      int64
	LastFailure       time.Time
	FailureMessages   map[string]int64
}

type PathConfig struct {
	Path            string
	MinSize         int
	Level           int
	CompressionRatio float64
	UpdateCount      int64
}

type CompressionMonitor struct {
	stats       map[string]*CompressionStats
	configs     map[string]*PathConfig
	mu          sync.RWMutex
	stopChan    chan struct{}
	logger      *slog.Logger
	done        chan struct{}
	stopping    atomic.Bool
}

func NewCompressionMonitor(ctx context.Context, logger *slog.Logger) *CompressionMonitor {
	cm := &CompressionMonitor{
		stats:     make(map[string]*CompressionStats),
		configs:   make(map[string]*PathConfig),
		stopChan:  make(chan struct{}),
		logger:    logger,
		done:      make(chan struct{}),
	}

	// Clean up resources daily
	cm.StartCleanup(ctx, 24 * time.Hour)

	return cm
}

// Get snapshot of current compression statistics
func (cm *CompressionMonitor) GetStats() *CompressionStats {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Create aggregated stats view
	totalStats := &CompressionStats{
		FailureMessages: make(map[string]int64),
	}

	// Aggregate stats across all paths
	for _, stats := range cm.stats {
		totalStats.RequestCount += stats.RequestCount
		totalStats.FailureCount += stats.FailureCount
		totalStats.OriginalSize += stats.OriginalSize
		totalStats.CompressedSize += stats.CompressedSize

		// Aggregate failure msgs
		for msg, count := range stats.FailureMessages {
			totalStats.FailureMessages[msg] += count
		}

		// Update last failure msg if more recent
		if stats.LastFailure.After(totalStats.LastFailure) {
			totalStats.LastFailure = stats.LastFailure
		}

		// Calculate overall compression ratio of compressed size by original size
		if totalStats.OriginalSize > 0 {
			totalStats.CompressionRatio = float64(totalStats.CompressedSize) / float64(totalStats.OriginalSize)
		}
	}

	return totalStats
}

func (cm *CompressionMonitor) StartCleanup(ctx context.Context, interval time.Duration) error{
	if interval < time.Minute {
		cm.logger.Warn("Cleanup interval too short, setting to 1 minute minimum",
				"requested", interval,
				"actual", time.Minute)
		interval = time.Minute
}

cm.logger.Info("Starting compression monitor cleanup routine",
		"interval", interval)

ticker := time.NewTicker(interval)
go func() {
		defer ticker.Stop()
		defer close(cm.done)

		for {
				select {
				case <-ticker.C:
						if err := cm.cleanup(ctx); err != nil {
								cm.logger.Error("Cleanup failed", "error", err)
						}
				case <-ctx.Done():
						cm.logger.Info("Cleanup routine stopped by context cancellation")
						return
				case <-cm.stopChan:
						cm.logger.Info("Cleanup routine stopped by stop signal")
						return
				}
		}
}()

return nil
}

func (cm *CompressionMonitor) Shutdown(ctx context.Context) error {
	// Ensure that Compression Monitor only stops once
	if !cm.stopping.CompareAndSwap(false, true) {
		return nil
	}

	cm.logger.Info("Starting compression monitor shutdown")

	// Signal cleanup to stop
	cm.mu.Lock()
	select {
	case <- cm.stopChan:
		cm.mu.Unlock()
		return nil
	default:
		close(cm.stopChan)
		cm.mu.Unlock()
	}

	// Wait for cleanup to finish or context to timeout
	select {
		case <-cm.done:
			cm.logger.Info("Compression monitor shutdown complete")
			return nil
		case <-ctx.Done():
			return fmt.Errorf("shutdown timeout: %w", ctx.Err())
	}
}

// Helpers
func (cm *CompressionMonitor) updateStats(
	path string,
	compressedSize int64,
	originalSize int64,
	latency time.Duration,
	) {
    // Guard clauses with logging instead of returns
    if path == "" {
			cm.logger.Error("Cannot update stats with empty path")
			return
		}
		if originalSize < 0 || compressedSize < 0 {
				cm.logger.Error("Invalid size values",
						"original", originalSize,
						"compressed", compressedSize)
				return
		}


	cm.mu.Lock()
	defer cm.mu.Unlock()

	stats, exists := cm.stats[path]
	if !exists {
		stats = &CompressionStats{
			Path:              path,
			FailureMessages:   make(map[string]int64),
		}
		cm.stats[path] = stats
	}

	stats.RequestCount++
	stats.CompressedSize += compressedSize
	stats.OriginalSize += originalSize
	stats.AverageLatency = (stats.AverageLatency * time.Duration(stats.RequestCount-1) +
			latency) / time.Duration(stats.RequestCount)
	stats.LastUpdated = time.Now()

	// Update compression ratio
	if stats.OriginalSize > 0 {
			stats.CompressionRatio = float64(stats.CompressedSize) / float64(stats.OriginalSize)
	}
}

func (cm *CompressionMonitor) getPathConfig(path string) *PathConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if config, exists := cm.configs[path]; exists {
		return config
	}

	return &PathConfig{
		Path:     path,
		MinSize:  1024,
		Level:    5,
	}
}

func (cm *CompressionMonitor) adjustCompressionLevel(path string, cfg AdaptiveCompressionConfig) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Get stats and validate sample size
	stats, exists := cm.stats[path]
	if !exists || stats.RequestCount < int64(cfg.SampleThreshold) {
		return
	}

	// Get path config
	config, exists := cm.configs[path]
	if !exists {
		return
	}

	// Ensure enough time has passed since last update
	if time.Since(stats.LastUpdated) < cfg.AdjustmentInterval {
		return
	}

	currentCompressionLevel := config.Level
	newCompressionLevel := currentCompressionLevel

	// Check compression ratio bounds
	if stats.CompressionRatio < cfg.MinCompressionRatio {

		// Compression too aggressive, decrease level
		newCompressionLevel = max(currentCompressionLevel - 1, cfg.MinLevel)
		cm.logger.Debug("Decreasing compression level due to low ratio",
			"path", path,
			"ratio", stats.CompressionRatio,
			"min_ratio", cfg.MinCompressionRatio)
	} else if stats.CompressionRatio > cfg.MaxCompressionRatio {

		// Compression too relaxed, increase level
		newCompressionLevel = min(currentCompressionLevel + 1, cfg.MaxLevel)
		cm.logger.Debug("Increasing compression level due to high ratio",
			"path", path,
			"ratio", stats.CompressionRatio,
			"max_ratio", cfg.MaxCompressionRatio)
	}

	// Check latency constraints
	if stats.AverageLatency > cfg.TargetLatency {
		// Response too slow, decrease compression
		newCompressionLevel = max(currentCompressionLevel - 1, cfg.MinLevel)
		cm.logger.Debug("Decreasing compression level due to high latency",
			"path", path,
			"latency", stats.AverageLatency,
			"target", cfg.TargetLatency)
	}

	// Update compression level if changed
	if newCompressionLevel != currentCompressionLevel {
		config.Level = newCompressionLevel
		config.UpdateCount++

		cm.logger.Info("Adjusted compression level",
		"path", path,
		"old_level", currentCompressionLevel,
		"new_level", newCompressionLevel,
		"ratio", stats.CompressionRatio,
		"latency", stats.AverageLatency)
	}
}

func (cm *CompressionMonitor) cleanup(ctx context.Context) error {
	defer func() {
		if r := recover(); r != nil {
				stack := debug.Stack()
				cm.logger.Error("Cleanup panic recovered",
						"error", r,
						"stack", string(stack))
		}
	}()

	cleanupCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
			cm.mu.Lock()
			defer cm.mu.Unlock()

			// Check before starting cleanup
			select {
			case <-cm.stopChan:
				done <- nil
				return
			default:
			}

			threshold := time.Now().Add(-24 * time.Hour)
			cleaned := 0

			for path, stats := range cm.stats {
            // Check both context and stopChan during iteration
            select {
            case <-cleanupCtx.Done():
                done <- cleanupCtx.Err()
                return
            case <-cm.stopChan:
                done <- nil
                return
            default:
                if stats.LastUpdated.Before(threshold) {
                    delete(cm.stats, path)
                    delete(cm.configs, path)
                    cleaned++
                }
            }
			}

			if cleaned > 0 {
					cm.logger.Info("Compression stats cleanup completed",
							"cleaned", cleaned,
							"timestamp", time.Now())
			}

			done <- nil
	}()

	select {
	case err := <-done:
			return err
	case <-cleanupCtx.Done():
			return fmt.Errorf("cleanup cancelled: %w", cleanupCtx.Err())
	}
}

func (cm *CompressionMonitor) recordFailure(path, reason string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	stats, exists := cm.stats[path]
	if !exists {
		stats = &CompressionStats{
			Path:               path,
			FailureMessages:    make(map[string]int64),
		}
		cm.stats[path] = stats
	}

	stats.FailureCount++
	stats.LastFailure = time.Now()
	stats.FailureMessages[reason]++
}
