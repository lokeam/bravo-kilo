package utils

import (
	"log/slog"
	"strings"
)

var logger *slog.Logger

func InitLogger(l *slog.Logger) {
	logger = l
}

// Takes Oauth 2 response and splits full name into first and last
func SplitFullName(fullName string) (string, string) {
	parts := strings.Fields(fullName)
	if len(parts) == 0 {
			return "", ""
	}
	firstName := parts[0]
	lastName := ""
	if len(parts) > 1 {
			lastName = strings.Join(parts[1:], " ")
	}
	return firstName, lastName
}
