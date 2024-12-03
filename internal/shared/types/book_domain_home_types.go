package types

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/lokeam/bravo-kilo/internal/books/repository"
	binaryMarshaler "github.com/lokeam/bravo-kilo/internal/shared/binary"
)

type HomePageData struct {
	Books           []repository.Book     `json:"books"`
	BooksByFormat   FormatCountStats      `json:"booksByFormat"`
	HomePageStats   HomePageStats         `json:"homepageStats"`
	logger          *slog.Logger
}

type FormatCountStats struct {
	Physical   int `json:"physical"`
	Digital    int `json:"eBook"`
	AudioBook  int `json:"audioBook"`
}

type HomePageStats struct {
	UserBkLang     LanguageStats   `json:"userBkLang"`
	UserBkGenre    GenreStats      `json:"userBkGenres"`
	UserTags       TagStats        `json:"userTags"`
	UserAuthors    AuthorStats     `json:"userAuthors"`
}

type StatItem struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

type LanguageStats struct {
	BooksByLang []StatItem `json:"booksByLang"`
}

type GenreStats struct {
	BooksByGenre []StatItem `json:"bookByGenre"`
}

type TagStats struct {
	UserTags []StatItem `json:"userTags"`
}

type AuthorStats struct {
	BooksByAuthor []StatItem `json:"booksByAuthor"`
}

type BookDomainHomeValidationConfig struct {
	statType string
	getItems func(book repository.Book) []string
	getStats func(h *HomePageData) []StatItem
}

func NewHomePageData(logger *slog.Logger) *HomePageData {
	if logger == nil {
		logger = slog.Default()
}

	return &HomePageData{
		Books:           make([]repository.Book, 0),
		BooksByFormat:   FormatCountStats{},
		HomePageStats:   HomePageStats{
			UserBkLang:     LanguageStats{BooksByLang: make([]StatItem, 0)},
			UserBkGenre:    GenreStats{BooksByGenre: make([]StatItem, 0)},
			UserTags:       TagStats{UserTags: make([]StatItem, 0)},
			UserAuthors:    AuthorStats{BooksByAuthor: make([]StatItem, 0)},
		},
		logger:          logger,
	}
}

func (h *HomePageData) Validate() error {
	// Early nil check before any operations
	if h == nil {
			return fmt.Errorf("HomePageData is nil")
	}

	// Initialize logger if needed
	if h.logger == nil {
			h.logger = slog.Default()
	}

	start := time.Now()
	defer func() {
			h.logger.Debug("validation completed",
					"duration", time.Since(start),
			)
	}()

	// Initialize all required structures
	if err := h.initializeStructures(); err != nil {
			return fmt.Errorf("initialization failed: %w", err)
	}

	// Validate data consistency
	if err := h.validateDataConsistency(); err != nil {
			return fmt.Errorf("consistency validation failed: %w", err)
	}

	return nil
}



// Helper functions

// Initializes required data structures to be used in HomePageData
func (h *HomePageData) initializeStructures() error {
	// Books initialization
	if h.Books == nil {
			h.Books = make([]repository.Book, 0)
	}

	// Initialize book fields if necessary
	for i := range h.Books {
			if h.Books[i].Authors == nil {
					h.Books[i].Authors = make([]string, 0)
			}
			if h.Books[i].Genres == nil {
					h.Books[i].Genres = make([]string, 0)
			}
			if h.Books[i].Tags == nil {
					h.Books[i].Tags = make([]string, 0)
			}
			if h.Books[i].Formats == nil {
					h.Books[i].Formats = make([]string, 0)
			}
	}

	// Initialize statistics structures
	if err := h.initializeStats(); err != nil {
			return fmt.Errorf("stats initialization failed: %w", err)
	}

	return nil
}

// Initialize structures for statistics
func (h *HomePageData) initializeStats() error {
	// Language stats initialization
	if h.HomePageStats.UserBkLang.BooksByLang == nil {
			h.HomePageStats.UserBkLang.BooksByLang = make([]StatItem, 0)
	}

	// Genre stats initialization
	if h.HomePageStats.UserBkGenre.BooksByGenre == nil {
			h.HomePageStats.UserBkGenre.BooksByGenre = make([]StatItem, 0)
	}

	// Tags stats initialization
	if h.HomePageStats.UserTags.UserTags == nil {
			h.HomePageStats.UserTags.UserTags = make([]StatItem, 0)
	}

	// Author stats initialization
	if h.HomePageStats.UserAuthors.BooksByAuthor == nil {
			h.HomePageStats.UserAuthors.BooksByAuthor = make([]StatItem, 0)
	}

	return nil
}

