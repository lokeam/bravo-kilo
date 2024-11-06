package types

import (
	"context"
)

type DomainHandler interface {
	// Returns domain identifier (e.g. "book")
	GetType() string

	// Fetches all items for a user within this domain
	GetLibraryItems(ctx context.Context, userID int) (interface{}, error)
}

// standardized domain data
type DomainResponse struct {
	Type   string       `json:"type"`
	Items  interface{}  `json:"items"`
}

// Represents a generic item within any domain
type LibraryItem struct {
	ID           int                           `json:"id"`
	DomainType   string                        `json:"domainType"`
	Title        string                        `json:"title"`
	Attributes   map[string]interface{}        `json:"attributes"`
}