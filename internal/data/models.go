package data

import (
	"context"
	"database/sql"
	"log/slog"
	"time"
)

const dbTimeout = time.Second * 3

func New(db *sql.DB, logger *slog.Logger) Models {
	return Models{
		User: UserModel{DB: db, Logger: logger},
		Token: TokenModel{DB: db, Logger: logger},
	}
}

type Models struct {
	User  UserModel
	Token TokenModel
}

type UserModel struct {
	DB     *sql.DB
	Logger *slog.Logger
}

type TokenModel struct {
	DB     *sql.DB
	Logger *slog.Logger
}

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name,omitempty"`
	LastName  string    `json:"last_name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Picture   string    `json:"picture,omitempty"`
}

type Token struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	RefreshToken string    `json:"refresh_token"`
	TokenExpiry  time.Time `json:"token_expiry"`
}

// User
func (u *UserModel) Insert(user User) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
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
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
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


// Token
func (t *TokenModel) Insert(token Token) error {
	// create context with a timeout to ensure db transaction doesn't go on forever
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)

	// ensure context is cancelled when fn runs
	defer cancel()

	// define SQL statement to do the things
	statement := `INSERT INTO tokens (user_id, refresh_token, token_expiry)
			VALUES ($1, $2, $3)`
	_, err := t.DB.ExecContext(ctx, statement,
		token.UserID,
		token.RefreshToken,
		token.TokenExpiry,
	)
	// if there was an error, log it and return
	if err != nil {
		t.Logger.Error("Token Model - Error inserting token", "error", err)
		return err
	}

	// rt nil if no error
	return nil
}