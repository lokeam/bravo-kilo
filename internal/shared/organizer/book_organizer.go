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

	bo.logger.Debug("ORGANIZER: Starting library organization",
	"component", "book_organizer",
	"function", "OrganizeForLibrary",
	"booksCount", len(books),
	"firstBookDetails", logBookDetails(books[0]), // Add this helper function
	)

	// Initialize result with empty collections
	result := &types.LibraryPageData{
		Books:          books,
		BooksByAuthors: types.AuthorData{AllAuthors: make([]string, 0), ByAuthor: make(map[string][]repository.Book)},
		BooksByGenres:  types.GenreData{AllGenres: make([]string, 0), ByGenre: make(map[string][]repository.Book)},
		BooksByFormat:  types.FormatData{AudioBook: make([]repository.Book, 0), EBook: make([]repository.Book, 0), Physical: make([]repository.Book, 0)},
		BooksByTags:    types.TagData{AllTags: make([]string, 0), ByTag: make(map[string][]repository.Book)},
	}

	// Track if we had any errors
	var hadErrors bool

	// Build author data
	if authors, err := bo.organizeByAuthors(ctx, books); err != nil {
		hadErrors = true
		bo.logger.Error("author organization failed, continuing with empty author data",
				"error", err)
		atomic.AddInt64(&bo.metrics.OrganizationErrors, 1)
	} else {
		result.BooksByAuthors = authors
	}

    // Build genre data
    if genres, err := bo.organizeByGenres(ctx, books); err != nil {
			hadErrors = true
			bo.logger.Error("genre organization failed, continuing with empty genre data",
					"error", err)
			atomic.AddInt64(&bo.metrics.OrganizationErrors, 1)
		} else {
			result.BooksByGenres = genres
		}

    // Build format data
    if formats, err := bo.organizeByFormats(ctx, books); err != nil {
			hadErrors = true
			bo.logger.Error("format organization failed, continuing with empty format data",
					"error", err)
			atomic.AddInt64(&bo.metrics.OrganizationErrors, 1)
		} else {
			result.BooksByFormat = formats
		}

    // Build tag data
    if tags, err := bo.organizeByTags(ctx, books); err != nil {
			hadErrors = true
			bo.logger.Error("tag organization failed, continuing with empty tag data",
					"error", err)
			atomic.AddInt64(&bo.metrics.OrganizationErrors, 1)
		} else {
			result.BooksByTags = tags
		}

    bo.logger.Debug("ORGANIZER: Completed library organization",
        "component", "book_organizer",
        "function", "OrganizeForLibrary",
        "resultBooksCount", len(result.Books),
        "firstResultBookDetails", logBookDetails(result.Books[0]),
        "hadErrors", hadErrors,
    )

    // Return partial results with error indication
    if hadErrors {
			return result, fmt.Errorf("some organization operations failed, partial results returned")
		}

		return result, nil
}


func (bo *BookOrganizer) GetMetrics() OrganizerMetrics {
	return OrganizerMetrics{
		OrganizationErrors: atomic.LoadInt64(&bo.metrics.OrganizationErrors),
		ItemsOrganized:     atomic.LoadInt64(&bo.metrics.ItemsOrganized),
	}
}

// Helper functions

func (bo *BookOrganizer) organizeByAuthors(ctx context.Context, books []repository.Book) (types.AuthorData, error) {
	bo.logger.Debug("ORGANIZER: raw books received",
        "component", "book_organizer",
        "booksCount", len(books),
        "firstBook", books[0])

	if err := ctx.Err(); err != nil {
			return types.AuthorData{}, fmt.Errorf("context cancelled: %w", err)
	}

	// Init empty result to return if books slice is nil
	emptyResult := types.AuthorData{
		AllAuthors: make([]string, 0),
		ByAuthor:   make(map[string][]repository.Book),
	}

	// Handle nil books slice
	if books == nil {
			return types.AuthorData{
					AllAuthors: make([]string, 0),
					ByAuthor:   make(map[string][]repository.Book),
			}, fmt.Errorf("books slice cannot be nil")
	}

	// Handle empty books slice
	if len(books) == 0 {
		return emptyResult, nil
	}

	bo.logger.Debug("starting author organization",
			"booksCount", len(books),
			"firstBookAuthors", books[0].Authors,
			"firstBookTitle", books[0].Title)

	result := types.AuthorData{
			AllAuthors: make([]string, 0),
			ByAuthor:   make(map[string][]repository.Book),
	}

	authorSet := make(map[string]struct{})  // Moved up before use

	for i, book := range books {
			bo.logger.Debug("processing book authors",
					"bookIndex", i,
					"bookTitle", book.Title,
					"rawAuthors", book.Authors)

			if len(book.Authors) == 0 {
					bo.logger.Warn("book has no authors",
							"bookIndex", i,
							"bookTitle", book.Title)
					continue
			}

			for _, author := range book.Authors {
					if author == "" {
							continue
					}
					bookCopy := book
					result.ByAuthor[author] = append(result.ByAuthor[author], bookCopy)
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

func logBookDetails(book repository.Book) map[string]interface{} {
	return map[string]interface{}{
			"id":           book.ID,
			"title":        book.Title,
			"authorCount":  len(book.Authors),
			"authors":      book.Authors,
			"genreCount":   len(book.Genres),
			"genres":       book.Genres,
			"formatCount":  len(book.Formats),
			"formats":      book.Formats,
			"tagCount":     len(book.Tags),
			"tags":         book.Tags,
	}
}