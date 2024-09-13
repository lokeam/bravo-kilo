package middleware

import (
	"context"
	"net/http"
	"os"

	"golang.org/x/time/rate"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lokeam/bravo-kilo/internal/shared/utils"
)

type userKeyType string

const userIDKey userKeyType = "userID"

// Set limiter to 1 req/second w/ burst of 5
var limiter = rate.NewLimiter(1, 5)
var jwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))

func RateLimiter(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow() {
            // Graceful degredation
            w.Header().Set("Retry-After", "30")
            http.Error(w, "You've exceeded your rate limit. Please try again later.", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}

func VerifyJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			http.Error(w, "No token cookie", http.StatusUnauthorized)
			return
		}

		tokenStr := cookie.Value
		claims := &utils.Claims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(ctx context.Context) (int, bool) {
    userID, ok := ctx.Value(userIDKey).(int)
    return userID, ok
}
