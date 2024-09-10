package services

import (
	"bravo-kilo/internal/data"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
)

type BookService interface {
	InsertBook(ctx context.Context, book data.Book, userID int) (int, error)
}

// BookServiceImpl implements BookService
type BookServiceImpl struct {
	repository data.BookRepository
	logger     *slog.Logger
}


// NewBookService creates a new instance of BookService
func NewBookService(bookRepo data.BookRepository, authorRepo data.AuthorRepository, genreManager data.GenreManager, logger *slog.Logger) *BookService {
    return &BookService{
        BookRepo:     bookRepo,
        AuthorRepo:   authorRepo,
        GenreManager: genreManager,
        Logger:       logger,
    }
}

func (s *BookServiceImpl) InsertBook(ctx context.Context, book data.Book, userID int) (int, error) {
	// Validate required fields
	if book.Title == "" || len(book.Authors) == 0 {
		return 0, errors.New("book title and authors are required")
	}

	// Format publish date if only year is provided
	book.PublishDate = formatPublishDate(book.PublishDate)

	// JSON encode tags
	tagsJSON, err := json.Marshal(book.Tags)
	if err != nil {
		s.logger.Error("Error marshaling tags", "error", err)
		return 0, err
	}

	// Start transaction
	tx, err := s.repository.BeginTransaction(ctx)
	if err != nil {
		s.logger.Error("Error starting transaction", "error", err)
		return 0, err
	}
	defer tx.Rollback()

	// Insert the book into the books table
	bookID, err := s.repository.InsertBook(ctx, tx, book, tagsJSON)  // The InsertBook method will now be a simple DB insert call.
	if err != nil {
		s.logger.Error("Error inserting book", "error", err)
		return 0, err
	}

	// Insert authors
	err = s.insertAuthors(ctx, tx, bookID, book.Authors)
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
	if err = tx.Commit(); err != nil {
		s.logger.Error("Error committing transaction", "error", err)
		return 0, err
	}

	return bookID, nil
}


// Helper to insert authors
func (s *BookServiceImpl) insertAuthors(ctx context.Context, tx data.Transaction, bookID int, authors []string) error {
	for _, author := range authors {
			err := s.repository.AddAuthor(ctx, tx, bookID, author)
			if err != nil {
					return err
			}
	}
	return nil
}

// Helper to insert genres
func (s *BookServiceImpl) insertGenres(ctx context.Context, tx data.Transaction, bookID int, genres []string) error {
	for _, genre := range genres {
			genreID, err := s.repository.AddOrGetGenreID(ctx, tx, genre)
			if err != nil {
					return err
			}
			err = s.repository.AddGenre(ctx, tx, bookID, genreID)
			if err != nil {
					return err
			}
	}
	return nil
}

// Helper to insert formats
func (s *BookServiceImpl) insertFormats(ctx context.Context, tx data.Transaction, bookID int, formats []string) error {
	for _, format := range formats {
			formatID, err := s.repository.AddOrGetFormatID(ctx, tx, format)
			if err != nil {
					return err
			}
			err = s.repository.AddFormat(ctx, tx, bookID, formatID)
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