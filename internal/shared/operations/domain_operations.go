package operations

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/lokeam/bravo-kilo/internal/books/repository"
	"github.com/lokeam/bravo-kilo/internal/shared/core"
	"github.com/lokeam/bravo-kilo/internal/shared/types"
)

// Local interface definition
type domainDataProvider interface {
	GetLibraryItems(ctx context.Context, userID int) ([]repository.Book, error)
}

type DomainOperation struct {
	*OperationExecutor[*types.LibraryPageData]
	handler domainDataProvider
	logger *slog.Logger
	domain core.DomainType
}

func NewDomainOperation(
	domain core.DomainType,
	handler domainDataProvider,
	logger *slog.Logger,
	) *DomainOperation {

		// default operation name and timeout
		const (
			defaultOperationName = "DomainOperation"
		defaultTimeout      = 30 * time.Second
	)

	return &DomainOperation{
		OperationExecutor: NewOperationExecutor[*types.LibraryPageData](
			defaultOperationName,
			defaultTimeout,
			logger,
		),
		domain: domain,
		handler: handler,
		logger: logger,
	}
}

func (d *DomainOperation) GetData(
	ctx context.Context,
	userID int,
	params *types.LibraryQueryParams,
) (*types.LibraryPageData, error) {
	return d.Execute(ctx, func(ctx context.Context) (*types.LibraryPageData, error) {
			books, err := d.handler.GetLibraryItems(ctx, userID)
			if err != nil {
					return nil, fmt.Errorf("failed to get library items: %w", err)
			}

			// Initialize page data and assign books directly
			pageData := types.NewLibraryPageData(d.logger)
			pageData.Books = books

			return pageData, nil
	})
}