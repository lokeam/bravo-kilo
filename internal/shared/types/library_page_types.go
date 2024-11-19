package types

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"sync"
	"time"

	"github.com/lokeam/bravo-kilo/internal/books/repository"
)

type LibraryPageData struct {
	mu              sync.RWMutex
	Books           []repository.Book            `json:"books"`
	Authors         map[string][]repository.Book `json:"authors"`
	Genres          map[string][]repository.Book `json:"genres"`
	Formats         map[string][]repository.Book `json:"formats"`
	Tags            map[string][]repository.Book `json:"tags"`
	logger          *slog.Logger
	validationConf  *ValidationConfig
	categoryPattern *regexp.Regexp
}

func NewLibraryPageData(logger *slog.Logger) *LibraryPageData {
	if logger == nil {
			logger = slog.Default()
	}

	conf := DefaultValidationConfig()
	pattern, err := regexp.Compile(conf.CategoryPattern)
	if err != nil {
			logger.Error("failed to compile category pattern", "error", err)
			// Use a safe default
			pattern = regexp.MustCompile(`^[a-zA-Z0-9\s\-_]+$`)
	}

	return &LibraryPageData{
			Books:           make([]repository.Book, 0),
			Authors:         make(map[string][]repository.Book),
			Genres:          make(map[string][]repository.Book),
			Formats:         make(map[string][]repository.Book),
			Tags:            make(map[string][]repository.Book),
			logger:          logger,
			validationConf:  conf,
			categoryPattern: pattern,
	}
}

