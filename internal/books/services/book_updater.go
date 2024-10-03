package services

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/lokeam/bravo-kilo/internal/books/repository"
	"github.com/lokeam/bravo-kilo/internal/shared/transaction"
)

type BookUpdaterService interface {
	UpdateBookEntry(ctx context.Context, book repository.Book) error
}

type BookUpdaterServiceImpl struct {
	DB           *sql.DB
	logger       *slog.Logger
	bookRepo     repository.BookRepository
	authorRepo   repository.AuthorRepository
	bookCache    repository.BookCache
	formatRepo   repository.FormatRepository
	genreRepo    repository.GenreRepository
	bookService  BookService
	dbManager    transaction.DBManager
}

func NewBookUpdaterService(
	db *sql.DB,
	logger *slog.Logger,
	bookRepo repository.BookRepository,
	authorRepo repository.AuthorRepository,
	bookCache repository.BookCache,
	formatRepo repository.FormatRepository,
	genreRepo repository.GenreRepository,
	bookService BookService,
	dbManager transaction.DBManager,
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
		authorRepo:  authorRepo,
		bookCache:   bookCache,
		bookService: bookService,
		dbManager:   dbManager,
	}, nil
}

func (b *BookUpdaterServiceImpl) UpdateBookEntry(ctx context.Context, book repository.Book) error {
	// Invalidate caches
	b.bookCache.InvalidateCaches(book.ID)
	b.logger.Info("Cache invalidated for book", "book", book.ID)

	// Normalize + sanitize book data before proceeding
	b.bookService.NormalizeBookData(&book)
	b.bookService.SanitizeBookData(&book)

	// Start transaction
	tx, err := b.dbManager.BeginTransaction(ctx)
	if err != nil {
		b.logger.Error("Update book entry, error starting transaction", "error", err)
		return err
	}
	defer tx.Rollback()

	err = b.bookRepo.UpdateBook(ctx, tx, book)
	if err != nil {
		return err
	}

	// Update authors
	err = b.bookService.CreateEntries(
		ctx,
		tx,
		book.ID,
		book.Authors,
		b.authorRepo.GetAuthorIDByName,
		b.authorRepo.InsertAuthor,
		b.authorRepo.AssociateBookWithAuthor,
	)
	if err != nil {
		b.logger.Error("Error updating authors", "error", err)
		return err
	}

	// Update genres
	err = b.bookService.CreateEntries(
		ctx,
		tx,
		book.ID,
		book.Genres,
		b.genreRepo.GetGenreIDByName,
		b.genreRepo.InsertGenre,
		b.genreRepo.AssociateBookWithGenre,
	)
	if err != nil {
		b.logger.Error("Error updating genres", "error", err)
		return err
	}

	// Update formats
	err = b.bookService.InsertFormats(ctx, tx, book.ID, book.Formats)
	if err != nil {
		b.logger.Error("Error updating formats", "error", err)
		return err
	}

	if err = b.dbManager.CommitTransaction(tx); err != nil {
		b.logger.Error("Book updater error commiting transaction", "error", err)
		return err
	}

	return nil
}
