package validator

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/lokeam/bravo-kilo/internal/books/repository"
	"github.com/lokeam/bravo-kilo/internal/shared/types"
)

const defaultValidationTimeout = 30 * time.Second

type BookValidator struct {
	logger       *slog.Logger
	metrics      *ValidationMetrics

	// Compile regex patterns at initialization
	regexPatterns struct {
		title      *regexp.Regexp
		format     *regexp.Regexp
		author     *regexp.Regexp
		publisher  *regexp.Regexp
		language   *regexp.Regexp
		isbn10     *regexp.Regexp
		isbn13     *regexp.Regexp
	}

	constraints struct {
		titleMaxLength          int
		descriptionMaxLength    int
		minPageCount            int
		maxPageCount            int
		maxAuthors              int
		maxGenres               int
		maxTags                 int
	}

	// Channel to signal cleanup
	done chan struct{}
}

// BookValidator Constructor
func NewBookValidator(logger *slog.Logger) (*BookValidator, error) {
	v := &BookValidator{
		logger: logger,
		metrics: &ValidationMetrics{
			ErrorTypes: make(map[string]int64),
			ValidationErrors: make(map[ValidationErrorCode]int64),
			LastValidationTime: time.Now(),
		},
		done: make(chan struct{}),
}

// Initialize constraints from your frontend validation
v.constraints = struct {
		titleMaxLength          int
		descriptionMaxLength    int
		minPageCount            int
		maxPageCount            int
		maxAuthors              int
		maxGenres               int
		maxTags                 int
}{
		titleMaxLength:         500,    // Matches your frontend validation
		descriptionMaxLength:   50000,  // Adjust based on your needs
		minPageCount:           1,
		maxPageCount:           100000,
		maxAuthors:             10,      // Reasonable limits
		maxGenres:              20,
		maxTags:                30,
}

	if err := v.compilePatterns(); err != nil {
		return nil, fmt.Errorf("error compiling regex patterns: %w", err)
	}

	return v, nil
}

// Method - Validate Redis Data
func (bv *BookValidator) ValidateRedisData(ctx context.Context, data *types.LibraryPageData) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	startTime := time.Now()
	defer func() {
			duration := time.Since(startTime)
			bv.metrics.AddTime(duration)
	}()

	if data == nil {
		return fmt.Errorf("received redis validator data cannot be nil")
	}

	// Guard cluase
	if len(data.Books) == 0 {
		return nil
	}

	return bv.BatchValidateBooks(ctx, data.Books)
}

