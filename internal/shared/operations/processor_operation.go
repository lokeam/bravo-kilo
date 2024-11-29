package operations

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/lokeam/bravo-kilo/internal/shared/organizer"
	"github.com/lokeam/bravo-kilo/internal/shared/processor/bookprocessor"
	"github.com/lokeam/bravo-kilo/internal/shared/types"
)

type ProcessorOperation struct {
    executor  types.OperationExecutor[*types.LibraryPageData]
    processor *bookprocessor.BookProcessor
    organizer *organizer.BookOrganizer
    logger    *slog.Logger
}

func NewProcessorOperation(
    processor *bookprocessor.BookProcessor,
    organizer *organizer.BookOrganizer,
    timeout time.Duration,
    logger *slog.Logger,
) (*ProcessorOperation, error) {
    if processor == nil {
        return nil, fmt.Errorf("processor cannot be nil")
    }
    if organizer == nil {
        return nil, fmt.Errorf("organizer cannot be nil")
    }
    if logger == nil {
        return nil, fmt.Errorf("logger cannot be nil")
    }

    return &ProcessorOperation{
        executor: NewOperationExecutor[*types.LibraryPageData](
            "processor",
            timeout,
            logger,
        ),
        processor: processor,
        organizer: organizer,
        logger:    logger,
    }, nil
}

func (p *ProcessorOperation) Process(ctx context.Context, data *types.LibraryPageData) (*types.LibraryPageData, error) {
    return p.executor.Execute(ctx, func(ctx context.Context) (*types.LibraryPageData, error) {
        if data == nil {
            return nil, fmt.Errorf("input data cannot be nil")
        }

        // Get first book details as map
        var firstBookDetails map[string]interface{}
        if len(data.Books) > 0 {
            firstBookDetails = logBookDetails(data.Books[0])
        }

        p.logger.Debug("PROCESSOR: Starting organization",
            "component", "processor_operation",
            "function", "Process",
            "inputBooksCount", len(data.Books),
            "firstBookDetails", firstBookDetails,
        )

        // Organize the processed data
        organizedData, err := p.organizer.OrganizeForLibrary(ctx, data)
        if err != nil {
            return nil, fmt.Errorf("organizing failed: %w", err)
        }

        // Get organized book details as map
        var firstOrganizedBookDetails map[string]interface{}
        if len(organizedData.Books) > 0 {
            firstOrganizedBookDetails = logBookDetails(organizedData.Books[0])
        }

        p.logger.Debug("PROCESSOR: Completed organizing",
            "component", "processor_operation",
            "function", "Process",
            "finalBooksCount", len(organizedData.Books),
            "firstFinalBookDetails", logBookDetails(firstOrganizedBookDetails),
        )

        return organizedData, nil
    })
}

