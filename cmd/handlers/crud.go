package handlers

import (
	"bravo-kilo/cmd/middleware"
	"bravo-kilo/internal/data"
	"bravo-kilo/internal/utils"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

// Validate Ownership
func (h *Handlers) ValidateBookOwnership(request *http.Request) (int, int, error) {
	// Extract user ID from JWT
	userID, err := utils.ExtractUserIDFromJWT(request)
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
	isOwner, err := h.models.Book.IsUserBookOwner(userID, bookID)
	if err != nil {
			return 0, 0, fmt.Errorf("error checking book ownership: %w", err)
	}

	if !isOwner {
			return 0, 0, fmt.Errorf("unauthorized")
	}

	return userID, bookID, nil
}

// Get all User Books
func (h *Handlers) HandleGetAllUserBooks(response http.ResponseWriter, request *http.Request) {
	// Set Content Security Policy headers
	utils.SetCSPHeaders(response)

	// Extract user ID from JWT
	userID, err := utils.ExtractUserIDFromJWT(request)
	if err != nil {
		h.logger.Error("Error extracting user ID", "error", err)
		http.Error(response, err.Error(), http.StatusUnauthorized)
		return
	}

	books, err := h.models.Book.GetAllBooksByUserID(userID)
	if err != nil {
		h.logger.Error("Error fetching books", "error", err)
		http.Error(response, "Error fetching books", http.StatusInternalServerError)
		return
	}

	dbResponse := map[string]interface{}{
		"books": books,
		"isInLibrary": true,
	}

	response.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(response).Encode(dbResponse); err != nil {
		http.Error(response, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// Retrieve books by a specific author
func (h *Handlers) HandleGetBooksByAuthors(response http.ResponseWriter, request *http.Request) {
	// Set Content Security Policy headers
	utils.SetCSPHeaders(response)

	// Grab token from cookie
	userID, err := utils.ExtractUserIDFromJWT(request)
	if err != nil {
		h.logger.Error("Error extracting user ID", "error", err)
		http.Error(response, err.Error(), http.StatusUnauthorized)
		return
	}
	h.logger.Info("Valid user ID received from token", "userID", userID)

	booksByAuthors, err := h.models.Book.GetAllBooksByAuthors(userID)
	if err != nil {
			h.logger.Error("Error fetching books by authors", "error", err)
			http.Error(response, "Error fetching books by authors", http.StatusInternalServerError)
			return
	}

	response.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(response).Encode(booksByAuthors); err != nil {
			http.Error(response, "Error encoding response", http.StatusInternalServerError)
	}
}

// Get Single Book by ID
func (h *Handlers) HandleGetBookByID(response http.ResponseWriter, request *http.Request) {
	// Set Content Security Policy headers
	utils.SetCSPHeaders(response)

	// Extract and ignore userID from JWT
	_, err := utils.ExtractUserIDFromJWT(request)
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
	book, err := h.models.Book.GetBookByID(bookID)
	if err != nil {
		h.logger.Error("Error fetching book", "error", err)
		http.Error(response, "Error fetching book", http.StatusInternalServerError)
		return
	}

	// Get formats using context from request
	formats, err := h.models.Book.GetFormats(request.Context(), bookID)
	if err != nil {
		h.logger.Error("Error fetching formats", "error", err)
		http.Error(response, "Error fetching formats", http.StatusInternalServerError)
		return
	}
	book.Formats = formats

	response.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(response).Encode(map[string]interface{}{"book": book}); err != nil {
		http.Error(response, "Error encoding response", http.StatusInternalServerError)
	}
}


// Get a Single Book's ID by title
func (h *Handlers) HandleGetBookIDByTitle(response http.ResponseWriter, request *http.Request) {
	// Extract and ignore user ID from JWT
	_, err := utils.ExtractUserIDFromJWT(request)
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
	bookID, err := h.models.Book.GetBookIdByTitle(bookTitle)
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
func (h *Handlers) HandleInsertBook(response http.ResponseWriter, request *http.Request) {
	// Register custom validations
	h.validate.RegisterValidation("isbn10", validateISBN10)
	h.validate.RegisterValidation("isbn13", validateISBN13)

	// Grab book data
	var book data.Book
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


// Helper fn for sanitization
// sanitizeAndUnescape is a helper function to sanitize input and then unescape HTML entities.
func (h *Handlers) sanitizeAndUnescape(input string) string {
	sanitized := h.sanitizer.Sanitize(input)
	return html.UnescapeString(sanitized)
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
func (h *Handlers) HandleUpdateBook(response http.ResponseWriter, request *http.Request) {
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
	var book data.Book
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

	// Sanitize input
	book.Title = h.sanitizer.Sanitize(book.Title)
	book.Subtitle = h.sanitizer.Sanitize(book.Subtitle)
	book.Description = h.sanitizer.Sanitize(book.Description)
	book.Language = h.sanitizer.Sanitize(book.Language)
	book.ImageLink = h.sanitizer.Sanitize(book.ImageLink)
	book.Notes = h.sanitizer.Sanitize(book.Notes)
	book.ISBN10 = h.sanitizer.Sanitize(book.ISBN10)
	book.ISBN13 = h.sanitizer.Sanitize(book.ISBN13)
	for i, author := range book.Authors {
		book.Authors[i] = h.sanitizer.Sanitize(author)
	}
	for i, genre := range book.Genres {
		book.Genres[i] = h.sanitizer.Sanitize(genre)
	}
	for i, tag := range book.Tags {
		book.Tags[i] = h.sanitizer.Sanitize(tag)
	}
	for i, format := range book.Formats {
		book.Formats[i] = h.sanitizer.Sanitize(format)
	}

	// Update the book
	err = h.models.Book.Update(book)
	if err != nil {
		h.logger.Error("Error updating book", "error", err)
		http.Error(response, "Error updating book", http.StatusInternalServerError)
		return
	}

	// Fetch current formats using context from request
	currentFormats, err := h.models.Book.GetFormats(request.Context(), book.ID)
	if err != nil {
		h.logger.Error("Error fetching current formats", "error", err)
		http.Error(response, "Error fetching current formats", http.StatusInternalServerError)
		return
	}

	// Determine formats to remove
	formatsToRemove := utils.FindDifference(currentFormats, book.Formats)

	// Remove specific format associations
	if len(formatsToRemove) > 0 {
		err = h.models.Book.RemoveSpecificFormats(request.Context(), book.ID, formatsToRemove)
		if err != nil {
			h.logger.Error("Error removing specific formats", "error", err)
			http.Error(response, "Error removing specific formats", http.StatusInternalServerError)
			return
		}
	}

	// Insert new formats and their associations
	for _, formatType := range book.Formats {
		formatID, err := h.models.Format.Insert(formatType)
		if err != nil {
			h.logger.Error("Error inserting format", "error", err)
			http.Error(response, "Error inserting format", http.StatusInternalServerError)
			return
		}

		// Wrap formatID in a slice and pass to AddFormats
		if err := h.models.Book.AddFormats(request.Context(), book.ID, []int{formatID}); err != nil {
			h.logger.Error("Error adding format association", "error", err)
			http.Error(response, "Error adding format association", http.StatusInternalServerError)
			return
		}
	}

	response.WriteHeader(http.StatusOK)
	json.NewEncoder(response).Encode(map[string]string{"message": "Book updated successfully"})
}

// Delete Book
func (h *Handlers) HandleDeleteBook(response http.ResponseWriter, request *http.Request) {
	// Validate book ownership
	_, bookID, err := h.ValidateBookOwnership(request)
	if err != nil {
		h.logger.Error("Book ownership validation failed", "error", err)
		http.Error(response, err.Error(), http.StatusUnauthorized)
		return
	}

	// Delete book
	err = h.models.Book.Delete(bookID)
	if err != nil {
		h.logger.Error("Error deleting book", "error", err)
		http.Error(response, "Error deleting book", http.StatusInternalServerError)
		return
	}

	response.WriteHeader(http.StatusOK)
	json.NewEncoder(response).Encode(map[string]string{"message": "Book deleted successfully"})
}

// Sorting - Get Books by Format
func (h *Handlers) HandleGetBooksByFormat(response http.ResponseWriter, request *http.Request) {
	// Set Content Security Policy headers
	utils.SetCSPHeaders(response)

	// Extract user ID from JWT
	userID, err := utils.ExtractUserIDFromJWT(request)
	if err != nil {
		h.logger.Error("Error extracting user ID", "error", err)
		http.Error(response, err.Error(), http.StatusUnauthorized)
		return
	}

	// Get books by format
	booksByFormat, err := h.models.Book.GetAllBooksByFormat(userID)
	if err != nil {
		h.logger.Error("Error fetching books by format", "error", err)
		http.Error(response, "Error fetching books by format", http.StatusInternalServerError)
		return
	}

	// h.logger.Info("Books fetched successfully", "booksByFormat", booksByFormat)

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(booksByFormat)
}

// Sorting - Get Books by Genre
func (h *Handlers) HandleGetBooksByGenres(response http.ResponseWriter, request *http.Request) {
	// Set Content Security Policy headers
	utils.SetCSPHeaders(response)

	// Extract user ID from JWT
	userID, err := utils.ExtractUserIDFromJWT(request)
	if err != nil {
		h.logger.Error("Error extracting user ID", "error", err)
		http.Error(response, err.Error(), http.StatusUnauthorized)
		return
	}
	h.logger.Info("Valid user ID received from token", "userID", userID)

	// Get the request context and pass it to GetAllBooksByGenres
	booksByGenres, err := h.models.Book.GetAllBooksByGenres(request.Context(), userID)
	if err != nil {
		h.logger.Error("Error fetching books by genres", "error", err)
		http.Error(response, "Error fetching books by genres", http.StatusInternalServerError)
		return
	}

	// h.logger.Info("Books fetched successfully", "booksByGenres", booksByGenres)

	response.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(response).Encode(booksByGenres); err != nil {
		http.Error(response, "Error encoding response", http.StatusInternalServerError)
	}
}

// Build Homepage Analytics Data Response
func (h *Handlers) HandleGetHomepageData(response http.ResponseWriter, request *http.Request) {
	// Set Content Security Policy headers
	utils.SetCSPHeaders(response)

	// Extract user ID from JWT
	userID, err := utils.ExtractUserIDFromJWT(request)
	if err != nil {
		h.logger.Error("Error extracting user ID", "error", err)
		http.Error(response, err.Error(), http.StatusUnauthorized)
		return
	}

	// Channels to receive data from goroutines
	userTagsChan := make(chan map[string]interface{}, 1)
	userLangChan := make(chan map[string]interface{}, 1)
	userGenresChan := make(chan map[string]interface{}, 1)
	errorChan := make(chan error, 1)

	// Goroutine for GetUserTags
	go func() {
		tags, err := h.models.Book.GetUserTags(request.Context(), userID)
		if err != nil {
			errorChan <- err
			return
		}
		userTagsChan <- tags
	}()

	// Goroutine for GetBooksByLanguage
	go func() {
		langs, err := h.models.Book.GetBooksByLanguage(request.Context(), userID)
		if err != nil {
			errorChan <- err
			return
		}
		userLangChan <- langs
	}()

	// Goroutine for GetBooksListByGenre
	go func() {
		genres, err := h.models.Book.GetBooksListByGenre(request.Context(), userID)
		if err != nil {
			errorChan <- err
			return
		}
		userGenresChan <- genres
	}()

	// Collect results from channels
	var userTags, userBkLang, userBkGenres interface{}

	for i := 0; i < 3; i++ {
		select {
		case tags := <-userTagsChan:
			userTags = tags
		case langs := <-userLangChan:
			userBkLang = langs
		case genres := <-userGenresChan:
			userBkGenres = genres
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
	}

	response.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(response).Encode(responseData); err != nil {
		http.Error(response, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
