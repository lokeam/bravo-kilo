package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"

	"github.com/lib/pq"
	"github.com/lokeam/bravo-kilo/internal/dbconfig"
)

type AuthorRepository interface {
	InitPreparedStatements() error
	InsertAuthor(ctx context.Context, tx *sql.Tx, author string) (int, error)
	AssociateBookWithAuthor(ctx context.Context, tx *sql.Tx, bookID, authorID int) error
	GetAllBooksByAuthors(userID int) (map[string]interface{}, error)
	GetAuthorsForBook(bookID int) ([]string, error)
	GetAuthorsForBooks(bookIDs []int) (map[int][]string, error)
	GetAuthorIDByName(ctx context.Context, tx *sql.Tx, authorName string, authorID *int) error
	GetBooksByAuthor(authorName string) ([]Book, error)
	BatchInsertAuthors(ctx context.Context, tx *sql.Tx, bookID int, authors []string) error
}

type AuthorRepositoryImpl struct {
	DB                        *sql.DB
	Logger                    *slog.Logger
	getAllBooksByAuthorsStmt  *sql.Stmt
	getAuthorsForBooksStmt    *sql.Stmt
}

func NewAuthorRepository(db *sql.DB, logger *slog.Logger) (AuthorRepository, error) {
	if db == nil || logger == nil {
		return nil, fmt.Errorf("new author repository, database or logger is nil")
	}

	return &AuthorRepositoryImpl{
		DB:      db,
		Logger: logger,
	}, nil
}

func (r *AuthorRepositoryImpl) InitPreparedStatements() error {
	var err error

	// Prepared select statement for GetAuthorsForBooks
	r.getAuthorsForBooksStmt, err = r.DB.Prepare(`
		SELECT ba.book_id, a.name
		FROM authors a
		JOIN book_authors ba ON a.id = ba.author_id
		WHERE ba.book_id = ANY($1)`)
	if err != nil {
		return err
	}

	// Prepared select statement for GetAllBooksByAuthors
	r.getAllBooksByAuthorsStmt, err = r.DB.Prepare(`
	SELECT b.id, b.title, b.subtitle, b.description, b.language, b.page_count, b.publish_date,
	       b.image_link, b.notes, b.created_at, b.last_updated, b.isbn_10, b.isbn_13,
	       a.name AS author_name,
	       json_agg(DISTINCT g.name) AS genres,
	       json_agg(DISTINCT f.format_type) AS formats,
	       json_agg(DISTINCT t.name) AS tags,
	       b.tags
	FROM books b
	INNER JOIN book_authors ba ON b.id = ba.book_id
	INNER JOIN authors a ON ba.author_id = a.id
	INNER JOIN user_books ub ON b.id = ub.book_id
	LEFT JOIN book_genres bg ON b.id = bg.book_id
	LEFT JOIN genres g ON bg.genre_id = g.id
	LEFT JOIN book_formats bf ON b.id = bf.book_id
	LEFT JOIN formats f ON bf.format_id = f.id
	LEFT JOIN book_tags bt ON b.id = bt.book_id
	LEFT JOIN tags t ON bt.tag_id = t.id
	WHERE ub.user_id = $1::integer
	GROUP BY b.id, a.name`)
	if err != nil {
		return err
	}

	return nil
}

func (r *AuthorRepositoryImpl) InsertAuthor(ctx context.Context, tx *sql.Tx, author string) (int, error) {
	var authorID int

	// Check if author already exists
	err := tx.QueryRowContext(ctx, `SELECT id FROM authors WHERE name = $1`, author).Scan(&authorID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Author doesn't exist, insert it
			err = tx.QueryRowContext(ctx, `INSERT INTO authors (name) VALUES ($1) RETURNING id`, author).Scan(&authorID)
			if err != nil {
				r.Logger.Error("Error inserting new author", "error", err, "author", author)
				return 0, err
			}
		} else {
			r.Logger.Error("Error checking if author exists", "error", err, "author", author)
			return 0, err
		}
	}

	return authorID, nil
}


