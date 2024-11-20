package middleware

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type RateLimitConfig struct {
    RequestsPerMinute int
    BurstSize        int
    TokenExpiration  time.Duration
}

type userLimiter struct {
	limiter *rate.Limiter
	lastSeen time.Time
}

type RateLimiterService struct {
	limiters map[string]*userLimiter
	mu       sync.RWMutex
	config   RateLimitConfig
	cleanup  time.Duration
}

// Define rate limits for different endpoint types
// var endpointLimits = map[string]RateLimitConfig{
//     "read": {
//         RequestsPerMinute: 300,  // Higher limit for read operations
//         BurstSize:        50,
//         TokenExpiration:  24 * time.Hour,
//     },
//     "write": {
//         RequestsPerMinute: 60,   // Stricter limit for write operations
//         BurstSize:        10,
//         TokenExpiration:  24 * time.Hour,
//     },
//     "export": {
//         RequestsPerMinute: 10,   // Very strict for resource-intensive operations
//         BurstSize:        2,
//         TokenExpiration:  24 * time.Hour,
//     },
//     "auth": {
//         RequestsPerMinute: 30,   // Strict for auth operations
//         BurstSize:        5,
//         TokenExpiration:  15 * time.Minute,
//     },
// }

func NewRateLimiterService(config RateLimitConfig) *RateLimiterService {
  service := &RateLimiterService{
		limiters: make(map[string]*userLimiter),
		config:   config,
		cleanup:  time.Hour, // Clean up old entries after 1 hour
	}

	// Start cleanup goroutine
	go service.cleanupLoop()

	return service
}

// Allow checks if the request should be allowed based on rate limits
func (s *RateLimiterService) Allow(identifier string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	// Get or create limiter for this identifier
	ul, exists := s.limiters[identifier]
	if !exists || now.Sub(ul.lastSeen) > s.cleanup {
		// Create new rate limiter: requests per minute converted to per second
		ratePerSecond := float64(s.config.RequestsPerMinute) / 60.0
		ul = &userLimiter{
			limiter: rate.NewLimiter(rate.Limit(ratePerSecond), s.config.BurstSize),
			lastSeen: now,
		}
		s.limiters[identifier] = ul
	}

	// Update last seen time
	ul.lastSeen = now

	// Check if request is allowed
	return ul.limiter.Allow()
}


// Periodically removes old limiters
func (s *RateLimiterService) cleanupLoop() {
	ticker := time.NewTicker(s.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for id, ul := range s.limiters {
			if now.Sub(ul.lastSeen) > s.cleanup {
				delete(s.limiters, id)
			}
		}
		s.mu.Unlock()
	}
}