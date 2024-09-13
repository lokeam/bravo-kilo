package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/lib/pq"
	"github.com/lokeam/bravo-kilo/internal/shared/utils"
)

type BookUpdater interface {
	UpdateBook(book Book) error
	AddGenre(ctx context.Context, bookID, genreID int) error
	RemoveSpecificFormats(ctx context.Context, bookID int, formats []string) error
	RemoveSpecificGenres(ctx context.Context, bookID int, genres []string) error
}

type BookUpdaterImpl struct {
	DB      *sql.DB
	Logger  *slog.Logger
}

func NewBookUpdater(db *sql.DB, logger *slog.Logger) (BookUpdater, error) {
	if db == nil || logger == nil {
		return nil, fmt.Errorf("book updater, database or logger is nil")
	}

	return &BookUpdaterImpl{
		DB:          db,
		Logger:      logger,
	}, nil
}

func (b *BookUpdaterImpl) UpdateBook(book Book) error {
	// Invalidate caches
	isbn10Cache.Delete(book.ID)
	isbn13Cache.Delete(book.ID)
	titleCache.Delete(book.ID)
	formatsCache.Delete(book.ID)
	genresCache.Delete(book.ID)
	b.Logger.Info("Cache invalidated for book", "book", book.ID)

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	tagsJSON, err := json.Marshal(book.Tags)
	if err != nil {
		b.Logger.Error("Error marshalling tags to JSON", "error", err)
		return err
	}

	statement := `UPDATE books SET title=$1, subtitle=$2, description=$3, language=$4, page_count=$5, publish_date=$6, image_link=$7, notes=$8, tags=$9, last_updated=$10, isbn_10=$11, isbn_13=$12 WHERE id=$13`
	_, err = b.DB.ExecContext(ctx, statement,
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
		b.Logger.Error("Book Model - Error updating book", "error", err)
		return err
	}

	// Update genres
	if err := b.updateGenres(ctx, book.ID, book.Genres); err != nil {
		return err
	}

	// Update formats
	if err := b.updateFormats(ctx, book.ID, book.Formats); err != nil {
		return err
	}

	// Update authors
	if err := b.updateAuthors(ctx, book.ID, book.Authors); err != nil {
		b.Logger.Error("Book Model - Error updating authors for book", "error", err)
		return err
	}

	return nil
}

func (b *BookUpdaterImpl) RemoveSpecificFormats(ctx context.Context, bookID int, formats []string) error {
	statement := `
		DELETE FROM book_formats
		WHERE book_id = $1
		AND format_id IN (
			SELECT id FROM formats WHERE format_type = ANY($2)
		)`

	_, err := b.DB.ExecContext(ctx, statement, bookID, pq.Array(formats))
	if err != nil {
		b.Logger.Error("Error removing specific formats", "error", err)
		return err
	}

	return nil
}

// Helper function for UpdateBooks
func (b *BookUpdaterImpl) updateAuthors(ctx context.Context, bookID int, authors []string) error {
	// Delete existing authors for the book
	deleteStatement := `DELETE FROM book_authors WHERE book_id = $1`
	if _, err := b.DB.ExecContext(ctx, deleteStatement, bookID); err != nil {
		b.Logger.Error("Error deleting existing authors for book", "error", err)
		return err
	}

	// Insert new authors for the book
	for _, authorName := range authors {
		// Check if author already exists in authors table
		var authorID int
		selectStatement := `SELECT id FROM authors WHERE name = $1`
		err := b.DB.QueryRowContext(ctx, selectStatement, authorName).Scan(&authorID)
		if err != nil {
			if err == sql.ErrNoRows {
				// Insert author if doesn't exist
				insertAuthorStatement := `INSERT INTO authors (name) VALUES ($1) RETURNING id`
				err := b.DB.QueryRowContext(ctx, insertAuthorStatement, authorName).Scan(&authorID)
				if err != nil {
					b.Logger.Error("Error inserting new author", "error", err)
					return err
				}
			} else {
				// Something else broke
				b.Logger.Error("Error checking for existing author", "error", err)
				return err
			}
		}

		// Link author to book
		insertLinkStatement := `INSERT INTO book_authors (book_id, author_id) VALUES ($1, $2)`
		_, err = b.DB.ExecContext(ctx, insertLinkStatement, bookID, authorID)
		if err != nil {
			b.Logger.Error("Error linking author to book", "error", err)
			return err
		}
	}

	return nil
}

