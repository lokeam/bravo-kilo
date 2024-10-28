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

type TagRepository interface {
	InitPreparedStatements() error
	InsertTag(ctx context.Context, tx *sql.Tx, tag string) (int, error)
	GetUserTags(ctx context.Context, userID int) (map[string]interface{}, error)
	GetTagsForBook(ctx context.Context, bookID int) ([]string, error)
	GetTagIDByName(ctx context.Context, tx *sql.Tx, tagName string, tagID *int) error
	GetAllBooksByTags(ctx context.Context, userID int) (map[string]interface{}, error)
	AssociateBookWithTag(ctx context.Context, tx *sql.Tx, bookID, tagID int) error
}

type TagRepositoryImpl struct {
	DB                     *sql.DB
	Logger                 *slog.Logger
	getUserTagsStmt        *sql.Stmt
	getAllBooksByTagsStmt  *sql.Stmt
}

func NewTagRepository(db *sql.DB, logger *slog.Logger) (TagRepository, error) {
	if db == nil || logger == nil {
		return nil, fmt.Errorf("database or logger is nil")
	}

	return &TagRepositoryImpl{
		DB:          db,
		Logger:      logger,
	}, nil
}

func (r *TagRepositoryImpl) InitPreparedStatements() error {
	var err error

	r.getUserTagsStmt, err = r.DB.Prepare(`
		SELECT t.name, b.title
		FROM books b
		INNER JOIN user_books ub ON b.id = ub.book_id
		INNER JOIN book_tags bt ON b.id = bt.book_id
		INNER JOIN tags t ON bt.tag_id = t.id
		WHERE ub.user_id = $1`)
	if err != nil {
		return err
	}

	r.getAllBooksByTagsStmt, err = r.DB.Prepare(`
	SELECT b.id, b.title, b.subtitle, b.description, b.language, b.page_count, b.publish_date,
				 b.image_link, b.notes, b.created_at, b.last_updated, b.isbn_10, b.isbn_13,
					json_agg(DISTINCT t.name) AS tags,
					json_agg(DISTINCT a.name) AS authors,
					json_agg(DISTINCT f.format_type) AS formats,
					json_agg(DISTINCT g.name) AS genres
	FROM books b
	INNER JOIN user_books ub ON b.id = ub.book_id
	LEFT JOIN book_tags bt ON b.id = bt.book_id
	LEFT JOIN tags t ON bt.tag_id = t.id
	LEFT JOIN book_authors ba ON b.id = ba.book_id
	LEFT JOIN authors a ON ba.author_id = a.id
	LEFT JOIN book_formats bf ON b.id = bf.book_id
	LEFT JOIN formats f ON bf.format_id = f.id
	LEFT JOIN book_genres bg ON b.id = bg.book_id
	LEFT JOIN genres g ON bg.genre_id = g.id
	WHERE ub.user_id = $1
	GROUP BY b.id`)
	if err != nil {
	return err
	}
	r.Logger.Info("Successfully prepared getAllBooksByTagsStmt")

	return nil
}

