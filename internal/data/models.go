package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"
)

const dbTimeout = time.Second * 3

func New(db *sql.DB, logger *slog.Logger) Models {
	return Models{
		User:      UserModel{DB: db, Logger: logger},
		Token:     TokenModel{DB: db, Logger: logger},
		Book:      BookModel{DB: db, Logger: logger},
		Category:  CategoryModel{DB: db, Logger: logger},
		Format:    FormatModel{DB: db, Logger: logger},
		Author:    AuthorModel{DB: db, Logger: logger},
		Genre:     GenreModel{DB: db, Logger: logger},
	}
}

type Models struct {
	User      UserModel
	Token     TokenModel
	Book      BookModel
  Category  CategoryModel
	Format    FormatModel
	Author    AuthorModel
	Genre     GenreModel
}

type UserModel struct {
	DB     *sql.DB
	Logger *slog.Logger
}

type TokenModel struct {
	DB     *sql.DB
	Logger *slog.Logger
}

type BookModel struct {
	DB     *sql.DB
	Logger *slog.Logger
}

type CategoryModel struct {
	DB     *sql.DB
	Logger *slog.Logger
}

type FormatModel struct {
	DB     *sql.DB
	Logger *slog.Logger
}

type AuthorModel struct {
	DB     *sql.DB
	Logger *slog.Logger
}

type GenreModel struct {
	DB     *sql.DB
	Logger *slog.Logger
}


type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"firstName,omitempty"`
	LastName  string    `json:"lastName,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Picture   string    `json:"picture,omitempty"`
}

type Token struct {
	ID             int       `json:"id"`
	UserID         int       `json:"user_id"`
	RefreshToken   string    `json:"refresh_token"`
	TokenExpiry    time.Time `json:"token_expiry"`
	PreviousToken  string    `json:"previous_token,omitempty"`
}

type Book struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Subtitle    string     `json:"subtitle"`
	Description string     `json:"description"`
	Language    string     `json:"language"`
	PageCount   int        `json:"pageCount"`
	PublishDate string     `json:"publishDate"`
	Authors     []string   `json:"authors"`
	ImageLinks  []string   `json:"imageLinks"`
	Genres      []string   `json:"genres"`
	Notes       string     `json:"notes"`
	Formats     []string   `json:"formats"`
	Tags        []string   `json:"tags"`
	CreatedAt   time.Time  `json:"created_at"`
	LastUpdated time.Time  `json:"lastUpdated"`
	ISBN10      string     `json:"isbn10"`
	ISBN13      string     `json:"isbn13"`
}

type Author struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Format struct {
	ID         int    `json:"id"`
	BookID     int    `json:"book_id"`
	FormatType string `json:"format_type"`
}

type Category struct {
	ID   int     `json:"id"`
	Name string  `json:"name"`
}

