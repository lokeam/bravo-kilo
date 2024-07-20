package utils

import (
	"errors"
	"log/slog"
	"strings"
)

// Book struct to hold simplified structure for Books
type Book struct {
	Authors     []string    `json:"authors"`
	ImageLinks  []string    `json:"imageLinks"`
	Title       string      `json:"title"`
	Subtitle    string      `json:"subtitle"`
	Details     BookDetails `json:"details"`
}

// BookDetails struct to hold product specific info
type BookDetails struct {
	Genres       []string      `json:"categories"`
	Description  string        `json:"description"`
	ISBN10       string        `json:"isbn10"`
	ISBN13       string        `json:"ibsn13"`
	Language     string        `json:"language"`
	PageCount    int           `json:"pageCount"`
	PublishDate  string        `json:"publishDate"`
}

type IndustryID struct {
	Indentifier  string  `json:"identifier"`
	Type         string  `json:"type"`
}

var logger *slog.Logger

func InitLogger(l *slog.Logger) {
	logger = l
}

// Takes Oauth 2 response and splits full name into first and last
func SplitFullName(fullName string) (string, string) {
	parts := strings.Fields(fullName)
	if len(parts) == 0 {
			return "", ""
	}
	firstName := parts[0]
	lastName := ""
	if len(parts) > 1 {
			lastName = strings.Join(parts[1:], " ")
	}
	return firstName, lastName
}

func TransformGoogleBooksResponse(searchResult map[string]interface{}) ([]Book, error) {
	items, ok := searchResult["items"].([]interface{})
	if !ok {
		logger.Error("Invalid response format")
		return nil, errors.New("invalid response format")
	}

	var books []Book

	// Check for book volume data
	for _, item := range items {
		volumeInfo, ok := item.(map[string]interface{})["volumeInfo"].(map[string]interface{})
		if !ok {
			logger.Error("Missing volumeInfo prop in Google API Book item response")
		}

		// Check for authors
		authors := []string{}
		if authorsData, exists := volumeInfo["authors"].([]interface{}); exists {
			for _, author := range authorsData {
				authors = append(authors, author.(string))
			}
		}

		// Check for img links
		imageLinks := []string{}
		if imageLinksData, exists := volumeInfo["imageLinks"].(map[string]interface{}); exists {
			for _, link := range imageLinksData {
				imageLinks = append(imageLinks, link.(string))
			}
		}

		// Check for ISBN
		var isbn10, isbn13 string
		if industryIDsData, exists := volumeInfo["industryIdentifiers"].([]interface{}); exists {
			for _, id := range industryIDsData {
				identifier := id.(map[string]interface{})
				switch identifier["type"].(string) {
				case "ISBN_10":
					isbn10 = identifier["identifier"].(string)
				case "ISBN_13":
					isbn13 = identifier["identifier"].(string)
				}
			}
		}

		// Build out book struct
		book := Book{
			Authors:     authors,
			ImageLinks:  imageLinks,
			Title:       volumeInfo["title"].(string),
			Subtitle:    getStringVal(volumeInfo, "subtitle"),
			Details:     BookDetails{
				Genres:      getStringArrVal(volumeInfo, "categories"),
				Description: getStringVal(volumeInfo, "description"),
				ISBN10:      isbn10,
				ISBN13:      isbn13,
				Language:    getStringVal(volumeInfo, "language"),
				PageCount:   getIntVal(volumeInfo, "pageCount"),
				PublishDate: getStringVal(volumeInfo, "publishedDate"),
			},
		}
		books = append(books, book)
	}
	return books, nil
}
	// Helpers to simply accessing types from map[string]interface{}
	func getStringVal(data map[string]interface{}, key string) string {
		if value, exists := data[key]; exists {
			return value.(string)
		}
		return ""
	}

	// Get string array values for response props
	func getStringArrVal(data map[string]interface{}, key string) []string {
		if value, exists := data[key]; exists {
			var result []string
			for _, item := range value.([]interface{}) {
				result = append(result, item.(string))
			}
			return result
		}
		return []string{}
	}


	// Get int values for any response props
	func getIntVal(data map[string]interface{}, key string) int {
		if value, exists := data[key]; exists {
			return int(value.(float64))
		}
		return 0
	}
