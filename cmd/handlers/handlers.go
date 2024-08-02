package handlers

import (
	"bravo-kilo/config"
	"bravo-kilo/internal/data"
	"bravo-kilo/internal/utils"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var jwtKey = []byte("extra-super-secret-256-bit-key")

type Claims struct {
	UserID int `json:"userId"`
	jwt.RegisteredClaims
}

// Handlers struct to hold the logger
type Handlers struct {
	logger *slog.Logger
	models data.Models
}

type jsonResponse struct {
	Error    bool        `json:"error"`
	Message  string      `json:"message"`
	Data     interface{} `json:"data,omitempty"`
}


// NewHandlers creates a new Handlers instance
func NewHandlers(logger *slog.Logger, models data.Models) *Handlers {
	return &Handlers{
		logger: logger,
		models: models,
	}
}

// Generate random state for CSRF protection
func generateState() string {
	byteSlice := make([]byte, 16)
	_, err := rand.Read(byteSlice)
	if err != nil {
		log.Fatalf("Error generating random state: %s", err)
	}

	return base64.URLEncoding.EncodeToString(byteSlice)
}

// Init OAuth with Google
func (h *Handlers) GoogleSignIn(response http.ResponseWriter, request *http.Request) {
	h.logger.Info("Handling Google OAuth callback")
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	var GoogleLoginConfig = oauth2.Config{
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URI"),
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	randomState := generateState()

	// Set the state as a cookie
	http.SetCookie(response, &http.Cookie{
		Name:     "oauthstate",
		Value:    randomState,
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
		Secure:   true,
	})

	url := GoogleLoginConfig.AuthCodeURL(randomState)

	http.Redirect(response, request, url, http.StatusSeeOther)
}

// Process Google OAuth callback
func (h *Handlers) GoogleCallback(response http.ResponseWriter, request *http.Request) {
	h.logger.Info("Handling Google OAuth callback")

	state := request.URL.Query().Get("state")

	// Retrieve the state from the cookie
	cookie, err := request.Cookie("oauthstate")
	if err != nil {
		h.logger.Error("Error: State cookie not found", "error", err)
		http.Error(response, "Error: State cookie not found", http.StatusBadRequest)
		return
	}

	if state != cookie.Value {
		h.logger.Error("Error: URL State and Cookie state don't match", "error", err)
		http.Error(response, "Error: States don't match", http.StatusBadRequest)
		return
	}

	code := request.URL.Query().Get("code")

	googleCfg := config.AppConfig.GoogleLoginConfig

	token, err := googleCfg.Exchange(context.Background(), code)
	if err != nil {
		h.logger.Error("Error exchanging code for token", "error", err)
		http.Error(response, "Error exchanging code for token", http.StatusInternalServerError)
		return
	}

	oauthResponse, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		h.logger.Error("Error getting user info", "error", err)
		http.Error(response, "Error getting user info", http.StatusInternalServerError)
		return
	}

	defer oauthResponse.Body.Close()

	var userInfo struct {
		Id       string `json:"id"`
		Email    string `json:"email"`
		Name     string `json:"name"`
		Picture  string `json:"picture"`
	}

	// Decode response
	if err := json.NewDecoder(oauthResponse.Body).Decode(&userInfo); err != nil {
		h.logger.Error("Error decoding user info response:", "error", err)
		http.Error(response, "Error decoding user info response", http.StatusInternalServerError)
		return
	}

	h.logger.Info("User info received!", "user", userInfo)

	firstName, lastName := utils.SplitFullName(userInfo.Name)

	// Check if user already exists in db
	existingUser, err := h.models.User.GetByEmail(userInfo.Email)
	if err != nil && err != sql.ErrNoRows {
		h.logger.Error("Error checking for existing user", "error", err)
		http.Error(response, "Error checking for existing user", http.StatusInternalServerError)
		return
	}

	var userId int
	if existingUser != nil {
		userId = existingUser.ID
	} else {
		// Save userInfo + access tokens
		user := data.User{
			Email:      userInfo.Email,
			FirstName:  firstName,
			LastName:   lastName,
			Picture:    userInfo.Picture,
		}

		userId, err = h.models.User.Insert(user)
		if err != nil {
			h.logger.Error("Error adding user to db", "error", err)
			http.Error(response, "Error adding user to db", http.StatusInternalServerError)
			return
		}
	}


	tokenRecord := data.Token{
		UserID:        userId,
		RefreshToken:  token.RefreshToken,
		TokenExpiry:   token.Expiry,
	}

	if err := h.models.Token.Insert(tokenRecord); err != nil {
		h.logger.Error("Error adding token to db", "error", err)
		http.Error(response, "Error adding token to db", http.StatusInternalServerError)
		return
	}

	h.logger.Info("User and tokens stored successfully")

	// Create JWT
	expirationTime := time.Now().Add(60 * time.Minute)
	claims := &Claims{
		UserID: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtString, err := jwtToken.SignedString(jwtKey)
	if err != nil {
		h.logger.Error("Error generating JWT", "error", err)
		http.Error(response, "Error generating JWT", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Generated JWT")

	// JWT as a cookie
	http.SetCookie(response, &http.Cookie{
		Name:     "token",
		Value:    jwtString,
		Expires:  expirationTime,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	// Redirect to FE dashboard with userID as query parameter
	dashboardURL := fmt.Sprintf("http://localhost:5173/library?userID=%d", userId)
	http.Redirect(response, request, dashboardURL, http.StatusSeeOther)

	h.logger.Info("JWT successfully sent to FE with status code: ", "info", http.StatusSeeOther)
}

// Verify JWT Token
func (h *Handlers) VerifyToken(response http.ResponseWriter, request *http.Request) {
	cookie, err := request.Cookie("token")
	if err != nil {
			h.logger.Error("Error: No token cookie", "error", err)
			http.Error(response, "Error: No token cookie", http.StatusUnauthorized)
			return
	}

	tokenStr := cookie.Value
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
	})
	if err != nil || !token.Valid {
			h.logger.Error("Error: Invalid token", "error", err)
			http.Error(response, "Invalid token", http.StatusUnauthorized)
			return
	}

	user, err := h.models.User.GetByID(claims.UserID)
	if err != nil {
			h.logger.Error("Error: User not found", "error", err)
			http.Error(response, "User not found", http.StatusInternalServerError)
			return
	}

	h.logger.Info("User info retrieved!", "user", user)

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(map[string]interface{}{
			"user": user,
	})
}

// Refresh Token
func (h *Handlers) RefreshToken(response http.ResponseWriter, request *http.Request) {
	// Grab refresh token from request
	cookie, err := request.Cookie("token")
	if err != nil {
		h.logger.Error("Error: No refresh token cookie", "error", err)
		http.Error(response, "Error: No refresh token cookie", http.StatusUnauthorized)
		return
	}

	oldTokenStr := cookie.Value

	// Parse old token
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(oldTokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		h.logger.Error("Error: Invalid refresh token", "error", err)
		http.Error(response, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	// Generate new token (1 week)
	expirationTime := time.Now().Add(7 * 24 * time.Hour)
	newClaims := &Claims{
		UserID: claims.UserID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	newTokenStr, err := newToken.SignedString(jwtKey)
	if err != nil {
		h.logger.Error("Error generating new JWT", "error", err)
		http.Error(response, "Error generating new JWT", http.StatusInternalServerError)
		return
	}

	// Rotate the refresh one
	err = h.models.Token.Rotate(claims.UserID, newTokenStr, oldTokenStr, expirationTime)
	if err != nil {
		h.logger.Error("Error rotating refresh token", "error", err)
		http.Error(response, "Error rotating refresh token", http.StatusInternalServerError)
		return
	}

	// Set the new refresh as a cookie
	http.SetCookie(response, &http.Cookie{
		Name:     "token",
		Value:    newTokenStr,
		Expires:  expirationTime,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	response.WriteHeader(http.StatusOK)
	response.Write([]byte("Token refreshed successfully"))
}

// Process user sign out
func (h *Handlers) SignOut(response http.ResponseWriter, request *http.Request) {
	// Clear token cookie
	http.SetCookie(response, &http.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	// Get the user ID from JWT token
	cookie, err := request.Cookie("token")
	if err != nil {
		h.logger.Error("Error: No token cookie", "error", err)
		http.Error(response, "Error: No token cookie", http.StatusUnauthorized)
		return
	}

	tokenStr := cookie.Value
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		h.logger.Error("Error: Invalid token", "error", err)
		http.Error(response, "Error: Invalid token", http.StatusUnauthorized)
		return
	}

	// Delete the refresh token from db
	if err := h.models.Token.DeleteByUserID(claims.UserID); err != nil {
		h.logger.Error("Error deleting refresh token by user ID", "error", err)
		http.Error(response, "Error deleting refresh token", http.StatusInternalServerError)
		return
	}

	h.logger.Info("User logged out, token cookie cleared, token deleted")
	response.WriteHeader(http.StatusOK)
	response.Write([]byte("Logged out successfully"))
}

// Process Google Books API Search
func (h *Handlers) SearchBooks(response http.ResponseWriter, request *http.Request) {
	query := request.URL.Query().Get("query")
	if query == "" {
		h.logger.Error("Search query is missing", "error", "missing query")
		http.Error(response, "Query parameter is required", http.StatusBadRequest)
		return
	}

	googleBooksAPI := "https://www.googleapis.com/books/v1/volumes"
	req, err := http.NewRequest("GET", googleBooksAPI, nil)
	if err != nil {
		h.logger.Error("Error creating request", "error", err)
		http.Error(response, "Error creating request", http.StatusInternalServerError)
		return
	}

	q := req.URL.Query()
	q.Add("q", query)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error("Error making request to Google Books API", "error", err)
		http.Error(response, "Error making request to Google Books API", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.logger.Error("Error response from Google Books API", "status", resp.Status)
		http.Error(response, "Error response from Google Books API", resp.StatusCode)
		return
	}

	var searchResult map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		h.logger.Error("Error decoding Google Books API response", "error", err)
		http.Error(response, "Error decoding Google Books API response", http.StatusInternalServerError)
		return
	}

	// Transform response for client
	books, err := utils.TransformGoogleBooksResponse(searchResult)
	if err != nil {
		h.logger.Error("Error transforming Google Books API response", "error", err)
		http.Error(response, "Error transforming Google Books API response", http.StatusInternalServerError)
		return
	}

	response.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(response).Encode(map[string]interface{}{"items": books}); err != nil {
		h.logger.Error("Error encoding response", "error", err)
		http.Error(response, "Error encoding response", http.StatusInternalServerError)
	}
}

// Get all User Books
func (h *Handlers) GetAllUserBooks(response http.ResponseWriter, request *http.Request) {
	// Grab token from cookie
	cookie, err := request.Cookie("token")
	if err != nil {
		h.logger.Error("No token cookie", "error", err)
		http.Error(response, "No token cookie", http.StatusUnauthorized)
		return
	}

	tokenStr := cookie.Value
	claims := &Claims{}

	// Parse JWT token to get userID
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		h.logger.Error("Invalid token", "error", err)
		http.Error(response, "Invalid token", http.StatusUnauthorized)
		return
	}

	userID := claims.UserID

	books, err := h.models.Book.GetAllBooksByUserID(userID)
	if err != nil {
		h.logger.Error("Error fetching books", "error", err)
		http.Error(response, "Error fetching books", http.StatusInternalServerError)
		return
	}

	dbResponse := map[string]interface{}{
		"books": books,
	}

	response.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(response).Encode(dbResponse); err != nil {
		http.Error(response, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// Get Single Book by ID
func (h *Handlers) GetBookByID(response http.ResponseWriter, request *http.Request) {
	// Grab token from cookie
	cookie, err := request.Cookie("token")
	if err != nil {
		h.logger.Error("No token cookie", "error", err)
		http.Error(response, "No token cookie", http.StatusUnauthorized)
		return
	}

	tokenStr := cookie.Value
	claims := &Claims{}

	// Parse JWT to get userID
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		h.logger.Error("Invalid token", "error", err)
		http.Error(response, "Invalid token", http.StatusUnauthorized)
		return
	}

	bookIDStr := chi.URLParam(request, "bookID")
	bookID, err := strconv.Atoi(bookIDStr)
	if err != nil {
		http.Error(response, "Invalid book ID", http.StatusBadRequest)
		return
	}

	book, err := h.models.Book.GetByID(bookID)
	if err != nil {
		h.logger.Error("Error fetching book", "error", err)
		http.Error(response, "Error fetching book", http.StatusInternalServerError)
		return
	}

	// Get formats
	formats, err := h.models.Book.GetFormats(bookID)
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

// Add Book
func (h *Handlers) InsertBook(response http.ResponseWriter, request *http.Request) {
	var book data.Book
	err := json.NewDecoder(request.Body).Decode(&book)
	if err != nil {
		h.logger.Error("Error decoding book data", "error", err)
		http.Error(response, "Error decoding book data - invalid input", http.StatusBadRequest)
		return
	}

	// Insert the book and get the ID
	bookID, err := h.models.Book.Insert(book)
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

		if err := h.models.Book.AddFormat(bookID, formatID); err != nil {
			h.logger.Error("Error adding format association", "error", err)
			http.Error(response, "Error adding format association", http.StatusInternalServerError)
			return
		}
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(map[string]int{"book_id": bookID})
}


// Update Book
func (h *Handlers) UpdateBook(response http.ResponseWriter, request *http.Request) {
	var book data.Book
	err := json.NewDecoder(request.Body).Decode(&book)
	if err != nil {
		h.logger.Error("Error decoding book data", "error", err)
		http.Error(response, "Error decoding book data - invalid input", http.StatusBadRequest)
		return
	}

	// Ensure book ID is provided in the URL and parse it
	bookIDStr := chi.URLParam(request, "bookID")
	bookID, err := strconv.Atoi(bookIDStr)
	if err != nil {
		h.logger.Error("Invalid book ID", "error", err)
		http.Error(response, "Invalid book ID", http.StatusBadRequest)
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

	// Remove old format associations
	err = h.models.Book.RemoveFormats(book.ID)
	if err != nil {
		h.logger.Error("Error removing old formats", "error", err)
		http.Error(response, "Error removing old formats", http.StatusInternalServerError)
		return
	}

	// Insert new formats and their associations
	for _, formatType := range book.Formats {
		formatID, err := h.models.Format.Insert(bookID, formatType)
		if err != nil {
			h.logger.Error("Error inserting format", "error", err)
			http.Error(response, "Error inserting format", http.StatusInternalServerError)
			return
		}

		if err := h.models.Book.AddFormat(book.ID, formatID); err != nil {
			h.logger.Error("Error adding format association", "error", err)
			http.Error(response, "Error adding format association", http.StatusInternalServerError)
			return
		}
	}

	response.WriteHeader(http.StatusOK)
	json.NewEncoder(response).Encode(map[string]string{"message": "Book updated successfully"})
}

// Delete Book
func (h *Handlers) DeleteBook(response http.ResponseWriter, request *http.Request) {
	bookIDStr := chi.URLParam(request, "bookID")
	bookID, err := strconv.Atoi(bookIDStr)
	if err != nil {
		h.logger.Error("Invalid book ID", "error", err)
		http.Error(response, "Invalid book ID", http.StatusBadRequest)
		return
	}

	err = h.models.Book.Delete(bookID)
	if err != nil {
		h.logger.Error("Error deleting book", "error", err)
		http.Error(response, "Error deleting book", http.StatusInternalServerError)
		return
	}

	response.WriteHeader(http.StatusOK)
	json.NewEncoder(response).Encode(map[string]string{"message": "Book deleted successfully"})
}

func (h *Handlers) GetBooksByFormat(response http.ResponseWriter, request *http.Request) {
	// Grab token from cookie
	cookie, err := request.Cookie("token")
	if err != nil {
		h.logger.Error("No token cookie", "error", err)
		http.Error(response, "No token cookie", http.StatusUnauthorized)
		return
	}

	tokenStr := cookie.Value
	claims := &Claims{}

	// Parse JWT token to get userID
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		h.logger.Error("Invalid token", "error", err)
		http.Error(response, "Invalid token", http.StatusUnauthorized)
		return
	}

	userID := claims.UserID
	h.logger.Info("Valid user ID received from token", "userID", userID)

	// Get books by format
	booksByFormat, err := h.models.Book.GetAllBooksByFormat(userID)
	if err != nil {
		h.logger.Error("Error fetching books by format", "error", err)
		http.Error(response, "Error fetching books by format", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Books fetched successfully", "booksByFormat", booksByFormat)

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(booksByFormat)
}