func (b *AuthorRepositoryImpl) AssociateBookWithAuthor(ctx context.Context, tx *sql.Tx, bookID, authorID int) error {
	statement := `INSERT INTO book_authors (book_id, author_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := tx.ExecContext(ctx, statement, bookID, authorID)
	if err != nil {
		b.Logger.Error("Error adding author association", "error", err)
		return err
	}

	return nil
}

func (r *AuthorRepositoryImpl) GetAllBooksByAuthors(userID int) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	// Single query to get all book details by authors
	query := `
	SELECT r.id, r.title, r.subtitle, r.description, r.language, r.page_count, r.publish_date,
	       r.image_link, r.notes, r.created_at, r.last_updated, r.isbn_10, r.isbn_13,
	       a.name AS author_name,
	       json_agg(DISTINCT g.name) AS genres,
	       json_agg(DISTINCT f.format_type) AS formats,
	       json_agg(DISTINCT t.name) AS tags
	FROM books r
	INNER JOIN book_authors ba ON r.id = ba.book_id
	INNER JOIN authors a ON ba.author_id = a.id
	INNER JOIN user_books ub ON r.id = ub.book_id
	LEFT JOIN book_genres bg ON r.id = bg.book_id
	LEFT JOIN genres g ON bg.genre_id = g.id
	LEFT JOIN book_formats bf ON r.id = bf.book_id
	LEFT JOIN formats f ON bf.format_id = f.id
	LEFT JOIN book_tags bt ON r.id = bt.book_id
	LEFT JOIN tags t ON bt.tag_id = t.id
	WHERE ub.user_id = $1
	GROUP BY r.id, a.name`

	rows, err := r.DB.QueryContext(ctx, query, userID)
	if err != nil {
		r.Logger.Error("Error retrieving books by authors", "error", err)
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
			r.Logger.Error("Error scanning book by author", "error", err)
			return nil, err
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(genresJSON, &book.Genres); err != nil {
			r.Logger.Error("Error unmarshalling genres JSON", "error", err)
			return nil, err
		}
		if err := json.Unmarshal(formatsJSON, &book.Formats); err != nil {
			r.Logger.Error("Error unmarshalling formats JSON", "error", err)
			return nil, err
		}
		if err := json.Unmarshal(tagsJSON, &book.Tags); err != nil {
			r.Logger.Error("Error unmarshalling tags JSON", "error", err)
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
		r.Logger.Error("Error with rows", "error", err)
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

func (r *AuthorRepositoryImpl) GetAuthorIDByName(ctx context.Context, tx *sql.Tx, authorName string, authorID *int) error {
	err := tx.QueryRowContext(ctx, `SELECT id FROM authors WHERE name = $1`, authorName).Scan(authorID)
	return err
}

func (r *AuthorRepositoryImpl) GetAuthorsForBook(bookID int) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	query := `
	SELECT array_to_json(array_agg(a.name))
	FROM authors a
	JOIN book_authors ba ON a.id = ba.author_id
	WHERE ba.book_id = $1`

	var authorsJSON []byte
	err := r.DB.QueryRowContext(ctx, query, bookID).Scan(&authorsJSON)
	if err != nil {
		r.Logger.Error("Error fetching authors for book", "error", err)
		return nil, err
	}

	// Unmarshal the JSON array of authors
	var authors []string
	if err := json.Unmarshal(authorsJSON, &authors); err != nil {
		r.Logger.Error("Error unmarshalling authors JSON", "error", err)
		return nil, err
	}

	return authors, nil
}

func (r *AuthorRepositoryImpl) GetAuthorsForBooks(bookIDs []int) (map[int][]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	var rows *sql.Rows
	var err error

	// Use prepared statement if available
	if r.getAuthorsForBooksStmt != nil {
		rows, err = r.getAuthorsForBooksStmt.QueryContext(ctx, pq.Array(bookIDs))
		if err != nil {
			r.Logger.Error("Error fetching authors for books using prepared statement", "error", err)
			return nil, err
		}
	} else {
		// Fallback if the prepared statement is unavailable
		query := `
		SELECT ba.book_id, a.name
		FROM authors a
		JOIN book_authors ba ON a.id = ba.author_id
		WHERE ba.book_id = ANY($1)`

		rows, err = r.DB.QueryContext(ctx, query, pq.Array(bookIDs))
		if err != nil {
			r.Logger.Error("Error fetching authors for books using fallback query", "error", err)
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
			r.Logger.Error("Error scanning author name", "error", err)
			return nil, err
		}
		authorsByBook[bookID] = append(authorsByBook[bookID], authorName)
	}

	return authorsByBook, nil
}

func (r *AuthorRepositoryImpl) GetBooksByAuthor(authorName string) ([]Book, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	// Fetch all books by the author
	query := `
	SELECT r.id, r.title, r.subtitle, r.description, r.language, r.page_count, r.publish_date,
	       r.image_link, r.notes, r.created_at, r.last_updated, r.isbn_10, r.isbn_13
	FROM books b
	INNER JOIN book_authors ba ON b.id = ba.book_id
	INNER JOIN authors a ON ba.author_id = a.id
	WHERE a.name = $1`

	rows, err := r.DB.QueryContext(ctx, query, authorName)
	if err != nil {
		r.Logger.Error("Error retrieving books by author", "error", err)
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
			r.Logger.Error("Error scanning book", "error", err)
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

	authorRows, err := r.DB.QueryContext(ctx, authorQuery, pq.Array(bookIDs))
	if err != nil {
		r.Logger.Error("Error retrieving authors for books", "error", err)
		return nil, err
	}
	defer authorRows.Close()

	for authorRows.Next() {
		var bookID int
		var authorName string
		if err := authorRows.Scan(&bookID, &authorName); err != nil {
			r.Logger.Error("Error scanning author", "error", err)
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

	genreRows, err := r.DB.QueryContext(ctx, genreQuery, pq.Array(bookIDs))
	if err != nil {
		r.Logger.Error("Error retrieving genres for books", "error", err)
		return nil, err
	}
	defer genreRows.Close()

	for genreRows.Next() {
		var bookID int
		var genreName string
		if err := genreRows.Scan(&bookID, &genreName); err != nil {
			r.Logger.Error("Error scanning genre", "error", err)
			return nil, err
		}
		bookIDMap[bookID].Genres = append(bookIDMap[bookID].Genres, genreName)
	}

	if err = rows.Err(); err != nil {
		r.Logger.Error("GetBooksByAuthor - Error with rows", "error", err)
		return nil, err
	}

	return books, nil
}

func (r *AuthorRepositoryImpl) BatchInsertAuthors(ctx context.Context, tx *sql.Tx, bookID int, authors []string) error {
	// Ensure authors slice is not empty
	if len(authors) == 0 {
			r.Logger.Warn("No authors provided for book", "bookID", bookID)
			return nil // Early return as there's nothing to insert
	}

	authorIDMap := make(map[string]int) // Store author name -> authorID

	for i, author := range authors {
			var authorID int
			// Log author processing step
			r.Logger.Info("Processing author", "index", i, "author", author)

			// Skip invalid author names
			if author == "" {
					r.Logger.Warn("Skipping empty author name", "bookID", bookID, "index", i)
					continue
			}

			// Check if the author already exists
			err := tx.QueryRowContext(ctx, `SELECT id FROM authors WHERE name = $1`, author).Scan(&authorID)
			if err != nil {
					if err == sql.ErrNoRows {
							// Insert new author
							r.Logger.Info("Inserting new author", "author", author)
							err = tx.QueryRowContext(ctx, `INSERT INTO authors (name) VALUES ($1) RETURNING id`, author).Scan(&authorID)
							if err != nil {
									r.Logger.Error("Failed to insert new author", "author", author, "error", err)
									return fmt.Errorf("error inserting new author: %s, err: %w", author, err)
							}
					} else {
							r.Logger.Error("Error querying author", "author", author, "error", err)
							return fmt.Errorf("error querying author: %s, err: %w", author, err)
					}
			}

			// Avoid inserting duplicate author associations
			if _, exists := authorIDMap[author]; exists {
					r.Logger.Warn("Skipping duplicate author", "author", author)
					continue
			}

			authorIDMap[author] = authorID

			// Insert the book-author association
			r.Logger.Info("Inserting book-author association", "bookID", bookID, "authorID", authorID)
			_, err = tx.ExecContext(ctx, `INSERT INTO book_authors (book_id, author_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, bookID, authorID)
			if err != nil {
					r.Logger.Error("Error inserting book-author association", "bookID", bookID, "authorID", authorID, "error", err)
					return fmt.Errorf("error inserting book-author association for author: %s, err: %w", author, err)
			}
	}


	r.Logger.Info("BatchInsertAuthors completed successfully", "bookID", bookID)
	return nil
}


// Helper function to get the last name from a full name
func getLastName(fullName string) string {
	parts := strings.Fields(fullName)
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}
