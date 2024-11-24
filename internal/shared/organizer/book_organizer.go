package organizer

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync/atomic"

	"github.com/lokeam/bravo-kilo/internal/books/repository"
	"github.com/lokeam/bravo-kilo/internal/shared/types"
)

const (
	FormatAudioBook = "audioBook"
	FormatEBook     = "eBook"
	FormatPhysical  = "physical"
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

// Sort these books into different views (by author, genre, etc.)
func (bo *BookOrganizer) OrganizeForLibrary(ctx context.Context, items *types.LibraryPageData) (*types.LibraryPageData, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}

	if err := ctx.Err(); err != nil {
		atomic.AddInt64(&bo.metrics.OrganizationErrors, 1)
		return nil, fmt.Errorf("context error before organization: %w", err)
	}

	if items == nil {
		atomic.AddInt64(&bo.metrics.OrganizationErrors, 1)
		return &types.LibraryPageData{}, fmt.Errorf("items cannot be nil")
	}

	books := items.Books

	if len(books) == 0 {
		return &types.LibraryPageData{
				Books:          books,
				BooksByAuthors: types.AuthorData{
						AllAuthors: make([]string, 0),
						ByAuthor:   make(map[string][]repository.Book),
				},
				BooksByGenres: types.GenreData{
						AllGenres: make([]string, 0),
						ByGenre:   make(map[string][]repository.Book),
				},
				BooksByFormat: types.FormatData{
						AudioBook: make([]repository.Book, 0),
						EBook:     make([]repository.Book, 0),
						Physical:  make([]repository.Book, 0),
				},
				BooksByTags: types.TagData{
						AllTags: make([]string, 0),
						ByTag:   make(map[string][]repository.Book),
				},
		}, nil
}

	atomic.AddInt64(&bo.metrics.ItemsOrganized, int64(len(books)))

	authors, err := bo.organizeByAuthors(ctx, books)
	if err != nil {
			atomic.AddInt64(&bo.metrics.OrganizationErrors, 1)
			return &types.LibraryPageData{}, fmt.Errorf("author organization failed: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled after author organization: %w", err)
	}

	genres, err := bo.organizeByGenres(ctx, books)
	if err != nil {
			atomic.AddInt64(&bo.metrics.OrganizationErrors, 1)
			return &types.LibraryPageData{}, fmt.Errorf("genre organization failed: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled after genre organization: %w", err)
	}

	formats, err := bo.organizeByFormats(ctx, books)
	if err != nil {
			atomic.AddInt64(&bo.metrics.OrganizationErrors, 1)
			return &types.LibraryPageData{}, fmt.Errorf("format organization failed: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled after format organization: %w", err)
	}

	tags, err := bo.organizeByTags(ctx,books)
	if err != nil {
			atomic.AddInt64(&bo.metrics.OrganizationErrors, 1)
			return &types.LibraryPageData{}, fmt.Errorf("tag organization failed: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled after tag organization: %w", err)
	}

	return &types.LibraryPageData{
		Books:          books,
		BooksByAuthors: authors,    // Changed from Authors
		BooksByGenres:  genres,     // Changed from Genres
		BooksByFormat:  formats,    // Changed from Formats
		BooksByTags:    tags,
	}, nil
}


func (bo *BookOrganizer) GetMetrics() OrganizerMetrics {
	return OrganizerMetrics{
		OrganizationErrors: atomic.LoadInt64(&bo.metrics.OrganizationErrors),
		ItemsOrganized:     atomic.LoadInt64(&bo.metrics.ItemsOrganized),
	}
}

// Helper functions

func (bo *BookOrganizer) organizeByAuthors(ctx context.Context, books []repository.Book) (types.AuthorData, error) {
	if err := ctx.Err(); err != nil {
		return types.AuthorData{}, fmt.Errorf("context cancelled: %w", err)
	}

	if books == nil {
			// Return empty AuthorData struct instead of nil
			return types.AuthorData{
					AllAuthors: make([]string, 0),
					ByAuthor:   make(map[string][]repository.Book),
			}, fmt.Errorf("books slice cannot be nil")
	}

	result := types.AuthorData{
			AllAuthors: make([]string, 0),
			ByAuthor:   make(map[string][]repository.Book),
	}

	authorSet := make(map[string]struct{})

	for i := range books {
		// Initialize nil slices
		if books[i].Authors == nil {
				books[i].Authors = make([]string, 0)
				bo.logger.Debug("initialized nil Authors slice",
						"bookIndex", i,
						"bookTitle", books[i].Title,
				)
				continue
		}

		for _, author := range books[i].Authors {
				if author == "" {
						continue
				}
				result.ByAuthor[author] = append(result.ByAuthor[author], books[i])
				authorSet[author] = struct{}{}
		}
}

	// Convert unique authors to slice
	for author := range authorSet {
			result.AllAuthors = append(result.AllAuthors, author)
	}

	return result, nil
}

