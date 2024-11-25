package validator

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"time"

	"github.com/lokeam/bravo-kilo/internal/shared/operations"
	"github.com/lokeam/bravo-kilo/internal/shared/types"
)

type QueryValidator struct {
	*BaseValidator
	executor types.OperationExecutor[*types.LibraryQueryParams]
}

func NewQueryValidator(logger *slog.Logger) (*QueryValidator, error) {
	baseValidator, err := NewBaseValidator(logger, "query")
	if err != nil {
		return nil, err
	}

	return &QueryValidator{
		BaseValidator: baseValidator,
		executor:      operations.NewOperationExecutor[*types.LibraryQueryParams](
			"query_validation",
			100 * time.Millisecond,
			logger,
		),
	}, nil
}

// Handle Validation of Parsing and Validation
func (qv *QueryValidator) ParseAndValidate(
	ctx context.Context,
	query url.Values,
	rules QueryValidationRules,
) (*types.LibraryQueryParams, error) {
		// Validate raw query params
		if errs := qv.ValidateQueryParams(ctx, query, rules); len(errs) > 0 {
			return nil, errs[0]
		}

		// Parse validated parameters
		validatedParams := &types.LibraryQueryParams{}
		if page := query.Get("page"); page != "" {
			val, err := strconv.Atoi(page)
			if err != nil {
				return nil, fmt.Errorf("invalid page parameter: %w", err)
			}
			validatedParams.Page = val
		}

		return validatedParams, nil
}