package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/lib/pq"
	"github.com/lokeam/bravo-kilo/internal/dbconfig"
)

// BookRepository interface defines methods related to book operations
type BookRepository interface {
	InitPreparedStatements() error
	InsertBook(ctx context.Context, tx *sql.Tx, book Book, userID int, tagsJSON []byte) (int, error)
	GetBookByID(id int) (*Book, error)
	GetBookIdByTitle(title string) (int, error)
	GetAllBooksByUserID(userID int) ([]Book, error)
	AddBookToUser(tx *sql.Tx, userID, bookID int) error
	IsUserBookOwner(userID, bookID int) (bool, error)
	UpdateBook(ctx context.Context, tx *sql.Tx, book Book, tagsJSON []byte) error
}

// BookRepositoryImpl implements BookRepository, separating SQL logic to `book_queries.go`
type BookRepositoryImpl struct {
	DB                         *sql.DB
	Logger                     *slog.Logger
	AuthorRepository           AuthorRepository
	GenreRepository            GenreRepository
	FormatRepository           FormatRepository
	insertBookStmt             *sql.Stmt
	getBookByIDStmt            *sql.Stmt
	addBookToUserStmt          *sql.Stmt
	getBookIdByTitleStmt       *sql.Stmt
	getAllBooksByUserIDStmt    *sql.Stmt
	isUserBookOwnerStmt        *sql.Stmt
	updateBookStmt             *sql.Stmt
}


// NewBookRepository initializes and returns a new instance of BookRepositoryImpl
func NewBookRepository(
	db *sql.DB,
	logger *slog.Logger,
	authorRepo AuthorRepository,
	genreRepo GenreRepository,
	formatRemo FormatRepository,
	) (BookRepository, error) {
	if db == nil || logger == nil {
		return nil, fmt.Errorf("database or logger is nil")
	}

	if authorRepo == nil {
		return nil, fmt.Errorf("author repository is nil")
	}

	if genreRepo == nil {
		return nil, fmt.Errorf("genreRepo is nil")
	}

	if formatRemo == nil {
		return nil, fmt.Errorf("formatRepo is nil")
	}

	return &BookRepositoryImpl{
		DB:                db,
		Logger:            logger,
		AuthorRepository:  authorRepo,
		GenreRepository:   genreRepo,
		FormatRepository:  formatRemo,
	}, nil
}

func (r *BookRepositoryImpl) InitPreparedStatements() error {
	var err error

	// DEBUG Prepared insert statement for books
	r.insertBookStmt, err = r.DB.Prepare(`
		INSERT INTO books (title, subtitle, description, language, page_count, publish_date, image_link, notes, tags, created_at, last_updated, isbn_10, isbn_13)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING id`)
	if err != nil {
		r.Logger.Error("Error preparing insertBookStmt", "error", err)
		return err
	}

	// Prepared insert statement for books
	r.getBookByIDStmt, err = r.DB.Prepare(`
	SELECT id, title, subtitle, description, language, page_count, publish_date,
		image_link, notes, created_at, last_updated, isbn_10, isbn_13 FROM books WHERE id = $1`)
	if err != nil {
		r.Logger.Error("Error preparing getBookByIDStmt", "error", err)
		return fmt.Errorf("failed to prepare getBookByIDStmt: %w", err)
	}

	// Prepared insert statement for adding books to user
	r.addBookToUserStmt, err = r.DB.Prepare(`INSERT INTO user_books (user_id, book_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`)
	if err != nil {
		r.Logger.Error("Error preparing addBookToUserStmt", "error", err)
		return fmt.Errorf("failed to prepare addBookToUserStmt: %w", err)
	}

	// Prepared select statement for getting book ID by title
	r.getBookIdByTitleStmt, err = r.DB.Prepare(`SELECT id FROM books WHERE title = $1`)
	if err != nil {
		r.Logger.Error("Error preparing getBookIdByTitleStmt", "error", err)
		return fmt.Errorf("failed to prepare getBookIdByTitleStmt: %w", err)
	}

	// Prepared select statement for GetAllBooksByUserID
	r.getAllBooksByUserIDStmt, err = r.DB.Prepare(`
	SELECT DISTINCT b.id, b.title, b.subtitle, b.description, b.language, b.page_count,
			b.publish_date, b.image_link, b.notes, b.created_at, b.last_updated,
			b.isbn_10, b.isbn_13
	FROM books b
	INNER JOIN user_books ub ON b.id = ub.book_id
	WHERE ub.user_id = $1`)
	if err != nil {
		r.Logger.Error("Error preparing getAllBooksByUserIDStmt", "error", err)
		return fmt.Errorf("failed to prepare getAllBooksByUserIDStmt: %w", err)
	}

	// Prepared statement for updating a book
	r.updateBookStmt, err = r.DB.Prepare(`
		UPDATE books SET title=$1, subtitle=$2, description=$3, language=$4, page_count=$5,
			publish_date=$6, image_link=$7, notes=$8, tags=$9, last_updated=$10,
			isbn_10=$11, isbn_13=$12 WHERE id=$13`)
	if err != nil {
		r.Logger.Error("Error preparing updateBookStmt", "error", err)
		return err
	}

	return nil
}

