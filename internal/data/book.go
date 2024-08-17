package data

import (
	"bravo-kilo/internal/data/collections"
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"
)

type BookModel struct {
	DB     *sql.DB
	Logger *slog.Logger
	Author *AuthorModel
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

	b.Logger.Info("Inserted book with ID", "bookID", newId)

	// Insert authors and associate them with the book
	for _, author := range book.Authors {
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

	// Insert genres and associate them with the book
	for _, genre := range book.Genres {
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

func (b *BookModel) GetByID(id int) (*Book, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var book Book
	var imageLinksJSON, tagsJSON []byte
	statement := `SELECT id, title, subtitle, description, language, page_count, publish_date, image_links, notes, tags, created_at, last_updated, isbn_10, isbn_13 FROM books WHERE id = $1`
	row := b.DB.QueryRowContext(ctx, statement, id)
	err := row.Scan(
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
	)
	if err != nil {
			if err == sql.ErrNoRows {
					return nil, nil
			} else {
					b.Logger.Error("Book Model - Error fetching book by ID", "error", err)
					return nil, err
			}
	}

	// Unmarshal JSONB fields
	if err := json.Unmarshal(imageLinksJSON, &book.ImageLinks); err != nil {
			b.Logger.Error("Error unmarshalling image links JSON", "error", err)
			return nil, err
	}
	if err := json.Unmarshal(tagsJSON, &book.Tags); err != nil {
			b.Logger.Error("Error unmarshalling tags JSON", "error", err)
			return nil, err
	}

	// Fetch authors for the book
	authorsQuery := `
	SELECT a.name
	FROM authors a
	JOIN book_authors ba ON a.id = ba.author_id
	WHERE ba.book_id = $1`
	authorRows, err := b.DB.QueryContext(ctx, authorsQuery, book.ID)
	if err != nil {
		b.Logger.Error("Error fetching authors for book", "error", err)
		return nil, err
	}
	defer authorRows.Close()

	var authors []string
	for authorRows.Next() {
		var authorName string
		if err := authorRows.Scan(&authorName); err != nil {
			b.Logger.Error("Error scanning author name", "error", err)
			return nil, err
		}
		authors = append(authors, authorName)
	}
	book.Authors = authors

	// Fetch genres
	genres, err := b.GetGenres(book.ID)
	if err != nil {
		b.Logger.Error("Error fetching genres", "error", err)
		return nil, err
	}

	book.Genres = genres

	// Fetch formats for the book
	formatsQuery := `
	SELECT f.format_type
	FROM formats f
	JOIN book_formats bf ON f.id = bf.format_id
	WHERE bf.book_id = $1`
	formatRows, err := b.DB.QueryContext(ctx, formatsQuery, book.ID)
	if err != nil {
		b.Logger.Error("Error fetching formats for book", "error", err)
		return nil, err
	}
	defer formatRows.Close()

	var formats []string
	for formatRows.Next() {
		var formatType string
		if err := formatRows.Scan(&formatType); err != nil {
			b.Logger.Error("Error scanning format type", "error", err)
			return nil, err
		}
		formats = append(formats, formatType)
	}
	book.Formats = formats

	// Mark book as inLibary
	book.IsInLibrary = true;

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

func (b *BookModel) GetAuthorsForBook(bookID int) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	authorsQuery := `
	SELECT a.name
	FROM authors a
	JOIN book_authors ba ON a.id = ba.author_id
	WHERE ba.book_id = $1`
	rows, err := b.DB.QueryContext(ctx, authorsQuery, bookID)
	if err != nil {
			b.Logger.Error("Error fetching authors for book", "error", err)
			return nil, err
	}
	defer rows.Close()

	var authors []string
	for rows.Next() {
			var authorName string
			if err := rows.Scan(&authorName); err != nil {
					b.Logger.Error("Error scanning author name", "error", err)
					return nil, err
			}
			authors = append(authors, authorName)
	}

	return authors, nil
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

	// Remove old genres
	if err := b.RemoveGenres(book.ID); err != nil {
		b.Logger.Error("Book Model - Error removing genres", "error", err)
		return err
	}

	// Add new genres
	for _, genre := range book.Genres {
		genreID, err := b.addOrGetGenreID(genre)
		if err != nil {
			b.Logger.Error("Error getting genre ID", "error", err)
			return err
		}

		err = b.AddGenre(book.ID, genreID)
		if err != nil {
			b.Logger.Error("Error adding genre association", "error", err)
			return err
		}
	}

	// Remove old formats
	if err := b.RemoveFormats(book.ID); err != nil {
		b.Logger.Error("Book Model - Error removing formats", "error", err)
		return err
	}

	// Add new formats
	for _, format := range book.Formats {
		formatID, err := b.addOrGetFormatID(format)
		if err != nil {
			b.Logger.Error("Error getting format ID", "error", err)
			return err
		}

		err = b.AddFormat(book.ID, formatID)
		if err != nil {
			b.Logger.Error("Error adding format association", "error", err)
			return err
		}
	}

	// Update Authors
	if err := b.UpdateAuthors(book.ID, book.Authors); err != nil {
		b.Logger.Error("Book Model - Error updating authors for book", "error", "err")
		return err
	}

	return nil
}

func (b *BookModel) Delete(id int) error {
	// Start a new context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// Start transaction
	tx, err := b.DB.BeginTx(ctx, nil)
	if err != nil {
			b.Logger.Error("Book Model - Error starting transaction", "error", err)
			return err
	}

	// Roll back in case of error
	defer func() {
		if p:= recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// Delete assoc. entries in users table
	deleteUserBookStatement := `DELETE FROM user_books WHERE book_id = $1`
	_, err = tx.ExecContext(ctx, deleteUserBookStatement, id)
	if err != nil {
		b.Logger.Error("Book Model - Error deleting from user_books", "error", err)
		return err
	}

	// Delete assoc entries in genres table
	deleteBookGenresStatement := `DELETE FROM book_genres WHERE book_id = $1`
	_, err = tx.ExecContext(ctx, deleteBookGenresStatement, id)
	if err != nil {
		b.Logger.Error("Book Model - Error deleting from book_genres", "error", err)
		return err
	}

	// Delete assoc entries in the authors table
	deleteBookAuthorsStatement := `DELETE FROM book_authors WHERE book_id = $1`
	_, err = tx.ExecContext(ctx, deleteBookAuthorsStatement, id)
	if err != nil {
		b.Logger.Error("Book Model - Error deleting from book_authors", "error", err)
		return err
	}

	// Delete book from books table
	deleteBookStatement := `DELETE from books WHERE id = $1`
	_, err = tx.ExecContext(ctx, deleteBookStatement, id)
	if err != nil {
		b.Logger.Error("Book Model - Error deleting book", "error", err)
		return err
	}

	// Commit if we're all gtg
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

	query := `
	SELECT b.id, b.title, b.subtitle, b.description, b.language, b.page_count, b.publish_date, b.image_links, b.notes, b.created_at, b.last_updated, b.isbn_10, b.isbn_13
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

		// Fetch genres
		genres, err := b.GetGenres(book.ID)
		if err != nil {
			b.Logger.Error("Error fetching genres", "error", err)
			return nil, err
		}
		book.Genres = genres

		// Get authors for each book
		authorsForEachBookQuery := `
			SELECT a.name
			FROM authors a
			JOIN book_authors ba ON a.id = ba.author_id
			WHERE ba.book_id = $1`
		authorRows, err := b.DB.QueryContext(ctx, authorsForEachBookQuery, book.ID)
		if err != nil {
			b.Logger.Error("Error retrieving authors for book", "error", err)
			return nil, err
		}

		var authors []string
		for authorRows.Next() {
			var author string
			if err := authorRows.Scan(&author); err != nil {
			  b.Logger.Error("GetBooksByAuthor - Error scanning author", "error", err)
				return nil, err
			}
			authors = append(authors, author)
		}
		authorRows.Close()

		book.Authors = authors

		// Mark book as in Library
		book.IsInLibrary = true;

		books = append(books, book)
	}

	if err = rows.Err(); err != nil {
		b.Logger.Error("GetBooksByAuthor - Error with rows", "error", err)
		return nil, err
	}

	return books, nil
}

func (b *BookModel) GetAllBooksByUserID(userID int) ([]Book, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
	SELECT b.id, b.title, b.subtitle, b.description, b.language, b.page_count, b.publish_date, b.image_links, b.notes, b.created_at, b.last_updated, b.isbn_10, b.isbn_13
	FROM books b
	INNER JOIN user_books ub ON b.id = ub.book_id
	WHERE ub.user_id = $1`

	rows, err := b.DB.QueryContext(ctx, query, userID)
	if err != nil {
		b.Logger.Error("Error retrieving books for user", "error", err)
		return nil, err
	}
	defer rows.Close()

	var books []Book
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

		// Unmarshal JSONB fields, handle null values
		if len(imageLinksJSON) > 0 {
			if err := json.Unmarshal(imageLinksJSON, &book.ImageLinks); err != nil {
				b.Logger.Error("Error unmarshalling image links JSON", "error", err, "data", string(imageLinksJSON))
				return nil, err
			}
		}

		// Fetch authors for the book
		authors, err := b.GetAuthorsForBook(book.ID)
		if err != nil {
				b.Logger.Error("Error fetching authors for book", "error", err)
				return nil, err
		}
		book.Authors = authors

		// Fetch formats for the book
		formats, err := b.GetFormats(book.ID)
		if err != nil {
				b.Logger.Error("Error fetching formats for book", "error", err)
				return nil, err
		}
		book.Formats = formats

		// Fetch genres for the book
		genres, err := b.GetGenres(book.ID)
		if err != nil {
				b.Logger.Error("Error fetching genres", "error", err)
				return nil, err
		}
		book.Genres = genres

		// Fetch tags for the book
		tags, err := b.GetTagsForBook(book.ID)
		if err != nil {
				b.Logger.Error("Error fetching tags for book", "error", err)
				return nil, err
		}
		book.Tags = tags

		// Mark book as in Library
		book.IsInLibrary = true;

		books = append(books, book)
	}

	if err = rows.Err(); err != nil {
		b.Logger.Error("Error with rows", "error", err)
		return nil, err
	}

	return books, nil
}

func (b *BookModel) GetTagsForBook(bookID int) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// SQL query to fetch tags JSON array
	query := `
	SELECT tags
	FROM books
	WHERE id = $1`

	var tagsJSON []byte
	err := b.DB.QueryRowContext(ctx, query, bookID).Scan(&tagsJSON)
	if err != nil {
			b.Logger.Error("Error fetching tags for book", "error", err)
			return nil, err
	}

	// Unmarshal the JSON array of tags
	var tags []string
	if err := json.Unmarshal(tagsJSON, &tags); err != nil {
			b.Logger.Error("Error unmarshalling tags JSON", "error", err)
			return nil, err
	}

	return tags, nil
}

func (b *BookModel) GetAllBooksByAuthors(userID int) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// SQL query to get all books and their authors for a user's books
	query := `
	SELECT b.id, b.title, b.subtitle, b.description, b.language, b.page_count, b.publish_date, b.image_links, b.notes, b.created_at, b.last_updated, b.isbn_10, b.isbn_13, a.name
	FROM books b
	INNER JOIN book_authors ba ON b.id = ba.book_id
	INNER JOIN authors a ON ba.author_id = a.id
	INNER JOIN user_books ub ON b.id = ub.book_id
	WHERE ub.user_id = $1
	`

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
		); err != nil {
			b.Logger.Error("Error scanning book by author", "error", err)
			return nil, err
		}

		// Unmarshal image links
		if err := json.Unmarshal(imageLinksJSON, &book.ImageLinks); err != nil {
			b.Logger.Error("Error unmarshalling image links JSON", "error", err)
			return nil, err
		}

		// Fetch authors for the book
		bookArrAuthors, err := b.GetAuthorsForBook(book.ID)
		if err != nil {
				b.Logger.Error("Error fetching authors for book", "error", err)
				return nil, err
		}
		book.Authors = bookArrAuthors

		// Fetch formats for the book
		formats, err := b.GetFormats(book.ID)
		if err != nil {
				b.Logger.Error("Error fetching formats for book", "error", err)
				return nil, err
		}
		book.Formats = formats

		// Fetch genres for the book
		genres, err := b.GetGenres(book.ID)
		if err != nil {
			b.Logger.Error("Error fetching genres", "error", err)
			return nil, err
		}
		book.Genres = genres

		// Fetch tags for the book
		tags, err := b.GetTagsForBook(book.ID)
		if err != nil {
				b.Logger.Error("Error fetching tags for book", "error", err)
				return nil, err
		}
		book.Tags = tags

		// Mark book as in Library
		book.IsInLibrary = true;

		// Add author to the list if not already present
		if _, found := booksByAuthor[authorName]; !found {
			authors = append(authors, authorName)
		}

		// Add book to the author's list
		booksByAuthor[authorName] = append(booksByAuthor[authorName], book)
	}

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

