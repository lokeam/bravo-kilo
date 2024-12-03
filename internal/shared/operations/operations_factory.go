package operations

import (
	"fmt"
	"log/slog"

	"github.com/lokeam/bravo-kilo/internal/shared/core"
)

type OperationFactory struct {
	bookHandlers BookOperationHandler
	// TODO: Add other domain repositories here
	logger *slog.Logger
}

func NewOperationFactory(
	bookHandlers BookOperationHandler,
	logger *slog.Logger,
) *OperationFactory {
	return &OperationFactory{
		bookHandlers: bookHandlers,
		logger:   logger,
	}
}

func (of *OperationFactory) CreateOperation(
	domain core.DomainType,
	pageType core.PageType,
) (DomainOperator, error) {
	of.logger.Debug("OPERATIONS_FACTORY: Creating operation",
	"component", "operations_factory",
		"function", "CreateOperation",
		"domain", domain,
		"pageType", pageType,
	)

	switch domain {
	case core.BookDomainType:
		return of.createBookOperation(pageType)
	default:
		err := fmt.Errorf("unsupported domain: %s", domain)
		of.logger.Error("OPERATIONS_FACTORY: Failed to create operation",
			"component", "operations_factory",
			"function", "CreateOperation",
			"error", err,
			"domain", domain,
		)

		return nil, fmt.Errorf("unsupported domain: %s", domain)
	}
}

func (of *OperationFactory) createBookOperation(pageType core.PageType) (DomainOperator, error) {
	of.logger.Debug("OPERATIONS_FACTORY: Creating book operation",
		"component", "operations_factory",
		"function", "createBookOperation",
		"pageType", pageType,
	)

	switch pageType {
	case core.LibraryPage:
		return NewLibraryOperation(of.bookHandlers, of.logger), nil
	case core.HomePage:
		return NewHomeOperation(of.bookHandlers, of.logger), nil
	default:
		err := fmt.Errorf("unsupported page type for books: %s", pageType)
		of.logger.Error("OPERATIONS_FACTORY: Failed to create book operation",
			"component", "operations_factory",
			"function", "createBookOperation",
			"error", err,
			"pageType", pageType,
		)

		return nil, fmt.Errorf("unsupported page type for books: %s", pageType)
	}
}