func (r *BookRepositoryImpl) InsertBook(ctx context.Context, tx *sql.Tx, book Book, userID int, tagsJSON []byte) (int, error) {
	var newId int
	formattedPublishDate := formatPublishDate(book.PublishDate)

	// Insert book into books table
	err := tx.StmtContext(ctx, r.insertBookStmt).QueryRowContext(ctx,
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

	if err != nil {
		r.Logger.Error("Error inserting book", "error", err)
		return 0, err
	}

	// Associate book with the user
	err = r.addBookToUser(tx, userID, newId)
	if err != nil {
		r.Logger.Error("Error associating book with user", "error", err)
		return 0, err
	}

	return newId, nil
}

func (r *BookRepositoryImpl) GetBookByID(id int) (*Book, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	var rows *sql.Rows
	var err error

	// Check if prepared statement is available
	if r.getBookByIDStmt != nil {
			r.Logger.Info("Using prepared statement for fetching book by ID")
			rows, err = r.getBookByIDStmt.QueryContext(ctx, id)
	} else {
			r.Logger.Warn("Prepared statement for fetching book by ID is not available. Falling back to raw SQL query")
			query := `SELECT id, title, subtitle, description, language, page_count, publish_date,
								image_link, notes, created_at, last_updated, isbn_10, isbn_13 FROM books WHERE id = $1`
			rows, err = r.DB.QueryContext(ctx, query, id)
	}

	if err != nil {
			r.Logger.Error("Error fetching book by ID from the database", "bookID", id, "error", err)
			return nil, fmt.Errorf("failed to fetch book by ID %d: %w", id, err)
	}
	defer rows.Close()

	var book Book
	var tagsJSON []byte

	// Check if rows were returned
	if rows.Next() {
			// Scan the data into the book struct
			err = rows.Scan(
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
			)

			if err != nil {
					r.Logger.Error("Error scanning book data", "bookID", id, "error", err)
					return nil, fmt.Errorf("failed to scan book data for ID %d: %w", id, err)
			}

			// Process tags JSON
			if len(tagsJSON) > 0 {
					err = json.Unmarshal(tagsJSON, &book.Tags)
					if err != nil {
							r.Logger.Error("Error unmarshaling tags JSON", "bookID", id, "error", err)
							return nil, fmt.Errorf("failed to unmarshal tags for book ID %d: %w", id, err)
					}
			}
	} else {
			// No rows returned, book not found
			r.Logger.Warn("Book not found", "bookID", id)
			return nil, fmt.Errorf("book with ID %d not found", id)
	}

	// Batch fetch related details like authors, genres, formats, etc.
	err = r.batchFetchBookDetails(ctx, []int{book.ID}, map[int]*Book{book.ID: &book})
	if err != nil {
			return nil, fmt.Errorf("failed to fetch additional details for book ID %d: %w", id, err)
	}

	book.IsInLibrary = true
	return &book, nil
}

func (r *BookRepositoryImpl) GetBookIdByTitle(title string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	var bookID int
	var err error

	if r.getBookIdByTitleStmt != nil {
		r.Logger.Info("Using prepared statement for fetching book ID by title")
		err = r.getBookIdByTitleStmt.QueryRowContext(ctx, title).Scan(&bookID)
	} else {
		// Fallback to raw SQL query if prepared statement is unavailable
		r.Logger.Warn("Prepared statement for fetching book ID by title is not available. Falling back to raw SQL query")
		statement := `SELECT id FROM books WHERE title = $1`
		err = r.DB.QueryRowContext(ctx, statement, title).Scan(&bookID)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			r.Logger.Warn("No book found with the given title", "title", title)
			return 0, nil
		}
		r.Logger.Error("Error fetching book ID by title", "error", err)
		return 0, err
	}

	return bookID, nil
}