func (b *BookModel) UpdateAuthors(bookID int, authors []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

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

	isbnSet := collections.NewSet()

	for rows.Next() {
		var bookTitle string
		if err := rows.Scan(&bookTitle); err != nil {
			b.Logger.Error("Error scanning book title", "error", err)
			return nil, err
		}
		isbnSet.Add(bookTitle)
	}

	if err = rows.Err(); err != nil {
		b.Logger.Error("Error with rows", "error", err)
		return nil, err
	}

	return isbnSet, nil
}

// (Return a HashMap)
func (b *BookModel) GetAllBooksPublishDate(userID int) (map[string]string, error) {
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

	bookMap := make(map[string]string)

	for rows.Next() {
		var title string
		var publishDate time.Time
		if err := rows.Scan(&title, &publishDate); err != nil {
			b.Logger.Error("Error scanning book title and publish date", "error", err)
			return nil, err
		}

		// Format publish date to "YYYY-MM-DD"
		formattedDate := publishDate.Format("2006-01-02")

		bookMap[title] = formattedDate
	}

	if err = rows.Err(); err != nil {
		b.Logger.Error("Error with rows", "error", err)
		return nil, err
	}

	return bookMap, nil
}


// Formats
func (b *BookModel) AddFormat(bookID, formatID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	statement := `INSERT INTO book_formats (book_id, format_id) VALUES ($1, $2)`
	_, err := b.DB.ExecContext(ctx, statement, bookID, formatID)
	if err != nil {
		b.Logger.Error("Book Model - Error adding format", "error", err)
		return err
	}

	return nil
}

