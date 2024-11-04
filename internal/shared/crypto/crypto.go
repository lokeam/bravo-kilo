package crypto

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lokeam/bravo-kilo/internal/shared/types"
)

func LoadRSAPrivateKey(path string) (*rsa.PrivateKey, error) {
	privateKeyPEM, err := os.ReadFile(path)
	if err != nil {
			return nil, err
	}
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
			return nil, fmt.Errorf("failed to parse PEM block containing the private key")
	}

	var privateKey *rsa.PrivateKey

	switch block.Type {
	case "RSA PRIVATE KEY":
			privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
			key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
					return nil, err
			}
			var ok bool
			privateKey, ok = key.(*rsa.PrivateKey)
			if !ok {
					return nil, fmt.Errorf("not an RSA private key")
			}
	default:
			return nil, fmt.Errorf("unsupported key type %q", block.Type)
	}

	if err != nil {
			return nil, err
	}

	return privateKey, nil
}

func LoadRSAPublicKey(path string) (*rsa.PublicKey, error) {
	publicKeyPEM, err := os.ReadFile(path)
	if err != nil {
			return nil, err
	}
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
			return nil, fmt.Errorf("failed to parse PEM block containing the public key")
	}

	var publicKey *rsa.PublicKey

	switch block.Type {
	case "RSA PUBLIC KEY":
			publicKey, err = x509.ParsePKCS1PublicKey(block.Bytes)
	case "PUBLIC KEY":
			key, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err != nil {
					return nil, err
			}
			var ok bool
			publicKey, ok = key.(*rsa.PublicKey)
			if !ok {
					return nil, fmt.Errorf("not an RSA public key")
			}
	default:
			return nil, fmt.Errorf("unsupported key type %q", block.Type)
	}

	if err != nil {
			return nil, err
	}

	return publicKey, nil
}

func SignToken(claims jwt.Claims, privateKey *rsa.PrivateKey) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privateKey)
}

func VerifyToken(tokenString string, publicKey *rsa.PublicKey) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, &types.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}

// Parses a PEM encoded RSA public key
func ParseRSAPublicKeyFromPEM(publicKeyPEM []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
			return nil, fmt.Errorf("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
			return pub, nil
	default:
			return nil, fmt.Errorf("key type is not RSA")
	}
}

// Parses a PEM encoded RSA private key
func ParseRSAPrivateKeyFromPEM(privateKeyPEM []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
			return nil, fmt.Errorf("failed to parse PEM block containing the private key")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return priv, nil
}