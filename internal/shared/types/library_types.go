package types

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/lokeam/bravo-kilo/internal/books/repository"
	"github.com/lokeam/bravo-kilo/internal/shared/core"
)

type LibraryResponse struct {
	RequestID   string            `json"requestId"`
	Data        *LibraryPageData  `json:"data"`
	Source      string            `json:"source"`
}

type LibraryQueryParams struct {
	Domain    core.DomainType    `json:"domain" validate:required,oneof=books, games movies`
	Page      int                 `json:"page" validate:"required,min=1,max=99999`
	Limit     int                 `json:"limit" validate:"required,min=1,max=999"`
}

type LibraryPageData struct {
	Books           []repository.Book  `json:"books"`
	BooksByAuthors  AuthorData        `json:"booksByAuthors"`
	BooksByGenres   GenreData         `json:"booksByGenres"`
	BooksByFormat   FormatData        `json:"booksByFormat"`
	BooksByTags     TagData           `json:"booksByTags"`
	logger          *slog.Logger
	validationConf  *ValidationConfig
}

// Match frontend contracts exactly
type AuthorData struct {
	AllAuthors []string                     `json:"allAuthors"`
	ByAuthor   map[string][]repository.Book `json:"byAuthor"`
}

type GenreData struct {
	AllGenres []string                     `json:"allGenres"`
	ByGenre   map[string][]repository.Book `json:"byGenre"`
}

type FormatData struct {
	AudioBook []repository.Book `json:"audioBook"`
	EBook     []repository.Book `json:"eBook"`
	Physical  []repository.Book `json:"physical"`
}

type TagData struct {
	AllTags []string                     `json:"allTags"`
	ByTag   map[string][]repository.Book `json:"byTag"`
}



func NewLibraryPageData(logger *slog.Logger) *LibraryPageData {
	if logger == nil {
			logger = slog.Default()
	}

	conf := DefaultValidationConfig()

	return &LibraryPageData{
		Books:    make([]repository.Book, 0),
		BooksByAuthors:  AuthorData{
				AllAuthors: make([]string, 0),
				ByAuthor:   make(map[string][]repository.Book),
		},
		BooksByGenres:   GenreData{
				AllGenres: make([]string, 0),
				ByGenre:   make(map[string][]repository.Book),
		},
		BooksByFormat:  FormatData{
				AudioBook: make([]repository.Book, 0),
				EBook:     make([]repository.Book, 0),
				Physical:  make([]repository.Book, 0),
		},
		BooksByTags:     TagData{
				AllTags: make([]string, 0),
				ByTag:   make(map[string][]repository.Book),
		},
		logger:          logger,
		validationConf:  conf,
	}
}

// Checks if LibraryPageData is valid
func (l *LibraryPageData) Validate() error {
	if l.logger == nil {
		l.logger = slog.Default()
	}
	start := time.Now()
	defer func() {
			l.logger.Debug("validation completed",
					"duration", time.Since(start),
			)
	}()

	l.logger.Debug("starting validation")

	if l == nil {
			return fmt.Errorf("LibraryPageData is nil")
	}

	validateStart := time.Now()
	if err := l.validateStructure(); err != nil {
			l.logger.Error("structure validation failed",
					"error", err,
					"duration", time.Since(validateStart),
			)
			return fmt.Errorf("structure validation failed: %w", err)
	}

	return nil
}

// New validation methods
func (l *LibraryPageData) validateBooks() error {
	if l.Books == nil {
			l.Books = make([]repository.Book, 0)
			l.logger.Debug("initialized nil Books slice")
			return nil
	}

	for i, book := range l.Books {
			if book.Title == "" {
					l.logger.Error("invalid book", "index", i, "error", "empty title")
					return fmt.Errorf("book at index %d has empty title", i)
			}
	}

	return nil
}

func (l *LibraryPageData) validateStructure() error {
	if l.Books == nil {
			return fmt.Errorf("books slice cannot be nil")
	}

	// Validate Authors
	if err := l.validateAuthorsData(); err != nil {
		return fmt.Errorf("authors validation failed: %w", err)
	}

	// Validate Genres
	if err := l.validateGenresData(); err != nil {
		return fmt.Errorf("genres validation failed: %w", err)
	}

	// Validate Formats (specific format validation)
	if err := l.validateFormatsData(); err != nil {
		return fmt.Errorf("format validation failed: %w", err)
	}

	// Validate Tags
	if err := l.validateTagsData(); err != nil {
		return fmt.Errorf("tags validation failed: %w", err)
	}

	return nil
}

func (l *LibraryPageData) validateAuthorsData() error {
	if l.BooksByAuthors.AllAuthors == nil {
			l.BooksByAuthors.AllAuthors = make([]string, 0)
			l.logger.Debug("initialized nil AllAuthors slice")
	}

	if l.BooksByAuthors.ByAuthor == nil {
			l.BooksByAuthors.ByAuthor = make(map[string][]repository.Book)
			l.logger.Debug("initialized nil ByAuthor map")
	}

	return nil
}

