package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"log/slog"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/unicode/norm"
	"golang.org/x/text/width"

	"github.com/lokeam/bravo-kilo/internal/books/repository"
	"github.com/lokeam/bravo-kilo/internal/shared/collections"
	"github.com/lokeam/bravo-kilo/internal/shared/transaction"
)

type BookService interface {
	CreateBookEntry(ctx context.Context, book repository.Book, userID int) (int, error)
	CreateEntries(
		ctx context.Context,
		tx *sql.Tx,
		bookID int,
		items []string,
		getIDByName func(ctx context.Context, tx *sql.Tx, name string, id *int) error,
		insertItem func(ctx context.Context, tx *sql.Tx, name string) (int, error),
		associateItem func(ctx context.Context, tx *sql.Tx, bookID, itemID int) error,
	) error
	InsertFormats(ctx context.Context, tx *sql.Tx, bookID int, formats []string) error
	ReverseNormalizeBookData(books *[]repository.Book)
	SanitizeAndUnescape(input string) string
	NormalizeBookData(book *repository.Book)
	SanitizeBookData(book *repository.Book)
}

// BookServiceImpl implements BookService
type BookServiceImpl struct {
	BookUpdaterService  BookUpdaterService
	bookRepository      repository.BookRepository
	authorRepository    repository.AuthorRepository
	genreRepository     repository.GenreRepository
	formatRepository    repository.FormatRepository
	tagRepository       repository.TagRepository
	sanitizer           *bluemonday.Policy
	dbManager           transaction.DBManager
	logger              *slog.Logger
}

// NewBookService creates a new instance of BookService
func NewBookService(
	bookRepo repository.BookRepository,
	authorRepo repository.AuthorRepository,
	genreRepo repository.GenreRepository,
	formatRepo repository.FormatRepository,
	tagRepo repository.TagRepository,
	logger *slog.Logger,
	dbManager transaction.DBManager,
) (BookService, error) {
	if bookRepo == nil {
		return nil, fmt.Errorf("book repository cannot be nil")
	}

	if authorRepo == nil {
		return nil, fmt.Errorf("author repository cannot be nil")
	}

	if genreRepo == nil {
		return nil, fmt.Errorf("genre repository cannot be nil")
	}

	if formatRepo == nil {
		return nil, fmt.Errorf("format repository cannot be nil")
	}

	if tagRepo == nil {
		return nil, fmt.Errorf("tag repository cannot be nil")
	}

	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	if dbManager == nil {
		return nil, fmt.Errorf("transaction manager cannot be nil")
	}

	sanitizer := bluemonday.StrictPolicy()
	if sanitizer == nil {
			return nil, fmt.Errorf("failed to initialize sanitizer")
	}

	return &BookServiceImpl{
		bookRepository:      bookRepo,
		authorRepository:    authorRepo,
		genreRepository:     genreRepo,
		formatRepository:    formatRepo,
		tagRepository:       tagRepo,
		sanitizer:           sanitizer,
		dbManager:           dbManager,
		logger:              logger,
	}, nil
}

// InsertBook creates a new book with its associated authors, genres, and formats
func (s *BookServiceImpl) CreateBookEntry(ctx context.Context, book repository.Book, userID int) (int, error) {
	// Normalize + sanitize book data before proceeding
	s.NormalizeBookData(&book)
	s.SanitizeBookData(&book)

	// Validate required fields
	if book.Title == "" || len(book.Authors) == 0 {
		return 0, errors.New("book title and authors are required")
	}

	// Format publish date if only year is provided
	book.PublishDate = formatPublishDate(book.PublishDate)

	// Marshal JSON tags
	tagsJSON, err := json.Marshal(book.Tags)
	if err != nil {
		s.logger.Error("Error marshaling tags", "error", err)
		return 0, err
	}

	// Start transaction
	tx, err := s.dbManager.BeginTransaction(ctx)
	if err != nil {
		s.logger.Error("Error starting transaction", "error", err)
		return 0, err
	}
	defer tx.Rollback()

	// Insert the book into the books table and associate with the user
	bookID, err := s.bookRepository.InsertBook(ctx, tx, book, userID, tagsJSON)
	if err != nil {
		s.logger.Error("Error inserting book", "error", err)
		return 0, err
	}

	// Create author entries
	err = s.CreateEntries(
		ctx,
		tx,
		bookID,
		book.Authors,
		s.authorRepository.GetAuthorIDByName,
		s.authorRepository.InsertAuthor,
		s.authorRepository.AssociateBookWithAuthor,
	)
	if err != nil {
		s.logger.Error("Error inserting authors", "error", err)
		return 0, err
	}

	// Create genres entries
	err = s.CreateEntries(
		ctx,
		tx,
		bookID,
		book.Genres,
		s.genreRepository.GetGenreIDByName,
		s.genreRepository.InsertGenre,
		s.genreRepository.AssociateBookWithGenre,
	)
	if err != nil {
		s.logger.Error("Error inserting genres", "error", err)
		return 0, err
	}

	// Insert formats
	err = s.InsertFormats(ctx, tx, bookID, book.Formats)
	if err != nil {
		s.logger.Error("Error inserting formats", "error", err)
		return 0, err
	}

	// Commit the transaction
	if err = s.dbManager.CommitTransaction(tx); err != nil {
		s.logger.Error("Error committing transaction", "error", err)
		return 0, err
	}

	return bookID, nil
}

