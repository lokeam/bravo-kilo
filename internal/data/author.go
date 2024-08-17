package data

import (
	"context"
	"database/sql"
	"log/slog"
)

type AuthorModel struct {
	DB     *sql.DB
	Logger *slog.Logger
}

type Author struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (a *AuthorModel) Insert(name string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var authorID int
	statement := `INSERT INTO authors (name) VALUES ($1) ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name RETURNING id`
	err := a.DB.QueryRowContext(ctx, statement, name).Scan(&authorID)
	if err != nil {
		a.Logger.Error("Author Model - Error inserting author", "error", err)
		return 0, err
	}

	return authorID, nil
}

func (a *AuthorModel) AddOrGetAuthorID(name string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var authorID int
	query := `SELECT id FROM authors WHERE name = $1`
	err := a.DB.QueryRowContext(ctx, query, name).Scan(&authorID)
	if err != nil && err != sql.ErrNoRows {
		a.Logger.Error("Error checking author existence", "error", err)
		return 0, err
	}

	if authorID == 0 {
		// Author does not exist, insert new author
		query = `INSERT INTO authors (name) VALUES ($1) RETURNING id`
		err = a.DB.QueryRowContext(ctx, query, name).Scan(&authorID)
		if err != nil {
			a.Logger.Error("Error inserting new author", "error", err)
			return 0, err
		}
	}

	return authorID, nil
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
