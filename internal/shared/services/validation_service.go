package services

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"time"

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

const (
	defaultTimeout = 30 * time.Second
)

func NewValidationService(
	baseValidator *validator.BaseValidator,
	queryValidator *validator.QueryValidator,
	logger *slog.Logger,
) (*ValidationService, error) {
	// Validate required dependencies
	if baseValidator == nil {
		return nil, fmt.Errorf("base validator cannot be nil")
	}
	if queryValidator == nil {
		return nil, fmt.Errorf("query validator cannot be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	executor := operations.NewOperationExecutor[*types.LibraryQueryParams](
		"validation",
		defaultTimeout,
		logger.With("component", "validation_executor"),
	)

	return &ValidationService{
		baseValidator:   baseValidator,
		queryValidator:  queryValidator,
		logger:          logger.With("component", "validation_service"),
		executor:        executor,
	}, nil
}

func (vs *ValidationService) ValidateLibraryRequest(
	ctx context.Context,
	query url.Values,
) (*types.LibraryQueryParams, error) {
	return vs.executor.Execute(ctx, func(opCtx context.Context) (*types.LibraryQueryParams, error) {
        // 1. Extract domain
        domainStr := query.Get("domain")

        // 2. Set default if empty
        if domainStr == "" {
            domainStr = string(core.BookDomainType)
        }

        // 3. Validate domain is one of the allowed values
        domain := core.DomainType(domainStr)
        if domain != core.BookDomainType &&
           domain != core.MovieDomainType &&
           domain != core.GameDomainType {
            return nil, fmt.Errorf("invalid domain: %s", domainStr)
        }

        params := &types.LibraryQueryParams{
            Domain: domain,
        }

        if err := vs.baseValidator.ValidateStruct(opCtx, params); err != nil {
            return nil, fmt.Errorf("validation failed: %v", err)
        }

        return params, nil
	})
}