// Optional method to kick off validation logic
func (h *HomePageData) validateDataConsistency() error {
	// Validate books field
	if err := h.validateBooksIntegrity(); err != nil {
			return fmt.Errorf("books validation failed: %w", err)
	}

	// Validate format counts field
	if err := h.validateFormatCounts(); err != nil {
			return fmt.Errorf("format counts validation failed: %w", err)
	}

	// Validate statistics
	if err := h.validateStatistics(); err != nil {
			return fmt.Errorf("statistics validation failed: %w", err)
	}

	return nil
}

// Validate all statistics using unified validation fn
func (h *HomePageData) validateStatistics() error {
	if err := h.validateLanguageStats(); err != nil {
			return fmt.Errorf("language stats validation failed: %w", err)
	}
	if err := h.validateGenreStats(); err != nil {
			return fmt.Errorf("genre stats validation failed: %w", err)
	}
	if err := h.validateTagStats(); err != nil {
			return fmt.Errorf("tag stats validation failed: %w", err)
	}
	if err := h.validateAuthorStats(); err != nil {
			return fmt.Errorf("author stats validation failed: %w", err)
	}
	return nil
}

// Unified validation fn for all stats
func (h *HomePageData) validateHomePageStatField(config BookDomainHomeValidationConfig) error {
	h.logger.Debug("starting validation",
			"statType", config.statType,
	)

	// Create a map to track counts
	counts := make(map[string]int)

	// Count items from books
	for _, book := range h.Books {
			items := config.getItems(book)
			for _, item := range items {
					if item == "" {
							h.logger.Error("empty value found",
									"statType", config.statType,
									"bookTitle", book.Title,
									"bookID", book.ID,
							)
							return fmt.Errorf("book %q (ID: %d) has empty %s",
									book.Title, book.ID, config.statType)
					}
					counts[item]++
			}
	}

	// Verify each stat matches book counts
	stats := config.getStats(h)
	for _, stat := range stats {
			if stat.Label == "" {
					h.logger.Error("empty label in stats",
							"statType", config.statType,
					)
					return fmt.Errorf("empty %s label in statistics", config.statType)
			}

			expectedCount, exists := counts[stat.Label]
			if !exists {
					h.logger.Error("stat not found in books",
							"statType", config.statType,
							"label", stat.Label,
					)
					return fmt.Errorf("%s %q in statistics not found in any book",
							config.statType, stat.Label)
			}

			if expectedCount != stat.Count {
					h.logger.Error("count mismatch",
							"statType", config.statType,
							"label", stat.Label,
							"expectedCount", expectedCount,
							"actualCount", stat.Count,
					)
					return fmt.Errorf("%s %q count mismatch: expected %d, got %d",
							config.statType, stat.Label, expectedCount, stat.Count)
			}
	}

	h.logger.Debug("validation completed",
			"statType", config.statType,
			"itemCount", len(counts),
			"statsCount", len(stats),
	)

	return nil
}

// Validate books
func (h *HomePageData) validateBooksIntegrity() error {
	for i, book := range h.Books {
			if book.Title == "" {
					h.logger.Error("invalid book",
							"index", i,
							"error", "empty title")
					return fmt.Errorf("book at index %d has empty title", i)
			}
	}
	return nil
}

// Validate format counts
func (h *HomePageData) validateFormatCounts() error {
	formatCount := struct {
			physical  int
			eBook     int
			audioBook int
	}{}

	// Count formats from books
	for _, book := range h.Books {
			for _, format := range book.Formats {
					switch format {
					case "physical":
							formatCount.physical++
					case "eBook":
							formatCount.eBook++
					case "audioBook":
							formatCount.audioBook++
					}
			}
	}

	// Verify counts match
	if formatCount.physical != h.BooksByFormat.Physical {
			return fmt.Errorf("physical book count mismatch: got %d, expected %d",
					h.BooksByFormat.Physical, formatCount.physical)
	}
	if formatCount.eBook != h.BooksByFormat.Digital {
			return fmt.Errorf("eBook count mismatch: got %d, expected %d",
					h.BooksByFormat.Digital, formatCount.eBook)
	}
	if formatCount.audioBook != h.BooksByFormat.AudioBook {
			return fmt.Errorf("audioBook count mismatch: got %d, expected %d",
					h.BooksByFormat.AudioBook, formatCount.audioBook)
	}

	return nil
}

