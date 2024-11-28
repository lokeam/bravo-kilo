package types

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/lokeam/bravo-kilo/internal/books/repository"
	"github.com/lokeam/bravo-kilo/internal/shared/core"
)

const maxMemoryLimit = 10 * 1024 * 1024 // 10MB limit

type LibraryResponse struct {
	RequestID   string            `json:"requestId"`
	Data        *LibraryPageData  `json:"data"`
	Source      string            `json:"source"`
}

type LibraryQueryParams struct {
	Domain    core.DomainType     `json:"domain" validate:"required,oneof=books games movies"`
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
	// Early nil check before any operations
	if l == nil {
			return fmt.Errorf("LibraryPageData is nil")
	}

	// Initialize logger if needed
	if l.logger == nil {
			l.logger = slog.Default()
	}

	start := time.Now()
	defer func() {
			l.logger.Debug("validation completed",
					"duration", time.Since(start),
			)
	}()

	// Initialize all required structures
	if err := l.initializeStructures(); err != nil {
			return fmt.Errorf("initialization failed: %w", err)
	}

	// Validate data consistency
	if err := l.validateDataConsistency(); err != nil {
			return fmt.Errorf("consistency validation failed: %w", err)
	}

	return nil
}

func (l *LibraryPageData) initializeStructures() error {
	// Books initialization
	if l.Books == nil {
			l.Books = make([]repository.Book, 0)
	}

	// Authors initialization
	if l.BooksByAuthors.AllAuthors == nil {
			l.BooksByAuthors.AllAuthors = make([]string, 0)
	}
	if l.BooksByAuthors.ByAuthor == nil {
			l.BooksByAuthors.ByAuthor = make(map[string][]repository.Book)
	}

	// Genres initialization
	if l.BooksByGenres.AllGenres == nil {
			l.BooksByGenres.AllGenres = make([]string, 0)
	}
	if l.BooksByGenres.ByGenre == nil {
			l.BooksByGenres.ByGenre = make(map[string][]repository.Book)
	}

	// Format initialization
	if l.BooksByFormat.AudioBook == nil {
			l.BooksByFormat.AudioBook = make([]repository.Book, 0)
	}
	if l.BooksByFormat.EBook == nil {
			l.BooksByFormat.EBook = make([]repository.Book, 0)
	}
	if l.BooksByFormat.Physical == nil {
			l.BooksByFormat.Physical = make([]repository.Book, 0)
	}

	// Tags initialization
	if l.BooksByTags.AllTags == nil {
			l.BooksByTags.AllTags = make([]string, 0)
	}
	if l.BooksByTags.ByTag == nil {
			l.BooksByTags.ByTag = make(map[string][]repository.Book)
	}

	// Initialize book fields if necessary
	for i := range l.Books {
			if l.Books[i].Authors == nil {
					l.Books[i].Authors = make([]string, 0)
			}
			if l.Books[i].Genres == nil {
					l.Books[i].Genres = make([]string, 0)
			}
			if l.Books[i].Tags == nil {
					l.Books[i].Tags = make([]string, 0)
			}
			if l.Books[i].Formats == nil {
					l.Books[i].Formats = make([]string, 0)
			}
	}

	return nil
}


func (l *LibraryPageData) validateDataConsistency() error {
	// Validate basic structure initialization
	if err := l.validateStructureInitialization(); err != nil {
			return fmt.Errorf("structure initialization validation failed: %w", err)
	}

	// Continue with existing validation...
	return nil
}

