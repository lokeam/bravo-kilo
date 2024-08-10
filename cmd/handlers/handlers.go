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
	"net/url"
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
			"https://www.googleapis.com/auth/books",
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

	url := GoogleLoginConfig.AuthCodeURL(
		randomState,
		oauth2.AccessTypeOffline,
	)

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
		h.logger.Error("Error: URL State and Cookie state don't match")
		http.Error(response, "Error: States don't match", http.StatusBadRequest)
		return
	}

	code := request.URL.Query().Get("code")

	googleCfg := config.AppConfig.GoogleLoginConfig

	// Exchange the authorization code for an access token
	token, err := googleCfg.Exchange(context.Background(), code, oauth2.AccessTypeOffline)
	if err != nil {
		h.logger.Error("Error exchanging code for token", "error", err)
		http.Error(response, "Error exchanging code for token", http.StatusInternalServerError)
		return
	}

	// Retrieve user info using the access token
	oauthResponse, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		h.logger.Error("Error getting user info", "error", err)
		http.Error(response, "Error getting user info", http.StatusInternalServerError)
		return
	}
	defer oauthResponse.Body.Close()

	// Decode the user info response
	var userInfo struct {
		Id      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(oauthResponse.Body).Decode(&userInfo); err != nil {
		h.logger.Error("Error decoding user info response:", "error", err)
		http.Error(response, "Error decoding user info response", http.StatusInternalServerError)
		return
	}

	h.logger.Info("User info received!", "user", userInfo)

	firstName, lastName := utils.SplitFullName(userInfo.Name)

	// Check if the user already exists in the database
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
		// Save user info in the database
		user := data.User{
			Email:     userInfo.Email,
			FirstName: firstName,
			LastName:  lastName,
			Picture:   userInfo.Picture,
		}

		userId, err = h.models.User.Insert(user)
		if err != nil {
			h.logger.Error("Error adding user to db", "error", err)
			http.Error(response, "Error adding user to db", http.StatusInternalServerError)
			return
		}
	}

	// Store refresh token if available
	if token.RefreshToken != "" {
		tokenRecord := data.Token{
			UserID:       userId,
			RefreshToken: token.RefreshToken,
			TokenExpiry:  token.Expiry,
		}

  if err := h.models.Token.Insert(tokenRecord); err != nil {
    h.logger.Error("Error adding token to db", "error", err)
    http.Error(response, "Error adding token to db", http.StatusInternalServerError)
		return
	}
  } else {
	  h.logger.Warn("No refresh token received, re-authentication needed for offline access")
  }

	// Create a JWT for the session
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

	// Send the JWT as a cookie
	http.SetCookie(response, &http.Cookie{
		Name:     "token",
		Value:    jwtString,
		Expires:  expirationTime,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	// Redirect to the frontend dashboard with userID as a query parameter
	dashboardURL := fmt.Sprintf("http://localhost:5173/library?userID=%d", userId)
	http.Redirect(response, request, dashboardURL, http.StatusSeeOther)

	h.logger.Info("JWT successfully sent to FE with status code: ", "info", http.StatusSeeOther)
}


// Retrieve Token
func (h *Handlers) getUserAccessToken(request *http.Request) (*oauth2.Token, error) {
	// Get userID from JWT
	cookie, err := request.Cookie("token")
	if err != nil {
		h.logger.Error("Error: No token cookie", "error", err)
		return nil, fmt.Errorf("no token cookie")
	}

	tokenStr := cookie.Value
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		h.logger.Error("Error: Invalid token", "error", err)
		return nil, fmt.Errorf("invalid token")
	}

	// Get the refresh token for the user from the database
	refreshToken, err := h.models.Token.GetRefreshTokenByUserID(claims.UserID)
	if err != nil {
		h.logger.Error("Error retrieving refresh token from DB", "error", err)
		return nil, fmt.Errorf("could not retrieve refresh token from DB")
	}
	if refreshToken == "" {
		h.logger.Error("No refresh token found for user", "userID", claims.UserID)
		return nil, fmt.Errorf("no refresh token found for user")
	}

	// Use the refresh token to obtain a new access token
	tokenSource := config.AppConfig.GoogleLoginConfig.TokenSource(context.Background(), &oauth2.Token{
		RefreshToken: refreshToken,
	})

	// Get a new access token
	newToken, err := tokenSource.Token()
	if err != nil {
		h.logger.Error("Error refreshing access token", "error", err)
		return nil, fmt.Errorf("could not refresh access token")
	}

	return newToken, nil
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
			http.Error(response, "Query parameter required in request", http.StatusBadRequest)
			return
	}

	// Debug - Set headers to prevent caching
	response.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	response.Header().Set("Pragma", "no-cache")
	response.Header().Set("Expires", "0")

	// Get user's access token
	accessToken, err := h.getUserAccessToken(request)
	if err != nil {
			h.logger.Error("Error retrieving user access token", "error", err)
			http.Error(response, "Error retrieving access token", http.StatusUnauthorized)
			return
	}

	// Use the access token to call the Google Books API
	client := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(accessToken))
	googleBooksURL := fmt.Sprintf("https://www.googleapis.com/books/v1/volumes?q=%s", url.QueryEscape(query))

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
	formattedBooks := h.FormatGoogleBooksResponse(response, booksData)
	// h.logger.Info("---------------")
	// h.logger.Info("Showing formattedBooks, pre-check:", "formattedBooks", formattedBooks)

	// Get user ID from JWT
	cookie, err := request.Cookie("token")
	if err != nil {
			h.logger.Error("No token cookie", "error", err)
			http.Error(response, "No token cookie", http.StatusUnauthorized)
			return
	}

	tokenStr := cookie.Value
	claims := &Claims{}
	jwtToken, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
	})
	if err != nil || !jwtToken.Valid {
			h.logger.Error("Invalid token", "error", err)
			http.Error(response, "Invalid token", http.StatusUnauthorized)
			return
	}

	userID := claims.UserID

	// Create hash sets for user's existing library data
	isbn10Set, err := h.models.Book.GetAllBooksISBN10(userID)
	if err != nil {
			h.logger.Error("Error retrieving user's ISBN10", "error", err)
			http.Error(response, "Error retrieving user's ISBN10", http.StatusInternalServerError)
			return
	}

	isbn13Set, err := h.models.Book.GetAllBooksISBN13(userID)
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

	bookMap, err := h.models.Book.GetAllBooksPublishDate(userID)
	if err != nil {
		h.logger.Error("Error retrieving user's book publish dates", "error", err)
		http.Error(response, "Error retrieving user's book publish dates", http.StatusInternalServerError)
		return
	}

	// Check bookMap
	h.logger.Info("======================================")
	fmt.Println("Checking bookmap: ", bookMap)

	// Check each book against the user's library
	for i := range formattedBooks {
		formattedBook := &formattedBooks[i]
		isInLibrary := (isbn10Set.Has(formattedBook.ISBN10) || isbn13Set.Has(formattedBook.ISBN13)) &&
			bookMap[formattedBook.Title] == formattedBook.PublishDate
			fmt.Println("checking formattedBook.publishDate", formattedBook.PublishDate)

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


// Format Google Books Response
func (h *Handlers) FormatGoogleBooksResponse(response http.ResponseWriter, booksData interface{}) []data.Book {
	var gBooksResponse []data.Book

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
		formattedBook := data.Book{
			Title:       utils.GetStringValOrDefault(volumeInfo, "title", ""),
			Subtitle:    utils.GetStringValOrDefault(volumeInfo, "subtitle", ""),
			Description: utils.GetStringValOrDefault(volumeInfo, "description", ""),
			Language:    utils.GetStringValOrDefault(volumeInfo, "language", ""),
			PageCount:   utils.GetIntValOrDefault(volumeInfo, "pageCount", 0),
			PublishDate: utils.GetStringValOrDefault(volumeInfo, "publishedDate", ""),
		}

		// Handle image links, ensuring it's always an array
		formattedBook.ImageLinks = []string{}
		if imageLinks, ok := volumeInfo["imageLinks"].(map[string]interface{}); ok {
			formattedBook.ImageLinks = append(formattedBook.ImageLinks,
				utils.GetStringValOrDefault(imageLinks, "thumbnail", ""),
				utils.GetStringValOrDefault(imageLinks, "smallThumbnail", ""),
			)
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

		gBooksResponse = append(gBooksResponse, formattedBook)
	}

	h.logger.Info("Formatted books", "books", gBooksResponse)
	return gBooksResponse
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
		"isInLibrary": true,
	}

	response.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(response).Encode(dbResponse); err != nil {
		http.Error(response, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// Retrieve books by a specific author
func (h *Handlers) GetBooksByAuthors(response http.ResponseWriter, request *http.Request) {
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

// Sorting - Get Books by Format
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

	// h.logger.Info("Books fetched successfully", "booksByFormat", booksByFormat)

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(booksByFormat)
}

// Sorting - Get Books by Genre
func (h *Handlers) GetBooksByGenres(response http.ResponseWriter, request *http.Request) {
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

	// Get books by genres
	booksByGenres, err := h.models.Book.GetAllBooksByGenres(userID)
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
