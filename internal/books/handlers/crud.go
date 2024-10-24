package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/lokeam/bravo-kilo/cmd/middleware"
	"github.com/lokeam/bravo-kilo/config"
	"github.com/lokeam/bravo-kilo/internal/books/repository"
	"github.com/lokeam/bravo-kilo/internal/shared/jwt"
	"github.com/lokeam/bravo-kilo/internal/shared/utils"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

// Validate Ownership
func (h *BookHandlers) ValidateBookOwnership(request *http.Request) (int, int, error) {
	// Extract user ID from JWT
	userID, err := jwt.ExtractUserIDFromJWT(request, config.AppConfig.JWTPublicKey)
	if err != nil {
			return 0, 0, fmt.Errorf("unauthorized: %w", err)
	}

	// Ensure book ID is provided in the URL and parse
	bookIDStr := chi.URLParam(request, "bookID")
	bookID, err := strconv.Atoi(bookIDStr)
	if err != nil {
			return 0, 0, fmt.Errorf("invalid book ID: %w", err)
	}

	// Check if the book is associated user in the user_books
	isOwner, err := h.bookRepo.IsUserBookOwner(userID, bookID)
	if err != nil {
			return 0, 0, fmt.Errorf("error checking book ownership: %w", err)
	}

	if !isOwner {
			return 0, 0, fmt.Errorf("unauthorized")
	}

	return userID, bookID, nil
}

// Get all User Books
func (h *BookHandlers) HandleGetAllUserBooks(response http.ResponseWriter, request *http.Request) {
	h.logger.Info("HandleGetAllUserBooks called")

	// Set Content Security Policy headers
	utils.SetCSPHeaders(response)

	// Extract user ID from JWT
	userID, err := jwt.ExtractUserIDFromJWT(request, config.AppConfig.JWTPublicKey)
	if err != nil {
		h.logger.Error("Error extracting user ID", "error", err)
		http.Error(response, err.Error(), http.StatusUnauthorized)
		return
	}
	h.logger.Info("User ID extracted successfully", "userID", userID)

	books, err := h.bookRepo.GetAllBooksByUserID(userID)
	if err != nil {
		h.logger.Error("Error fetching books", "error", err)
		http.Error(response, "Error fetching books", http.StatusInternalServerError)
		return
	}
	h.logger.Info("Books fetched successfully", "count", len(books))

	// Reverse Normalize Book Data
	h.bookService.ReverseNormalizeBookData(&books)

	dbResponse := map[string]interface{}{
		"books": books,
		"isInLibrary": true,
	}

	response.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(response).Encode(dbResponse); err != nil {
		h.logger.Error("Error encoding response", "error", err)
		http.Error(response, "Error encoding response", http.StatusInternalServerError)
		return
	}
	//h.logger.Info("HandleGetAllUserBooks completed successfully")
}

// Retrieve books by a specific author
func (h *BookHandlers) HandleGetBooksByAuthors(response http.ResponseWriter, request *http.Request) {
	h.logger.Info("HandleGetBooksByAuthors called")

	// Set Content Security Policy headers
	utils.SetCSPHeaders(response)

	// Grab token from cookie
	userID, err := jwt.ExtractUserIDFromJWT(request, config.AppConfig.JWTPublicKey)
	if err != nil {
		h.logger.Error("Error extracting user ID", "error", err)
		http.Error(response, err.Error(), http.StatusUnauthorized)
		return
	}
	h.logger.Info("Valid user ID received from token", "userID", userID)

	booksByAuthors, err := h.authorRepo.GetAllBooksByAuthors(userID)
	if err != nil {
			h.logger.Error("Error fetching books by authors", "error", err)
			http.Error(response, "Error fetching books by authors", http.StatusInternalServerError)
			return
	}
	//h.logger.Info("Books by authors fetched successfully", "authorCount", len(booksByAuthors))

	response.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(response).Encode(booksByAuthors); err != nil {
			h.logger.Error("Error encoding response", "error", err)
			http.Error(response, "Error encoding response", http.StatusInternalServerError)
	}
	//h.logger.Info("HandleGetBooksByAuthors completed successfully")
}