// Helper fn for UpdateBook
func (b *BookUpdaterImpl) updateFormats(ctx context.Context, bookID int, newFormats []string) error {
	// Invalidate cache for bookID
	formatsCache.Delete(bookID)
	b.Logger.Info("Invalidating formats cache for book", "bookID", bookID)

	// Fetch current formats for the book, passing context
	currentFormats, err := b.GetFormats(ctx, bookID)
	if err != nil {
		b.Logger.Error("Book Model - Error fetching current formats", "error", err)
		return err
	}

	// Find formats to remove and formats to add
	formatsToRemove := utils.FindDifference(currentFormats, newFormats)
	formatsToAdd := utils.FindDifference(newFormats, currentFormats)

	// Remove specific formats
	if len(formatsToRemove) > 0 {
		err := b.RemoveSpecificFormats(ctx, bookID, formatsToRemove)
		if err != nil {
			b.Logger.Error("Book Model - Error removing specific formats", "error", err)
			return err
		}
	}

	// Add new formats
	var formatIDs []int
	for _, format := range formatsToAdd {
		formatID, err := b.addOrGetFormatID(ctx, format)
		if err != nil {
			b.Logger.Error("Error getting format ID", "error", err)
			return err
		}
		formatIDs = append(formatIDs, formatID)
	}

	if len(formatIDs) > 0 {
		err = b.AddFormats(ctx, bookID, formatIDs)
		if err != nil {
			b.Logger.Error("Error adding format associations", "error", err)
			return err
		}
	}

	return nil
}

// Helper fn for UpdateBook
func (b *BookUpdaterImpl) updateGenres(ctx context.Context, bookID int, newGenres []string) error {
	genresCache.Delete(bookID)
	b.Logger.Info("Invalidating genres cache for book", "bookID", bookID)

	// Fetch current genres for the book with context
	currentGenres, err := b.GetGenres(ctx, bookID)
	if err != nil {
		b.Logger.Error("Book Model - Error fetching current genres", "error", err)
		return err
	}

	// Find genres to remove and genres to add
	genresToRemove := utils.FindDifference(currentGenres, newGenres)
	genresToAdd := utils.FindDifference(newGenres, currentGenres)

	// Remove specific genres
	if len(genresToRemove) > 0 {
		err := b.RemoveSpecificGenres(ctx, bookID, genresToRemove) // Pass ctx to this method
		if err != nil {
			b.Logger.Error("Book Model - Error removing specific genres", "error", err)
			return err
		}
	}

	// Add new genres
	for _, genre := range genresToAdd {
		genreID, err := b.addOrGetGenreID(ctx, genre) // Use the updated method with ctx
		if err != nil {
			b.Logger.Error("Error getting genre ID", "error", err)
			return err
		}
		err = b.AddGenre(ctx, bookID, genreID) // Use the refactored AddGenre method
		if err != nil {
			b.Logger.Error("Error adding genre association", "error", err)
			return err
		}
	}

	return nil
}

func (b *BookUpdaterImpl) addOrGetFormatID(ctx context.Context, format string) (int, error) {
	var formatID int
	statement := `
		INSERT INTO formats (format_type)
		VALUES ($1)
		ON CONFLICT (format_type) DO UPDATE
		SET format_type = EXCLUDED.format_type
		RETURNING id`
	err := b.DB.QueryRowContext(ctx, statement, format).Scan(&formatID)
	if err != nil {
		b.Logger.Error("Error inserting or updating format", "error", err)
		return 0, err
	}
	return formatID, nil
}

// Genres
func (b *BookUpdaterImpl) addOrGetGenreID(ctx context.Context, genreName string) (int, error) {
	var genreID int
	statement := `
		INSERT INTO genres (name)
		VALUES ($1)
		ON CONFLICT (name) DO UPDATE
		SET name = EXCLUDED.name
		RETURNING id`
	err := b.DB.QueryRowContext(ctx, statement, genreName).Scan(&genreID)
	if err != nil {
		b.Logger.Error("Error inserting or updating genre", "error", err)
		return 0, err
	}
	return genreID, nil
}

func (b *BookUpdaterImpl) RemoveSpecificGenres(ctx context.Context, bookID int, genres []string) error {

	statement := `
		DELETE FROM book_genres
		WHERE book_id = $1
		AND genre_id IN (
			SELECT id FROM genres WHERE name = ANY($2)
		)`

	_, err := b.DB.ExecContext(ctx, statement, bookID, pq.Array(genres))
	if err != nil {
		b.Logger.Error("Error removing specific genres", "error", err)
		return err
	}

	return nil
}

func (b *BookUpdaterImpl) AddGenre(ctx context.Context, bookID, genreID int) error {
	statement := `INSERT INTO book_genres (book_id, genre_id) VALUES ($1, $2)`
	_, err := b.DB.ExecContext(ctx, statement, bookID, genreID)
	if err != nil {
		b.Logger.Error("Error adding genre association", "error", err)
		return err
	}

	return nil
}
