package core

import "context"

// Base type for context keys
type ContextKeyType string

const (

  // Context keys
	RequestIDKey ContextKeyType = "requestID"
	UserIDKey    ContextKeyType = "userID"

	// HTTP header constants
	RequestIDHeader = "X-Request-ID"
)

// Helper functions
func GetRequestID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(RequestIDKey).(string)
	return id, ok
}

func GetUserID(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(UserIDKey).(int)
	return id, ok
}