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
	switch domain {
	case core.BookDomainType:
		return of.createBookOperation(pageType)
	default:
		return nil, fmt.Errorf("unsupported domain: %s", domain)
	}
}

func (of *OperationFactory) createBookOperation(pageType core.PageType) (DomainOperator, error) {
	switch pageType {
	case core.LibraryPage:
		return NewLibraryOperation(of.bookHandlers, of.logger), nil
	case core.HomePage:
		return NewHomeOperation(of.bookHandlers, of.logger), nil
	default:
		return nil, fmt.Errorf("unsupported page type for books: %s", pageType)
	}
}