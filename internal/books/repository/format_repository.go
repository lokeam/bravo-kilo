package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/lib/pq"
	"github.com/lokeam/bravo-kilo/internal/dbconfig"
)

type FormatRepository interface {
	InitPreparedStatements() error
	AddFormats(tx *sql.Tx, ctx context.Context, bookID int, formatTypes []string) error
	AddOrGetFormatID(ctx context.Context, tx *sql.Tx, format string) (int, error)
	AssociateFormatWithBooks(ctx context.Context, tx *sql.Tx, bookID, authorID int) error
	GetAllBooksByFormat(userID int) (map[string][]Book, error)
	GetFormats(ctx context.Context, bookID int) ([]string, error)
	GetOrInsertFormat(ctx context.Context, formatType string) (int, error)
	RemoveSpecificFormats(ctx context.Context, bookID int, formats []string) error
}

type FormatRepositoryImpl struct {
	DB                        *sql.DB
	Logger                    *slog.Logger
	getAllBooksByFormatStmt   *sql.Stmt
	getFormatsStmt            *sql.Stmt
}

func NewFormatRepository(db *sql.DB, logger *slog.Logger) (FormatRepository, error) {
	if db == nil || logger == nil {
		return nil, fmt.Errorf("database or logger is nil")
	}

	return &FormatRepositoryImpl{
		DB:          db,
		Logger:      logger,
	}, nil
}

func (r *FormatRepositoryImpl) InitPreparedStatements() error {
	var err error

	// Prepared statement for GetAllBooksByFormat
	r.getAllBooksByFormatStmt, err = r.DB.Prepare(`
	SELECT
		b.id, b.title, b.subtitle, b.description, b.language, b.page_count, b.publish_date,
		b.image_link, b.notes, b.created_at, b.last_updated, b.isbn_10, b.isbn_13,
		f.format_type,
		array_to_json(array_agg(DISTINCT a.name)) as authors,
		array_to_json(array_agg(DISTINCT g.name)) as genres,
		COALESCE(b.tags, '[]') AS tags
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
	r.getFormatsStmt, err = r.DB.Prepare(`
		SELECT f.format_type
		FROM formats f
		JOIN book_formats bf ON f.id = bf.format_id
		WHERE bf.book_id = $1`)
	if err != nil {
		return err
	}

	return nil
}

// AddFormats method using the new getOrInsertFormat helper
func (r *FormatRepositoryImpl) AddFormats(tx *sql.Tx, ctx context.Context, bookID int, formatTypes []string) error {
	if len(formatTypes) == 0 {
		return nil
	}

	valueStrings := make([]string, len(formatTypes))
	valueArgs := make([]interface{}, 0, len(formatTypes)*2)

	for i, formatType := range formatTypes {
		// Get or insert the format
		formatID, err := r.GetOrInsertFormat(ctx, formatType)
		if err != nil {
			return err
		}

		// Prepare the value string for insertion into the book_formats table
		valueStrings[i] = fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2)
		valueArgs = append(valueArgs, bookID, formatID)
	}

	statement := fmt.Sprintf("INSERT INTO book_formats (book_id, format_id) VALUES %s ON CONFLICT DO NOTHING", strings.Join(valueStrings, ","))

	// Use the passed-in transaction
	_, err := tx.ExecContext(ctx, statement, valueArgs...)
	if err != nil {
		r.Logger.Error("Error adding formats", "bookID", bookID, "error", err)
		return err
	}

	return nil
}

func (b *FormatRepositoryImpl) AddOrGetFormatID(ctx context.Context, tx *sql.Tx, format string) (int, error) {
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

func (r *FormatRepositoryImpl) GetAllBooksByFormat(userID int) (map[string][]Book, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	var rows *sql.Rows
	var err error

	// If prepared statement is nil, try reinitializing it
	if r.getAllBooksByFormatStmt == nil {
		r.Logger.Warn("getAllBooksByFormatStmt is nil, attempting to reinitialize")
		err = r.InitPreparedStatements()
		if err != nil {
			r.Logger.Error("Failed to reinitialize prepared statements", "error", err)
			return nil, fmt.Errorf("failed to reinitialize prepared statements: %w", err)
		}
	}

	// Use the prepared statement if available
	if r.getAllBooksByFormatStmt != nil {
		r.Logger.Info("Using prepared statement for retrieving books by format")
		rows, err = r.getAllBooksByFormatStmt.QueryContext(ctx, userID)
	} else {
		// Fallback to raw SQL query if reinitialization fails
		r.Logger.Warn("Prepared statement for retrieving books by format is still unavailable, using fallback query")
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
		rows, err = r.DB.QueryContext(ctx, query, userID)
	}

	// Handle query execution error
	if err != nil {
		r.Logger.Error("Error retrieving books by format", "error", err)
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
			r.Logger.Error("Error scanning book", "error", err)
			return nil, err
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(authorsJSON, &book.Authors); err != nil {
			r.Logger.Error("Error unmarshalling authors JSON", "error", err)
			return nil, err
		}
		if err := json.Unmarshal(genresJSON, &book.Genres); err != nil {
			r.Logger.Error("Error unmarshalling genres JSON", "error", err)
			return nil, err
		}
		if err := json.Unmarshal(tagsJSON, &book.Tags); err != nil {
			r.Logger.Error("Error unmarshalling tags JSON", "error", err)
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
		r.Logger.Error("Error with rows", "error", err)
		return nil, err
	}

	return booksByFormat, nil
}

func (b *FormatRepositoryImpl) GetFormats(ctx context.Context, bookID int) ([]string, error) {
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

func (r *FormatRepositoryImpl) GetOrInsertFormat(ctx context.Context, formatType string) (int, error) {
	var formatID int

	// First, check if the format already exists
	err := r.DB.QueryRowContext(ctx, `SELECT id FROM formats WHERE format_type = $1`, formatType).Scan(&formatID)

	if err == sql.ErrNoRows {
		// Insert the new format if it doesn't exist
		err = r.DB.QueryRowContext(ctx, `INSERT INTO formats (format_type) VALUES ($1) RETURNING id`, formatType).Scan(&formatID)
		if err != nil {
			r.Logger.Error("Error inserting new format", "error", err)
			return 0, err
		}
	} else if err != nil {
		r.Logger.Error("Error checking format existence", "error", err)
		return 0, err
	}

	return formatID, nil
}

func (b *FormatRepositoryImpl) AssociateFormatWithBooks(ctx context.Context, tx *sql.Tx, bookID, authorID int) error {
	statement := `INSERT INTO book_authors (book_id, author_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := tx.ExecContext(ctx, statement, bookID, authorID)
	if err != nil {
		b.Logger.Error("Error adding author association", "error", err)
		return err
	}

	return nil
}

func (b *FormatRepositoryImpl) RemoveSpecificFormats(ctx context.Context, bookID int, formats []string) error {
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


