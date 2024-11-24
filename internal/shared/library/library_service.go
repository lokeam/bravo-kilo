package library

import (
	"context"
	"log/slog"

	"github.com/lokeam/bravo-kilo/internal/shared/operations"
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

// 1. Primary business logic flow
func (ls *LibraryService) GetLibraryData(ctx context.Context, userID int, params *types.LibraryQueryParams) (*types.LibraryResponse, error) {
	/*
	Responsibilities:
		1. Try cache first
		2. If cache miss, get domain data
		3. Process domain data
		4. Cache results async
		5. Return response


		Flow:
			- Generate cache key
			- Try cache operation
			- On cache miss, run domain operation
			- Run processor operation (transform data into library page response format)
			- Cache update async
			- Return formatted response
	*/

	// Try cache
	data, err := ls.operations.Cache.Get(ctx, userID, params)
	if err != nil && !errors.Is(err, cache.ErrNotFound){
		return nil, fmt.Errorf("cache operation failed: %w", err)
	}

	// Get domain data
	data, err = ls.operations.Domain.GetDomainData(ctx, userID, params)
	if err != nil {
		return nil, err
	}

	// Process data into correct format
	data, err = ls.operations.Processor.Process(ctx, data)
	if err != nil {
		return nil, err
	}

	// Cache results
	go func(data *types.LibraryPageData) {
		// Todo: instead of magic number for context timeout, use config
		ctx, cancel := context.WithTimeout(context.Backgroud(), 5 * time.Second)
		defer cancel()
		if err := ls.operations.Cache.Set(ctx, userID, params, data); err != nil {
			ls.logger.Error("async redis cache update operation failed", "error", err)
		}
	}(data.Clone())

	// Build response
	return ls.buildResponse(ctx.Value(RequestIDKey).(string), data, "database"), nil
}

// Helpers

// Construct response
func (s *LibraryService) buildResponse(requestID string, data *types.LibraryPageData, source string) *types.LibraryResponse {
	/*
	Responsibilities:
		- Construct standard response format
		- Add metadata (requestID, source)
		- Enforce type safety
	*/
}


// Error handling
func (s *LibraryService) handleOperationError(ctx, context.Context, operation string, err error) {
	/*
	Responsibilities:
		- Log operation errors
		- Wrap errors with context
		- Preserve error types
	*/
}