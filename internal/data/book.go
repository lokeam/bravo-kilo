package data

import (
	"bravo-kilo/internal/data/collections"
	"bravo-kilo/internal/utils"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
)

type BookModel struct {
	DB     *sql.DB
	Logger *slog.Logger
	Author *AuthorModel
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
	ImageLinks      []string   `json:"imageLinks"`
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

	imageLinksJSON, err := json.Marshal(book.ImageLinks)
	if err != nil {
		b.Logger.Error("Error marshalling image links to JSON", "error", err)
		return 0, err
	}
	tagsJSON, err := json.Marshal(book.Tags)
	if err != nil {
		b.Logger.Error("Error marshalling tags to JSON", "error", err)
		return 0, err
	}

	// Insert book into books table
	statement := `INSERT INTO books (title, subtitle, description, language, page_count, publish_date, image_links, notes, tags, created_at, last_updated, isbn_10, isbn_13)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING id`

	err = tx.QueryRowContext(ctx, statement,
		book.Title,
		book.Subtitle,
		book.Description,
		book.Language,
		book.PageCount,
		book.PublishDate,
		imageLinksJSON,
		book.Notes,
		tagsJSON,
		time.Now(),
		time.Now(),
		book.ISBN10,
		book.ISBN13,
	).Scan(&newId)

	if err != nil {
		b.Logger.Error("Book Model - Error inserting book", "error", err)
		return 0, err
	}

	// Use sets to eliminate duplicates
	authorsSet := collections.NewSet()
	genresSet := collections.NewSet()

	for _, author := range book.Authors {
		authorsSet.Add(author)
	}

	for _, genre := range book.Genres {
		genresSet.Add(genre)
	}

	// Batch insert authors and associate them with the book
	for _, author := range authorsSet.Elements() {
		var authorID int
		err = tx.QueryRowContext(ctx, `SELECT id FROM authors WHERE name = $1`, author).Scan(&authorID)
		if err != nil {
			if err == sql.ErrNoRows {
				err = tx.QueryRowContext(ctx, `INSERT INTO authors (name) VALUES ($1) RETURNING id`, author).Scan(&authorID)
				if err != nil {
					b.Logger.Error("Error inserting author", "error", err)
					return 0, err
				}
			} else {
				b.Logger.Error("Error querying author", "error", err)
				return 0, err
			}
		}

		_, err = tx.ExecContext(ctx, `INSERT INTO book_authors (book_id, author_id) VALUES ($1, $2)`, newId, authorID)
		if err != nil {
			b.Logger.Error("Error inserting book_author association", "error", err)
			return 0, err
		}
	}

	// Batch insert genres and associate them with the book
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

	// Associate the book with the user
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

func (b *BookModel) AddBookToUser(tx *sql.Tx, userID, bookID int) error {
	statement := `INSERT INTO user_books (user_id, book_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := tx.ExecContext(context.Background(), statement, userID, bookID)
	if err != nil {
		b.Logger.Error("Error adding book to user", "error", err)
		return err
	}
	return nil
}

func (b *BookModel) GetBookByID(id int) (*Book, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// Batch query to fetch book, authors, genres, and formats
	query := `
		WITH book_data AS (
			SELECT
				b.id, b.title, b.subtitle, b.description, b.language, b.page_count, b.publish_date,
				b.image_links, b.notes, b.tags, b.created_at, b.last_updated, b.isbn_10, b.isbn_13
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
			b.image_links, b.notes, b.tags, b.created_at, b.last_updated, b.isbn_10, b.isbn_13,
			a.name AS author_name, g.name AS genre_name, f.format_type AS format_type
		FROM book_data b
		LEFT JOIN authors_data a ON true
		LEFT JOIN genres_data g ON true
		LEFT JOIN formats_data f ON true
	`

	rows, err := b.DB.QueryContext(ctx, query, id)
	if err != nil {
		b.Logger.Error("Error fetching book by ID with batch queries", "error", err)
		return nil, err
	}
	defer rows.Close()

	var book Book
	var imageLinksJSON, tagsJSON []byte
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
			&imageLinksJSON,
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

	// Unmarshal JSON fields
	if err := json.Unmarshal(imageLinksJSON, &book.ImageLinks); err != nil {
		b.Logger.Error("Error unmarshalling image links JSON", "error", err)
		return nil, err
	}
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
	statement := `SELECT id FROM books WHERE title = $1`
	err := b.DB.QueryRowContext(ctx, statement, title).Scan(&bookID)
	if err != nil {
			if err == sql.ErrNoRows {
					b.Logger.Error("Book Model - No book found with the given title", "title", title)
					return 0, nil
			}
			b.Logger.Error("Book Model - Error fetching book ID by title", "error", err)
			return 0, err
	}

	return bookID, nil
}

func (b *BookModel) GetAuthorsForBooks(bookIDs []int) (map[int][]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
	SELECT ba.book_id, a.name
	FROM authors a
	JOIN book_authors ba ON a.id = ba.author_id
	WHERE ba.book_id = ANY($1)`

	rows, err := b.DB.QueryContext(ctx, query, pq.Array(bookIDs))
	if err != nil {
		b.Logger.Error("Error fetching authors for books", "error", err)
		return nil, err
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
	var exists bool

	query := `SELECT EXISTS(SELECT 1 FROM user_books WHERE user_id = $1 AND book_id = $2)`
	err := b.DB.QueryRow(query, userID, bookID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (b *BookModel) GetAllBooksByUserID(userID int) ([]Book, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// Fetch the list of books for the user
	bookQuery := `
		SELECT b.id, b.title, b.subtitle, b.description, b.language, b.page_count, b.publish_date,
					 b.image_links, b.notes, b.created_at, b.last_updated, b.isbn_10, b.isbn_13
		FROM books b
		INNER JOIN user_books ub ON b.id = ub.book_id
		WHERE ub.user_id = $1`

	rows, err := b.DB.QueryContext(ctx, bookQuery, userID)
	if err != nil {
		b.Logger.Error("Error retrieving books for user", "error", err)
		return nil, err
	}
	defer rows.Close()

	bookIDMap := make(map[int]*Book)
	var bookIDs []int
	for rows.Next() {
		var book Book
		var imageLinksJSON []byte

		if err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.Subtitle,
			&book.Description,
			&book.Language,
			&book.PageCount,
			&book.PublishDate,
			&imageLinksJSON,
			&book.Notes,
			&book.CreatedAt,
			&book.LastUpdated,
			&book.ISBN10,
			&book.ISBN13,
		); err != nil {
			b.Logger.Error("Error scanning book", "error", err)
			return nil, err
		}

		if len(imageLinksJSON) > 0 {
			if err := json.Unmarshal(imageLinksJSON, &book.ImageLinks); err != nil {
				b.Logger.Error("Error unmarshalling image links JSON", "error", err)
				return nil, err
			}
		}

		book.IsInLibrary = true
		bookIDMap[book.ID] = &book
		bookIDs = append(bookIDs, book.ID)
	}

	// Fetch authors, formats, genres, and tags in batches
	if err := b.batchFetchBookDetails(ctx, bookIDs, bookIDMap); err != nil {
		return nil, err
	}

	// Collect books from the map into a slice, check for empty fields
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
		{len(book.ImageLinks) == 0, "imageLinks"},
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
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	imageLinksJSON, err := json.Marshal(book.ImageLinks)
	if err != nil {
		b.Logger.Error("Error marshalling image links to JSON", "error", err)
		return err
	}
	tagsJSON, err := json.Marshal(book.Tags)
	if err != nil {
		b.Logger.Error("Error marshalling tags to JSON", "error", err)
		return err
	}

	statement := `UPDATE books SET title=$1, subtitle=$2, description=$3, language=$4, page_count=$5, publish_date=$6, image_links=$7, notes=$8, tags=$9, last_updated=$10, isbn_10=$11, isbn_13=$12 WHERE id=$13`
	_, err = b.DB.ExecContext(ctx, statement,
		book.Title,
		book.Subtitle,
		book.Description,
		book.Language,
		book.PageCount,
		book.PublishDate,
		imageLinksJSON,
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
	       b.image_links, b.notes, b.created_at, b.last_updated, b.isbn_10, b.isbn_13
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
		var imageLinksJSON []byte
		if err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.Subtitle,
			&book.Description,
			&book.Language,
			&book.PageCount,
			&book.PublishDate,
			&imageLinksJSON,
			&book.Notes,
			&book.CreatedAt,
			&book.LastUpdated,
			&book.ISBN10,
			&book.ISBN13,
		); err != nil {
			b.Logger.Error("Error scanning book", "error", err)
			return nil, err
		}

		// Unmarshal JSONB fields
		if err := json.Unmarshal(imageLinksJSON, &book.ImageLinks); err != nil {
			b.Logger.Error("Error unmarshalling image links JSON", "error", err)
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

	// Batch all books and their authors for a user in a single query
	query := `
	SELECT b.id, b.title, b.subtitle, b.description, b.language, b.page_count, b.publish_date,
				 b.image_links, b.notes, b.created_at, b.last_updated, b.isbn_10, b.isbn_13,
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

	rows, err := b.DB.QueryContext(ctx, query, userID)
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
		var imageLinksJSON []byte
		var genresJSON, formatsJSON []byte
		var tagsJSON []byte

		// Scan the result
		if err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.Subtitle,
			&book.Description,
			&book.Language,
			&book.PageCount,
			&book.PublishDate,
			&imageLinksJSON,
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

		// Debugging: Log the retrieved author name
		b.Logger.Info("Retrieved author", "authorName", authorName)

		// Unmarshal JSON fields
		if err := json.Unmarshal(imageLinksJSON, &book.ImageLinks); err != nil {
			b.Logger.Error("Error unmarshalling image links JSON", "error", err)
			return nil, err
		}
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

	return isbnSet, nil
}

// (Returns a HashSet)
func (b *BookModel) GetAllBooksISBN13(userID int) (*collections.Set, error) {
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

	return isbnSet, nil
}

// (Returns a HashSet)
func (b *BookModel) GetAllBooksTitles(userID int) (*collections.Set, error) {
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

	// Build query dynamically based on number of formatIDs
	valueStrings := make([]string, 0, len(formatIDs))
	valueArgs := make([]interface{}, 0, len(formatIDs)*2)

	for i, formatID := range formatIDs {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		valueArgs = append(valueArgs, bookID, formatID)
	}

	statement := fmt.Sprintf(`INSERT INTO book_formats (book_id, format_id) VALUES %s`, strings.Join(valueStrings, ","))
	_, err := b.DB.ExecContext(ctx, statement, valueArgs...)
	if err != nil {
		b.Logger.Error("Book Model - Error adding formats", "bookID", bookID, "formatIDs", formatIDs, "error", err)
		return err
	}

	return nil
}

func (b *BookModel) addOrGetFormatID(ctx context.Context, format string) (int, error) {
	var formatID int
	err := b.DB.QueryRowContext(ctx, "SELECT id FROM formats WHERE format_type = $1", format).Scan(&formatID)
	if err != nil {
		if err == sql.ErrNoRows {
			err = b.DB.QueryRowContext(ctx, "INSERT INTO formats (format_type) VALUES ($1) RETURNING id", format).Scan(&formatID)
			if err != nil {
				return 0, err
			}
		} else {
			return 0, err
		}
	}
	return formatID, nil
}

func (b *BookModel) GetAllBooksByFormat(userID int) (map[string][]Book, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
	SELECT
		b.id, b.title, b.subtitle, b.description, b.language, b.page_count, b.publish_date,
		b.image_links, b.notes, b.created_at, b.last_updated, b.isbn_10, b.isbn_13,
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
	GROUP BY b.id, f.format_type
	`

	rows, err := b.DB.QueryContext(ctx, query, userID)
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
		var imageLinksJSON, authorsJSON, genresJSON, tagsJSON []byte
		var formatType string

		if err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.Subtitle,
			&book.Description,
			&book.Language,
			&book.PageCount,
			&book.PublishDate,
			&imageLinksJSON,
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
		if err := json.Unmarshal(imageLinksJSON, &book.ImageLinks); err != nil {
			b.Logger.Error("Error unmarshalling image links JSON", "error", err)
			return nil, err
		}
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

	query := `
	SELECT f.format_type
	FROM formats f
	JOIN book_formats bf ON f.id = bf.format_id
	WHERE bf.book_id = $1`

	rows, err := b.DB.QueryContext(ctx, query, bookID)
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

	return formats, nil
}


// Helper fn for UpdateBook
func (b *BookModel) updateFormats(ctx context.Context, bookID int, newFormats []string) error {
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
	// Check if the genre already exists
	statement := `SELECT id FROM genres WHERE name = $1`
	err := b.DB.QueryRowContext(ctx, statement, genreName).Scan(&genreID)
	if err != nil && err != sql.ErrNoRows {
		b.Logger.Error("Error checking genre existence", "error", err)
		return 0, err
	}

	if genreID == 0 { // If the genre does not exist, insert it
		statement := `INSERT INTO genres (name) VALUES ($1) RETURNING id`
		err = b.DB.QueryRowContext(ctx, statement, genreName).Scan(&genreID)
		if err != nil {
			b.Logger.Error("Error inserting new genre", "error", err)
			return 0, err
		}
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

	query := `
	SELECT g.name
	FROM genres g
	JOIN book_genres bg ON g.id = bg.genre_id
	WHERE bg.book_id = $1`

	// Execute the query with the provided context
	rows, err := b.DB.QueryContext(ctx, query, bookID)
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
		b.Logger.Error("Error with rows", "error", err)
		return nil, err
	}

	return genres, nil
}

func (b *BookModel) GetAllBooksByGenres(ctx context.Context, userID int) (map[string]interface{}, error) {
	// Batch get all books and genres for a user's books
	query := `
	SELECT
		b.id, b.title, b.subtitle, b.description, b.language, b.page_count,
		b.publish_date, b.image_links, b.notes, b.created_at, b.last_updated,
		b.isbn_10, b.isbn_13, b.genres, json_agg(DISTINCT f.format_type) AS formats, -- Aggregate formats as JSON
		a.name AS author_name, b.tags
	FROM books b
	INNER JOIN book_genres bg ON b.id = bg.book_id
	INNER JOIN genres g ON bg.genre_id = g.id
	INNER JOIN user_books ub ON b.id = ub.book_id
	LEFT JOIN book_authors ba ON b.id = ba.book_id
	LEFT JOIN authors a ON ba.author_id = a.id
	LEFT JOIN book_formats bf ON b.id = bf.book_id -- Join book_formats table
	LEFT JOIN formats f ON bf.format_id = f.id     -- Join formats table
	WHERE ub.user_id = $1
	GROUP BY b.id, a.name, b.genres`  // Group by necessary fields

	rows, err := b.DB.QueryContext(ctx, query, userID)
	if err != nil {
		b.Logger.Error("Error retrieving books by genres", "error", err)
		return nil, err
	}
	defer rows.Close()

	// Store genres and books
	genres := []string{}
	booksByGenre := map[string][]Book{}

	for rows.Next() {
		var book Book
		var authorName string
		var imageLinksJSON, genresJSON, formatsJSON, tagsJSON []byte

		// Scan the row
		if err := rows.Scan(
			&book.ID, &book.Title, &book.Subtitle, &book.Description, &book.Language, &book.PageCount,
			&book.PublishDate, &imageLinksJSON, &book.Notes, &book.CreatedAt, &book.LastUpdated,
			&book.ISBN10, &book.ISBN13, &genresJSON, &formatsJSON, &authorName, &tagsJSON,
		); err != nil {
			b.Logger.Error("Error scanning book by genre", "error", err)
			return nil, err
		}

		// Unmarshal image links
		if err := json.Unmarshal(imageLinksJSON, &book.ImageLinks); err != nil {
			b.Logger.Error("Error unmarshalling image links JSON", "error", err)
			return nil, err
		}

		// Unmarshal genres
		if len(genresJSON) > 0 {
			if err := json.Unmarshal(genresJSON, &book.Genres); err != nil {
				b.Logger.Error("Error unmarshalling genres JSON", "error", err)
				return nil, err
			}
		}

		// Unmarshal formats
		if len(formatsJSON) > 0 {
			if err := json.Unmarshal(formatsJSON, &book.Formats); err != nil {
				b.Logger.Error("Error unmarshalling formats JSON", "error", err)
				return nil, err
			}
		}

		// Unmarshal tags
		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &book.Tags); err != nil {
				b.Logger.Error("Error unmarshalling tags JSON", "error", err)
				return nil, err
			}
		}

		// Assign the author
		if authorName != "" {
			book.Authors = append(book.Authors, authorName)
		}

		// Add genres to the list
		for _, genre := range book.Genres {
			if _, found := booksByGenre[genre]; !found {
				genres = append(genres, genre)
			}

			// Add book to the genre's list
			booksByGenre[genre] = append(booksByGenre[genre], book)
		}
	}

	// Check for row errors
	if err = rows.Err(); err != nil {
		b.Logger.Error("Error with rows", "error", err)
		return nil, err
	}

	// Sort genres alphabetically
	sort.Strings(genres)

	// Create the result map with index keys for genres
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

		// Prepare genreImgs array
		genreImgs := make([]string, len(booksByGenre[genre]))
		for j, book := range booksByGenre[genre] {
			if len(book.ImageLinks) > 0 {
				genreImgs[j] = book.ImageLinks[0]
			}
		}

		result[key] = map[string]interface{}{
			"genreImgs": genreImgs,
			"bookList":  booksByGenre[genre],
		}
	}

	return result, nil
}

// Helper fn for UpdateBook
func (b *BookModel) updateGenres(ctx context.Context, bookID int, newGenres []string) error {
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

