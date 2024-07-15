package data

import "time"

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