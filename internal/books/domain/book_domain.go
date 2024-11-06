package domain

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/lokeam/bravo-kilo/internal/books/repository"
	"github.com/lokeam/bravo-kilo/internal/shared/redis"
)

type BookDomain struct {
    bookRepo    *repository.BookRepository
    authorRepo  *repository.AuthorRepository
    formatRepo  *repository.FormatRepository
    genreRepo   *repository.GenreRepository
    tagRepo     *repository.TagRepository
    redisClient *redis.RedisClient
    logger      *slog.Logger
}

// LibraryData represents all book data needed for the library page
type LibraryData struct {
    Books         []repository.Book                  `json:"books"`
    BooksByAuthor map[string][]repository.Book       `json:"booksByAuthor"`
    BooksByFormat map[string][]repository.Book       `json:"booksByFormat"`
    BooksByGenre  map[string][]repository.Book       `json:"booksByGenre"`
    BooksByTag    map[string][]repository.Book       `json:"booksByTag"`
}

func (d *BookDomain) GetLibraryData(ctx context.Context, userID int) (*LibraryData, error) {
    var (
        wg sync.WaitGroup
        mu sync.Mutex
        result LibraryData
        errs []error
    )

    // Concurrent fetches using existing repo methods
    wg.Add(5)

    // Fetch books
    go func() {
        defer wg.Done()
        books, err := d.bookRepo.GetAllBooksByUserID(userID)
        mu.Lock()
        if err != nil {
            errs = append(errs, fmt.Errorf("books: %w", err))
        } else {
            result.Books = books
        }
        mu.Unlock()
    }()

    // Fetch books by author
    go func() {
        defer wg.Done()
        booksByAuthor, err := d.authorRepo.GetAllBooksByAuthors(userID)
        mu.Lock()
        if err != nil {
            errs = append(errs, fmt.Errorf("authors: %w", err))
        } else {
            result.BooksByAuthor = booksByAuthor
        }
        mu.Unlock()
    }()

    // Similar goroutines for format, genre, and tags...

    wg.Wait()

    if len(errs) > 0 {
        return nil, fmt.Errorf("errors fetching library data: %v", errs)
    }

    return &result, nil
}