func (b *BookModel) addOrGetFormatID(format string) (int, error) {
	var formatID int
	err := b.DB.QueryRow("SELECT id FROM formats WHERE format_type = $1", format).Scan(&formatID)
	if err != nil {
		if err == sql.ErrNoRows {
			err = b.DB.QueryRow("INSERT INTO formats (format_type) VALUES ($1) RETURNING id", format).Scan(&formatID)
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

	// b.Logger.Info("GetAllBooksByFormat, about to run query ","userID", userID)

	query := `
	SELECT b.id, b.title, b.subtitle, b.description, b.language, b.page_count, b.publish_date, b.image_links, b.notes, b.created_at, b.last_updated, b.isbn_10, b.isbn_13, f.format_type
	FROM books b
	INNER JOIN book_formats bf ON b.id = bf.book_id
	INNER JOIN formats f ON bf.format_id = f.id
	INNER JOIN user_books ub ON b.id = ub.book_id
	WHERE ub.user_id = $1`

	rows, err := b.DB.QueryContext(ctx, query, userID)
	if err != nil {
		b.Logger.Error("Error retrieving books by format", "error", err)
		return nil, err
	}
	defer rows.Close()

	booksByFormat := map[string][]Book{
		"audioBooks":   {},
		"eBooks":       {},
		"physicalBooks": {},
	}

	// Track unique book ids for each format
	uniqueAudioBooks := make(map[int]bool)
	uniqueEBooks := make(map[int]bool)
	uniquePhysicalBooks := make(map[int]bool)

	for rows.Next() {
		var book Book
		var imageLinksJSON []byte
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
		); err != nil {
			b.Logger.Error("Error scanning book", "error", err)
			return nil, err
		}
		// b.Logger.Info("--------------")
		// b.Logger.Info("Retrieved format type from DB", "formatType", formatType)

		if err := json.Unmarshal(imageLinksJSON, &book.ImageLinks); err != nil {
			b.Logger.Error("Error unmarshalling image links JSON", "error", err)
			return nil, err
		}

		// Fetch genres for the book
		genres, err := b.GetGenres(book.ID)
		if err != nil {
			b.Logger.Error("Error fetching genres", "error", err)
			return nil, err
		}
		book.Genres = genres

		// Fetch authors for the book
		authors, err := b.GetAuthorsForBook(book.ID)
		if err != nil {
				b.Logger.Error("Error fetching authors for book", "error", err)
				return nil, err
		}
		book.Authors = authors

		// Fetch formats for the book
		formats, err := b.GetFormats(book.ID)
		if err != nil {
				b.Logger.Error("Error fetching formats for book", "error", err)
				return nil, err
		}
		book.Formats = formats

		// Fetch tags for the book
		tags, err := b.GetTagsForBook(book.ID)
		if err != nil {
				b.Logger.Error("Error fetching tags for book", "error", err)
				return nil, err
		}
		book.Tags = tags

		// Mark book as in Library
		book.IsInLibrary = true;

		// b.Logger.Info("Processing book in format", "formatType", formatType, "bookID", book.ID)

		// Add book to the appropriate format list
		switch formatType {
		case "audioBook":
			if !uniqueAudioBooks[book.ID] {
				booksByFormat["audioBooks"] = append(booksByFormat["audioBooks"], book)
				uniqueAudioBooks[book.ID] = true
			}

		case "eBook":
			if !uniqueEBooks[book.ID] {
				booksByFormat["eBooks"] = append(booksByFormat["eBooks"], book)
				uniqueEBooks[book.ID] = true
			}

		case "physical":
			if !uniquePhysicalBooks[book.ID] {
				booksByFormat["physicalBooks"] = append(booksByFormat["physicalBooks"], book)
				uniquePhysicalBooks[book.ID] = true
			}

		default:
			b.Logger.Warn("Unknown format type encountered", "formatType", formatType, "bookID", book.ID)
		}
	}

	// Debug - Remove after UAT before prod push
	if len(booksByFormat["audioBooks"]) == 0 && len(booksByFormat["eBooks"]) == 0 && len(booksByFormat["physicalBooks"]) == 0 {
		b.Logger.Warn("No books found for user", "userID", userID)
	} else {
		// b.Logger.Info("Books found for user", "userID", userID, "audioBooks", len(booksByFormat["audioBooks"]), "eBooks", len(booksByFormat["eBooks"]), "physicalBooks", len(booksByFormat["physicalBooks"]))
	}

	if err := rows.Err(); err != nil {
		b.Logger.Error("Error with rows", "error", err)
		return nil, err
	}

	return booksByFormat, nil
}

func (b *BookModel) GetFormats(bookID int) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

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

	var formats []string
	for rows.Next() {
		var format string
		if err := rows.Scan(&format); err != nil {
			b.Logger.Error("Error scanning format", "error", err)
			return nil, err
		}
		formats = append(formats, format)
	}

	return formats, nil
}

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
func (b *BookModel) addOrGetGenreID(genreName string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

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

func (b *BookModel) AddGenre(bookID, genreID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	statement := `INSERT INTO book_genres (book_id, genre_id) VALUES ($1, $2)`
	_, err := b.DB.ExecContext(ctx, statement, bookID, genreID)
	if err != nil {
		b.Logger.Error("Error adding genre association", "error", err)
		return err
	}

	return nil
}

func (b *BookModel) GetGenres(bookID int) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
	SELECT g.name
	FROM genres g
	JOIN book_genres bg ON g.id = bg.genre_id
	WHERE bg.book_id = $1`
	rows, err := b.DB.QueryContext(ctx, query, bookID)
	if err != nil {
		b.Logger.Error("Error fetching genres", "error", err)
		return nil, err
	}
	defer rows.Close()

	var genres []string
	for rows.Next() {
		var genre string
		if err := rows.Scan(&genre); err != nil {
			b.Logger.Error("Error scanning genre", "error", err)
			return nil, err
		}
		genres = append(genres, genre)
	}

	return genres, nil
}

func (b *BookModel) GetAllBooksByGenres(userID int) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// SQL query to get all books and their genres for a user's books
	query := `
	SELECT b.id, b.title, b.subtitle, b.description, b.language, b.page_count, b.publish_date, b.image_links, b.notes, b.created_at, b.last_updated, b.isbn_10, b.isbn_13, g.name
	FROM books b
	INNER JOIN book_genres bg ON b.id = bg.book_id
	INNER JOIN genres g ON bg.genre_id = g.id
	INNER JOIN user_books ub ON b.id = ub.book_id
	WHERE ub.user_id = $1
	`

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
			var genreName string
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
					&genreName,
			); err != nil {
					b.Logger.Error("Error scanning book by genre", "error", err)
					return nil, err
			}

			// Unmarshal image links
			if err := json.Unmarshal(imageLinksJSON, &book.ImageLinks); err != nil {
					b.Logger.Error("Error unmarshalling image links JSON", "error", err)
					return nil, err
			}

			// Fetch authors for the book
			bookArrAuthors, err := b.GetAuthorsForBook(book.ID)
			if err != nil {
					b.Logger.Error("Error fetching authors for book", "error", err)
					return nil, err
			}
			book.Authors = bookArrAuthors

			// Add genre to the list if not already present
			if _, found := booksByGenre[genreName]; !found {
					genres = append(genres, genreName)
			}

			// Add book to the genre's list
			booksByGenre[genreName] = append(booksByGenre[genreName], book)
	}

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
					return getLastName(booksByGenre[genre][i].Authors[0]) < getLastName(booksByGenre[genre][j].Authors[0])
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
