package services

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/lokeam/bravo-kilo/internal/shared/core"
	"github.com/lokeam/bravo-kilo/internal/shared/operations"
	"github.com/lokeam/bravo-kilo/internal/shared/types"
	"github.com/lokeam/bravo-kilo/internal/shared/validator"
)

type ValidationService struct {
	queryValidator         *validator.QueryValidator
	baseValidator          *validator.BaseValidator
	logger                 *slog.Logger
	executor               *operations.OperationExecutor[*types.LibraryQueryParams]
}

func (vs *ValidationService) ValidateLibraryRequest(
	ctx context.Context,
	query url.Values,
) (*types.LibraryQueryParams, error) {
	return vs.executor.Execute(ctx, func(opCtx context.Context) (*types.LibraryQueryParams, error) {
		// 1. Grab domain from query parameters
		params := &types.LibraryQueryParams{
			Domain: core.DomainType(query.Get("domain")),
		}

		// Validate domain
		if params.Domain == "" {
			params.Domain = core.BookDomainType
		}

		// The library page does not have any business rules to validate, skip to structure validation

		// 2. Validate structure
		if err := vs.baseValidator.ValidateStruct(opCtx, params); err != nil {
			return nil, fmt.Errorf("structure validation failed: %w", err)
		}

		return params, nil
	})
}