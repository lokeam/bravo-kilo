package bookprocessor

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lokeam/bravo-kilo/internal/books/repository"
	"github.com/lokeam/bravo-kilo/internal/shared/core"
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

func (bp *BookProcessor) GetDomainType() core.DomainType {
	return core.BookDomainType
}

func (bp *BookProcessor) ProcessLibraryItems(ctx context.Context, items []core.LibraryItem) (*types.LibraryPageData, error) {
	bp.logger.Debug("PROCESSOR: Starting ProcessLibraryItems",
		"component", "book_processor",
		"function", "ProcessLibraryItems",
		"itemsCount", len(items),
			"firstItemDetails", logFirstItem(items),
		)

	if items == nil {
		return nil, fmt.Errorf("items slice cannot be nil")
	}

	// 1. Guard clause: if no items, return nil and print warning
	if len(items) == 0 {
		return nil, fmt.Errorf("no items to process")
	}

	// Guard clause: Context timeout check
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}

	// Create a buffered channel for results
	resultChan := make(chan repository.Book, len(items))
	errorChan := make(chan error, 1)

	// Process items in chunks of 50 (adjustable based on your needs)
	const chunkSize = 50
	for i := 0; i < len(items); i += chunkSize {
			end := i + chunkSize
			if end > len(items) {
					end = len(items)
			}

			chunk := items[i:end]
			bp.logger.Debug("PROCESSOR: Processing chunk",
				"component", "book_processor",
				"function", "ProcessLibraryItems",
				"chunkSize", len(chunk),
				"chunkStartIndex", i,
				"chunkEndIndex", end,
			)

			// Process chunk
			go func(chunk []core.LibraryItem) {
					for _, item := range chunk {
							select {
							case <-ctx.Done():
									errorChan <- ctx.Err()
									return
							default:
									bp.logger.Debug("PROCESSOR: Converting item to book",
										"component", "book_processor",
										"function", "ProcessLibraryItems",
										"itemID", item.ID,
										"itemTitle", item.Title,
									)
									bp.metrics.ItemsProcessed.Add(1)
									resultChan <- repository.Book{
											ID:    item.ID,
											Title: item.Title,
									}
							}
					}
			}(chunk)
	}

	// 2. Init new variable to hold books
	books := make([]repository.Book, 0, len(items))
	remaining := len(items)

	bp.logger.Debug("PROCESSOR: Starting to collect results",
		"component", "book_processor",
		"function", "ProcessLibraryItems",
		"expectedBooks", remaining,
	)

	// 3. Walk through slice of items
	for remaining > 0 {
		select {
		case <-ctx.Done():
				return nil, fmt.Errorf("processing timed out: %w", ctx.Err())
		case err := <-errorChan:
				return nil, fmt.Errorf("processing error: %w", err)
		case book := <-resultChan:
				books = append(books, book)
				remaining--

				if len(books) == 1 {
					bp.logger.Debug("PROCESSOR: First book processed",
							"component", "book_processor",
							"function", "ProcessLibraryItems",
							"firstBookID", book.ID,
							"firstBookTitle", book.Title,
					)
				}
		}
	}

	// 4. If slice of items is empty, return nil and print warning
	if len(books) == 0 {
		return nil, fmt.Errorf("no valid books found in items")
	}

	// 5. Return LibraryData
	result := &types.LibraryPageData{
    Books: books,
    BooksByAuthors: types.AuthorData{
        AllAuthors: make([]string, 0),
        ByAuthor:   make(map[string][]repository.Book),
    },
    BooksByGenres: types.GenreData{
        AllGenres: make([]string, 0),
        ByGenre:   make(map[string][]repository.Book),
    },
    BooksByFormat: types.FormatData{
        AudioBook: make([]repository.Book, 0),
        EBook:     make([]repository.Book, 0),
        Physical:  make([]repository.Book, 0),
    },
    BooksByTags: types.TagData{
        AllTags: make([]string, 0),
        ByTag:   make(map[string][]repository.Book),
    },
	}

	bp.logger.Debug("PROCESSOR: Completed ProcessLibraryItems",
		"component", "book_processor",
		"function", "ProcessLibraryItems",
		"processedBooksCount", len(books),
		"firstProcessedBook", logFirstBook(books),
	)

	return result, nil
}

