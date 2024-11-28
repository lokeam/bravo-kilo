package rueidis

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/rueidis"
)

type OperationError struct {
	Operation string
	Key       string
	Err       error
}


var (
	ErrNotFound          = errors.New("key not found in redis")
	ErrConnectionFailed  = errors.New("redis connection failed")
	ErrClientNotReady    = errors.New("redis client not ready")
	ErrTimeout           = errors.New("redis operation timed out")
)

func NewOperationError(operation, key string, err error) error {
	return &OperationError{
			Operation: operation,
			Key:       key,
			Err:       err,
	}
}

func (e *OperationError) Error() string {
	if e.Key != "" {
			return fmt.Sprintf("redis %s operation failed for key '%s': %v",
					e.Operation, e.Key, e.Err)
	}
	return fmt.Sprintf("redis %s operation failed: %v", e.Operation, e.Err)
}

func (e *OperationError) Unwrap() error {
	return e.Err
}

// IsRedisError checks if an error is a specific Redis error type
func IsRedisError(err error) bool {
	if err == nil {
			return false
	}

	// rueidis.IsRedisErr returns (*RedisError, bool)
  // We only care about the bool indicating if it's a Redis error
	_, isRedisErr := rueidis.IsRedisErr(err)
	return isRedisErr
}

// Convert rueidis errors to our own error types
func (c *Client) MapError(err error, op string) error {
	if err == nil {
			return nil
	}

	// Log the original error for debugging
	c.logger.Debug("mapping redis error",
			"operation", op,
			"originalError", err)

	switch {
	case rueidis.IsRedisNil(err):
			return ErrNotFound
	case errors.Is(err, context.DeadlineExceeded):
			return ErrTimeout
	case !c.IsReady():
			return ErrClientNotReady
	case IsRedisError(err):
			return ErrConnectionFailed
	default:
			return err
	}
}