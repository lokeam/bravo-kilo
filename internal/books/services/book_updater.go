package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/lokeam/bravo-kilo/internal/books/repository"
	"github.com/lokeam/bravo-kilo/internal/dbconfig"
	"github.com/lokeam/bravo-kilo/internal/shared/collections"
	"github.com/lokeam/bravo-kilo/internal/shared/utils"
)

type BookUpdaterService interface {
	UpdateBookEntry(tx *sql.Tx, book repository.Book) error
	UpdateFormats(ctx context.Context, bookID int, newFormats []string) error
}

type BookUpdaterServiceImpl struct {
	DB         *sql.DB
	logger     *slog.Logger
	bookRepo   repository.BookRepository
	bookCache  repository.BookCache
	formatRepo repository.FormatRepository
	genreRepo  repository.GenreRepository
}

func NewBookUpdaterService(
	db *sql.DB,
	logger *slog.Logger,
	bookRepo repository.BookRepository,
	bookCache repository.BookCache,
	formatRepo repository.FormatRepository,
	genreRepo repository.GenreRepository,
	) (BookUpdaterService, error) {
	if db == nil || logger == nil {
		return nil, fmt.Errorf("book updater, database or logger is nil")
	}

	if formatRepo == nil {
		return nil, fmt.Errorf("book updater, error initializing format repo")
	}

	if genreRepo == nil {
		return nil, fmt.Errorf("book updater, error initializing genre repo")
	}

	return &BookUpdaterServiceImpl{
		DB:          db,
		logger:      logger,
		formatRepo:  formatRepo,
		genreRepo:   genreRepo,
		bookRepo:    bookRepo,
		bookCache:   bookCache,
	}, nil
}

func (b *BookUpdaterServiceImpl) UpdateBookEntry(tx *sql.Tx, book repository.Book) error {
	// Invalidate caches
	b.bookCache.InvalidateCaches(book.ID)
	b.logger.Info("Cache invalidated for book", "book", book.ID)

	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	tagsJSON, err := json.Marshal(book.Tags)
	if err != nil {
			b.logger.Error("Error marshalling tags to JSON", "error", err)
			return err
	}

	statement := `UPDATE books SET title=$1, subtitle=$2, description=$3, language=$4, page_count=$5,
								 publish_date=$6, image_link=$7, notes=$8, tags=$9, last_updated=$10,
								 isbn_10=$11, isbn_13=$12 WHERE id=$13`
	_, err = tx.ExecContext(ctx, statement,
			book.Title,
			book.Subtitle,
			book.Description,
			book.Language,
			book.PageCount,
			book.PublishDate,
			book.ImageLink,
			book.Notes,
			tagsJSON,
			time.Now(),
			book.ISBN10,
			book.ISBN13,
			book.ID,
	)
	if err != nil {
			b.logger.Error("Book Model - Error updating book", "error", err)
			return err
	}

	// Update genres
	if err := b.updateGenres(ctx, book.ID, book.Genres); err != nil {
			return err
	}

	// Update formats
	if err := b.UpdateFormats(ctx, book.ID, book.Formats); err != nil {
			return err
	}

	// Update authors
	if err := b.updateAuthors(ctx, book.ID, book.Authors); err != nil {
			b.logger.Error("Book Model - Error updating authors for book", "error", err)
			return err
	}

	return nil
}

func (b *BookUpdaterServiceImpl) UpdateFormats(ctx context.Context, bookID int, formats []string) error {
	// Remove duplicates before processing
	formats = utils.RemoveDuplicates(formats)

	// Fetch existing formats
	currentFormats, err := b.formatRepo.GetFormats(ctx, bookID)
	if err != nil {
		b.logger.Error("Error fetching current formats", "error", err)
		return err
	}

	// Determine formats to add (difference between new formats and current formats)
	formatsToAdd := utils.FindDifference(formats, currentFormats)

	// Begin transaction
	tx, err := b.DB.BeginTx(ctx, nil)
	if err != nil {
		b.logger.Error("Error starting transaction", "error", err)
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			b.logger.Error("Transaction rolled back", "error", err)
		} else {
			tx.Commit()
			b.logger.Info("Transaction committed successfully")
		}
	}()

	// Add new formats to the book using the existing AddFormats method
	if len(formatsToAdd) > 0 {
		err = b.formatRepo.AddFormats(tx, ctx, bookID, formatsToAdd)
		if err != nil {
			b.logger.Error("Error adding formats", "error", err)
			return err
		}
	}

	return nil
}

// Helper function for UpdateBooks
func (b *BookUpdaterServiceImpl) updateAuthors(ctx context.Context, bookID int, authors []string) error {
	// Delete existing authors for the book
	deleteStatement := `DELETE FROM book_authors WHERE book_id = $1`
	if _, err := b.DB.ExecContext(ctx, deleteStatement, bookID); err != nil {
		b.logger.Error("Error deleting existing authors for book", "error", err)
		return err
	}

	// Insert new authors without duplicates
	authorSet := collections.NewSet() // Ensure no duplicates are reinserted
	for _, authorName := range authors {
		if authorSet.Has(authorName) {
			continue // Skip if already processed
		}
		authorSet.Add(authorName)

		var authorID int
		selectStatement := `SELECT id FROM authors WHERE name = $1`
		err := b.DB.QueryRowContext(ctx, selectStatement, authorName).Scan(&authorID)

		if err != nil && err == sql.ErrNoRows {
			// Insert author if not exists
			insertAuthorStatement := `INSERT INTO authors (name) VALUES ($1) RETURNING id`
			err := b.DB.QueryRowContext(ctx, insertAuthorStatement, authorName).Scan(&authorID)
			if err != nil {
				b.logger.Error("Error inserting new author", "error", err)
				return err
			}
		} else if err != nil {
			b.logger.Error("Error checking for existing author", "error", err)
			return err
		}

		// Link author to book without duplicates
		insertLinkStatement := `INSERT INTO book_authors (book_id, author_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
		_, err = b.DB.ExecContext(ctx, insertLinkStatement, bookID, authorID)
		if err != nil {
			b.logger.Error("Error linking author to book", "error", err)
			return err
		}
	}
	return nil
}

func (b *BookUpdaterServiceImpl) updateGenres(ctx context.Context, bookID int, genres []string) error {
	// Remove duplicates
	genres = utils.RemoveDuplicates(genres)

	// Fetch existing genres
	currentGenres, err := b.genreRepo.GetGenres(ctx, bookID)
	if err != nil {
		b.logger.Error("Error fetching current genres", "error", err)
		return err
	}

	genresToAdd := utils.FindDifference(genres, currentGenres)

	// Insert new genres
	genreSet := collections.NewSet() // Ensure no duplicates are reinserted
	for _, genre := range genresToAdd {
		if genreSet.Has(genre) {
			continue // Skip if already processed
		}
		genreSet.Add(genre)

		genreID, err := b.genreRepo.AddOrGetGenreID(ctx, genre)
		if err != nil {
			b.logger.Error("Error getting genre ID", "error", err)
			return err
		}

		// Link genre to book
		err = b.genreRepo.AddGenre(ctx, bookID, genreID)
		if err != nil {
			b.logger.Error("Error adding genre association", "error", err)
			return err
		}
	}
	return nil
}
