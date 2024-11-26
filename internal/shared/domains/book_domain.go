package domains

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/lokeam/bravo-kilo/internal/books/handlers"
	"github.com/lokeam/bravo-kilo/internal/shared/core"
)


type BookDomainHandler struct {
    bookHandlers *handlers.BookHandlers
    logger       *slog.Logger
}

type BookDomainError struct {
    Source string
    Err    error
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

// GetType implements core.DomainHandler
func (h *BookDomainHandler) GetType() core.DomainType {
    return core.BookDomainType
}

// GetLibraryItems implements core.DomainHandler
func (h *BookDomainHandler) GetLibraryItems(ctx context.Context, userID int) ([]core.LibraryItem, error) {
    h.logger.Debug("fetching library items",
        "userID", userID,
        "domain", "books",
    )

    start := time.Now()

    // Call refactored Getter from crud.go
    books, err := h.bookHandlers.GetAllUserBooksDomain(ctx, userID)
    if err != nil {
        h.logger.Error("failed to get user books",
            "userID", userID,
            "error", err,
            "duration", time.Since(start),
        )

        return nil, &BookDomainError{
            Source: "GetLibraryItems",
            Err:    err,
        }
    }

    h.logger.Debug("successfully retrieved books",
    "userID", userID,
    "bookCount", len(books),
        "duration", time.Since(start),
    )

    items := make([]core.LibraryItem, len(books))
    for i, book := range books {
        items[i] = core.LibraryItem{
            ID:          book.ID,
            Title:       book.Title,
            Type:        core.BookDomainType,
            DateAdded:   book.CreatedAt.Format(time.RFC3339),
            LastUpdated: book.LastUpdated.Format(time.RFC3339),
        }
    }

    h.logger.Debug("completed GetLibraryItems",
        "userID", userID,
        "itemCount", len(items),
        "totalDuration", time.Since(start),
    )

    return items, nil
}

func (h *BookDomainHandler) GetMetadata() (core.DomainMetadata, error){
    return core.DomainMetadata{
        DomainType: core.BookDomainType,  // Use correct field name
        Label:      "Books",               // Use correct field name
    }, nil
}

func (e *BookDomainError) Error() string {
    return fmt.Sprintf("book domain error in %s: %v", e.Source, e.Err)
}