package data

import (
	"bravo-kilo/internal/data/collections"
	"bravo-kilo/internal/data/customheap"
	"bravo-kilo/internal/utils"
	"container/heap"
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/lib/pq"
)

type BookModel struct {
	DB                       *sql.DB
	Logger                   *slog.Logger
	Author                   *AuthorModel
	insertBookStmt           *sql.Stmt
	getBookByIDStmt          *sql.Stmt
	addBookToUserStmt        *sql.Stmt
	getBookIdByTitleStmt     *sql.Stmt
	getAuthorsForBooksStmt   *sql.Stmt
	isUserBookOwnerStmt      *sql.Stmt
	getAllBooksByUserIDStmt  *sql.Stmt
	getAllBooksByAuthorsStmt *sql.Stmt
	getAllBooksByGenresStmt  *sql.Stmt
	getAllBooksByFormatStmt  *sql.Stmt
	getFormatsStmt           *sql.Stmt
	getGenresStmt            *sql.Stmt
	getAllLangStmt           *sql.Stmt
	getBookListByGenreStmt   *sql.Stmt
	getUserTagsStmt          *sql.Stmt
}

type BookInfo struct {
	Title       string
	PublishDate string
}

type Book struct {
	ID              int        `json:"id"`
	Title           string     `json:"title"`
	Subtitle        string     `json:"subtitle"`
	Description     string     `json:"description"`
	Language        string     `json:"language"`
	PageCount       int        `json:"pageCount"`
	PublishDate     string     `json:"publishDate"`
	Authors         []string   `json:"authors"`
	ImageLink       string     `json:"imageLink"`
	Genres          []string   `json:"genres"`
	Notes           string     `json:"notes"`
	Formats         []string   `json:"formats"`
	Tags            []string   `json:"tags"`
	CreatedAt       time.Time  `json:"created_at"`
	LastUpdated     time.Time  `json:"lastUpdated"`
	ISBN10          string     `json:"isbn10"`
	ISBN13          string     `json:"isbn13"`
	IsInLibrary     bool       `json:"isInLibrary"`
	HasEmptyFields  bool       `json:"hasEmptyFields"`
	EmptyFields     []string   `json:"emptyFields"`
}

type UserTagsCacheEntry struct {
	data      map[string]interface{}
	timestamp time.Time
}

type BooksByGenresCacheEntry struct {
	data      map[string]interface{}
	timestamp time.Time
}

var isbn10Cache         sync.Map
var isbn13Cache         sync.Map
var titleCache          sync.Map
var formatsCache        sync.Map
var genresCache         sync.Map
var booksByLangCache    sync.Map
var booksByGenresCache  sync.Map
var userTagsCache       sync.Map

// Prepared Statements
func (b *BookModel) InitPreparedStatements() error {
	var err error

	// Prepared insert statement for books
	b.insertBookStmt, err = b.DB.Prepare(`
		INSERT INTO books (title, subtitle, description, language, page_count, publish_date, image_link, notes, tags, created_at, last_updated, isbn_10, isbn_13)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING id`)
	if err != nil {
		return err
	}

	// Prepared select statement for GetBookByID
	b.getBookByIDStmt, err = b.DB.Prepare(`
		WITH book_data AS (
			SELECT
				b.id, b.title, b.subtitle, b.description, b.language, b.page_count, b.publish_date,
				b.image_link, b.notes, b.tags, b.created_at, b.last_updated, b.isbn_10, b.isbn_13
			FROM books b
			WHERE b.id = $1
		),
		authors_data AS (
			SELECT a.name
			FROM authors a
			JOIN book_authors ba ON a.id = ba.author_id
			WHERE ba.book_id = $1
		),
		genres_data AS (
			SELECT g.name
			FROM genres g
			JOIN book_genres bg ON g.id = bg.genre_id
			WHERE bg.book_id = $1
		),
		formats_data AS (
			SELECT f.format_type
			FROM formats f
			JOIN book_formats bf ON f.id = bf.format_id
			WHERE bf.book_id = $1
		)
		SELECT
			b.id, b.title, b.subtitle, b.description, b.language, b.page_count, b.publish_date,
			b.image_link, b.notes, b.tags, b.created_at, b.last_updated, b.isbn_10, b.isbn_13,
			a.name AS author_name, g.name AS genre_name, f.format_type AS format_type
		FROM book_data b
		LEFT JOIN authors_data a ON true
		LEFT JOIN genres_data g ON true
		LEFT JOIN formats_data f ON true`)
	if err != nil {
		return err
	}

	// Prepared insert statement for adding books to user
	b.addBookToUserStmt, err = b.DB.Prepare(`INSERT INTO user_books (user_id, book_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`)
	if err != nil {
		return err
	}

	// Prepared select statement for getting book ID by title
	b.getBookIdByTitleStmt, err = b.DB.Prepare(`SELECT id FROM books WHERE title = $1`)
	if err != nil {
		return err
	}

	// Prepared select statement for GetAuthorsForBooks
	b.getAuthorsForBooksStmt, err = b.DB.Prepare(`
		SELECT ba.book_id, a.name
		FROM authors a
		JOIN book_authors ba ON a.id = ba.author_id
		WHERE ba.book_id = ANY($1)`)
	if err != nil {
		return err
	}

	// Prepared select statement for IsUserBookOwner
	b.isUserBookOwnerStmt, err = b.DB.Prepare(`
		SELECT EXISTS(SELECT 1 FROM user_books WHERE user_id = $1 AND book_id = $2)`)
	if err != nil {
		return err
	}

	// Prepared select statement for GetAllBooksByUserID
	b.getAllBooksByUserIDStmt, err = b.DB.Prepare(`
		SELECT b.id, b.title, b.subtitle, b.description, b.language, b.page_count, b.publish_date,
					 b.image_link, b.notes, b.created_at, b.last_updated, b.isbn_10, b.isbn_13
		FROM books b
		INNER JOIN user_books ub ON b.id = ub.book_id
		WHERE ub.user_id = $1`)
	if err != nil {
		return err
	}

	// Prepared select statement for GetAllBooksByAuthors
	b.getAllBooksByAuthorsStmt, err = b.DB.Prepare(`
	SELECT b.id, b.title, b.subtitle, b.description, b.language, b.page_count, b.publish_date,
				 b.image_link, b.notes, b.created_at, b.last_updated, b.isbn_10, b.isbn_13,
				 a.name AS author_name,
				 json_agg(DISTINCT g.name) AS genres,
				 json_agg(DISTINCT f.format_type) AS formats,
				 b.tags
	FROM books b
	INNER JOIN book_authors ba ON b.id = ba.book_id
	INNER JOIN authors a ON ba.author_id = a.id
	INNER JOIN user_books ub ON b.id = ub.book_id
	LEFT JOIN book_genres bg ON b.id = bg.book_id
	LEFT JOIN genres g ON bg.genre_id = g.id
	LEFT JOIN book_formats bf ON b.id = bf.book_id
	LEFT JOIN formats f ON bf.format_id = f.id
	WHERE ub.user_id = $1::integer  -- Explicitly cast the user_id to integer
	GROUP BY b.id, a.name`)
	if err != nil {
		return err
	}

	// Prepared statment for GetAllBooksByGenres
	b.getAllBooksByGenresStmt, err = b.DB.Prepare(`
	SELECT b.id, b.title, b.subtitle, b.description, b.language, b.page_count, b.publish_date,
								 b.image_link, b.notes, b.created_at, b.last_updated, b.isbn_10, b.isbn_13,
								 json_agg(DISTINCT g.name) AS genres,
								 json_agg(DISTINCT a.name) AS authors,
								 json_agg(DISTINCT f.format_type) AS formats,
								 b.tags
					FROM books b
					INNER JOIN user_books ub ON b.id = ub.book_id
					LEFT JOIN book_genres bg ON b.id = bg.book_id
					LEFT JOIN genres g ON bg.genre_id = g.id
					LEFT JOIN book_authors ba ON b.id = ba.book_id
					LEFT JOIN authors a ON ba.author_id = a.id
					LEFT JOIN book_formats bf ON b.id = bf.book_id
					LEFT JOIN formats f ON bf.format_id = f.id
					WHERE ub.user_id = $1
					GROUP BY b.id`)
	if err != nil {
		return err
	}

	// Prepared statement for GetAllBooksByFormat
	b.getAllBooksByFormatStmt, err = b.DB.Prepare(`
	SELECT
		b.id, b.title, b.subtitle, b.description, b.language, b.page_count, b.publish_date,
		b.image_link, b.notes, b.created_at, b.last_updated, b.isbn_10, b.isbn_13,
		f.format_type,
		array_to_json(array_agg(DISTINCT a.name)) as authors,
		array_to_json(array_agg(DISTINCT g.name)) as genres,
		b.tags
	FROM books b
	INNER JOIN book_formats bf ON b.id = bf.book_id
	INNER JOIN formats f ON bf.format_id = f.id
	INNER JOIN user_books ub ON b.id = ub.book_id
	LEFT JOIN book_authors ba ON b.id = ba.book_id
	LEFT JOIN authors a ON ba.author_id = a.id
	LEFT JOIN book_genres bg ON b.id = bg.book_id
	LEFT JOIN genres g ON bg.genre_id = g.id
	WHERE ub.user_id = $1
	GROUP BY b.id, f.format_type`)
	if err != nil {
	return err
	}

	// Prepared select statement for GetFormats
	b.getFormatsStmt, err = b.DB.Prepare(`
		SELECT f.format_type
		FROM formats f
		JOIN book_formats bf ON f.id = bf.format_id
		WHERE bf.book_id = $1`)
	if err != nil {
		return err
	}

	// Prepared select statement for GetGenres
	b.getGenresStmt, err = b.DB.Prepare(`
		SELECT g.name
		FROM genres g
		JOIN book_genres bg ON g.id = bg.genre_id
		WHERE bg.book_id = $1`)
	if err != nil {
		return err
	}

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

	b.getBookListByGenreStmt, err = b.DB.Prepare(`
	SELECT g.name, COUNT(DISTINCT b.id) AS total_books
		FROM books b
		INNER JOIN book_genres bg ON b.id = bg.book_id
		INNER JOIN genres g ON bg.genre_id = g.id
		INNER JOIN user_books ub ON b.id = ub.book_id
		WHERE ub.user_id = $1
		GROUP BY g.name
		ORDER BY total_books DESC`)
	if err != nil {
		return err
	}

	b.getUserTagsStmt, err = b.DB.Prepare(`
	SELECT b.tags
		FROM books b
		INNER JOIN user_books ub ON b.id = ub.book_id
		WHERE ub.user_id = $1
	`)
	if err != nil {
		return err
	}

	return nil
}

