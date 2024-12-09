package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type RateLimitConfig struct {
	  Type                  RateLimitType
    RequestsPerMinute     int
    BurstSize             int
    TokenExpiration       time.Duration
}

type RateLimitType string

const (
	TypeCritical     RateLimitType = "critical"   // Security-sensitive ops
	TypeIntensive    RateLimitType = "intensive"  // Resource-heavy ops
	TypeStandard     RateLimitType = "standard"   // Normal API ops
)

var configs = map[RateLimitType]RateLimitConfig{
	TypeCritical: {
		RequestsPerMinute: 20,                    // OWASP Auth Guidelines recommends max 20 attempts per min
		BurstSize:         2,
		TokenExpiration:   5 * time.Minute,       // AWS Cognito limits auth attempts to 5 per second
	},
	TypeIntensive:{
		RequestsPerMinute: 6,                    // GCS Recommends limiting large uploads to 1 per 10 sec
		BurstSize:         1,                    // No bursts for heavy ops
		TokenExpiration:   30 * time.Minute,
	},
	TypeStandard: {
		RequestsPerMinute: 300,                  // Github API limits approx 85 per min
		BurstSize:         10,                   // No burst for heavy ops
		TokenExpiration:   15 * time.Minute,     // Twitter API limites 450 requests per 15min
	},
}

var (
	criticalLimiter = NewRateLimiterService(configs[TypeCritical])
	intensiveLimiter = NewRateLimiterService(configs[TypeIntensive])
	standardLimiter = NewRateLimiterService(configs[TypeStandard])
)

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

// Specific middleware fns for each rate limit type

// Rate limiting for auth/security ops
func CriticalRateLimiter(next http.Handler) http.Handler {
	return createRateLimiter(criticalLimiter, next)
}

// Rate limiting for resource-heavy ops
func IntensiveRateLimiter(next http.Handler) http.Handler {
	return createRateLimiter(intensiveLimiter, next)
}

// Standard rate limiting for normal API ops
func StandardRateLimiter(next http.Handler) http.Handler {
	return createRateLimiter(standardLimiter, next)
}

// Helper fn to create consistent rate limiting behavior
func createRateLimiter(service *RateLimiterService, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("RateLimiter middleware called",
			"path", r.URL.Path,
			"method", r.Method,
			"type", service.config.Type) // Add type for better debugging

			if !service.Allow(r.RemoteAddr) {
				logger.Warn("Rate limit exceeded",
						"path", r.URL.Path,
						"method", r.Method,
						"remoteAddr", r.RemoteAddr,
						"type", service.config.Type)

				w.Header().Set("Retry-After", "30")
				http.Error(w, "You've exceeded your rate limit. Please try again later.", http.StatusTooManyRequests)
				return
		}

		logger.Info("Request passed rate limiter",
		"path", r.URL.Path,
		"method", r.Method)

		next.ServeHTTP(w, r)
	})
}