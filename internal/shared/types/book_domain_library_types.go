package types

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/lokeam/bravo-kilo/internal/books/repository"
	binaryMarshaler "github.com/lokeam/bravo-kilo/internal/shared/binary"
	"github.com/lokeam/bravo-kilo/internal/shared/core"
)

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
	// Directly use the shared binary marshaler
	data, err := binaryMarshaler.MarshalBinary(lpd)
	if err != nil {
		lpd.logger.Error("library types binary marshal failed",
			"error", err,
		)
		return nil, fmt.Errorf("library types binary marshal failed: %w", err)
	}
	return data, nil
}

func (lpd *LibraryPageData) UnmarshalBinary(data []byte) error {
	lpd.logger.Debug("starting binary unmarshaling",
			"dataSize", len(data),
	)

	// Validate size, min and max
	if len(data) < 4 {
			lpd.logger.Error("data too short for length prefix",
					"dataSize", len(data),
					"minimumRequired", 4,
			)
			return fmt.Errorf("data too short: %d bytes", len(data))
	}

	// Total size check
	totalSize := len(data)
	if totalSize > binaryMarshaler.MaxMemoryLimit {
		lpd.logger.Error("total data exceeds limit",
      "size", totalSize,
      "limit", binaryMarshaler.MaxMemoryLimit,
    )
		return fmt.Errorf("total data size %d exceeds limit %d", totalSize, binaryMarshaler.MaxMemoryLimit)
	}

	// Read and validate length prefix
	var claimedLength uint32
	lengthReader := bytes.NewReader(data[:4])
	if err := binary.Read(lengthReader, binary.LittleEndian, &claimedLength); err != nil {
			lpd.logger.Error("failed to read length prefix",
					"error", err,
			)
			return fmt.Errorf("failed to read length prefix: %w", err)
	}

	// Validate claimed length before using it
	if claimedLength > uint32(binaryMarshaler.MaxMemoryLimit) {
		lpd.logger.Error("claimed length exceeds limit",
				"claimedLength", claimedLength,
				"limit", binaryMarshaler.MaxMemoryLimit,
		)
		return fmt.Errorf("claimed length %d exceeds limit %d", claimedLength, binaryMarshaler.MaxMemoryLimit)
	}

	// Verify actual data matches claimed length
	actualDataLength := uint32(len(data) - 4)
	if actualDataLength != claimedLength {
			lpd.logger.Error("length mismatch",
					"claimed", claimedLength,
					"actual", actualDataLength,
					"totalSize", len(data),
			)
			return fmt.Errorf("length mismatch: claimed %d, actual %d", claimedLength, actualDataLength)
	}

	// JSON Validation
	jsonData := data[4:]
	if !json.Valid(jsonData) {
		lpd.logger.Error("invalid JSON structure",
				"dataSize", len(jsonData),
				"claimedLength", claimedLength,
				"totalSize", len(data),
		)
		return fmt.Errorf("invalid JSON structure in binary data")
	}

	// Unmarshal into temporary structure
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

	// Pre unmarshal data logging
	lpd.logger.Debug("pre-unmarshal data",
    "jsonPreview", string(data[4:min(len(data), 104)]), // First 100 chars after length prefix
		"dataSize", len(data)-4,
	)

	// Unmarshal JSON portion
	if err := json.Unmarshal(data[4:], &temp); err != nil {
			lpd.logger.Error("json unmarshal failed",
					"error", err,
					"jsonSize", claimedLength,
			)
			return fmt.Errorf("json unmarshal failed: %w", err)
	}


	if temp.Authors.AllAuthors == nil {
    temp.Authors.AllAuthors = make([]string, 0)
    lpd.logger.Debug("initialized nil AllAuthors slice")
	}
	if temp.Authors.ByAuthor == nil {
			temp.Authors.ByAuthor = make(map[string][]repository.Book)
			lpd.logger.Debug("initialized nil ByAuthor map")
	}
	if temp.Genres.AllGenres == nil {
			temp.Genres.AllGenres = make([]string, 0)
			lpd.logger.Debug("initialized nil AllGenres slice")
	}
	if temp.Genres.ByGenre == nil {
			temp.Genres.ByGenre = make(map[string][]repository.Book)
			lpd.logger.Debug("initialized nil ByGenre map")
	}
	if temp.Formats.AudioBook == nil {
			temp.Formats.AudioBook = make([]repository.Book, 0)
			lpd.logger.Debug("initialized nil AudioBook slice")
	}
	if temp.Formats.EBook == nil {
			temp.Formats.EBook = make([]repository.Book, 0)
			lpd.logger.Debug("initialized nil EBook slice")
	}
	if temp.Formats.Physical == nil {
			temp.Formats.Physical = make([]repository.Book, 0)
			lpd.logger.Debug("initialized nil Physical slice")
	}
	if temp.Tags.AllTags == nil {
			temp.Tags.AllTags = make([]string, 0)
			lpd.logger.Debug("initialized nil AllTags slice")
	}
	if temp.Tags.ByTag == nil {
			temp.Tags.ByTag = make(map[string][]repository.Book)
			lpd.logger.Debug("initialized nil ByTag map")
	}


	// post unmarshal validation
	if temp.Books == nil {
    lpd.logger.Error("unmarshaled nil Books slice")
			return fmt.Errorf("unmarshaled nil Books slice")
	}
	lpd.logger.Debug("unmarshal results",
    "bookCount", len(temp.Books),
    "firstBook", func() string {
        if len(temp.Books) > 0 {
            return fmt.Sprintf("%+v", temp.Books[0])
        }
        return "no books"
		}(),
	)

	// Validate individual books
	for i, book := range temp.Books {
		if len(book.Authors) > 0 && book.Authors[0] == "" {
				lpd.logger.Error("book has empty author",
						"bookIndex", i,
						"bookTitle", book.Title,
				)
				return fmt.Errorf("book %q has empty author", book.Title)
		}
	}

	// Pre assignment validation
	for i, book := range temp.Books {
		if book.Authors == nil {
				lpd.logger.Error("nil Authors slice detected",
						"bookIndex", i,
						"bookTitle", book.Title,
				)
				// Initialize empty slice instead of failing
				temp.Books[i].Authors = make([]string, 0)
		}
		if book.Genres == nil {
			lpd.logger.Error("nil Genres slice detected",
					"bookIndex", i,
					"bookTitle", book.Title,
			)
			temp.Books[i].Genres = make([]string, 0)
		}
    if book.Tags == nil {
			temp.Books[i].Tags = make([]string, 0)     // Change: Use temp.Books[i] instead of book
		}
		if book.Formats == nil {
				temp.Books[i].Formats = make([]string, 0)  // Change: Use temp.Books[i] instead of book
		}
	}

	// Update fields with proper type conversion
	lpd.Books = temp.Books

	// Post assignment validation
	lpd.logger.Debug("field assignment verification",
			"sourceAuthorsLen", func() int {
					if len(temp.Books) > 0 {
							return len(temp.Books[0].Authors)
					}
					return -1
			}(),
			"destAuthorsLen", func() int {
					if len(lpd.Books) > 0 {
							return len(lpd.Books[0].Authors)
					}
					return -1
			}(),
	)


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
			"jsonSize", claimedLength,
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