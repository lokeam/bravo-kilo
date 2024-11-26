package library

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

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
	if err != nil && !errors.Is(err, redis.ErrNotFound){
		// Only return error if not cache miss
		return nil, fmt.Errorf("cache operation failed: %w", err)
	}

	// 2.If cache miss, get fresh data
	if errors.Is(err, redis.ErrNotFound) {
		// Get domain data
		data, err = ls.operations.Domain.GetData(ctx, userID, params)
		if err != nil {
			return nil, err
		}

		// 3.Process data into correct format
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
func (ls *LibraryService) buildResponse(
	requestID string,
	data *types.LibraryPageData,
	source string,
	) *types.LibraryResponse {
	/*
	Responsibilities:
		- Construct standard response format
		- Add metadata (requestID, source)
		- Enforce type safety
	*/

	// Validate inputs
	if data == nil {
		ls.logger.Error("attempt to build response with nil data")

		// Return empty response instead of nil in order to maintain contract
		return &types.LibraryResponse{
			RequestID: requestID,
			Source:    source,
		}
	}

	// Validate data structure
	if err := data.Validate(); err != nil {
		ls.logger.Error("library page response data validation failed",
		"requestID", requestID,
		"error", err,
		)
	}

	// Create response
	response := &types.LibraryResponse{
		RequestID: requestID,
		Data:      data,
		Source:    source,
	}

	ls.logger.Info("library page response created",
		"requestID", requestID,
		"source", source,
		"bookCount", len(data.Books),
	)

	return response
}
