package processor

import (
	"context"
	"sync/atomic"

	"github.com/lokeam/bravo-kilo/internal/shared/core"
)

// Defines how each domain processes its library items
type DomainProcessor interface {
	// Converts domain items into page-specific data
	ProcessLibraryItems(ctx context.Context, items []core.LibraryItem) (interface{}, error)

	// Return type of domain that is being processed
	GetDomainType() core.DomainType

	GetMetrics() *ProcessorMetrics
}


// Tracks processing operations
type ProcessorMetrics struct {
	ProcessingErrors atomic.Int64
	ItemsProcessed   atomic.Int64
	InvalidItems     atomic.Int64
}