func (l *LibraryPageData) validateGenresData() error {
	// Initialize nil slices/maps with zero length
	if l.BooksByGenres.AllGenres == nil {
			l.BooksByGenres.AllGenres = make([]string, 0)
			l.logger.Debug("initialized nil AllGenres slice")
	}

	if l.BooksByGenres.ByGenre == nil {
			l.BooksByGenres.ByGenre = make(map[string][]repository.Book)
			l.logger.Debug("initialized nil ByGenre map")
	}

	// Check each genre in AllGenres
	for _, genre := range l.BooksByGenres.AllGenres {
			// Verify non-empty genre
			if genre == "" {
					l.logger.Error("invalid genre", "error", "empty genre name")
					return fmt.Errorf("empty genre name in AllGenres")
			}

			// Ensure genre exists in map
			if _, exists := l.BooksByGenres.ByGenre[genre]; !exists {
					l.logger.Error("inconsistent genre data", "genre", genre)
					return fmt.Errorf("genre %q in AllGenres but missing from ByGenre map", genre)
			}
	}

	return nil
}

func (l *LibraryPageData) validateFormatsData() error {
	// Initialize nil slices
	if l.BooksByFormat.AudioBook == nil {
			l.BooksByFormat.AudioBook = make([]repository.Book, 0)
			l.logger.Debug("initialized nil AudioBook slice")
	}
	if l.BooksByFormat.EBook == nil {
			l.BooksByFormat.EBook = make([]repository.Book, 0)
			l.logger.Debug("initialized nil EBook slice")
	}
	if l.BooksByFormat.Physical == nil {
			l.BooksByFormat.Physical = make([]repository.Book, 0)
			l.logger.Debug("initialized nil Physical slice")
	}

	// Validate AudioBook entries
	for i, book := range l.BooksByFormat.AudioBook {
			if !hasFormat(book, "audioBook") {
					l.logger.Error("format mismatch",
							"book", book.Title,
							"index", i,
							"expected", "audioBook",
					)
					return fmt.Errorf("book %q in AudioBook list but missing audioBook format", book.Title)
			}
	}

	// Validate EBook entries
	for i, book := range l.BooksByFormat.EBook {
			if !hasFormat(book, "eBook") {
					l.logger.Error("format mismatch",
							"book", book.Title,
							"index", i,
							"expected", "eBook",
					)
					return fmt.Errorf("book %q in EBook list but missing eBook format", book.Title)
			}
	}

	// Validate Physical entries
	for i, book := range l.BooksByFormat.Physical {
			if !hasFormat(book, "physical") {
					l.logger.Error("format mismatch",
							"book", book.Title,
							"index", i,
							"expected", "physical",
					)
					return fmt.Errorf("book %q in Physical list but missing physical format", book.Title)
			}
	}

	return nil
}

func hasFormat(book repository.Book, format string) bool {
	for _, f := range book.Formats {
			if f == format {
					return true
			}
	}
	return false
}

