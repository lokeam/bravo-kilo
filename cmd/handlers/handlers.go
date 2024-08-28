package handlers

import (
	"bravo-kilo/internal/data"
	"log/slog"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/time/rate"
)

var jwtKey = []byte("extra-super-secret-256-bit-key")

type Claims struct {
	UserID int `json:"userId"`
	jwt.RegisteredClaims
}

// Handlers struct to hold the logger
type Handlers struct {
	logger *slog.Logger
	models data.Models
	exportLimiter *rate.Limiter
}

type jsonResponse struct {
	Error    bool        `json:"error"`
	Message  string      `json:"message"`
	Data     interface{} `json:"data,omitempty"`
}


// NewHandlers creates a new Handlers instance
func NewHandlers(logger *slog.Logger, models data.Models) *Handlers {
	return &Handlers{
		logger: logger,
		models: models,
	}
}
