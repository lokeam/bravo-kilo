package handlers

import (
	"bravo-kilo/internal/data"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/time/rate"
)

var jwtKey = []byte("extra-super-secret-256-bit-key")

type Claims struct {
	UserID int `json:"userId"`
	jwt.RegisteredClaims
}

// Handlers struct to hold the logger, models, and new components
type Handlers struct {
	logger        *slog.Logger
	models        data.Models
	exportLimiter *rate.Limiter
	validate      *validator.Validate
	sanitizer     *bluemonday.Policy
}

type jsonResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// NewHandlers creates a new Handlers instance
func NewHandlers(logger *slog.Logger, models data.Models) *Handlers {
	validate := validator.New()
	sanitizer := bluemonday.UGCPolicy()

	return &Handlers{
		logger:        logger,
		models:        models,
		exportLimiter: rate.NewLimiter(rate.Limit(1), 3),
		validate:      validate,
		sanitizer:     sanitizer,
	}
}