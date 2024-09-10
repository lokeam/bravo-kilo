package services

import (
	"bravo-kilo/internal/data"
	"bravo-kilo/internal/data/collections"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
)

type BookService interface {
	CreateBookEntry(ctx context.Context, book data.Book, userID int) (int, error)
}

// BookServiceImpl implements BookService
type BookServiceImpl struct {
	bookRepository   data.BookRepository
	authorRepository data.AuthorRepository
	genreRepository  data.GenreRepository
	formatRepository data.FormatRepository
	dbManager        data.DBManager
	logger           *slog.Logger
}

// NewBookService creates a new instance of BookService
func NewBookService(
	bookRepo data.BookRepository,
	authorRepo data.AuthorRepository,
	genreRepo data.GenreRepository,
	formatRepo data.FormatRepository,
	logger *slog.Logger,
	dbManager data.DBManager,
) BookService {
	return &BookServiceImpl{
		bookRepository:   bookRepo,
		authorRepository: authorRepo,
		genreRepository:  genreRepo,
		formatRepository: formatRepo,
		dbManager:        dbManager,
		logger:           logger,
	}
}

// InsertBook creates a new book with its associated authors, genres, and formats
func (s *BookServiceImpl) CreateBookEntry(ctx context.Context, book data.Book, userID int) (int, error) {
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

	// Insert genres
	err = s.insertGenres(ctx, tx, bookID, book.Genres)
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
func (s *BookServiceImpl) insertGenres(ctx context.Context, tx *sql.Tx, bookID int, genres []string) error {
	for _, genre := range genres {
		genreID, err := s.genreRepository.AddOrGetGenreID(ctx, tx, genre)
		if err != nil {
			return err
		}
		err = s.genreRepository.AddGenre(ctx, tx, bookID, genreID)
		if err != nil {
			return err
		}
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
