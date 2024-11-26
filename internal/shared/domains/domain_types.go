package domains

import (
	"context"

	"github.com/lokeam/bravo-kilo/internal/shared/core"
)

// DomainHandler defines the interface for domain-specific operations
type DomainHandler interface {
	GetType() core.DomainType
	GetLibraryItems(ctx context.Context, userID int) ([]core.LibraryItem, error)
    GetMetadata() (core.DomainMetadata, error)
}

// DomainRegistry manages available domains
type DomainRegistry interface {
    RegisterDomain(handler DomainHandler) error
    GetHandler(domainType core.DomainType) (DomainHandler, error)
    GetEnabledDomains() []core.DomainType
}
