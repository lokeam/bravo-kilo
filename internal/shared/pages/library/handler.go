package library

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/lokeam/bravo-kilo/config"
	"github.com/lokeam/bravo-kilo/internal/books/handlers"
	"github.com/lokeam/bravo-kilo/internal/shared/cache"
	"github.com/lokeam/bravo-kilo/internal/shared/jwt"
	"github.com/lokeam/bravo-kilo/internal/shared/redis"
	"github.com/lokeam/bravo-kilo/internal/shared/types"
)

type Handler struct {
	domains map[string]types.DomainHandler
	redisClient   *redis.RedisClient
	logger        *slog.Logger
	cacheManager  *cache.CacheManager
	metrics       *LibraryMetrics
}

type LibraryResponse struct {
	RequestID string      `json:"requestId"`
	Data     interface{}  `json:"data"`
	Source   string       `json:"source"` // either "cache" || "database"
}

type LibraryMetrics struct {
	CacheHits   int64
	CacheMisses int64
	Errors      int64
}

type DomainType string

const (
    BooksDomain DomainType = "books"
    GamesDomain DomainType = "games"
)

func NewHandler(
	bookHandlers *handlers.BookHandlers, // initial domain
	redisClient *redis.RedisClient,
	logger *slog.Logger,
	cacheManager *cache.CacheManager,
) (*Handler, error) {

	if bookHandlers == nil {
		return nil, fmt.Errorf("bookHandlers cannot be nil")
	}
	if redisClient == nil {
			return nil, fmt.Errorf("redisClient cannot be nil")
	}
	if logger == nil {
			return nil, fmt.Errorf("logger cannot be nil")
	}
	if cacheManager == nil {
			return nil, fmt.Errorf("cacheManager cannot be nil")
	}

	return &Handler{
		domains:       make(map[string]types.DomainHandler),
		redisClient:   redisClient,
		logger:        logger,
		cacheManager:  cacheManager,
		metrics:       &LibraryMetrics{},
	}, nil
}

func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// 1. Auth check
	userID, err := jwt.ExtractUserIDFromJWT(r, config.AppConfig.JWTPublicKey)
	if err != nil {
			h.respondWithError(w, err, http.StatusUnauthorized)
			return
	}

	// 2. Get domain from query param
	domainType := DomainType(r.URL.Query().Get("domain"))
	if domainType == "" {
		domainType = BooksDomain // Default to books if not specified
	}

	// 3. Validate domain
	domainHandler, exists := h.domains[string(domainType)]
	if !exists {
			h.respondWithError(w, fmt.Errorf("invalid domain: %s", domainType), http.StatusBadRequest)
			return
	}

	// 4. Check cache
	cacheKey := fmt.Sprintf("page:library:%d:domain:%s", userID, domainType)
	if cached, err := h.redisClient.Get(ctx, cacheKey); err == nil {
			// Add metrics for cache hits/misses
			atomic.AddInt64(&h.metrics.CacheHits, 1)

			h.respondWithJSON(w, LibraryResponse{
					RequestID: uuid.New().String(),
					Data:     cached,
					Source:   "cache",
			})
			return
	} else {
		atomic.AddInt64(&h.metrics.CacheMisses, 1)
	}

	// Add error metrics
	if err != nil {
		atomic.AddInt64(&h.metrics.Errors, 1)
		h.respondWithError(w, err, http.StatusInternalServerError)
		return
	}

	// 5. Get domain data, call various domain handler methods (BookDomainHandler is default)
	data, err := domainHandler.GetLibraryItems(ctx, userID)
	if err != nil {
			h.respondWithError(w, err, http.StatusInternalServerError)
			return
	}

	// 6. Build response
	response := LibraryResponse{
			RequestID: uuid.New().String(),
			Data:     data,
			Source:   "database",
	}

	// 7. Cache response
	// todo: replace constant ttl with proper cacheManager config
	h.redisClient.Set(
		ctx,
		cacheKey,
		data,
		h.redisClient.GetConfig().CacheConfig.DefaultTTL,
	)

	// 8. Return response
	h.respondWithJSON(w, response)
}

func (h *Handler) RegisterDomain(domain types.DomainHandler) {
	if domain == nil {
		h.logger.Error("library handler attempted to register nil domain handler")
		return
	}

	if domain.GetType() == "" {
		h.logger.Error("domain type cannot be empty")
		return
	}

	domainType := domain.GetType()
	h.logger.Info("registering domain handler",
		"type", domainType,
	)

	h.domains[domainType] = domain
}

func (h *Handler) Cleanup() error {
	h.logger.Info("library handler cleanup",
	"cacheHits", atomic.LoadInt64(&h.metrics.CacheHits),
	"cacheMisses", atomic.LoadInt64(&h.metrics.CacheMisses),
	"errors", atomic.LoadInt64(&h.metrics.Errors),
)
return nil
}

// Helper functions
// Error response following your existing pattern
func (h *Handler) respondWithError(w http.ResponseWriter, err error, status int) {
	h.logger.Error("handler error",
			"error", err,
			"status", status,
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
	})
}

func (h *Handler) respondWithJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
			h.respondWithError(w, err, http.StatusInternalServerError)
			return
	}
}
