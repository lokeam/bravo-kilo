package operations

import (
	"context"

	"github.com/lokeam/bravo-kilo/internal/shared/processor/bookprocessor"
	"github.com/lokeam/bravo-kilo/internal/shared/types"
)

type ProcessorOperation struct {
	*OperationExecutor[*types.LibraryPageData]
	processor *bookprocessor.BookProcessor
}

// Transform data for presentation on frontend
// Maintains own timeout and error handling via executor

func (p *ProcessorOperation) Process(ctx context.Context, data *types.LibraryPageData) (*types.LibraryPageData, error) {
	return p.Execute(ctx, func(ctx context.Context) (*types.LibraryPageData, error) {
		// Existing data transformation logic from library_page_handler.go
		return p.processor.Process(ctx, data)
	})
}