package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lokeam/bravo-kilo/internal/dbconfig"
	"github.com/lokeam/bravo-kilo/internal/shared/collections"
)

var isbn10Cache              sync.Map
var isbn13Cache              sync.Map
var titleCache               sync.Map
var formatsCache             sync.Map
var booksByLangCache         sync.Map
var bookCountByGenreCache    sync.Map
var allBooksByGenresCache    sync.Map
var userTagsCache            sync.Map

const (
	ShortTTL  = 15 * time.Minute // Frequently changing data
	MediumTTL = 1 * time.Hour   // Session-related data
	LongTTL   = 24 * time.Hour  // Static data
)

var cacheTTL = map[string]time.Duration{
	"isbn10":                LongTTL,   // Book detail/form meta data
	"isbn13":                LongTTL,   // Book detail/form meta data
	"titles":                MediumTTL, // Book detail/form meta data
	"formats":               LongTTL,   // Home page statistics and Library page sorting
	"booksByLang":           ShortTTL,  // Home page statistics
	"bookCountByGenre":      ShortTTL,  // Home page statistics
	"allBooksByGenres":      ShortTTL,  // Library page sorting
	"userTags":              ShortTTL,  // Book detail/form meta data
}

type BookCache interface {
	InitPreparedStatements() error
	CleanupPreparedStatements() error
	GetAllBooksISBN10(userID int) (*collections.Set, error)
	GetAllBooksISBN13(userID int) (*collections.Set, error)
	GetAllBooksTitles(userID int) (*collections.Set, error)
	GetAllBooksPublishDate(userID int) ([]BookInfo, error)
	GetBooksByLanguage(ctx context.Context, userID int) (map[string]interface{}, error)
	InvalidateCaches(bookID int, userID int)
	StopCleanupWorker()
	FormatCacheKey(prefix string, userID int) string
	RecordCacheHit(item *CacheItem, cacheType string)
	RecordCacheMiss(item *CacheItem, cacheType string)
}

type BookCacheImpl struct {
	DB               *sql.DB
	Logger           *slog.Logger
	getAllLangStmt   *sql.Stmt
	stopCleanup      chan struct{} // Signal channel to stop cleanup goroutine
	isRunning        atomic.Bool // Thread-safe boolean for worker status
}

type CacheItem struct {
	Value       interface{}
	ExpireTime  time.Time
	Hits        int64 // For monitoring
	Misses      int64 // For monitoring
}

func NewBookCache(ctx context.Context, db *sql.DB, logger *slog.Logger) (BookCache, error) {
	if db == nil || logger == nil {
		return nil, fmt.Errorf("book cache, database or logger is nil")
	}

	bc := &BookCacheImpl{
		DB:     db,
		Logger: logger,
	}

	bc.startCleanupWorker(ctx)
	return bc, nil
}

func (b *BookCacheImpl) InitPreparedStatements() error {
	var err error

	// Prepared statment for GetLanguages
	b.getAllLangStmt, err = b.DB.Prepare(`
	SELECT language, COUNT(*) AS total
		FROM books
		INNER JOIN user_books ub ON books.id = ub.book_id
		WHERE ub.user_id = $1
		GROUP BY language
		ORDER BY total DESC`)
	if err != nil {
		return err
	}

	return nil
}

