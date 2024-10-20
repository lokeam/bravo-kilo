package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/lokeam/bravo-kilo/cmd/middleware"
	authhandlers "github.com/lokeam/bravo-kilo/internal/auth/handlers"
	"github.com/lokeam/bravo-kilo/internal/books/repository"
	"github.com/lokeam/bravo-kilo/internal/shared/utils"
	"golang.org/x/oauth2"
)

type SearchHandlers struct {
	logger        *slog.Logger
	bookRepo      repository.BookRepository
	bookCache     repository.BookCache
	authHandlers  *authhandlers.AuthHandlers
}

func NewSearchHandlers(
	logger *slog.Logger,
	bookRepo repository.BookRepository,
	bookCache repository.BookCache,
	authHandlers *authhandlers.AuthHandlers,
	) (*SearchHandlers, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	if bookRepo == nil {
		return nil, fmt.Errorf("failed to initialize bookRepo")
	}

	if bookCache == nil {
		return nil, fmt.Errorf("failed to initialize bookCache")
	}

	return &SearchHandlers{
		logger:   logger,
		bookRepo: bookRepo,
		bookCache: bookCache,
		authHandlers: authHandlers,
	}, nil
}


// Helper fn to check for empty fields in GBooks Response
func checkEmptyFields(book repository.Book) (bool, []string) {
	emptyFields := []string{}

	if book.Title == "" {
		emptyFields = append(emptyFields, "Title")
	}
	if book.PublishDate == "" {
		emptyFields = append(emptyFields, "Publish date")
	}
	if book.ISBN10 == "" {
		emptyFields = append(emptyFields, "ISBN-10")
	}
	if book.ISBN13 == "" {
		emptyFields = append(emptyFields, "ISBN-13")
	}

	hasEmptyFields := len(emptyFields) > 0

	return hasEmptyFields, emptyFields
}

// Format Google Books Response
func (h *SearchHandlers) formatGoogleBooksResponse(response http.ResponseWriter, booksData interface{}) []repository.Book {
	var gBooksResponse []repository.Book

	// Ensure that the data is correctly cast to the expected format
	dataMap, ok := booksData.(map[string]interface{})
	if !ok {
		h.logger.Error("Invalid books data format")
		http.Error(response, "Invalid books data format", http.StatusInternalServerError)
		return nil
	}

	items, ok := dataMap["items"].([]interface{})
	if !ok {
		h.logger.Warn("No items in books data")
		return gBooksResponse // Return an empty slice if no items are found
	}

	for _, item := range items {
		volumeInfo, ok := item.(map[string]interface{})["volumeInfo"].(map[string]interface{})
		if !ok {
			h.logger.Warn("Invalid volumeInfo format", "item", item)
			continue // Skip items with invalid format
		}

		// Use utility functions to safely retrieve string and integer values with defaults
		formattedBook := repository.Book{
			Title:       utils.GetStringValOrDefault(volumeInfo, "title", ""),
			Subtitle:    utils.GetStringValOrDefault(volumeInfo, "subtitle", ""),
			Description: utils.GetStringValOrDefault(volumeInfo, "description", ""),
			Language:    utils.GetStringValOrDefault(volumeInfo, "language", ""),
			PageCount:   utils.GetIntValOrDefault(volumeInfo, "pageCount", 0),
			PublishDate: utils.GetStringValOrDefault(volumeInfo, "publishedDate", ""),
		}

		// Handle image link, selecting the largest available thumbnail
		if imageLinks, ok := volumeInfo["imageLinks"].(map[string]interface{}); ok {
			if largeThumbnail, ok := imageLinks["thumbnail"].(string); ok {
					formattedBook.ImageLink = utils.CleanImageLink(largeThumbnail)
			} else if smallThumbnail, ok := imageLinks["smallThumbnail"].(string); ok {
					formattedBook.ImageLink = utils.CleanImageLink(smallThumbnail)
			} else {
					formattedBook.ImageLink = "" // Handle cases where no image link is available
			}
		}


		// Handle ISBN numbers
		if industryIdentifiers, ok := volumeInfo["industryIdentifiers"].([]interface{}); ok {
			for _, id := range industryIdentifiers {
				if identifier, ok := id.(map[string]interface{}); ok {
					if utils.GetStringValOrDefault(identifier, "type", "") == "ISBN_13" {
						formattedBook.ISBN13 = utils.GetStringValOrDefault(identifier, "identifier", "")
					}
					if utils.GetStringValOrDefault(identifier, "type", "") == "ISBN_10" {
						formattedBook.ISBN10 = utils.GetStringValOrDefault(identifier, "identifier", "")
					}
				}
			}
		}

		// Handle genres, ensuring it's always an array
		formattedBook.Genres = []string{}
		if categories, ok := volumeInfo["categories"].([]interface{}); ok {
			for _, category := range categories {
				if categoryStr, ok := category.(string); ok {
					formattedBook.Genres = append(formattedBook.Genres, categoryStr)
				}
			}
		}

		// Handle authors, ensuring it's always an array
		formattedBook.Authors = []string{}
		if authors, ok := volumeInfo["authors"].([]interface{}); ok {
			for _, author := range authors {
				if authorStr, ok := author.(string); ok {
					formattedBook.Authors = append(formattedBook.Authors, authorStr)
				}
			}
		}

		// Check for empty fields
		hasEmptyFields, emptyFields := checkEmptyFields(formattedBook)
		formattedBook.HasEmptyFields = hasEmptyFields
		formattedBook.EmptyFields = emptyFields

		gBooksResponse = append(gBooksResponse, formattedBook)
	}

	h.logger.Info("Formatted books", "books", gBooksResponse)
	return gBooksResponse
}