// General Book
func (b *BookModel) InsertBook(book Book, userID int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	tx, err := b.DB.BeginTx(ctx, nil)
	if err != nil {
		b.Logger.Error("Error beginning transaction", "error", err)
		return 0, err
	}
	defer tx.Rollback()

	var newId int

	tagsJSON, err := json.Marshal(book.Tags)
	if err != nil {
		b.Logger.Error("Error marshalling tags to JSON", "error", err)
		return 0, err
	}

	// Format publish date in case Google Books only returns year
	formattedPublishDate := formatPublishDate(book.PublishDate)

	// Check if prepared statement is available
	if b.insertBookStmt != nil {
		b.Logger.Info("Using prepared statement for inserting a book")
		err = tx.StmtContext(ctx, b.insertBookStmt).QueryRowContext(ctx,
			book.Title,
			book.Subtitle,
			book.Description,
			book.Language,
			book.PageCount,
			formattedPublishDate,
			book.ImageLink,
			book.Notes,
			tagsJSON,
			time.Now(),
			time.Now(),
			book.ISBN10,
			book.ISBN13,
		).Scan(&newId)
	} else {
		// Fallback to using raw SQL query if prepared statement unavailable
		b.Logger.Warn("Prepared statement for inserting a book is not available. Falling back to raw SQL query")
		statement := `INSERT INTO books (title, subtitle, description, language, page_count, publish_date, image_link, notes, tags, created_at, last_updated, isbn_10, isbn_13)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING id`
		err = tx.QueryRowContext(ctx, statement,
			book.Title,
			book.Subtitle,
			book.Description,
			book.Language,
			book.PageCount,
			formattedPublishDate,
			book.ImageLink,
			book.Notes,
			tagsJSON,
			time.Now(),
			time.Now(),
			book.ISBN10,
			book.ISBN13,
		).Scan(&newId)
	}

	if err != nil {
		b.Logger.Error("Book Model - Error inserting book", "error", err)
		return 0, err
	}

	// Insert authors + genres
	authorsSet := collections.NewSet()
	genresSet := collections.NewSet()

	b.Logger.Info("Authors received: ", "authors", book.Authors)

	for _, author := range book.Authors {
		authorsSet.Add(author)
	}

	for _, genre := range book.Genres {
		genresSet.Add(genre)
	}

	// Batch insert authors, associate them with the book
	// Batch insert authors, associate them with the book
	for _, author := range authorsSet.Elements() {
		if author == "" {
				b.Logger.Error("Author name is empty, skipping insertion")
				continue
		}

		b.Logger.Info("Inserting/Querying author", "author", author)

		var authorID int
		err = tx.QueryRowContext(ctx, `SELECT id FROM authors WHERE name = $1`, author).Scan(&authorID)
		if err != nil {
				if err == sql.ErrNoRows {
						// Author doesn't exist, so insert it
						b.Logger.Info("Author not found, inserting new author", "author", author)
						err = tx.QueryRowContext(ctx, `INSERT INTO authors (name) VALUES ($1) RETURNING id`, author).Scan(&authorID)
						if err != nil {
								b.Logger.Error("Error inserting author", "error", err, "author", author)
								return 0, err
						}
				} else {
						b.Logger.Error("Error querying author", "error", err, "author", author)
						return 0, err
				}
		}

		// Log successful author retrieval/insertion
		b.Logger.Info("Successfully retrieved or inserted author", "authorID", authorID, "author", author)

		// Insert the book-author association
		_, err = tx.ExecContext(ctx, `INSERT INTO book_authors (book_id, author_id) VALUES ($1, $2)`, newId, authorID)
		if err != nil {
				b.Logger.Error("Error inserting book_author association", "error", err, "bookID", newId, "authorID", authorID)
				return 0, err
		}

		b.Logger.Info("Successfully inserted book-author association", "bookID", newId, "authorID", authorID)
	}


	// Batch insert genres, associate them with the book
	for _, genre := range genresSet.Elements() {
		var genreID int
		err = tx.QueryRowContext(ctx, `SELECT id FROM genres WHERE name = $1`, genre).Scan(&genreID)
		if err != nil {
			if err == sql.ErrNoRows {
				err = tx.QueryRowContext(ctx, `INSERT INTO genres (name) VALUES ($1) RETURNING id`, genre).Scan(&genreID)
				if err != nil {
					b.Logger.Error("Error inserting genre", "error", err)
					return 0, err
				}
			} else {
				b.Logger.Error("Error querying genre", "error", err)
				return 0, err
			}
		}

		_, err = tx.ExecContext(ctx, `INSERT INTO book_genres (book_id, genre_id) VALUES ($1, $2)`, newId, genreID)
		if err != nil {
			b.Logger.Error("Error inserting book_genre association", "error", err)
			return 0, err
		}
	}

	// Associate book with user
	if err := b.AddBookToUser(tx, userID, newId); err != nil {
		b.Logger.Error("Error adding book to user", "error", err)
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		b.Logger.Error("Error committing transaction for book", "bookID", newId, "error", err)
		return 0, err
	}

	return newId, nil
}

// Helper fn function for Insert Book to format publish date
func formatPublishDate(dateStr string) string {
	// If publish date only lists year, append "-01-01"
	if len(dateStr) == 4 {
		return dateStr + "-01-01"
	}
	return dateStr
}