// Get Single Book by ID
func (h *BookHandlers) HandleGetBookByID(response http.ResponseWriter, request *http.Request) {
	// Set Content Security Policy headers
	utils.SetCSPHeaders(response)

	// Extract and ignore userID from JWT
	_, err := jwt.ExtractUserIDFromJWT(request, config.AppConfig.JWTPublicKey)
	if err != nil {
		h.logger.Error("Error extracting user ID", "error", err)
		http.Error(response, err.Error(), http.StatusUnauthorized)
		return
	}

	bookIDStr := chi.URLParam(request, "bookID")
	bookID, err := strconv.Atoi(bookIDStr)

	if err != nil {
		http.Error(response, "Invalid book ID", http.StatusBadRequest)
		return
	}

	// Fetch the book by ID
	book, err := h.bookRepo.GetBookByID(bookID)
	if err != nil {
		h.logger.Error("Error fetching book", "error", err)
		http.Error(response, "Error fetching book", http.StatusInternalServerError)
		return
	}

	caser := cases.Title(language.Und)
	book.Title = caser.String(book.Title)

	response.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(response).Encode(map[string]interface{}{"book": book}); err != nil {
		http.Error(response, "Error encoding response", http.StatusInternalServerError)
	}
}


// Get a Single Book's ID by title
func (h *BookHandlers) HandleGetBookIDByTitle(response http.ResponseWriter, request *http.Request) {
	// Extract and ignore user ID from JWT
	_, err := jwt.ExtractUserIDFromJWT(request, config.AppConfig.JWTPublicKey)
	if err != nil {
		h.logger.Error("Error extracting userID", "error", err)
		http.Error(response, err.Error(), http.StatusUnauthorized)
		return
	}

	// Get book title from query parameter
	bookTitle := request.URL.Query().Get("title")
	if bookTitle == "" {
			http.Error(response, "Book title is required", http.StatusBadRequest)
			return
	}

	// Retrieve the book ID by title
	bookID, err := h.bookRepo.GetBookIdByTitle(bookTitle)
	if err != nil {
			h.logger.Error("Error fetching book ID by title", "error", err)
			http.Error(response, "Error fetching book ID", http.StatusInternalServerError)
			return
	}

	if bookID == 0 {
			http.Error(response, "Book not found", http.StatusNotFound)
			return
	}

	response.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(response).Encode(map[string]interface{}{"bookID": bookID}); err != nil {
			http.Error(response, "Error encoding response", http.StatusInternalServerError)
	}
}

// Add Book
func (h *BookHandlers) HandleInsertBook(response http.ResponseWriter, request *http.Request) {
	// Register custom validations
	h.validate.RegisterValidation("isbn10", validateISBN10)
	h.validate.RegisterValidation("isbn13", validateISBN13)

	// Grab book data
	var book repository.Book
	err := json.NewDecoder(request.Body).Decode(&book)
	if err != nil {
			h.logger.Error("Error decoding book data", "error", err)
			http.Error(response, "Error decoding book data - invalid input", http.StatusBadRequest)
			return
	}

	// Authenticate
	userID, ok := middleware.GetUserID(request.Context())
	if !ok {
			http.Error(response, "User ID not found", http.StatusUnauthorized)
			return
	}

	// Call service create a book entry then insert the book
	bookID, err := h.bookService.CreateBookEntry(request.Context(), book, userID)
	if err != nil {
			http.Error(response, "Error inserting book", http.StatusInternalServerError)
			return
	}

	// Send response back to FE
	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(map[string]int{"book_id": bookID})
}


// Helper fns for validation
func validateISBN10(fl validator.FieldLevel) bool {
	isbn := fl.Field().String()
	if len(isbn) != 10 {
		return false
	}
	sum := 0
	for i := 0; i < 9; i++ {
		digit := int(isbn[i] - '0')
		if digit < 0 || digit > 9 {
			return false
		}
		sum += digit * (10 - i)
	}
	checksum := (11 - sum%11) % 11
	return (checksum == 10 && isbn[9] == 'X') || (int(isbn[9]-'0') == checksum)
}

func validateISBN13(fl validator.FieldLevel) bool {
	isbn := fl.Field().String()
	if len(isbn) != 13 {
		return false
	}
	sum := 0
	for i := 0; i < 12; i++ {
		digit := int(isbn[i] - '0')
		if digit < 0 || digit > 9 {
			return false
		}
		if i%2 == 0 {
			sum += digit
		} else {
			sum += 3 * digit
		}
	}
	checksum := (10 - sum%10) % 10
	return int(isbn[12]-'0') == checksum
}

