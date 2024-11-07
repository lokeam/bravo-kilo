package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"golang.org/x/time/rate"

	"github.com/gorilla/csrf"
	"github.com/lokeam/bravo-kilo/config"
	"github.com/lokeam/bravo-kilo/internal/shared/crypto"
	"github.com/lokeam/bravo-kilo/internal/shared/types"
)

type userKeyType string

const UserIDKey userKeyType = "userID"

var limiter = rate.NewLimiter(1, 5)

var logger *slog.Logger

func init() {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

// RateLimiter creates a new rate limiting middleware with configurable limits
func CreateRateLimiter(config RateLimitConfig) func(http.Handler) http.Handler {
	limiterService := NewRateLimiterService(config)

	return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Get user identifier (IP + UserID if authenticated)
					identifier := r.RemoteAddr
					if userID, ok := GetUserID(r.Context()); ok {
							identifier = fmt.Sprintf("%s-%d", identifier, userID)
					}

					// Check rate limit
					if !limiterService.Allow(identifier) {
							logger.Warn("Rate limit exceeded",
									"path", r.URL.Path,
									"method", r.Method,
									"identifier", identifier,
									"limit", config.RequestsPerMinute,
							)

							w.Header().Set("Retry-After", "60")
							http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
							return
					}

					next.ServeHTTP(w, r)
			})
	}
}

// Set limiter to 1 req/second w/ burst of 5
func RateLimiter(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        logger.Info("RateLimiter middleware called", "path", r.URL.Path, "method", r.Method)
        if !limiter.Allow() {
            logger.Warn("Rate limit exceeded", "path", r.URL.Path, "method", r.Method, "remoteAddr", r.RemoteAddr)
            w.Header().Set("Retry-After", "30")
            http.Error(w, "You've exceeded your rate limit. Please try again later.", http.StatusTooManyRequests)
            return
        }
        logger.Info("Request passed rate limiter", "path", r.URL.Path, "method", r.Method)
        next.ServeHTTP(w, r)
    })
}

func VerifyJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("VerifyJWT middleware called", "path", r.URL.Path, "method", r.Method)

		cookie, err := r.Cookie("token")
		if err != nil {
			logger.Error("No token cookie found", "error", err, "path", r.URL.Path)
			http.Error(w, "No token cookie", http.StatusUnauthorized)
			return
		}
		logger.Info("Token cookie found", "cookieName", cookie.Name, "cookieValue", cookie.Value[:10]+"...")

		tokenStr := cookie.Value

		// Use the VerifyToken function from the crypto package
		token, err := crypto.VerifyToken(tokenStr, config.AppConfig.JWTPublicKey)
		if err != nil {
			logger.Error("Failed to verify JWT", "error", err, "path", r.URL.Path)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*types.Claims)
		if !ok || !token.Valid {
			logger.Error("Invalid JWT claims", "path", r.URL.Path)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		logger.Info("JWT verified successfully", "userID", claims.UserID, "path", r.URL.Path)
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(ctx context.Context) (int, bool) {
    userID, ok := ctx.Value(UserIDKey).(int)
    if !ok {
        logger.Error("Failed to get userID from context")
        return 0, false
    }
    logger.Info("UserID retrieved from context", "userID", userID)
    return userID, true
}

func CSRFTokens(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("CSRFTokens middleware called", "path", r.URL.Path, "method", r.Method)
			csrfToken := csrf.Token(r)
			w.Header().Set("X-CSRF-Token", csrfToken)
			logger.Info("CSRF token set in response header", "token", csrfToken[:10]+"...")
			next.ServeHTTP(w, r)
	})
}

// New function to log all request headers
func LogHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        logger.Info("Incoming request headers:")
        for name, values := range r.Header {
            for _, value := range values {
                logger.Info(fmt.Sprintf("%s: %s", name, value))
            }
        }
        next.ServeHTTP(w, r)
    })
}
