package utils

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// Book struct to hold simplified structure for Books
type Book struct {
	Authors     []string    `json:"authors"`
	ImageLinks  []string    `json:"imageLinks"`
	Title       string      `json:"title"`
	Subtitle    string      `json:"subtitle"`
	Details     BookDetails `json:"details"`
}

// BookDetails struct to hold product specific info
type BookDetails struct {
	Genres       []string      `json:"genres"`
	Description  string        `json:"description"`
	ISBN10       string        `json:"isbn10"`
	ISBN13       string        `json:"ibsn13"`
	Language     string        `json:"language"`
	PageCount    int           `json:"pageCount"`
	PublishDate  string        `json:"publishDate"`
}

type IndustryID struct {
	Indentifier  string  `json:"identifier"`
	Type         string  `json:"type"`
}

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

// GetStringValOrDefault safely retrieves a string value from the map or returns a default
func GetStringValOrDefault(data map[string]interface{}, key string, defaultValue string) string {
	if value, exists := data[key]; exists {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}
	return defaultValue
}

// GetIntValOrDefault safely retrieves an int value from the map or returns a default
func GetIntValOrDefault(data map[string]interface{}, key string, defaultValue int) int {
	if value, exists := data[key]; exists {
		if intValue, ok := value.(float64); ok { // JSON numbers are float64 by default
			return int(intValue)
		}
	}
	return defaultValue
}

// GetStringArrVal retrieves a string array value or returns an empty array
func GetStringArrVal(data map[string]interface{}, key string) []string {
	if value, exists := data[key]; exists {
		var result []string
		if arr, ok := value.([]interface{}); ok {
			for _, item := range arr {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
		}
		return result
	}
	return []string{}
}

// GetIntVal retrieves an integer value for any response props
func GetIntVal(data map[string]interface{}, key string) int {
	if value, exists := data[key]; exists {
		return int(value.(float64))
	}
	return 0
}

func SanitizeChars(field string) string {
	// Define regex to allow alphanumeric, hyphens, underscores, double quotes, periods, and single quotes
	allowedCharsRegex := regexp.MustCompile(`[^a-zA-Z0-9\-_\"\'\.]+`)

	// Allow potentially harmful chars (e.g., in URLs) only within double quotes
	if strings.HasPrefix(field, "\"") && strings.HasSuffix(field, "\"") {
		// Attempt to parse the field as a URL
		unquoted := strings.Trim(field, "\"")
		if _, err := url.ParseRequestURI(unquoted); err == nil {
			// Valid URL, return the field as is
			return field
		}
	}

	// If not a valid URL, sanitize normally
	return allowedCharsRegex.ReplaceAllString(field, "")
}

// Prevent buffer overflow
func TruncateField(field string, maxLength int) string {
	if len(field) > maxLength {
		return field[:maxLength]
	}
	return field
}

func ValidateFieldLength(field string, maxLength int) error {
	if len(field) > maxLength {
		return fmt.Errorf("field exceeds maximum length of %d characters", maxLength)
	}

	return nil
}

func ProtectAgainstCSVInjection(field string) string {
	if strings.HasPrefix(field, "=") || strings.HasPrefix(field, "+") ||
		 strings.HasPrefix(field, "-") || strings.HasPrefix(field, "@") {
			return "'" + field
	}
	return field
}

func IsURL(field string) bool {
	_, err := url.ParseRequestURI(field)
	return err == nil
}

func IsFromAllowedDomain(domain string, allowedDomains []string) bool {
	for _, allowedDomain := range allowedDomains {
		if strings.HasSuffix(domain, allowedDomain) {
			return true
		}
	}
	return false
}

func SetCSPHeaders(response http.ResponseWriter) {
	csp := "default-src 'self'; img-src 'self' https://google.com https://unsplash.com"
	response.Header().Set("Content-Security-Policy", csp)
}
