package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"github.com/lib/pq"
)

func (r *BookRepositoryImpl) GetAllBooksByAuthors(userID int) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// Single query to get all book details by authors
	query := `
	SELECT r.id, r.title, r.subtitle, r.description, r.language, r.page_count, r.publish_date,
	       r.image_link, r.notes, r.created_at, r.last_updated, r.isbn_10, r.isbn_13,
	       a.name AS author_name,
	       json_agg(DISTINCT g.name) AS genres,
	       json_agg(DISTINCT f.format_type) AS formats,
	       r.tags
	FROM books r
	INNER JOIN book_authors ba ON r.id = ba.book_id
	INNER JOIN authors a ON ba.author_id = a.id
	INNER JOIN user_books ub ON r.id = ub.book_id
	LEFT JOIN book_genres bg ON r.id = bg.book_id
	LEFT JOIN genres g ON bg.genre_id = g.id
	LEFT JOIN book_formats bf ON r.id = bf.book_id
	LEFT JOIN formats f ON bf.format_id = f.id
	WHERE ub.user_id = $1
	GROUP BY r.id, a.name`

	rows, err := r.DB.QueryContext(ctx, query, userID)
	if err != nil {
		r.Logger.Error("Error retrieving books by authors", "error", err)
		return nil, err
	}
	defer rows.Close()

	// Store authors and books
	authors := []string{}
	booksByAuthor := map[string][]Book{}

	for rows.Next() {
		var book Book
		var authorName string
		var genresJSON, formatsJSON, tagsJSON []byte

		// Scan the result
		if err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.Subtitle,
			&book.Description,
			&book.Language,
			&book.PageCount,
			&book.PublishDate,
			&book.ImageLink,
			&book.Notes,
			&book.CreatedAt,
			&book.LastUpdated,
			&book.ISBN10,
			&book.ISBN13,
			&authorName,
			&genresJSON,
			&formatsJSON,
			&tagsJSON,
		); err != nil {
			r.Logger.Error("Error scanning book by author", "error", err)
			return nil, err
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(genresJSON, &book.Genres); err != nil {
			r.Logger.Error("Error unmarshalling genres JSON", "error", err)
			return nil, err
		}
		if err := json.Unmarshal(formatsJSON, &book.Formats); err != nil {
			r.Logger.Error("Error unmarshalling formats JSON", "error", err)
			return nil, err
		}
		if err := json.Unmarshal(tagsJSON, &book.Tags); err != nil {
			r.Logger.Error("Error unmarshalling tags JSON", "error", err)
			return nil, err
		}

		// Mark book as in Library
		book.IsInLibrary = true

		// Populate the Authors field directly
		book.Authors = []string{authorName}

		// Add author to the list if not already present
		if _, found := booksByAuthor[authorName]; !found {
			authors = append(authors, authorName)
		}

		// Add book to the author's list
		booksByAuthor[authorName] = append(booksByAuthor[authorName], book)
	}

	// Check for row errors
	if err = rows.Err(); err != nil {
		r.Logger.Error("Error with rows", "error", err)
		return nil, err
	}

	// Sort authors by last name
	sort.Slice(authors, func(i, j int) bool {
		return getLastName(authors[i]) < getLastName(authors[j])
	})

	// Create the result map with index keys for authors
	result := map[string]interface{}{
		"allAuthors": authors,
	}

	for i, author := range authors {
		key := strconv.Itoa(i)
		result[key] = booksByAuthor[author]
	}

	return result, nil
}

