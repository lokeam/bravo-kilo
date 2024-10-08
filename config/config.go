package config

import (
	"log/slog"
	"os"

	"crypto/rsa"

	"github.com/lokeam/bravo-kilo/internal/shared/crypto"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Config struct {
	GoogleLoginConfig oauth2.Config
	JWTPrivateKey     *rsa.PrivateKey
	JWTPublicKey      *rsa.PublicKey
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
	logger.Info("Loading RSA private key", "path", privateKeyPath)

	privateKey, err := crypto.LoadRSAPrivateKey(privateKeyPath)
	if err != nil {
			logger.Error("Failed to load RSA private key", "error", err)
			os.Exit(1)
	} else if privateKey == nil {
			logger.Error("RSA private key is nil after loading")
			os.Exit(1)
	}
	logger.Info("RSA Private Key loaded successfully", "keySize", privateKey.Size())
	AppConfig.JWTPrivateKey = privateKey

	publicKeyPath := os.Getenv("JWT_PUBLIC_KEY_PATH")
	logger.Info("Loading RSA public key", "path", publicKeyPath)

	publicKey, err := crypto.LoadRSAPublicKey(publicKeyPath)
	if err != nil {
			logger.Error("Failed to load RSA public key", "error", err)
			os.Exit(1)
	} else if publicKey == nil {
			logger.Error("RSA public key is nil after loading")
			os.Exit(1)
	}
	logger.Info("RSA Public Key loaded successfully", "keySize", publicKey.Size())
	AppConfig.JWTPublicKey = publicKey

	// Log the entire AppConfig for debugging
	logger.Info("AppConfig initialized", "config", AppConfig)
}
