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
	GetBooksListByGenre(ctx context.Context, userID int) (map[string]interface{}, error)
	GetGenreIDByName(ctx context.Context, tx *sql.Tx, genreName string, genreID *int) error
	AssociateBookWithGenre(ctx context.Context, tx *sql.Tx, bookID, genreID int) error
}

type GenreRepositoryImpl struct {
	DB                        *sql.DB
	Logger                    *slog.Logger
	getAllBooksByGenresStmt   *sql.Stmt
	getBookListByGenreStmt    *sql.Stmt
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
	SELECT b.id, b.title, b.subtitle, b.description, b.language, b.page_count, b.publish_date,
	       b.image_link, b.notes, b.created_at, b.last_updated, b.isbn_10, b.isbn_13,
	       json_agg(DISTINCT g.name) AS genres,
	       json_agg(DISTINCT a.name) AS authors,
	       json_agg(DISTINCT f.format_type) AS formats,
	       json_agg(DISTINCT t.name) AS tags
	FROM books b
	INNER JOIN user_books ub ON b.id = ub.book_id
	LEFT JOIN book_genres bg ON b.id = bg.book_id
	LEFT JOIN genres g ON bg.genre_id = g.id
	LEFT JOIN book_authors ba ON b.id = ba.book_id
	LEFT JOIN authors a ON ba.author_id = a.id
	LEFT JOIN book_formats bf ON b.id = bf.book_id
	LEFT JOIN formats f ON bf.format_id = f.id
	LEFT JOIN book_tags bt ON b.id = bt.book_id
	LEFT JOIN tags t ON bt.tag_id = t.id
	WHERE ub.user_id = $1
	GROUP BY b.id`)
if err != nil {
	return err
}
	r.Logger.Info("Successfully prepared getAllBooksByGenresStmt")

	r.getBookListByGenreStmt, err = r.DB.Prepare(`
	SELECT g.name, COUNT(DISTINCT b.id) AS total_books
		FROM books b
		INNER JOIN book_genres bg ON b.id = bg.book_id
		INNER JOIN genres g ON bg.genre_id = g.id
		INNER JOIN user_books ub ON b.id = ub.book_id
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

	// Try re-initializing if the prepared statement is nil
	if r.getAllBooksByGenresStmt == nil {
			r.Logger.Warn("getAllBooksByGenresStmt is nil, attempting to reinitialize")
			err = r.InitPreparedStatements()
			if err != nil {
					r.Logger.Error("Failed to re-initialize prepared statements", "error", err)
					return nil, fmt.Errorf("failed to initialize prepared statements: %w", err)
			}
			r.Logger.Info("Successfully reinitialized getAllBooksByGenresStmt")
	}

	// Use the prepared statement if available, else fall back to a raw query
	if r.getAllBooksByGenresStmt != nil {
			r.Logger.Info("Using prepared statement for retrieving books by genres")
			rows, err = r.getAllBooksByGenresStmt.QueryContext(ctx, userID)
	} else {
			r.Logger.Info("Prepared statement unavailable, using fallback query for retrieving books by genres")
			query := `
					SELECT b.id, b.title, b.subtitle, b.description, b.language, b.page_count, b.publish_date,
								 b.image_link, b.notes, b.created_at, b.last_updated, b.isbn_10, b.isbn_13,
								 json_agg(DISTINCT g.name) AS genres,
								 json_agg(DISTINCT a.name) AS authors,
								 json_agg(DISTINCT f.format_type) AS formats,
								 json_agg(DISTINCT t.name) AS tags,
					FROM books b
					INNER JOIN user_books ub ON b.id = ub.book_id
					LEFT JOIN book_genres bg ON b.id = bg.book_id
					LEFT JOIN genres g ON bg.genre_id = g.id
					LEFT JOIN book_authors ba ON b.id = ba.book_id
					LEFT JOIN authors a ON ba.author_id = a.id
					LEFT JOIN book_formats bf ON b.id = bf.book_id
					LEFT JOIN formats f ON bf.format_id = f.id
					LEFT JOIN book_tags bt ON b.id = bt.book_id
					LEFT JOIN tags t ON bt.tag_id = t.id
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

func (b *GenreRepositoryImpl) AssociateBookWithGenre(ctx context.Context, tx *sql.Tx, bookID, genreID int) error {
	statement := `INSERT INTO book_genres (book_id, genre_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := tx.ExecContext(ctx, statement, bookID, genreID)
	if err != nil {
		b.Logger.Error("Error adding author association", "error", err)
		return err
	}

	return nil
}
