package domains

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/lokeam/bravo-kilo/internal/books/handlers"
	"github.com/lokeam/bravo-kilo/internal/books/repository"
)

const (
	BookDomainType = "books"
)

type BookDomainHandler struct {
	bookHandlers    *handlers.BookHandlers
	logger          *slog.Logger
}

type BookDomainError struct {
	Source string
	Err    error
}

// LibraryData represents all book data needed for the library page
type LibraryData struct {
    Books         []repository.Book                  `json:"books"`
    BooksByAuthor map[string][]repository.Book       `json:"booksByAuthor"`
    BooksByFormat map[string][]repository.Book       `json:"booksByFormat"`
    BooksByGenre  map[string][]repository.Book       `json:"booksByGenre"`
    BooksByTag    map[string][]repository.Book       `json:"booksByTag"`
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

func (h *BookDomainHandler) GetType() string {
	return BookDomainType
}

func (h *BookDomainHandler) GetLibraryItems(ctx context.Context, userID int) (interface{}, error) {
	var (
			wg sync.WaitGroup
			mu sync.Mutex
			result LibraryData
			errs []error
	)

	h.logger.Info("fetching library items",
		"userID", userID,
		"domainType", h.GetType(),
	)

	wg.Add(5)

	// Fetch all books
	go func() {
			defer wg.Done()
			books, err := h.bookHandlers.HandleGetAllUserBooks(ctx, userID)
			mu.Lock()
			if err != nil {
					h.logger.Error("failed to fetch books", "error", err)
					errs = append(errs, fmt.Errorf("books: %w", err))
			} else {
					result.Books = books
			}
			mu.Unlock()
	}()

	// Fetch books by author
	go func() {
			defer wg.Done()
			authors, err := h.bookHandlers.HandleGetBooksByAuthors(ctx, userID)
			mu.Lock()
			if err != nil {
					h.logger.Error("failed to fetch books by authors", "error", err)
					errs = append(errs, fmt.Errorf("authors: %w", err))
			} else {
					result.BooksByAuthor = authors
			}
			mu.Unlock()
	}()

	// Fetch books by format
	go func() {
			defer wg.Done()
			formats, err := h.bookHandlers.HandleGetBooksByFormat(ctx, userID)
			mu.Lock()
			if err != nil {
					h.logger.Error("failed to fetch books by format", "error", err)
					errs = append(errs, fmt.Errorf("formats: %w", err))
			} else {
					result.BooksByFormat = formats
			}
			mu.Unlock()
	}()

	// Fetch books by genre
	go func() {
			defer wg.Done()
			genres, err := h.bookHandlers.HandleGetBooksByGenres(ctx, userID)
			mu.Lock()
			if err != nil {
					h.logger.Error("failed to fetch books by genre", "error", err)
					errs = append(errs, fmt.Errorf("genres: %w", err))
			} else {
					result.BooksByGenre = genres
			}
			mu.Unlock()
	}()

	// Fetch books by tag
	go func() {
			defer wg.Done()
			tags, err := h.bookHandlers.HandleGetBooksByTags(ctx, userID)
			mu.Lock()
			if err != nil {
					h.logger.Error("failed to fetch books by tags", "error", err)
					errs = append(errs, fmt.Errorf("tags: %w", err))
			} else {
					result.BooksByTag = tags
			}
			mu.Unlock()
	}()

	wg.Wait()

	if len(errs) > 0 {
			return nil, &BookDomainError{
				Source:   "GetLibraryItems",
				Err:      fmt.Errorf("multiple errors: %v", errs),
			}
	}

	h.logger.Info("library items fetched successfully",
		"userID", userID,
		"bookCount", len(result.Books),
		"authorCount", len(result.BooksByAuthor),
		"formatCount", len(result.BooksByFormat),
		"genreCount", len(result.BooksByGenre),
		"tagCount", len(result.BooksByTag),
	)

	return result, nil
}



func (e *BookDomainError) Error() string {
	return fmt.Sprintf("book domain error from %s: %v ", e.Source, e.Err)
}

