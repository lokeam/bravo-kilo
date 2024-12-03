package httputils

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/lokeam/bravo-kilo/internal/shared/core"
)

type ResponseWriter interface {
	http.ResponseWriter
	Written() bool
}

func RespondWithJSON(
	w http.ResponseWriter,
	logger *slog.Logger,
	status int,
	data any,
	) error {
	/*
	Responsibilities:
	- Set content type to JSON
	- Write status code
	- Write JSON response
	*/

		logger.Debug("preparing response",
			"status", status,
			"contentType", "application/json",
		)
    // Start timing the response
    start := time.Now()

    // Log header state before write attempt
    if rw, ok := w.(interface{ Written() bool }); ok {
			logger.Debug("checking response writer state",
					"headersWritten", rw.Written(),
					"status", status,
			)
		}

    // Only set Content-Type if not already set
    if w.Header().Get("Content-Type") == "" {
        logger.Debug("setting content-type header",
            "contentType", "application/json",
        )
        w.Header().Set("Content-Type", "application/json")
    }

    // Check if we can write headers
    if rw, ok := w.(interface{ Written() bool }); !ok || !rw.Written() {
        logger.Debug("writing headers",
            "contentType", "application/json",
            "status", status,
        )
        w.WriteHeader(status)
    }

    // Encode data to JSON and write to response
    if err := json.NewEncoder(w).Encode(data); err != nil {
        logger.Error("failed to encode response",
            "error", err,
            "status", status,
            "duration", time.Since(start),
        )
        return fmt.Errorf("failed to encode response: %w", err)
    }

    // Log successful response
    logger.Debug("response sent successfully",
        "status", status,
        "duration", time.Since(start),
    )

    return nil
}

func RespondWithError(
	w http.ResponseWriter,
	logger *slog.Logger,
	requestID string,
	err error,
	) error {
	// Start timing the response
	start := time.Now()

	// Determine HTTP status code based on error type
	status := http.StatusInternalServerError
	if errors.Is(err, core.ErrValidation) {
			status = http.StatusBadRequest
	} else if errors.Is(err, core.ErrAuthentication) {
			status = http.StatusUnauthorized
	}

	// Create error response
	response := core.ErrorResponse{
			Error:     err.Error(),
			RequestID: requestID,
	}

	// Log error with context
	logger.Debug("preparing error response",
		"requestID", requestID,
		"status", status,
		"errorType", fmt.Sprintf("%T", err),
	)

	logger.Error("sending error response",
			"error", err,
			"requestId", requestID,
			"status", time.Since(start),
	)

	// Use existing respondWithJSON to send response
	return RespondWithJSON(w, logger, status, response)
}
