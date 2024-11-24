package types

import "context"

// Domain type constants
const (
	BookDomainType  DomainType = "books"
	GameDomainType  DomainType = "games"
	MovieDomainType DomainType = "movies"
)

// DomainType represents the type of content domain
type DomainType string

// DomainHandler defines the interface for domain-specific operations
type DomainHandler interface {
	GetType() DomainType
	GetLibraryItems(ctx context.Context, userID int) ([]LibraryItem, error)
    GetMetadata() (DomainMetadata, error)
}

// LibraryItem represents a generic item in the library
type LibraryItem struct {
    ID          int         `json:"id"`
    Title       string      `json:"title"`
    Type        DomainType  `json:"type"`
    DateAdded   string      `json:"dateAdded"`
    LastUpdated string      `json:"lastUpdated"`
    Metadata    interface{} `json:"metadata"`
}

// DomainMetadata represents domain-specific metadata
type DomainMetadata struct {
    DomainType DomainType  // Field is DomainType, not Type
    Label      string      // Field is Label, not Name
}

// DomainRegistry manages available domains
type DomainRegistry interface {
    RegisterDomain(handler DomainHandler) error
    GetHandler(domainType DomainType) (DomainHandler, error)
    GetEnabledDomains() []DomainType
}
