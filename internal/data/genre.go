package data

import (
	"context"
	"database/sql"
	"log/slog"
)

type GenreModel struct {
	DB     *sql.DB
	Logger *slog.Logger
}

type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

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
