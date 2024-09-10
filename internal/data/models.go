package data

import (
	"context"
	"database/sql"
	"log/slog"
	"time"
)

const dbTimeout = time.Second * 3

type Models struct {
	User      UserModel
	Token     TokenModel
	Book      BookRepository
  Category  CategoryModel
}

type CategoryModel struct {
	DB     *sql.DB
	Logger *slog.Logger
}

type Category struct {
	ID   int     `json:"id"`
	Name string  `json:"name"`
}

func New(db *sql.DB, logger *slog.Logger) (Models, error) {
	models := Models{
		User:      UserModel{DB: db, Logger: logger},
		Token:     TokenModel{DB: db, Logger: logger},
		Book:      BookModel{DB: db, Logger: logger, Author: &AuthorModel{DB: db, Logger: logger}},
		Category:  CategoryModel{DB: db, Logger: logger},
	}

	// Init prepared statements for BookModel
	if err := models.Book.InitPreparedStatements(); err != nil {
		logger.Error("Error initializing prepared statements for BookModel", "error", err)
		return Models{}, err
	}

	// Init additional models for prepared statements below:

	return models, nil
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
			return nil, err
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
