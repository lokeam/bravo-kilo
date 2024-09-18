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
}

// BookServiceImpl implements BookService
type BookServiceImpl struct {
	bookRepository   repository.BookRepository
	authorRepository repository.AuthorRepository
	genreRepository  repository.GenreRepository
	formatRepository repository.FormatRepository
	tagRepository    repository.TagRepository
	sanitizer        *bluemonday.Policy
	dbManager        transaction.DBManager
	logger           *slog.Logger
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
		bookRepository:   bookRepo,
		authorRepository: authorRepo,
		genreRepository:  genreRepo,
		formatRepository: formatRepo,
		tagRepository:    tagRepo,
		sanitizer:        sanitizer,
		dbManager:        dbManager,
		logger:           logger,
	}, nil
}

// InsertBook creates a new book with its associated authors, genres, and formats
func (s *BookServiceImpl) CreateBookEntry(ctx context.Context, book repository.Book, userID int) (int, error) {
	// Normalize + sanitize book data before proceeding
	s.normalizeBookData(&book)
	s.sanitizeBookData(&book)

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
	err = s.createAuthorEntries(ctx, tx, bookID, book.Authors)
	if err != nil {
		s.logger.Error("Error inserting authors", "error", err)
		return 0, err
	}

	// Create genres
	err = s.createGenreEntries(ctx, tx, bookID, book.Genres)
	if err != nil {
		s.logger.Error("Error inserting genres", "error", err)
		return 0, err
	}

	// Insert formats
	err = s.insertFormats(ctx, tx, bookID, book.Formats)
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

// Helper to insert authors
func (s *BookServiceImpl) createAuthorEntries(ctx context.Context, tx *sql.Tx, bookID int, authors []string) error {
	authorsSet := collections.NewSet()

	// Deduplicate authors using the set
	for _, author := range authors {
		if author != "" {
			authorsSet.Add(author)
		}
	}

	for _, author := range authorsSet.Elements() {
		var authorID int
		// Query the author once
		err := s.authorRepository.GetAuthorIDByName(ctx, tx, author, &authorID)
		if err == sql.ErrNoRows {
			authorID, err = s.authorRepository.InsertAuthor(ctx, tx, author)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		// Associate author with the book
		err = s.authorRepository.AssociateBookWithAuthor(ctx, tx, bookID, authorID)
		if err != nil {
			return err
		}
	}

	return nil
}

// Helper to insert genres
func (s *BookServiceImpl) createGenreEntries(ctx context.Context, tx *sql.Tx, bookID int, genres []string) error {
	genresSet := collections.NewSet()

	for _, genre := range genres {
		if genre != "" {
			genresSet.Add(genre)
		} else {
			s.logger.Warn("Skipping empty genre")
		}
	}

	for _, genre := range genresSet.Elements() {
		var genreID int
		err := s.genreRepository.GetGenreIDByName(ctx, tx, genre, &genreID)
		if err != nil {
			if err == sql.ErrNoRows {
				s.logger.Info("Genre not found, inserting new genre", "genre", genre)
				genreID, err = s.genreRepository.InsertGenre(ctx, tx, genre)
				if err != nil {
					s.logger.Error("Error inserting genre", "error", err)
					return err
				}
			} else {
				s.logger.Error("Error querying genre", "error", err)
				return err
			}
		}

		// Validate genreID
		if genreID == 0 {
			s.logger.Error("Invalid genreID found", "genre", genre)
			return errors.New("invalid genreID")
		}

		err = s.genreRepository.AssociateBookWithGenre(ctx, tx, bookID, genreID)
		if err != nil {
			s.logger.Error("Error adding genre association", "error", err)
			return err
		}

		s.logger.Info("Successfully associated book with genre", "bookID", bookID, "genreID", genreID)
	}

	return nil
}

// Helper to insert formats
func (s *BookServiceImpl) insertFormats(ctx context.Context, tx *sql.Tx, bookID int, formats []string) error {
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

func (s *BookServiceImpl) sanitizeAndUnescape(input string) string {
	sanitized := s.sanitizer.Sanitize(input)
	return html.UnescapeString(sanitized)
}

// Helper method to normalize book data
func (s *BookServiceImpl) normalizeBookData(book *repository.Book) {
	caser := cases.Lower(language.Und)

	// Normalize book title and description
	book.Title = strings.TrimSpace(caser.String(norm.NFC.String(book.Title)))
	book.Subtitle = strings.TrimSpace(norm.NFC.String(book.Subtitle))
	book.Description = strings.TrimSpace(caser.String(norm.NFC.String(book.Description)))
	book.Language = strings.TrimSpace(caser.String(norm.NFC.String(book.Language)))

	// Normalize authors (trim only, no lowercase conversion)
	for i := range book.Authors {
		book.Authors[i] = strings.TrimSpace(width.Narrow.String(norm.NFC.String(book.Authors[i])))
	}

	// Normalize genres, formats, tags
	for i := range book.Genres {
		book.Genres[i] = strings.TrimSpace(caser.String(norm.NFC.String(book.Genres[i])))
	}

	for i := range book.Formats {
		book.Formats[i] = strings.TrimSpace(caser.String(norm.NFC.String(book.Formats[i])))
	}

	for i := range book.Tags {
		book.Tags[i] = strings.TrimSpace(caser.String(norm.NFC.String(book.Tags[i])))
	}

}

func (s *BookServiceImpl) sanitizeBookData(book *repository.Book) {
	// Sanitize book fields
	book.Title = s.sanitizeAndUnescape(book.Title)
	book.Subtitle = s.sanitizeAndUnescape(book.Subtitle)
	book.Description = s.sanitizeAndUnescape(book.Description)
	book.Language = s.sanitizeAndUnescape(book.Language)

	// Sanitize genres, formats, and tags
	for i := range book.Genres {
		book.Genres[i] = s.sanitizeAndUnescape(book.Genres[i])
	}

	for i := range book.Formats {
		book.Formats[i] = s.sanitizeAndUnescape(book.Formats[i])
	}

	for i := range book.Tags {
		book.Tags[i] = s.sanitizeAndUnescape(book.Tags[i])
	}

	// Authors only need sanitization, no case normalization
	for i := range book.Authors {
		book.Authors[i] = s.sanitizeAndUnescape(book.Authors[i])
	}
}

// Helper to format the publish date
func formatPublishDate(dateStr string) string {
	if len(dateStr) == 4 {
		return dateStr + "-01-01"
	}
	return dateStr
}