type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// User
func (u *UserModel) Insert(user User) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var newId int
	statement := `INSERT INTO users (email, first_name, last_name, created_at, updated_at, picture)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	err := u.DB.QueryRowContext(ctx, statement,
		user.Email,
		user.FirstName,
		user.LastName,
		time.Now(),
		time.Now(),
		user.Picture,
	).Scan(&newId)
	if err != nil {
		u.Logger.Error("User Model - Error inserting user", "error", err)
		return 0, err
	}

	return newId, nil
}

func (u *UserModel) GetByID(id int) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var user User
	statement := `SELECT id, email, first_name, last_name, created_at, updated_at, picture FROM users WHERE id = $1`
	row := u.DB.QueryRowContext(ctx, statement, id)
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Picture,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			u.Logger.Error("User Model - Error fetching user by ID", "error", err)
			return nil, err
		}
	}

	return &user, nil
}

func (u *UserModel) GetByEmail(email string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var user User
	statement := `SELECT id, email, first_name, last_name, created_at, updated_at, picture FROM users WHERE email = $1`
	row := u.DB.QueryRowContext(ctx, statement, email)
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Picture,
	)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, nil
		} else {
			u.Logger.Error("User Model - Error fetching user by email", "error", err)
			return nil, err
		}
	}
	return &user, nil
}

// Token
func (t *TokenModel) Insert(token Token) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	statement := `INSERT INTO tokens (user_id, refresh_token, token_expiry, previous_token)
		VALUES ($1, $2, $3, $4)`
	_, err := t.DB.ExecContext(ctx, statement,
		token.UserID,
		token.RefreshToken,
		token.TokenExpiry,
		token.PreviousToken,
	)
	if err != nil {
		t.Logger.Error("Token Model - Error inserting token", "error", err)
		return err
	}

	return nil
}

func (t *TokenModel) Rotate(userID int, newToken, oldToken string, expiry time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	statement := `UPDATE tokens SET refresh_token = $1, previous_token = $2, token_expiry = $3 WHERE user_id = $4 AND refresh_token = $5`
	_, err := t.DB.ExecContext(ctx, statement, newToken, oldToken, expiry, userID, oldToken)
	if err != nil {
		t.Logger.Error("Token Model - Error rotating token", "error", err)
		return err
	}
	return nil
}

func (t *TokenModel) DeleteByUserID(userID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	statement := `DELETE FROM tokens WHERE user_id = $1`
	_, err := t.DB.ExecContext(ctx, statement, userID)
	if err != nil {
		t.Logger.Error("Token Model - Error deleting token by user ID", "error", err)
		return err
	}

	return nil
}

// Book
func (b *BookModel) Insert(book Book) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var newId int

	// Marshal arrays to JSON strings
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

	statement := `INSERT INTO books (title, subtitle, description, language, page_count, publish_date, image_links, notes, tags, created_at, last_updated, isbn_10, isbn_13)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING id`

	err = b.DB.QueryRowContext(ctx, statement,
			book.Title,
			book.Subtitle,
			book.Description,
			book.Language,
			book.PageCount,
			book.PublishDate,
			imageLinksJSON,   // Insert JSON string
			book.Notes,
			tagsJSON,         // Insert JSON string
			time.Now(),
			time.Now(),
			book.ISBN10,
			book.ISBN13,
	).Scan(&newId)

	if err != nil {
			b.Logger.Error("Book Model - Error inserting book", "error", err)
			return 0, err
	}

	// Insert genres into the book_genres table
	for _, genre := range book.Genres {
		genreID, err := b.addOrGetGenreID(genre)
		if err != nil {
			b.Logger.Error("Error getting genre ID", "error", err)
			return 0, err
		}

		err = b.AddGenre(newId, genreID)
		if err != nil {
			b.Logger.Error("Error adding genre association", "error", err)
			return 0, err
		}
	}

	// Insert formats into the book_formats table
	for _, format := range book.Formats {
		formatID, err := b.addOrGetFormatID(format)
		if err != nil {
			b.Logger.Error("Error getting format ID", "error", err)
			return 0, err
		}

		err = b.AddFormat(newId, formatID)
		if err != nil {
			b.Logger.Error("Error adding format association", "error", err)
			return 0, err
		}
	}

	return newId, nil
}

func (b *BookModel) GetByID(id int) (*Book, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var book Book
	var imageLinksJSON, tagsJSON []byte
	statement := `SELECT id, title, subtitle, description, language, page_count, publish_date, image_links, notes, formats, tags, created_at, last_updated, isbn_10, isbn_13 FROM books WHERE id = $1`
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

	return &book, nil
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

		books = append(books, book)
	}

	if err = rows.Err(); err != nil {
		b.Logger.Error("Error with rows", "error", err)
		return nil, err
	}

	return books, nil
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

	statement := `UPDATE books SET title=$1, subtitle=$2, description=$3, language=$4, page_count=$5, publish_date=$6, image_links=$7, genres=$8, notes=$9, formats=$10, tags=$11, last_updated=$12, isbn_10=$13, isbn_13=$14 WHERE id=$15`
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
		books = append(books, book)
	}

	if err = rows.Err(); err != nil {
		b.Logger.Error("GetBooksByAuthor - Error with rows", "error", err)
		return nil, err
	}

	return books, nil
}

// Method to update authors for a book
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

func (b *BookModel) Delete(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	statement := `DELETE FROM books WHERE id = $1`
	_, err := b.DB.ExecContext(ctx, statement, id)
	if err != nil {
		b.Logger.Error("Book Model - Error deleting book", "error", err)
		return err
	}

	return nil
}

// Method to fetch all books grouped by format type for a specific user
func (b *BookModel) GetAllBooksByFormat(userID int) (map[string][]Book, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	b.Logger.Info("GetAllBooksByFormat, about to run query ","userID", userID)

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
		b.Logger.Info("--------------")
		b.Logger.Info("Retrieved format type from DB", "formatType", formatType)

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

		b.Logger.Info("Processing book in format", "formatType", formatType, "bookID", book.ID)

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

	if len(booksByFormat["audioBooks"]) == 0 && len(booksByFormat["eBooks"]) == 0 && len(booksByFormat["physicalBooks"]) == 0 {
		b.Logger.Warn("No books found for user", "userID", userID)
	} else {
		b.Logger.Info("Books found for user", "userID", userID, "audioBooks", len(booksByFormat["audioBooks"]), "eBooks", len(booksByFormat["eBooks"]), "physicalBooks", len(booksByFormat["physicalBooks"]))
	}

	if err := rows.Err(); err != nil {
		b.Logger.Error("Error with rows", "error", err)
		return nil, err
	}

	return booksByFormat, nil
}

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


// Format
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

func (f *FormatModel) Insert(bookID int, formatType string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var formatID int
	statement := `INSERT INTO formats (book_id, format_type) VALUES ($1, $2) RETURNING id`
	err := f.DB.QueryRowContext(ctx, statement, bookID, formatType).Scan(&formatID)
	if err != nil {
			f.Logger.Error("Format Model - Error inserting format", "error", err)
			return 0, err
	}

	return formatID, nil
}

func (f *FormatModel) GetByBookID(bookID int) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `SELECT format_type FROM formats WHERE book_id = $1`
	rows, err := f.DB.QueryContext(ctx, query, bookID)
	if err != nil {
		f.Logger.Error("Format Model - Error fetching formats by book ID", "error", err)
		return nil, err
	}
	defer rows.Close()

	var formats []string
	for rows.Next() {
		var format string
		if err := rows.Scan(&format); err != nil {
			f.Logger.Error("Error scanning format", "error", err)
			return nil, err
		}
		formats = append(formats, format)
	}

	return formats, nil
}

func (f *FormatModel) DeleteByBookID(bookID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	statement := `DELETE FROM formats WHERE book_id = $1`
	_, err := f.DB.ExecContext(ctx, statement, bookID)
	if err != nil {
		f.Logger.Error("Format Model - Error deleting formats by book ID", "error", err)
		return err
	}

	return nil
}


// Category
func (c *CategoryModel) Insert(category Category) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var newId int
	statement := `INSERT INTO categories (name) VALUES ($1) RETURNING id`
	err := c.DB.QueryRowContext(ctx, statement, category.Name).Scan(&newId)
	if err != nil {
		c.Logger.Error("Category Model - Error inserting category", "error", err)
		return 0, err
	}

	return newId, nil
}

func (c *CategoryModel) GetByID(id int) (*Category, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var category Category
	statement := `SELECT id, name FROM categories WHERE id = $1`
	row := c.DB.QueryRowContext(ctx, statement, id)
	err := row.Scan(
		&category.ID,
		&category.Name,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			c.Logger.Error("Category Model - Error fetching category by ID", "error", err)
			return nil ,err
		}
	}

	return &category, nil
}

func (c *CategoryModel) GetByName(name string) (*Category, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var category Category
	statement := `SELECT id, name FROM categories WHERE name = $1`
	row := c.DB.QueryRowContext(ctx, statement, name)
	err := row.Scan(
		&category.ID,
		&category.Name,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			c.Logger.Error("Category Model - Error fetching category by name", "error", err)
			return nil, err
		}
	}

	return &category, nil
}


// Author
func (a *AuthorModel) Insert(name string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var newID int
	statement := `INSERT INTO authors (name) VALUES ($1) RETURNING id`
	err := a.DB.QueryRowContext(ctx, statement, name).Scan(&newID)
	if err != nil {
		a.Logger.Error("Author Model - Error inserting author", "error", err)
		return 0, err
	}

	return newID, nil
}

func (a *AuthorModel) GetByID(id int) (*Author, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var author Author
	statement := `SELECT id, name FROM authors WHERE id = $1`
	row := a.DB.QueryRowContext(ctx, statement, id)
	err := row.Scan(&author.ID, &author.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			a.Logger.Error("Author Model - GetByID, no rows", "error", err)
			return nil, nil
		}
		a.Logger.Error("Author Model - Error fetching author by ID", "error", err)
		return nil, err
	}

	return &author, nil
}

// Genre
// Insert a new genre
func (g *GenreModel) Insert(name string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var newID int
	statement := `INSERT INTO genres (name) VALUES ($1) RETURNING id`
	err := g.DB.QueryRowContext(ctx, statement, name).Scan(&newID)
	if err != nil {
		g.Logger.Error("Genre Model - Error inserting genre", "error", err)
		return 0, err
	}

	return newID, nil
}

// Get a genre by ID
func (g *GenreModel) GetByID(id int) (*Genre, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var genre Genre
	statement := `SELECT id, name FROM genres WHERE id = $1`
	row := g.DB.QueryRowContext(ctx, statement, id)
	err := row.Scan(&genre.ID, &genre.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		g.Logger.Error("Genre Model - Error fetching genre by ID", "error", err)
		return nil, err
	}

	return &genre, nil
}

// Get genre for specific book
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

