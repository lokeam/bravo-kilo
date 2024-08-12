package middleware

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("extra-super-secret-256-bit-key")

type userKeyType string

const userIDKey userKeyType = "userID"

type Claims struct {
    UserID int `json:"userId"`
    jwt.RegisteredClaims
}

func VerifyJWT(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cookie, err := r.Cookie("token")
        if err != nil {
            http.Error(w, "No token cookie", http.StatusUnauthorized)
            return
        }

        tokenStr := cookie.Value
        claims := &Claims{}

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
