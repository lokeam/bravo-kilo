package operations

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/lokeam/bravo-kilo/internal/shared/types"
)

type HomeOperation struct {
	*OperationExecutor[*types.HomePageData]
	bookHandlers BookOperationHandler
	logger *slog.Logger
}

func NewHomeOperation(
	bookHandlers BookOperationHandler,
	logger *slog.Logger,
	) *HomeOperation {
	return &HomeOperation{
		OperationExecutor: NewOperationExecutor[*types.HomePageData](
			"BookHomeOperation",
			30 * time.Second,
			logger,
		),
		bookHandlers: bookHandlers,
		logger: logger,
	}
}

// Library Operation can call the GetData method
func (lo *HomeOperation) GetData(
	ctx context.Context,
	userID int,
	params *types.PageQueryParams,
) (any, error) {
	return lo.Execute(ctx, func(ctx context.Context) (*types.HomePageData, error) {
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

		pageData := types.NewHomePageData(lo.logger)
		pageData.Books = books

		// Validation happens here
		if err := pageData.Validate(); err != nil {
			return nil, fmt.Errorf("home page data validation failed: %w", err)
		}

		return pageData, nil
	})
}