// Higher order helper fn to insert author, genre, tag entries
func (s *BookServiceImpl) CreateEntries(
	ctx context.Context,
	tx *sql.Tx,
	bookID int,
	items []string,
	getIDByName func(ctx context.Context, tx *sql.Tx, name string, id *int) error,
	insertItem func(ctx context.Context, tx *sql.Tx, name string) (int, error),
	associateItem func(ctx context.Context, tx *sql.Tx, bookID, itemID int) error,
) error {
	itemSet := collections.NewSet()
	itemMap := make(map[string] string)

	// Dedupe and normalize for comparison, keep original for insertion
	for _, item := range items {
		normalizedItem := strings.TrimSpace(width.Narrow.String(norm.NFC.String(strings.ToLower(item))))
		if normalizedItem != "" {
			itemSet.Add(normalizedItem)
			itemMap[normalizedItem] = item
		}
	}

	// Look at each normalized entry, check existence and insert or associate
	for _, normalizedItem := range itemSet.Elements() {
		originalItem := itemMap[normalizedItem]
		var itemID int

		// Check item existence
		err := getIDByName(ctx, tx, normalizedItem, &itemID)
		if err == sql.ErrNoRows {
			// Insert original item into db if not found
			itemID, err = insertItem(ctx, tx, originalItem)
			if err != nil {
				return fmt.Errorf("failed to insert item: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("failed to query item: %w", err)
		}

		// Associate item with book
		err = associateItem(ctx, tx, bookID, itemID)
		if err != nil {
			return fmt.Errorf("failed to associate item with book: %w", err)
		}
	}
	return nil
}

// Helper to insert formats
func (s *BookServiceImpl) InsertFormats(ctx context.Context, tx *sql.Tx, bookID int, formats []string) error {
	if len(formats) == 0 {
		return nil
	}

	formatsSet := collections.NewSet()
	for _, format := range formats {
		if format != "" {
			formatsSet.Add(format)
		} else {
			s.logger.Warn("Skipping empty format")
		}
	}

	// Add the formats using AddFormats
	err := s.formatRepository.AddFormats(tx, ctx, bookID, formatsSet.Elements())
	if err != nil {
		s.logger.Error("Error inserting formats", "error", err)
		return err
	}

	return nil
}

func (s *BookServiceImpl) SanitizeAndUnescape(input string) string {
	sanitized := s.sanitizer.Sanitize(input)
	return html.UnescapeString(sanitized)
}

func (s *BookServiceImpl) ReverseNormalizeBookData(books *[]repository.Book) {
	caser := cases.Title(language.Und)

	for i := range *books {
		book := &(*books)[i]

		// Apply sentence case to book fields
		book.Title = caser.String(book.Title)
		book.Subtitle = caser.String(book.Subtitle)

		for j := range book.Genres {
			book.Genres[j] = caser.String(book.Genres[j])
		}

		for j := range book.Formats {
			book.Formats[j] = caser.String(book.Formats[j])
		}

		for j := range book.Tags {
			book.Tags[j] = caser.String(book.Tags[j])
		}
	}
}

// Helper method to normalize book data
func (s *BookServiceImpl) NormalizeBookData(book *repository.Book) {
	caser := cases.Lower(language.Und)
	titleCaser := cases.Title(language.Und)

	// Normalize book title and description
	book.Title = strings.TrimSpace(caser.String(norm.NFC.String(book.Title)))
	book.Subtitle = strings.TrimSpace(norm.NFC.String(book.Subtitle))
	book.Language = strings.TrimSpace(caser.String(norm.NFC.String(book.Language)))

	// Normalize authors (trim only, no lowercase conversion)
	for i := range book.Authors {
		book.Authors[i] = strings.TrimSpace(width.Narrow.String(norm.NFC.String(book.Authors[i])))
	}

	// Normalize genres, formats, tags
	for i := range book.Genres {
		book.Genres[i] = strings.TrimSpace(titleCaser.String(norm.NFC.String(book.Genres[i])))
	}

	for i := range book.Formats {
		book.Formats[i] = strings.TrimSpace(norm.NFC.String(book.Formats[i]))
	}

	for i := range book.Tags {
		book.Tags[i] = strings.TrimSpace(titleCaser.String(norm.NFC.String(book.Tags[i])))
	}
}

func (s *BookServiceImpl) SanitizeBookData(book *repository.Book) {
	// Sanitize book fields
	book.Title = s.SanitizeAndUnescape(book.Title)
	book.Subtitle = s.SanitizeAndUnescape(book.Subtitle)
	book.Language = s.SanitizeAndUnescape(book.Language)
	book.Description = s.SanitizeAndUnescape(book.Description)
	book.Notes = s.SanitizeAndUnescape(book.Notes)

	// Sanitize genres, formats, and tags
	for i := range book.Genres {
		book.Genres[i] = s.SanitizeAndUnescape(book.Genres[i])
	}

	for i := range book.Formats {
		book.Formats[i] = s.SanitizeAndUnescape(book.Formats[i])
	}

	for i := range book.Tags {
		book.Tags[i] = s.SanitizeAndUnescape(book.Tags[i])
	}

	// Authors only need sanitization, no case normalization
	for i := range book.Authors {
		book.Authors[i] = s.SanitizeAndUnescape(book.Authors[i])
	}
}

// Helper to format the publish date
func formatPublishDate(dateStr string) string {
	if len(dateStr) == 4 {
		return dateStr + "-01-01"
	}
	return dateStr
}
