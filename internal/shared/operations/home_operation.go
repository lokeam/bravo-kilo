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
func (ho *HomeOperation) GetData(
	ctx context.Context,
	userID int,
	params *types.PageQueryParams,
) (any, error) {
	return ho.Execute(ctx, func(ctx context.Context) (*types.HomePageData, error) {
		books, err := ho.bookHandlers.GetAllUserBooksDomain(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get library items: %w", err)
		}

		ho.logger.Debug("DOMAIN_OP: Library data retrieved",
		"component", "library_operation",
		"function", "GetData",
		"dataType", fmt.Sprintf("%T", books),
			"hasData", books != nil,
		)

		pageData := types.NewHomePageData(ho.logger)
		pageData.Books = books

		ho.logger.Debug("DOMAIN_OP: Starting format count calculation",
				"component", "library_operation",
				"function", "GetData",
				"bookCount", len(books))

		// Validation happens here before any processing,
		if err := pageData.CalculateFormatCounts(); err != nil {
				return nil, fmt.Errorf("failed to calculate format counts: %w", err)
		}

		ho.logger.Debug("DOMAIN_OP: Completed format count calculation",
				"component", "library_operation",
				"function", "GetData",
				"formatCounts", fmt.Sprintf("%+v", pageData.BooksByFormat))

		if err := pageData.Validate(); err != nil {
				return nil, fmt.Errorf("home page data validation failed: %w", err)
		}

		return pageData, nil
	})
}