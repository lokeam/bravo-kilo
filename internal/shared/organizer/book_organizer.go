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

// Struct for homepage stat calculator
type statCalculator struct {
	getItems func(book repository.Book) []string
	targetSlice *[]types.StatItem
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
func (bo *BookOrganizer) OrganizeForLibrary(
	ctx context.Context,
	items *types.LibraryPageData,
	) (*types.LibraryPageData, error) {
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

func (bo *BookOrganizer) OrganizeForHome(ctx context.Context, items *types.HomePageData) (*types.HomePageData, error) {
	// 1. Guard clauses
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}
	if err := ctx.Err(); err != nil {
			atomic.AddInt64(&bo.metrics.OrganizationErrors, 1)
			return nil, fmt.Errorf("context error before organization: %w", err)
	}
	if items == nil {
			atomic.AddInt64(&bo.metrics.OrganizationErrors, 1)
			return &types.HomePageData{}, fmt.Errorf("items cannot be nil")
	}

	// 2. Initialize result with empty collections
	books := items.Books

	bo.logger.Debug("ORGANIZER: Starting home organization",
			"component", "book_organizer",
			"function", "OrganizeForHome",
			"booksCount", len(books),
	)

	result := types.NewHomePageData(bo.logger)
	result.Books = books

	// 3. Track organization errors
	var hadErrors bool

	// 4. Organize format counts
	formatCounts, err := bo.CalculateFormatCounts(books)
	if err != nil {
			hadErrors = true
			bo.logger.Error("format count calculation failed",
					"error", err)
			atomic.AddInt64(&bo.metrics.OrganizationErrors, 1)
	} else {
			result.BooksByFormat = formatCounts
	}

	// 5. Organize homepage statistics
	stats, err := bo.calculateHomePageStats(books)
	if err != nil {
			hadErrors = true
			bo.logger.Error("homepage stats calculation failed",
					"error", err)
			atomic.AddInt64(&bo.metrics.OrganizationErrors, 1)
	} else {
			result.HomePageStats = stats
	}

	bo.logger.Debug("ORGANIZER: Completed home organization",
	"component", "book_organizer",
	"function", "OrganizeForHome",
	"resultBooksCount", len(result.Books),
	"hadErrors", hadErrors,
	)

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

// Helper functions - OrganizeForLibrary

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



// Helper functions - OrganizeForHome
// Calculate stats for individual book attribute
func (bo *BookOrganizer) calculateStats(
	books []repository.Book,
	getItems func(repository.Book) []string,
	target *[]types.StatItem,
) {
	// Count occurences of each item in the target slice
	countMap := make(map[string]int)
	for _, book := range books {
		for _, item := range getItems(book) {
			countMap[item]++
		}
	}

	// Convert to StatItem type slice
	*target = make([]types.StatItem, 0, len(countMap))
	for label, count := range countMap {
		*target = append(*target, types.StatItem{
			Label: label,
			Count: count,
		})
	}
}


func (bo *BookOrganizer) calculateHomePageStats(
	books []repository.Book,
) (types.HomePageStats, error) {
	stats := types.HomePageStats{
		UserBkLang:  types.LanguageStats{BooksByLang: make([]types.StatItem, 0)},
		UserBkGenre: types.GenreStats{BooksByGenre: make([]types.StatItem, 0)},
		UserTags:    types.TagStats{UserTags: make([]types.StatItem, 0)},
		UserAuthors: types.AuthorStats{BooksByAuthor: make([]types.StatItem, 0)},
	}

	// Define calculations
	calculations := []statCalculator{
		{
				// Language stats (single string)
				getItems:    func(book repository.Book) []string { return []string{book.Language} },
				targetSlice: &stats.UserBkLang.BooksByLang,
		},
		{
				// Genre stats (slice)
				getItems:    func(book repository.Book) []string { return book.Genres },
				targetSlice: &stats.UserBkGenre.BooksByGenre,
		},
		{
				// Tag stats (slice)
				getItems:    func(book repository.Book) []string { return book.Tags },
				targetSlice: &stats.UserTags.UserTags,
		},
		{
				// Author stats (slice)
				getItems:    func(book repository.Book) []string { return book.Authors },
				targetSlice: &stats.UserAuthors.BooksByAuthor,
		},
	}

	// Calculate all stats using the same pattern
	for _, calc := range calculations {
			bo.calculateStats(books, calc.getItems, calc.targetSlice)
	}

return stats, nil


}

func (bo *BookOrganizer) CalculateFormatCounts(books []repository.Book) (types.FormatCountStats, error) {
	counts := types.FormatCountStats{}

	bo.logger.Debug("ORGANIZER: Starting format count calculation",
			slog.String("function", "github.com/lokeam/bravo-kilo/internal/shared/organizer.CalculateFormatCounts"),
			slog.String("file", "internal/shared/organizer/book_organizer.go"),
			slog.Int("total_books", len(books)))

	for _, book := range books {
			bo.logger.Debug("ORGANIZER: Processing book formats",
					slog.String("function", "github.com/lokeam/bravo-kilo/internal/shared/organizer.CalculateFormatCounts"),
					slog.String("file", "internal/shared/organizer/book_organizer.go"),
					slog.Int("bookID", book.ID),
					slog.Any("formats", book.Formats))

			for _, format := range book.Formats {
					normalizedFormat := strings.ToLower(format)

					bo.logger.Debug("ORGANIZER: Processing format",
							slog.String("function", "github.com/lokeam/bravo-kilo/internal/shared/organizer.CalculateFormatCounts"),
							slog.String("file", "internal/shared/organizer/book_organizer.go"),
							slog.Int("bookID", book.ID),
							slog.String("original_format", format),
							slog.String("normalized_format", normalizedFormat))

					switch normalizedFormat {
					case "physical":
							counts.Physical++
							bo.logger.Debug("ORGANIZER: Incremented physical count",
									slog.String("function", "github.com/lokeam/bravo-kilo/internal/shared/organizer.CalculateFormatCounts"),
									slog.String("file", "internal/shared/organizer/book_organizer.go"),
									slog.Int("current_count", counts.Physical))
					case "digital":
							counts.Digital++
					case "audiobook":
							counts.AudioBook++
					}
			}
	}

	bo.logger.Debug("ORGANIZER: Final format counts",
			slog.String("function", "github.com/lokeam/bravo-kilo/internal/shared/organizer.CalculateFormatCounts"),
			slog.String("file", "internal/shared/organizer/book_organizer.go"),
			slog.Int("physical", counts.Physical),
			slog.Int("digital", counts.Digital),
			slog.Int("audiobook", counts.AudioBook))

	return counts, nil
}