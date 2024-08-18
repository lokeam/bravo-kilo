package handlers

import (
	"bravo-kilo/cmd/middleware"
	"bravo-kilo/internal/data"
	"bravo-kilo/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
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

	h.logger.Info("GetBookIDByTitle handler, bookID: ", "bookID", bookID)

	response.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(response).Encode(map[string]interface{}{"bookID": bookID}); err != nil {
			http.Error(response, "Error encoding response", http.StatusInternalServerError)
	}
}

// Add Book
func (h *Handlers) HandleInsertBook(response http.ResponseWriter, request *http.Request) {
	// Grab book data
	var book data.Book
	err := json.NewDecoder(request.Body).Decode(&book)
	if err != nil {
		h.logger.Error("Error decoding book data", "error", err)
		http.Error(response, "Error decoding book data - invalid input", http.StatusBadRequest)
		return
	}

	// Retrieve user ID from context
	userID, ok := middleware.GetUserID(request.Context())
	if !ok {
		h.logger.Error("User ID not found in context")
		http.Error(response, "User ID not found", http.StatusInternalServerError)
		return
	}

	// Insert the book and associate it with the user
	bookID, err := h.models.Book.InsertBook(book, userID)
	if err != nil {
		h.logger.Error("Error inserting book", "error", err)
		http.Error(response, "Error inserting book", http.StatusInternalServerError)
		return
	}

	// Insert formats and their associations
	for _, formatType := range book.Formats {
		formatID, err := h.models.Format.Insert(bookID, formatType)
		if err != nil {
			h.logger.Error("Error inserting format", "error", err)
			http.Error(response, "Error inserting format", http.StatusInternalServerError)
			return
		}

		// Pass the context to the AddFormats method and wrap formatID in a slice
		if err := h.models.Book.AddFormats(request.Context(), bookID, []int{formatID}); err != nil {
			h.logger.Error("Error adding format association", "error", err)
			http.Error(response, "Error adding format association", http.StatusInternalServerError)
			return
		}
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(map[string]int{"book_id": bookID})
}

// Update Book
func (h *Handlers) HandleUpdateBook(response http.ResponseWriter, request *http.Request) {
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
		formatID, err := h.models.Format.Insert(bookID, formatType)
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
