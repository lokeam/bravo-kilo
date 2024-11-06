package pages

import "context"

// DomainType represents supported content domains
type DomainType string

const (
    BookDomain  DomainType = "books"
    GameDomain  DomainType = "games"
    MovieDomain DomainType = "movies"
)

// DomainService defines what each domain must implement
type DomainService interface {
    GetType() DomainType
    GetLibraryItems(ctx context.Context, userID int) ([]LibraryItem, error)
    GetMetadata(ctx context.Context, userID int) (DomainMetadata, error)
}

// DomainRegistry manages available domains
type DomainRegistry interface {
    RegisterDomain(service DomainService) error
    GetService(domainType DomainType) (DomainService, error)
    GetEnabledDomains() []DomainType
}