package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"time"
)

type GenreRepository interface {
	InitPreparedStatements() error
	InsertGenre(ctx context.Context, tx *sql.Tx, genre string) (int, error)
	GetAllBooksByGenres(ctx context.Context, userID int) (map[string]interface{}, error)
	GetGenres(ctx context.Context, bookID int) ([]string, error)
	GetBooksListByGenre(ctx context.Context, userID int) (map[string]interface{}, error)
	GetGenreIDByName(ctx context.Context, tx *sql.Tx, genreName string, genreID *int) error
	AssociateBookWithGenre(ctx context.Context, tx *sql.Tx, bookID, genreID int) error
}

type GenreRepositoryImpl struct {
	DB                        *sql.DB
	Logger                    *slog.Logger
	getAllBooksByGenresStmt   *sql.Stmt
	getBookListByGenreStmt    *sql.Stmt
	getGenresStmt             *sql.Stmt
}


func NewGenreRepository(db *sql.DB, logger *slog.Logger) (GenreRepository, error) {
	if db == nil || logger == nil {
		return nil, fmt.Errorf("database or logger is nil")
	}

	return &GenreRepositoryImpl{
		DB:          db,
		Logger:      logger,
	}, nil
}

func (r *GenreRepositoryImpl) InitPreparedStatements() error {
	var err error

	// Prepared statment for GetAllBooksByGenres
	r.getAllBooksByGenresStmt, err = r.DB.Prepare(`
	SELECT r.id, r.title, r.subtitle, r.description, r.language, r.page_count, r.publish_date,
								 r.image_link, r.notes, r.created_at, r.last_updated, r.isbn_10, r.isbn_13,
								 json_agg(DISTINCT g.name) AS genres,
								 json_agg(DISTINCT a.name) AS authors,
								 json_agg(DISTINCT f.format_type) AS formats,
								 r.tags
					FROM books b
					INNER JOIN user_books ub ON r.id = ub.book_id
					LEFT JOIN book_genres bg ON r.id = bg.book_id
					LEFT JOIN genres g ON bg.genre_id = g.id
					LEFT JOIN book_authors ba ON r.id = ba.book_id
					LEFT JOIN authors a ON ba.author_id = a.id
					LEFT JOIN book_formats bf ON r.id = bf.book_id
					LEFT JOIN formats f ON bf.format_id = f.id
					WHERE ub.user_id = $1
					GROUP BY r.id`)
	if err != nil {
		return err
	}

	// Prepared select statement for GetGenres
	r.getGenresStmt, err = r.DB.Prepare(`
		SELECT g.name
		FROM genres g
		JOIN book_genres bg ON g.id = bg.genre_id
		WHERE bg.book_id = $1`)
	if err != nil {
		return err
	}

	r.getBookListByGenreStmt, err = r.DB.Prepare(`
	SELECT g.name, COUNT(DISTINCT r.id) AS total_books
		FROM books b
		INNER JOIN book_genres bg ON r.id = bg.book_id
		INNER JOIN genres g ON bg.genre_id = g.id
		INNER JOIN user_books ub ON r.id = ub.book_id
		WHERE ub.user_id = $1
		GROUP BY g.name
		ORDER BY total_books DESC`)
	if err != nil {
		return err
	}

	return nil
}


func (r *GenreRepositoryImpl) InsertGenre(ctx context.Context, tx *sql.Tx, genre string) (int, error) {
	var genreID int

	// Check if genre already exists
	err := tx.QueryRowContext(ctx, `SELECT id FROM genres WHERE name = $1`, genre).Scan(&genreID)
	if err == sql.ErrNoRows {
			// Genre doesn't exist, insert it
			err = tx.QueryRowContext(ctx, `INSERT INTO genres (name) VALUES ($1) RETURNING id`, genre).Scan(&genreID)
			if err != nil {
					r.Logger.Error("Error inserting genre", "error", genre)
					return 0, err
			}
	} else if err != nil {
			r.Logger.Error("Error checking if genre exists", "error", genre)
			return 0, err
	}

	return genreID, nil
}