func (r *BookRepositoryImpl) GetAllBooksByUserID(userID int) ([]Book, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	var rows *sql.Rows
	var err error

	// Check if prepared statement is available, otherwise try re-initializing
	if r.getAllBooksByUserIDStmt == nil {
		r.Logger.Warn("Prepared statement for retrieving books by user ID is nil, attempting to re-initialize")
		err = r.InitPreparedStatements()
		if err != nil {
			r.Logger.Error("Failed to re-initialize prepared statements", "error", err)
			return nil, fmt.Errorf("failed to initialize prepared statements: %w", err)
		}
	}

	// Now attempt to use the prepared statement
	if r.getAllBooksByUserIDStmt != nil {
		r.Logger.Info("Using prepared statement for retrieving books by user ID")
		rows, err = r.getAllBooksByUserIDStmt.QueryContext(ctx, userID)
		if err != nil {
			r.Logger.Error("Error executing getAllBooksByUserIDStmt", "error", err)
			return nil, fmt.Errorf("failed to execute prepared statement: %w", err)
		}
	} else {
		// Fallback to raw SQL query if prepared statement is unavailable
		r.Logger.Warn("Prepared statement for retrieving books by user ID is still uninitialized, using fallback query")
		query := `
			SELECT b.id, b.title, b.subtitle, b.description, b.language, b.page_count, b.publish_date,
						 b.image_link, b.notes, b.created_at, b.last_updated, b.isbn_10, b.isbn_13
			FROM books b
			INNER JOIN user_books ub ON b.id = ub.book_id
			WHERE ub.user_id = $1`
		rows, err = r.DB.QueryContext(ctx, query, userID)
		if err != nil {
			r.Logger.Error("Error executing fallback query for GetAllBooksByUserID", "error", err)
			return nil, fmt.Errorf("failed to execute fallback query: %w", err)
		}
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
			r.Logger.Error("Error scanning book", "error", err)
			return nil, fmt.Errorf("failed to scan book row: %w", err)
		}

		book.IsInLibrary = true
		bookIDMap[book.ID] = &book
		bookIDs = append(bookIDs, book.ID)
	}

	// Batch Fetch authors, formats, genres, and tags
	if err := r.batchFetchBookDetails(ctx, bookIDs, bookIDMap); err != nil {
		return nil, fmt.Errorf("failed to fetch additional book details: %w", err)
	}

	// Collect books from map into a slice, check for empty fields
	var books []Book
	for _, book := range bookIDMap {
		book.EmptyFields, book.HasEmptyFields = r.findEmptyFields(book)
		books = append(books, *book)
	}

	return books, nil
}

func (r *BookRepositoryImpl) UpdateBook(ctx context.Context, tx *sql.Tx, book Book, tagsJSON []byte) error {

	// Check if the prepared statement is available
	if r.updateBookStmt != nil {
		_, err := tx.StmtContext(ctx, r.updateBookStmt).ExecContext(ctx,
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
			r.Logger.Error("Error updating book using prepared statement", "error", err)
			return err
		}
	} else {
		// Fallback to raw SQL query
		r.Logger.Warn("Prepared statement is nil, using raw SQL query")
		query := `UPDATE books SET title=$1, subtitle=$2, description=$3, language=$4, page_count=$5,
				  publish_date=$6, image_link=$7, notes=$8, tags=$9, last_updated=$10,
				  isbn_10=$11, isbn_13=$12 WHERE id=$13`

		_, err := tx.ExecContext(ctx, query,
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
			r.Logger.Error("Error updating book using raw SQL", "error", err)
			return err
		}
	}


	return nil
}


