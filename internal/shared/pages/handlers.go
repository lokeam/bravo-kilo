package pages

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/lokeam/bravo-kilo/internal/books/handlers"
	"github.com/lokeam/bravo-kilo/internal/shared/cache"
	"github.com/lokeam/bravo-kilo/internal/shared/pages/library"
	"github.com/lokeam/bravo-kilo/internal/shared/pages/library/domains"
	"github.com/lokeam/bravo-kilo/internal/shared/redis"
)

type PageHandlers struct {
	Library *library.Handler  // Points to the implementation in library/handler.go
	logger  *slog.Logger
	metrics *PageMetrics
}

type PageMetrics struct {
	LibraryErrors int64
}

func NewPageHandlers(
	bookHandlers *handlers.BookHandlers,
	redisClient *redis.RedisClient,
	logger *slog.Logger,
	cacheManager *cache.CacheManager,
) (*PageHandlers, error) {
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

	// Create domain handler
	bookDomain := domains.NewBookDomainHandler(
		bookHandlers,
		logger.With("component", "book_domain"),
	)


	// Initialize library handler
	libraryHandler, err := library.NewHandler(
		bookHandlers,
		redisClient,
		logger.With("component", "library_handler"),
		cacheManager,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize library handler: %w", err)
	}

	// Regisrer domain handler
	libraryHandler.RegisterDomain(bookDomain)

	return &PageHandlers{
		Library: libraryHandler,
		logger: logger.With("component", "page_handlers"),
		metrics: &PageMetrics{},
	}, nil
}

// Handle graceful shutdown + metrics reporting
func (h *PageHandlers) Cleanup() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()


	// Add cleanup channels
	done := make(chan error)
	go func() {
		done <- h.Library.Cleanup()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("cleanup timed out")
	}

	h.logger.Info("page handlers cleanup successful")
	return nil
}