func (r *GenreRepositoryImpl) GetGenreIDByName(ctx context.Context, tx *sql.Tx, genreName string, genreID *int) error {
	err := tx.QueryRowContext(ctx, `SELECT id FROM genres WHERE name = $1`, genreName).Scan(genreID)
	return err
}

func (r *GenreRepositoryImpl) GetAllBooksByGenres(ctx context.Context, userID int) (map[string]interface{}, error) {
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

func (r *GenreRepositoryImpl) GetGenres(ctx context.Context, bookID int) ([]string, error) {
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

func (b *GenreRepositoryImpl) GetBooksListByGenre(ctx context.Context, userID int) (map[string]interface{}, error) {
	const cacheTTL = time.Hour

	// Check cache with TTL
	if cacheEntry, found := booksByGenresCache.Load(userID); found {
			entry := cacheEntry.(BooksByGenresCacheEntry)
			if time.Since(entry.timestamp) < cacheTTL {
					b.Logger.Info("Fetching genres info from cache for user", "userID", userID)
					return entry.data, nil
			}
			booksByGenresCache.Delete(userID) // Cache entry expired, delete it
	}

	var rows *sql.Rows
	var err error

	// Use prepared statement if available
	if b.getBookListByGenreStmt != nil {
			b.Logger.Info("Using prepared statement for fetching genres")
			rows, err = b.getBookListByGenreStmt.QueryContext(ctx, userID)
	} else {
			b.Logger.Warn("Prepared statement for fetching genres unavailable. Falling back to raw SQL query")
			query := `
					SELECT g.name, COUNT(DISTINCT b.id) AS total_books
					FROM books b
					INNER JOIN book_genres bg ON b.id = bg.book_id
					INNER JOIN genres g ON bg.genre_id = g.id
					INNER JOIN user_books ub ON b.id = ub.book_id
					WHERE ub.user_id = $1
					GROUP BY g.name
					ORDER BY total_books DESC`
			rows, err = b.DB.QueryContext(ctx, query, userID)
	}

	if err != nil {
			b.Logger.Error("Error fetching books by genre", "error", err)
			return nil, err
	}
	defer rows.Close()

	// Collect genres and total book count
	var booksByGenre []map[string]interface{}

	for rows.Next() {
			var genre string
			var count int
			if err := rows.Scan(&genre, &count); err != nil {
					b.Logger.Error("Error scanning genre data", "error", err)
					return nil, err
			}

			booksByGenre = append(booksByGenre, map[string]interface{}{
					"label": genre,
					"count": count,
			})
	}

	if err = rows.Err(); err != nil {
			b.Logger.Error("Error iterating rows", "error", err)
			return nil, err
	}

	// Prepare the result
	result := map[string]interface{}{
			"booksByGenre": booksByGenre,
	}

	// Cache the result with TTL
	booksByGenresCache.Store(userID, BooksByGenresCacheEntry{
			data:      result,
			timestamp: time.Now(),
	})
	b.Logger.Info("Caching genres info for user", "userID", userID)

	return result, nil
}

func (b *BookModel) addOrGetGenreID(ctx context.Context, genreName string) (int, error) {
	var genreID int
	statement := `
		INSERT INTO genres (name)
		VALUES ($1)
		ON CONFLICT (name) DO UPDATE
		SET name = EXCLUDED.name
		RETURNING id`
	err := b.DB.QueryRowContext(ctx, statement, genreName).Scan(&genreID)
	if err != nil {
		b.Logger.Error("Error inserting or updating genre", "error", err)
		return 0, err
	}
	return genreID, nil
}

func (b *GenreRepositoryImpl) AssociateBookWithGenre(ctx context.Context, tx *sql.Tx, bookID, genreID int) error {
	statement := `INSERT INTO book_genres (book_id, genre_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := tx.ExecContext(ctx, statement, bookID, genreID)
	if err != nil {
		b.Logger.Error("Error adding author association", "error", err)
		return err
	}

	return nil
}