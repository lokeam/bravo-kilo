package init

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/lokeam/bravo-kilo/config"
	authhandlers "github.com/lokeam/bravo-kilo/internal/auth/handlers"
	authservices "github.com/lokeam/bravo-kilo/internal/auth/services"
	"github.com/lokeam/bravo-kilo/internal/books"
	bookcache "github.com/lokeam/bravo-kilo/internal/books/cache"
	"github.com/lokeam/bravo-kilo/internal/books/handlers"
	"github.com/lokeam/bravo-kilo/internal/books/repository"
	"github.com/lokeam/bravo-kilo/internal/books/services"
	"github.com/lokeam/bravo-kilo/internal/shared/cache"
	"github.com/lokeam/bravo-kilo/internal/shared/models"
	"github.com/lokeam/bravo-kilo/internal/shared/pages"
	library "github.com/lokeam/bravo-kilo/internal/shared/pages/library"
	"github.com/lokeam/bravo-kilo/internal/shared/redis"
	"github.com/lokeam/bravo-kilo/internal/shared/transaction"
	"github.com/lokeam/bravo-kilo/internal/shared/workers"
)

type Factory struct {
    RedisClient    *redis.RedisClient
    BookHandlers   *handlers.BookHandlers
    SearchHandlers *handlers.SearchHandlers
    AuthHandlers   *authhandlers.AuthHandlers
    DeletionWorker *workers.DeletionWorker
    CacheWorker    *workers.CacheWorker
    CacheManager   *cache.CacheManager
    LibraryHandler *library.Handler
    PageHandlers   *pages.PageHandlers
}

func NewFactory(ctx context.Context, db *sql.DB, redisClient *redis.RedisClient, log *slog.Logger) (*Factory, error) {
    if redisClient == nil {
        return nil, fmt.Errorf("error initializing factory: redis client is required")
    }

    // Initialize transaction manager
    transactionManager, err := transaction.NewDBManager(db, log)
    if err != nil {
        log.Error("Error initializing transaction manager", "error", err)
        return nil, err
    }

    // Initialize cache components
    cacheWorker := workers.NewCacheWorker(redisClient, log, 3)
    if cacheWorker == nil {
        return nil, fmt.Errorf("error initializing factory: cache worker is required")
    }

    // Initialize book-related repositories
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

    // Initialize cache invalidation components
    bookCacheInvalidator := bookcache.NewBookCacheInvalidator(
        bookCache,
        redisClient,
        log.With("component", "book_cache_invalidator"),
    )

    cacheManager := cache.NewCacheManager(
        bookCacheInvalidator,
        log,
    )

    // Initialize models
    bookModels, err := books.New(db, log)
    if err != nil {
        return nil, err
    }

    // Initialize auth-related repositories and models
    userRepo, err := models.NewUserRepository(db, log)
    if err != nil {
        log.Error("Error initializing user repository", "error", err)
        return nil, err
    }

    tokenModel := models.NewTokenModel(db, log)

    // Initialize auth-related services
    tokenService := authservices.NewTokenService(
			log.With("service", "token"),
			tokenModel,
			config.AppConfig.JWTPrivateKey,
			config.AppConfig.JWTPublicKey,
			os.Getenv("ENV") == "production",
		)

    oauthService, err := authservices.NewOAuthService(
			log.With("service", "oauth"),
			tokenModel,  // Changed: Pass tokenModel instead of credentials
		)
		if err != nil {
				log.Error("Error initializing OAuth service", "error", err)
				return nil, err
		}

    userDeletionService := workers.NewUserDeletionService(
        redisClient,
        log.With("service", "user_deletion"),
    )

    authService := authservices.NewAuthService(
			log.With("service", "auth"),
			transactionManager,
			userRepo,
			tokenModel,
			oauthService,
			tokenService,
			userDeletionService,
		)

    // Initialize book-related services
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
        log.Error("Error initializing book service", "error", err)
        return nil, err
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
        log.Error("Error initializing book updater service", "error", err)
        return nil, err
    }

    exportService, err := services.NewExportService(
        log,
        bookRepo,
    )
    if err != nil {
        log.Error("Error initializing export service", "error", err)
        return nil, err
    }

    bookCacheService := services.NewBookCacheService(
        redisClient,
        log.With("service", "book_cache"),
    )

    // Initialize handlers
    authHandlers := authhandlers.NewAuthHandlers(
        log.With("handler", "auth"),
        authService,
        oauthService,
        tokenService,
    )

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
        bookCacheService,
        exportService,
        redisClient,
        cacheManager,
        cacheWorker,
    )
    if err != nil {
        return nil, err
    }

    searchHandlers, err := handlers.NewSearchHandlers(
        log,
        bookRepo,
        bookCache,
        authHandlers,
    )
    if err != nil {
        return nil, err
    }

    pageHandlers, err := pages.NewPageHandlers(
        bookHandlers,
        redisClient,
        log.With("component", "page_handlers"),
        cacheManager,
    )

    // Initialize workers
    deletionWorker := workers.NewDeletionWorker(
        24*time.Hour,
        authService,
        log.With("worker", "deletion"),
    )

    return &Factory{
        RedisClient:    redisClient,
        BookHandlers:   bookHandlers,
        AuthHandlers:   authHandlers,
        SearchHandlers: searchHandlers,
        DeletionWorker: deletionWorker,
        CacheWorker:    cacheWorker,
        CacheManager:   cacheManager,
        PageHandlers: pageHandlers,
    }, nil
}