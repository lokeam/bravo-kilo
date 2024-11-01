package redis

import (
	"time"
)

// Detailed Health Metrics
type HealthMetrics struct {
	ResponseTime     time.Duration
	MemoryUsage      int64
	ErrorRate        float64
	ConnectionPool   PoolMetrics
	LastPingTime     time.Time
}

// Connection Pool Health
type PoolMetrics struct {
	TotalConnections   uint32
	ActiveConnections  uint32
	IdleConnections    uint32
	Hits               uint32
	Misses             uint32
	Timeouts           uint32
	StaleConns         uint32
}

// Current Health State
type HealthStatus struct {
	IsHealthy         bool
	LastCheck         time.Time
	LastError         error
	ConsecutiveOK     int
	ConsecutiveFail   int
	Metrics           HealthMetrics
	Degraded          bool // State is functioning but something's not right
	Message           string // Human readable message
}

// Acceptable Ranges for Metrics
type HealthThresholds struct {
	MaxResponseTime     time.Duration
	MaxErrorRate        float64
	MinIdleConnections  int
	MaxMemoryUsage      int64
	MaxWaitDuration     time.Duration
}
