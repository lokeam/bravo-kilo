package data

import (
	"context"
	"database/sql"
	"log/slog"
)

type FormatModel struct {
	DB     *sql.DB
	Logger *slog.Logger
}

type Format struct {
	ID         int    `json:"id"`
	BookID     int    `json:"book_id"`
	FormatType string `json:"format_type"`
}

func (f *FormatModel) Insert(bookID int, formatType string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var formatID int
	statement := `INSERT INTO formats (book_id, format_type) VALUES ($1, $2) RETURNING id`
	err := f.DB.QueryRowContext(ctx, statement, bookID, formatType).Scan(&formatID)
	if err != nil {
			f.Logger.Error("Format Model - Error inserting format", "error", err)
			return 0, err
	}

	return formatID, nil
}

func (f *FormatModel) GetByBookID(bookID int) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `SELECT format_type FROM formats WHERE book_id = $1`
	rows, err := f.DB.QueryContext(ctx, query, bookID)
	if err != nil {
		f.Logger.Error("Format Model - Error fetching formats by book ID", "error", err)
		return nil, err
	}
	defer rows.Close()

	var formats []string
	for rows.Next() {
		var format string
		if err := rows.Scan(&format); err != nil {
			f.Logger.Error("Error scanning format", "error", err)
			return nil, err
		}
		formats = append(formats, format)
	}

	return formats, nil
}

func (f *FormatModel) DeleteByBookID(bookID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	statement := `DELETE FROM formats WHERE book_id = $1`
	_, err := f.DB.ExecContext(ctx, statement, bookID)
	if err != nil {
		f.Logger.Error("Format Model - Error deleting formats by book ID", "error", err)
		return err
	}

	return nil
}
