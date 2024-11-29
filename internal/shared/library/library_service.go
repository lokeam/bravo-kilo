package library

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/lokeam/bravo-kilo/internal/books/repository"
	"github.com/lokeam/bravo-kilo/internal/shared/core"
	"github.com/lokeam/bravo-kilo/internal/shared/operations"
	"github.com/lokeam/bravo-kilo/internal/shared/redis"
	"github.com/lokeam/bravo-kilo/internal/shared/services"
	"github.com/lokeam/bravo-kilo/internal/shared/types"
)

type LibraryService struct {
	operations *operations.Manager
	logger     *slog.Logger
}

// RESPONSIBILITIES:
/*
	- Service focus on ORCHESTRATION
	- No direct cache/db access
	- Use operations for all external calls
	- Make sure error handling chain is CLEAN
	- Type safe response building
*/
func NewLibraryService(
	operationsManager *operations.Manager,
	validationService *services.ValidationService,
	logger *slog.Logger,
) (*LibraryService, error) {
	if operationsManager == nil {
		return nil, fmt.Errorf("operations manager cannot be nil")
	}
	if validationService == nil {
		panic("validation service is required")
	}
	if logger == nil {
		panic("logger is required")
	}
	return &LibraryService{
		operations: operationsManager,
		logger:              logger,
	}, nil
}



// 1. Primary business logic flow
func (ls *LibraryService) GetLibraryData(ctx context.Context, userID int, params *types.LibraryQueryParams) (*types.LibraryResponse, error) {
	/*
	Responsibilities:
		1. Try cache first
		2. If cache miss, get domain data
		3. Process domain data
		4. Cache results
		5. Return response


		Flow:
			- Generate cache key
			- Try cache operation
			- On cache miss, run domain operation
			- Run processor operation (transform data into library page response format)
			- Cache update
			- Return formatted response
	*/

	// 1. Try cache
	data, err := ls.operations.Cache.Get(ctx, userID, params)
	if err != nil && !errors.Is(err, redis.ErrNotFound) {
			// Only return error if not cache miss
			return nil, fmt.Errorf("cache operation failed: %w", err)
	}

	// 2. If cache miss OR data is nil, get fresh data
	if errors.Is(err, redis.ErrNotFound) || data == nil {
			// Get domain data
			data, err = ls.operations.Domain.GetData(ctx, userID, params)
			if err != nil {
					return nil, err
			}

			// 3. Organize data into correct shape
			data, err = ls.operations.Processor.Process(ctx, data)
			if err != nil {
					return nil, err
			}

			// 4. Cache processed data
			if err := ls.operations.Cache.Set(ctx, userID, params, data); err != nil {
					// Log but don't fail if cache update fails
					ls.logger.Error("failed to cache library data",
							"userId", userID,
							"error", err,
					)
			}
	}

// 5. Build response
	return ls.buildResponse(
    ctx.Value(core.RequestIDKey).(string),
    data,
    "database",
	), nil
}

// Helper - Construct response
// buildResponse constructs a validated LibraryResponse ensuring all required fields meet frontend contract
func (ls *LibraryService) buildResponse(
	requestID string,
	data *types.LibraryPageData,
	source string,
) *types.LibraryResponse {
	if data == nil {
			ls.logger.Error("attempt to build response with nil data",
					"component", "library_service",
					"requestID", requestID,
					"source", source)

			// Initialize with empty but valid data structures
			data = &types.LibraryPageData{
					Books:          make([]repository.Book, 0),
					BooksByAuthors: types.AuthorData{
							AllAuthors: make([]string, 0),
							ByAuthor:   make(map[string][]repository.Book),
					},
					BooksByGenres: types.GenreData{
							AllGenres: make([]string, 0),
							ByGenre:   make(map[string][]repository.Book),
					},
					BooksByTags: types.TagData{
							AllTags: make([]string, 0),
							ByTag:   make(map[string][]repository.Book),
					},
					BooksByFormat: types.FormatData{
							Physical:  make([]repository.Book, 0),
							EBook:     make([]repository.Book, 0),
							AudioBook: make([]repository.Book, 0),
					},
			}
	}

	// Ensure all books meet frontend contract requirements
	for i := range data.Books {
			// Handle description (required string)
			if data.Books[i].Description.Ops == nil {
					data.Books[i].Description = repository.RichText{
							Ops: []repository.DeltaOp{{Insert: "No description available"}},
					}
			}

			// Handle publishDate (required string)
			if data.Books[i].PublishDate == "" {
					data.Books[i].PublishDate = "Unknown"
			}

			// Handle imageLink (required string)
			if data.Books[i].ImageLink == "" {
					data.Books[i].ImageLink = "/default-book-cover.jpg"
			}

			// Handle formats (required array)
			if len(data.Books[i].Formats) == 0 {
					data.Books[i].Formats = []string{"physical"} // Default to physical if none specified
			}
	}

	// Apply same validations to books in categorized collections
	ls.validateBookCollection(data.BooksByAuthors.ByAuthor)
	ls.validateBookCollection(data.BooksByGenres.ByGenre)
	ls.validateBookCollection(data.BooksByTags.ByTag)
	ls.validateFormatBooks(&data.BooksByFormat)

	return &types.LibraryResponse{
			RequestID: requestID,
			Data:     data,
			Source:   source,
	}
}

// validateBookCollection ensures all books in a map collection meet frontend requirements
func (ls *LibraryService) validateBookCollection(books map[string][]repository.Book) {
	for category := range books {
			for i := range books[category] {
					if books[category][i].Description.Ops == nil {
							books[category][i].Description = repository.RichText{
									Ops: []repository.DeltaOp{{Insert: "No description available"}},
							}
					}
					if books[category][i].PublishDate == "" {
							books[category][i].PublishDate = "Unknown"
					}
					if books[category][i].ImageLink == "" {
							books[category][i].ImageLink = "/default-book-cover.jpg"
					}
					if len(books[category][i].Formats) == 0 {
							books[category][i].Formats = []string{"physical"}
					}
			}
	}
}

// validateFormatBooks ensures all books in format collections meet frontend requirements
func (ls *LibraryService) validateFormatBooks(formatData *types.FormatData) {
	for i := range formatData.Physical {
			ls.normalizeSingleBook(&formatData.Physical[i])
	}
	for i := range formatData.EBook {
			ls.normalizeSingleBook(&formatData.EBook[i])
	}
	for i := range formatData.AudioBook {
			ls.normalizeSingleBook(&formatData.AudioBook[i])
	}
}

// validateSingleBook ensures a single book meets frontend requirements
func (ls *LibraryService) normalizeSingleBook(book *repository.Book) {
	if book.Description.Ops == nil {
			book.Description = repository.RichText{
					Ops: []repository.DeltaOp{{Insert: "No description available"}},
			}
	}
	if book.PublishDate == "" {
			book.PublishDate = "Unknown"
	}
	if book.ImageLink == "" {
			book.ImageLink = "/default-book-cover.jpg"
	}
	if len(book.Formats) == 0 {
			book.Formats = []string{"physical"}
	}
}