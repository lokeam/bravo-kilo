package books

import (
	"context"
	"log/slog"
)

// BookDomain implements the DomainService interface
type BookDomain struct {
    bookService *service.BookService
    logger      *slog.Logger
}

func (d *BookDomain) GetType() pages.DomainType {
    return pages.BookDomain
}

func (d *BookDomain) GetLibraryItems(ctx context.Context, userID int) ([]pages.LibraryItem, error) {
    books, err := d.bookService.GetUserBooks(ctx, userID)
    if err != nil {
        return nil, err
    }

    // Convert domain-specific books to generic LibraryItems
    items := make([]pages.LibraryItem, len(books))
    for i, book := range books {
        items[i] = pages.LibraryItem{
            ID:         book.ID,
            DomainType: string(pages.BookDomain),
            Title:      book.Title,
            Attributes: book.ToMap(),
        }
    }
    return items, nil
}