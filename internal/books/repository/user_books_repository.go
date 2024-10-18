package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/lokeam/bravo-kilo/internal/dbconfig"
)

type UserBooksRepository interface {
	DeleteUserBooks(userID int) error
}

type UserBooksRepositoryImpl struct {
	DB     *sql.DB
	Logger *slog.Logger
}

func NewUserBooksRepository(db *sql.DB, logger *slog.Logger) (UserBooksRepository, error) {
	if db == nil || logger == nil {
		return nil, fmt.Errorf("db or logger cannot be nil")
	}

	return &UserBooksRepositoryImpl{
		DB:     db,
		Logger: logger,
	}, nil
}

func (u *UserBooksRepositoryImpl) DeleteUserBooks(userID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	deleteStatement := `DELETE FROM user_books WHERE user_id = $1`
	_, err := u.DB.ExecContext(ctx, deleteStatement, userID)
	if err != nil {
		u.Logger.Error("Error deleting user books", "error", err)
		return err
	}

	u.Logger.Info("User books deleted successfully", "userID", userID)
	return nil
}
