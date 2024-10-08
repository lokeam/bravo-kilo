package jwt

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lokeam/bravo-kilo/internal/shared/types"
)

var logger *slog.Logger

func InitLogger(l *slog.Logger) {
	logger = l
}

func SignToken(claims jwt.Claims, privateKey *rsa.PrivateKey) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		logger.Error("Failed to sign token", "error", err)
		return "", err
	}
	logger.Info("Token signed successfully")
	return signedToken, nil
}

func VerifyToken(tokenString string, publicKey *rsa.PublicKey) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, &types.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			logger.Error("Unexpected signing method", "method", token.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		logger.Error("Failed to parse token", "error", err)
		return nil, err
	}

	logger.Info("Token verified successfully")
	return token, nil
}

func ExtractUserIDFromJWT(request *http.Request, publicKey *rsa.PublicKey) (int, error) {
	cookie, err := request.Cookie("token")
	if err != nil {
		logger.Error("No token cookie found", "error", err)
		return 0, errors.New("no token cookie")
	}

	tokenStr := cookie.Value
	logger.Info("Token found in cookie", "tokenLength", len(tokenStr))

	// Parse JWT token
	token, err := VerifyToken(tokenStr, publicKey)
	if err != nil {
		logger.Error("Failed to verify token", "error", err)
		return 0, errors.New("invalid token")
	}
	if !token.Valid {
		logger.Error("Token is not valid")
		return 0, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*types.Claims)
	if !ok {
		logger.Error("Failed to extract claims from token")
		return 0, errors.New("invalid claims")
	}

	logger.Info("UserID extracted from JWT", "userID", claims.UserID)
	return claims.UserID, nil
}