func (bp *BookProcessor) ProcessBooks(books []repository.Book) (types.LibraryPageData, error) {
	bp.logger.Debug("PROCESSOR: After organization",
        "component", "book_processor",
        "booksCount", len(books),
        "firstBookTitle", books[0].Title,
        "firstBookAuthors", books[0].Authors)

	// Handle nil input
	if books == nil {
			return types.LibraryPageData{
					Books: []repository.Book{},
					BooksByAuthors: types.AuthorData{
							AllAuthors: []string{},
							ByAuthor:   make(map[string][]repository.Book),
					},
					BooksByGenres: types.GenreData{
							AllGenres: []string{},
							ByGenre:   make(map[string][]repository.Book),
					},
					BooksByFormat: types.FormatData{
							AudioBook: []repository.Book{},
							EBook:     []repository.Book{},
							Physical:  []repository.Book{},
					},
					BooksByTags: types.TagData{
							AllTags: []string{},
							ByTag:   make(map[string][]repository.Book),
					},
			}, fmt.Errorf("books slice cannot be nil")
	}

	// Handle empty input
	if len(books) == 0 {
			return types.LibraryPageData{
					Books: []repository.Book{},
					BooksByAuthors: types.AuthorData{
							AllAuthors: []string{},
							ByAuthor:   make(map[string][]repository.Book),
					},
					BooksByGenres: types.GenreData{
							AllGenres: []string{},
							ByGenre:   make(map[string][]repository.Book),
					},
					BooksByFormat: types.FormatData{
							AudioBook: []repository.Book{},
							EBook:     []repository.Book{},
							Physical:  []repository.Book{},
					},
					BooksByTags: types.TagData{
							AllTags: []string{},
							ByTag:   make(map[string][]repository.Book),
					},
			}, nil
	}

	// Process valid books
	result := types.LibraryPageData{
			Books: books,
			BooksByAuthors: types.AuthorData{
					AllAuthors: []string{},
					ByAuthor:   make(map[string][]repository.Book),
			},
			BooksByGenres: types.GenreData{
					AllGenres: []string{},
					ByGenre:   make(map[string][]repository.Book),
			},
			BooksByFormat: types.FormatData{
					AudioBook: []repository.Book{},
					EBook:     []repository.Book{},
					Physical:  []repository.Book{},
			},
			BooksByTags: types.TagData{
					AllTags: []string{},
					ByTag:   make(map[string][]repository.Book),
			},
	}

	// Process the books...
	return result, nil
}

func (bp *BookProcessor) GetMetrics() *processor.ProcessorMetrics {
	metrics := &processor.ProcessorMetrics{}

	// Initialize the atomic values
	metrics.ProcessingErrors.Store(bp.metrics.ProcessingErrors.Load())
	metrics.ItemsProcessed.Store(bp.metrics.ItemsProcessed.Load())
	metrics.InvalidItems.Store(bp.metrics.InvalidItems.Load())

	return metrics
}

// Helper function to safely log first item details
func logFirstItem(items []core.LibraryItem) map[string]interface{} {
	if len(items) == 0 {
			return nil
	}
	return map[string]interface{}{
			"id":    items[0].ID,
			"title": items[0].Title,
			"type":  items[0].Type,
	}
}

// Helper function to safely log first book details
func logFirstBook(books []repository.Book) map[string]interface{} {
	if len(books) == 0 {
			return nil
	}
	return map[string]interface{}{
			"id":    books[0].ID,
			"title": books[0].Title,
	}
}