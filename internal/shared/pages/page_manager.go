package pages

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/lokeam/bravo-kilo/internal/books/handlers"
	"github.com/lokeam/bravo-kilo/internal/shared/cache"
	"github.com/lokeam/bravo-kilo/internal/shared/pages/library"
	"github.com/lokeam/bravo-kilo/internal/shared/pages/library/domains"
	"github.com/lokeam/bravo-kilo/internal/shared/redis"
	"github.com/lokeam/bravo-kilo/internal/shared/validator"
)

// Coordinates all pages

// Initializes page handlers

// Manages shared resources

type PageManager struct {
	Library *library.LibraryPageHandler
	logger  *slog.Logger
	metrics *PageMetrics
}

type PageMetrics struct {
	LibraryErrors int64
	// Future metrics can be added here
}

func NewPageManager(
	bookHandlers *handlers.BookHandlers,
	redisClient *redis.RedisClient,
	logger *slog.Logger,
	cacheManager *cache.CacheManager,
	baseValidator *validator.BaseValidator,
) (*PageManager, error) {
	if bookHandlers == nil {
		return nil, fmt.Errorf("bookHandlers cannot be nil")
	}
	if redisClient == nil {
			return nil, fmt.Errorf("redisClient cannot be nil")
	}
	if logger == nil {
			return nil, fmt.Errorf("logger cannot be nil")
	}
	if cacheManager == nil {
			return nil, fmt.Errorf("cacheManager cannot be nil")
	}
	if baseValidator == nil {
		return nil, fmt.Errorf("baseValidator cannot be nil")
	}

	// Create domain handler
	bookDomain := domains.NewBookDomainHandler(
		bookHandlers,
		logger.With("component", "book_domain"),
	)


	// Initialize library handler
	libraryPage, err := library.NewLibraryPageHandler(
		bookHandlers,
		redisClient,
		logger.With("component", "library_handler"),
		cacheManager,
		baseValidator,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize library handler: %w", err)
	}

	// Regisrer domain handler
	libraryPage.RegisterDomain(bookDomain)

	return &PageManager{
		Library: libraryPage,
		logger: logger.With("component", "page_handlers"),
		metrics: &PageMetrics{},
	}, nil
}

// Handle graceful shutdown + metrics reporting
func (pm *PageManager) Cleanup() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan error)
	go func() {
			var err error
			defer func() {
					if r := recover(); r != nil {
							err = fmt.Errorf("panic during cleanup: %v", r)
					}
					done <- err
			}()

			err = pm.Library.Cleanup()
			if err != nil {
					pm.logger.Error("library cleanup failed",
							"error", err,
							"errors_count", pm.metrics.LibraryErrors,
					)
					return
			}

			pm.logger.Info("page manager cleanup successful",
					"library_errors", pm.metrics.LibraryErrors,
			)
	}()

	select {
	case err := <-done:
			return err
	case <-ctx.Done():
			return fmt.Errorf("cleanup timed out after %v", 5*time.Second)
	}
}

// New method to get metrics
func (pm *PageManager) GetMetrics() PageMetrics {
	return *pm.metrics
}

// New method to increment error count
func (pm *PageManager) IncrementLibraryErrors() {
	// Using atomic operation for thread safety
	atomic.AddInt64(&pm.metrics.LibraryErrors, 1)
}

// New method to reset metrics
func (pm *PageManager) ResetMetrics() {
	atomic.StoreInt64(&pm.metrics.LibraryErrors, 0)
}

func (pm *PageManager) IsHealthy() bool {
	return pm.Library != nil && pm.metrics != nil
}