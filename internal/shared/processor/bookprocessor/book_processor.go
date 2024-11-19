package bookprocessor

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lokeam/bravo-kilo/internal/books/repository"
	"github.com/lokeam/bravo-kilo/internal/shared/processor"
	"github.com/lokeam/bravo-kilo/internal/shared/types"
)

type BookProcessor struct {
	logger   *slog.Logger
	metrics  *processor.ProcessorMetrics
}

func NewBookProcessor(logger *slog.Logger) (*BookProcessor, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	return &BookProcessor{
		logger:   logger,
		metrics:  &processor.ProcessorMetrics{},
	}, nil
}

func (bp *BookProcessor) GetDomainType() types.DomainType {
	return types.BookDomainType
}

func (bp *BookProcessor) ProcessLibraryItems(ctx context.Context, items []types.LibraryItem) (interface{}, error) {

	// 1. Guard clause: if no items, return nil and print warning
	if len(items) == 0 {
		return nil, fmt.Errorf("no items to process")
	}

	// Guard clause: Context timeout check
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}

	// 2. Init new variable to hold books
	books := make([]repository.Book, 0, len(items))

	// 3. Walk through slice of items
	for i, item := range items {
		// For each item:
		select {
			// A. If the context completes, return nill and print error if applicable
		case <- ctx.Done():
			return nil, ctx.Err()

			// B. Else
		default:
				// a. Record in processor metrics
			bp.metrics.ItemsProcessed.Add(1)

				// b. If item has book metadata, add to books slice
				if book, err := bp.processItem(item); err != nil {
					bp.metrics.ProcessingErrors.Add(1)
					bp.logger.Error("item processing failed",
							"itemIndex", i,
							"itemID", item.ID,
							"error", err,
							"metadata", fmt.Sprintf("%+v", item.Metadata),
					)
					continue
			} else if book != nil {
					books = append(books, *book)
			}

				// c. If item is of an object (map) type, run convertMapToBook
				if mapData, ok := item.Metadata.(map[string]interface{}); ok {
					book, err := bp.convertMapToBook(mapData)

					// i. If we encounter an error, record in processor metrics
					if err != nil {
						bp.metrics.ProcessingErrors.Add(1)
						bp.logger.Error("failed to convert map to book",
							"itemID", item.ID,
							"error", err,
						)
						continue
					}

					// ii. Else append book to books slice
					books = append(books, book)
					continue
				}

				// d. If item is of an unexpected type, record in metrics
			bp.metrics.InvalidItems.Add(1)
				bp.logger.Error("unexpected metadata type",
					"itemID", item.ID,
					"type", fmt.Sprintf("%T", item.Metadata),
				)
		}
	}

	// 4. If slice of items is empty, return nil and print warning
	if len(books) == 0 {
		return nil, fmt.Errorf("no valid books found in items")
	}
	// 5. Return LibraryData

	return types.LibraryPageData{
		// Note: Organizer package will fill in the rest of the data
		Books: books,
		Authors: make(map[string][]repository.Book),
		Genres:  make(map[string][]repository.Book),
		Formats: make(map[string][]repository.Book),
		Tags:    make(map[string][]repository.Book),
	}, nil
}

func (bp *BookProcessor) GetMetrics() *processor.ProcessorMetrics {
	metrics := &processor.ProcessorMetrics{}

	// Initialize the atomic values
	metrics.ProcessingErrors.Store(bp.metrics.ProcessingErrors.Load())
	metrics.ItemsProcessed.Store(bp.metrics.ItemsProcessed.Load())
	metrics.InvalidItems.Store(bp.metrics.InvalidItems.Load())

	return metrics
}


// Helper functions

// Process individual Library Item
func (bp *BookProcessor) processItem(item types.LibraryItem) (*repository.Book, error) {
	// Guard clause for type assertion
	if book, ok := item.Metadata.(repository.Book); ok {
		return &book, nil
	}

	// Try as pointer
	bookPtr, ok := item.Metadata.(*repository.Book)
	if ok {
			return bookPtr, nil
	}

	// Map conversion as fallback
	if mapData, ok := item.Metadata.(map[string]interface{}); ok {
		book, err := bp.convertMapToBook(mapData)
		if err != nil {
				return nil, fmt.Errorf("map conversion failed: %w", err)
		}
		return &book, nil
	}

	// Handle unknown type
	bp.metrics.InvalidItems.Add(1)
	return nil, fmt.Errorf("unexpected metadata type: %T", item.Metadata)
}


// Converts a map to Book type
func (bp *BookProcessor) convertMapToBook(data map[string]interface{}) (repository.Book, error) {
	// Init variable to hold book w/ default values
	var book repository.Book

	// Set required title field\
	title, ok := data["title"].(string)
	if !ok {
		return book, fmt.Errorf("invalid or missing title")
	}
	book.Title = title

	// If we have an array of authors, set this field

	// Note: initial recommended type assertion returns an interface{},
	//       we should be changing this on the front end, but adjust this value if needed
	if authors, ok := data["authors"].([]string); ok {
			book.Authors = append(book.Authors, authors...)
	}

	// Do same thing for genres, tags, etc
	if genres, ok := data["genres"].([]string); ok {
			book.Genres = append(book.Genres, genres...)
	}

	if tags, ok := data["tags"].([]string); ok {
			book.Tags = append(book.Tags, tags...)
	}

	// Set other fields: description, publisher, published date, ISBN10, ISBN13, pages, language

	// Return book and nil
	return book, nil
}