// Ensure language statistics are consistent with book data - uses validateHomePageStatField
func (h *HomePageData) validateLanguageStats() error {
	return h.validateHomePageStatField(BookDomainHomeValidationConfig{
			statType: "language",
			getItems: func(book repository.Book) []string {
					return []string{book.Language} // Language is a single string
			},
			getStats: func(h *HomePageData) []StatItem {
					return h.HomePageStats.UserBkLang.BooksByLang
			},
	})
}

// Ensure genre statistics are consistent with book data
func (h *HomePageData) validateGenreStats() error {
	return h.validateHomePageStatField(BookDomainHomeValidationConfig{
			statType: "genre",
			getItems: func(book repository.Book) []string {
					return book.Genres
			},
			getStats: func(h *HomePageData) []StatItem {
					return h.HomePageStats.UserBkGenre.BooksByGenre
			},
	})
}

// Ensure tag statistics are consistent with book data
func (h *HomePageData) validateTagStats() error {
	return h.validateHomePageStatField(BookDomainHomeValidationConfig{
			statType: "tag",
			getItems: func(book repository.Book) []string {
					return book.Tags
			},
			getStats: func(h *HomePageData) []StatItem {
					return h.HomePageStats.UserTags.UserTags
			},
	})
}


// Ensure author statistics are consistent with book data
func (h *HomePageData) validateAuthorStats() error {
	return h.validateHomePageStatField(BookDomainHomeValidationConfig{
			statType: "author",
			getItems: func(book repository.Book) []string {
					return book.Authors
			},
			getStats: func(h *HomePageData) []StatItem {
					return h.HomePageStats.UserAuthors.BooksByAuthor
			},
	})
}


// MarshalBinary implements the encoding.BinaryMarshaler interface
func (hpd *HomePageData) MarshalBinary() ([]byte, error) {
	// Directly use the shared binary marshaler
	data, err := binaryMarshaler.MarshalBinary(hpd)
	if err != nil {
		hpd.logger.Error("library types binary marshal failed",
			"error", err,
		)
		return nil, fmt.Errorf("library types binary marshal failed: %w", err)
	}
	return data, nil
}

