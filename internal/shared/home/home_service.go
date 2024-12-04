package home

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/lokeam/bravo-kilo/internal/books/repository"
	"github.com/lokeam/bravo-kilo/internal/shared/core"
	"github.com/lokeam/bravo-kilo/internal/shared/operations"
	"github.com/lokeam/bravo-kilo/internal/shared/organizer"
	"github.com/lokeam/bravo-kilo/internal/shared/redis"
	"github.com/lokeam/bravo-kilo/internal/shared/services"
	"github.com/lokeam/bravo-kilo/internal/shared/types"
)

type HomeService struct {
	operations *operations.Manager
	operationFactory *operations.OperationFactory
	organizerFactory *organizer.OrganizerFactory
	logger *slog.Logger
}

// RESPONSIBILITIES:
/*
	- Service focus on ORCHESTRATION
	- No direct cache/db access
	- Use operations for all external calls
	- Make sure error handling chain is CLEAN
	- Type safe response building
*/

func NewHomeService(
	operationsManager *operations.Manager,
	operationFactory *operations.OperationFactory,
	organizerFactory *organizer.OrganizerFactory,
	validationService *services.ValidationService,
	logger *slog.Logger,
) (*HomeService, error) {
	if operationFactory == nil {
		return nil, fmt.Errorf("operation factory cannot be nil")
	}
	if organizerFactory == nil {
		return nil, fmt.Errorf("organizor factory cannot be nil")
	}
	if operationsManager == nil {
		return nil, fmt.Errorf("operations manager cannot be nil")
	}
	if validationService == nil {
		return nil, fmt.Errorf("validation service cannot be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	return &HomeService{
		operations: operationsManager,
		operationFactory: operationFactory,
		organizerFactory: organizerFactory,
		logger:              logger,
	}, nil
}

func (hs *HomeService) GetHomeData(
	ctx context.Context,
	userID int,
	params *types.PageQueryParams,
) (*types.HomeResponse, error) {
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
	data, err := hs.operations.Cache.Get(ctx, userID, params)
	if err != nil && !errors.Is(err, redis.ErrNotFound) {
		// Only return error if not cache miss
		return nil, fmt.Errorf("cache operation failed: %w", err)
	}

	// Type assertion to ensure data is of correct type
	organizedData, ok := data.(*types.HomePageData)
	if !ok && data != nil {
		return nil, fmt.Errorf("unexpected data type during home service type assertion: %T", data)
	}

	// 2. If cache miss OR data is nil, get fresh data
	if errors.Is(err, redis.ErrNotFound) || data == nil {
		// 3. Process domain data
		pageOperation, err := hs.operationFactory.CreateOperation(core.BookDomainType, core.HomePage)
		if err != nil {
			return nil, fmt.Errorf("failed to create operation: %w", err)
		}

		// 4. Use page operation to get domain data from db
		data, err := pageOperation.GetData(ctx, userID, params)
		if err != nil {
			return nil, fmt.Errorf("failed to get data: %w", err)
		}

		// 5. Type assertion to ensure data is of correct type
		homePageData, ok := data.(*types.HomePageData)
		if !ok {
			return nil, fmt.Errorf("unexpected data type during library service type assertion: %T", data)
		}

		// 6. Use organizer factory to organize data based upon domain
		organizer, err := hs.organizerFactory.GetOrganizer(params.Domain, params)
		if err != nil {
			return nil, fmt.Errorf("failed to organize data based upon domain: %w", err)
		}

		// 7. Organize data into correct shape for home or library page
		organizedData, err = organizer.OrganizeForHome(ctx, homePageData)
		if err != nil {
			return nil, fmt.Errorf("failed to process data: %w", err)
		}

		// 8. Cache processed data
		if err := hs.operations.Cache.Set(ctx, userID, params, organizedData); err != nil {
			// Log but don't fail if cache update fails
			hs.logger.Error("failed to cache library data",
					"userId", userID,
					"error", err,
			)
		}

	}

	// 5. Build response
	return hs.buildResponse(
		ctx.Value(core.RequestIDKey).(string),
		organizedData,
		"database",
	), nil
}

// Helper - Construct response
// buildResponse constructs a validated Home pageResponse ensuring all required fields meet frontend contract
func (hs *HomeService) buildResponse(
	requestID string,
	data *types.HomePageData,
	source string,
) *types.HomeResponse {
	if data == nil {
			hs.logger.Error("attempt to build response with nil data",
					"component", "home_service",
					"requestID", requestID,
					"source", source)

			// Initialize with empty but valid data structures
			data = types.NewHomePageData(hs.logger)
	}

	// Validate and populate book data
	for i := range data.Books {
			if data.Books[i].Title == "" {
					data.Books[i].Title = "Unknown Title"
			}
			if data.Books[i].Description.Ops == nil {
					data.Books[i].Description = repository.RichText{
							Ops: []repository.DeltaOp{{Insert: "No description available"}},
					}
			}
			if data.Books[i].PublishDate == "" {
					data.Books[i].PublishDate = "Unknown"
			}
			if data.Books[i].ImageLink == "" {
					data.Books[i].ImageLink = "/default-book-cover.jpg"
			}
			if len(data.Books[i].Formats) == 0 {
					data.Books[i].Formats = []string{"physical"}
			}
	}

	// Validate statistics
	if err := data.Validate(); err != nil {
			hs.logger.Error("validation failed",
					"component", "home_service",
					"requestID", requestID,
					"error", err)
	}

	return &types.HomeResponse{
			RequestID: requestID,
			Data:      data,
			Source:    source,
	}
}

// validateBookCollection ensures all books in a map collection meet frontend requirements
func (ls *HomeService) validateBookCollection(books map[string][]repository.Book) {
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
func (ls *HomeService) validateFormatBooks(formatData *types.FormatData) {
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
func (ls *HomeService) normalizeSingleBook(book *repository.Book) {
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
