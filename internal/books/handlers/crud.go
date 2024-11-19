package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/lokeam/bravo-kilo/cmd/middleware"
	"github.com/lokeam/bravo-kilo/config"
	"github.com/lokeam/bravo-kilo/internal/books/repository"
	"github.com/lokeam/bravo-kilo/internal/shared/jwt"
	"github.com/lokeam/bravo-kilo/internal/shared/redis"
	"github.com/lokeam/bravo-kilo/internal/shared/utils"
	"github.com/lokeam/bravo-kilo/internal/shared/workers"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	goredis "github.com/redis/go-redis/v9"
)

type JSONResponse struct {
	Data        interface{} `json:"data,omitempty"`
	Error       string      `json:"error,omitempty"`
	StatusCode  int         `json:"-"` // Do not include in JSON response
}

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
	requestID := uuid.New().String()
	startTime := time.Now()

	h.logger.Info("=== HandleGetAllUserBooks Trace Start ===", "requestID", requestID)
	defer func() {
		h.logger.Info("=== HandleGetAllUserBooks Trace End ===",
			"requestID", requestID,
			"duration", time.Since(startTime))
	}()

	// Set Content Security Policy headers
	utils.SetCSPHeaders(response)

	// Extract user ID from JWT
	userID, err := jwt.ExtractUserIDFromJWT(request, config.AppConfig.JWTPublicKey)
	if err != nil {
			h.logger.Error("Error extracting user ID",
					"error", err,
					"requestID", requestID)
			http.Error(response, err.Error(), http.StatusUnauthorized)
			return
	}

	// Check Redis client status
	h.logger.Info("Redis client status check",
			"requestID", requestID,
			"status", h.redisClient.GetStatus(),
			"isReady", h.redisClient.IsReady())

	// Define cache key before potential early returns
	cacheKey := fmt.Sprintf("book:%d", userID)

	// Attempt cache retrieval if Redis is ready
	if h.redisClient.IsReady() {
			if breaker := h.redisClient.GetCircuitBreaker(); breaker != nil && breaker.GetState() != redis.StateOpen {
					h.logger.Info("Attempting cache retrieval",
							"requestID", requestID,
							"cacheKey", cacheKey)

					cachedData, err := h.redisClient.Get(request.Context(), cacheKey)
					if err == nil {
							var books []repository.Book
							if err := json.Unmarshal([]byte(cachedData), &books); err == nil {
									h.logger.Info("Cache hit: returning books from cache",
											"requestID", requestID,
											"bookCount", len(books))

									h.sendJSONResponse(response, JSONResponse{
											Data: map[string]interface{}{
													"books": books,
													"source": "cache",
											},
									})
									return
							}
							h.logger.Error("Failed to unmarshal cached data",
									"error", err,
									"requestID", requestID)
					} else if err != goredis.Nil {
							h.logger.Error("Redis operation failed",
									"error", err,
									"requestID", requestID)
					}
			}
	}

	// Database fetch
	h.logger.Info("Attempting database fetch",
			"requestID", requestID,
			"userID", userID)

	books, err := h.bookRepo.GetAllBooksByUserID(userID)
	if err != nil {
			h.logger.Error("Database fetch failed",
					"requestID", requestID,
					"error", err,
					"errorType", fmt.Sprintf("%T", err))
			http.Error(response, "Error fetching books", http.StatusInternalServerError)
			return
	}

	h.logger.Info("Database fetch successful",
			"requestID", requestID,
			"bookCount", len(books))

	// Reverse normalize the book data
	h.logger.Info("Applying reverse normalization",
			"requestID", requestID,
			"bookCount", len(books))
	h.bookService.ReverseNormalizeBookData(&books)

	// Only attempt caching if Redis is ready and circuit breaker allows
	if h.redisClient.IsReady() {
			if breaker := h.redisClient.GetCircuitBreaker(); breaker != nil && breaker.GetState() != redis.StateOpen {
					h.logger.Info("Attempting to cache results",
							"requestID", requestID,
							"bookCount", len(books))

					if booksJSON, err := json.Marshal(books); err == nil {
							cacheDuration := h.redisClient.GetConfig().CacheConfig.BookList
							h.logger.Info("Marshaled data for caching",
									"requestID", requestID,
									"dataSize", len(booksJSON),
									"cacheDuration", cacheDuration)

							if err := h.redisClient.Set(request.Context(), cacheKey, booksJSON, cacheDuration); err != nil {
									h.logger.Error("Cache update failed",
									"error", err,
									"cacheKey", cacheKey,
									"dataSize", len(booksJSON),
									"cacheDuration", cacheDuration,
									"redisStatus", h.redisClient.GetStatus(),
									"circuitState", h.redisClient.GetCircuitBreaker().GetState())
							} else {
									h.logger.Info("Cache update successful", "requestID", requestID)
							}
					}
			}
	}

	// Send response
	h.sendJSONResponse(response, JSONResponse{
			Data: map[string]interface{}{
					"books": books,
					"source": "db",
			},
	})
}

