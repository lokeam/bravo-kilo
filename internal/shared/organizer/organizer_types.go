package organizer

import (
	"github.com/lokeam/bravo-kilo/internal/shared/types"
)

type DomainOrganizer interface {
	// Organizes items for Library Page
	OrganizeForLibrary(items interface{}) (types.LibraryPageData, error)

	// Organizes items for Home page
	OrganizeForStats(items interface{}) (interface{}, error)

	// Returns metrics for Organizer
	GetMetrics() OrganizerMetrics
}

type OrganizerMetrics struct {
	OrganizationErrors int64
	ItemsOrganized     int64
}
