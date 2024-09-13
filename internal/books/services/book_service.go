package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

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

	return &BookServiceImpl{
		bookRepository:   bookRepo,
		authorRepository: authorRepo,
		genreRepository:  genreRepo,
		formatRepository: formatRepo,
		tagRepository:    tagRepo,
		dbManager:        dbManager,
		logger:           logger,
	}, nil
}

// InsertBook creates a new book with its associated authors, genres, and formats
func (s *BookServiceImpl) CreateBookEntry(ctx context.Context, book repository.Book, userID int) (int, error) {
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

	// Utilize custom Set to store unique author names
	authorsSet := collections.NewSet()
	for _, author := range authors {
		authorsSet.Add(author)
	}

	for _, author := range authorsSet.Elements() {
		s.logger.Info("Inserting/Querying author", "author", author)

		var authorID int
		err := s.authorRepository.GetAuthorIDByName(ctx, tx, author, &authorID)
		if err != nil {
			if err == sql.ErrNoRows {

				s.logger.Info("Author not found, inserting new author", "author", author)
				authorID, err = s.authorRepository.InsertAuthor(ctx, tx, author)
				if err != nil {
					s.logger.Error("Error inserting author", "error", err)
					return err
				}
			} else {
				s.logger.Error("Error querying author", "error", err)
				return err
			}
		}

		// Associate the book with the author
		err = s.authorRepository.AssociateBookWithAuthor(ctx, tx, bookID, authorID)
		if err != nil {
			s.logger.Error("Error adding author association", "error", err)
			return err
		}

		s.logger.Info("Successfully associated book with author", "bookID", bookID, "authorID", authorID)
	}
	return nil
}

// Helper to insert genres
func (s *BookServiceImpl) createGenreEntries(ctx context.Context, tx *sql.Tx, bookID int, genres []string) error {

	// Utilize custom Set to store unique author names
	genresSet := collections.NewSet()
	for _, genre := range genres {
		genresSet.Add(genre)
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
	for _, format := range formats {
		formatID, err := s.formatRepository.AddOrGetFormatID(ctx, tx, format)
		if err != nil {
			return err
		}
		err = s.formatRepository.AddFormat(ctx, tx, bookID, formatID)
		if err != nil {
			return err
		}
	}
	return nil
}

// Helper to format the publish date
func formatPublishDate(dateStr string) string {
	if len(dateStr) == 4 {
		return dateStr + "-01-01"
	}
	return dateStr
}
