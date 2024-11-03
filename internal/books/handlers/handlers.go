package handlers

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/lokeam/bravo-kilo/internal/books"
	"github.com/lokeam/bravo-kilo/internal/books/repository"
	"github.com/lokeam/bravo-kilo/internal/books/services"
	"github.com/lokeam/bravo-kilo/internal/shared/redis"
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
	bookRedisCache          repository.BookRedisCache
	bookDeleter             repository.BookDeleter
	bookUpdater             services.BookUpdaterService
	bookService             services.BookService
	exportService           services.ExportService
	exportLimiter           *rate.Limiter
	logger                  *slog.Logger
	bookModels              books.Models
	redisClient             *redis.RedisClient
	cacheWorker             *workers.CacheWorker
	sanitizer               *bluemonday.Policy
	validate                *validator.Validate
	DB                      *sql.DB
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
	bookCache repository.BookCache,
	bookRedisCache repository.BookRedisCache,
	bookDeleter repository.BookDeleter,
	bookUpdater services.BookUpdaterService,
	bookService services.BookService,
	exportService services.ExportService,
	redisClient *redis.RedisClient,
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

	if bookCache == nil {
		return nil, fmt.Errorf("bookCache cannot be nil")
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
		BookCache:         bookCache,
		bookRedisCache:    bookRedisCache,
		bookDeleter:       bookDeleter,
		bookService:       bookService,
		bookUpdater:       bookUpdater,
		exportService:     exportService,
		exportLimiter:     rate.NewLimiter(rate.Limit(1), 3),
		validate:          validate,
		sanitizer:         sanitizer,
		redisClient:       redisClient,
		cacheWorker:       cacheWorker,
	}, nil
}