func (l *LibraryPageData) validateStructureInitialization() error {
	// Check Books array
	if l.Books == nil {
			return fmt.Errorf("Books array is nil")
	}

	// Check BooksByAuthors
	if l.BooksByAuthors.AllAuthors == nil {
			return fmt.Errorf("AllAuthors array is nil")
	}
	if l.BooksByAuthors.ByAuthor == nil {
			return fmt.Errorf("ByAuthor map is nil")
	}

	// Check BooksByGenres
	if l.BooksByGenres.AllGenres == nil {
			return fmt.Errorf("AllGenres array is nil")
	}
	if l.BooksByGenres.ByGenre == nil {
			return fmt.Errorf("ByGenre map is nil")
	}

	// Check BooksByFormat
	if l.BooksByFormat.AudioBook == nil {
			return fmt.Errorf("AudioBook array is nil")
	}
	if l.BooksByFormat.EBook == nil {
			return fmt.Errorf("EBook array is nil")
	}
	if l.BooksByFormat.Physical == nil {
			return fmt.Errorf("Physical array is nil")
	}

	// Check BooksByTags
	if l.BooksByTags.AllTags == nil {
			return fmt.Errorf("AllTags array is nil")
	}
	if l.BooksByTags.ByTag == nil {
			return fmt.Errorf("ByTag map is nil")
	}

	return nil
}

func (l *LibraryPageData) validateBooksIntegrity() error {
	for i, book := range l.Books {
			if book.Title == "" {
					l.logger.Error("invalid book", "index", i, "error", "empty title")
					return fmt.Errorf("book at index %d has empty title", i)
			}
	}
	return nil
}

func (l *LibraryPageData) validateAuthorsConsistency() error {
	// Track all unique authors from Books collection
	authorsSet := make(map[string]bool)
	for _, book := range l.Books {
			for _, author := range book.Authors {
					if author != "" {
							authorsSet[author] = true
					}
			}
	}

	// Verify AllAuthors matches Books collection
	for _, author := range l.BooksByAuthors.AllAuthors {
			if !authorsSet[author] {
					l.logger.Error("inconsistent author data",
							"author", author,
							"error", "author in AllAuthors but not in Books")
					return fmt.Errorf("author %q in AllAuthors but not found in Books", author)
			}
	}

	// Verify ByAuthor map consistency
	for author, books := range l.BooksByAuthors.ByAuthor {
			if !authorsSet[author] {
					l.logger.Error("inconsistent author mapping",
							"author", author,
							"error", "author in ByAuthor but not in Books")
					return fmt.Errorf("author %q in ByAuthor but not found in Books", author)
			}

			// Verify each book in the author's collection
			for _, book := range books {
					found := false
					for _, mainBook := range l.Books {
							if mainBook.ID == book.ID {
									found = true
									break
							}
					}
					if !found {
							l.logger.Error("inconsistent book reference",
									"author", author,
									"bookID", book.ID,
									"error", "book in ByAuthor but not in Books")
							return fmt.Errorf("book ID %d referenced in ByAuthor[%q] but not found in Books", book.ID, author)
					}
			}
	}

	return nil
}

func (l *LibraryPageData) validateGenresConsistency() error {
	// Track all unique genres from Books collection
	genresSet := make(map[string]bool)
	for _, book := range l.Books {
			for _, genre := range book.Genres {
					if genre != "" {
							genresSet[genre] = true
					}
			}
	}

	// Verify AllGenres matches Books collection
	for _, genre := range l.BooksByGenres.AllGenres {
			if !genresSet[genre] {
					l.logger.Error("inconsistent genre data",
							"genre", genre,
							"error", "genre in AllGenres but not in Books")
					return fmt.Errorf("genre %q in AllGenres but not found in Books", genre)
			}
	}

	// Verify ByGenre map consistency
	for genre, books := range l.BooksByGenres.ByGenre {
			if !genresSet[genre] {
					l.logger.Error("inconsistent genre mapping",
							"genre", genre,
							"error", "genre in ByGenre but not in Books")
					return fmt.Errorf("genre %q in ByGenre but not found in Books", genre)
			}

			// Verify each book in the genre's collection
			for _, book := range books {
					found := false
					for _, mainBook := range l.Books {
							if mainBook.ID == book.ID {
									found = true
									break
							}
					}
					if !found {
							l.logger.Error("inconsistent book reference",
									"genre", genre,
									"bookID", book.ID,
									"error", "book in ByGenre but not in Books")
							return fmt.Errorf("book ID %d referenced in ByGenre[%q] but not found in Books", book.ID, genre)
					}
			}
	}

	return nil
}

