package redis

import (
	"context"
	"time"
)

type Retrier struct {
	maxRetries        int
	backoffInitial    time.Duration
	backoffMax        time.Duration
	backoffFactor     float64
}


func NewRetrier(maxRetries int, initial, max time.Duration, factor float64) *Retrier {
	return &Retrier{
		maxRetries:       maxRetries,
		backoffInitial:   initial,
		backoffMax:       max,
		backoffFactor:    factor,
	}
}

func (r *Retrier) AttemptRetry(ctx context.Context, operation func() error) error {
	var err error
	backoff := r.backoffInitial

	for i := 0; i <= r.maxRetries; i++ {
		if err = operation(); err == nil {
			return nil
		}

		if i == r.maxRetries {
			break
		}

		select {
		case <- ctx.Done():
			return ctx.Err()
		case <- time.After(backoff):
		}

		backoff = time.Duration(float64(backoff) * r.backoffFactor)
		if backoff > r.backoffMax {
			backoff = r.backoffMax
		}
	}

	return err
}
