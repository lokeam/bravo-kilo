package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/lokeam/bravo-kilo/internal/dbconfig"
	"github.com/lokeam/bravo-kilo/internal/shared/collections"
)

var isbn10Cache         sync.Map
var isbn13Cache         sync.Map
var titleCache          sync.Map
var formatsCache        sync.Map
var genresCache         sync.Map
var booksByLangCache    sync.Map
var booksByGenresCache  sync.Map
var userTagsCache       sync.Map

type BookCache interface {
	InitPreparedStatements() error
	GetAllBooksISBN10(userID int) (*collections.Set, error)
	GetAllBooksISBN13(userID int) (*collections.Set, error)
	GetAllBooksTitles(userID int) (*collections.Set, error)
	GetAllBooksPublishDate(userID int) ([]BookInfo, error)
	GetBooksByLanguage(ctx context.Context, userID int) (map[string]interface{}, error)
	InvalidateCaches(bookID int)
}

type BookCacheImpl struct {
	DB               *sql.DB
	Logger           *slog.Logger
	getAllLangStmt   *sql.Stmt
}

func NewBookCache(db *sql.DB, logger *slog.Logger) (BookCache, error) {
	if db == nil || logger == nil {
		return nil, fmt.Errorf("book cache, database or logger is nil")
	}

	return &BookCacheImpl{
		DB:              db,
		Logger:          logger,
	}, nil
}

func (b *BookCacheImpl) InitPreparedStatements() error {
	var err error

	// Prepared statment for GetLanguages
	b.getAllLangStmt, err = b.DB.Prepare(`
	SELECT language, COUNT(*) AS total
		FROM books
		INNER JOIN user_books ub ON books.id = ub.book_id
		WHERE ub.user_id = $1
		GROUP BY language
		ORDER BY total DESC`)
	if err != nil {
		return err
	}

	return nil
}

