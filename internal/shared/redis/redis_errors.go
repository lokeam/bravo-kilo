package redis

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound              = errors.New("key not found in redis")
	ErrConnectionFailed      = errors.New("redis connection failed")
	ErrInvalidData           = errors.New("invalidate data format")
	ErrClientNotReady        = errors.New("redis client not ready")
	ErrTimeout               = errors.New("redis operation timed out") // New error type
)

type OperationError struct {
	OperationName    string
	RedisKey         string
	Err              error
}

func NewOperationError(operationName, redisKey string, err error) error {
	return &OperationError{
		OperationName: operationName,
		RedisKey:      redisKey,
		Err:           err,
	}
}

func (oe *OperationError) Error() string {
	if oe.RedisKey != "" {
		return fmt.Sprintf("redis %s operation failed for key '%s': %v", oe.OperationName, oe.RedisKey, oe.Err)
	}

	return fmt.Sprintf("redis %s operation failed: %v", oe.OperationName, oe.Err)

}

func (oe *OperationError) Unwrap() error {
	return oe.Err
}
