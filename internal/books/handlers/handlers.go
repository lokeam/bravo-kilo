package handlers

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	"github.com/lokeam/bravo-kilo/internal/books"
	"github.com/lokeam/bravo-kilo/internal/books/repository"
	"github.com/lokeam/bravo-kilo/internal/books/services"

	"github.com/go-playground/validator/v10"
	"github.com/lokeam/bravo-kilo/internal/shared/utils"
	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/time/rate"
)

var claims utils.Claims
var jwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))

// Handlers struct to hold the logger, models, and new components
type BookHandlers struct {
	authorRepo        repository.AuthorRepository
	bookRepo          repository.BookRepository
	formatRepo        repository.FormatRepository
	genreRepo         repository.GenreRepository
	tagRepo           repository.TagRepository
	bookCache         repository.BookCache
	bookDeleter       repository.BookDeleter
	bookUpdater       services.BookUpdaterService
	bookService       services.BookService
	exportService     services.ExportService
	exportLimiter     *rate.Limiter
	logger            *slog.Logger
	bookModels        books.Models
	sanitizer         *bluemonday.Policy
	searchHandlers    *SearchHandlers
	validate          *validator.Validate
	DB                *sql.DB
}

type jsonResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
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
	bookCache repository.BookCache,
	bookDeleter repository.BookDeleter,
	bookUpdater services.BookUpdaterService,
	bookService services.BookService,
	exportService services.ExportService,
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

	return &BookHandlers{
		DB:                db,
		logger:            logger,
		bookModels:        bookModels,
		authorRepo:        authorRepo,
		bookRepo:          bookRepo,
		formatRepo:        formatRepo,
		genreRepo:         genreRepo,
		tagRepo:           tagRepo,
		bookCache:         bookCache,
		bookDeleter:       bookDeleter,
		bookService:       bookService,
		bookUpdater:       bookUpdater,
		exportService:     exportService,
		exportLimiter:     rate.NewLimiter(rate.Limit(1), 3),
		validate:          validate,
		sanitizer:         sanitizer,
	}, nil
}
