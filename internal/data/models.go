package data

import (
	"database/sql"
	"log/slog"
	"time"
)

const dbTimeout = time.Second * 3

type Models struct {
	User         UserModel
	Token        TokenModel
	Book         BookRepository
	Author       AuthorRepository
	Genre        GenreRepository
	Format       FormatRepository
	Tag          TagRepository
	Transaction  DBManager
}

func New(db *sql.DB, logger *slog.Logger) (Models, error) {
	models := Models{
		User: UserModel{
			DB: db, Logger: logger,
		},
		Token: TokenModel{
			DB: db, Logger: logger,
		},
		Book: &BookRepositoryImpl{
			DB: db, Logger: logger,
		},
		Author: &AuthorRepositoryImpl{
			DB: db, Logger: logger,
		},
		Genre: &GenreRepositoryImpl{
			DB: db, Logger: logger,
		},
		Format: &FormatRepositoryImpl{
			DB: db, Logger: logger,
		},
		Tag: &TagRepositoryImpl{
			DB: db, Logger: logger,
		},
		Transaction: &DBManagerImpl{
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