// Domain refactor - GetAllUserBooks
func (h *BookHandlers) GetAllUserBooksDomain(ctx context.Context, userID int) ([]repository.Book, error) {
	// Database fetch
	books, err := h.bookRepo.GetAllBooksByUserID(userID)
	if err != nil {
		h.logger.Error(
			"GetAllUserBooksDomain - Database fetch failed",
			"error", err,
			"errorType", fmt.Errorf("error fetching books %w", err))
		return nil, fmt.Errorf("error fetching books: %w", err)
	}

	// Reverse normalize book data
	h.bookService.ReverseNormalizeBookData(&books)

	return books, nil
}


// Retrieve books by a specific author
func (h *BookHandlers) HandleGetBooksByAuthors(response http.ResponseWriter, request *http.Request) {
	requestID := uuid.New().String()
	h.logger.Info("=== HandleGetBooksByAuthors Trace Start ===", "requestID", requestID)
	defer h.logger.Info("=== HandleGetBooksByAuthors Trace End ===", "requestID", requestID)

	// Set Content Security Policy headers
	utils.SetCSPHeaders(response)
	h.logger.Info("CSP headers set", "requestID", requestID)

	// Grab token from cookie
	userID, err := jwt.ExtractUserIDFromJWT(request, config.AppConfig.JWTPublicKey)
	if err != nil {
			h.logger.Error("Error extracting user ID", "error", err, "requestID", requestID)
			http.Error(response, err.Error(), http.StatusUnauthorized)
			return
	}
	h.logger.Info("Valid user ID received from token", "userID", userID, "requestID", requestID)

	// Cache check
	cacheKey := fmt.Sprintf("%s%d", redis.PrefixBookAuthor, userID)
	h.logger.Info("Attempting cache retrieval", "requestID", requestID, "cacheKey", cacheKey)

	cachedData, err := h.redisClient.Get(request.Context(), cacheKey)
	if err != nil {
			h.logger.Info("Cache miss or error", "requestID", requestID, "error", err)
	} else {
			h.logger.Info("Cache hit", "requestID", requestID, "dataLength", len(cachedData))
	}

	if cachedData != "" {
			var booksByAuthors map[string][]repository.Book
			if err := json.Unmarshal([]byte(cachedData), &booksByAuthors); err == nil {
					h.logger.Info("Cache data unmarshaled successfully", "requestID", requestID, "authorCount", len(booksByAuthors))
					h.sendJSONResponse(response, JSONResponse{
							Data: map[string]interface{}{
									"booksByAuthors": booksByAuthors,
									"source":         "cache",
							},
					})
					return
			}
			h.logger.Error("Failed to unmarshal cached data", "requestID", requestID, "error", err, "cachedData", cachedData[:100]) // Log first 100 chars of cached data
	}

	// Database fetch
	h.logger.Info("Attempting to fetch books by authors from database",
	"userID", userID,
	"requestID", requestID)
	h.logger.Info("Fetching books from database", "requestID", requestID, "userID", userID)
	booksByAuthors, err := h.authorRepo.GetAllBooksByAuthors(userID)
	if err != nil {
			h.logger.Error("Database fetch failed",
					"requestID", requestID,
					"error", err,
					"errorType", fmt.Sprintf("%T", err))
			http.Error(response, "Error fetching books by authors", http.StatusInternalServerError)
			return
	}
	h.logger.Info("Database fetch successful", "requestID", requestID, "authorCount", len(booksByAuthors))

	// Cache update
	if booksByAuthorsJSON, err := json.Marshal(booksByAuthors); err == nil {
			cacheDuration := h.redisClient.GetConfig().CacheConfig.BooksByAuthor
			h.logger.Info("Attempting to cache results",
					"requestID", requestID,
					"dataSize", len(booksByAuthorsJSON),
					"cacheDuration", cacheDuration)

			if err := h.redisClient.Set(request.Context(), cacheKey, booksByAuthorsJSON, cacheDuration); err != nil {
					h.logger.Error("Cache update failed", "requestID", requestID, "error", err)
			} else {
					h.logger.Info("Cache update successful", "requestID", requestID)
			}
	}

	// Response preparation
	h.logger.Info("Preparing response", "requestID", requestID, "authorCount", len(booksByAuthors))
	h.sendJSONResponse(response, JSONResponse{
			Data: map[string]interface{}{
					"booksByAuthors": booksByAuthors,
					"source": "db",
			},
	})
	h.logger.Info("Response sent successfully", "requestID", requestID)
}