func (l *LibraryPageData) validateTagsData() error {
	// Initialize nil slices/maps
	if l.BooksByTags.AllTags == nil {
			l.BooksByTags.AllTags = make([]string, 0)
			l.logger.Debug("initialized nil AllTags slice")
	}

	if l.BooksByTags.ByTag == nil {
			l.BooksByTags.ByTag = make(map[string][]repository.Book)
			l.logger.Debug("initialized nil ByTag map")
	}

	// Check each tag in AllTags
	for _, tag := range l.BooksByTags.AllTags {
			// Verify non-empty tag
			if tag == "" {
					l.logger.Error("invalid tag", "error", "empty tag name")
					return fmt.Errorf("empty tag name in AllTags")
			}

			// Ensure tag exists in map
			if _, exists := l.BooksByTags.ByTag[tag]; !exists {
					l.logger.Error("inconsistent tag data", "tag", tag)
					return fmt.Errorf("tag %q in AllTags but missing from ByTag map", tag)
			}
	}

	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (lpd *LibraryPageData) MarshalBinary() ([]byte, error) {
	if lpd == nil {
			return nil, fmt.Errorf("cannot marshal nil LibraryPageData")
	}

	if err := lpd.Validate(); err != nil {
			return nil, fmt.Errorf("validation failed before marshal: %w", err)
	}

	// Create a copy with the correct types
	dataCopy := struct {
			Books    []repository.Book `json:"books"`
			Authors  AuthorData       `json:"booksByAuthors"`
			Genres   GenreData        `json:"booksByGenres"`
			Formats  FormatData       `json:"booksByFormat"`
			Tags     TagData          `json:"booksByTags"`
	}{
			Books:    lpd.Books,
			Authors:  lpd.BooksByAuthors,
			Genres:   lpd.BooksByGenres,
			Formats:  lpd.BooksByFormat,
			Tags:     lpd.BooksByTags,
	}

	return json.Marshal(dataCopy)
}

func (lpd *LibraryPageData) UnmarshalBinary(data []byte) error {
	if data == nil {
			return fmt.Errorf("cannot unmarshal nil data")
	}

	// Temporary struct for unmarshaling that matches the JSON structure
	var temp struct {
			Books    []repository.Book `json:"books"`
			Authors  struct {
					AllAuthors []string                     `json:"allAuthors"`
					ByAuthor   map[string][]repository.Book `json:"-"`
			} `json:"booksByAuthors"`
			Genres struct {
					AllGenres []string                     `json:"allGenres"`
					ByGenre   map[string][]repository.Book `json:"-"`
			} `json:"booksByGenres"`
			Formats struct {
					AudioBook []repository.Book `json:"audioBook"`
					EBook     []repository.Book `json:"eBook"`
					Physical  []repository.Book `json:"physical"`
			} `json:"booksByFormat"`
			Tags struct {
					AllTags []string                     `json:"allTags"`
					ByTag   map[string][]repository.Book `json:"-"`
			} `json:"booksByTags"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
			return fmt.Errorf("unmarshal failed: %w", err)
	}

	// Update fields with proper type conversion
	lpd.Books = temp.Books
	lpd.BooksByAuthors = AuthorData{
			AllAuthors: temp.Authors.AllAuthors,
			ByAuthor:   temp.Authors.ByAuthor,
	}
	lpd.BooksByGenres = GenreData{
			AllGenres: temp.Genres.AllGenres,
			ByGenre:   temp.Genres.ByGenre,
	}
	lpd.BooksByFormat = FormatData{
			AudioBook: temp.Formats.AudioBook,
			EBook:     temp.Formats.EBook,
			Physical:  temp.Formats.Physical,
	}
	lpd.BooksByTags = TagData{
			AllTags: temp.Tags.AllTags,
			ByTag:   temp.Tags.ByTag,
	}

	// Validate after unmarshaling
	if err := lpd.Validate(); err != nil {
			return fmt.Errorf("validation failed after unmarshal: %w", err)
	}

	return nil
}

func (lpd *LibraryPageData) MarshalJSON() ([]byte, error) {
	if err := lpd.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Define the view structure with its own AuthorData type
	type AuthorView struct {
			AllAuthors []string                     `json:"allAuthors"`
			ByAuthor   map[string][]repository.Book `json:"byAuthor"`
	}

	type JSONView struct {
			Books          []repository.Book `json:"books"`
			BooksByAuthors AuthorView       `json:"booksByAuthors"`
			// ... other fields
	}

	// Create the view using the defined types
	view := JSONView{
			Books: lpd.Books,
			BooksByAuthors: AuthorView{
					AllAuthors: lpd.BooksByAuthors.AllAuthors,
					ByAuthor:   lpd.BooksByAuthors.ByAuthor,
			},
	}

	return json.Marshal(view)
}

func (lpd *LibraryPageData) UnmarshalJSON(data []byte) error {
	if data == nil {
			return fmt.Errorf("cannot unmarshal nil data")
	}

	// Add debug logging
	lpd.logger.Debug("starting unmarshal",
			"dataLength", len(data),
			"rawData", string(data), // Log first 100 chars only in production
	)

	// Temporary struct to avoid recursive unmarshaling
	type Alias LibraryPageData
	temp := struct {
			*Alias
			Authors AuthorData `json:"booksByAuthors"`
			Genres  GenreData  `json:"booksByGenres"`
			Formats FormatData `json:"booksByFormat"`
			Tags    TagData    `json:"booksByTags"`
	}{
			Alias: (*Alias)(lpd),
	}

	if err := json.Unmarshal(data, &temp); err != nil {
			return fmt.Errorf("unmarshal failed: %w", err)
	}

	// Add debug logging for book data
	lpd.logger.Debug("unmarshaled book data",
			"bookCount", len(temp.Books),
			"firstBookTitle", func() string {
					if len(temp.Books) > 0 {
							return temp.Books[0].Title
					}
					return "no books"
			}(),
	)

	// Update the fields
	lpd.Books = temp.Books
	lpd.BooksByAuthors = temp.Authors
	lpd.BooksByGenres = temp.Genres
	lpd.BooksByFormat = temp.Formats
	lpd.BooksByTags = temp.Tags

	// Add debug logging post-assignment
	lpd.logger.Debug("after field assignment",
			"bookCount", len(lpd.Books),
			"authorCount", len(lpd.BooksByAuthors.AllAuthors),
			"genreCount", len(lpd.BooksByGenres.AllGenres),
	)

	// Validate after unmarshaling
	if err := lpd.Validate(); err != nil {
			return fmt.Errorf("validation failed after unmarshal: %w", err)
	}

	return nil
}