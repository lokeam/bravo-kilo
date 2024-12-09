package models

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
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

	// Token Family Pattern
	FamilyID      string    `json:"family_id"`
	IsRevoked     bool      `json:"is_revoked"`
}

// Define the interface
type TokenModel interface {
	Insert(token Token) error
	Delete(userID int) error
	GetRefreshTokenByUserID(userID int) (string, error)
	Rotate(ctx context.Context, userID int, newToken, oldToken string, expiry time.Time) error
	DeleteByUserID(userID int) error
	DeleteExpiredTokens(ctx context.Context) error
	UpdateLastUsed(tokenID int) error
	IncrementRefreshCount(tokenID int) error


	// Monitoring
	GetLatestActiveToken(ctx context.Context, userID int) (*Token, error)
  RecordRefreshAttempt(tokenID int, success bool, error string) error

	// Token Family Pattern
	ValidateRefreshToken(refreshToken string) (*Token, error)
	RevokeFamilyByID(familyID string) error
	IsFamilyRevoked(ctx context.Context, familyID string) (bool, error)
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

	// Generate new family ID if one is not provided
	if token.FamilyID == "" {
		token.FamilyID = uuid.New().String()
	}

	statement := `
		INSERT INTO tokens (
				user_id, refresh_token, token_expiry,
				previous_token, created_at, last_used_at,
				family_id, is_revoked
		) VALUES ($1, $2, $3, $4, NOW(), NOW(), $5, false)
	`

	t.logger.Debug("Inserting new token",
		"userID", token.UserID,
		"expiry", token.TokenExpiry,
		"familyID", token.FamilyID,
		"component", "token_model",
	)

	_, err := t.db.ExecContext(ctx, statement,
		token.UserID,
		token.RefreshToken,
		token.TokenExpiry,
		token.PreviousToken,
		token.FamilyID,
	)
	if err != nil {
		t.logger.Error("Failed to insert token",
			"error", err,
			"userID", token.UserID,
			"familyID", token.FamilyID,
			"component", "token_model",
    )
		return err
	}



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

func (t *TokenModelImpl) Rotate(ctx context.Context,userID int, newToken, oldToken string, expiry time.Time) error {
	// Create child context with timeout
	timeoutCtx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	tx, err := t.db.BeginTx(timeoutCtx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	t.logger.Debug("Attempting token rotation",
		"userID", userID,
		"tokenExpiry", expiry,
		"component", "token_model",
  )

	// 1. Validate old token
	var currentToken Token
	err = tx.QueryRowContext(ctx, `
		SELECT id, family_id, is_revoked, token_expiry
		FROM tokens
		WHERE user_id = $1 AND refresh_token = $2
	`, userID, oldToken).Scan(
		&currentToken.ID,
		&currentToken.FamilyID,
		&currentToken.IsRevoked,
		&currentToken.TokenExpiry,
	)

	if err == sql.ErrNoRows {
		t.logger.Error("Token not found during rotation",
			"userID", userID,
			"component", "token_model",
		)
		return fmt.Errorf("token not found: %w", err)
	}
	if err != nil {
		return fmt.Errorf("query token: %w", err)
	}

	// 2. Check if token is revoked
	if currentToken.IsRevoked {
		t.logger.Warn("Attempt to rotate revoked token",
			"userID", userID,
			"familyID", currentToken.FamilyID,
			"component", "token_model",
		)

		// Revoke entire family and return error
		if err := t.RevokeFamilyByID(currentToken.FamilyID); err != nil {
			t.logger.Error("Failed to revoke token family",
				"error", err,
				"familyID", currentToken.FamilyID,
				"component", "token_model",
			)
		}
		return fmt.Errorf("token is revoked: %w", err)
	}

	// 3. Insert new token in same family
	_, err = tx.ExecContext(ctx, `
        INSERT INTO tokens (
            user_id, refresh_token, token_expiry,
            family_id, last_used_at, is_revoked,
            refresh_count
        ) VALUES ($1, $2, $3, $4, NOW(), false, 0)
	`, userID, newToken, expiry, currentToken.FamilyID)

	if err != nil {
		return fmt.Errorf("insert new token: %w", err)
	}

	// 4. Revoke old token
	_, err = tx.ExecContext(ctx, `
        UPDATE tokens
        SET is_revoked = true,
            last_used_at = NOW()
        WHERE id = $1
	`)

	if err != nil {
		return fmt.Errorf("revoke old token: %w", err)
	}

	// 5. Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	t.logger.Info("Token rotation completed",
		"userID", userID,
		"familyID", currentToken.FamilyID,
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

func (t *TokenModelImpl) GetLatestActiveToken(ctx context.Context,userID int) (*Token, error) {
	// Create child context with Timeout
	timeoutCtx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
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

	err := t.db.QueryRowContext(timeoutCtx, statement, userID).Scan(
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

// Looks up token record by refresh token value to check validity + family status
func (t *TokenModelImpl) ValidateRefreshToken(refreshToken string) (*Token, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	t.logger.Debug("Validating refresh token",
		"refreshToken", refreshToken,
		"component", "token_model",
	)

	var token Token
	statement := `
		SELECT id, user_id, refresh_token, token_expiry,
						family_id, is_revoked, created_at, last_used_at,
						refresh_count
		FROM tokens
		WHERE refresh_token = $1
	`

	err := t.db.QueryRowContext(ctx, statement,  refreshToken).Scan(
		&token.ID,
		&token.UserID,
		&token.RefreshToken,
		&token.TokenExpiry,
		&token.FamilyID,
		&token.IsRevoked,
		&token.CreatedAt,
		&token.LastUsedAt,
		&token.RefreshCount,
	)

	if err == sql.ErrNoRows {
		t.logger.Warn("Refresh token not found",
			"component", "token_model",
		)
		return nil, nil
	}
	if err != nil {
		t.logger.Error("Database error validating refresh token",
				"error", err,
				"component", "token_model",
		)
		return nil, fmt.Errorf("query token: %w", err)
	}

	// Explicit expiration check
	if time.Now().After(token.TokenExpiry) {
		t.logger.Warn("Token expired",
			"tokenID", token.ID,
			"expiry", token.TokenExpiry,
			"component", "token_model",
		)
		return nil, nil
	}

	t.logger.Debug("Refresh token validated",
		"tokenID", token.ID,
		"userID", token.UserID,
		"isRevoked", token.IsRevoked,
		"expiresIn", time.Until(token.TokenExpiry),
		"component", "token_model",
	)

	return &token, nil
}

// Marks all tokens in a family as revoked
func (t *TokenModelImpl) RevokeFamilyByID(familyID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	// Start transaction
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
			return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	t.logger.Debug("Revoking family by ID",
		"familyID", familyID,
		"component", "token_model",
	)

	// Check if family exists
	var familyExists bool
	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 from tokens WHERE family_id = $1
		)
	`, familyID).Scan(&familyExists)

	if err != nil {
		return fmt.Errorf("check family existence: %w", err)
	}
	if !familyExists {
		t.logger.Warn("Token family not found",
			"familyID", familyID,
			"component", "token_model",
		)
		return nil
	}

	// Revoke all tokens in family
	// Track when revocation occurred, only update non-revoked tokens
	statement := `
        UPDATE tokens
        SET is_revoked = true,
            last_used_at = NOW()
        WHERE family_id = $1
          AND is_revoked = false
        RETURNING id
	`

	rows, err := t.db.QueryContext(ctx, statement, familyID)
	if err != nil {
		t.logger.Error("Failed to revoke token family",
				"error", err,
				"familyID", familyID,
				"component", "token_model",
		)
		return fmt.Errorf("revoke family: %w", err)
	}
	defer rows.Close()

	// Count the number of affected tokens
	var revokedTokens []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
				return fmt.Errorf("scan token id: %w", err)
		}
		revokedTokens = append(revokedTokens, id)
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("rows error: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	t.logger.Info("Token family revoked",
		"familyID", familyID,
		"tokensRevoked", revokedTokens,
		"component", "token_model",
	)

	return nil
}

// Check if token family is revoked
func (t *TokenModelImpl) IsFamilyRevoked(ctx context.Context, familyID string) (bool, error) {
	var isRevoked bool
	statement := `
        SELECT EXISTS (
            SELECT 1
            FROM tokens
            WHERE family_id = $1
            AND is_revoked = true
        )
	`

	err := t.db.QueryRowContext(ctx, statement, familyID).Scan(&isRevoked)
	if err != nil {
			t.logger.Error("Failed to check family revocation status",
					"error", err,
					"familyID", familyID,
					"component", "token_model",
			)
			return false, fmt.Errorf("check family revocation: %w", err)
	}

	return isRevoked, nil
}
