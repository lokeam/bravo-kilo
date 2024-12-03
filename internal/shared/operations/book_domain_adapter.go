package operations

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lokeam/bravo-kilo/internal/books/repository"
)

// Must have interface for required repository methods
type bookRepository interface {
	GetAllBooksByUserID(userID int) ([]repository.Book, error)
}

type BookDomainAdapter struct {
	bookRepo   bookRepository
	logger     *slog.Logger
}

// Constructor
func NewBookDomainAdapter(
	bookRepo repository.BookRepository,
	logger *slog.Logger,
) *BookDomainAdapter {
	if bookRepo == nil {
		panic("bookRepo is nil")
	}
	if logger == nil {
		panic("logger is nil")
	}

	return &BookDomainAdapter{
		bookRepo: bookRepo,
		logger:   logger.With("component", "book_domain_adapter"),
	}
}


// Get Library Items implements domainDataProvider interface
func (a *BookDomainAdapter) GetAllUserBooksDomain(ctx context.Context, userID int) ([]repository.Book, error) {
	// Use existing BookRepo methods to get library items
	books, err := a.bookRepo.GetAllBooksByUserID(userID)  // Use correct method name
	if err != nil {
			a.logger.Error("failed to get user books",
					"userID", userID,
					"error", err,
			)
			return nil, fmt.Errorf("failed to get user books: %w", err)
	}
	return books, nil
}