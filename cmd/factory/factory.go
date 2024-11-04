package init

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	authhandlers "github.com/lokeam/bravo-kilo/internal/auth/handlers"
	"github.com/lokeam/bravo-kilo/internal/books"
	bookcache "github.com/lokeam/bravo-kilo/internal/books/cache"
	"github.com/lokeam/bravo-kilo/internal/books/handlers"
	"github.com/lokeam/bravo-kilo/internal/books/repository"
	"github.com/lokeam/bravo-kilo/internal/books/services"
	"github.com/lokeam/bravo-kilo/internal/shared/cache"
	"github.com/lokeam/bravo-kilo/internal/shared/models"
	"github.com/lokeam/bravo-kilo/internal/shared/redis"
	"github.com/lokeam/bravo-kilo/internal/shared/transaction"
	"github.com/lokeam/bravo-kilo/internal/shared/workers"
)

// Factory initializes all components and returns them
type Factory struct {
	RedisClient     *redis.RedisClient
	BookHandlers    *handlers.BookHandlers
	SearchHandlers  *handlers.SearchHandlers
	AuthHandlers    *authhandlers.AuthHandlers
	DeletionWorker  *workers.DeletionWorker
	CacheWorker     *workers.CacheWorker
	CacheManager    *cache.CacheManager
}

// NewFactory initializes repositories, services, and handlers
func NewFactory(ctx context.Context, db *sql.DB, redisClient *redis.RedisClient, log *slog.Logger) (*Factory, error) {
	if redisClient == nil {
		return nil, fmt.Errorf("error initializing factory: redis client is required")
	}

	cacheWorker := workers.NewCacheWorker(redisClient, log, 3)
	if cacheWorker == nil {
		return nil, fmt.Errorf("error initializing factory: cache worker is required")
	}

	// Initialize repositories
	authorRepo, err := repository.NewAuthorRepository(db, log)
	if err != nil {
		log.Error("Error initializing author repository", "error", err)
		return nil, err
	}

	bookCache, err := repository.NewBookCache(ctx, db, log)
	if err != nil {
		log.Error("Error initializing book cache", "error", err)
		return nil, err
	}

	genreRepo, err := repository.NewGenreRepository(db, log, bookCache)
	if err != nil {
		log.Error("Error initializing genre repository", "error", err)
		return nil, err
	}

	formatRepo, err := repository.NewFormatRepository(db, log, bookCache)
	if err != nil {
		log.Error("Error initializing format repository", "error", err)
		return nil, err
	}

	tagRepo, err := repository.NewTagRepository(db, log)
	if err != nil {
		log.Error("Error initializing tag repository", "error", err)
		return nil, err
	}

	bookRepo, err := repository.NewBookRepository(db, log, authorRepo, genreRepo, formatRepo)
	if err != nil {
		log.Error("Error initializing book repository", "error", err)
		return nil, err
	}

	bookDeleter, err := repository.NewBookDeleter(db, log)
	if err != nil {
		log.Error("Error initializing book deleter", "error", err)
		return nil, err
	}

	userBooksRepo, err := repository.NewUserBooksRepository(db, log)
	if err != nil {
		log.Error("Error initializing user books repository", "error", err)
		return nil, err
	}

	bookCacheInvalidator := bookcache.NewBookCacheInvalidator(
		bookCache,   // L1 cache
		redisClient, // L2 cache
		log.With("component", "book_cache_invalidator"),
	)

	cacheManager := cache.NewCacheManager(
		bookCacheInvalidator,
		log,
	)

	// Initialize transaction manager
	transactionManager, err := transaction.NewDBManager(db, log)
	if err != nil {
		log.Error("Error initializing transaction manager", "error", err)
		os.Exit(1)
	}

	// Initialize services
	bookService, err := services.NewBookService(
		bookRepo,
		authorRepo,
		genreRepo,
		formatRepo,
		tagRepo,
		log,
		transactionManager,
	)
	if err != nil {
		log.Error("Error initializing book service manager", "error", err)
		os.Exit(1)
	}

	bookUpdaterService, err := services.NewBookUpdaterService(
		db,
		log,
		bookRepo,
		authorRepo,
		bookCache,
		formatRepo,
		genreRepo,
		tagRepo,
		bookService,
		transactionManager,
	)
	if err != nil {
		log.Error("Error initializing book updater service manager", "error", err)
		os.Exit(1)
	}

	exportService, err := services.NewExportService(
		log,
		bookRepo,
	)
	if err != nil {
		log.Error("Error initializing export service manager", "error", err)
		os.Exit(1)
	}

	bookCacheService := services.NewBookCacheService(
		redisClient,
		log.With("service", "book_cache"),
	)

	// Initialize models
	bookModels, err := books.New(db, log)
	if err != nil {
		return nil, err
	}

	// Initialize models for auth
	userRepo, err := models.NewUserRepository(db, log)
	if err != nil {
		log.Error("Error initializing user repository", "error", err)
		return nil, err
	}

	authModels := models.Models{
		User:  userRepo,
		Token: &models.TokenModel{DB: db, Logger: log},
	}

	// Initialize handlers
	bookHandlers, err := handlers.NewBookHandlers(
    db,
    log,
    bookModels,
    authorRepo,
    bookRepo,
    formatRepo,
    genreRepo,
    tagRepo,
    userBooksRepo,
    bookCache,
    bookDeleter,
    bookUpdaterService,
    bookService,
    bookCacheService,      // Changed
    exportService,         // Changed
    redisClient,
    cacheManager,         // Changed
    cacheWorker,          // Changed
	)
	if err != nil {
		return nil, err
	}

	authHandlers, err := authhandlers.NewAuthHandlers(
		log,
		authModels,
		transactionManager,
		bookCacheService,
		db,
	)
	if err != nil {
			return nil, err
	}

	searchHandlers, err := handlers.NewSearchHandlers(log, bookRepo, bookCache, authHandlers)
	if err != nil {
		return nil, err
	}

	deletionWorker := workers.NewDeletionWorker(24*time.Hour, authHandlers)

// Return all handlers and services inside the Factory
	return &Factory{
		RedisClient:    redisClient,
		BookHandlers:   bookHandlers,
		AuthHandlers:   authHandlers,
		SearchHandlers: searchHandlers,
		DeletionWorker: deletionWorker,
		CacheWorker:    cacheWorker,
		CacheManager:   cacheManager,
	}, nil
}