func (b *TagRepositoryImpl) GetAllBooksByTags(ctx context.Context, userID int) (map[string]interface{}, error) {
	var rows *sql.Rows
	var err error

	// Try re-initializing if the prepared statement is nil
	if b.getAllBooksByTagsStmt == nil {
			b.Logger.Warn("getAllBooksByTagsStmt is nil, attempting to reinitialize")
			err = b.InitPreparedStatements()
			if err != nil {
					b.Logger.Error("Failed to re-initialize prepared statements", "error", err)
					return nil, fmt.Errorf("failed to initialize prepared statements: %w", err)
			}
			b.Logger.Info("Successfully reinitialized getAllBooksByTagsStmt")
	}

	// Use the prepared statement if available, else fall back to a raw query
	if b.getAllBooksByTagsStmt != nil {
			b.Logger.Info("Using prepared statement for retrieving books by tags")
			rows, err = b.getAllBooksByTagsStmt.QueryContext(ctx, userID)
	} else {
			b.Logger.Info("Prepared statement unavailable, using fallback query for retrieving books by genres")
			query := `
					SELECT b.id, b.title, b.subtitle, b.description, b.language, b.page_count, b.publish_date,
								 b.image_link, b.notes, b.created_at, b.last_updated, b.isbn_10, b.isbn_13,
								 json_agg(DISTINCT t.name) AS tags,
								 json_agg(DISTINCT a.name) AS authors,
								 json_agg(DISTINCT f.format_type) AS formats,
								 json_agg(DISTINCT g.name) AS genres
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
			rows, err = b.DB.QueryContext(ctx, query, userID)
	}

	if err != nil {
			b.Logger.Error("Error retrieving books by tags", "error", err)
			return nil, err
	}
	defer rows.Close()

	tagsSet := make(map[string]struct{}) // Track unique tag
	booksByTag := map[string][]Book{}    // Store books by tag

	// Iterate through the rows and process the results
	for rows.Next() {
			var book Book
			var tagsJSON, authorsJSON, formatsJSON, genresJSON, descriptionJSON, notesJSON []byte

			// Ensure the scan order matches the SQL query's column order
			if err := rows.Scan(
					&book.ID, &book.Title, &book.Subtitle, &descriptionJSON, &book.Language, &book.PageCount,
					&book.PublishDate, &book.ImageLink, &notesJSON, &book.CreatedAt, &book.LastUpdated,
					&book.ISBN10, &book.ISBN13, &tagsJSON, &authorsJSON, &formatsJSON, &genresJSON,
			); err != nil {
					b.Logger.Error("Error scanning book by genre", "error", err)
					return nil, err
			}

			// Unmarshal JSON fields into the respective slices
			if err := json.Unmarshal(descriptionJSON, &book.Description); err != nil {
				b.Logger.Error("Error unmarshalling description JSON", "error", err)
				return nil, err
			}
			if err := json.Unmarshal(notesJSON, &book.Notes); err != nil {
				b.Logger.Error("Error unmarshalling notes JSON", "error", err)
				return nil, err
			}
			if err := json.Unmarshal(tagsJSON, &book.Tags); err != nil {
				b.Logger.Error("Error unmarshalling tags JSON", "error", err)
				return nil, err
			}
			if err := json.Unmarshal(authorsJSON, &book.Authors); err != nil {
					b.Logger.Error("Error unmarshalling authors JSON", "error", err)
					return nil, err
			}
			if err := json.Unmarshal(formatsJSON, &book.Formats); err != nil {
					b.Logger.Error("Error unmarshalling formats JSON", "error", err)
					return nil, err
			}
			if err := json.Unmarshal(genresJSON, &book.Genres); err != nil {
				b.Logger.Error("Error unmarshalling genres JSON", "error", err)
				return nil, err
		}

			// Add tags to the tags set, ignore empty strings
			for _, tag := range book.Tags {
					if tag != "" {
						tagsSet[tag] = struct{}{}
					}
			}

			// Add book to the booksByTag map
			for _, tag := range book.Tags {
				booksByTag[tag] = append(booksByTag[tag], book)
			}
	}

	if err = rows.Err(); err != nil {
			b.Logger.Error("Error with rows", "error", err)
			return nil, err
	}

	// Convert the genres set to a sorted list
	var tags []string
	for tag := range tagsSet {
			tags = append(tags, tag)
	}
	sort.Strings(tags)

	// Prepare the final result with tags and their associated books
	result := map[string]interface{}{
			"allTags": tags,
	}

	for i, tag := range tags {
			key := strconv.Itoa(i)

			// Sort the books for each tag by author's last name
			sort.Slice(booksByTag[tag], func(i, j int) bool {
				if len(booksByTag[tag][i].Authors) > 0 && len(booksByTag[tag][j].Authors) > 0 {
						return getLastName(booksByTag[tag][i].Authors[0]) < getLastName(booksByTag[tag][j].Authors[0])
				}
				return false
			})

			// Extract the first image for each book in the genre
			tagImgs := make([]string, len(booksByTag[tag]))
			for j, book := range booksByTag[tag] {
					if book.ImageLink != "" {
							tagImgs[j] = book.ImageLink
					}
			}

			result[key] = map[string]interface{}{
				"tagImgs":  tagImgs,  // Correctly named here
				"bookList": booksByTag[tag],
		}
	}

	// b.Logger.Info("Tag Repo,  result being sent to the frontend", "result", result)
	return result, nil
}

func (b *TagRepositoryImpl) GetUserTags(ctx context.Context, userID int) (map[string]interface{}, error) {
	// Check cache with TTL
	if cacheEntry, found := userTagsCache.Load(userID); found {
		entry := cacheEntry.(UserTagsCacheEntry)
		if time.Since(entry.timestamp) < time.Hour {
			b.Logger.Info("Fetching user tags from cache for user", "userID", userID)
			return entry.data, nil
		}
		// Cache entry expired, delete it
		userTagsCache.Delete(userID)
	}

	var rows *sql.Rows
	var err error

	// Use prepared statement if available
	if b.getUserTagsStmt != nil {
		b.Logger.Info("Using prepared statement for fetching user tags")
		rows, err = b.getUserTagsStmt.QueryContext(ctx, userID)
	} else {
		b.Logger.Warn("Prepared statement for fetching user tags unavailable. Falling back to raw SQL query")
		query := `
		SELECT t.name, b.title
		FROM books b
		INNER JOIN user_books ub ON b.id = ub.book_id
		INNER JOIN book_tags bt ON b.id = bt.book_id
		INNER JOIN tags t ON bt.tag_id = t.id
		WHERE ub.user_id = $1`
		rows, err = b.DB.QueryContext(ctx, query, userID)
	}

	if err != nil {
		b.Logger.Error("Error fetching user tags", "error", err)
		return nil, err
	}
	defer rows.Close()

	// Process the tags and count occurrences, also group books by tags
	tagBooksMap := make(map[string][]string)
	tagCount := make(map[string]int)

	for rows.Next() {
		var tagName, bookTitle string
		if err := rows.Scan(&tagName, &bookTitle); err != nil {
			b.Logger.Error("Error scanning tags and books", "error", err)
			return nil, err
		}

		// Update tag count and group books
		tagCount[tagName]++
		tagBooksMap[tagName] = append(tagBooksMap[tagName], bookTitle)
	}

	// Create the result array with tags and associated books
	userTags := make([]map[string]interface{}, 0, len(tagCount))
	for tagName, count := range tagCount {
		userTags = append(userTags, map[string]interface{}{
			"label": tagName,
			"count": count,
			"books": tagBooksMap[tagName],
		})
	}

	// Prepare the result
	result := map[string]interface{}{
		"userTags": userTags,
	}

	// Cache the result
	userTagsCache.Store(userID, UserTagsCacheEntry{data: result, timestamp: time.Now()})
	b.Logger.Info("Caching user tags for user", "userID", userID)

	return result, nil
}

func (b *TagRepositoryImpl) GetTagsForBook(ctx context.Context, bookID int) ([]string, error) {

	query := `
	SELECT tags
	FROM books
	WHERE id = $1`

	var tagsJSON []byte
	err := b.DB.QueryRowContext(ctx, query, bookID).Scan(&tagsJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			b.Logger.Warn("No tags found for book", "bookID", bookID)
			// Return empty slice if no tags are found
			return nil, nil
		}
		b.Logger.Error("Error fetching tags for book", "error", err)
		return nil, err
	}

	// Return an empty slice if tagsJSON is null
	if tagsJSON == nil {
		return []string{}, nil
	}

	// Unmarshal the JSON array of tags
	var tags []string
	if err := json.Unmarshal(tagsJSON, &tags); err != nil {
		b.Logger.Error("Error unmarshalling tags JSON", "error", err)
		return nil, err
	}

	return tags, nil
}

func (b *TagRepositoryImpl) GetTagIDByName(ctx context.Context, tx *sql.Tx, tagName string, tagID *int) error {
	err := tx.QueryRowContext(ctx, `SELECT id FROM tags WHERE name = $1`, tagName).Scan(tagID)
	return err
}

func (b *TagRepositoryImpl) InsertTag(ctx context.Context, tx *sql.Tx, tag string) (int, error) {
	b.Logger.Info("calling insert Tag")
	var tagID int

	err := tx.QueryRowContext(ctx, `SELECT id FROM tags WHERE name = $1`, tag).Scan(&tagID)
	if err == sql.ErrNoRows {
		// Tag doesn't exist, insert it
		err = tx.QueryRowContext(ctx, `INSERT INTO tags (name) VALUES ($1) RETURNING id`, tag).Scan(&tagID)
		if err != nil {
			b.Logger.Error("Error inserting new tag", "error", err, "tag", tag)
			return 0, err
		}
	} else if err != nil {
		b.Logger.Error("Error checking if tag exists", "error", err, "tag", tag)
		return 0, err
	}

	return tagID, nil
}

func (b *TagRepositoryImpl) AssociateBookWithTag(ctx context.Context, tx *sql.Tx, bookID, tagID int) error {
	statement := `INSERT INTO book_tags (book_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := tx.ExecContext(ctx, statement, bookID, tagID)
	if err != nil {
		b.Logger.Error("Error adding tag association", "error", err)
		return err
	}

	b.Logger.Info("associated book with tag")
	return nil
}
