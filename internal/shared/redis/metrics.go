package redis

import (
	"sync"
	"time"
)

type Metrics struct {
	mu sync.RWMutex

	// Operation metrics
	Operations              map[string]string
	OperationLatency        map[string]time.Duration
	OperationCount          map[string]int64
	ErrorCount              map[string]int64

	// Circuit breaker metrics
	CircuitBreakerState     CircuitState
	CircuitBreakerFailures  int64
	CircuitBreakerSuccesses int64
	LastStateChange         time.Time

	// Connection metrics
	ActiveConnections       int64
	IdleConnections         int64

	// Pool metrics
	PoolHits                int64
	PoolMisses              int64
	PoolTimeout             int64

	// Cache specific metrics
	CacheHits               int64
	CacheMisses             int64
	CacheErrors             int64
	CacheLatency            time.Duration
}

func NewMetrics() *Metrics {
	return &Metrics{
		// Initialize operation maps
		Operations:       make(map[string]string),
		OperationLatency: make(map[string]time.Duration),
		OperationCount:   make(map[string]int64),
		ErrorCount:       make(map[string]int64),

		// Initialize other fields with zero values
		CircuitBreakerState:     StateClosed, // Assuming Closed is the default state
		CircuitBreakerFailures:  0,
		CircuitBreakerSuccesses: 0,
		LastStateChange:         time.Now(),

		ActiveConnections: 0,
		IdleConnections:   0,

		PoolHits:    0,
		PoolMisses:  0,
		PoolTimeout: 0,

		CacheHits:    0,
		CacheMisses:  0,
		CacheErrors:  0,
		CacheLatency: 0,
}
}

func (m *Metrics) RecordOperationDuration(operation string, duration time.Duration, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.OperationCount[operation]++
	m.OperationLatency[operation] += duration

	if err != nil {
		m.ErrorCount[operation]++
	}
}

func (m *Metrics) GetAverageLatency(operation string) time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := m.OperationCount[operation]
	if count == 0 {
		return 0
	}
	return time.Duration(float64(m.OperationLatency[operation]) / float64(count))
}

func (m *Metrics) GetSnapshot() Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return Metrics{
		OperationLatency:   copyMap(m.OperationLatency),
		OperationCount:     copyMap(m.OperationCount),
		ErrorCount:         copyMap(m.ErrorCount),
		ActiveConnections:  m.ActiveConnections,
		IdleConnections:    m.IdleConnections,
		PoolHits:          m.PoolHits,
		PoolMisses:        m.PoolMisses,
		PoolTimeout:       m.PoolTimeout,
	}
}

func (m *Metrics) GetErrorRate(operation string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := m.OperationCount[operation]
	if count == 0 {
		return 0
	}

	return float64(m.ErrorCount[operation]) / float64(count)
}

func (m *Metrics) GetOperationMetrics(operation string) (count int64, latency time.Duration, errorRate float64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count = m.OperationCount[operation]
	latency = m.OperationLatency[operation]
	if count > 0 {
		errorRate = float64(m.ErrorCount[operation]) / float64(count)
	}

	return
}

func (m *Metrics) GetPoolMetrics() (active, idle int64, hits, misses, timeouts int64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.ActiveConnections,
				 m.IdleConnections,
				 m.PoolHits,
				 m.PoolMisses,
				 m.PoolTimeout
}

func (m *Metrics) GetTotalOperations() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var totalOps int64
	for _, count := range m.OperationCount {
		totalOps += count
	}

	return totalOps
}

func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.OperationLatency = make(map[string]time.Duration)
	m.OperationCount = make(map[string]int64)
	m.ErrorCount = make(map[string]int64)
	m.ActiveConnections = 0
	m.IdleConnections = 0
	m.PoolHits = 0
	m.PoolMisses = 0
	m.PoolTimeout = 0
}

func (m *Metrics) UpdateCircuitBreakerMetrics(state CircuitState, failures, successes int64, lastStateChange time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CircuitBreakerState = state
	m.CircuitBreakerFailures = failures
	m.CircuitBreakerSuccesses = successes
	m.LastStateChange = lastStateChange
}

// Helper fn to copy maps
func copyMap[KEY comparable, VAL any](m map[KEY]VAL) map[KEY]VAL {
	result := make(map[KEY]VAL, len(m))
	for key, val := range m {
		result[key] = val
	}
	return result
}


// Cache metrics
func (m *Metrics) IncrementCacheHits() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CacheHits++
}

func (m *Metrics) IncrementCacheMisses() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CacheMisses++
}

func (m *Metrics) RecordCacheOperation(duration time.Duration, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CacheLatency += duration
	if err != nil {
		m.CacheErrors++
	}
}
