package operations

import (
	"context"

	"github.com/lokeam/bravo-kilo/internal/shared/types"
)

type DomainOperation struct {
	*OperationExecutor[*types.LibraryPageData]
	handler types.DomainHandler
}

// Handle database access
// Maintains own timeout and error handling via executor

func (d *DomainOperation) GetData(
	ctx context.Context,
	userID int,
	params *types.LibraryQueryParams,
	) (*types.LibraryPageData, error) {
	return d.Execute(ctx, func(ctx context.Context) (*types.LibraryPageData, error) {
		// Existing domain logic from library_page_handler.go
		return d.handler.GetLibraryData(ctx, userID, params)
	})
}
