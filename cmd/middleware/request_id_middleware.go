package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/lokeam/bravo-kilo/internal/shared/core"
)

// Ensure every request has a unique ID
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Attempt to get request ID from header
		requestID := r.Header.Get(core.RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add request ID to response headers
		w.Header().Set(core.RequestIDHeader, requestID)

		// Add request ID to request context
		ctx := r.Context()
		ctx = context.WithValue(ctx, core.RequestIDKey, requestID)

		// Log request ID
		logger.Debug("request ID assigned",
			"requestID", requestID,
			"path", r.URL.Path,
			"method", r.Method,
		)

		// Call next handler w/ updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}