func (r *BookRepositoryImpl) GetAuthorsForBook(bookID int) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
	SELECT array_to_json(array_agg(a.name))
	FROM authors a
	JOIN book_authors ba ON a.id = ba.author_id
	WHERE ba.book_id = $1`

	var authorsJSON []byte
	err := b.DB.QueryRowContext(ctx, query, bookID).Scan(&authorsJSON)
	if err != nil {
		b.Logger.Error("Error fetching authors for book", "error", err)
		return nil, err
	}

	// Unmarshal the JSON array of authors
	var authors []string
	if err := json.Unmarshal(authorsJSON, &authors); err != nil {
		b.Logger.Error("Error unmarshalling authors JSON", "error", err)
		return nil, err
	}

	return authors, nil
}

func (r *BookRepositoryImpl) GetAuthorsForBooks(bookIDs []int) (map[int][]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var rows *sql.Rows
	var err error

	// Use prepared statement if available
	if r.getAuthorsForBooksStmt != nil {
		rows, err = r.getAuthorsForBooksStmt.QueryContext(ctx, pq.Array(bookIDs))
		if err != nil {
			r.Logger.Error("Error fetching authors for books using prepared statement", "error", err)
			return nil, err
		}
	} else {
		// Fallback if the prepared statement is unavailable
		query := `
		SELECT ba.book_id, a.name
		FROM authors a
		JOIN book_authors ba ON a.id = ba.author_id
		WHERE ba.book_id = ANY($1)`

		rows, err = r.DB.QueryContext(ctx, query, pq.Array(bookIDs))
		if err != nil {
			r.Logger.Error("Error fetching authors for books using fallback query", "error", err)
			return nil, err
		}
	}
	defer rows.Close()

	// Map to store authors for each bookID
	authorsByBook := make(map[int][]string)
	for rows.Next() {
		var bookID int
		var authorName string
		if err := rows.Scan(&bookID, &authorName); err != nil {
			r.Logger.Error("Error scanning author name", "error", err)
			return nil, err
		}
		authorsByBook[bookID] = append(authorsByBook[bookID], authorName)
	}

	return authorsByBook, nil
}

func (r *BookRepositoryImpl) GetBooksByAuthor(authorName string) ([]Book, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// Fetch all books by the author
	query := `
	SELECT r.id, r.title, r.subtitle, r.description, r.language, r.page_count, r.publish_date,
	       r.image_link, r.notes, r.created_at, r.last_updated, r.isbn_10, r.isbn_13
	FROM books b
	INNER JOIN book_authors ba ON b.id = ba.book_id
	INNER JOIN authors a ON ba.author_id = a.id
	WHERE a.name = $1`

	rows, err := r.DB.QueryContext(ctx, query, authorName)
	if err != nil {
		r.Logger.Error("Error retrieving books by author", "error", err)
		return nil, err
	}
	defer rows.Close()

	var books []Book
	bookIDMap := make(map[int]*Book)
	bookIDs := []int{}

	for rows.Next() {
		var book Book

		if err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.Subtitle,
			&book.Description,
			&book.Language,
			&book.PageCount,
			&book.PublishDate,
			&book.ImageLink,
			&book.Notes,
			&book.CreatedAt,
			&book.LastUpdated,
			&book.ISBN10,
			&book.ISBN13,
		); err != nil {
			r.Logger.Error("Error scanning book", "error", err)
			return nil, err
		}

		book.IsInLibrary = true
		books = append(books, book)
		bookIDMap[book.ID] = &book
		bookIDs = append(bookIDs, book.ID)
	}

	// Fetch all authors for the books in a single query
	authorQuery := `
	SELECT ba.book_id, a.name
	FROM book_authors ba
	INNER JOIN authors a ON ba.author_id = a.id
	WHERE ba.book_id = ANY($1)`

	authorRows, err := r.DB.QueryContext(ctx, authorQuery, pq.Array(bookIDs))
	if err != nil {
		r.Logger.Error("Error retrieving authors for books", "error", err)
		return nil, err
	}
	defer authorRows.Close()

	for authorRows.Next() {
		var bookID int
		var authorName string
		if err := authorRows.Scan(&bookID, &authorName); err != nil {
			r.Logger.Error("Error scanning author", "error", err)
			return nil, err
		}
		bookIDMap[bookID].Authors = append(bookIDMap[bookID].Authors, authorName)
	}

	// Fetch all genres for the books in a single query
	genreQuery := `
	SELECT bg.book_id, g.name
	FROM book_genres bg
	INNER JOIN genres g ON bg.genre_id = g.id
	WHERE bg.book_id = ANY($1)`

	genreRows, err := r.DB.QueryContext(ctx, genreQuery, pq.Array(bookIDs))
	if err != nil {
		r.Logger.Error("Error retrieving genres for books", "error", err)
		return nil, err
	}
	defer genreRows.Close()

	for genreRows.Next() {
		var bookID int
		var genreName string
		if err := genreRows.Scan(&bookID, &genreName); err != nil {
			r.Logger.Error("Error scanning genre", "error", err)
			return nil, err
		}
		bookIDMap[bookID].Genres = append(bookIDMap[bookID].Genres, genreName)
	}

	if err = rows.Err(); err != nil {
		r.Logger.Error("GetBooksByAuthor - Error with rows", "error", err)
		return nil, err
	}

	return books, nil
}

// Helper function to get the last name from a full name
func getLastName(fullName string) string {
	parts := strings.Fields(fullName)
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}