func (bv *BookValidator) BatchValidateBooks(ctx context.Context, books []repository.Book) error {
	ctx, cancel := context.WithTimeout(ctx, defaultValidationTimeout)
	defer cancel()
	// Don't use concurrency for small batches
	if len(books) < 10 {
		for _, book := range books {
			if err := bv.validateBook(book); err != nil {
				if validationErr, ok := err.(ValidationError); ok {
					bv.metrics.IncrementErrorType(ValidationErrorCode(validationErr.Code))
				} else {
					bv.metrics.IncrementErrorType(ErrInvalidContent)
				}
				return fmt.Errorf("book %d validation failed: %w", book.ID, err)
			}
			bv.metrics.IncrementValid()
		}
		return nil
	}

	// For large batches, use worker pool
	errChan := make(chan error, len(books))
	semaphore := make(chan struct{}, 5) // Limit to 5 goroutines

	var wg sync.WaitGroup
	for _, book := range books {
		wg.Add(1)
		go func(b repository.Book) {
				defer wg.Done()

				select {
				case semaphore <- struct{}{}: // Acquire semaphore
						defer func() { <-semaphore }() // Release semaphore
				case <-ctx.Done():
						errChan <- ctx.Err()
						return
				}

				if err := bv.validateBook(b); err != nil {
						bv.metrics.IncrementErrorType(ErrRequired)
						errChan <- fmt.Errorf("book %d validation failed: %w", b.ID, err)
						return
				}
				bv.metrics.IncrementValid()
		}(book)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}


// Cleanup performs necessary cleanup operations for the BookValidator.
// Called when the validator is no longer needed.
func (bv *BookValidator) Cleanup() {
	select {
	case <-bv.done:
			return // Already closed
	default:
			bv.metrics = NewValidationMetrics()
			close(bv.done)
	}
}

// Helper Methods


// initialize regex patterns
func (bv *BookValidator) compilePatterns() error {
	var err error

	// Title: alphanumeric w/ basic punctuation
	bv.regexPatterns.title, err = regexp.Compile(`^[\w\s\-\.,:;'"!?()]{1,500}$`)
	if err != nil {
		return fmt.Errorf("failed to compile title pattern: %w", err)
	}

	// Format: exact match for book formats
	bv.regexPatterns.format, err = regexp.Compile(`^(physical|eBook|audioBook)$`)
	if err != nil {
		return fmt.Errorf("failed to compile format pattern: %w", err)
	}

	// Language: ISO 639-1 language codes
	bv.regexPatterns.language, err = regexp.Compile(`^[a-z]{2}(-[A-Z]{2})?$`)
	if err != nil {
		return fmt.Errorf("failed to compile language pattern: %w", err)
	}

	// ISBN-10
	bv.regexPatterns.isbn10, err = regexp.Compile(`^\d{9}[\d|X]$`)
	if err != nil {
		return fmt.Errorf("failed to compile ISBN-10 pattern: %w", err)
	}

	// ISBN-13
	bv.regexPatterns.isbn13, err = regexp.Compile(`^\d{13}$`)
	if err != nil {
		return fmt.Errorf("failed to compile ISBN-13 pattern: %w", err)
	}

	return nil
}

func (bv *BookValidator) validateRequiredFields(book repository.Book) error {
	if len(strings.TrimSpace(book.Title)) == 0 {
		bv.metrics.IncrementErrorType(ErrRequired)
		return NewValidationError("title", ErrRequired, "title is required")
	}

	// Handle RichText type for description
	if book.Description.IsRichTextEmpty() {
		bv.metrics.IncrementErrorType(ErrRequired)
		return NewValidationError("description", ErrRequired, "description is required")
	}

	if len(strings.TrimSpace(book.Language)) == 0 {
		bv.metrics.IncrementErrorType(ErrRequired)
		return NewValidationError("language", ErrRequired, "language is required")
	}

	// Validate field lengths
	if len(book.Title) > bv.constraints.titleMaxLength {
		bv.metrics.IncrementErrorType(ErrMaxExceeded)
		return NewValidationError("title", ErrMaxExceeded, "title exceeds maximum length")
	}

	// Validate RichText length
	if book.Description.CheckRichTextLength() > bv.constraints.descriptionMaxLength {
    bv.metrics.IncrementErrorType(ErrMaxExceeded)
    return NewValidationError("description", ErrMaxExceeded, "description exceeds maximum length")
}

	return nil
}


func (bv *BookValidator) validateBook(book repository.Book) error {
	// Start collecting metrics
	startTime := time.Now()
	defer func() {
			bv.metrics.AddTime(time.Since(startTime))
	}()

	var err error
	// Step 1: Validate required fields
	if err = bv.validateRequiredFields(book); err != nil {
			return err // Error type already incremented in validateRequiredFields
	}

	// Step 2: Validate content format
	if err = bv.validateContentFormat(book); err != nil {
			return err // Error type already incremented in validateContentFormat
	}

	// Step 3: Validate array fields
	if err = bv.validateArrayFields(book); err != nil {
			return err // Error type already incremented in validateArrayFields
	}

	if err == nil {
			bv.metrics.IncrementValid()
	}
	return nil
}

func (bv *BookValidator) validateContentFormat(book repository.Book) error {
    // ISBN validation
    if book.ISBN10 != "" && !bv.regexPatterns.isbn10.MatchString(book.ISBN10) {
			bv.metrics.IncrementErrorType(ErrInvalidISBN)
			return NewValidationError("isbn10", ErrInvalidISBN,
					"ISBN-10 must be 10 digits with possible 'X' at the end")
	}

	// Language validation
	if !bv.regexPatterns.language.MatchString(book.Language) {
			bv.metrics.IncrementErrorType(ErrInvalidLanguage)
			return NewValidationError("language", ErrInvalidLanguage,
					"Language must be a valid ISO 639-1 code")
	}

	// Title validation
	if !bv.regexPatterns.title.MatchString(book.Title) {
			bv.metrics.IncrementErrorType(ErrInvalidTitle)
			return NewValidationError("title", ErrInvalidTitle,
					"Title contains invalid characters or exceeds length limit")
	}

	return nil
}

func (bv *BookValidator) validateArrayFields(book repository.Book) error {
	// Authors validation
	if len(book.Authors) == 0 {
			bv.metrics.IncrementErrorType(ErrRequired)
			return NewValidationError("authors", ErrRequired, "at least one author is required")
	}
	if len(book.Authors) > bv.constraints.maxAuthors {
			bv.metrics.IncrementErrorType(ErrMaxExceeded)
			return NewValidationError("authors", ErrMaxExceeded,
					fmt.Sprintf("exceeds maximum number of authors (%d)", bv.constraints.maxAuthors))
	}

	// Genres validation
	if len(book.Genres) > bv.constraints.maxGenres {
			bv.metrics.IncrementErrorType(ErrMaxExceeded)
			return NewValidationError("genres", ErrMaxExceeded,
					fmt.Sprintf("exceeds maximum number of genres (%d)", bv.constraints.maxGenres))
	}

	// Tags validation
	if len(book.Tags) > bv.constraints.maxTags {
			bv.metrics.IncrementErrorType(ErrMaxExceeded)
			return NewValidationError("tags", ErrMaxExceeded,
					fmt.Sprintf("exceeds maximum number of tags (%d)", bv.constraints.maxTags))
	}

	// Formats validation
	if len(book.Formats) == 0 {
			bv.metrics.IncrementErrorType(ErrRequired)
			return NewValidationError("formats", ErrRequired, "at least one format is required")
	}

	for _, bookFormat := range book.Formats {
			if !bv.regexPatterns.format.MatchString(string(bookFormat)) {
					bv.metrics.IncrementErrorType(ErrInvalidFormat)
					return NewValidationError("format", ErrInvalidFormat,
							fmt.Sprintf("invalid format: %s", bookFormat))
			}
	}

	return nil
}
