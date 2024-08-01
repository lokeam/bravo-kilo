package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"time"
)

const dbTimeout = time.Second * 3

func New(db *sql.DB, logger *slog.Logger) Models {
	return Models{
		User:      UserModel{DB: db, Logger: logger},
		Token:     TokenModel{DB: db, Logger: logger},
		Book:      BookModel{DB: db, Logger: logger},
		Category:  CategoryModel{DB: db, Logger: logger},
	}
}

type Models struct {
	User      UserModel
	Token     TokenModel
	Book      BookModel
  Category  CategoryModel
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

type Category struct {
	ID   int     `json:"id"`
	Name string  `json:"name"`
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

func (b *BookModel) Insert(book Book) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var newId int

	// Marshal arrays to JSON strings
	authorsJSON, err := json.Marshal(book.Authors)
	if err != nil {
			b.Logger.Error("Error marshalling authors to JSON", "error", err)
			return 0, err
	}
	imageLinksJSON, err := json.Marshal(book.ImageLinks)
	if err != nil {
			b.Logger.Error("Error marshalling image links to JSON", "error", err)
			return 0, err
	}
	genresJSON, err := json.Marshal(book.Genres)
	if err != nil {
			b.Logger.Error("Error marshalling genres to JSON", "error", err)
			return 0, err
	}
	formatsJSON, err := json.Marshal(book.Formats)
	if err != nil {
			b.Logger.Error("Error marshalling formats to JSON", "error", err)
			return 0, err
	}
	tagsJSON, err := json.Marshal(book.Tags)
	if err != nil {
			b.Logger.Error("Error marshalling tags to JSON", "error", err)
			return 0, err
	}

	statement := `INSERT INTO books (title, subtitle, description, language, page_count, publish_date, authors, image_links, genres, notes, formats, tags, created_at, last_updated, isbn_10, isbn_13)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16) RETURNING id`

	err = b.DB.QueryRowContext(ctx, statement,
			book.Title,
			book.Subtitle,
			book.Description,
			book.Language,
			book.PageCount,
			book.PublishDate,
			authorsJSON,      // Insert JSON string
			imageLinksJSON,   // Insert JSON string
			genresJSON,       // Insert JSON string
			book.Notes,
			formatsJSON,      // Insert JSON string
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

	return newId, nil
}

func (b *BookModel) GetByID(id int) (*Book, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var book Book
	var authorsJSON, imageLinksJSON, genresJSON, formatsJSON, tagsJSON []byte
	statement := `SELECT id, title, subtitle, description, language, page_count, publish_date, authors, image_links, genres, notes, formats, tags, created_at, last_updated, isbn_10, isbn_13 FROM books WHERE id = $1`
	row := b.DB.QueryRowContext(ctx, statement, id)
	err := row.Scan(
			&book.ID,
			&book.Title,
			&book.Subtitle,
			&book.Description,
			&book.Language,
			&book.PageCount,
			&book.PublishDate,
			&authorsJSON,
			&imageLinksJSON,
			&genresJSON,
			&book.Notes,
			&formatsJSON,
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
	if err := json.Unmarshal(authorsJSON, &book.Authors); err != nil {
			b.Logger.Error("Error unmarshalling authors JSON", "error", err)
			return nil, err
	}
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

	return &book, nil
}

func (b *BookModel) GetAllBooksByUserID(userID int) ([]Book, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
	SELECT b.id, b.title, b.subtitle, b.description, b.language, b.page_count, b.publish_date, b.authors, b.genres, b.formats, b.tags, b.image_links, b.created_at, b.last_updated, b.isbn_10, b.isbn_13
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
		var authorsJSON, genresJSON, formatsJSON, tagsJSON, imageLinksJSON []byte
		if err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.Subtitle,
			&book.Description,
			&book.Language,
			&book.PageCount,
			&book.PublishDate,
			&authorsJSON,
			&genresJSON,
			&formatsJSON,
			&tagsJSON,
			&imageLinksJSON,
			&book.CreatedAt,
			&book.LastUpdated,
			&book.ISBN10,
			&book.ISBN13,
		); err != nil {
			b.Logger.Error("Error scanning book", "error", err)
			return nil, err
		}

		// Unmarshal JSONB fields, handle null values
		if len(authorsJSON) > 0 {
			if err := json.Unmarshal(authorsJSON, &book.Authors); err != nil {
				b.Logger.Error("Error unmarshalling authors JSON", "error", err, "data", string(authorsJSON))
				return nil, err
			}
		}
		if len(genresJSON) > 0 {
			if err := json.Unmarshal(genresJSON, &book.Genres); err != nil {
				b.Logger.Error("Error unmarshalling genres JSON", "error", err, "data", string(genresJSON))
				return nil, err
			}
		}
		if len(formatsJSON) > 0 {
			if err := json.Unmarshal(formatsJSON, &book.Formats); err != nil {
				b.Logger.Error("Error unmarshalling formats JSON", "error", err, "data", string(formatsJSON))
				return nil, err
			}
		} else {
			book.Formats = []string{}
		}
		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &book.Tags); err != nil {
				b.Logger.Error("Error unmarshalling tags JSON", "error", err, "data", string(tagsJSON))
				return nil, err
			}
		} else {
			book.Tags = []string{}
		}
		if len(imageLinksJSON) > 0 {
			if err := json.Unmarshal(imageLinksJSON, &book.ImageLinks); err != nil {
				b.Logger.Error("Error unmarshalling image links JSON", "error", err, "data", string(imageLinksJSON))
				return nil, err
			}
		}

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

	authorsJSON, err := json.Marshal(book.Authors)
	if err != nil {
			b.Logger.Error("Error marshalling authors to JSON", "error", err)
			return err
	}
	imageLinksJSON, err := json.Marshal(book.ImageLinks)
	if err != nil {
			b.Logger.Error("Error marshalling image links to JSON", "error", err)
			return err
	}
	genresJSON, err := json.Marshal(book.Genres)
	if err != nil {
			b.Logger.Error("Error marshalling genres to JSON", "error", err)
			return err
	}
	formatsJSON, err := json.Marshal(book.Formats)
	if err != nil {
			b.Logger.Error("Error marshalling formats to JSON", "error", err)
			return err
	}
	tagsJSON, err := json.Marshal(book.Tags)
	if err != nil {
			b.Logger.Error("Error marshalling tags to JSON", "error", err)
			return err
	}

	statement := `UPDATE books SET title=$1, subtitle=$2, description=$3, language=$4, page_count=$5, publish_date=$6, authors=$7, image_links=$8, genres=$9, notes=$10, formats=$11, tags=$12, last_updated=$13, isbn_10=$14, isbn_13=$15 WHERE id=$16`
	_, err = b.DB.ExecContext(ctx, statement,
			book.Title,
			book.Subtitle,
			book.Description,
			book.Language,
			book.PageCount,
			book.PublishDate,
			authorsJSON,
			imageLinksJSON,
			genresJSON,
			book.Notes,
			formatsJSON,
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

func (b *BookModel) CountBooksByFormat(userID int) (map[string]int, error) {
	counts := make(map[string]int)

	query := `
	SELECT format, COUNT(*)
	FROM (
		SELECT jsonb_array_elements_text(formats) AS format
		FROM books b
		JOIN user_books ub ON b.id = ub.book_id
		WHERE ub.user_id = $1
	) AS format_counts
	GROUP BY format`

	rows, err := b.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var format string
		var count int
		if err := rows.Scan(&format, &count); err != nil {
			return nil, err
		}
		counts[format] = count
	}

	return counts, nil
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
