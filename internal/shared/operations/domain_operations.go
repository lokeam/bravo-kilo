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
		  // Call book repository to get books
			books, err := d.handler.GetLibraryItems(ctx, userID)
			if err != nil {
				return nil, fmt.Errorf("failed to get library items: %w", err)
			}

			d.logger.Debug("DOMAIN_OP: Data before sending to organizer",
			"component", "domain_operation",
			"function", "GetData",
			"dataType", fmt.Sprintf("%T", books),
			"hasData", books != nil,
			// If it's a slice of books, log the first book's details
			"firstBookDetails", logBookDetails(books), // Create this helper function
		)

			// Initialize page data and assign books directly
			pageData := types.NewLibraryPageData(d.logger)
			pageData.Books = books

			return pageData, nil
	})
}

func logBookDetails(data interface{}) map[string]interface{} {
	if books, ok := data.([]repository.Book); ok && len(books) > 0 {
			book := books[0]
			return map[string]interface{}{
					"id":           book.ID,
					"title":        book.Title,
					"authorCount":  len(book.Authors),
					"authors":      book.Authors,
					"genreCount":   len(book.Genres),
					"genres":       book.Genres,
					"formatCount":  len(book.Formats),
					"formats":      book.Formats,
					"tagCount":     len(book.Tags),
					"tags":         book.Tags,
			}
	}
	return nil
}