func (r *BookRepositoryImpl) AddBookToUser(tx *sql.Tx, userID, bookID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	// Use prepared statement if available
	if r.addBookToUserStmt != nil {
		r.Logger.Info("Using prepared statement for adding a book to user")
		_, err := tx.StmtContext(ctx, r.addBookToUserStmt).ExecContext(ctx, userID, bookID)
		if err != nil {
			r.Logger.Error("Error adding book to user using prepared statement", "error", err)
			return err
		}
	} else {
		// Fallback to raw SQL query
		r.Logger.Warn("Prepared statement for adding a book to user not available, falling back to raw SQL query")
		statement := `INSERT INTO user_books (user_id, book_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
		_, err := tx.ExecContext(ctx, statement, userID, bookID)
		if err != nil {
			r.Logger.Error("Error adding book to user using raw SQL query", "error", err)
			return err
		}
	}

	return nil
}

func (r *BookRepositoryImpl) IsUserBookOwner(userID, bookID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	var exists bool
	var err error

	// Use prepared statement if unavailable
	if r.isUserBookOwnerStmt != nil {
		err = r.isUserBookOwnerStmt.QueryRowContext(ctx, userID, bookID).Scan(&exists)
		if err != nil {
			r.Logger.Error("Error checking book ownership using prepared statement", "error", err)
			return false, err
		}
	} else {
		// Fallback if prepared statement is unavailable
		query := `SELECT EXISTS(SELECT 1 FROM user_books WHERE user_id = $1 AND book_id = $2)`
		err = r.DB.QueryRowContext(ctx, query, userID, bookID).Scan(&exists)
		if err != nil {
			r.Logger.Error("Error checking book ownership using fallback query", "error", err)
			return false, err
		}
	}

	return exists, nil
}


func (r *BookRepositoryImpl) addBookToUser(tx *sql.Tx, userID, bookID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	if r.addBookToUserStmt != nil {
		_, err := tx.StmtContext(ctx, r.addBookToUserStmt).ExecContext(ctx, userID, bookID)
		if err != nil {
			r.Logger.Error("Error adding book to user using prepared statement", "error", err)
			return err
		}
	} else {
		// Fallback to raw SQL query if prepared statement is unavailable
		_, err := tx.ExecContext(ctx, `INSERT INTO user_books (user_id, book_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, userID, bookID)
		if err != nil {
			r.Logger.Error("Error adding book to user", "error", err)
			return err
		}
	}

	return nil
}

func (r *BookRepositoryImpl) batchFetchBookDetails(ctx context.Context, bookIDs []int, bookIDMap map[int]*Book) error {
	query := `
	SELECT
		b.id AS book_id,
		ARRAY_AGG(DISTINCT a.name) AS author_names,
		ARRAY_AGG(DISTINCT g.name) AS genre_names,
		ARRAY_AGG(DISTINCT f.format_type) AS format_types,
		b.tags
	FROM books b
	LEFT JOIN book_authors ba ON b.id = ba.book_id
	LEFT JOIN authors a ON ba.author_id = a.id
	LEFT JOIN book_genres bg ON b.id = bg.book_id
	LEFT JOIN genres g ON bg.genre_id = g.id
	LEFT JOIN book_formats bf ON b.id = bf.book_id
	LEFT JOIN formats f ON bf.format_id = f.id
	WHERE b.id = ANY($1)
	GROUP BY b.id, b.tags`

	rows, err := r.DB.QueryContext(ctx, query, pq.Array(bookIDs))
	if err != nil {
		r.Logger.Error("Error batch fetching book details", "error", err)
		return err
	}
	defer rows.Close()

	// Processing the result rows
	for rows.Next() {
		var bookID int
		var authorNames, genreNames, formatTypes pq.StringArray
		var tagsJSON []byte

		// Scan the row
		if err := rows.Scan(&bookID, &authorNames, &genreNames, &formatTypes, &tagsJSON); err != nil {
			r.Logger.Error("Error scanning book details", "error", err)
			return err
		}

		// Fetch book from map
		book := bookIDMap[bookID]

		// Unmarshal tags if needed
		if len(tagsJSON) > 0 && len(book.Tags) == 0 {
			if err := json.Unmarshal(tagsJSON, &book.Tags); err != nil {
				r.Logger.Error("Error unmarshalling tags JSON", "error", err)
				return err
			}
		}

		// Deduplicated authors, genres, and formats are added
		book.Authors = append(book.Authors, authorNames...)
		book.Genres = append(book.Genres, genreNames...)
		book.Formats = append(book.Formats, formatTypes...)
	}

	// Check for errors after row iteration
	if err = rows.Err(); err != nil {
		r.Logger.Error("Error with rows in batch fetch", "error", err)
		return err
	}

	return nil
}

// Helper fn function for Insert Book to format publish date
func formatPublishDate(dateStr string) string {
	// If publish date only lists year, append "-01-01"
	if len(dateStr) == 4 {
		return dateStr + "-01-01"
	}
	return dateStr
}

// Helper fn for GetAllBooksByUserID
func (r *BookRepositoryImpl) findEmptyFields(book *Book) ([]string, bool) {
	// Define a slice of field checks
	checks := []struct {
		value interface{}
		name  string
	}{
		{book.Title, "title"},
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
