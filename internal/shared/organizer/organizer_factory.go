package organizer

import (
	"fmt"
	"log/slog"

	"github.com/lokeam/bravo-kilo/internal/shared/core"
	"github.com/lokeam/bravo-kilo/internal/shared/types"
)

// OrganizerFactory creates and manages domain-specific organizers
type OrganizerFactory struct {
    bookOrganizer *BookOrganizer
    logger        *slog.Logger
}

func NewOrganizerFactory(
    bookOrganizer *BookOrganizer,
    logger *slog.Logger,
) (*OrganizerFactory, error) {
    if bookOrganizer == nil {
        return nil, fmt.Errorf("book organizer cannot be nil")
    }
    if logger == nil {
        return nil, fmt.Errorf("logger cannot be nil")
    }

    return &OrganizerFactory{
        bookOrganizer: bookOrganizer,
        logger:        logger,
    }, nil
}

// GetOrganizer returns the appropriate organizer based on domain type
func (of *OrganizerFactory) GetOrganizer(
    domain core.DomainType,
		params *types.PageQueryParams,
) (DomainOrganizer, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}

	of.logger.Debug("ORGANIZER_FACTORY: Getting organizer",
		"component", "organizer_factory",
		"function", "GetOrganizer",
		"domain", domain,
		"params", params,
  )

	switch domain {
	case core.BookDomainType:
			return of.bookOrganizer, nil
	default:
			return nil, fmt.Errorf("unsupported domain type: %s", domain)
	}
}