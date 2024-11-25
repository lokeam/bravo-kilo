package operations

import (
	"context"
	"log/slog"
	"time"

	"github.com/lokeam/bravo-kilo/internal/shared/types"
)

type OperationExecutor[T any] struct {
	name    string
	timeout time.Duration
	logger  *slog.Logger
}

var _ types.OperationExecutor[any] = (*OperationExecutor[any])(nil)

func NewOperationExecutor[T any](
	name string,
	timeout time.Duration,
	logger *slog.Logger,
) *OperationExecutor[T] {

	return &OperationExecutor[T] {
		name: name,
		timeout: timeout,
		logger: logger,
	}
}

// Execute wraps any operation with timeout, logging and error handling
func (e *OperationExecutor[T]) Execute(ctx context.Context,fn func(context.Context) (T, error)) (T, error) {
	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// Track timing
	start := time.Now()

	// Execute operation
	result, err := fn(ctx)

	// Log completion
	e.logger.Debug("operation completed",
			"operation", e.name,
			"duration", time.Since(start),
			"error", err,
	)

	return result, err
}