// Process Google Books API Search
func (h *SearchHandlers) HandleSearchBooks(response http.ResponseWriter, request *http.Request) {
	query := request.URL.Query().Get("query")
	if query == "" {
			http.Error(response, "Query parameter required in request", http.StatusBadRequest)
			return
	}

	// Debug - Set headers to prevent caching
	response.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	response.Header().Set("Pragma", "no-cache")
	response.Header().Set("Expires", "0")


	// Get userID from context
	userID, ok := request.Context().Value(middleware.UserIDKey).(int)
	if !ok {
			h.logger.Error("Failed to get userID from context")
			http.Error(response, "Unauthorized", http.StatusUnauthorized)
			return
	}

	// Get user's access token
	accessToken, err := h.authHandlers.GetUserAccessToken(request)
	if err != nil {
		h.logger.Error("Error retrieving user access token", "error", err)
		if err == authhandlers.ErrNoRefreshToken {
				http.Error(response, "No valid refresh token found", http.StatusUnauthorized)
		} else {
				http.Error(response, "Error retrieving access token", http.StatusInternalServerError)
		}
		return
	}

	// Use the access token to call the Google Books API
	client := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(accessToken))
	googleBooksURL := fmt.Sprintf("https://www.googleapis.com/books/v1/volumes?q=%s&maxResults=35", url.QueryEscape(query))

	h.logger.Info("Requesting Google Books API", "url", googleBooksURL)

	booksResponse, err := client.Get(googleBooksURL)
	if err != nil {
			h.logger.Error("Error calling Google Books API", "error", err)
			http.Error(response, "Error calling Google Books API", http.StatusInternalServerError)
			return
	}
	defer booksResponse.Body.Close()

	if booksResponse.StatusCode != http.StatusOK {
			var errorResponse map[string]interface{}
			json.NewDecoder(booksResponse.Body).Decode(&errorResponse)
			h.logger.Error("Google Books API responded with non-OK status", "status", booksResponse.StatusCode, "body", errorResponse)
			http.Error(response, "Google Books API error", booksResponse.StatusCode)
			return
	}

	// Decode the Google Books API response
	var booksData interface{}
	if err := json.NewDecoder(booksResponse.Body).Decode(&booksData); err != nil {
			h.logger.Error("Error decoding Google Books API response", "error", err)
			http.Error(response, "Error decoding response", http.StatusInternalServerError)
			return
	}

	// Format the books response
	formattedBooks := h.formatGoogleBooksResponse(response, booksData)
	// h.logger.Info("---------------")
	// h.logger.Info("Showing formattedBooks, pre-check:", "formattedBooks", formattedBooks)

	// Create hash sets for user's existing library data
	isbn10Set, err := h.bookCache.GetAllBooksISBN10(userID)
	if err != nil {
			h.logger.Error("Error retrieving user's ISBN10", "error", err)
			http.Error(response, "Error retrieving user's ISBN10", http.StatusInternalServerError)
			return
	}

	isbn13Set, err := h.bookCache.GetAllBooksISBN13(userID)
	if err != nil {
			h.logger.Error("Error retrieving user's ISBN13", "error", err)
			http.Error(response, "Error retrieving user's ISBN13", http.StatusInternalServerError)
			return
	}

	// Debug - temporarily remove book title check
	// titleSet, err := h.models.Book.GetAllBooksTitles(userID)
	// if err != nil {
	// 		h.logger.Error("Error retrieving user's book titles", "error", err)
	// 		http.Error(response, "Error retrieving user's book titles", http.StatusInternalServerError)
	// 		return
	// }

    // Fetch the user's book publish dates as a slice of BookInfo structs
    bookList, err := h.bookCache.GetAllBooksPublishDate(userID)
    if err != nil {
        h.logger.Error("Error retrieving user's book publish dates", "error", err)
        http.Error(response, "Error retrieving user's book publish dates", http.StatusInternalServerError)
        return
    }

	// Check bookMap
	h.logger.Info("======================================")
	fmt.Println("Checking bookmap: ", bookList)

    // Helper function to check if a book is in the user's library
  bookExistsInLibrary := func(title, publishDate string) bool {
		for _, book := range bookList {
			if book.Title == title && book.PublishDate == publishDate {
				return true
			}
		}
		return false
	}

	// Check each book against the user's library
	for i := range formattedBooks {
		formattedBook := &formattedBooks[i]
		isInLibrary := (isbn10Set.Has(formattedBook.ISBN10) || isbn13Set.Has(formattedBook.ISBN13)) &&
				bookExistsInLibrary(formattedBook.Title, formattedBook.PublishDate)

		formattedBook.IsInLibrary = isInLibrary
	}

	// h.logger.Info("===================")
	// h.logger.Info("Showing formattedBooks, post check, about to send:", "formattedBooks", formattedBooks)

	dbResponse := map[string]interface{}{
		"books": formattedBooks,
		"isSearchPage": true,
	}

	response.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(response).Encode(dbResponse); err != nil {
			h.logger.Error("Error encoding response", "error", err)
			http.Error(response, "Error encoding response", http.StatusInternalServerError)
	}
}
