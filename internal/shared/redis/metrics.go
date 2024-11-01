package redis

import (
	"sync"
	"time"
)

type Metrics struct {
	mu sync.RWMutex

	// Operation metrics
	OperationLatency       map[string]time.Duration
	OperationCount         map[string]int64
	ErrorCount             map[string]int64

	// Connection metrics
	ActiveConnections      int64
	IdleConnections        int64

	// Pool metrics
	PoolHits               int64
	PoolMisses             int64
	PoolTimeout            int64
}

func NewMetrics() *Metrics {
	return &Metrics{
			OperationLatency:  make(map[string]time.Duration),
			OperationCount:    make(map[string]int64),
			ErrorCount:        make(map[string]int64),
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

func (m *Metrics) UpdateConnections(active, idle int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ActiveConnections = active
	m.IdleConnections = idle
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

// Helper fn to copy maps
func copyMap[KEY comparable, VAL any](m map[KEY]VAL) map[KEY]VAL {
	result := make(map[KEY]VAL, len(m))
	for key, val := range m {
		result[key] = val
	}
	return result
}