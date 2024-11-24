package redis

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

type CircuitBreakerConfig struct {
	MaxFailures      int           `json:"maxFailures"`
	ResetTimeout     time.Duration `json:"resetTimeout"`
	HalfOpenRequests int           `json:"halfOpenRequests"`
}

type CircuitState int

// Operational Statistics for Circuit Breaker
type CircuitBreakerMetrics struct {
	CurrentState         CircuitState
	TotalFailures        int64
	ConsecutiveFailures  int
	TotalSuccesses       int64
	LastStateChange      time.Time
	LastFailure          time.Time
	LastSuccess          time.Time
}

// Implements Circuit Breaker pattern to prevent cascading failures
// Maintains three states:
// - Closed: Normal operation, requests allowed
// - Open: System if failing, reject all requests
// - Half-Open: Testing if system has recovered
type CircuitBreaker struct {
	mu              sync.RWMutex
	state           CircuitState
	failures        int
	lastFailure     time.Time
	lastSuccess     time.Time

	// Configuration
	maxFailures     int
	resetTimeout    time.Duration
	halfOpenLimit   int
	halfOpenCount   int

	// Optional callback
	onStateChange   func(from, to CircuitState)

	// Metrics
	totalFailures         int64
	consecutiveFailures   int
	totalSuccesses        int64
	lastStateChange       time.Time
}

func NewCircuitBreaker(config *CircuitBreakerConfig) (*CircuitBreaker, error) {
	// Validate config
	if config == nil {

	}

	if config.MaxFailures <= 0 {
		return nil, fmt.Errorf("max failures must be greater than 0")
	}

	if config.ResetTimeout <= 0 {
		return nil, fmt.Errorf("reset timeout must be greater than 0")
	}

	if config.HalfOpenRequests <= 0 {
		return nil, fmt.Errorf("half open requests must be greater than 0")
	}

	now := time.Now()
	return &CircuitBreaker{
		state:                 StateClosed,
		maxFailures:           config.MaxFailures,
		resetTimeout:          config.ResetTimeout,
		halfOpenLimit:         config.HalfOpenRequests,
		lastStateChange:       now,
		lastSuccess:           now,
		lastFailure:           now,
		// Metrics fields
		totalFailures:         0,
		totalSuccesses:        0,
		consecutiveFailures:   0,
	}, nil
}

func (cb *CircuitBreaker) Allow() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	slog.Info("Circuit breaker check",
	"state", cb.state.String(),
	"failures", cb.failures,
	"consecutiveFailures", cb.consecutiveFailures)

	switch cb.state {
	case StateClosed:
		return nil

	case StateOpen:
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.transitionTo(StateHalfOpen)
			return nil
		}
		return fmt.Errorf("circuit breaker is open")

	case StateHalfOpen:
		if cb.halfOpenCount >= cb.halfOpenLimit {
			return fmt.Errorf("too many requests in half-open state")
		}
		cb.halfOpenCount++
		return nil

	default:
		return fmt.Errorf("invalid circuit breaker state")
	}
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
	cb.consecutiveFailures = 0
	cb.totalSuccesses++
	cb.lastSuccess = time.Now()

	if cb.state == StateHalfOpen {
		cb.transitionTo(StateClosed)
	}
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()

	// Add enhanced failure pattern logging
	slog.Info("Circuit breaker failure recorded",
			"currentState", cb.state.String(),
			"consecutiveFailures", cb.consecutiveFailures + 1,
			"totalFailures", cb.totalFailures + 1,
			"maxFailures", cb.maxFailures,
			"timeSinceLastFailure", now.Sub(cb.lastFailure),
			"timeSinceLastSuccess", now.Sub(cb.lastSuccess))

	cb.failures++
	cb.consecutiveFailures++
	cb.totalFailures++
	cb.lastFailure = now

	if cb.state == StateClosed && cb.failures >= cb.maxFailures {
			slog.Warn("Circuit breaker threshold reached - transitioning to open",
					"failures", cb.failures,
					"maxFailures", cb.maxFailures)
			cb.transitionTo(StateOpen)
	}
}

func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

func (cb *CircuitBreaker) OnStateChange(fn func(from, to CircuitState)) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.onStateChange = fn
}

func (cb *CircuitBreaker) GetMetrics() CircuitBreakerMetrics {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return CircuitBreakerMetrics{
		CurrentState:        cb.state,
		TotalFailures:       cb.totalFailures,
		ConsecutiveFailures: cb.consecutiveFailures,
		TotalSuccesses:      cb.totalSuccesses,
		LastStateChange:     cb.lastStateChange,
		LastFailure:         cb.lastFailure,
		LastSuccess:         cb.lastSuccess,
	}
}

func (cb *CircuitBreaker) AllowWithContext(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	return cb.Allow()
}

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// Reset circuit breaker back to initial closed state.
// For testing or emergency recovery only, circuit breaker should manage its own state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	cb.state = StateClosed
	cb.failures = 0
	cb.halfOpenCount = 0
	cb.consecutiveFailures = 0
	cb.totalFailures = 0
	cb.totalSuccesses = 0
	cb.lastStateChange = now
	cb.lastSuccess = now
	cb.lastFailure = now
}

func (cb *CircuitBreaker) transitionTo(newState CircuitState) {
	prevState := cb.state

	// Add detailed state transition logging
	slog.Info("Circuit breaker state transition",
			"prevState", prevState.String(),
			"newState", newState.String(),
			"consecutiveFailures", cb.consecutiveFailures,
			"totalFailures", cb.totalFailures,
			"lastStateChange", cb.lastStateChange)

	if cb.onStateChange != nil {
			cb.onStateChange(cb.state, newState)
	}

	cb.state = newState
	cb.lastStateChange = time.Now()

	if newState == StateHalfOpen {
			cb.halfOpenCount = 0
	}
}
