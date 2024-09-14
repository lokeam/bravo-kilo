package services

import (
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"unicode"

	"github.com/lokeam/bravo-kilo/internal/books/repository"
)

type ExportService interface {
	GenerateBookCSV(userID int, writer io.Writer) error
}

type ExportServiceImpl struct {
  bookRepository  repository.BookRepository
	logger          *slog.Logger
}

func NewExportService (
	logger *slog.Logger,
	bookRepo repository.BookRepository,
) (ExportService, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger is nil")
	}

	return &ExportServiceImpl{
		logger:         logger,
		bookRepository: bookRepo,
	}, nil
}

// Generate Book CSV
func (e *ExportServiceImpl) GenerateBookCSV(userID int, writer io.Writer) error {
	books, err := e.bookRepository.GetAllBooksByUserID(userID)
	if err != nil {
		e.logger.Error("Failed to fetch books for user", "error", err)
		return err
	}

	csvWriter := csv.NewWriter(writer)
	defer func() {
		csvWriter.Flush()
		if err := csvWriter.Error(); err != nil {
			e.logger.Error("Error flushing CSV writer: ", "error", err)
		}
	}()

	if _, err := writer.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		return fmt.Errorf("failed to write UTF-8 BOM: %w", err)
	}

	header := []string{"Title", "Authors", "ISBN", "Language", "PageCount", "Publish Date"}
	if err := csvWriter.Write(header); err != nil {
		e.logger.Error("Failed to write CSV header", "error", err)
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
			e.logger.Error("Failed to write CSV row", "bookID", book.ID, "error", err )
			return err
		}
	}

	return nil
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
