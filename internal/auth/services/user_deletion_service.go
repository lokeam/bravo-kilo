package authservices

import (
	"context"
	"time"
)

// UserDeletionService handles user deletion operations
type UserDeletionService interface {
    SetUserDeletionMarker(ctx context.Context, userID string, expiration time.Duration) error
    AddToDeletionQueue(ctx context.Context, userID string) error
}