package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
)

type BookDeleter interface {
  Delete(id int) error
}

type BookDeleterImpl struct {
	DB      *sql.DB
	Logger  *slog.Logger
}

func NewBookDeleter(db *sql.DB, logger *slog.Logger) (BookDeleter, error) {
	if db == nil || logger == nil {
		return nil, fmt.Errorf("book updater, database or logger is nil")
	}

	return &BookDeleterImpl{
		DB:          db,
		Logger:      logger,
	}, nil
}

func (b *BookDeleterImpl) Delete(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// Start a new transaction
	tx, err := b.DB.BeginTx(ctx, nil)
	if err != nil {
		b.Logger.Error("Book Model - Error starting transaction", "error", err)
		return err
	}

	// Roll back in case of error
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// Perform all deletions
	if err = b.deleteAssociations(ctx, tx, id); err != nil {
		return err
	}

	// Delete the book
	deleteBookStatement := `DELETE FROM books WHERE id = $1`
	if _, err = tx.ExecContext(ctx, deleteBookStatement, id); err != nil {
		b.Logger.Error("Book Model - Error deleting book", "error", err)
		return err
	}

	return nil
}

// Helper fn for Delete, handles deleting associated records in related tables
func (b *BookDeleterImpl) deleteAssociations(ctx context.Context, tx *sql.Tx, bookID int) error {
	// Delete associated user_books entries
	deleteUserBookStatement := `DELETE FROM user_books WHERE book_id = $1`
	if _, err := tx.ExecContext(ctx, deleteUserBookStatement, bookID); err != nil {
		b.Logger.Error("Book Model - Error deleting from user_books", "error", err)
		return err
	}

	// Delete associated book_genres entries
	deleteBookGenresStatement := `DELETE FROM book_genres WHERE book_id = $1`
	if _, err := tx.ExecContext(ctx, deleteBookGenresStatement, bookID); err != nil {
		b.Logger.Error("Book Model - Error deleting from book_genres", "error", err)
		return err
	}

	// Delete associated book_authors entries
	deleteBookAuthorsStatement := `DELETE FROM book_authors WHERE book_id = $1`
	if _, err := tx.ExecContext(ctx, deleteBookAuthorsStatement, bookID); err != nil {
		b.Logger.Error("Book Model - Error deleting from book_authors", "error", err)
		return err
	}

	// Delete associated book_formats entries
	deleteBookFormatsStatement := `DELETE FROM book_formats WHERE book_id = $1`
	if _, err := tx.ExecContext(ctx, deleteBookFormatsStatement, bookID); err != nil {
		b.Logger.Error("Book Model - Error deleting from book_formats", "error", err)
		return err
	}

	return nil
}