// Update Book
func (h *BookHandlers) HandleUpdateBook(response http.ResponseWriter, request *http.Request) {
	// Register custom validations
	h.validate.RegisterValidation("isbn10", validateISBN10)
	h.validate.RegisterValidation("isbn13", validateISBN13)

	// Validate book ownership
	_, bookID, err := h.ValidateBookOwnership(request)
	if err != nil {
			h.logger.Error("Validation failed", "error", err)
			http.Error(response, err.Error(), http.StatusUnauthorized)
			return
	}

	// Grab book data from request body
	var book repository.Book
	err = json.NewDecoder(request.Body).Decode(&book)
	if err != nil {
			h.logger.Error("Error decoding book data", "error", err)
			http.Error(response, "Error decoding book data - invalid input", http.StatusBadRequest)
			return
	}

	book.ID = bookID

	// Validate struct
	err = h.validate.Struct(book)
	if err != nil {
			h.logger.Error("Validation error", "error", err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
	}

	// Update the book
	err = h.bookUpdater.UpdateBookEntry(request.Context(), book)
	if err != nil {
			h.logger.Error("Error updating book", "error", err)
			http.Error(response, "Error updating book", http.StatusInternalServerError)
			return
	}

	response.WriteHeader(http.StatusOK)
	json.NewEncoder(response).Encode(map[string]string{"message": "Book updated successfully"})
}


// Delete Book
func (h *BookHandlers) HandleDeleteBook(response http.ResponseWriter, request *http.Request) {
	// Validate book ownership
	_, bookID, err := h.ValidateBookOwnership(request)
	if err != nil {
		h.logger.Error("Book ownership validation failed", "error", err)
		http.Error(response, err.Error(), http.StatusUnauthorized)
		return
	}

	// Delete book
	err = h.bookDeleter.Delete(bookID)
	if err != nil {
		h.logger.Error("Error deleting book", "error", err)
		http.Error(response, "Error deleting book", http.StatusInternalServerError)
		return
	}

	response.WriteHeader(http.StatusOK)
	json.NewEncoder(response).Encode(map[string]string{"message": "Book deleted successfully"})
}

// Sorting - Get Books by Format
func (h *BookHandlers) HandleGetBooksByFormat(response http.ResponseWriter, request *http.Request) {
	h.logger.Info("HandleGetBooksByFormat called")

	// Set Content Security Policy headers
	utils.SetCSPHeaders(response)

	// Extract user ID from JWT
	userID, err := jwt.ExtractUserIDFromJWT(request, config.AppConfig.JWTPublicKey)
	if err != nil {
		h.logger.Error("Error extracting user ID", "error", err)
		http.Error(response, err.Error(), http.StatusUnauthorized)
		return
	}
	h.logger.Info("Valid user ID received from token", "userID", userID)

	// Get books by format
	booksByFormat, err := h.formatRepo.GetAllBooksByFormat(userID)
	if err != nil {
		h.logger.Error("Error fetching books by format", "error", err)
		http.Error(response, "Error fetching books by format", http.StatusInternalServerError)
		return
	}
	h.logger.Info("Books by format fetched successfully", "formatCount", len(booksByFormat))

	for format, books := range booksByFormat {
		h.bookService.ReverseNormalizeBookData(&books)
		booksByFormat[format] = books
	}

	response.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(response).Encode(booksByFormat); err != nil {
		h.logger.Error("Error encoding response", "error", err)
		http.Error(response, "Error encoding response", http.StatusInternalServerError)
	}
	//h.logger.Info("HandleGetBooksByFormat completed successfully")
}

// Sorting - Get Books by Genre
func (h *BookHandlers) HandleGetBooksByGenres(response http.ResponseWriter, request *http.Request) {
	h.logger.Info("HandleGetBooksByGenres called")

	// Set Content Security Policy headers
	utils.SetCSPHeaders(response)

	// Extract user ID from JWT
	userID, err := jwt.ExtractUserIDFromJWT(request, config.AppConfig.JWTPublicKey)
	if err != nil {
		h.logger.Error("Error extracting user ID", "error", err)
		http.Error(response, err.Error(), http.StatusUnauthorized)
		return
	}
	h.logger.Info("Valid user ID received from token", "userID", userID)

	// Get the request context and pass it to GetAllBooksByGenres
	booksByGenres, err := h.genreRepo.GetAllBooksByGenres(request.Context(), userID)
	if err != nil {
		h.logger.Error("Error fetching books by genres", "error", err)
		http.Error(response, "Error fetching books by genres", http.StatusInternalServerError)
		return
	}
	//h.logger.Info("Books by genres fetched successfully", "genreCount", len(booksByGenres))

	response.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(response).Encode(booksByGenres); err != nil {
		h.logger.Error("Error encoding response", "error", err)
		http.Error(response, "Error encoding response", http.StatusInternalServerError)
	}
	//h.logger.Info("HandleGetBooksByGenres completed successfully")
}

// Sorting - Get Books by Tag
func (h *BookHandlers) HandleGetBooksByTags(response http.ResponseWriter, request *http.Request) {
	h.logger.Info("HandleGetBooksByTags called")

	// Set Content Security Policy headers
	utils.SetCSPHeaders(response)

	// Extract user ID from JWT
	userID, err := jwt.ExtractUserIDFromJWT(request, config.AppConfig.JWTPublicKey)
	if err != nil {
		h.logger.Error("Error extracting user ID", "error", err)
		http.Error(response, err.Error(), http.StatusUnauthorized)
		return
	}
	h.logger.Info("Valid user ID received from token", "userID", userID)

	// Get the request context and pass it to GetAllBooksByTags
	booksByTags, err := h.tagRepo.GetAllBooksByTags(request.Context(), userID)
	if err != nil {
		h.logger.Error("Error fetching books by tags", "error", err)
		http.Error(response, "Error fetching books by tags", http.StatusInternalServerError)
		return
	}
	//h.logger.Info("Books by tags fetched successfully", "tagCount", len(booksByTags))

	// Set content type and respond with JSON
	response.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(response).Encode(booksByTags); err != nil {
		h.logger.Error("Error encoding books by tags response", "error", err)
		http.Error(response, "Error encoding response", http.StatusInternalServerError)
	}
	//h.logger.Info("HandleGetBooksByTags completed successfully")
}


// Build Homepage Analytics Data Response
func (h *BookHandlers) HandleGetHomepageData(response http.ResponseWriter, request *http.Request) {
	// Set Content Security Policy headers
	utils.SetCSPHeaders(response)

	// Extract user ID from JWT
	userID, err := jwt.ExtractUserIDFromJWT(request, config.AppConfig.JWTPublicKey)
	if err != nil {
		h.logger.Error("Error extracting user ID", "error", err)
		http.Error(response, err.Error(), http.StatusUnauthorized)
		return
	}

	// Channels to receive data from goroutines
	userTagsChan := make(chan map[string]interface{}, 1)
	userLangChan := make(chan map[string]interface{}, 1)
	userGenresChan := make(chan map[string]interface{}, 1)
	userAuthorsChan := make(chan map[string]interface{}, 1)
	errorChan := make(chan error, 1)

	// Goroutine for GetUserTags
	go func() {
		tags, err := h.tagRepo.GetUserTags(request.Context(), userID)
		if err != nil {
			errorChan <- err
			return
		}
		userTagsChan <- tags
	}()

	// Goroutine for GetBooksByLanguage
	go func() {
		langs, err := h.bookCache.GetBooksByLanguage(request.Context(), userID)
		if err != nil {
			errorChan <- err
			return
		}
		userLangChan <- langs
	}()

	// Goroutine for GetBooksListByGenre
	go func() {
		genres, err := h.genreRepo.GetBooksListByGenre(request.Context(), userID)
		if err != nil {
			errorChan <- err
			return
		}

		userGenresChan <- genres
	}()

	// Goroutine for GetAuthorsListWithBookCount
	go func() {
		authors, err := h.authorRepo.GetAuthorsListWithBookCount(request.Context(), userID)
		if err != nil {
			errorChan <- err
			return
		}
		userAuthorsChan <- authors
	}()

	// Collect results from channels
	var userTags, userBkLang, userBkGenres, userAuthors interface{}

	for i := 0; i < 4; i++ {
		select {
		case tags := <-userTagsChan:
			userTags = tags
		case langs := <-userLangChan:
			userBkLang = langs
		case genres := <-userGenresChan:
			userBkGenres = genres
		case authors := <-userAuthorsChan:
			userAuthors = authors
		case err := <-errorChan:
			h.logger.Error("Error fetching homepage data", "error", err)
			http.Error(response, "Error fetching homepage data", http.StatusInternalServerError)
			return
		}
	}

	// Prepare the JSON response
	responseData := map[string]interface{}{
		"userTags":    userTags,
		"userBkLang":  userBkLang,
		"userBkGenres": userBkGenres,
		"userAuthors": userAuthors,
	}

	response.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(response).Encode(responseData); err != nil {
		http.Error(response, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