// Domain refactor - GetBooksByAuthors
func (h *BookHandlers) GetBooksByAuthorsDomain(ctx context.Context, userID int) (map[string]interface{}, error) {
	h.logger.Info("GetBooksByAuthorsDomain called",
			"userID", userID)

	// Database fetch
	booksByAuthors, err := h.authorRepo.GetAllBooksByAuthors(userID)
	if err != nil {
			h.logger.Error("Database fetch failed",
					"error", err,
					"errorType", fmt.Sprintf("%T", err))
			return nil, fmt.Errorf("error fetching books by authors: %w", err)
	}

	// Apply any necessary data transformations here
	// Note: Based on the original handler, no transformations were needed

	return booksByAuthors, nil
}



// Get Single Book by ID
func (h *BookHandlers) HandleGetBookByID(response http.ResponseWriter, request *http.Request) {
	// Set Content Security Policy headers
	utils.SetCSPHeaders(response)

	// Extract userID from JWT
	userID, err := jwt.ExtractUserIDFromJWT(request, config.AppConfig.JWTPublicKey)
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

		// Attempt to fetch books from Redis cache
		cacheKey := fmt.Sprintf("%s:%d:%d", redis.PrefixBookDetail, userID, bookID)
		if cachedData, err := h.redisClient.Get(request.Context(), cacheKey); err == nil {
			var book repository.Book
			if err := json.Unmarshal([]byte(cachedData), &book); err == nil {
				h.logger.Info("Crud.go - HandleGetBookByID - Cache hit: returning books from cache")

				// Apply title casing
				caser := cases.Title(language.Und)
				book.Title = caser.String(book.Title)

				h.sendJSONResponse(response, JSONResponse{
					Data: map[string]interface{}{
						"book":       book,
						"source":      "cache",
					},
				})
				return
			}
			h.logger.Error("Crud.go - HandleGetBookByID - Failed to unmarshal cached data", "error", err)
		}

	// Cache miss - fetch book by ID from db
	book, err := h.bookRepo.GetBookByID(bookID)
	if err != nil {
		h.logger.Error("Error fetching book", "error", err)
		http.Error(response, "Error fetching book", http.StatusInternalServerError)
		return
	}

	// Apply title casing
	caser := cases.Title(language.Und)
	book.Title = caser.String(book.Title)

	// Cache book data
	if bookJSON, err := json.Marshal(book); err == nil {
		cacheDuration := h.redisClient.GetConfig().CacheConfig.BookDetail
		if err := h.redisClient.Set(request.Context(), cacheKey, bookJSON, cacheDuration); err != nil {
			h.logger.Error("Crud.go - HandleGetBookByID - Failed to cache book data", "error", err, "bookID", bookID)
		}
	}

	// Send response
	h.sendJSONResponse(response, JSONResponse{
		Data: map[string]interface{}{
			"book":   book,
			"source": "db",
		},
	})
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

	// Invalidate L1 caches after inserting a book
	h.BookCache.InvalidateCaches(bookID, userID)

	// Prepare cache keys for Redis invalidation
	cacheKeys := []string{
		fmt.Sprintf("%s%d", redis.PrefixBook, userID),
		fmt.Sprintf("%s%d", redis.PrefixBookAuthor, userID),
		fmt.Sprintf("%s%d", redis.PrefixBookFormat, userID),
		fmt.Sprintf("%s%d", redis.PrefixBookGenre, userID),
		fmt.Sprintf("%s%d", redis.PrefixBookTag, userID),
		fmt.Sprintf("%s%d", redis.PrefixBookHomepage, userID),
	}

	// Attempt immediate cache invalidation
	ctx, cancel := context.WithTimeout(request.Context(), h.redisClient.GetConfig().TimeoutConfig.Write)
	defer cancel()

	if err := h.redisClient.Delete(ctx, cacheKeys...); err != nil {
		// If immediate invalidation fails, queue for async retry
		h.logger.Warn("Immediate cache invalidation failed, queueing for retry",
				"error", err,
				"userID", userID,
				"bookID", bookID,
		)

		if queueErr := h.cacheWorker.EnqueueInvalidationJob(request.Context(), workers.CacheInvalidationJob{
				Keys:      cacheKeys,
				UserID:    userID,
				BookID:    bookID,
				Timestamp: time.Now(),
		}); queueErr != nil {
				h.logger.Error("Failed to queue cache invalidation job",
						"error", queueErr,
						"userID", userID,
						"bookID", bookID,
				)
		}
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
	userID, bookID, err := h.ValidateBookOwnership(request)
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
	err = h.bookUpdater.UpdateBookEntry(request.Context(), book, userID)
	if err != nil {
			h.logger.Error("Error updating book", "error", err)
			http.Error(response, "Error updating book", http.StatusInternalServerError)
			return
	}

	// Invalidate L1 cache
	h.BookCache.InvalidateCaches(bookID, userID)

	// Invalidate L2 cache



	response.WriteHeader(http.StatusOK)
	json.NewEncoder(response).Encode(map[string]string{"message": "Book updated successfully"})
}


// Delete Book
func (h *BookHandlers) HandleDeleteBook(response http.ResponseWriter, request *http.Request) {
	// Validate book ownership
	userID, bookID, err := h.ValidateBookOwnership(request)
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

	// Invalidate caches after successful deletion
	h.BookCache.InvalidateCaches(bookID, userID)

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

	// Attempt to fetch from Redis cache
	cacheKey := fmt.Sprintf("%s%d", redis.PrefixBookFormat, userID)
	if cachedData, err := h.redisClient.Get(request.Context(), cacheKey); err == nil {
		var booksByFormat map[string][]repository.Book
		if err := json.Unmarshal([]byte(cachedData), &booksByFormat); err == nil {
			h.logger.Info("Crud.go - HandleGetBooksByFormat - Cache hit: returning books from cache")

			h.sendJSONResponse(response, JSONResponse{
				Data: map[string]interface{}{
					"booksByFormat": booksByFormat,
					"source":        "cache",
				},
			})
			return
		}
		h.logger.Error("Crud.go - HandleGetBooksByFormat - Failed to unmarshal cached data", "error", err)
	}

	// Cache miss - fetch books by format from db
	booksByFormat, err := h.formatRepo.GetAllBooksByFormat(userID)
	if err != nil {
		h.logger.Error("Error fetching books by format", "error", err)
		http.Error(response, "Error fetching books by format", http.StatusInternalServerError)
		return
	}

	// Apply reverse normalization to each book in each format
	for format, books := range booksByFormat {
		h.bookService.ReverseNormalizeBookData(&books)
		booksByFormat[format] = books
	}

	// Cache normalized data
	if booksByFormatJSON, err := json.Marshal(booksByFormat); err == nil {
		cacheDuration := h.redisClient.GetConfig().CacheConfig.BooksByFormat
		if err := h.redisClient.Set(request.Context(), cacheKey, booksByFormatJSON, cacheDuration); err != nil {
			h.logger.Error("Crud.go - HandleGetBooksByFormat - Failed to cache books by format", "error", err)
		}
	}

	// Send response
	h.sendJSONResponse(response, JSONResponse{
		Data: map[string]interface{}{
			"booksByFormat": booksByFormat,
			"source":        "db",
		},
	})
	//h.logger.Info("HandleGetBooksByFormat completed successfully")
}

// Domain refactor - GetBooksByFormatDomain
func (h *BookHandlers) GetBooksByFormatDomain(ctx context.Context, userID int) (map[string]interface{}, error) {
	// Fetch original data
	booksByFormat, err := h.formatRepo.GetAllBooksByFormat(userID)
	if err != nil {
		h.logger.Error("Database fetch failed",
			"error", err,
			"errorType", fmt.Sprintf("%T", err))
		return nil, fmt.Errorf("error fetching books by format: %w", err	)
	}

	// Transform data
	result := make(map[string]interface{})
	for format, books := range booksByFormat {
		h.bookService.ReverseNormalizeBookData(&books)
		booksByFormat[format] = books
	}

	return result, nil
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

	// Attempt to fetch from Redis cache
	cacheKey := fmt.Sprintf("%s%d", redis.PrefixBookGenre, userID)
	if cachedData, err := h.redisClient.Get(request.Context(), cacheKey); err == nil {
		var booksByGenres map[string][]repository.Book
		if err := json.Unmarshal([]byte(cachedData), &booksByGenres); err == nil {
			h.logger.Info("Crud.go - HandleGetBooksByGenres - Cache hit: returning books from cache")

			h.sendJSONResponse(response, JSONResponse{
				Data: map[string]interface{}{
					"booksByGenres": booksByGenres,
					"source":        "cache",
				},
			})
			return
		}
		h.logger.Error("Crud.go - HandleGetBooksByGenres - Failed to unmarshal cached data", "error", err)
	}

	// Cache miss - fetch books by genres from db
	booksByGenres, err := h.genreRepo.GetAllBooksByGenres(request.Context(), userID)
	if err != nil {
		h.logger.Error("Error fetching books by genres", "error", err)
		http.Error(response, "Error fetching books by genres", http.StatusInternalServerError)
		return
	}

	// Cache the data
	if booksByGenresJSON, err := json.Marshal(booksByGenres); err == nil {
		cacheDuration := h.redisClient.GetConfig().CacheConfig.BooksByGenre
		if err := h.redisClient.Set(request.Context(), cacheKey, booksByGenresJSON, cacheDuration); err != nil {
			h.logger.Error("Crud.go - HandleGetBooksByGenres - Failed to cache books by genres", "error", err)
		}
	}

	// Send response
	h.sendJSONResponse(response, JSONResponse{
		Data: map[string]interface{}{
			"booksByGenres": booksByGenres,
			"source":        "db",
		},
	})

	//h.logger.Info("HandleGetBooksByGenres completed successfully")
}

// Domain refactor - GetBooksByGenresDomain
func (h *BookHandlers) GetBooksByGenreDomain(ctx context.Context, userID int) (map[string]interface{}, error) {
	// Fetch original data
	booksByGenre, err := h.genreRepo.GetAllBooksByGenres(ctx, userID)
	if err != nil {
		h.logger.Error("Database fetch failed",
			"error", err,
			"errorType", fmt.Sprintf("%T", err))
		return nil, fmt.Errorf("error fetching books by genre: %w", err)
	}

	return booksByGenre, nil
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

	// Attempt to fetch from Redis cache
	cacheKey := fmt.Sprintf("%s%d", redis.PrefixBookTag, userID)
	if cachedData, err := h.redisClient.Get(request.Context(), cacheKey); err == nil {
			var booksByTags map[string][]repository.Book
			if err := json.Unmarshal([]byte(cachedData), &booksByTags); err == nil {
					h.logger.Info("Crud.go - HandleGetBooksByTags - Cache hit: returning books from cache")

					h.sendJSONResponse(response, JSONResponse{
							Data: map[string]interface{}{
									"booksByTags": booksByTags,
									"source":      "cache",
							},
					})
					return
			}
			h.logger.Error("Crud.go - HandleGetBooksByTags - Failed to unmarshal cached data", "error", err)
	}

	// Cache miss - fetch books by tags from db
	booksByTags, err := h.tagRepo.GetAllBooksByTags(request.Context(), userID)
	if err != nil {
		h.logger.Error("Error fetching books by tags", "error", err)
		http.Error(response, "Error fetching books by tags", http.StatusInternalServerError)
		return
	}

	// Cache the data
	if booksByTagsJSON, err := json.Marshal(booksByTags); err == nil {
		cacheDuration := h.redisClient.GetConfig().CacheConfig.BooksByTag
		if err := h.redisClient.Set(request.Context(), cacheKey, booksByTagsJSON, cacheDuration); err != nil {
				h.logger.Error("Crud.go - HandleGetBooksByTags - Failed to cache books by tags", "error", err)
		}
	}

	// Send response
	h.sendJSONResponse(response, JSONResponse{
		Data: map[string]interface{}{
			"booksByTags": booksByTags,
			"source":      "db",
		},
	})
}

// Domain refactor - GetBooksByTagsDomain
func (h *BookHandlers) GetBooksByTagsDomain(ctx context.Context, userID int) (map[string][]repository.Book, error) {
	booksByTags, err := h.tagRepo.GetAllBooksByTags(ctx, userID)
	if err != nil {
			h.logger.Error("Database fetch failed",
					"error", err,
					"errorType", fmt.Sprintf("%T", err))
			return nil, fmt.Errorf("error fetching books by tags: %w", err)
	}

	// Type assertion to ensure correct type
	result := make(map[string][]repository.Book)
	for tag, books := range booksByTags {
			if booksSlice, ok := books.([]repository.Book); ok {
					result[tag] = booksSlice
			} else {
					return nil, fmt.Errorf("invalid data structure for tag %s", tag)
			}
	}

	return result, nil
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

	// Attempt to fetch from Redis cache
	cacheKey := fmt.Sprintf("%s%d", redis.PrefixBookHomepage, userID)
	if cachedData, err := h.redisClient.Get(request.Context(), cacheKey); err == nil {
			var homepageData map[string]interface{}
			if err := json.Unmarshal([]byte(cachedData), &homepageData); err == nil {
					h.logger.Info("Crud.go - HandleGetHomepageData - Cache hit: returning homepage data from cache")

					h.sendJSONResponse(response, JSONResponse{
							Data: map[string]interface{}{
									"data": homepageData,
									"source": "cache",
							},
					})
					return
			}
			h.logger.Error("Crud.go - HandleGetHomepageData - Failed to unmarshal cached data", "error", err)
	}

	// Cache miss - fetch data using goroutines
	userTagsChan := make(chan map[string]interface{}, 1)
	userLangChan := make(chan map[string]interface{}, 1)
	userGenresChan := make(chan map[string]interface{}, 1)
	userAuthorsChan := make(chan map[string]interface{}, 1)
	errorChan := make(chan error, 4) // match number of error channels to goroutines

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
		langs, err := h.BookCache.GetBooksByLanguage(request.Context(), userID)
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

	// Cache the aggregated data
	if homepageJSON, err := json.Marshal(responseData); err == nil {
		cacheDuration := h.redisClient.GetConfig().CacheConfig.BookHomepage
		if err := h.redisClient.Set(request.Context(), cacheKey, homepageJSON, cacheDuration); err != nil {
				h.logger.Error("Crud.go - HandleGetHomepageData - Failed to cache homepage data", "error", err)
		}
	}

	// Send response
	h.sendJSONResponse(response, JSONResponse{
    Data: map[string]interface{}{
        "data": responseData,
        "source": "db",
    },
	})
}

// Helper fn to send consistent JSON responses
func (h *BookHandlers) sendJSONResponse(w http.ResponseWriter, response JSONResponse) {
	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Set status code
	if response.StatusCode == 0 {
			response.StatusCode = http.StatusOK
	}
	w.WriteHeader(response.StatusCode)

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response.Data); err != nil {
			h.logger.Error("Failed to encode JSON response",
					"error", err,
					"data", response.Data,
					"statusCode", response.StatusCode,
			)

			// Set error status code before writing response
			w.WriteHeader(http.StatusInternalServerError)

			errResponse := map[string]string{
					"error": "Internal server error",
			}

			if encErr := json.NewEncoder(w).Encode(errResponse); encErr != nil {
					h.logger.Error("Failed to encode error response",
							"originalError", err,
							"encodingError", encErr,
					)
					// At this point, we can only try to write a plain text response
					w.Write([]byte("Internal server error"))
			}
			return // Important: return after sending error response
	}
}