// ISBN10 + ISBN13 (Returns a HashSet)
func (b *BookCacheImpl) GetAllBooksISBN10(userID int) (*collections.Set, error) {
	cacheKey := b.FormatCacheKey("isbn10", userID)
	// Check cache
	b.Logger.Info("Fetching ISBN10 from cache")
	if cacheData, found := isbn10Cache.Load(cacheKey); found {
		if item, ok := cacheData.(*CacheItem); ok {
			b.RecordCacheHit(item, "isbn10")
			return item.Value.(*collections.Set), nil
		}
	} else {
		b.RecordCacheMiss(nil, "isbn10")
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	query := `
	SELECT b.isbn_10
	FROM books b
	INNER JOIN user_books ub ON b.id = ub.book_id
	WHERE ub.user_id = $1`

	rows, err := b.DB.QueryContext(ctx, query, userID)
	if err != nil {
		b.Logger.Error("Error retrieving ISBN10 numbers", "error", err)
		return nil, err
	}
	defer rows.Close()

	isbnSet := collections.NewSet()

	for rows.Next() {
		var isbn10 string
		if err := rows.Scan(&isbn10); err != nil {
			b.Logger.Error("Error scanning ISBN10", "error", err)
			return nil, err
		}
		isbnSet.Add(isbn10)
	}

	// Checking for errors after scanning all rows
	if err = rows.Err(); err != nil {
		b.Logger.Error("Error with rows", "error", err)
		return nil, err
	}

	// Cache result w/ metrics
	cacheItem := &CacheItem{
		Value:      isbnSet,
		ExpireTime: time.Now().Add(cacheTTL["isbn10"]),
		Hits:       0,
		Misses:     1, // First time cache miss
	}
	isbn10Cache.Store(cacheKey, cacheItem)
	b.Logger.Info("Caching ISBN10 for user", "userID", userID)

	return isbnSet, nil
}

// (Returns a HashSet)
func (b *BookCacheImpl) GetAllBooksISBN13(userID int) (*collections.Set, error) {
	cacheKey := b.FormatCacheKey("isbn13", userID)
	// Check cache
	b.Logger.Info("Fetching ISBN13 from cache")
	if cacheData, found := isbn13Cache.Load(cacheKey); found {
		if item, ok := cacheData.(*CacheItem); ok {
			b.RecordCacheHit(item, "isbn13")
			return item.Value.(*collections.Set), nil
		}
	} else {
		b.RecordCacheMiss(nil, "isbn13")
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	query := `
	SELECT b.isbn_13
	FROM books b
	INNER JOIN user_books ub ON b.id = ub.book_id
	WHERE ub.user_id = $1`

	rows, err := b.DB.QueryContext(ctx, query, userID)
	if err != nil {
		b.Logger.Error("Error retrieving ISBN13 numbers", "error", err)
		return nil, err
	}
	defer rows.Close()

	isbnSet := collections.NewSet()

	for rows.Next() {
		var isbn13 string
		if err := rows.Scan(&isbn13); err != nil {
			b.Logger.Error("Error scanning ISBN13", "error", err)
			return nil, err
		}
		isbnSet.Add(isbn13)
	}

	// Checking for errors after scanning all rows
	if err = rows.Err(); err != nil {
		b.Logger.Error("Error with rows", "error", err)
		return nil, err
	}

	// Cache the result w/ metrics
	cacheItem := &CacheItem{
		Value:      isbnSet,
		ExpireTime: time.Now().Add(cacheTTL["isbn13"]),
		Hits:       0,
		Misses:     1,
	}
	isbn13Cache.Store(cacheKey, cacheItem)
	b.Logger.Info("Caching ISBN13 for user", "userID", userID)

	return isbnSet, nil
}

// (Returns a HashSet)
func (b *BookCacheImpl) GetAllBooksTitles(userID int) (*collections.Set, error) {
	cacheKey := b.FormatCacheKey("titles", userID)
	// Check cache
	b.Logger.Info("Fetching Title info from cache")
	if cacheData, found := titleCache.Load(cacheKey); found {
		if item, ok := cacheData.(*CacheItem); ok {
			b.RecordCacheHit(item, "titles")
			return item.Value.(*collections.Set), nil
		}
	} else {
		b.RecordCacheMiss(nil, "titles")
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	query := `
	SELECT b.title
	FROM books b
	INNER JOIN user_books ub ON b.id = ub.book_id
	WHERE ub.user_id = $1`

	rows, err := b.DB.QueryContext(ctx, query, userID)
	if err != nil {
		b.Logger.Error("Error retrieving book titles", "error", err)
		return nil, err
	}
	defer rows.Close()

	titleSet := collections.NewSet()

	for rows.Next() {
		var bookTitle string
		if err := rows.Scan(&bookTitle); err != nil {
			b.Logger.Error("Error scanning book title", "error", err)
			return nil, err
		}
		titleSet.Add(bookTitle)
	}

	if err = rows.Err(); err != nil {
		b.Logger.Error("Error with rows", "error", err)
		return nil, err
	}

	// Cache result
	cacheItem := &CacheItem{
		Value:      titleSet,
		ExpireTime: time.Now().Add(cacheTTL["titles"]),
		Hits:       0,
		Misses:     1,
	}
	titleCache.Store(cacheKey, cacheItem)
	b.Logger.Info("Caching Title info for user", "userID", userID)

	return titleSet, nil
}

// (Return a Slice of BookInfo Structs to handle books with duplicate titles)
func (b *BookCacheImpl) GetAllBooksPublishDate(userID int) ([]BookInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbconfig.DBTimeout)
	defer cancel()

	query := `
	SELECT b.title, b.publish_date
	FROM books b
	INNER JOIN user_books ub ON b.id = ub.book_id
	WHERE ub.user_id = $1`

	rows, err := b.DB.QueryContext(ctx, query, userID)
	if err != nil {
			b.Logger.Error("Error retrieving book titles and publish dates", "error", err)
			return nil, err
	}
	defer rows.Close()

	var books []BookInfo

	for rows.Next() {
			var book BookInfo
			var publishDate time.Time
			if err := rows.Scan(&book.Title, &publishDate); err != nil {
					b.Logger.Error("Error scanning book title and publish date", "error", err)
					return nil, err
			}

			// Format publish date to "YYYY-MM-DD"
			book.PublishDate = publishDate.Format("2006-01-02")
			books = append(books, book)
	}

	if err = rows.Err(); err != nil {
			b.Logger.Error("Error with rows", "error", err)
			return nil, err
	}

	return books, nil
}

// Languages
func (b *BookCacheImpl) GetBooksByLanguage(ctx context.Context, userID int) (map[string]interface{}, error) {
	// Check cache
	cacheKey := b.FormatCacheKey("booksByLang", userID)
	b.Logger.Info("Fetching Title info from cache")
    // Check cache and validate data
    if cacheData, found := booksByLangCache.Load(cacheKey); found {
			if item, ok := cacheData.(*CacheItem); ok {
					if data, ok := item.Value.(map[string]interface{}); ok {
							if langs, exists := data["booksByLang"].([]map[string]interface{}); exists && len(langs) > 0 {
									b.RecordCacheHit(item, "booksByLang")
									return data, nil
							}
					}
			}
	} else {
		b.RecordCacheMiss(nil, "booksByLang")
	}

	// Use prepared statement if available
	var rows *sql.Rows
	var err error

	if b.getAllLangStmt != nil {
			b.Logger.Info("Using prepared statement for fetching books by language")
			rows, err = b.getAllLangStmt.QueryContext(ctx, userID)
	} else {
			b.Logger.Warn("Prepared statement for fetching books by language is unavailable. Falling back to raw query")
			query := `
			SELECT language, COUNT(*) AS total
			FROM books
			INNER JOIN user_books ub ON books.id = ub.book_id
			WHERE ub.user_id = $1
			GROUP BY language
			ORDER BY total DESC`
			rows, err = b.DB.QueryContext(ctx, query, userID)
	}

	if err != nil {
			b.Logger.Error("Error fetching books by language", "error", err)
			return nil, err
	}
	defer rows.Close()

	var booksByLang []map[string]interface{}

	// Process the results
	for rows.Next() {
			var language string
			var total int

			if err := rows.Scan(&language, &total); err != nil {
					b.Logger.Error("Error scanning books by language", "error", err)
					return nil, err
			}

			booksByLang = append(booksByLang, map[string]interface{}{
					"label": language,
					"count": total,
			})
	}

	if err = rows.Err(); err != nil {
			b.Logger.Error("Error finalizing rows", "error", err)
			return nil, err
	}

	// Add validation before caching
	if len(booksByLang) == 0 {
		b.Logger.Warn("No languages found for user", "userID", userID)
	} else {
		b.Logger.Info("Caching languages for user",
				"userID", userID,
				"languagesCount", len(booksByLang))
	}

	result := map[string]interface{}{
		"booksByLang": booksByLang,
	}

	// Cache the result
	cacheItem := &CacheItem{
		Value:      result,
		ExpireTime: time.Now().Add(cacheTTL["booksByLang"]),
		Hits:       0,
		Misses:     1, // First miss that caused cache entry
	}
	booksByLangCache.Store(cacheKey, cacheItem)
	b.Logger.Info("Caching books by language", "userID", userID)

	return result, nil
}

// Invalidate Caches
func (b *BookCacheImpl) InvalidateCaches(bookID int, userID int) {

    // Book-specific caches
    if bookID > 0 {
			formatsCache.Delete(b.FormatCacheKey("formats", bookID))
	}

    // Book-specific caches
    isbn10Cache.Delete(b.FormatCacheKey("isbn10", userID))
    isbn13Cache.Delete(b.FormatCacheKey("isbn13", userID))
    titleCache.Delete(b.FormatCacheKey("titles", userID))

    // User-specific caches
    booksByLangCache.Delete(b.FormatCacheKey("booksByLang", userID))
    bookCountByGenreCache.Delete(b.FormatCacheKey("bookCountByGenre", userID))
    allBooksByGenresCache.Delete(b.FormatCacheKey("allBooksByGenres", userID))
    userTagsCache.Delete(b.FormatCacheKey("userTags", userID))

    b.Logger.Info("Caches invalidated", "bookID", bookID, "userID", userID)
}

// Format cache key to ensure only exactly one key format
func (b *BookCacheImpl) FormatCacheKey(prefix string, userID int) string {
	return fmt.Sprintf("%s:%d", prefix, userID)
}

// Cleanup prepared statements
func (b *BookCacheImpl) CleanupPreparedStatements() error {
	if b.getAllLangStmt != nil {
		if err := b.getAllLangStmt.Close(); err != nil {
			b.Logger.Error("Failed to close getAllLangStmt", "error", err)
			return fmt.Errorf("error closing getAllLangStmt: %w", err)
		}
		b.getAllLangStmt = nil
	}

	b.Logger.Info("Successfully cleaned up prepared statements")
	return nil
}

// Start cleanup worker
func (b *BookCacheImpl) startCleanupWorker(ctx context.Context) {
	b.stopCleanup = make(chan struct{})
	b.isRunning.Store(true)

	go func() {
		ticker := time.NewTicker(15 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <- ticker.C:
				b.cleanup()
			case <- b.stopCleanup:
				b.Logger.Info("Cleanup worker stopped")
				return
			case <- ctx.Done():
				b.Logger.Info("Cleanup worker stopped due to context cancellation")
				return
			}
		}
	}()

	b.Logger.Info("Cache cleanup worker started")
}

// Cleanup method
func (b *BookCacheImpl) cleanup() {
	b.Logger.Info("Starting cache cleanup")
	start := time.Now()
	cleanupCache(&isbn10Cache, "isbn10", b.Logger)
	cleanupCache(&isbn13Cache, "isbn13", b.Logger)
	cleanupCache(&titleCache, "titles", b.Logger)
	cleanupCache(&formatsCache, "formats", b.Logger)
	cleanupCache(&booksByLangCache, "booksByLang", b.Logger)
	cleanupCache(&bookCountByGenreCache, "bookCountByGenre", b.Logger)
	cleanupCache(&allBooksByGenresCache, "allBooksByGenres", b.Logger)
	cleanupCache(&userTagsCache, "userTags", b.Logger)

	b.Logger.Info("Cache cleanup completed",
		"duration", time.Since(start),
	)
}

// Helper fn to cleanup individual caches
func cleanupCache(cache *sync.Map, cacheType string, logger *slog.Logger) {
	var deleted int

	cache.Range(func(key, value interface{}) bool {
		if item, ok := value.(*CacheItem); ok {
			if time.Now().After(item.ExpireTime) {
				cache.Delete(key)
				deleted++
			}
		}
		return true
	})

	if deleted > 0 {
		logger.Info("Cleaned up expired items from cache", "cache", cacheType, "deleted", deleted)
	}
}

// Method to stop the cleanup worker
func (b *BookCacheImpl) StopCleanupWorker() {
	if b.isRunning.Load() {
		close(b.stopCleanup)
		b.isRunning.Store(false)
		b.Logger.Info("Cache cleanup worker stop signal sent")
	}
}

// Helper methods for tracking metrics
func (b *BookCacheImpl) RecordCacheHit(item *CacheItem, cacheType string) {
	atomic.AddInt64(&item.Hits, 1)
	b.Logger.Debug("Cache hit", "type", cacheType, "hits", item.Hits, "misses", item.Misses)
}

func (b *BookCacheImpl) RecordCacheMiss(item *CacheItem, cacheType string) {
	if item != nil {
		atomic.AddInt64(&item.Misses, 1)
		b.Logger.Debug("Cache miss", "type", cacheType, "hits", item.Hits, "misses", item.Misses)
	}
}