package home

import (
	"errors"
	"log/slog"
	"net/http"
)

type ErrorResponse struct {
	Error     string `json:"error"`
	RequestID string `json:"requestId"`
}

type HomeHandler struct {
	logger *slog.Logger
}

var (
	ErrValidation     = errors.New("validation error")
	ErrAuthentication = errors.New("authentication error")
)

func NewHomeHandler(logger *slog.Logger) *HomeHandler {
	if logger == nil {
		panic("logger is required")
	}

	return &HomeHandler{}
}

func (h *HomeHandler) HandleGetHomePageData(w http.ResponseWriter, r *http.Request) {
	// Setup

	// Auth

	// Validate / parse params

	// Get home page data

	// Send response back to frontend
}
