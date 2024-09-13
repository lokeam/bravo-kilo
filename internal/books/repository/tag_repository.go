package repository

import (
	"container/heap"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/lokeam/bravo-kilo/internal/shared/customheap"
)

type TagRepository interface {
	InitPreparedStatements() error
	GetUserTags(ctx context.Context, userID int) (map[string]interface{}, error)
	GetTagsForBook(ctx context.Context, bookID int) ([]string, error)
}

type TagRepositoryImpl struct {
	DB               *sql.DB
	Logger           *slog.Logger
	getUserTagsStmt  *sql.Stmt
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
	SELECT r.tags
		FROM books b
		INNER JOIN user_books ub ON r.id = ub.book_id
		WHERE ub.user_id = $1
	`)
	if err != nil {
		return err
	}

	return nil
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
			SELECT b.tags
			FROM books b
			INNER JOIN user_books ub ON b.id = ub.book_id
			WHERE ub.user_id = $1`
			rows, err = b.DB.QueryContext(ctx, query, userID)
	}

	if err != nil {
			b.Logger.Error("Error fetching user tags", "error", err)
			return nil, err
	}
	defer rows.Close()

	// Process the tags and count occurrences
	tagCount := make(map[string]int)

	for rows.Next() {
			var tagsJSON []byte
			if err := rows.Scan(&tagsJSON); err != nil {
					b.Logger.Error("Error scanning tags", "error", err)
					return nil, err
			}

			var tags []string
			if err := json.Unmarshal(tagsJSON, &tags); err != nil {
					b.Logger.Error("Error unmarshalling tags JSON", "error", err)
					return nil, err
			}

			// Convert spaces to underscores and count occurrences
			for _, tag := range tags {
					formattedTag := strings.ReplaceAll(tag, " ", "_")
					tagCount[formattedTag]++
			}
	}

	// Create a max heap for more efficient sorting
	h := &customheap.TagHeap{}
	heap.Init(h)

	for tag, count := range tagCount {
		heap.Push(h, customheap.TagCount{Tag: tag, Count: count})
	}

	// Create the result array with the new format
	userTags := make([]map[string]interface{}, 0, len(tagCount))
	for h.Len() > 0 {
		tagCount := heap.Pop(h).(customheap.TagCount)
		userTags = append(userTags, map[string]interface{}{
					"label":   tagCount.Tag,
					"count": tagCount.Count,
			})
	}

	// Prepare the result
	result := map[string]interface{}{
			"userTags":       userTags,
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