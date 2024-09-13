package models

import (
	"context"
	"database/sql"
	"log/slog"
	"time"
)

type TokenModel struct {
	DB     *sql.DB
	Logger *slog.Logger
}

type Token struct {
	ID             int       `json:"id"`
	UserID         int       `json:"user_id"`
	RefreshToken   string    `json:"refresh_token"`
	TokenExpiry    time.Time `json:"token_expiry"`
	PreviousToken  string    `json:"previous_token,omitempty"`
}

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

func (t *TokenModel) GetRefreshTokenByUserID(userID int) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var refreshToken string
	statement := `SELECT refresh_token FROM tokens WHERE user_id = $1 LIMIT 1`
	err := t.DB.QueryRowContext(ctx, statement, userID).Scan(&refreshToken)
	if err != nil {
    if err == sql.ErrNoRows {
      t.Logger.Error("Token Model - No refresh token found for user", "userID", userID)
      return "", nil
    }
    t.Logger.Error("Token Model - Error fetching refresh token", "error", err)
    return "", err
  }

	return refreshToken, nil
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
