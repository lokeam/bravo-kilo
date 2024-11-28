package handlers

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/lokeam/bravo-kilo/internal/books"
	"github.com/lokeam/bravo-kilo/internal/books/repository"
	"github.com/lokeam/bravo-kilo/internal/books/services"
	"github.com/lokeam/bravo-kilo/internal/shared/cache"
	"github.com/lokeam/bravo-kilo/internal/shared/rueidis"
	"github.com/lokeam/bravo-kilo/internal/shared/workers"

	"github.com/go-playground/validator/v10"
	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/time/rate"
)

// Handlers struct to hold the logger, models, and new components
type BookHandlers struct {
	authorRepo              repository.AuthorRepository
	bookRepo                repository.BookRepository
	formatRepo              repository.FormatRepository
	genreRepo               repository.GenreRepository
	tagRepo                 repository.TagRepository
	BookCache               repository.BookCache
	bookDeleter             repository.BookDeleter
	bookUpdater             services.BookUpdaterService
	bookService             services.BookService
	bookCacheService        services.BookCacheService
	exportService           services.ExportService
	exportLimiter           *rate.Limiter
	logger                  *slog.Logger
	bookModels              books.Models
	redisClient             *rueidis.Client
	cacheManager            *cache.CacheManager
	cacheWorker             *workers.CacheWorker
	sanitizer               *bluemonday.Policy
	validate                *validator.Validate
	DB                      *sql.DB

	cacheDurations struct {
		BookList          time.Duration
		BookDetail        time.Duration
		BooksByAuthor     time.Duration
		BooksByFormat     time.Duration
		BooksByGenre      time.Duration
		BooksByTag        time.Duration
		BookHomepage      time.Duration
		UserData          time.Duration
		DefaultTTL        time.Duration
	}
}

// Create a new Handlers instance
func NewBookHandlers(
	db *sql.DB,
	logger *slog.Logger,
	bookModels books.Models,
	authorRepo repository.AuthorRepository,
	bookRepo repository.BookRepository,
	formatRepo repository.FormatRepository,
	genreRepo repository.GenreRepository,
	tagRepo repository.TagRepository,
	userBooksRepo repository.UserBooksRepository,
	BookCache repository.BookCache,
	bookDeleter repository.BookDeleter,
	bookUpdater services.BookUpdaterService,
	bookService services.BookService,
	bookCacheService services.BookCacheService,
	exportService services.ExportService,
	redisClient *rueidis.Client,
	cacheManager *cache.CacheManager,
	cacheWorker *workers.CacheWorker,
	) (*BookHandlers, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	if authorRepo == nil {
		return nil, fmt.Errorf("authorRepo cannot be nil")
	}

	if bookRepo == nil {
		return nil, fmt.Errorf("bookRepo cannot be nil")
	}

	if formatRepo == nil {
		return nil, fmt.Errorf("formatRepo cannot be nil")
	}

	if genreRepo == nil {
		return nil, fmt.Errorf("genreRepo cannot be nil")
	}

	if exportService == nil {
		return nil, fmt.Errorf("exportService cannot be nil")
	}

	if BookCache == nil {
		return nil, fmt.Errorf("bookCache cannot be nil")
	}

	if bookCacheService == nil {
		return nil, fmt.Errorf("bookCacheService cannot be nil")
	}

	if bookDeleter == nil {
		return nil, fmt.Errorf("bookDeleter cannot be nil")
	}

	validate := validator.New()
	if validate == nil {
		return nil, fmt.Errorf("failed to initialize validator")
	}

	sanitizer := bluemonday.StrictPolicy()
	if sanitizer == nil {
			return nil, fmt.Errorf("failed to initialize sanitizer")
	}

	if redisClient == nil {
		return nil, fmt.Errorf("redisClient cannot be nil")
	}

	if cacheManager == nil {
		return nil, fmt.Errorf("cacheManager cannot be nil")
	}

	if cacheWorker == nil {
		return nil, fmt.Errorf("cacheWorker cannot be nil")
	}


	return &BookHandlers{
		DB:                db,
		logger:            logger,
		bookModels:        bookModels,
		authorRepo:        authorRepo,
		bookRepo:          bookRepo,
		formatRepo:        formatRepo,
		genreRepo:         genreRepo,
		tagRepo:           tagRepo,
		BookCache:         BookCache,
		bookDeleter:       bookDeleter,
		bookService:       bookService,
		bookCacheService:  bookCacheService,
		bookUpdater:       bookUpdater,
		exportService:     exportService,
		exportLimiter:     rate.NewLimiter(rate.Limit(1), 3),
		validate:          validate,
		sanitizer:         sanitizer,
		redisClient:       redisClient,
		cacheManager:      cacheManager,
		cacheWorker:       cacheWorker,

		cacheDurations: struct {
			BookList          time.Duration
			BookDetail        time.Duration
			BooksByAuthor     time.Duration
			BooksByFormat     time.Duration
			BooksByGenre      time.Duration
			BooksByTag        time.Duration
			BookHomepage      time.Duration
			UserData          time.Duration
			DefaultTTL        time.Duration
		}{
			BookList:          15 * time.Minute,
			BookDetail:        30 * time.Minute,
			BooksByAuthor:     15 * time.Minute,
			BooksByFormat:     15 * time.Minute,
			BooksByGenre:      15 * time.Minute,
			BooksByTag:        15 * time.Minute,
			BookHomepage:      5 * time.Minute,
			UserData:          30 * time.Minute,
			DefaultTTL:        15 * time.Minute,
		},
	}, nil
}