func (l *LibraryPageData) validateFormatsConsistency() error {
	// Create a map of all books by ID for quick lookup
	// Change the map key type to int to match book.ID
	booksMap := make(map[int]repository.Book)
	for _, book := range l.Books {
			booksMap[book.ID] = book
	}

	// Helper function to validate book list
	validateBookList := func(books []repository.Book, formatName string) error {
			for _, book := range books {
					// Now the types match for the map lookup
					if _, exists := booksMap[book.ID]; !exists {
							l.logger.Error("inconsistent format data",
									"format", formatName,
									"bookID", book.ID,
									"error", "book in format but not in Books")
							return fmt.Errorf("book ID %d in %s format but not found in Books", book.ID, formatName)
					}
			}
			return nil
	}

	// Rest of the function remains the same
	if err := validateBookList(l.BooksByFormat.AudioBook, "AudioBook"); err != nil {
			return err
	}
	if err := validateBookList(l.BooksByFormat.EBook, "EBook"); err != nil {
			return err
	}
	if err := validateBookList(l.BooksByFormat.Physical, "Physical"); err != nil {
			return err
	}

	return nil
}

func (l *LibraryPageData) validateTagsConsistency() error {
	// Track all unique tags from Books collection
	tagsSet := make(map[string]bool)
	for _, book := range l.Books {
			for _, tag := range book.Tags {
					if tag != "" {
							tagsSet[tag] = true
					}
			}
	}

	// Verify AllTags matches Books collection
	for _, tag := range l.BooksByTags.AllTags {
			if !tagsSet[tag] {
					l.logger.Error("inconsistent tag data",
							"tag", tag,
							"error", "tag in AllTags but not in Books")
					return fmt.Errorf("tag %q in AllTags but not found in Books", tag)
			}
	}

	// Verify ByTag map consistency
	for tag, books := range l.BooksByTags.ByTag {
			if !tagsSet[tag] {
					l.logger.Error("inconsistent tag mapping",
							"tag", tag,
							"error", "tag in ByTag but not in Books")
					return fmt.Errorf("tag %q in ByTag but not found in Books", tag)
			}

			// Verify each book in the tag's collection
			for _, book := range books {
					found := false
					for _, mainBook := range l.Books {
							if mainBook.ID == book.ID {
									found = true
									break
							}
					}
					if !found {
							l.logger.Error("inconsistent book reference",
									"tag", tag,
									"bookID", book.ID,
									"error", "book in ByTag but not in Books")
							return fmt.Errorf("book ID %d referenced in ByTag[%q] but not found in Books", book.ID, tag)
					}
			}
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
    // First convert the struct to JSON
    jsonData, err := json.Marshal(lpd)
    if err != nil {
			lpd.logger.Error("json marshal failed",
					"error", err,
					"dataType", fmt.Sprintf("%T", lpd),
			)
			return nil, fmt.Errorf("json marshal failed: %w", err)
		}

    // Check memory limit before proceeding
    if len(jsonData) > maxMemoryLimit {
			lpd.logger.Error("data exceeds memory limit",
					"size", len(jsonData),
					"limit", maxMemoryLimit,
			)
			return nil, fmt.Errorf("data size %d exceeds memory limit %d", len(jsonData), maxMemoryLimit)
	}

    // Create a buffer with capacity hint to avoid reallocations
    buf := bytes.NewBuffer(make([]byte, 0, len(jsonData)+4)) // +4 for length prefix

    // Write length as uint32 (4 bytes)
    if err := binary.Write(buf, binary.LittleEndian, uint32(len(jsonData))); err != nil {
			lpd.logger.Error("failed to write length prefix",
					"error", err,
					"dataLength", len(jsonData),
			)
			return nil, fmt.Errorf("failed to write length prefix: %w", err)
	}

    // Write JSON data
    if _, err := buf.Write(jsonData); err != nil {
			lpd.logger.Error("failed to write json data",
					"error", err,
					"bufferSize", buf.Len(),
					"jsonLength", len(jsonData),
			)
			return nil, fmt.Errorf("failed to write json data: %w", err)
		}

    lpd.logger.Debug("binary marshaling completed",
        "totalSize", buf.Len(),
        "jsonSize", len(jsonData),
    )

    return buf.Bytes(), nil
}

func (lpd *LibraryPageData) UnmarshalBinary(data []byte) error {
	lpd.logger.Debug("starting binary unmarshaling",
			"dataSize", len(data),
	)

	// Need at least 4 bytes for length prefix
	if len(data) < 4 {
			lpd.logger.Error("data too short for length prefix",
					"dataSize", len(data),
					"minimumRequired", 4,
			)
			return fmt.Errorf("data too short: %d bytes", len(data))
	}

	// Read length prefix
	var length uint32
	if err := binary.Read(bytes.NewReader(data[:4]), binary.LittleEndian, &length); err != nil {
			lpd.logger.Error("failed to read length prefix",
					"error", err,
					"dataSize", len(data),
			)
			return fmt.Errorf("failed to read length prefix: %w", err)
	}

	// Check memory limit
	if length > uint32(maxMemoryLimit) {
			lpd.logger.Error("data exceeds memory limit",
					"length", length,
					"limit", maxMemoryLimit,
			)
			return fmt.Errorf("data size %d exceeds memory limit %d", length, maxMemoryLimit)
	}

	// Verify length
	if uint32(len(data)-4) != length {
			lpd.logger.Error("invalid data length",
					"expected", length,
					"actual", len(data)-4,
					"totalSize", len(data),
			)
			return fmt.Errorf("invalid data length: expected %d, got %d", length, len(data)-4)
	}

	// Temporary struct for unmarshaling that matches the JSON structure
	var temp struct {
			Books    []repository.Book `json:"books"`
			Authors  struct {
					AllAuthors []string                     `json:"allAuthors"`
					ByAuthor   map[string][]repository.Book `json:"byAuthor"`
			} `json:"booksByAuthors"`
			Genres struct {
					AllGenres []string                     `json:"allGenres"`
					ByGenre   map[string][]repository.Book `json:"byGenre"`
			} `json:"booksByGenres"`
			Formats struct {
					AudioBook []repository.Book `json:"audioBook"`
					EBook     []repository.Book `json:"eBook"`
					Physical  []repository.Book `json:"physical"`
			} `json:"booksByFormat"`
			Tags struct {
					AllTags []string                     `json:"allTags"`
					ByTag   map[string][]repository.Book `json:"byTag"`
			} `json:"booksByTags"`
	}

	// Unmarshal JSON portion
	if err := json.Unmarshal(data[4:], &temp); err != nil {
			lpd.logger.Error("json unmarshal failed",
					"error", err,
					"jsonSize", length,
			)
			return fmt.Errorf("json unmarshal failed: %w", err)
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
			lpd.logger.Error("validation failed after unmarshal",
					"error", err,
			)
			return fmt.Errorf("validation failed after unmarshal: %w", err)
	}

	lpd.logger.Debug("binary unmarshaling completed",
			"jsonSize", length,
			"totalSize", len(data),
	)

	return nil
}

func (lpd *LibraryPageData) MarshalJSON() ([]byte, error) {
	// Validate before marshaling
	if err := lpd.Validate(); err != nil {
			return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Ensure all slices are initialized
	if lpd.Books == nil {
			lpd.Books = []repository.Book{}
	}
	if lpd.BooksByAuthors.AllAuthors == nil {
			lpd.BooksByAuthors.AllAuthors = []string{}
	}
	if lpd.BooksByGenres.AllGenres == nil {
			lpd.BooksByGenres.AllGenres = []string{}
	}

	// Use a complete view structure
	view := struct {
			Books          []repository.Book `json:"books"`
			BooksByAuthors AuthorData       `json:"booksByAuthors"`
			BooksByGenres  GenreData        `json:"booksByGenres"`
			BooksByFormat  FormatData       `json:"booksByFormat"`
			BooksByTags    TagData          `json:"booksByTags"`
	}{
			Books:          lpd.Books,
			BooksByAuthors: lpd.BooksByAuthors,
			BooksByGenres:  lpd.BooksByGenres,
			BooksByFormat:  lpd.BooksByFormat,
			BooksByTags:    lpd.BooksByTags,
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