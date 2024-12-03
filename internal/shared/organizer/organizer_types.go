package organizer

import (
	"context"

	"github.com/lokeam/bravo-kilo/internal/shared/types"
)

// DomainOrganizer defines the interface for all domain organizers
type DomainOrganizer interface {
	// OrganizeForLibrary handles library page organization
	OrganizeForLibrary(ctx context.Context, data *types.LibraryPageData) (*types.LibraryPageData, error)

	// OrganizeForHome handles home page organization
	OrganizeForHome(ctx context.Context, data *types.HomePageData) (*types.HomePageData, error)

	// Returns metrics for Organizer
	GetMetrics() OrganizerMetrics
}


type OrganizerMetrics struct {
	OrganizationErrors int64
	ItemsOrganized     int64
}
