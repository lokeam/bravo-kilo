package models

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/lokeam/bravo-kilo/internal/dbconfig"
)

type Token struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	RefreshToken  string    `json:"refresh_token"`
	TokenExpiry   time.Time `json:"token_expiry"`
	PreviousToken string    `json:"previous_token,omitempty"`
}

// Define the interface
type TokenModel interface {
	Insert(token Token) error
	Delete(userID int) error
	GetRefreshTokenByUserID(userID int) (string, error)
	Rotate(userID int, newToken, oldToken string, expiry time.Time) error
	DeleteByUserID(userID int) error
}

// Implementation struct
type TokenModelImpl struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewTokenModel(db *sql.DB, logger *slog.Logger) TokenModel {
	return &TokenModelImpl{
			db:     db,
			logger: logger,
	}
}

func (t *TokenModelImpl) Insert(token Token) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	statement := `INSERT INTO tokens (user_id, refresh_token, token_expiry, previous_token)
		VALUES ($1, $2, $3, $4)`
	_, err := t.db.ExecContext(ctx, statement,
		token.UserID,
		token.RefreshToken,
		token.TokenExpiry,
		token.PreviousToken,
	)
	if err != nil {
		t.logger.Error("Token Model - Error inserting token", "error", err)
		return err
	}

	return nil
}

func (t *TokenModelImpl) GetRefreshTokenByUserID(userID int) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	var refreshToken string
	statement := `SELECT refresh_token FROM tokens WHERE user_id = $1 LIMIT 1`
	err := t.db.QueryRowContext(ctx, statement, userID).Scan(&refreshToken)
	if err != nil {
    if err == sql.ErrNoRows {
      t.logger.Error("Token Model - No refresh token found for user", "userID", userID)
      return "", nil
    }
    t.logger.Error("Token Model - Error fetching refresh token", "error", err)
    return "", err
  }

	return refreshToken, nil
}

func (t *TokenModelImpl) Rotate(userID int, newToken, oldToken string, expiry time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	statement := `UPDATE tokens SET refresh_token = $1, previous_token = $2, token_expiry = $3 WHERE user_id = $4 AND refresh_token = $5`
	_, err := t.db.ExecContext(ctx, statement, newToken, oldToken, expiry, userID, oldToken)
	if err != nil {
		t.logger.Error("Token Model - Error rotating token", "error", err)
		return err
	}
	return nil
}

func (t *TokenModelImpl) DeleteByUserID(userID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	statement := `DELETE FROM tokens WHERE user_id = $1`
	_, err := t.db.ExecContext(ctx, statement, userID)
	if err != nil {
		t.logger.Error("Token Model - Error deleting token by user ID", "error", err)
		return err
	}

	return nil
}

func (t *TokenModelImpl) Delete(userID int) error {
    return t.DeleteByUserID(userID)
}
