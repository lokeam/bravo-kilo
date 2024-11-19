package organizer

import (
	"fmt"
	"log/slog"
	"sync/atomic"

	"github.com/lokeam/bravo-kilo/internal/books/repository"
	"github.com/lokeam/bravo-kilo/internal/shared/types"
)

type BookOrganizer struct {
	logger    *slog.Logger
	metrics  *OrganizerMetrics
}

func NewBookOrganizer(logger *slog.Logger) (*BookOrganizer, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	return &BookOrganizer{
		logger:    logger,
		metrics:   &OrganizerMetrics{},
	}, nil
}

func (bo *BookOrganizer) OrganizeForLibrary(items interface{}) (types.LibraryPageData, error) {
	if items == nil {
		atomic.AddInt64(&bo.metrics.OrganizationErrors, 1)
		return types.LibraryPageData{}, fmt.Errorf("items cannot be nil")
	}

	pageData, ok := items.(*types.LibraryPageData)
	if !ok {
		return types.LibraryPageData{}, fmt.Errorf("expected types.LibraryPageData, got %T", items)
	}

	books := pageData.Books

	if len(books) == 0 {
		return types.LibraryPageData{
			Books:    []repository.Book{},
			Authors:  make(map[string][]repository.Book),
			Genres:   make(map[string][]repository.Book),
			Formats:  make(map[string][]repository.Book),
			Tags:     make(map[string][]repository.Book),
		}, nil
	}

	atomic.AddInt64(&bo.metrics.ItemsOrganized, int64(len(books)))

	authors, err := bo.organizeByAuthors(books)
	if err != nil {
			atomic.AddInt64(&bo.metrics.OrganizationErrors, 1)
			return types.LibraryPageData{}, fmt.Errorf("author organization failed: %w", err)
	}

	genres, err := bo.organizeByGenres(books)
	if err != nil {
			atomic.AddInt64(&bo.metrics.OrganizationErrors, 1)
			return types.LibraryPageData{}, fmt.Errorf("genre organization failed: %w", err)
	}

	formats, err := bo.organizeByFormats(books)
	if err != nil {
			atomic.AddInt64(&bo.metrics.OrganizationErrors, 1)
			return types.LibraryPageData{}, fmt.Errorf("format organization failed: %w", err)
	}

	tags, err := bo.organizeByTags(books)
	if err != nil {
			atomic.AddInt64(&bo.metrics.OrganizationErrors, 1)
			return types.LibraryPageData{}, fmt.Errorf("tag organization failed: %w", err)
	}

	return types.LibraryPageData{
		Books:    books,
		Authors:  authors,
		Genres:   genres,
		Formats:  formats,
		Tags:     tags,
	}, nil
}


func (bo *BookOrganizer) GetMetrics() OrganizerMetrics {
	return OrganizerMetrics{
		OrganizationErrors: atomic.LoadInt64(&bo.metrics.OrganizationErrors),
		ItemsOrganized:     atomic.LoadInt64(&bo.metrics.ItemsOrganized),
	}
}

// Helper functions

func (bo *BookOrganizer) organizeByAuthors(books []repository.Book) (map[string][]repository.Book, error) {
	if books == nil {
			return nil, fmt.Errorf("books slice cannot be nil")
	}

	result := make(map[string][]repository.Book)

	for i, book := range books {
			if book.Authors == nil {
					bo.logger.Warn("book has nil Authors slice",
							"bookIndex", i,
							"bookTitle", book.Title,
					)
					continue
			}

			for _, author := range book.Authors {
					if author == "" {
							bo.logger.Warn("empty author name found",
									"bookIndex", i,
									"bookTitle", book.Title,
							)
							continue
					}
					result[author] = append(result[author], book)
			}
	}

	return result, nil
}

func (bo *BookOrganizer) organizeByGenres(books []repository.Book) (map[string][]repository.Book, error) {
	if books == nil {
			return nil, fmt.Errorf("books slice cannot be nil")
	}

	result := make(map[string][]repository.Book)

	for i, book := range books {
			if book.Genres == nil {
					bo.logger.Warn("book has nil Genres slice",
							"bookIndex", i,
							"bookTitle", book.Title,
					)
					continue
			}

			for _, genre := range book.Genres {
					if genre == "" {
							bo.logger.Warn("empty genre found",
									"bookIndex", i,
									"bookTitle", book.Title,
							)
							continue
					}
					result[genre] = append(result[genre], book)
			}
	}

	return result, nil
}

func (bo *BookOrganizer) organizeByFormats(books []repository.Book) (map[string][]repository.Book, error) {
	if books == nil {
			return nil, fmt.Errorf("books slice cannot be nil")
	}

	result := make(map[string][]repository.Book)

	for i, book := range books {
			if book.Formats == nil {
					bo.logger.Warn("book has nil Formats slice",
							"bookIndex", i,
							"bookTitle", book.Title,
					)
					continue
			}

			for _, format := range book.Formats {
					if format == "" {
							bo.logger.Warn("empty format found",
									"bookIndex", i,
									"bookTitle", book.Title,
							)
							continue
					}
					result[format] = append(result[format], book)
			}
	}

	return result, nil
}

func (bo *BookOrganizer) organizeByTags(books []repository.Book) (map[string][]repository.Book, error) {
	if books == nil {
			return nil, fmt.Errorf("books slice cannot be nil")
	}

	result := make(map[string][]repository.Book)

	for i, book := range books {
			if book.Tags == nil {
					bo.logger.Warn("book has nil Tags slice",
							"bookIndex", i,
							"bookTitle", book.Title,
					)
					continue
			}

			for _, tag := range book.Tags {
					if tag == "" {
							bo.logger.Warn("empty tag found",
									"bookIndex", i,
									"bookTitle", book.Title,
							)
							continue
					}
					result[tag] = append(result[tag], book)
			}
	}

	return result, nil
}