func (hpd *HomePageData) UnmarshalBinary(data []byte) error {
	hpd.logger.Debug("starting binary unmarshaling",
			"dataSize", len(data),
	)

	// Validate size, min and max
	if len(data) < 4 {
			hpd.logger.Error("data too short for length prefix",
					"dataSize", len(data),
					"minimumRequired", 4,
			)
			return fmt.Errorf("data too short: %d bytes", len(data))
	}

	// Total size check
	totalSize := len(data)
	if totalSize > binaryMarshaler.MaxMemoryLimit {
		hpd.logger.Error("total data exceeds limit",
      "size", totalSize,
      "limit", binaryMarshaler.MaxMemoryLimit,
    )
		return fmt.Errorf("total data size %d exceeds limit %d", totalSize, binaryMarshaler.MaxMemoryLimit)
	}

	// Read and validate length prefix
	var claimedLength uint32
	lengthReader := bytes.NewReader(data[:4])
	if err := binary.Read(lengthReader, binary.LittleEndian, &claimedLength); err != nil {
			hpd.logger.Error("failed to read length prefix",
					"error", err,
			)
			return fmt.Errorf("failed to read length prefix: %w", err)
	}

	// Validate claimed length before using it
	if claimedLength > uint32(binaryMarshaler.MaxMemoryLimit) {
		hpd.logger.Error("claimed length exceeds limit",
				"claimedLength", claimedLength,
				"limit", binaryMarshaler.MaxMemoryLimit,
		)
		return fmt.Errorf("claimed length %d exceeds limit %d", claimedLength, binaryMarshaler.MaxMemoryLimit)
	}

	// Verify actual data matches claimed length
	actualDataLength := uint32(len(data) - 4)
	if actualDataLength != claimedLength {
			hpd.logger.Error("length mismatch",
					"claimed", claimedLength,
					"actual", actualDataLength,
					"totalSize", len(data),
			)
			return fmt.Errorf("length mismatch: claimed %d, actual %d", claimedLength, actualDataLength)
	}

	// JSON Validation
	jsonData := data[4:]
	if !json.Valid(jsonData) {
		hpd.logger.Error("invalid JSON structure",
				"dataSize", len(jsonData),
				"claimedLength", claimedLength,
				"totalSize", len(data),
		)
		return fmt.Errorf("invalid JSON structure in binary data")
	}

	// Unmarshal into temporary structure specific to HomePageData
	var temp struct {
		Books         []repository.Book `json:"books"`
		BooksByFormat struct {
				Physical  int `json:"physical"`
				Digital   int `json:"eBook"`
				AudioBook int `json:"audioBook"`
		} `json:"booksByFormat"`
		HomePageStats struct {
				UserBkLang  struct {
						BooksByLang []StatItem `json:"booksByLang"`
				} `json:"userBkLang"`
				UserBkGenre struct {
						BooksByGenre []StatItem `json:"bookByGenre"`
				} `json:"userBkGenres"`
				UserTags struct {
						UserTags []StatItem `json:"userTags"`
				} `json:"userTags"`
				UserAuthors struct {
						BooksByAuthor []StatItem `json:"booksByAuthor"`
				} `json:"userAuthors"`
		} `json:"homepageStats"`
	}

	// Pre unmarshal data logging
	hpd.logger.Debug("pre-unmarshal data",
    "jsonPreview", string(data[4:min(len(data), 104)]), // First 100 chars after length prefix
		"dataSize", len(data)-4,
	)

	// Unmarshal JSON portion
	if err := json.Unmarshal(data[4:], &temp); err != nil {
			hpd.logger.Error("json unmarshal failed",
					"error", err,
					"jsonSize", claimedLength,
			)
			return fmt.Errorf("json unmarshal failed: %w", err)
	}

    // 9. Initialize nil slices/maps
    if temp.Books == nil {
			temp.Books = make([]repository.Book, 0)
		}
		if temp.HomePageStats.UserBkLang.BooksByLang == nil {
				temp.HomePageStats.UserBkLang.BooksByLang = make([]StatItem, 0)
		}
		if temp.HomePageStats.UserBkGenre.BooksByGenre == nil {
				temp.HomePageStats.UserBkGenre.BooksByGenre = make([]StatItem, 0)
		}
		if temp.HomePageStats.UserTags.UserTags == nil {
				temp.HomePageStats.UserTags.UserTags = make([]StatItem, 0)
		}
		if temp.HomePageStats.UserAuthors.BooksByAuthor == nil {
				temp.HomePageStats.UserAuthors.BooksByAuthor = make([]StatItem, 0)
		}

    // 10. Validate books
    for i, book := range temp.Books {
			if book.Title == "" {
					hpd.logger.Error("book has empty title",
							"bookIndex", i,
					)
					return fmt.Errorf("book at index %d has empty title", i)
			}
		}

    // 11. Assign values
    hpd.Books = temp.Books
    hpd.BooksByFormat = FormatCountStats{
        Physical:  temp.BooksByFormat.Physical,
        Digital:   temp.BooksByFormat.Digital,
        AudioBook: temp.BooksByFormat.AudioBook,
    }
    hpd.HomePageStats = HomePageStats{
        UserBkLang:  LanguageStats{BooksByLang: temp.HomePageStats.UserBkLang.BooksByLang},
        UserBkGenre: GenreStats{BooksByGenre: temp.HomePageStats.UserBkGenre.BooksByGenre},
        UserTags:    TagStats{UserTags: temp.HomePageStats.UserTags.UserTags},
        UserAuthors: AuthorStats{BooksByAuthor: temp.HomePageStats.UserAuthors.BooksByAuthor},
    }

		// 12. Final validation
		if err := hpd.Validate(); err != nil {
			hpd.logger.Error("validation failed after unmarshal",
					"error", err,
			)
			return fmt.Errorf("validation failed after unmarshal: %w", err)
		}

    hpd.logger.Debug("binary unmarshaling completed",
        "jsonSize", claimedLength,
        "totalSize", len(data),
    )

	return nil
}
