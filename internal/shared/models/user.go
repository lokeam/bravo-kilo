package models

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/lokeam/bravo-kilo/internal/dbconfig"
)

type UserModel struct {
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

func (u *UserModel) Insert(user User) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
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
	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
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
	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
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
