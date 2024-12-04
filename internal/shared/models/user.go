package models

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/lokeam/bravo-kilo/internal/dbconfig"
)

type UserRepository interface {
	Insert(user User) (int, error)
	GetByID(id int) (*User, error)
	GetUserBookIDs(userID int) ([]int, error)
	GetByEmail(email string) (*User, error)
	MarkForDeletion(ctx context.Context, tx *sql.Tx, userID int, deletionTime time.Time) error
	Delete(userID int) error
}


type UserRepositoryImpl struct {
	DB     *sql.DB
	Logger *slog.Logger
}

type User struct {
	ID                  int       `json:"id"`
	Email               string    `json:"email"`
	FirstName           string    `json:"firstName,omitempty"`
	LastName            string    `json:"lastName,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updatedAt"`
	Picture             string    `json:"picture,omitempty"`
	IsPendingDeletion   bool      `json:"isPendingDeletion"`
	DeletionRequestedAt time.Time `json:"deletionRequestedAt"`
}

func NewUserRepository(db *sql.DB, logger *slog.Logger) (UserRepository, error) {
	if db == nil || logger == nil {
		return nil, fmt.Errorf("database or logger is nil")
	}

	return &UserRepositoryImpl{
		DB:     db,
		Logger: logger,
	}, nil
}

func (u *UserRepositoryImpl) Insert(user User) (int, error) {
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

func (u *UserRepositoryImpl) GetByID(id int) (*User, error) {
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

func (u *UserRepositoryImpl) GetUserBookIDs(userID int) ([]int, error) {
	query := `SELECT book_id FROM user_books WHERE user_id = $1`
	rows, err := u.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookIDs []int
	for rows.Next() {
		var bookID int
		err := rows.Scan(&bookID)
		if err != nil {
			return nil, err
		}
		bookIDs = append(bookIDs, bookID)
	}
	return bookIDs, nil
}

func (u *UserRepositoryImpl) GetByEmail(email string) (*User, error) {
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

func (u *UserRepositoryImpl) MarkForDeletion(ctx context.Context, tx *sql.Tx, userID int, deletionTime time.Time) error {
	// Update the user record to mark for deletion
	statement := `UPDATE users
		SET is_pending_deletion = true,
			deletion_requested_at = $1,
			updated_at = $2
		WHERE id = $3`
	_, err := tx.ExecContext(ctx, statement, deletionTime, time.Now(), userID)
	if err != nil {
		u.Logger.Error("User Model - Error marking user for deletion", "error", err)
		return err
	}

	u.Logger.Info("User marked for deletion", "userID", userID)
	return nil
}

func (u *UserRepositoryImpl) Delete(userID int) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := u.DB.Exec(query, userID)
	if err != nil {
		u.Logger.Error("User Model - Error deleting user", "error", err)
		return err
	}

	return nil
}

