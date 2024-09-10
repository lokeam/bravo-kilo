package data

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type BookModel struct {
	Logger        *slog.Logger
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
		Logger: logger,
	}
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
