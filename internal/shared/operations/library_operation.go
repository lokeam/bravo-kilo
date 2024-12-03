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
		logger.Debug("LIBRARY_OP: Creating new library operation",
			"component", "library_operation",
			"function", "NewLibraryOperation",
		)

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
	lo.logger.Debug("LIBRARY_OP: Starting GetData execution",
		"component", "library_operation",
		"function", "GetData",
		"userID", userID,
		"params", params,
	)

	return lo.Execute(ctx, func(ctx context.Context) (*types.LibraryPageData, error) {
		lo.logger.Debug("LIBRARY_OP: Fetching user books",
			"component", "library_operation",
			"function", "GetData.Execute",
			"userID", userID,
		)

		books, err := lo.bookHandlers.GetAllUserBooksDomain(ctx, userID)
		if err != nil {
			lo.logger.Error("LIBRARY_OP: Failed to get library items",
				"component", "library_operation",
				"function", "GetData.Execute",
				"error", err,
				"userID", userID,
			)
			return nil, fmt.Errorf("failed to get library items: %w", err)
		}

		lo.logger.Debug("DOMAIN_OP: Library data retrieved",
		"component", "library_operation",
		"function", "GetData",
		"dataType", fmt.Sprintf("%T", books),
			"hasData", books != nil,
		)

		pageData := types.NewLibraryPageData(lo.logger)
		lo.logger.Debug("LIBRARY_OP: Created page data",
			"component", "library_operation",
			"function", "GetData.Execute",
			"pageDataType", fmt.Sprintf("%T", pageData),
		)

		pageData.Books = books
		return pageData, nil
	})
}