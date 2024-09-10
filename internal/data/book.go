package data

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type BookModel struct {
	Deleter       BookDeleter
	Logger        *slog.Logger
	Repository    BookRepository
	Updater       BookUpdater
}

type BookInfo struct {
	Title       string
	PublishDate string
}

type Book struct {
	ID              int        `json:"id"`
	Title           string     `json:"title"`
	Subtitle        string     `json:"subtitle"`
	Description     string     `json:"description"`
	Language        string     `json:"language"`
	PageCount       int        `json:"pageCount"`
	PublishDate     string     `json:"publishDate"`
	Authors         []string   `json:"authors"`
	ImageLink       string     `json:"imageLink"`
	Genres          []string   `json:"genres"`
	Notes           string     `json:"notes"`
	Formats         []string   `json:"formats"`
	Tags            []string   `json:"tags"`
	CreatedAt       time.Time  `json:"created_at"`
	LastUpdated     time.Time  `json:"lastUpdated"`
	ISBN10          string     `json:"isbn10"`
	ISBN13          string     `json:"isbn13"`
	IsInLibrary     bool       `json:"isInLibrary"`
	HasEmptyFields  bool       `json:"hasEmptyFields"`
	EmptyFields     []string   `json:"emptyFields"`
}

type UserTagsCacheEntry struct {
	data      map[string]interface{}
	timestamp time.Time
}

type BooksByGenresCacheEntry struct {
	data      map[string]interface{}
	timestamp time.Time
}

func NewBookModel(db *sql.DB, logger *slog.Logger) *BookModel {
	return &BookModel{
		Deleter: &BookDeleterImpl{
			DB: db,
			Logger: logger,
		},
		Logger: logger,
		Repository: &BookRepositoryImpl{
			DB:      db,
			Logger:  logger,
		},
		Updater: &BookUpdaterImpl{
			DB: db,
			Logger: logger,
		},
	}
}

func (b *BookModel) GetBooksListByGenre(ctx context.Context, userID int) (map[string]interface{}, error) {
	const cacheTTL = time.Hour

	// Check cache with TTL
	if cacheEntry, found := booksByGenresCache.Load(userID); found {
			entry := cacheEntry.(BooksByGenresCacheEntry)
			if time.Since(entry.timestamp) < cacheTTL {
					b.Logger.Info("Fetching genres info from cache for user", "userID", userID)
					return entry.data, nil
			}
			booksByGenresCache.Delete(userID) // Cache entry expired, delete it
	}

	var rows *sql.Rows
	var err error

	// Use prepared statement if available
	if b.getBookListByGenreStmt != nil {
			b.Logger.Info("Using prepared statement for fetching genres")
			rows, err = b.getBookListByGenreStmt.QueryContext(ctx, userID)
	} else {
			b.Logger.Warn("Prepared statement for fetching genres unavailable. Falling back to raw SQL query")
			query := `
					SELECT g.name, COUNT(DISTINCT b.id) AS total_books
					FROM books b
					INNER JOIN book_genres bg ON b.id = bg.book_id
					INNER JOIN genres g ON bg.genre_id = g.id
					INNER JOIN user_books ub ON b.id = ub.book_id
					WHERE ub.user_id = $1
					GROUP BY g.name
					ORDER BY total_books DESC`
			rows, err = b.DB.QueryContext(ctx, query, userID)
	}

	if err != nil {
			b.Logger.Error("Error fetching books by genre", "error", err)
			return nil, err
	}
	defer rows.Close()

	// Collect genres and total book count
	var booksByGenre []map[string]interface{}

	for rows.Next() {
			var genre string
			var count int
			if err := rows.Scan(&genre, &count); err != nil {
					b.Logger.Error("Error scanning genre data", "error", err)
					return nil, err
			}

			booksByGenre = append(booksByGenre, map[string]interface{}{
					"label": genre,
					"count": count,
			})
	}

	if err = rows.Err(); err != nil {
			b.Logger.Error("Error iterating rows", "error", err)
			return nil, err
	}

	// Prepare the result
	result := map[string]interface{}{
			"booksByGenre": booksByGenre,
	}

	// Cache the result with TTL
	booksByGenresCache.Store(userID, BooksByGenresCacheEntry{
			data:      result,
			timestamp: time.Now(),
	})
	b.Logger.Info("Caching genres info for user", "userID", userID)

	return result, nil
}

func (b *BookModel) GetTagsForBook(ctx context.Context, bookID int) ([]string, error) {

	query := `
	SELECT tags
	FROM books
	WHERE id = $1`

	var tagsJSON []byte
	err := b.DB.QueryRowContext(ctx, query, bookID).Scan(&tagsJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			b.Logger.Warn("No tags found for book", "bookID", bookID)
			// Return empty slice if no tags are found
			return nil, nil
		}
		b.Logger.Error("Error fetching tags for book", "error", err)
		return nil, err
	}

	// Return an empty slice if tagsJSON is null
	if tagsJSON == nil {
		return []string{}, nil
	}

	// Unmarshal the JSON array of tags
	var tags []string
	if err := json.Unmarshal(tagsJSON, &tags); err != nil {
		b.Logger.Error("Error unmarshalling tags JSON", "error", err)
		return nil, err
	}

	return tags, nil
}


// Helper fn: Sanitize Fields
func sanitizeCSVField(field string) string {
	// Guard Clause
	if field == "" {
		return ""
	}

	// Prevent formula injection
	if strings.ContainsAny(field, "=+-@") {
		field = "'" + field
	}

	// Escape double quotes
	field = strings.ReplaceAll(field, "\"", "\"\"")

	// Remove/replace ctl chars
	var sanitized strings.Builder
	for _, r := range field {
		if unicode.IsControl(r) {
			sanitized.WriteRune(' ')
		} else {
			sanitized.WriteRune(r)
		}
	}
	return sanitized.String()
}

// Generate Book CSV
func (b *BookModel) GenerateBookCSV(userID int, writer io.Writer) error {
	books, err := b.GetAllBooksByUserID(userID)
	if err != nil {
		b.Logger.Error("Failed to fetch books for user", "error", err)
		return err
	}

	csvWriter := csv.NewWriter(writer)
	defer func() {
		csvWriter.Flush()
		if err := csvWriter.Error(); err != nil {
			b.Logger.Error("Error flushing CSV writer: ", "error", err)
		}
	}()

	if _, err := writer.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		return fmt.Errorf("failed to write UTF-8 BOM: %w", err)
	}

	header := []string{"Title", "Authors", "ISBN", "Language", "PageCount", "Publish Date"}
	if err := csvWriter.Write(header); err != nil {
		b.Logger.Error("Failed to write CSV header", "error", err)
		return err
	}

	// Write data
	for _, book := range books {
		row := []string{
			sanitizeCSVField(book.Title),
			sanitizeCSVField(strings.Join(book.Authors, ", ")),
			sanitizeCSVField(book.ISBN13),
			sanitizeCSVField(book.Language),
			sanitizeCSVField(strconv.Itoa(book.PageCount)),
			sanitizeCSVField(book.PublishDate),
		}
		// Add UTF-8 Byte Order Mark for Excel
		if err := csvWriter.Write(row); err != nil {
			b.Logger.Error("Failed to write CSV row", "bookID", book.ID, "error", err )
			return err
		}
	}

	return nil
}