// ISBN10 + ISBN13 (Returns a HashSet)
func (b *BookCacheImpl) GetAllBooksISBN10(userID int) (*collections.Set, error) {
	// Check cache
	if cache, found := isbn10Cache.Load(userID); found {
		b.Logger.Info("Fetching ISBN10 from cache")
		return cache.(*collections.Set), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	query := `
	SELECT b.isbn_10
	FROM books b
	INNER JOIN user_books ub ON b.id = ub.book_id
	WHERE ub.user_id = $1`

	rows, err := b.DB.QueryContext(ctx, query, userID)
	if err != nil {
		b.Logger.Error("Error retrieving ISBN10 numbers", "error", err)
		return nil, err
	}
	defer rows.Close()

	isbnSet := collections.NewSet()

	for rows.Next() {
		var isbn10 string
		if err := rows.Scan(&isbn10); err != nil {
			b.Logger.Error("Error scanning ISBN10", "error", err)
			return nil, err
		}
		isbnSet.Add(isbn10)
	}

	// Checking for errors after scanning all rows
	if err = rows.Err(); err != nil {
		b.Logger.Error("Error with rows", "error", err)
		return nil, err
	}

	// Cache the result
	isbn10Cache.Store(userID, isbnSet)
	b.Logger.Info("Caching ISBN10 for user", "userID", userID)

	return isbnSet, nil
}

// (Returns a HashSet)
func (b *BookCacheImpl) GetAllBooksISBN13(userID int) (*collections.Set, error) {
	// Check cache
	if cache, found := isbn13Cache.Load(userID); found {
		b.Logger.Info("Fetching ISBN13 from cache")
		return cache.(*collections.Set), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	query := `
	SELECT b.isbn_13
	FROM books b
	INNER JOIN user_books ub ON b.id = ub.book_id
	WHERE ub.user_id = $1`

	rows, err := b.DB.QueryContext(ctx, query, userID)
	if err != nil {
		b.Logger.Error("Error retrieving ISBN13 numbers", "error", err)
		return nil, err
	}
	defer rows.Close()

	isbnSet := collections.NewSet()

	for rows.Next() {
		var isbn13 string
		if err := rows.Scan(&isbn13); err != nil {
			b.Logger.Error("Error scanning ISBN13", "error", err)
			return nil, err
		}
		isbnSet.Add(isbn13)
	}

	// Checking for errors after scanning all rows
	if err = rows.Err(); err != nil {
		b.Logger.Error("Error with rows", "error", err)
		return nil, err
	}

	// Cache the result
	isbn13Cache.Store(userID, isbnSet)
	b.Logger.Info("Caching ISBN13 for user", "userID", userID)

	return isbnSet, nil
}

// (Returns a HashSet)
func (b *BookCacheImpl) GetAllBooksTitles(userID int) (*collections.Set, error) {
	// Check cache
	if cache, found := titleCache.Load(userID); found {
		b.Logger.Info("Fetching Title info from cache")
		return cache.(*collections.Set), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	query := `
	SELECT b.title
	FROM books b
	INNER JOIN user_books ub ON b.id = ub.book_id
	WHERE ub.user_id = $1`

	rows, err := b.DB.QueryContext(ctx, query, userID)
	if err != nil {
		b.Logger.Error("Error retrieving book titles", "error", err)
		return nil, err
	}
	defer rows.Close()

	titleSet := collections.NewSet()

	for rows.Next() {
		var bookTitle string
		if err := rows.Scan(&bookTitle); err != nil {
			b.Logger.Error("Error scanning book title", "error", err)
			return nil, err
		}
		titleSet.Add(bookTitle)
	}

	if err = rows.Err(); err != nil {
		b.Logger.Error("Error with rows", "error", err)
		return nil, err
	}

	// Cache result
	titleCache.Store(userID, titleSet)
	b.Logger.Info("Caching Title info for user", "userID", userID)

	return titleSet, nil
}

// (Return a Slice of BookInfo Structs to handle books with duplicate titles)
func (b *BookCacheImpl) GetAllBooksPublishDate(userID int) ([]BookInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	query := `
	SELECT b.title, b.publish_date
	FROM books b
	INNER JOIN user_books ub ON b.id = ub.book_id
	WHERE ub.user_id = $1`

	rows, err := b.DB.QueryContext(ctx, query, userID)
	if err != nil {
			b.Logger.Error("Error retrieving book titles and publish dates", "error", err)
			return nil, err
	}
	defer rows.Close()

	var books []BookInfo

	for rows.Next() {
			var book BookInfo
			var publishDate time.Time
			if err := rows.Scan(&book.Title, &publishDate); err != nil {
					b.Logger.Error("Error scanning book title and publish date", "error", err)
					return nil, err
			}

			// Format publish date to "YYYY-MM-DD"
			book.PublishDate = publishDate.Format("2006-01-02")
			books = append(books, book)
	}

	if err = rows.Err(); err != nil {
			b.Logger.Error("Error with rows", "error", err)
			return nil, err
	}

	return books, nil
}

// Languages
func (b *BookCacheImpl) GetBooksByLanguage(ctx context.Context, userID int) (map[string]interface{}, error) {
	// Check cache
	cacheKey := fmt.Sprintf("booksByLang:%d", userID)
	if cache, found := booksByLangCache.Load(cacheKey); found {
			b.Logger.Info("Fetching books by language from cache", "userID", userID)
			return cache.(map[string]interface{}), nil
	}

	// Use prepared statement if available
	var rows *sql.Rows
	var err error

	if b.getAllLangStmt != nil {
			b.Logger.Info("Using prepared statement for fetching books by language")
			rows, err = b.getAllLangStmt.QueryContext(ctx, userID)
	} else {
			b.Logger.Warn("Prepared statement for fetching books by language is unavailable. Falling back to raw query")
			query := `
			SELECT language, COUNT(*) AS total
			FROM books
			INNER JOIN user_books ub ON books.id = ub.book_id
			WHERE ub.user_id = $1
			GROUP BY language
			ORDER BY total DESC`
			rows, err = b.DB.QueryContext(ctx, query, userID)
	}

	if err != nil {
			b.Logger.Error("Error fetching books by language", "error", err)
			return nil, err
	}
	defer rows.Close()

	var booksByLang []map[string]interface{}

	// Process the results
	for rows.Next() {
			var language string
			var total int

			if err := rows.Scan(&language, &total); err != nil {
					b.Logger.Error("Error scanning books by language", "error", err)
					return nil, err
			}

			booksByLang = append(booksByLang, map[string]interface{}{
					"label": language,
					"count": total,
			})
	}

	if err = rows.Err(); err != nil {
			b.Logger.Error("Error finalizing rows", "error", err)
			return nil, err
	}

	result := map[string]interface{}{
			"booksByLang": booksByLang,
	}

	// Cache the result
	booksByLangCache.Store(cacheKey, result)
	b.Logger.Info("Caching books by language", "userID", userID)

	return result, nil
}

// Invalidate Caches
func (b *BookCacheImpl) InvalidateCaches(bookID int) {
	isbn10Cache.Delete(bookID)
	isbn13Cache.Delete(bookID)
	titleCache.Delete(bookID)
	formatsCache.Delete(bookID)
	genresCache.Delete(bookID)
	b.Logger.Info("Caches invalidated for book", "bookID", bookID)
}