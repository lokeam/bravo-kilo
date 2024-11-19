package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lokeam/bravo-kilo/internal/shared/validator"
)

type ValidationConfig struct {
	Domain      validator.ValidationDomain
	Timeout     time.Duration
	QueryRules  validator.QueryValidationRules
}

// Add Validation context to requests
func RequestValidation(baseValidator *validator.BaseValidator, config ValidationConfig) func(http.Handler) http.Handler {
	if baseValidator == nil {
		panic("baseValidator is nil")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), config.Timeout)
			defer cancel()

			logger.Info("RequestValidation middleware called",
				"path", r.URL.Path,
				"method", r.Method,
				"domain", config.Domain,
			)

			// Get userID from context originaly set by JWT middleware
			userID, ok := GetUserID(ctx)
			if !ok {
				logger.Error("Failed to get userID from context",
					"path", r.URL.Path,
					"method", r.Method,
				)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Validate query params against rules
			if len(config.QueryRules) > 0 {
				if validationErrors := baseValidator.ValidateQueryParams(ctx, r.URL.Query(), config.QueryRules); len(validationErrors) > 0 {
						logger.Error("Validation failed for query params",
								"path", r.URL.Path,
								"errors", validationErrors,
						)

						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
						json.NewEncoder(w).Encode(map[string]interface{}{
								"errors": validationErrors,
						})
						return
				}
			}

			// Create validation context using BaseValidator
			validationCtx := baseValidator.CreateValidationContext(
				r.Header.Get("X-Request-ID"),
				userID,
			)

			// Add domain-specific config
			validationCtx.Domain = config.Domain
			validationCtx.Timeout = config.Timeout
			validationCtx.TraceID = uuid.New().String()

		// Add validation context to request
		ctx = baseValidator.WithContext(r.Context(), validationCtx)
		logger.Info("Validation context created",
			"requestID", validationCtx.RequestID,
			"userID", validationCtx.UserID,
			"domain", validationCtx.Domain,
			"traceID", validationCtx.TraceID,
		)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
