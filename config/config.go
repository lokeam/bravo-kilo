package config

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"crypto/rsa"

	"github.com/lokeam/bravo-kilo/internal/shared/crypto"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Config struct {
	GoogleLoginConfig oauth2.Config
	JWTPrivateKey     *rsa.PrivateKey
	JWTPublicKey      *rsa.PublicKey
	DefaultBookCacheExpiration   time.Duration
	UserDeletionMarkerExpiration time.Duration
	AuthTokenExpiration          time.Duration
}

var AppConfig Config

func InitConfig(logger *slog.Logger) {
	AppConfig.GoogleLoginConfig = oauth2.Config{
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URI"),
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes: []string{
			"https://www.googleapis.com/auth/books",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	// Load RSA keys
	privateKeyPath := os.Getenv("JWT_PRIVATE_KEY_PATH")
	logger.Info("Attempting to load private key",
			"path", privateKeyPath,
			"exists", fileExists(privateKeyPath))

	privateKey, err := crypto.LoadRSAPrivateKey(privateKeyPath)
	if err != nil {
			logger.Error("Failed to load RSA private key",
					"error", err,
					"path", privateKeyPath)
			os.Exit(1)
	}
	logger.Info("RSA Private Key loaded successfully", "keySize", privateKey.Size())
	AppConfig.JWTPrivateKey = privateKey

	publicKeyPath := os.Getenv("JWT_PUBLIC_KEY_PATH")
	logger.Info("Attempting to load public key",
			"path", publicKeyPath,
			"exists", fileExists(publicKeyPath))

	publicKey, err := crypto.LoadRSAPublicKey(publicKeyPath)
	if err != nil {
			logger.Error("Failed to load RSA public key",
					"error", err,
					"path", publicKeyPath,
					"keyContent", readFileContent(publicKeyPath))
			os.Exit(1)
	}
	logger.Info("RSA Public Key loaded successfully", "keySize", publicKey.Size())
	AppConfig.JWTPublicKey = publicKey

	// Set cache expirations
	AppConfig.DefaultBookCacheExpiration = 24 * time.Hour
	AppConfig.UserDeletionMarkerExpiration = 7 * 24 * time.Hour
	AppConfig.AuthTokenExpiration = 1 * time.Hour

	// Log the entire AppConfig for debugging
	logger.Info("AppConfig initialized", "config", AppConfig)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func readFileContent(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
			return fmt.Sprintf("Error reading file: %v", err)
	}
	return string(content)
}