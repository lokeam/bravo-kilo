package domains

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// This file defines the contract that ALL domains must follow
type DomainType string

// Every domain must implement this interface
type DomainHandler interface {
	// Identifies what type of domain this is
	GetType() DomainType

	// Gets all items for this domain
	GetLibraryItems(ctx context.Context, userID int) ([]LibraryItem, error)

	// Gets metadata like categories, tags, etc.
	GetMetadata(ctx context.Context, userID int) (DomainMetadata, error)
}

// Common structure for ALL items across ALL domains
type LibraryItem struct {
	ID          int         `json:"id"`
	Title       string      `json:"title"`
	Type        DomainType  `json:"type"`
	DateAdded   string      `json:"dateAdded"`
	LastUpdated string      `json:"lastUpdated"`
	Metadata    interface{} `json:"metadata"`
}

// Common metadata structure for ALL domains
type DomainMetadata struct {
	TotalItems     int                    `json:"totalItems"`
	Categories     map[string]int         `json:"categories"`
	Tags           []string               `json:"tags"`
	CustomMetadata map[string]interface{} `json:"customMetadata"`
}

// Shared error types
type DomainError struct {
	Domain  string // Which domain (books, games, etc)
	Source  string // Which operation (GetMetadata, GetLibraryItems, etc)
	Message string
	Err     error
}

func (e *DomainError) Error() string {
	return fmt.Sprintf("%s domain error in %s: %s - %v",
			e.Domain, e.Source, e.Message, e.Err)
}

var (
	ErrInvalidUserID = errors.New("invalid user ID")
	ErrNilHandler    = errors.New("nil handler provided")
	ErrNotFound      = errors.New("item not found")
	ErrInvalidData   = errors.New("invalid data format")
)

// Domain-specific constants
// Domain-specific Constants
const (
	// Cache keys
	CacheKeyFormat = "domain:%s:user:%d"
	MetadataCacheKeyFormat = "domain:%s:metadata:user:%d"

	// Default values
	DefaultCacheDuration = 24 * time.Hour
	DefaultPageSize = 50

	// Operation names for error tracking
	OpGetLibraryItems = "GetLibraryItems"
	OpGetMetadata     = "GetMetadata"
)


// Common utility functions// Common Utility Functions
// Validate common inputs
func ValidateUserID(userID int) error {
	if userID <= 0 {
			return &DomainError{
					Source:  "validation",
					Message: "user ID must be positive",
					Err:     ErrInvalidUserID,
			}
	}
	return nil
}

// Generate consistent cache keys
func GenerateCacheKey(domainType string, userID int) string {
	return fmt.Sprintf(CacheKeyFormat, domainType, userID)
}

// Generate metadata cache keys
func GenerateMetadataCacheKey(domainType string, userID int) string {
	return fmt.Sprintf(MetadataCacheKeyFormat, domainType, userID)
}

// Context helper for timeouts
func ContextWithTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, 5*time.Second)
}

// Helper to create domain errors
func NewDomainError(domain, source, message string, err error) error {
	return &DomainError{
			Domain:  domain,
			Source:  source,
			Message: message,
			Err:     err,
	}
}