func (b *BookModel) AddBookToUser(tx *sql.Tx, userID, bookID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// Check if prepared statement is available
	if b.addBookToUserStmt != nil {
		b.Logger.Info("Using prepared statement for adding a book to user")
		_, err := tx.StmtContext(ctx, b.addBookToUserStmt).ExecContext(ctx, userID, bookID)
		if err != nil {
			b.Logger.Error("Error adding book to user using prepared statement", "error", err)
			return err
		}
	} else {
		// Fallback to using raw SQL query if prepared statement is unavailable
		b.Logger.Warn("Prepared statement for adding a book to user is not available. Falling back to raw SQL query")
		statement := `INSERT INTO user_books (user_id, book_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
		_, err := tx.ExecContext(ctx, statement, userID, bookID)
		if err != nil {
			b.Logger.Error("Error adding book to user using raw SQL query", "error", err)
			return err
		}
	}

	return nil
}

func (b *BookModel) GetBookByID(id int) (*Book, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// Check if prepared statement is available
	var rows *sql.Rows
	var err error

	if b.getBookByIDStmt != nil {
		b.Logger.Info("Using prepared statement for fetching book by ID")
		rows, err = b.getBookByIDStmt.QueryContext(ctx, id)
	} else {
		// Fallback to raw SQL query if prepared statement is unavailable
		b.Logger.Warn("Prepared statement for fetching book by ID is not available. Falling back to raw SQL query")
		query := `
		WITH book_data AS (
			SELECT
				b.id, b.title, b.subtitle, b.description, b.language, b.page_count, b.publish_date,
				b.image_link, b.notes, b.tags, b.created_at, b.last_updated, b.isbn_10, b.isbn_13
			FROM books b
			WHERE b.id = $1
		),
		authors_data AS (
			SELECT a.name
			FROM authors a
			JOIN book_authors ba ON a.id = ba.author_id
			WHERE ba.book_id = $1
		),
		genres_data AS (
			SELECT g.name
			FROM genres g
			JOIN book_genres bg ON g.id = bg.genre_id
			WHERE bg.book_id = $1
		),
		formats_data AS (
			SELECT f.format_type
			FROM formats f
			JOIN book_formats bf ON f.id = bf.format_id
			WHERE bf.book_id = $1
		)
		SELECT
			b.id, b.title, b.subtitle, b.description, b.language, b.page_count, b.publish_date,
			b.image_link, b.notes, b.tags, b.created_at, b.last_updated, b.isbn_10, b.isbn_13,
			a.name AS author_name, g.name AS genre_name, f.format_type AS format_type
		FROM book_data b
		LEFT JOIN authors_data a ON true
		LEFT JOIN genres_data g ON true
		LEFT JOIN formats_data f ON true
		`
		rows, err = b.DB.QueryContext(ctx, query, id)
	}

	if err != nil {
		b.Logger.Error("Error fetching book by ID", "error", err)
		return nil, err
	}
	defer rows.Close()

	var book Book
	var tagsJSON []byte
	authorsSet := collections.NewSet()
	genresSet := collections.NewSet()
	formatsSet := collections.NewSet()

	for rows.Next() {
		var authorName, genreName, formatType sql.NullString

		if err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.Subtitle,
			&book.Description,
			&book.Language,
			&book.PageCount,
			&book.PublishDate,
			&book.ImageLink,
			&book.Notes,
			&tagsJSON,
			&book.CreatedAt,
			&book.LastUpdated,
			&book.ISBN10,
			&book.ISBN13,
			&authorName,
			&genreName,
			&formatType,
		); err != nil {
			b.Logger.Error("Error scanning book with batch queries", "error", err)
			return nil, err
		}

		if authorName.Valid {
			authorsSet.Add(authorName.String)
		}
		if genreName.Valid {
			genresSet.Add(genreName.String)
		}
		if formatType.Valid {
			formatsSet.Add(formatType.String)
		}
	}

	// Unmarshal JSON field
	if err := json.Unmarshal(tagsJSON, &book.Tags); err != nil {
		b.Logger.Error("Error unmarshalling tags JSON", "error", err)
		return nil, err
	}

	// Convert sets to slices
	book.Authors = authorsSet.Elements()
	book.Genres = genresSet.Elements()
	book.Formats = formatsSet.Elements()

	book.IsInLibrary = true
	return &book, nil
}

func (b *BookModel) GetBookIdByTitle(title string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var bookID int
	var err error

	if b.getBookIdByTitleStmt != nil {
		b.Logger.Info("Using prepared statement for fetching book ID by title")
		err = b.getBookIdByTitleStmt.QueryRowContext(ctx, title).Scan(&bookID)
	} else {
		// Fallback to raw SQL query if prepared statement is unavailable
		b.Logger.Warn("Prepared statement for fetching book ID by title is not available. Falling back to raw SQL query")
		statement := `SELECT id FROM books WHERE title = $1`
		err = b.DB.QueryRowContext(ctx, statement, title).Scan(&bookID)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			b.Logger.Warn("No book found with the given title", "title", title)
			return 0, nil
		}
		b.Logger.Error("Error fetching book ID by title", "error", err)
		return 0, err
	}

	return bookID, nil
}

func (b *BookModel) GetAuthorsForBooks(bookIDs []int) (map[int][]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var rows *sql.Rows
	var err error

	// Use prepared statement if available
	if b.getAuthorsForBooksStmt != nil {
		rows, err = b.getAuthorsForBooksStmt.QueryContext(ctx, pq.Array(bookIDs))
		if err != nil {
			b.Logger.Error("Error fetching authors for books using prepared statement", "error", err)
			return nil, err
		}
	} else {
		// Fallback if the prepared statement is unavailable
		query := `
		SELECT ba.book_id, a.name
		FROM authors a
		JOIN book_authors ba ON a.id = ba.author_id
		WHERE ba.book_id = ANY($1)`

		rows, err = b.DB.QueryContext(ctx, query, pq.Array(bookIDs))
		if err != nil {
			b.Logger.Error("Error fetching authors for books using fallback query", "error", err)
			return nil, err
		}
	}
	defer rows.Close()

	// Map to store authors for each bookID
	authorsByBook := make(map[int][]string)
	for rows.Next() {
		var bookID int
		var authorName string
		if err := rows.Scan(&bookID, &authorName); err != nil {
			b.Logger.Error("Error scanning author name", "error", err)
			return nil, err
		}
		authorsByBook[bookID] = append(authorsByBook[bookID], authorName)
	}

	return authorsByBook, nil
}

func (b *BookModel) IsUserBookOwner(userID, bookID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var exists bool
	var err error

	// Use prepared statement if unavailable
	if b.isUserBookOwnerStmt != nil {
		err = b.isUserBookOwnerStmt.QueryRowContext(ctx, userID, bookID).Scan(&exists)
		if err != nil {
			b.Logger.Error("Error checking book ownership using prepared statement", "error", err)
			return false, err
		}
	} else {
		// Fallback if prepared statement is unavailable
		query := `SELECT EXISTS(SELECT 1 FROM user_books WHERE user_id = $1 AND book_id = $2)`
		err = b.DB.QueryRowContext(ctx, query, userID, bookID).Scan(&exists)
		if err != nil {
			b.Logger.Error("Error checking book ownership using fallback query", "error", err)
			return false, err
		}
	}

	return exists, nil
}

func (b *BookModel) GetAllBooksByUserID(userID int) ([]Book, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var rows *sql.Rows
	var err error

	if b.getAllBooksByUserIDStmt != nil {
		b.Logger.Info("Using prepared statement for retrieving books by user ID")
		rows, err = b.getAllBooksByUserIDStmt.QueryContext(ctx, userID)
	} else {
		b.Logger.Warn("Prepared statement for retrieving books by user ID not initialized, using fallback query")
		query := `
			SELECT b.id, b.title, b.subtitle, b.description, b.language, b.page_count, b.publish_date,
						 b.image_link, b.notes, b.created_at, b.last_updated, b.isbn_10, b.isbn_13
			FROM books b
			INNER JOIN user_books ub ON b.id = ub.book_id
			WHERE ub.user_id = $1`
		rows, err = b.DB.QueryContext(ctx, query, userID)
	}

	if err != nil {
		b.Logger.Error("Error retrieving books for user", "error", err)
		return nil, err
	}
	defer rows.Close()

	bookIDMap := make(map[int]*Book)
	var bookIDs []int
	for rows.Next() {
		var book Book

		if err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.Subtitle,
			&book.Description,
			&book.Language,
			&book.PageCount,
			&book.PublishDate,
			&book.ImageLink,
			&book.Notes,
			&book.CreatedAt,
			&book.LastUpdated,
			&book.ISBN10,
			&book.ISBN13,
		); err != nil {
			b.Logger.Error("Error scanning book", "error", err)
			return nil, err
		}

		book.IsInLibrary = true
		bookIDMap[book.ID] = &book
		bookIDs = append(bookIDs, book.ID)
	}

	// Batch Fetch authors, formats, genres, and tags
	if err := b.batchFetchBookDetails(ctx, bookIDs, bookIDMap); err != nil {
		return nil, err
	}

	// Collect books from map into a slice, check for empty fields
	var books []Book
	for _, book := range bookIDMap {
		book.EmptyFields, book.HasEmptyFields = b.findEmptyFields(book)
		books = append(books, *book)
	}

	return books, nil
}

// Helper fn for GetAllBooksByUserID
func (b *BookModel) findEmptyFields(book *Book) ([]string, bool) {
	// Define a slice of field checks
	checks := []struct {
		value interface{}
		name  string
	}{
		{book.Title, "title"},
		{book.Subtitle, "subtitle"},
		{book.Description, "description"},
		{book.Language, "language"},
		{book.PageCount == 0, "pageCount"},
		{book.PublishDate == "", "publishDate"},
		{len(book.Authors) == 0, "authors"},
		{book.ImageLink == "", "imageLink"},
		{len(book.Genres) == 0, "genres"},
		{len(book.Formats) == 0, "formats"},
		{len(book.Tags) == 0, "tags"},
	}

	var emptyFields []string
	hasEmptyFields := false

	// Loop through the checks
	for _, check := range checks {
		// Check for empty fields
		switch v := check.value.(type) {
		case string:
			if v == "" {
				emptyFields = append(emptyFields, check.name)
				hasEmptyFields = true
			}
		case bool:
			if v {
				emptyFields = append(emptyFields, check.name)
				hasEmptyFields = true
			}
		}
	}

	return emptyFields, hasEmptyFields
}


func (b *BookModel) batchFetchBookDetails(ctx context.Context, bookIDs []int, bookIDMap map[int]*Book) error {
	// Batch fetch authors
	authorQuery := `
		SELECT ba.book_id, a.name
		FROM book_authors ba
		INNER JOIN authors a ON ba.author_id = a.id
		WHERE ba.book_id = ANY($1)`

	authorRows, err := b.DB.QueryContext(ctx, authorQuery, pq.Array(bookIDs))
	if err != nil {
		b.Logger.Error("Error retrieving authors for books", "error", err)
		return err
	}
	defer authorRows.Close()

	for authorRows.Next() {
		var bookID int
		var authorName string
		if err := authorRows.Scan(&bookID, &authorName); err != nil {
			b.Logger.Error("Error scanning author", "error", err)
			return err
		}
		bookIDMap[bookID].Authors = append(bookIDMap[bookID].Authors, authorName)
	}

	// Batch fetch genres
	genreQuery := `
		SELECT bg.book_id, g.name
		FROM book_genres bg
		INNER JOIN genres g ON bg.genre_id = g.id
		WHERE bg.book_id = ANY($1)`

	genreRows, err := b.DB.QueryContext(ctx, genreQuery, pq.Array(bookIDs))
	if err != nil {
		b.Logger.Error("Error retrieving genres for books", "error", err)
		return err
	}
	defer genreRows.Close()

	for genreRows.Next() {
		var bookID int
		var genreName string
		if err := genreRows.Scan(&bookID, &genreName); err != nil {
			b.Logger.Error("Error scanning genre", "error", err)
			return err
		}
		bookIDMap[bookID].Genres = append(bookIDMap[bookID].Genres, genreName)
	}

	// Batch fetch formats
	formatQuery := `
		SELECT bf.book_id, f.format_type
		FROM book_formats bf
		INNER JOIN formats f ON bf.format_id = f.id
		WHERE bf.book_id = ANY($1)`

	formatRows, err := b.DB.QueryContext(ctx, formatQuery, pq.Array(bookIDs))
	if err != nil {
		b.Logger.Error("Error retrieving formats for books", "error", err)
		return err
	}
	defer formatRows.Close()

	for formatRows.Next() {
		var bookID int
		var formatType string
		if err := formatRows.Scan(&bookID, &formatType); err != nil {
			b.Logger.Error("Error scanning format", "error", err)
			return err
		}
		bookIDMap[bookID].Formats = append(bookIDMap[bookID].Formats, formatType)
	}

	// Batch fetch tags stored as JSON
	tagQuery := `
		SELECT b.id, b.tags
		FROM books b
		WHERE b.id = ANY($1)`

	tagRows, err := b.DB.QueryContext(ctx, tagQuery, pq.Array(bookIDs))
	if err != nil {
		b.Logger.Error("Error retrieving tags for books", "error", err)
		return err
	}
	defer tagRows.Close()

	for tagRows.Next() {
		var bookID int
		var tagsJSON []byte
		if err := tagRows.Scan(&bookID, &tagsJSON); err != nil {
			b.Logger.Error("Error scanning tags", "error", err)
			return err
		}

		var tags []string
		if err := json.Unmarshal(tagsJSON, &tags); err != nil {
			b.Logger.Error("Error unmarshalling tags JSON", "error", err)
			return err
		}
		bookIDMap[bookID].Tags = tags
	}

	return nil
}

func (b *BookModel) Update(book Book) error {
	// Invalidate caches
	isbn10Cache.Delete(book.ID)
	isbn13Cache.Delete(book.ID)
	titleCache.Delete(book.ID)
	formatsCache.Delete(book.ID)
	genresCache.Delete(book.ID)
	b.Logger.Info("Cache invalidated for book", "book", book.ID)

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	tagsJSON, err := json.Marshal(book.Tags)
	if err != nil {
		b.Logger.Error("Error marshalling tags to JSON", "error", err)
		return err
	}

	statement := `UPDATE books SET title=$1, subtitle=$2, description=$3, language=$4, page_count=$5, publish_date=$6, image_link=$7, notes=$8, tags=$9, last_updated=$10, isbn_10=$11, isbn_13=$12 WHERE id=$13`
	_, err = b.DB.ExecContext(ctx, statement,
		book.Title,
		book.Subtitle,
		book.Description,
		book.Language,
		book.PageCount,
		book.PublishDate,
		book.ImageLink,
		book.Notes,
		tagsJSON,
		time.Now(),
		book.ISBN10,
		book.ISBN13,
		book.ID,
	)
	if err != nil {
		b.Logger.Error("Book Model - Error updating book", "error", err)
		return err
	}

	// Update genres
	if err := b.updateGenres(ctx, book.ID, book.Genres); err != nil {
		return err
	}

	// Update formats
	if err := b.updateFormats(ctx, book.ID, book.Formats); err != nil {
		return err
	}

	// Update authors
	if err := b.updateAuthors(ctx, book.ID, book.Authors); err != nil {
		b.Logger.Error("Book Model - Error updating authors for book", "error", err)
		return err
	}

	return nil
}

func (b *BookModel) Delete(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// Start a new transaction
	tx, err := b.DB.BeginTx(ctx, nil)
	if err != nil {
		b.Logger.Error("Book Model - Error starting transaction", "error", err)
		return err
	}

	// Roll back in case of error
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// Perform all deletions
	if err = b.deleteAssociations(ctx, tx, id); err != nil {
		return err
	}

	// Delete the book
	deleteBookStatement := `DELETE FROM books WHERE id = $1`
	if _, err = tx.ExecContext(ctx, deleteBookStatement, id); err != nil {
		b.Logger.Error("Book Model - Error deleting book", "error", err)
		return err
	}

	return nil
}

// Helper fn for Delete, handles deleting associated records in related tables
func (b *BookModel) deleteAssociations(ctx context.Context, tx *sql.Tx, bookID int) error {
	// Delete associated user_books entries
	deleteUserBookStatement := `DELETE FROM user_books WHERE book_id = $1`
	if _, err := tx.ExecContext(ctx, deleteUserBookStatement, bookID); err != nil {
		b.Logger.Error("Book Model - Error deleting from user_books", "error", err)
		return err
	}

	// Delete associated book_genres entries
	deleteBookGenresStatement := `DELETE FROM book_genres WHERE book_id = $1`
	if _, err := tx.ExecContext(ctx, deleteBookGenresStatement, bookID); err != nil {
		b.Logger.Error("Book Model - Error deleting from book_genres", "error", err)
		return err
	}

	// Delete associated book_authors entries
	deleteBookAuthorsStatement := `DELETE FROM book_authors WHERE book_id = $1`
	if _, err := tx.ExecContext(ctx, deleteBookAuthorsStatement, bookID); err != nil {
		b.Logger.Error("Book Model - Error deleting from book_authors", "error", err)
		return err
	}

	// Delete associated book_formats entries
	deleteBookFormatsStatement := `DELETE FROM book_formats WHERE book_id = $1`
	if _, err := tx.ExecContext(ctx, deleteBookFormatsStatement, bookID); err != nil {
		b.Logger.Error("Book Model - Error deleting from book_formats", "error", err)
		return err
	}

	return nil
}


// Authors
func (b *BookModel) AddAuthor(bookID, authorID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	statement := `INSERT INTO book_authors (book_id, author_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := b.DB.ExecContext(ctx, statement, bookID, authorID)
	if err != nil {
		b.Logger.Error("Error adding author association", "error", err)
		return err
	}

	return nil
}

func (b *BookModel) GetBooksByAuthor(authorName string) ([]Book, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// Fetch all books by the author
	query := `
	SELECT b.id, b.title, b.subtitle, b.description, b.language, b.page_count, b.publish_date,
	       b.image_link, b.notes, b.created_at, b.last_updated, b.isbn_10, b.isbn_13
	FROM books b
	INNER JOIN book_authors ba ON b.id = ba.book_id
	INNER JOIN authors a ON ba.author_id = a.id
	WHERE a.name = $1`

	rows, err := b.DB.QueryContext(ctx, query, authorName)
	if err != nil {
		b.Logger.Error("Error retrieving books by author", "error", err)
		return nil, err
	}
	defer rows.Close()

	var books []Book
	bookIDMap := make(map[int]*Book)
	bookIDs := []int{}

	for rows.Next() {
		var book Book

		if err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.Subtitle,
			&book.Description,
			&book.Language,
			&book.PageCount,
			&book.PublishDate,
			&book.ImageLink,
			&book.Notes,
			&book.CreatedAt,
			&book.LastUpdated,
			&book.ISBN10,
			&book.ISBN13,
		); err != nil {
			b.Logger.Error("Error scanning book", "error", err)
			return nil, err
		}

		book.IsInLibrary = true
		books = append(books, book)
		bookIDMap[book.ID] = &book
		bookIDs = append(bookIDs, book.ID)
	}

	// Fetch all authors for the books in a single query
	authorQuery := `
	SELECT ba.book_id, a.name
	FROM book_authors ba
	INNER JOIN authors a ON ba.author_id = a.id
	WHERE ba.book_id = ANY($1)`

	authorRows, err := b.DB.QueryContext(ctx, authorQuery, pq.Array(bookIDs))
	if err != nil {
		b.Logger.Error("Error retrieving authors for books", "error", err)
		return nil, err
	}
	defer authorRows.Close()

	for authorRows.Next() {
		var bookID int
		var authorName string
		if err := authorRows.Scan(&bookID, &authorName); err != nil {
			b.Logger.Error("Error scanning author", "error", err)
			return nil, err
		}
		bookIDMap[bookID].Authors = append(bookIDMap[bookID].Authors, authorName)
	}

	// Fetch all genres for the books in a single query
	genreQuery := `
	SELECT bg.book_id, g.name
	FROM book_genres bg
	INNER JOIN genres g ON bg.genre_id = g.id
	WHERE bg.book_id = ANY($1)`

	genreRows, err := b.DB.QueryContext(ctx, genreQuery, pq.Array(bookIDs))
	if err != nil {
		b.Logger.Error("Error retrieving genres for books", "error", err)
		return nil, err
	}
	defer genreRows.Close()

	for genreRows.Next() {
		var bookID int
		var genreName string
		if err := genreRows.Scan(&bookID, &genreName); err != nil {
			b.Logger.Error("Error scanning genre", "error", err)
			return nil, err
		}
		bookIDMap[bookID].Genres = append(bookIDMap[bookID].Genres, genreName)
	}

	if err = rows.Err(); err != nil {
		b.Logger.Error("GetBooksByAuthor - Error with rows", "error", err)
		return nil, err
	}

	return books, nil
}

func (b *BookModel) GetAuthorsForBook(bookID int) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
	SELECT array_to_json(array_agg(a.name))
	FROM authors a
	JOIN book_authors ba ON a.id = ba.author_id
	WHERE ba.book_id = $1`

	var authorsJSON []byte
	err := b.DB.QueryRowContext(ctx, query, bookID).Scan(&authorsJSON)
	if err != nil {
		b.Logger.Error("Error fetching authors for book", "error", err)
		return nil, err
	}

	// Unmarshal the JSON array of authors
	var authors []string
	if err := json.Unmarshal(authorsJSON, &authors); err != nil {
		b.Logger.Error("Error unmarshalling authors JSON", "error", err)
		return nil, err
	}

	return authors, nil
}

func (b *BookModel) GetAllBooksByAuthors(userID int) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var rows *sql.Rows
	var err error

	// Use the prepared statement if available, otherwise fall back to raw query
	if b.getAllBooksByAuthorsStmt != nil {
		b.Logger.Info("Using prepared statement for retrieving books by authors")
		rows, err = b.getAllBooksByAuthorsStmt.QueryContext(ctx, userID)
	} else {
		b.Logger.Warn("Prepared statement for retrieving books by authors not initialized, using fallback query")
		query := `
			SELECT b.id, b.title, b.subtitle, b.description, b.language, b.page_count, b.publish_date,
						 b.image_link, b.notes, b.created_at, b.last_updated, b.isbn_10, b.isbn_13,
						 a.name AS author_name,
						 json_agg(DISTINCT g.name) AS genres,
						 json_agg(DISTINCT f.format_type) AS formats,
						 b.tags
			FROM books b
			INNER JOIN book_authors ba ON b.id = ba.book_id
			INNER JOIN authors a ON ba.author_id = a.id
			INNER JOIN user_books ub ON b.id = ub.book_id
			LEFT JOIN book_genres bg ON b.id = bg.book_id
			LEFT JOIN genres g ON bg.genre_id = g.id
			LEFT JOIN book_formats bf ON b.id = bf.book_id
			LEFT JOIN formats f ON bf.format_id = f.id
			WHERE ub.user_id = $1
			GROUP BY b.id, a.name`
		rows, err = b.DB.QueryContext(ctx, query, userID)
	}

	if err != nil {
		b.Logger.Error("Error retrieving books by authors", "error", err)
		return nil, err
	}
	defer rows.Close()

	// Store authors and books
	authors := []string{}
	booksByAuthor := map[string][]Book{}

	for rows.Next() {
		var book Book
		var authorName string
		var genresJSON, formatsJSON, tagsJSON []byte

		// Scan the result
		if err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.Subtitle,
			&book.Description,
			&book.Language,
			&book.PageCount,
			&book.PublishDate,
			&book.ImageLink,
			&book.Notes,
			&book.CreatedAt,
			&book.LastUpdated,
			&book.ISBN10,
			&book.ISBN13,
			&authorName,
			&genresJSON,
			&formatsJSON,
			&tagsJSON,
		); err != nil {
			b.Logger.Error("Error scanning book by author", "error", err)
			return nil, err
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(genresJSON, &book.Genres); err != nil {
			b.Logger.Error("Error unmarshalling genres JSON", "error", err)
			return nil, err
		}
		if err := json.Unmarshal(formatsJSON, &book.Formats); err != nil {
			b.Logger.Error("Error unmarshalling formats JSON", "error", err)
			return nil, err
		}
		if err := json.Unmarshal(tagsJSON, &book.Tags); err != nil {
			b.Logger.Error("Error unmarshalling tags JSON", "error", err)
			return nil, err
		}

		// Mark book as in Library
		book.IsInLibrary = true

		// Populate the Authors field directly
		book.Authors = []string{authorName}

		// Add author to the list if not already present
		if _, found := booksByAuthor[authorName]; !found {
			authors = append(authors, authorName)
		}

		// Add book to the author's list
		booksByAuthor[authorName] = append(booksByAuthor[authorName], book)
	}

	// Check for row errors
	if err = rows.Err(); err != nil {
		b.Logger.Error("Error with rows", "error", err)
		return nil, err
	}

	// Sort authors by last name
	sort.Slice(authors, func(i, j int) bool {
		return getLastName(authors[i]) < getLastName(authors[j])
	})

	// Create the result map with index keys for authors
	result := map[string]interface{}{
		"allAuthors": authors,
	}

	for i, author := range authors {
		key := strconv.Itoa(i)
		result[key] = booksByAuthor[author]
	}

	return result, nil
}

// Helper function to get the last name from a full name
func getLastName(fullName string) string {
	parts := strings.Fields(fullName)
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// Helper function for UpdateBooks
func (b *BookModel) updateAuthors(ctx context.Context, bookID int, authors []string) error {
	// Delete existing authors for the book
	deleteStatement := `DELETE FROM book_authors WHERE book_id = $1`
	if _, err := b.DB.ExecContext(ctx, deleteStatement, bookID); err != nil {
		b.Logger.Error("Error deleting existing authors for book", "error", err)
		return err
	}

	// Insert new authors for the book
	for _, authorName := range authors {
		// Check if author already exists in authors table
		var authorID int
		selectStatement := `SELECT id FROM authors WHERE name = $1`
		err := b.DB.QueryRowContext(ctx, selectStatement, authorName).Scan(&authorID)
		if err != nil {
			if err == sql.ErrNoRows {
				// Insert author if doesn't exist
				insertAuthorStatement := `INSERT INTO authors (name) VALUES ($1) RETURNING id`
				err := b.DB.QueryRowContext(ctx, insertAuthorStatement, authorName).Scan(&authorID)
				if err != nil {
					b.Logger.Error("Error inserting new author", "error", err)
					return err
				}
			} else {
				// Something else broke
				b.Logger.Error("Error checking for existing author", "error", err)
				return err
			}
		}

		// Link author to book
		insertLinkStatement := `INSERT INTO book_authors (book_id, author_id) VALUES ($1, $2)`
		_, err = b.DB.ExecContext(ctx, insertLinkStatement, bookID, authorID)
		if err != nil {
			b.Logger.Error("Error linking author to book", "error", err)
			return err
		}
	}

	return nil
}

// ISBN10 + ISBN13 (Returns a HashSet)
func (b *BookModel) GetAllBooksISBN10(userID int) (*collections.Set, error) {
	// Check cache
	if cache, found := isbn10Cache.Load(userID); found {
		b.Logger.Info("Fetching ISBN10 from cache")
		return cache.(*collections.Set), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
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
func (b *BookModel) GetAllBooksISBN13(userID int) (*collections.Set, error) {
	// Check cache
	if cache, found := isbn13Cache.Load(userID); found {
		b.Logger.Info("Fetching ISBN13 from cache")
		return cache.(*collections.Set), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
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
func (b *BookModel) GetAllBooksTitles(userID int) (*collections.Set, error) {
	// Check cache
	if cache, found := titleCache.Load(userID); found {
		b.Logger.Info("Fetching Title info from cache")
		return cache.(*collections.Set), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
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
func (b *BookModel) GetAllBooksPublishDate(userID int) ([]BookInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
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


// Formats
func (b *BookModel) AddFormats(ctx context.Context, bookID int, formatIDs []int) error {
	if len(formatIDs) == 0 {
			return nil
	}

	// Build query using parameter placeholders
	valueStrings := make([]string, len(formatIDs))
	valueArgs := make([]interface{}, 0, len(formatIDs)*2)

	for i, formatID := range formatIDs {
			valueStrings[i] = fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2)
			valueArgs = append(valueArgs, bookID, formatID)
	}

	statement := fmt.Sprintf("INSERT INTO book_formats (book_id, format_id) VALUES %s", strings.Join(valueStrings, ","))

	// Execute parameterized query with arguments
	_, err := b.DB.ExecContext(ctx, statement, valueArgs...)
	if err != nil {
			b.Logger.Error("Book Model - Error adding formats", "bookID", bookID, "formatIDs", formatIDs, "error", err)
			return err
	}

	return nil
}


func (b *BookModel) addOrGetFormatID(ctx context.Context, format string) (int, error) {
	var formatID int
	statement := `
		INSERT INTO formats (format_type)
		VALUES ($1)
		ON CONFLICT (format_type) DO UPDATE
		SET format_type = EXCLUDED.format_type
		RETURNING id`
	err := b.DB.QueryRowContext(ctx, statement, format).Scan(&formatID)
	if err != nil {
		b.Logger.Error("Error inserting or updating format", "error", err)
		return 0, err
	}
	return formatID, nil
}

func (b *BookModel) GetAllBooksByFormat(userID int) (map[string][]Book, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var rows *sql.Rows
	var err error

	if b.getAllBooksByFormatStmt != nil {
		b.Logger.Info("Using prepared statement for retrieving books by format")
		rows, err = b.getAllBooksByFormatStmt.QueryContext(ctx, userID)
	} else {
		b.Logger.Warn("Prepared statement for retrieving books by format is not available. Falling back to raw SQL query")
		query := `
		SELECT
			b.id, b.title, b.subtitle, b.description, b.language, b.page_count, b.publish_date,
			b.image_link, b.notes, b.created_at, b.last_updated, b.isbn_10, b.isbn_13,
			f.format_type,
			array_to_json(array_agg(DISTINCT a.name)) as authors,
			array_to_json(array_agg(DISTINCT g.name)) as genres,
			b.tags
		FROM books b
		INNER JOIN book_formats bf ON b.id = bf.book_id
		INNER JOIN formats f ON bf.format_id = f.id
		INNER JOIN user_books ub ON b.id = ub.book_id
		LEFT JOIN book_authors ba ON b.id = ba.book_id
		LEFT JOIN authors a ON ba.author_id = a.id
		LEFT JOIN book_genres bg ON b.id = bg.book_id
		LEFT JOIN genres g ON bg.genre_id = g.id
		WHERE ub.user_id = $1
		GROUP BY b.id, f.format_type`
		rows, err = b.DB.QueryContext(ctx, query, userID)
	}

	if err != nil {
		b.Logger.Error("Error retrieving books by format", "error", err)
		return nil, err
	}
	defer rows.Close()

	booksByID := make(map[int]*Book)
	bookFormats := make(map[int][]string)
	booksByFormat := make(map[string][]Book)

	for rows.Next() {
		var book Book
		var authorsJSON, genresJSON, tagsJSON []byte
		var formatType string

		if err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.Subtitle,
			&book.Description,
			&book.Language,
			&book.PageCount,
			&book.PublishDate,
			&book.ImageLink,
			&book.Notes,
			&book.CreatedAt,
			&book.LastUpdated,
			&book.ISBN10,
			&book.ISBN13,
			&formatType,
			&authorsJSON,
			&genresJSON,
			&tagsJSON,
		); err != nil {
			b.Logger.Error("Error scanning book", "error", err)
			return nil, err
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(authorsJSON, &book.Authors); err != nil {
			b.Logger.Error("Error unmarshalling authors JSON", "error", err)
			return nil, err
		}
		if err := json.Unmarshal(genresJSON, &book.Genres); err != nil {
			b.Logger.Error("Error unmarshalling genres JSON", "error", err)
			return nil, err
		}
		if err := json.Unmarshal(tagsJSON, &book.Tags); err != nil {
			b.Logger.Error("Error unmarshalling tags JSON", "error", err)
			return nil, err
		}

		// Track formats for the book
		bookFormats[book.ID] = append(bookFormats[book.ID], formatType)

		// If the book is not yet in the map, add it
		if _, exists := booksByID[book.ID]; !exists {
			book.IsInLibrary = true
			booksByID[book.ID] = &book
		}
	}

	// After collecting all formats, populate the booksByFormat map
	for bookID, book := range booksByID {
		book.Formats = bookFormats[bookID]
		for _, format := range book.Formats {
			booksByFormat[format] = append(booksByFormat[format], *book)
		}
	}

	if err := rows.Err(); err != nil {
		b.Logger.Error("Error with rows", "error", err)
		return nil, err
	}

	return booksByFormat, nil
}

func (b *BookModel) GetFormats(ctx context.Context, bookID int) ([]string, error) {
	// Check cache
	if cache, found := formatsCache.Load(bookID); found {
		b.Logger.Info("Fetching formats book info from cache", "bookID", bookID)
		cachedFormats := cache.([]string)
		return append([]string(nil), cachedFormats...), nil
	}

	var rows *sql.Rows
	var err error

	// Use prepared statement if available
	if b.getFormatsStmt != nil {
		b.Logger.Info("Using prepared statement for fetching formats")
		rows, err = b.getFormatsStmt.QueryContext(ctx, bookID)
	} else {
		b.Logger.Warn("Prepared statement for fetching formats is not available. Falling back to raw SQL query")
		query := `
		SELECT f.format_type
		FROM formats f
		JOIN book_formats bf ON f.id = bf.format_id
		WHERE bf.book_id = $1`
		rows, err = b.DB.QueryContext(ctx, query, bookID)
	}

	if err != nil {
		b.Logger.Error("Book Model - Error fetching formats", "error", err)
		return nil, err
	}
	defer rows.Close()

	// Collect the format types in a slice
	var formats []string
	for rows.Next() {
		var format string
		if err := rows.Scan(&format); err != nil {
			b.Logger.Error("Book Model - Error scanning format", "error", err)
			return nil, err
		}
		formats = append(formats, format)
	}

	// Handle any errors encountered during iteration
	if err = rows.Err(); err != nil {
		b.Logger.Error("Book Model - Error with rows during formats fetch", "error", err)
		return nil, err
	}

	// Cache result
	formatsCache.Store(bookID, formats)
	b.Logger.Info("Caching formats for book", "bookID", bookID)

	return formats, nil
}

// Helper fn for UpdateBook
func (b *BookModel) updateFormats(ctx context.Context, bookID int, newFormats []string) error {
	// Invalidate cache for bookID
	formatsCache.Delete(bookID)
	b.Logger.Info("Invalidating formats cache for book", "bookID", bookID)

	// Fetch current formats for the book, passing context
	currentFormats, err := b.GetFormats(ctx, bookID)
	if err != nil {
		b.Logger.Error("Book Model - Error fetching current formats", "error", err)
		return err
	}

	// Find formats to remove and formats to add
	formatsToRemove := utils.FindDifference(currentFormats, newFormats)
	formatsToAdd := utils.FindDifference(newFormats, currentFormats)

	// Remove specific formats
	if len(formatsToRemove) > 0 {
		err := b.RemoveSpecificFormats(ctx, bookID, formatsToRemove)
		if err != nil {
			b.Logger.Error("Book Model - Error removing specific formats", "error", err)
			return err
		}
	}

	// Add new formats
	var formatIDs []int
	for _, format := range formatsToAdd {
		formatID, err := b.addOrGetFormatID(ctx, format)
		if err != nil {
			b.Logger.Error("Error getting format ID", "error", err)
			return err
		}
		formatIDs = append(formatIDs, formatID)
	}

	if len(formatIDs) > 0 {
		err = b.AddFormats(ctx, bookID, formatIDs)
		if err != nil {
			b.Logger.Error("Error adding format associations", "error", err)
			return err
		}
	}

	return nil
}

func (b *BookModel) RemoveSpecificFormats(ctx context.Context, bookID int, formats []string) error {
	statement := `
		DELETE FROM book_formats
		WHERE book_id = $1
		AND format_id IN (
			SELECT id FROM formats WHERE format_type = ANY($2)
		)`

	_, err := b.DB.ExecContext(ctx, statement, bookID, pq.Array(formats))
	if err != nil {
		b.Logger.Error("Error removing specific formats", "error", err)
		return err
	}

	return nil
}

// DEBUG - to possibly remove
func (b *BookModel) RemoveFormats(bookID int) error {
	formatsCache.Delete(bookID)

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	statement := `DELETE FROM book_formats WHERE book_id = $1`
	_, err := b.DB.ExecContext(ctx, statement, bookID)
	if err != nil {
		b.Logger.Error("Book Model - Error removing formats", "error", err)
		return err
	}

	return nil
}


// Genres
func (b *BookModel) addOrGetGenreID(ctx context.Context, genreName string) (int, error) {
	var genreID int
	statement := `
		INSERT INTO genres (name)
		VALUES ($1)
		ON CONFLICT (name) DO UPDATE
		SET name = EXCLUDED.name
		RETURNING id`
	err := b.DB.QueryRowContext(ctx, statement, genreName).Scan(&genreID)
	if err != nil {
		b.Logger.Error("Error inserting or updating genre", "error", err)
		return 0, err
	}
	return genreID, nil
}

func (b *BookModel) AddGenre(ctx context.Context, bookID, genreID int) error {
	statement := `INSERT INTO book_genres (book_id, genre_id) VALUES ($1, $2)`
	_, err := b.DB.ExecContext(ctx, statement, bookID, genreID)
	if err != nil {
		b.Logger.Error("Error adding genre association", "error", err)
		return err
	}

	return nil
}

func (b *BookModel) GetGenres(ctx context.Context, bookID int) ([]string, error) {
	// Check cache
	if cache, found := genresCache.Load(bookID); found {
		b.Logger.Info("Fetching genres from cache for book", "bookID", bookID)
		cachedGenres := cache.([]string)
		return append([]string(nil), cachedGenres...), nil
	}

	var rows *sql.Rows
	var err error

	// Use prepared statement if available
	if b.getGenresStmt != nil {
		b.Logger.Info("Using prepared statement for fetching genres")
		rows, err = b.getGenresStmt.QueryContext(ctx, bookID)
	} else {
		b.Logger.Warn("Prepared statement for fetching genres is not available. Falling back to raw SQL query")
		query := `
		SELECT g.name
		FROM genres g
		JOIN book_genres bg ON g.id = bg.genre_id
		WHERE bg.book_id = $1`
		rows, err = b.DB.QueryContext(ctx, query, bookID)
	}

	if err != nil {
		b.Logger.Error("Error fetching genres", "error", err)
		return nil, err
	}
	defer rows.Close()

	// Collect genres
	var genres []string
	for rows.Next() {
		var genre string
		if err := rows.Scan(&genre); err != nil {
			b.Logger.Error("Error scanning genre", "error", err)
			return nil, err
		}
		genres = append(genres, genre)
	}

	// Check for errors after looping through the rows
	if err = rows.Err(); err != nil {
		b.Logger.Error("Error with rows during genres fetch", "error", err)
		return nil, err
	}

	// Cache the result
	genresCache.Store(bookID, genres)
	b.Logger.Info("Caching genres for book", "bookID", bookID)

	return genres, nil
}

func (b *BookModel) GetAllBooksByGenres(ctx context.Context, userID int) (map[string]interface{}, error) {
	var rows *sql.Rows
	var err error

	// Use the prepared statement if available, else fall back to a raw query
	if b.getAllBooksByGenresStmt != nil {
			b.Logger.Info("Using prepared statement for retrieving books by genres")
			rows, err = b.getAllBooksByGenresStmt.QueryContext(ctx, userID)
	} else {
			b.Logger.Info("Prepared statement unavailable, using fallback query for retrieving books by genres")
			query := `
					SELECT b.id, b.title, b.subtitle, b.description, b.language, b.page_count, b.publish_date,
								 b.image_link, b.notes, b.created_at, b.last_updated, b.isbn_10, b.isbn_13,
								 json_agg(DISTINCT g.name) AS genres,
								 json_agg(DISTINCT a.name) AS authors,
								 json_agg(DISTINCT f.format_type) AS formats,
								 b.tags
					FROM books b
					INNER JOIN user_books ub ON b.id = ub.book_id
					LEFT JOIN book_genres bg ON b.id = bg.book_id
					LEFT JOIN genres g ON bg.genre_id = g.id
					LEFT JOIN book_authors ba ON b.id = ba.book_id
					LEFT JOIN authors a ON ba.author_id = a.id
					LEFT JOIN book_formats bf ON b.id = bf.book_id
					LEFT JOIN formats f ON bf.format_id = f.id
					WHERE ub.user_id = $1
					GROUP BY b.id`
			rows, err = b.DB.QueryContext(ctx, query, userID)
	}

	if err != nil {
			b.Logger.Error("Error retrieving books by genres", "error", err)
			return nil, err
	}
	defer rows.Close()

	genresSet := make(map[string]struct{}) // Track unique genres
	booksByGenre := map[string][]Book{}    // Store books by genre

	// Iterate through the rows and process the results
	for rows.Next() {
			var book Book
			var genresJSON, authorsJSON, formatsJSON, tagsJSON []byte

			// Ensure the scan order matches the SQL query's column order
			if err := rows.Scan(
					&book.ID, &book.Title, &book.Subtitle, &book.Description, &book.Language, &book.PageCount,
					&book.PublishDate, &book.ImageLink, &book.Notes, &book.CreatedAt, &book.LastUpdated,
					&book.ISBN10, &book.ISBN13, &genresJSON, &authorsJSON, &formatsJSON, &tagsJSON,
			); err != nil {
					b.Logger.Error("Error scanning book by genre", "error", err)
					return nil, err
			}

			// Unmarshal JSON fields into the respective slices
			if err := json.Unmarshal(genresJSON, &book.Genres); err != nil {
					b.Logger.Error("Error unmarshalling genres JSON", "error", err)
					return nil, err
			}
			if err := json.Unmarshal(authorsJSON, &book.Authors); err != nil {
					b.Logger.Error("Error unmarshalling authors JSON", "error", err)
					return nil, err
			}
			if err := json.Unmarshal(formatsJSON, &book.Formats); err != nil {
					b.Logger.Error("Error unmarshalling formats JSON", "error", err)
					return nil, err
			}
			if err := json.Unmarshal(tagsJSON, &book.Tags); err != nil {
					b.Logger.Error("Error unmarshalling tags JSON", "error", err)
					return nil, err
			}

			// Add genres to the genres set
			for _, genre := range book.Genres {
					genresSet[genre] = struct{}{}
			}

			// Add book to the booksByGenre map
			for _, genre := range book.Genres {
					booksByGenre[genre] = append(booksByGenre[genre], book)
			}
	}

	if err = rows.Err(); err != nil {
			b.Logger.Error("Error with rows", "error", err)
			return nil, err
	}

	// Convert the genres set to a sorted list
	var genres []string
	for genre := range genresSet {
			genres = append(genres, genre)
	}
	sort.Strings(genres)

	// Prepare the final result with genres and their associated books
	result := map[string]interface{}{
			"allGenres": genres,
	}

	for i, genre := range genres {
			key := strconv.Itoa(i)

			// Sort the books for each genre by author's last name
			sort.Slice(booksByGenre[genre], func(i, j int) bool {
					if len(booksByGenre[genre][i].Authors) > 0 && len(booksByGenre[genre][j].Authors) > 0 {
							return getLastName(booksByGenre[genre][i].Authors[0]) < getLastName(booksByGenre[genre][j].Authors[0])
					}
					return false
			})

			// Extract the first image for each book in the genre
			genreImgs := make([]string, len(booksByGenre[genre]))
			for j, book := range booksByGenre[genre] {
					if book.ImageLink != "" {
							genreImgs[j] = book.ImageLink
					}
			}

			result[key] = map[string]interface{}{
					"genreImgs": genreImgs,
					"bookList":  booksByGenre[genre],
			}
	}

	//b.Logger.Info("Final result being sent to the frontend", "result", result)
	return result, nil
}

func (b *BookModel) GetBooksListByGenre(ctx context.Context, userID int) (map[string]interface{}, error) {
	const cacheTTL = time.Hour

	// Check cache with TTL
	if cacheEntry, found := booksByGenresCache.Load(userID); found {
			entry := cacheEntry.(BooksByGenresCacheEntry)
			if time.Since(entry.timestamp) < cacheTTL {
					b.Logger.Info("Fetching genres info from cache for user", "userID", userID)
					return entry.data, nil
			}
			booksByGenresCache.Delete(userID) // Cache entry expired, delete it
	}

	var rows *sql.Rows
	var err error

	// Use prepared statement if available
	if b.getBookListByGenreStmt != nil {
			b.Logger.Info("Using prepared statement for fetching genres")
			rows, err = b.getBookListByGenreStmt.QueryContext(ctx, userID)
	} else {
			b.Logger.Warn("Prepared statement for fetching genres unavailable. Falling back to raw SQL query")
			query := `
					SELECT g.name, COUNT(DISTINCT b.id) AS total_books
					FROM books b
					INNER JOIN book_genres bg ON b.id = bg.book_id
					INNER JOIN genres g ON bg.genre_id = g.id
					INNER JOIN user_books ub ON b.id = ub.book_id
					WHERE ub.user_id = $1
					GROUP BY g.name
					ORDER BY total_books DESC`
			rows, err = b.DB.QueryContext(ctx, query, userID)
	}

	if err != nil {
			b.Logger.Error("Error fetching books by genre", "error", err)
			return nil, err
	}
	defer rows.Close()

	// Collect genres and total book count
	var booksByGenre []map[string]interface{}

	for rows.Next() {
			var genre string
			var count int
			if err := rows.Scan(&genre, &count); err != nil {
					b.Logger.Error("Error scanning genre data", "error", err)
					return nil, err
			}

			booksByGenre = append(booksByGenre, map[string]interface{}{
					"label": genre,
					"count": count,
			})
	}

	if err = rows.Err(); err != nil {
			b.Logger.Error("Error iterating rows", "error", err)
			return nil, err
	}

	// Prepare the result
	result := map[string]interface{}{
			"booksByGenre": booksByGenre,
	}

	// Cache the result with TTL
	booksByGenresCache.Store(userID, BooksByGenresCacheEntry{
			data:      result,
			timestamp: time.Now(),
	})
	b.Logger.Info("Caching genres info for user", "userID", userID)

	return result, nil
}


// Helper fn for UpdateBook
func (b *BookModel) updateGenres(ctx context.Context, bookID int, newGenres []string) error {
	genresCache.Delete(bookID)
	b.Logger.Info("Invalidating genres cache for book", "bookID", bookID)

	// Fetch current genres for the book with context
	currentGenres, err := b.GetGenres(ctx, bookID)
	if err != nil {
		b.Logger.Error("Book Model - Error fetching current genres", "error", err)
		return err
	}

	// Find genres to remove and genres to add
	genresToRemove := utils.FindDifference(currentGenres, newGenres)
	genresToAdd := utils.FindDifference(newGenres, currentGenres)

	// Remove specific genres
	if len(genresToRemove) > 0 {
		err := b.RemoveSpecificGenres(ctx, bookID, genresToRemove) // Pass ctx to this method
		if err != nil {
			b.Logger.Error("Book Model - Error removing specific genres", "error", err)
			return err
		}
	}

	// Add new genres
	for _, genre := range genresToAdd {
		genreID, err := b.addOrGetGenreID(ctx, genre) // Use the updated method with ctx
		if err != nil {
			b.Logger.Error("Error getting genre ID", "error", err)
			return err
		}
		err = b.AddGenre(ctx, bookID, genreID) // Use the refactored AddGenre method
		if err != nil {
			b.Logger.Error("Error adding genre association", "error", err)
			return err
		}
	}

	return nil
}

func (b *BookModel) RemoveSpecificGenres(ctx context.Context, bookID int, genres []string) error {

	statement := `
		DELETE FROM book_genres
		WHERE book_id = $1
		AND genre_id IN (
			SELECT id FROM genres WHERE name = ANY($2)
		)`

	_, err := b.DB.ExecContext(ctx, statement, bookID, pq.Array(genres))
	if err != nil {
		b.Logger.Error("Error removing specific genres", "error", err)
		return err
	}

	return nil
}

func (b *BookModel) GetTagsForBook(ctx context.Context, bookID int) ([]string, error) {

	query := `
	SELECT tags
	FROM books
	WHERE id = $1`

	var tagsJSON []byte
	err := b.DB.QueryRowContext(ctx, query, bookID).Scan(&tagsJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			b.Logger.Warn("No tags found for book", "bookID", bookID)
			// Return empty slice if no tags are found
			return nil, nil
		}
		b.Logger.Error("Error fetching tags for book", "error", err)
		return nil, err
	}

	// Return an empty slice if tagsJSON is null
	if tagsJSON == nil {
		return []string{}, nil
	}

	// Unmarshal the JSON array of tags
	var tags []string
	if err := json.Unmarshal(tagsJSON, &tags); err != nil {
		b.Logger.Error("Error unmarshalling tags JSON", "error", err)
		return nil, err
	}

	return tags, nil
}


// DEBUG - to possibly remove
func (b *BookModel) RemoveGenres(bookID int) error {
	genresCache.Delete(bookID)

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	statement := `DELETE FROM book_genres WHERE book_id = $1`
	_, err := b.DB.ExecContext(ctx, statement, bookID)
	if err != nil {
		b.Logger.Error("Error removing genres", "error", err)
		return err
	}

	return nil
}

// Languages
func (b *BookModel) GetBooksByLanguage(ctx context.Context, userID int) (map[string]interface{}, error) {
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

// Tags
func (b *BookModel) GetUserTags(ctx context.Context, userID int) (map[string]interface{}, error) {
	// Check cache with TTL
	if cacheEntry, found := userTagsCache.Load(userID); found {
			entry := cacheEntry.(UserTagsCacheEntry)
			if time.Since(entry.timestamp) < time.Hour {
					b.Logger.Info("Fetching user tags from cache for user", "userID", userID)
					return entry.data, nil
			}
			// Cache entry expired, delete it
			userTagsCache.Delete(userID)
	}

	var rows *sql.Rows
	var err error

	// Use prepared statement if available
	if b.getUserTagsStmt != nil {
			b.Logger.Info("Using prepared statement for fetching user tags")
			rows, err = b.getUserTagsStmt.QueryContext(ctx, userID)
	} else {
			b.Logger.Warn("Prepared statement for fetching user tags unavailable. Falling back to raw SQL query")
			query := `
			SELECT b.tags
			FROM books b
			INNER JOIN user_books ub ON b.id = ub.book_id
			WHERE ub.user_id = $1`
			rows, err = b.DB.QueryContext(ctx, query, userID)
	}

	if err != nil {
			b.Logger.Error("Error fetching user tags", "error", err)
			return nil, err
	}
	defer rows.Close()

	// Process the tags and count occurrences
	tagCount := make(map[string]int)

	for rows.Next() {
			var tagsJSON []byte
			if err := rows.Scan(&tagsJSON); err != nil {
					b.Logger.Error("Error scanning tags", "error", err)
					return nil, err
			}

			var tags []string
			if err := json.Unmarshal(tagsJSON, &tags); err != nil {
					b.Logger.Error("Error unmarshalling tags JSON", "error", err)
					return nil, err
			}

			// Convert spaces to underscores and count occurrences
			for _, tag := range tags {
					formattedTag := strings.ReplaceAll(tag, " ", "_")
					tagCount[formattedTag]++
			}
	}

	// Create a max heap for more efficient sorting
	h := &customheap.TagHeap{}
	heap.Init(h)

	for tag, count := range tagCount {
		heap.Push(h, customheap.TagCount{Tag: tag, Count: count})
	}

	// Create the result array with the new format
	userTags := make([]map[string]interface{}, 0, len(tagCount))
	for h.Len() > 0 {
		tagCount := heap.Pop(h).(customheap.TagCount)
		userTags = append(userTags, map[string]interface{}{
					"label":   tagCount.Tag,
					"count": tagCount.Count,
			})
	}

	// Prepare the result
	result := map[string]interface{}{
			"userTags":       userTags,
	}

	// Cache the result
	userTagsCache.Store(userID, UserTagsCacheEntry{data: result, timestamp: time.Now()})
	b.Logger.Info("Caching user tags for user", "userID", userID)

	return result, nil
}


// Helper fn: Sanitize Fields
func sanitizeCSVField(field string) string {
	// Guard Clause
	if field == "" {
		return ""
	}

	// Prevent formula injection
	if strings.ContainsAny(field, "=+-@") {
		field = "'" + field
	}

	// Escape double quotes
	field = strings.ReplaceAll(field, "\"", "\"\"")

	// Remove/replace ctl chars
	var sanitized strings.Builder
	for _, r := range field {
		if unicode.IsControl(r) {
			sanitized.WriteRune(' ')
		} else {
			sanitized.WriteRune(r)
		}
	}
	return sanitized.String()
}

// Generate Book CSV
func (b *BookModel) GenerateBookCSV(userID int, writer io.Writer) error {
	books, err := b.GetAllBooksByUserID(userID)
	if err != nil {
		b.Logger.Error("Failed to fetch books for user", "error", err)
		return err
	}

	csvWriter := csv.NewWriter(writer)
	defer func() {
		csvWriter.Flush()
		if err := csvWriter.Error(); err != nil {
			b.Logger.Error("Error flushing CSV writer: ", "error", err)
		}
	}()

	if _, err := writer.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		return fmt.Errorf("failed to write UTF-8 BOM: %w", err)
	}

	header := []string{"Title", "Authors", "ISBN", "Language", "PageCount", "Publish Date"}
	if err := csvWriter.Write(header); err != nil {
		b.Logger.Error("Failed to write CSV header", "error", err)
		return err
	}

	// Write data
	for _, book := range books {
		row := []string{
			sanitizeCSVField(book.Title),
			sanitizeCSVField(strings.Join(book.Authors, ", ")),
			sanitizeCSVField(book.ISBN13),
			sanitizeCSVField(book.Language),
			sanitizeCSVField(strconv.Itoa(book.PageCount)),
			sanitizeCSVField(book.PublishDate),
		}
		// Add UTF-8 Byte Order Mark for Excel
		if err := csvWriter.Write(row); err != nil {
			b.Logger.Error("Failed to write CSV row", "bookID", book.ID, "error", err )
			return err
		}
	}

	return nil
}
