package operations

import (
	"context"

	"github.com/lokeam/bravo-kilo/internal/books/repository"
	"github.com/lokeam/bravo-kilo/internal/shared/types"
)

type DomainOperator interface {
	GetData(ctx context.Context, userID int, params *types.PageQueryParams) (interface{}, error)
}

// Define what we need from BookHandlers
type BookOperationHandler interface {
    GetAllUserBooksDomain(ctx context.Context, userID int) ([]repository.Book, error)
}