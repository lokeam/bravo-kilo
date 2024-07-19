package data

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/lib/pq"
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
	FirstName string    `json:"first_name,omitempty"`
	LastName  string    `json:"last_name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Picture   string    `json:"picture,omitempty"`
}

type Token struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	RefreshToken string    `json:"refresh_token"`
	TokenExpiry  time.Time `json:"token_expiry"`
}

type Book struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Subtitle    string     `json:"subtitle"`
	Description string     `json:"description"`
	Language    string     `json:"language"`
	PageCount   int        `json:"page_count"`
	PublishDate string     `json:"publish_date"`
	Authors     []string   `json:"authors"`
	ImageLinks  []string   `json:"image_links"`
	Categories  []string   `json:"categories"`
	CreatedAt   time.Time  `json:"created_at"`
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


// Token
func (t *TokenModel) Insert(token Token) error {
	// create context with a timeout to ensure db transaction doesn't go on forever
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)

	// ensure context is cancelled when fn runs
	defer cancel()

	// define SQL statement to do the things
	statement := `INSERT INTO tokens (user_id, refresh_token, token_expiry)
			VALUES ($1, $2, $3)`
	_, err := t.DB.ExecContext(ctx, statement,
		token.UserID,
		token.RefreshToken,
		token.TokenExpiry,
	)
	// if there was an error, log it and return
	if err != nil {
		t.Logger.Error("Token Model - Error inserting token", "error", err)
		return err
	}

	// rt nil if no error
	return nil
}

// Book
func (b *BookModel) Insert(book Book) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var newId int
	statement := `INSERT INTO books (title, subtitle, description, language, page_count, publish_date, authors, image_links, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`

	err := b.DB.QueryRowContext(ctx, statement,
		book.Title,
		book.Subtitle,
		book.Description,
		book.Language,
		book.PageCount,
		book.PublishDate,
		pq.Array(book.Authors),
		pq.Array(book.ImageLinks),
		time.Now(),
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
	statement := `SELECT id, title, subtitle, description, language, page_count, publish_date, authors, image_links, created_at FROM books WHERE id = $1`
	row := b.DB.QueryRowContext(ctx, statement, id)
	err := row.Scan(
		&book.ID,
		&book.Title,
		&book.Subtitle,
		&book.Description,
		&book.Language,
		&book.PageCount,
		&book.PublishDate,
		pq.Array(&book.Authors),
		pq.Array(&book.ImageLinks),
		&book.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			b.Logger.Error("Book Model - Error fetching book by ID", "error", err)
			return nil, err
		}
	}

	return &book, nil
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
