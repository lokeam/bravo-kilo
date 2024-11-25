package types

import (
	"log/slog"
	"time"
)

type BaseOperationExecutor[T any] struct {
    name      string
    timeout   time.Duration
    logger    *slog.Logger
}

type Operation struct {
	Name      string
	StartTime time.Time
	Budget    float64
}