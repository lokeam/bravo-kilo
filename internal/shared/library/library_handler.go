package library

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/lokeam/bravo-kilo/config"
	"github.com/lokeam/bravo-kilo/internal/shared/core"
	"github.com/lokeam/bravo-kilo/internal/shared/jwt"
	"github.com/lokeam/bravo-kilo/internal/shared/operations"
	"github.com/lokeam/bravo-kilo/internal/shared/services"
)

type ErrorResponse struct {
	Error     string `json:"error"`
	RequestID string `json:"requestId"`
}

type LibraryHandler struct {
	authExecutor              *operations.OperationExecutor[int]
	libraryService            *LibraryService
	validationService         *services.ValidationService
	logger                    *slog.Logger
}

var (
	ErrValidation     = errors.New("validation error")
	ErrAuthentication = errors.New("authentication error")
)

func NewLibraryHandler(
	libraryService *LibraryService,
	validationService *services.ValidationService,
	logger *slog.Logger,
) *LibraryHandler {
	if libraryService == nil {
		panic("library service is required")
	}
	if validationService == nil {
		panic("validation service is required")
	}
	if logger == nil {
		panic("logger is required")
	}
	return &LibraryHandler{
		authExecutor:        operations.NewOperationExecutor[int]("auth_request", 5*time.Second, logger),
		libraryService:      libraryService,
		validationService:   validationService,
		logger:              logger,
	}
}

// RESPONSIBILITIES:

// 1. HTTP Request Setup
func (h *LibraryHandler) HandleGetLibraryPageData(w http.ResponseWriter, r *http.Request) {
    /*
    Responsibilities:
    - Extract request ID (from header or generate new)
    - Basic request logging
    - Call auth middleware for userID
    - Parse query params into LibraryQueryParams
    - Call service layer
    - Handle response writing
    */

	// Setup
	ctx := r.Context()
	requestID, ok := ctx.Value(core.RequestIDKey).(string)
	if !ok {
		h.logger.Error("request ID not found in context")
		h.respondWithError(w, "", fmt.Errorf("internal server error"))
		return
	}

	// Auth
	userID, err := h.authenticateRequest(ctx, r)
	if err != nil {
		h.respondWithError(w, requestID, err)
		return
	}

	// Validate and parse params
	params, err := h.validationService.ValidateLibraryRequest(ctx, r.URL.Query())
	if err != nil {
		h.respondWithError(w, requestID, err)
		return
	}


	// Get library data
	libraryData, err := h.libraryService.GetLibraryData(ctx, userID, params)
	if err != nil {
		h.respondWithError(w, requestID, err)
		return
	}

	// Send response back to frontend
	if err := h.respondWithJSON(w, http.StatusOK, libraryData); err != nil {
		h.logger.Error("failed to send response",
			"error", err,
		)
	}

	return
}

// Helpers
func (h *LibraryHandler) authenticateRequest(ctx context.Context, r *http.Request) (int, error) {
	return h.authExecutor.Execute(ctx, func(ctx context.Context) (int, error) {
		userID, err := jwt.ExtractUserIDFromJWT(r, config.AppConfig.JWTPublicKey)
		if err != nil {
				return 0, ErrAuthentication
		}
		return userID, nil
	})
}

func (h *LibraryHandler) respondWithJSON(w http.ResponseWriter, status int, data any) error {
	/*
	Responsibilities:
	- Set content type to JSON
	- Write status code
	- Write JSON response
	*/

		h.logger.Debug("preparing response",
			"status", status,
			"contentType", "application/json",
		)
    // Start timing the response
    start := time.Now()

    // Log header state before write attempt
    if rw, ok := w.(interface{ Written() bool }); ok {
			h.logger.Debug("checking response writer state",
					"headersWritten", rw.Written(),
					"status", status,
			)
		}

    // Only set Content-Type if not already set
    if w.Header().Get("Content-Type") == "" {
        h.logger.Debug("setting content-type header",
            "contentType", "application/json",
        )
        w.Header().Set("Content-Type", "application/json")
    }

    // Check if we can write headers
    if rw, ok := w.(interface{ Written() bool }); !ok || !rw.Written() {
        h.logger.Debug("writing headers",
            "contentType", "application/json",
            "status", status,
        )
        w.WriteHeader(status)
    }

    // Encode data to JSON and write to response
    if err := json.NewEncoder(w).Encode(data); err != nil {
        h.logger.Error("failed to encode response",
            "error", err,
            "status", status,
            "duration", time.Since(start),
        )
        return fmt.Errorf("failed to encode response: %w", err)
    }

    // Log successful response
    h.logger.Debug("response sent successfully",
        "status", status,
        "duration", time.Since(start),
    )

    return nil
}

func (h *LibraryHandler) respondWithError(w http.ResponseWriter, requestID string, err error) error {
	// Start timing the response
	start := time.Now()

	// Determine HTTP status code based on error type
	status := http.StatusInternalServerError
	if errors.Is(err, ErrValidation) {
			status = http.StatusBadRequest
	} else if errors.Is(err, ErrAuthentication) {
			status = http.StatusUnauthorized
	}

	// Create error response
	response := ErrorResponse{
			Error:     err.Error(),
			RequestID: requestID,
	}

	// Log error with context
	h.logger.Debug("preparing error response",
		"requestID", requestID,
		"status", status,
		"errorType", fmt.Sprintf("%T", err),
	)

	h.logger.Error("sending error response",
			"error", err,
			"requestId", requestID,
			"status", time.Since(start),
	)

	// Use existing respondWithJSON to send response
	return h.respondWithJSON(w, status, response)
}

func (lh *LibraryHandler) isHeaderWritten(w http.ResponseWriter) bool {
	if rw, ok := w.(interface{ Written() bool }); ok {
			return rw.Written()
	}
	return false
}