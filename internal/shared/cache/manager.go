package cache

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/lokeam/bravo-kilo/internal/shared/utils"
)

type L1CacheInvalidator interface {
    InvalidateL1Cache(itemID, userID int) error
    GetType() string
}

type L2CacheInvalidator interface {
    InvalidateL2Cache(ctx context.Context, keys []string) error
    GetCacheKeys(itemID, userID int) []string
}

type CacheManager struct {
    l2Invalidator L2CacheInvalidator
    logger        *slog.Logger
    metrics       *CacheMetrics
    mu            sync.RWMutex
}

type CacheMetrics struct {
		mu             sync.RWMutex
		Operations     map[string]int64

		L1Failures     int64
        L2Failures     int64

		L2Hits         int64
		L2Misses       int64
		Errors         int64
		UnmarshalErrs  int64

    TotalOps       int64
    LastError      error
    LastErrorTime  time.Time

		// Operation metrics
		OperationLatency  map[string]time.Duration
		OperationCount    map[string]int64
		ErrorCount        map[string]int64
}

func NewCacheMetrics() *CacheMetrics {
	return &CacheMetrics{
		OperationLatency: make(map[string]time.Duration),
		OperationCount:   make(map[string]int64),
		ErrorCount:       make(map[string]int64),

	}
}

func NewCacheManager(l2Invalidator L2CacheInvalidator, logger *slog.Logger) *CacheManager {
    if l2Invalidator == nil {
        panic("l2Invalidator cannot be nil")
    }
    if logger == nil {
        panic("logger cannot be nil")
    }

    return &CacheManager{
        l2Invalidator: l2Invalidator,
        logger:        logger,
        metrics:       NewCacheMetrics(),
    }
}

// InvalidateCache handles cache invalidation with proper error handling and metrics
func (cm *CacheManager) InvalidateCache(ctx context.Context, l1Invalidator L1CacheInvalidator, itemID, userID int) error {
    if ctx == nil {
        return fmt.Errorf("context cannot be nil")
    }
    if l1Invalidator == nil {
        return fmt.Errorf("l1Invalidator cannot be nil")
    }

    start := time.Now()
    defer func() {
        cm.recordOperationMetrics("cache_invalidation", time.Since(start))
    }()

    // Track total operations
    cm.mu.Lock()
    cm.metrics.TotalOps++
    cm.mu.Unlock()

    // Create error channel for collecting errors
    errChan := make(chan error, 2)
    var wg sync.WaitGroup

    // L1 Cache Invalidation
    wg.Add(1)
    go func() {
        defer wg.Done()
        if err := l1Invalidator.InvalidateL1Cache(itemID, userID); err != nil {
            cm.recordL1Failure(err)
            errChan <- fmt.Errorf("L1 cache invalidation failed for %s: %w",
                l1Invalidator.GetType(), err)

            cm.logger.Error("L1 cache invalidation failed",
                "error", err,
                "type", l1Invalidator.GetType(),
                "userID", userID,
                "itemID", itemID,
            )
        }
    }()

    // L2 Cache Invalidation
    wg.Add(1)
    go func() {
        defer wg.Done()

        cacheKeys := cm.l2Invalidator.GetCacheKeys(itemID, userID)

        // Invalidate L2 cache with parent context
        if err := cm.l2Invalidator.InvalidateL2Cache(ctx, cacheKeys); err != nil {
            cm.recordL2Failure(err)
            errChan <- fmt.Errorf("L2 cache invalidation failed: %w", err)

            cm.logger.Error("L2 cache invalidation failed",
                "error", err,
                "userID", userID,
                "itemID", itemID,
                "keys", cacheKeys,
            )
        }
    }()

    // Wait for all operations to complete
    go func() {
        wg.Wait()
        close(errChan)
    }()

    // Collect errors
    var errors []error
    for err := range errChan {
        errors = append(errors, err)
    }

    // Return combined errors if any occurred
    if len(errors) > 0 {
        return fmt.Errorf("cache invalidation errors: %v", errors)
    }

    return nil
}

// Helper methods for metrics
func (cm *CacheManager) recordOperationMetrics(operation string, duration time.Duration) {
	cm.metrics.mu.Lock()
	defer cm.metrics.mu.Unlock()

	cm.metrics.OperationCount[operation]++
	cm.metrics.OperationLatency[operation] += duration
}

func (cm *CacheManager) recordL1Failure(err error) {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    cm.metrics.L1Failures++
    cm.metrics.LastError = err
    cm.metrics.LastErrorTime = time.Now()
		cm.metrics.ErrorCount["l1_invalidation"]++
}

func (cm *CacheManager) recordL2Failure(err error) {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    cm.metrics.L2Failures++
    cm.metrics.LastError = err
    cm.metrics.LastErrorTime = time.Now()
		cm.metrics.ErrorCount["l2_invalidation"]++
}

// GetMetrics returns a snapshot of current metrics
func (cm *CacheManager) GetMetrics() CacheMetrics {
    cm.mu.RLock()
    defer cm.mu.RUnlock()

    return CacheMetrics{
			L1Failures:        cm.metrics.L1Failures,
			L2Failures:        cm.metrics.L2Failures,
			TotalOps:          cm.metrics.TotalOps,
			LastError:         cm.metrics.LastError,
			LastErrorTime:     cm.metrics.LastErrorTime,
			OperationLatency:  utils.CopyMap(cm.metrics.OperationLatency),
			OperationCount:    utils.CopyMap(cm.metrics.OperationCount),
			ErrorCount:        utils.CopyMap(cm.metrics.ErrorCount),
		}
}

// Health check method
func (cm *CacheManager) HealthCheck(ctx context.Context) error {
    cm.mu.RLock()
    defer cm.mu.RUnlock()

    // Consider system unhealthy if error rate is too high
    failureRate := float64(cm.metrics.L1Failures+cm.metrics.L2Failures) /
        float64(cm.metrics.TotalOps)

    if failureRate > 0.15 { // 15% failure rate threshold
        return fmt.Errorf("cache system unhealthy: failure rate %.2f%%", failureRate*100)
    }

    return nil
}