package init

import (
	"database/sql"
	"log/slog"
	"os"

	"github.com/lokeam/bravo-kilo/internal/books"
	"github.com/lokeam/bravo-kilo/internal/books/handlers"
	"github.com/lokeam/bravo-kilo/internal/books/repository"
	"github.com/lokeam/bravo-kilo/internal/books/services"
	auth "github.com/lokeam/bravo-kilo/internal/shared/handlers/auth"
	"github.com/lokeam/bravo-kilo/internal/shared/models"
	"github.com/lokeam/bravo-kilo/internal/shared/transaction"
)

// Factory initializes all components and returns them
type Factory struct {
	BookHandlers    *handlers.BookHandlers
	SearchHandlers  *handlers.SearchHandlers
	AuthHandlers    *auth.AuthHandlers
}

// NewFactory initializes repositories, services, and handlers
func NewFactory(db *sql.DB, log *slog.Logger) (*Factory, error) {
	// Initialize repositories
	authorRepo, err := repository.NewAuthorRepository(db, log)
	if err != nil {
		log.Error("Error initializing author repository", "error", err)
		return nil, err
	}

	genreRepo, err := repository.NewGenreRepository(db, log)
	if err != nil {
		log.Error("Error initializing genre repository", "error", err)
		return nil, err
	}

	formatRepo, err := repository.NewFormatRepository(db, log)
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

	bookCache, err := repository.NewBookCache(db, log)
	if err != nil {
		log.Error("Error initializing book cache", "error", err)
		return nil, err
	}

	bookDeleter, err := repository.NewBookDeleter(db, log)
	if err != nil {
		log.Error("Error initializing book deleter", "error", err)
		return nil, err
	}

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

	// Initialize models
	bookModels, err := books.New(db, log)
	if err != nil {
		return nil, err
	}

	// Initialize models for auth
	authModels := models.Models{
		User:  &models.UserModel{DB: db, Logger: log},
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
		bookCache,
		bookDeleter,
		bookUpdaterService,
		bookService,
		exportService,
	)
	if err != nil {
		return nil, err
	}

	authHandlers, err := auth.NewAuthHandlers(log, authModels, transactionManager)
	if err != nil {
		return nil, err
	}

	searchHandlers, err := handlers.NewSearchHandlers(log, bookRepo, bookCache, authHandlers)
	if err != nil {
		return nil, err
	}

	// Return all handlers and services inside the Factory
	return &Factory{
		BookHandlers: bookHandlers,
		AuthHandlers: authHandlers,
		SearchHandlers: searchHandlers,
	}, nil
}
