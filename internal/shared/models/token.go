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
	CreatedAt     time.Time `json:"created_at"`
	LastUsedAt    time.Time `json:"last_used_at"`
	RefreshCount  int       `json:"refresh_count"`
}

// Define the interface
type TokenModel interface {
	Insert(token Token) error
	Delete(userID int) error
	GetRefreshTokenByUserID(userID int) (string, error)
	Rotate(userID int, newToken, oldToken string, expiry time.Time) error
	DeleteByUserID(userID int) error
	DeleteExpiredTokens(ctx context.Context) error
	UpdateLastUsed(tokenID int) error
	IncrementRefreshCount(tokenID int) error

	// Monitoring
	GetLatestActiveToken(userID int) (*Token, error)
  RecordRefreshAttempt(tokenID int, success bool, error string) error
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

	// Set expiry to a future time (e.g., 7 days from now)
	token.TokenExpiry = time.Now().Add(7 * 24 * time.Hour)

	statement := `
		INSERT INTO tokens (
				user_id, refresh_token, token_expiry,
				previous_token, created_at, last_used_at
		) VALUES ($1, $2, $3, $4, NOW(), NOW())
	`

	t.logger.Debug("Inserting new token",
		"userID", token.UserID,
		"expiry", token.TokenExpiry,
		"component", "token_model",
	)

	_, err := t.db.ExecContext(ctx, statement,
		token.UserID,
		token.RefreshToken,
		token.TokenExpiry,
		token.PreviousToken,
	)
	if err != nil {
		t.logger.Error("Failed to insert token",
			"error", err,
			"userID", token.UserID,
			"component", "token_model",
    )
		return err
	}

	t.logger.Debug("Setting token expiry",
		"userID", token.UserID,
		"expiry", token.TokenExpiry,
		"component", "token_model",
	)

	return nil
}

func (t *TokenModelImpl) GetRefreshTokenByUserID(userID int) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	t.logger.Debug("Starting refresh token retrieval",
			"userID", userID,
			"component", "token_model",
	)

	var refreshToken string
	statement := `
			SELECT refresh_token
			FROM tokens
			WHERE user_id = $1
				AND token_expiry > NOW()
			ORDER BY created_at DESC
			LIMIT 1`

	t.logger.Debug("Executing token retrieval query",
			"userID", userID,
			"query", statement,
			"component", "token_model",
	)

	err := t.db.QueryRowContext(ctx, statement, userID).Scan(&refreshToken)
	if err != nil {
			if err == sql.ErrNoRows {
					t.logger.Error("No refresh token found for user",
							"userID", userID,
							"component", "token_model",
							"error", "no_rows",
					)
					return "", nil
			}
			t.logger.Error("Failed to fetch refresh token",
					"userID", userID,
					"component", "token_model",
					"error", err,
			)
			return "", err
	}

	t.logger.Debug("Successfully retrieved refresh token",
			"userID", userID,
			"component", "token_model",
			"tokenLength", len(refreshToken),
			"tokenPrefix", refreshToken[:10]+"...",
	)

	return refreshToken, nil
}

