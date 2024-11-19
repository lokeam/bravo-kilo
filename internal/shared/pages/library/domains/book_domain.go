package domains

import (
	"context"
	"log/slog"
	"time"

	"github.com/lokeam/bravo-kilo/internal/books/handlers"
	"github.com/lokeam/bravo-kilo/internal/shared/types"
)


type BookDomainHandler struct {
    bookHandlers *handlers.BookHandlers
    logger       *slog.Logger
}

func NewBookDomainHandler(
    bookHandlers *handlers.BookHandlers,
    logger *slog.Logger,
) *BookDomainHandler {
    if bookHandlers == nil {
        panic("bookHandlers cannot be nil")
    }
    if logger == nil {
        panic("logger cannot be nil")
    }
    return &BookDomainHandler{
        bookHandlers: bookHandlers,
        logger:       logger,
    }
}

// GetType implements types.DomainHandler
func (h *BookDomainHandler) GetType() types.DomainType {
    return types.BookDomainType
}

// GetLibraryItems implements types.DomainHandler
func (h *BookDomainHandler) GetLibraryItems(ctx context.Context, userID int) ([]types.LibraryItem, error) {
    h.logger.Debug("fetching library items",
        "userID", userID,
        "domain", "books",
    )

    // Call refactored Getter from crud.go
    books, err := h.bookHandlers.GetAllUserBooksDomain(ctx, userID)
    if err != nil {
        return nil, &BookDomainError{
            Source: "GetLibraryItems",
            Err:    err,
        }
    }

    items := make([]types.LibraryItem, len(books))
    for i, book := range books {
        items[i] = types.LibraryItem{
            ID:          book.ID,
            Title:       book.Title,
            Type:        types.BookDomainType,
            DateAdded:   book.CreatedAt.Format(time.RFC3339),
            LastUpdated: book.LastUpdated.Format(time.RFC3339),
            Metadata:    book,
        }
    }

    return items, nil
}
