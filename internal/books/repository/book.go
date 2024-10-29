package repository

import (
	"database/sql"
	"log/slog"
	"time"
)

type BookModel struct {
	Logger        *slog.Logger
}

type BookInfo struct {
	Title       string
	PublishDate string
}

type RichText struct {
	Ops []DeltaOp `json:"ops"`
}

type DeltaOp struct {
	Insert     interface{}            `json:"insert,omitempty"`     // Changed from *string to interface{}
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

type Book struct {
	ID              int                 `json:"id"`
	Title           string              `json:"title"`
	Subtitle        string              `json:"subtitle"`
	Description     RichText            `json:"description"`
	Notes           RichText            `json:"notes"`
	Language        string              `json:"language"`
	PageCount       int                 `json:"pageCount"`
	PublishDate     string              `json:"publishDate"`
	Authors         []string            `json:"authors"`
	ImageLink       string              `json:"imageLink"`
	Genres          []string            `json:"genres"`
	Formats         []string            `json:"formats"`
	Tags            []string            `json:"tags"`
	CreatedAt       time.Time           `json:"created_at"`
	LastUpdated     time.Time           `json:"lastUpdated"`
	ISBN10          string              `json:"isbn10"`
	ISBN13          string              `json:"isbn13"`
	IsInLibrary     bool                `json:"isInLibrary"`
	HasEmptyFields  bool                `json:"hasEmptyFields"`
	EmptyFields     []string            `json:"emptyFields"`
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
