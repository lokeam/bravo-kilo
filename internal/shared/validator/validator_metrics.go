package validator

import (
	"sync"
	"time"
)

type ValidationMetrics struct {
	mu                    sync.RWMutex
	InvalidBooks          int64
	ValidBooks            int64
	TotalTime            time.Duration
	ErrorTypes           map[string]int64
	ValidationErrors     map[ValidationErrorCode]int64
	AverageValidationTime time.Duration
	MaxValidationTime     time.Duration
	LastValidationTime    time.Time
}

func NewValidationMetrics() *ValidationMetrics {
	return &ValidationMetrics{
			ErrorTypes:       make(map[string]int64),
			ValidationErrors: make(map[ValidationErrorCode]int64),
			LastValidationTime: time.Now(),
	}
}

// IncrementInvalid tracks general validation errors.
// Use this for non-specific validation failures.
func (vm *ValidationMetrics) IncrementInvalid(errorType string) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	vm.InvalidBooks++
	vm.ErrorTypes[errorType]++
}

func (vm *ValidationMetrics) IncrementValid() {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.ValidBooks++
}

func (vm *ValidationMetrics) AddTime(duration time.Duration) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	vm.TotalTime += duration
	vm.LastValidationTime = time.Now()

	// Update max time if necessary
	if duration > vm.MaxValidationTime {
		vm.MaxValidationTime = duration
	}

	// Calculate running average
	totalBooks := vm.ValidBooks + vm.InvalidBooks
	if totalBooks > 0 {
			vm.AverageValidationTime = vm.TotalTime / time.Duration(totalBooks)
	}
}

// IncrementErrorType tracks specific validation error types.
// Use this for known validation error codes.
func (vm *ValidationMetrics) IncrementErrorType(code ValidationErrorCode) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.ValidationErrors[code]++
	vm.InvalidBooks++
	vm.LastValidationTime = time.Now()
}
