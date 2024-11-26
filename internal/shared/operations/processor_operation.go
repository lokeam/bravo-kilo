package operations

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/lokeam/bravo-kilo/internal/shared/core"
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

        libraryItems := make([]core.LibraryItem, len(data.Books))
        for i, book := range data.Books {
            libraryItems[i] = core.LibraryItem{
                ID:          book.ID,
                Title:      book.Title,
                Type:       core.BookDomainType,
            }
        }

        // Process the items
        processedData, err := p.processor.ProcessLibraryItems(ctx, libraryItems)
        if err != nil {
            return nil, fmt.Errorf("processing failed: %w", err)
        }

        // Organize the processed data
        organizedData, err := p.organizer.OrganizeForLibrary(ctx, processedData)
        if err != nil {
            return nil, fmt.Errorf("organizing failed: %w", err)
        }

        return organizedData, nil
    })
}