func (l *LibraryPageData) SetValidationConfig(conf *ValidationConfig) error {
	if conf == nil {
			return fmt.Errorf("validation config cannot be nil")
	}

	pattern, err := regexp.Compile(conf.CategoryPattern)
	if err != nil {
			return fmt.Errorf("invalid category pattern: %w", err)
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.validationConf = conf
	l.categoryPattern = pattern
	return nil
}

// Checks if LibraryPageData is valid
func (l *LibraryPageData) Validate() error {
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

	l.mu.RLock()
	defer l.mu.RUnlock()

	validateStart := time.Now()
	if err := l.validateStructure(); err != nil {
			l.logger.Error("structure validation failed",
					"error", err,
					"duration", time.Since(validateStart),
			)
			return fmt.Errorf("structure validation failed: %w", err)
	}
	l.logger.Debug("structure validation completed",
			"duration", time.Since(validateStart),
	)

	// Validate map contents with performance tracking
	for _, validation := range []struct {
			name string
			data map[string][]repository.Book
	}{
			{"Authors", l.Authors},
			{"Genres", l.Genres},
			{"Formats", l.Formats},
			{"Tags", l.Tags},
	} {
			mapStart := time.Now()
			if err := l.validateBookMap(validation.name, validation.data); err != nil {
					l.logger.Error("map validation failed",
							"map", validation.name,
							"error", err,
							"duration", time.Since(mapStart),
					)
					return fmt.Errorf("invalid %s: %w", validation.name, err)
			}
			l.logger.Debug("map validation completed",
					"map", validation.name,
					"duration", time.Since(mapStart),
			)
	}

	return nil
}

func (l *LibraryPageData) validateStructure() error {
	if l.Books == nil {
			return fmt.Errorf("books slice is nil")
	}
	if l.Authors == nil {
			return fmt.Errorf("authors map is nil")
	}
	if l.Genres == nil {
			return fmt.Errorf("genres map is nil")
	}
	if l.Formats == nil {
			return fmt.Errorf("formats map is nil")
	}
	if l.Tags == nil {
			return fmt.Errorf("tags map is nil")
	}
	return nil
}

func (l *LibraryPageData) validateBookMap(mapName string, books map[string][]repository.Book) error {
	for category, bookList := range books {
			start := time.Now()
			if err := l.validateCategory(category); err != nil {
					return fmt.Errorf("invalid category %q: %w", category, err)
			}
			if err := l.validateBookList(category, bookList); err != nil {
					return fmt.Errorf("invalid books for category %q: %w", category, err)
			}
			l.logger.Debug("category validation completed",
					"map", mapName,
					"category", category,
					"bookCount", len(bookList),
					"duration", time.Since(start),
			)
	}
	return nil
}

func (l *LibraryPageData) validateCategory(category string) error {
	if len(category) == 0 {
			return fmt.Errorf("empty category")
	}
	if len(category) > l.validationConf.MaxCategoryLength {
			return fmt.Errorf("category exceeds maximum length of %d", l.validationConf.MaxCategoryLength)
	}
	if !l.categoryPattern.MatchString(category) {
			return fmt.Errorf("category contains invalid characters")
	}
	return nil
}

func (l *LibraryPageData) validateBookList(category string, books []repository.Book) error {
	if books == nil {
			return fmt.Errorf("nil book slice")
	}

	for i, book := range books {
			start := time.Now()
			if err := l.validateBook(book); err != nil {
					l.logger.Error("book validation failed",
							"category", category,
							"bookIndex", i,
							"error", err,
							"duration", time.Since(start),
					)
					return fmt.Errorf("invalid book at index %d: %w", i, err)
			}
	}
	return nil
}

func (l *LibraryPageData) validateBook(book repository.Book) error {
	if book.ID == 0 {
			return fmt.Errorf("book ID cannot be zero")
	}

	if len(book.Title) < l.validationConf.MinTitleLength {
			return fmt.Errorf("title length below minimum (%d)", l.validationConf.MinTitleLength)
	}

	if len(book.Title) > l.validationConf.MaxBookTitleLength {
			return fmt.Errorf("title exceeds maximum length (%d)", l.validationConf.MaxBookTitleLength)
	}

	return nil
}

// DeepCopy creates a thread-safe deep copy
func (l *LibraryPageData) DeepCopy(source *LibraryPageData) error {
	if source == nil {
			return fmt.Errorf("source data cannot be nil")
	}
	if source.logger == nil {
			return fmt.Errorf("source logger cannot be nil")
	}

	// Lock both source and destination
	source.mu.RLock()
	l.mu.Lock()
	defer source.mu.RUnlock()
	defer l.mu.Unlock()

	// Helper function to safely copy book slice
	copyBooks := func(books []repository.Book) []repository.Book {
			if len(books) == 0 {
					return make([]repository.Book, 0)
			}
			copied := make([]repository.Book, len(books))
			copy(copied, books)
			return copied
	}

	// Helper function to safely copy book map
	copyBookMap := func(m map[string][]repository.Book) map[string][]repository.Book {
			copied := make(map[string][]repository.Book, len(m))
			for k, v := range m {
					copied[k] = copyBooks(v)
			}
			return copied
	}

	// Copy all fields
	l.Books = copyBooks(source.Books)
	l.Authors = copyBookMap(source.Authors)
	l.Genres = copyBookMap(source.Genres)
	l.Formats = copyBookMap(source.Formats)
	l.Tags = copyBookMap(source.Tags)
	l.logger = source.logger

	if source.validationConf != nil {
		l.validationConf = &ValidationConfig{
			MaxCategoryLength:   source.validationConf.MaxCategoryLength,
			MinTitleLength:      source.validationConf.MinTitleLength,
			MaxBookTitleLength:  source.validationConf.MaxBookTitleLength,
			CategoryPattern:     source.validationConf.CategoryPattern,
		}
	}

	if source.categoryPattern != nil {
		l.categoryPattern = source.categoryPattern
	}

	// Validate the copy
	if err := l.Validate(); err != nil {
			return fmt.Errorf("validation failed after copy: %w", err)
	}

	return nil
}


// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (lpd *LibraryPageData) MarshalBinary() ([]byte, error) {
	if lpd == nil {
		return nil, fmt.Errorf("cannot marshal nil LibraryPageData")
	}

	lpd.mu.RLock()
	defer lpd.mu.RUnlock()

	if err := lpd.Validate(); err != nil {
			return nil, fmt.Errorf("validation failed before marshal: %w", err)
	}

	// Create a copy without mutex and logger for marshaling
	dataCopy := struct {
			Books    []repository.Book            `json:"books"`
			Authors  map[string][]repository.Book `json:"authors"`
			Genres   map[string][]repository.Book `json:"genres"`
			Formats  map[string][]repository.Book `json:"formats"`
			Tags     map[string][]repository.Book `json:"tags"`
	}{
			Books:    lpd.Books,
			Authors:  lpd.Authors,
			Genres:   lpd.Genres,
			Formats:  lpd.Formats,
			Tags:     lpd.Tags,
	}

	return json.Marshal(dataCopy)
}

func (lpd *LibraryPageData) UnmarshalBinary(data []byte) error {
	if data == nil {
		return fmt.Errorf("cannot unmarshal nil data")
	}
	lpd.mu.Lock()
	defer lpd.mu.Unlock()

	// Temporary struct for unmarshaling
	var temp struct {
			Books    []repository.Book            `json:"books"`
			Authors  map[string][]repository.Book `json:"authors"`
			Genres   map[string][]repository.Book `json:"genres"`
			Formats  map[string][]repository.Book `json:"formats"`
			Tags     map[string][]repository.Book `json:"tags"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
			return fmt.Errorf("unmarshal failed: %w", err)
	}

	// Update fields
	lpd.Books = temp.Books
	lpd.Authors = temp.Authors
	lpd.Genres = temp.Genres
	lpd.Formats = temp.Formats
	lpd.Tags = temp.Tags

	// Validate after unmarshaling
	if err := lpd.Validate(); err != nil {
			return fmt.Errorf("validation failed after unmarshal: %w", err)
	}

	return nil
}