func (bo *BookOrganizer) organizeByGenres(ctx context.Context, books []repository.Book) (types.GenreData, error) {
	if err := ctx.Err(); err != nil {
		return types.GenreData{}, fmt.Errorf("context cancelled: %w", err)
	}

	if books == nil {
			return types.GenreData{}, fmt.Errorf("books slice cannot be nil")
	}

	result := types.GenreData{
			AllGenres: make([]string, 0),
			ByGenre:   make(map[string][]repository.Book),
	}

	genreSet := make(map[string]struct{})

	for i := range books {
			if books[i].Genres == nil {
					books[i].Genres = make([]string, 0)
					bo.logger.Warn("book has nil Genres slice",
							"bookIndex", i,
							"bookTitle", books[i].Title,
					)
					continue
			}

			for _, genre := range books[i].Genres {
					if genre == "" {
							bo.logger.Warn("empty genre found",
									"bookIndex", i,
									"bookTitle", books[i].Title,
							)
							continue
					}
					result.ByGenre[genre] = append(result.ByGenre[genre], books[i])
					genreSet[genre] = struct{}{} // Mark genre as seen
			}
	}

	// Convert unique genres to slice
	for genre := range genreSet {
			result.AllGenres = append(result.AllGenres, genre)
	}

	return result, nil
}

func (bo *BookOrganizer) organizeByFormats(ctx context.Context, books []repository.Book) (types.FormatData, error) {
	if err := ctx.Err(); err != nil {
		return types.FormatData{}, fmt.Errorf("context cancelled: %w", err)
	}

	if books == nil {
			return types.FormatData{}, fmt.Errorf("books slice cannot be nil")
	}

	result := types.FormatData{
			AudioBook: make([]repository.Book, 0),
			EBook:     make([]repository.Book, 0),
			Physical:  make([]repository.Book, 0),
	}

	for i, book := range books {
		// Check formats array instead of single format
		for _, format := range book.Formats {
				// Normalize format to lowercase for comparison
				format = strings.ToLower(format)

				switch format {
				case FormatAudioBook:
						result.AudioBook = append(result.AudioBook, book)
				case FormatEBook:
						result.EBook = append(result.EBook, book)
				case FormatPhysical:
						result.Physical = append(result.Physical, book)
				default:
						bo.logger.Warn("unknown format found",
								"format", format,
								"bookIndex", i,
								"bookTitle", book.Title)
				}
		}
}

	return result, nil
}

func (bo *BookOrganizer) organizeByTags(ctx context.Context, books []repository.Book) (types.TagData, error) {
	if err := ctx.Err(); err != nil {
		return types.TagData{}, fmt.Errorf("context cancelled: %w", err)
	}

	if books == nil {
			return types.TagData{}, fmt.Errorf("books slice cannot be nil")
	}

	result := types.TagData{
			AllTags: make([]string, 0),
			ByTag:   make(map[string][]repository.Book),
	}

	tagSet := make(map[string]struct{})

	for i := range books {
		if i > 0 && i%100 == 0 { // Check every 100 items
			if err := ctx.Err(); err != nil {
					return types.TagData{}, fmt.Errorf("context cancelled during tag processing: %w", err)
			}
		}

		// Initialize nil slices
		if books[i].Tags == nil {
				books[i].Tags = make([]string, 0)
				bo.logger.Debug("initialized nil Tags slice",
						"bookIndex", i,
						"bookTitle", books[i].Title,
				)
				continue
		}

		for _, tag := range books[i].Tags {
				if tag == "" {
						bo.logger.Debug("skipping empty tag",
								"bookIndex", i,
								"bookTitle", books[i].Title,
						)
						continue
				}
				result.ByTag[tag] = append(result.ByTag[tag], books[i])
				tagSet[tag] = struct{}{}
		}
	}

	// Convert unique tags to slice
	for tag := range tagSet {
			result.AllTags = append(result.AllTags, tag)
	}

	return result, nil
}