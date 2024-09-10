package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"sort"
	"strconv"
)


func (r *BookRepositoryImpl) GetAllBooksByGenres(ctx context.Context, userID int) (map[string]interface{}, error) {
	var rows *sql.Rows
	var err error

	// Use the prepared statement if available, else fall back to a raw query
	if r.getAllBooksByGenresStmt != nil {
			r.Logger.Info("Using prepared statement for retrieving books by genres")
			rows, err = r.getAllBooksByGenresStmt.QueryContext(ctx, userID)
	} else {
			r.Logger.Info("Prepared statement unavailable, using fallback query for retrieving books by genres")
			query := `
					SELECT r.id, r.title, r.subtitle, r.description, r.language, r.page_count, r.publish_date,
								 r.image_link, r.notes, r.created_at, r.last_updated, r.isbn_10, r.isbn_13,
								 json_agg(DISTINCT g.name) AS genres,
								 json_agg(DISTINCT a.name) AS authors,
								 json_agg(DISTINCT f.format_type) AS formats,
								 r.tags
					FROM books b
					INNER JOIN user_books ub ON b.id = ub.book_id
					LEFT JOIN book_genres bg ON b.id = bg.book_id
					LEFT JOIN genres g ON bg.genre_id = g.id
					LEFT JOIN book_authors ba ON b.id = ba.book_id
					LEFT JOIN authors a ON ba.author_id = a.id
					LEFT JOIN book_formats bf ON b.id = bf.book_id
					LEFT JOIN formats f ON bf.format_id = f.id
					WHERE ub.user_id = $1
					GROUP BY b.id`
			rows, err = r.DB.QueryContext(ctx, query, userID)
	}

	if err != nil {
			r.Logger.Error("Error retrieving books by genres", "error", err)
			return nil, err
	}
	defer rows.Close()

	genresSet := make(map[string]struct{}) // Track unique genres
	booksByGenre := map[string][]Book{}    // Store books by genre

	// Iterate through the rows and process the results
	for rows.Next() {
			var book Book
			var genresJSON, authorsJSON, formatsJSON, tagsJSON []byte

			// Ensure the scan order matches the SQL query's column order
			if err := rows.Scan(
					&book.ID, &book.Title, &book.Subtitle, &book.Description, &book.Language, &book.PageCount,
					&book.PublishDate, &book.ImageLink, &book.Notes, &book.CreatedAt, &book.LastUpdated,
					&book.ISBN10, &book.ISBN13, &genresJSON, &authorsJSON, &formatsJSON, &tagsJSON,
			); err != nil {
					r.Logger.Error("Error scanning book by genre", "error", err)
					return nil, err
			}

			// Unmarshal JSON fields into the respective slices
			if err := json.Unmarshal(genresJSON, &book.Genres); err != nil {
					r.Logger.Error("Error unmarshalling genres JSON", "error", err)
					return nil, err
			}
			if err := json.Unmarshal(authorsJSON, &book.Authors); err != nil {
					r.Logger.Error("Error unmarshalling authors JSON", "error", err)
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

			// Add genres to the genres set
			for _, genre := range book.Genres {
					genresSet[genre] = struct{}{}
			}

			// Add book to the booksByGenre map
			for _, genre := range book.Genres {
					booksByGenre[genre] = append(booksByGenre[genre], book)
			}
	}

	if err = rows.Err(); err != nil {
			r.Logger.Error("Error with rows", "error", err)
			return nil, err
	}

	// Convert the genres set to a sorted list
	var genres []string
	for genre := range genresSet {
			genres = append(genres, genre)
	}
	sort.Strings(genres)

	// Prepare the final result with genres and their associated books
	result := map[string]interface{}{
			"allGenres": genres,
	}

	for i, genre := range genres {
			key := strconv.Itoa(i)

			// Sort the books for each genre by author's last name
			sort.Slice(booksByGenre[genre], func(i, j int) bool {
					if len(booksByGenre[genre][i].Authors) > 0 && len(booksByGenre[genre][j].Authors) > 0 {
							return getLastName(booksByGenre[genre][i].Authors[0]) < getLastName(booksByGenre[genre][j].Authors[0])
					}
					return false
			})

			// Extract the first image for each book in the genre
			genreImgs := make([]string, len(booksByGenre[genre]))
			for j, book := range booksByGenre[genre] {
					if book.ImageLink != "" {
							genreImgs[j] = book.ImageLink
					}
			}

			result[key] = map[string]interface{}{
					"genreImgs": genreImgs,
					"bookList":  booksByGenre[genre],
			}
	}

	//r.Logger.Info("Final result being sent to the frontend", "result", result)
	return result, nil
}

func (r *BookRepositoryImpl) GetGenres(ctx context.Context, bookID int) ([]string, error) {
	// Check cache
	if cache, found := genresCache.Load(bookID); found {
		r.Logger.Info("Fetching genres from cache for book", "bookID", bookID)
		cachedGenres := cache.([]string)
		return append([]string(nil), cachedGenres...), nil
	}

	var rows *sql.Rows
	var err error

	// Use prepared statement if available
	if r.getGenresStmt != nil {
		r.Logger.Info("Using prepared statement for fetching genres")
		rows, err = r.getGenresStmt.QueryContext(ctx, bookID)
	} else {
		r.Logger.Warn("Prepared statement for fetching genres is not available. Falling back to raw SQL query")
		query := `
		SELECT g.name
		FROM genres g
		JOIN book_genres bg ON g.id = bg.genre_id
		WHERE bg.book_id = $1`
		rows, err = r.DB.QueryContext(ctx, query, bookID)
	}

	if err != nil {
		r.Logger.Error("Error fetching genres", "error", err)
		return nil, err
	}
	defer rows.Close()

	// Collect genres
	var genres []string
	for rows.Next() {
		var genre string
		if err := rows.Scan(&genre); err != nil {
			r.Logger.Error("Error scanning genre", "error", err)
			return nil, err
		}
		genres = append(genres, genre)
	}

	// Check for errors after looping through the rows
	if err = rows.Err(); err != nil {
		r.Logger.Error("Error with rows during genres fetch", "error", err)
		return nil, err
	}

	// Cache the result
	genresCache.Store(bookID, genres)
	r.Logger.Info("Caching genres for book", "bookID", bookID)

	return genres, nil
}
