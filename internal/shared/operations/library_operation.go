package operations

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/lokeam/bravo-kilo/internal/shared/types"
)

type LibraryOperation struct {
	*OperationExecutor[*types.LibraryPageData]
	bookHandlers BookOperationHandler
	logger *slog.Logger
}

func NewLibraryOperation(
	bookHandlers BookOperationHandler,
	logger *slog.Logger,
	) *LibraryOperation {
	return &LibraryOperation{
		OperationExecutor: NewOperationExecutor[*types.LibraryPageData](
			"BookLibraryOperation",
			30 * time.Second,
			logger,
		),
		bookHandlers: bookHandlers,
		logger: logger,
	}
}

// Library Operation can call the GetData method
func (lo *LibraryOperation) GetData(
	ctx context.Context,
	userID int,
	params *types.PageQueryParams,
) (any, error) {
	return lo.Execute(ctx, func(ctx context.Context) (*types.LibraryPageData, error) {
		books, err := lo.bookHandlers.GetAllUserBooksDomain(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get library items: %w", err)
		}

		lo.logger.Debug("DOMAIN_OP: Library data retrieved",
		"component", "library_operation",
		"function", "GetData",
		"dataType", fmt.Sprintf("%T", books),
			"hasData", books != nil,
		)

		pageData := types.NewLibraryPageData(lo.logger)
		pageData.Books = books
		return pageData, nil
	})
}