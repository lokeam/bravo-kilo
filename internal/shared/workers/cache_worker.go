package workers

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/lokeam/bravo-kilo/internal/shared/redis"
)

var (
	ErrQueueFull = fmt.Errorf("cache invalidation queue is full")
)

type CacheInvalidationJob struct {
	Keys         []string
	UserID       int
	BookID       int
	Timestamp    time.Time
	Attempts     int
}

type CacheWorker struct {
	jobs             chan CacheInvalidationJob
	workers          int
	redisClient      *redis.RedisClient
	logger           *slog.Logger
	maxAttempts      int
	retryInterval    time.Duration
	metrics          *CacheWorkerMetrics
	ctx              context.Context
	cancel           context.CancelFunc
}

type CacheWorkerMetrics struct {
	JobsProcessed      int64
	JobsFailed         int64
	JobsRetried        int64
	KeysProcessed      int64
	KeysFailed         int64
	ProcessingTime     time.Duration
	mu                 sync.RWMutex
}

type MetricsSnapshot struct {
	JobsProcessed   int64
	JobsFailed      int64
	JobsRetried     int64
	KeysProcessed   int64
	KeysFailed      int64
	ProcessingTime  time.Duration
}

func NewCacheWorker(redisClient *redis.RedisClient, logger *slog.Logger, workers int) (*CacheWorker) {
	if redisClient == nil {
		return nil
	}
	if logger == nil {
		return nil
	}
	if workers < 1 {
		workers = 1
	}

	ctx, cancel := context.WithCancel(context.Background())
	w := &CacheWorker{
		jobs:            make(chan CacheInvalidationJob, 1000),
		workers:         workers,
		redisClient:     redisClient,
		logger:          logger,
		maxAttempts:     3,
		retryInterval:   time.Second * 5,
		metrics:         &CacheWorkerMetrics{},
		ctx:             ctx,
		cancel:          cancel,
	}

	w.startWorkers()
	return w
}

func (w *CacheWorker) startWorkers() {
	for i := 0; i < w.workers; i++ {
		go w.worker(i)
	}
}

func (w *CacheWorker) worker(id int) {
	w.logger.Info("Starting cache invalidation worker", "workerID", id)

	for {
		select {
		case <-w.ctx.Done():
			w.logger.Info("Shutting down cache invalidation worker", "workerID", id)
			return
		case job := <-w.jobs:
			start := time.Now()

			if err := w.processInvalidateQueueJob(job); err != nil {
				w.handleFailedInvalidationJob(job, err)
			}

			w.updateMetrics(time.Since(start))
		}
	}
}

func (w *CacheWorker) processInvalidateQueueJob(job CacheInvalidationJob) error {
	ctx, cancel := context.WithTimeout(w.ctx, time.Second*5)
	defer cancel()

	// Delete each cache key individually
	for _, key := range job.Keys {
		if err := w.redisClient.Delete(ctx, key); err != nil {
			w.metrics.mu.Lock()
			w.metrics.KeysFailed++
			w.metrics.mu.Unlock()

			return fmt.Errorf("error deleting cache key %s: %w", key, err)
		}

		w.metrics.mu.Lock()
		w.metrics.KeysProcessed++
		w.metrics.mu.Unlock()
	}

	w.logger.Info("Successfully invalidated caches",
			"userID", job.UserID,
			"bookID", job.BookID,
			"keysCount", fmt.Sprintf("%d", len(job.Keys)),
	)

	return nil
}

func (w *CacheWorker) handleFailedInvalidationJob(job CacheInvalidationJob, err error) {
	w.metrics.mu.Lock()
	w.metrics.JobsFailed++
	w.metrics.mu.Unlock()

	if job.Attempts < w.maxAttempts {
		job.Attempts++
		w.metrics.mu.Lock()
		w.metrics.JobsRetried++
		w.metrics.mu.Unlock()

		// Exponential backoff
		time.Sleep(w.retryInterval * time.Duration(job.Attempts*job.Attempts))
		w.jobs <- job

		w.logger.Warn("Retrying cache invalidation",
			"attempt", job.Attempts,
			"userID", job.UserID,
			"bookID", job.BookID,
			"error", err,
		)
	} else {
		w.logger.Error("Cache invalidation failed after max attempts",
			"attempt", job.Attempts,
			"userID", job.UserID,
			"bookID", job.BookID,
			"error", err,
		)
	}
}

func (w *CacheWorker) updateMetrics(duration time.Duration) {
	w.metrics.mu.Lock()
	defer w.metrics.mu.Unlock()

	w.metrics.JobsProcessed++
	w.metrics.ProcessingTime += duration
}

func (w *CacheWorker) EnqueueInvalidationJob(ctx context.Context, job CacheInvalidationJob) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	select {
	case w.jobs <- job:
		return nil
	case <- ctx.Done():
		return ctx.Err()
	default:
		// Queue is full
		w.logger.Error("Cache invalidation queue is full, dropping job",
			"userID", job.UserID,
			"bookID", job.BookID,
			"keyesCount", len(job.Keys),
		)
		return ErrQueueFull
	}
}

func (w *CacheWorker) Shutdown() {
	// Signal workers to stop
	w.cancel()

	// Wait for all jobs to complete or context to cancel
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if len(w.jobs) == 0 {
				close(w.jobs)
				return
			}
		case <- time.After(30 * time.Second):
			// Force shutdown after timeout
			close(w.jobs)
			return
		}
	}
}

func (w *CacheWorker) GetMetrics() MetricsSnapshot {
	w.metrics.mu.RLock()
	defer w.metrics.mu.RUnlock()

	return MetricsSnapshot{
		JobsProcessed:   w.metrics.JobsProcessed,
		JobsFailed:      w.metrics.JobsFailed,
		JobsRetried:     w.metrics.JobsRetried,
		KeysProcessed:   w.metrics.KeysProcessed,
		KeysFailed:      w.metrics.KeysFailed,
		ProcessingTime:  w.metrics.ProcessingTime,
	}
}