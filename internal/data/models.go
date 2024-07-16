package data

import (
	"database/sql"
	"log/slog"
	"time"
)

const dbTimeout = time.Second * 3
var db *sql.DB

type Models struct {
	User UserModel
	Token TokenModel
}

func New(db *sql.DB, logger *slog.Logger) Models {
	return Models{
		User: UserModel{DB: db, Logger: logger},
		Token: TokenModel{DB: db, Logger: logger},
	}
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
	ID         int       `json:"id"`
	Email      string    `json:"email"`
	FirstName  string    `json:"first_name,omitempty"`
	LastName   string    `json:"last_name,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Token struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	RefreshToken  string    `json:"refresh_token"`
	TokenExpiry   time.Time `json:"token_expiry"`
}