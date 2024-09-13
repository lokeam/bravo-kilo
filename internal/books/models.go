package books

import (
	"database/sql"
	"log/slog"
	"time"

	"github.com/lokeam/bravo-kilo/internal/books/repository"
	"github.com/lokeam/bravo-kilo/internal/shared/transaction"
)

const dbTimeout = time.Second * 3

type Models struct {
	Book         repository.BookRepository
	Author       repository.AuthorRepository
	Genre        repository.GenreRepository
	Format       repository.FormatRepository
	Tag          repository.TagRepository
	Transaction  transaction.DBManager
}

func New(db *sql.DB, logger *slog.Logger) (Models, error) {
	models := Models{
		Book: &repository.BookRepositoryImpl{
			DB: db, Logger: logger,
		},
		Author: &repository.AuthorRepositoryImpl{
			DB: db, Logger: logger,
		},
		Genre: &repository.GenreRepositoryImpl{
			DB: db, Logger: logger,
		},
		Format: &repository.FormatRepositoryImpl{
			DB: db, Logger: logger,
		},
		Tag: &repository.TagRepositoryImpl{
			DB: db, Logger: logger,
		},
		Transaction: &transaction.DBManagerImpl{
			DB: db, Logger: logger,
		},
	}

	// Initialize prepared statements for all repositories that need them
	if err := models.Book.InitPreparedStatements(); err != nil {
		logger.Error("Error initializing prepared statements for BookRepository", "error", err)
		return Models{}, err
	}
	if err := models.Author.InitPreparedStatements(); err != nil {
		logger.Error("Error initializing prepared statements for AuthorRepository", "error", err)
		return Models{}, err
	}
	if err := models.Genre.InitPreparedStatements(); err != nil {
		logger.Error("Error initializing prepared statements for GenreRepository", "error", err)
		return Models{}, err
	}
	if err := models.Format.InitPreparedStatements(); err != nil {
		logger.Error("Error initializing prepared statements for FormatRepository", "error", err)
		return Models{}, err
	}
	if err := models.Tag.InitPreparedStatements(); err != nil {
		logger.Error("Error initializing prepared statements for TagRepository", "error", err)
		return Models{}, err
	}

	// Init additional repositories if needed

	return models, nil
}