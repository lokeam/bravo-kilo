package data

import (
	"context"
	"database/sql"
	"log/slog"
)

type BookRepository interface {
	BeginTransaction(ctx context.Context) (*sql.Tx, error)
	GetBookByID(id int) (*Book, error)
	GetBookIdByTitle(title string) (int, error)
	GetAllBooksByUserID(userID int) ([]Book, error)
	GetBooksByAuthor(authorName string) ([]Book, error)
	GetAllBooksByAuthors(userID int) (map[string]interface{}, error)
	GetAuthorsForBook(bookID int) ([]string, error)
	GetAllBooksByFormat(userID int) (map[string][]Book, error)
	GetOrInsertFormat(ctx context.Context, formatType string) (int, error)
	GetAllBooksByGenres(ctx context.Context, userID int) (map[string]interface{}, error)
}

type BookRepositoryImpl struct {
	DB                          *sql.DB
	Logger                      *slog.Logger
	insertBookStmt              *sql.Stmt
	getBookByIDStmt             *sql.Stmt
	addBookToUserStmt           *sql.Stmt
	getBookIdByTitleStmt        *sql.Stmt
	getAuthorsForBooksStmt      *sql.Stmt
	isUserBookOwnerStmt         *sql.Stmt
	getAllBooksByUserIDStmt     *sql.Stmt
	getAllBooksByAuthorsStmt    *sql.Stmt
	getAllBooksByGenresStmt     *sql.Stmt
	getAllBooksByFormatStmt     *sql.Stmt
	getFormatsStmt              *sql.Stmt
	getGenresStmt               *sql.Stmt
	getAllLangStmt              *sql.Stmt
	getBookListByGenreStmt      *sql.Stmt
	getUserTagsStmt             *sql.Stmt
}
