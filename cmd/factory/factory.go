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
	bookservices "github.com/lokeam/bravo-kilo/internal/books/services"
	"github.com/lokeam/bravo-kilo/internal/shared/cache"
	"github.com/lokeam/bravo-kilo/internal/shared/core"
	"github.com/lokeam/bravo-kilo/internal/shared/domains"
	"github.com/lokeam/bravo-kilo/internal/shared/library"
	"github.com/lokeam/bravo-kilo/internal/shared/models"
	"github.com/lokeam/bravo-kilo/internal/shared/operations"
	"github.com/lokeam/bravo-kilo/internal/shared/organizer"
	"github.com/lokeam/bravo-kilo/internal/shared/processor/bookprocessor"
	"github.com/lokeam/bravo-kilo/internal/shared/rueidis"
	sharedservices "github.com/lokeam/bravo-kilo/internal/shared/services"
	"github.com/lokeam/bravo-kilo/internal/shared/transaction"
	"github.com/lokeam/bravo-kilo/internal/shared/validator"
	"github.com/lokeam/bravo-kilo/internal/shared/workers"
)

type Factory struct {
    RedisClient           *rueidis.Client
    BookHandlers          *handlers.BookHandlers
    SearchHandlers        *handlers.SearchHandlers
    AuthHandlers          *authhandlers.AuthHandlers
    DeletionWorker        *workers.DeletionWorker
    CacheWorker           *workers.CacheWorker
    CacheManager          *cache.CacheManager
    LibraryHandler        *library.LibraryHandler
    BaseValidator         *validator.BaseValidator
}

func NewFactory(
    ctx context.Context,
    db *sql.DB,
    redisClient *rueidis.Client,
    log *slog.Logger,
    ) (*Factory, error) {
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

    bookProcessor, err := bookprocessor.NewBookProcessor(
        log.With("component", "book_processor"),
    )
    if err != nil {
        log.Error("failed to create book processor", "error", err)
        return nil, fmt.Errorf("failed to create book processor: %w", err)
    }


    bookOrganizer, err := organizer.NewBookOrganizer(
        log.With("component", "book_organizer"),
    )
    if err != nil {
        log.Error("failed to create book organizer", "error", err)
        return nil, fmt.Errorf("failed to create book organizer: %w", err)
    }

    bookDomainAdapter := operations.NewBookDomainAdapter(
        bookRepo,
        log.With("component", "book_domain_adapter"),
    )

    domainOperation := operations.NewDomainOperation(
        core.BookDomainType,
        bookDomainAdapter,
        log.With("component", "domain_operation"),
    )

    baseValidator, err := validator.NewBaseValidator(
        log.With("component", "validator"),
        core.BookDomainType,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to initialize base validator: %w", err)
    }


    cacheOperation := operations.NewCacheOperation(
        redisClient,
        30 * time.Second,
        log.With("component", "cache_operation"),
        baseValidator,
    )
    processorOperation, err := operations.NewProcessorOperation(
        bookProcessor,
        bookOrganizer,
        30 * time.Second,
        log.With("component", "processor_operation"),
    )
    if err != nil {
        log.Error("failed to create processor operation", "error", err)
        return nil, fmt.Errorf("failed to create processor operation: %w", err)
    }


    // Initialize auth-related services
    operationsManager := operations.NewManager(
        cacheOperation,
        domainOperation,
        processorOperation,
    )

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
    bookService, err := bookservices.NewBookService(
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

    bookUpdaterService, err := bookservices.NewBookUpdaterService(
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

    exportService, err := bookservices.NewExportService(
        log,
        bookRepo,
    )
    if err != nil {
        log.Error("Error initializing export service", "error", err)
        return nil, err
    }

    bookCacheService := bookservices.NewBookCacheService(
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

    queryValidator, err := validator.NewQueryValidator(
        log.With("component", "query_validator"),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to initialize query validator: %w", err)
    }

    validationService, err := sharedservices.NewValidationService(
        baseValidator,
        queryValidator,
        log.With("component", "validation_service"),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to initialize validation service: %v", err)
    }

    libraryService, err := library.NewLibraryService(
        operationsManager,
        validationService,
        log.With("component", "library_service"),
    )
    if err != nil {
        return nil, fmt.Errorf("error initializing library service: %v", err)
    }

    libraryHandler := library.NewLibraryHandler(
        libraryService,
        validationService,
        log,
    )

    bookDomainHandler := domains.NewBookDomainHandler(bookHandlers, log.With("domain", "books"))
    if err := operationsManager.RegisterDomain(bookDomainHandler); err != nil {
        return nil, fmt.Errorf("failed to register book domain handler: %w", err)
    }

    // Initialize workers
    deletionWorker := workers.NewDeletionWorker(
        24*time.Hour,
        authService,
        log.With("worker", "deletion"),
    )

    return &Factory{
        RedisClient:           redisClient,
        BookHandlers:          bookHandlers,
        AuthHandlers:          authHandlers,
        SearchHandlers:        searchHandlers,
        DeletionWorker:        deletionWorker,
        CacheWorker:           cacheWorker,
        CacheManager:          cacheManager,
        LibraryHandler:    libraryHandler,
        BaseValidator:         baseValidator,
    }, nil
}