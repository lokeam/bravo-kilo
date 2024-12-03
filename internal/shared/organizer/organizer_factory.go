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
	logger.Debug("ORGANIZER_FACTORY: Creating new organizer factory",
	"component", "organizer_factory",
		"function", "NewOrganizerFactory",
		"hasBookOrganizer", bookOrganizer != nil,
	)

	if bookOrganizer == nil {
		logger.Error("ORGANIZER_FACTORY: Book organizer is nil",
			"component", "organizer_factory",
			"function", "NewOrganizerFactory",
		)

     return nil, fmt.Errorf("book organizer cannot be nil")
  }
  if logger == nil {
    	return nil, fmt.Errorf("logger cannot be nil")
  }

  factory := &OrganizerFactory{
    bookOrganizer: bookOrganizer,
    logger:        logger,
  }

	logger.Debug("ORGANIZER_FACTORY: Successfully created organizer factory",
		"component", "organizer_factory",
		"function", "NewOrganizerFactory",
	)

	return factory, nil

}

// GetOrganizer returns the appropriate organizer based on domain type
func (of *OrganizerFactory) GetOrganizer(
    domain core.DomainType,
		params *types.PageQueryParams,
) (DomainOrganizer, error) {
	of.logger.Debug("ORGANIZER_FACTORY: Starting organizer request",
		"component", "organizer_factory",
		"function", "GetOrganizer",
		"domain", domain,
		"hasParams", params != nil,
	)

	if params == nil {
		of.logger.Error("ORGANIZER_FACTORY: Params are nil",
			"component", "organizer_factory",
			"function", "GetOrganizer",
			"domain", domain,
		)
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
    of.logger.Debug("ORGANIZER_FACTORY: Retrieved book organizer",
        "component", "organizer_factory",
        "function", "GetOrganizer",
        "organizerType", fmt.Sprintf("%T", of.bookOrganizer),
    )

		return of.bookOrganizer, nil
	default:
		return nil, fmt.Errorf("unsupported domain type: %s", domain)
	}
}