func (t *TokenModelImpl) Rotate(userID int, newToken, oldToken string, expiry time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	t.logger.Debug("Attempting token rotation",
		"userID", userID,
		"tokenExpiry", expiry,
		"component", "token_model",
  )

	statement := `UPDATE tokens
	SET refresh_token = $1,
			previous_token = $2,
			token_expiry = $3
	WHERE user_id = $4 AND refresh_token = $5`

	result, err := t.db.ExecContext(ctx, statement, newToken, oldToken, expiry, userID, oldToken)
	if err != nil {
		t.logger.Error("Failed to rotate token",
				"error", err,
				"userID", userID,
				"component", "token_model",
		)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		t.logger.Error("Error checking rows affected during rotation",
				"error", err,
				"userID", userID,
				"component", "token_model",
		)
	} else if rowsAffected == 0 {
		t.logger.Warn("No tokens were rotated",
				"userID", userID,
				"component", "token_model",
		)
	}

	t.logger.Info("Token rotation completed",
		"userID", userID,
		"rowsAffected", rowsAffected,
		"newExpiry", expiry,
		"component", "token_model",
	)
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

func (t *TokenModelImpl) DeleteExpiredTokens(ctx context.Context) error {
	t.logger.Debug("Starting expired tokens cleanup",
			"component", "token_model",
	)

	statement := `
			WITH deleted AS (
					DELETE FROM tokens
					WHERE token_expiry < NOW()
					RETURNING id, user_id, token_expiry
			)
			SELECT COUNT(*),
						 array_agg(id),
						 array_agg(user_id),
						 MIN(token_expiry),
						 MAX(token_expiry)
			FROM deleted`

	var (
			count int
			tokenIDs []int
			userIDs []int
			oldestExpiry, newestExpiry time.Time
	)

	err := t.db.QueryRowContext(ctx, statement).Scan(
			&count,
			&tokenIDs,
			&userIDs,
			&oldestExpiry,
			&newestExpiry,
	)
	if err != nil && err != sql.ErrNoRows {
			t.logger.Error("Failed to delete expired tokens",
					"error", err,
					"component", "token_model",
			)
			return err
	}

	if count > 0 {
			t.logger.Info("Expired tokens deleted",
					"count", count,
					"tokenIDs", tokenIDs,
					"userIDs", userIDs,
					"oldestExpiry", oldestExpiry,
					"newestExpiry", newestExpiry,
					"component", "token_model",
			)
	} else {
			t.logger.Debug("No expired tokens found",
					"component", "token_model",
			)
	}

	return nil
}

func (t *TokenModelImpl) UpdateLastUsed(tokenID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	t.logger.Debug("Updating token last used timestamp",
			"tokenID", tokenID,
			"component", "token_model",
	)

	statement := `
			UPDATE tokens
			SET last_used_at = NOW()
			WHERE id = $1`

	result, err := t.db.ExecContext(ctx, statement, tokenID)
	if err != nil {
			t.logger.Error("Failed to update last used timestamp",
					"error", err,
					"tokenID", tokenID,
					"component", "token_model",
			)
			return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
			t.logger.Error("Error checking rows affected for last used update",
					"error", err,
					"tokenID", tokenID,
					"component", "token_model",
			)
	} else if rowsAffected == 0 {
			t.logger.Warn("No token found to update last used timestamp",
					"tokenID", tokenID,
					"component", "token_model",
			)
	}

	t.logger.Info("Token last used timestamp updated",
			"tokenID", tokenID,
			"component", "token_model",
	)
	return nil
}

func (t *TokenModelImpl) GetLatestActiveToken(userID int) (*Token, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	var token Token
	statement := `
			SELECT id, user_id, refresh_token, token_expiry, previous_token, created_at, last_used_at
			FROM tokens
			WHERE user_id = $1
				AND token_expiry > NOW()
			ORDER BY token_expiry DESC
			LIMIT 1
	`

	err := t.db.QueryRowContext(ctx, statement, userID).Scan(
			&token.ID,
			&token.UserID,
			&token.RefreshToken,
			&token.TokenExpiry,
			&token.PreviousToken,
			&token.CreatedAt,
			&token.LastUsedAt,
	)

	if err == sql.ErrNoRows {
			return nil, nil
	}
	if err != nil {
			t.logger.Error("Token Model - Error fetching latest active token", "error", err)
			return nil, err
	}

	return &token, nil
}

func (t *TokenModelImpl) RecordRefreshAttempt(tokenID int, success bool, errorMsg string) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	t.logger.Debug("Recording refresh attempt",
			"tokenID", tokenID,
			"success", success,
			"hasError", errorMsg != "",
			"component", "token_model",
	)

	statement := `
			INSERT INTO token_events (
					token_id,
					event_type,
					success,
					error_message,
					created_at
			) VALUES ($1, 'refresh_attempt', $2, $3, NOW())`

	_, err := t.db.ExecContext(ctx, statement,
			tokenID,
			success,
			errorMsg,
	)
	if err != nil {
			t.logger.Error("Failed to record refresh attempt",
					"error", err,
					"tokenID", tokenID,
					"success", success,
					"component", "token_model",
			)
			return err
	}

	t.logger.Info("Refresh attempt recorded",
			"tokenID", tokenID,
			"success", success,
			"component", "token_model",
	)
	return nil
}

func (t *TokenModelImpl) IncrementRefreshCount(tokenID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	t.logger.Debug("Incrementing refresh count",
			"tokenID", tokenID,
			"component", "token_model",
	)

	statement := `
			UPDATE tokens
			SET refresh_count = refresh_count + 1
			WHERE id = $1
			RETURNING refresh_count`

	var newCount int
	err := t.db.QueryRowContext(ctx, statement, tokenID).Scan(&newCount)
	if err != nil {
			t.logger.Error("Failed to increment refresh count",
					"error", err,
					"tokenID", tokenID,
					"component", "token_model",
			)
			return err
	}

	t.logger.Info("Refresh count incremented",
			"tokenID", tokenID,
			"newCount", newCount,
			"component", "token_model",
	)
	return nil
}