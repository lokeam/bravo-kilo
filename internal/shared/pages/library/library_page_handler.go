package library

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/lokeam/bravo-kilo/config"
	"github.com/lokeam/bravo-kilo/internal/books/handlers"
	"github.com/lokeam/bravo-kilo/internal/shared/cache"
	"github.com/lokeam/bravo-kilo/internal/shared/jwt"
	"github.com/lokeam/bravo-kilo/internal/shared/organizer"
	"github.com/lokeam/bravo-kilo/internal/shared/processor/bookprocessor"
	"github.com/lokeam/bravo-kilo/internal/shared/redis"
	"github.com/lokeam/bravo-kilo/internal/shared/types"
	"github.com/lokeam/bravo-kilo/internal/shared/validator"
	goredis "github.com/redis/go-redis/v9"
)

type LibraryPageHandler struct {
	mu                 sync.RWMutex
	domainHandlers     map[types.DomainType]types.DomainHandler
	processor          *bookprocessor.BookProcessor
	redisClient        *redis.RedisClient
	logger             *slog.Logger
	cacheManager       *cache.CacheManager
	metrics            *LibraryMetrics
	organizer          *organizer.BookOrganizer
	requestTimeout     time.Duration
	baseValidator      *validator.BaseValidator
	validationRules    validator.QueryValidationRules
}

type LibraryResponse struct {
	RequestID string      `json:"requestId"`
	Data     interface{}  `json:"data"`
	Source   string       `json:"source"` // either "cache" || "database"
}

type LibraryMetrics struct {
	CacheHits                     int64
	CacheMisses                   int64
	Errors                        int64
	CacheOperationDuration        int64
	ValidationTotalAttempts       int64
	ValidationTotalErrors         int64
	ValidationTotalDuration       int64
	ValidationMaxDuration         int64
	ValidationMinDuration         int64
}

type LibraryQueryParams struct {
	Page  int    `json:"page" validate:"required,min=1,max=99999"`
	Limit int    `json:"limit" validate:"required,min=1,max=999"`
	Sort  string `json:"sort" validate:"omitempty,oneof=asc desc"`
}

const (
    BooksDomain types.DomainType = "books"
    GamesDomain types.DomainType = "games"
)

func NewLibraryPageHandler(
	bookHandlers *handlers.BookHandlers,
	redisClient *redis.RedisClient,
	logger *slog.Logger,
	cacheManager *cache.CacheManager,
	baseValidator *validator.BaseValidator,
) (*LibraryPageHandler, error) {

	// Validate dependencies
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
	if baseValidator == nil {
		return nil, fmt.Errorf("baseValidator cannot be nil")
	}

	// Init processor + organizer
	processor, err := bookprocessor.NewBookProcessor(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create book processor: %w", err)
	}

	organizer, err := organizer.NewBookOrganizer(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create book organizer: %w", err)
	}

	// Init validator
	baseValidator, err = validator.NewBaseValidator(logger, validator.BookDomain)
	if err != nil {
			return nil, fmt.Errorf("failed to initialize validator: %w", err)
	}

	// Set validation rules
	validationRules := validator.QueryValidationRules{
		"page": {
			Required: true,
			Type:     validator.QueryTypeInt,
			MinLength: 1,
			MaxLength: 5,
		},
		"limit": {
				Required: true,
				Type:     validator.QueryTypeInt,
				MinLength: 1,
				MaxLength: 3,
		},
		"sort": {
				Required: false,
				Type:     validator.QueryTypeString,
				AllowedValues: []string{"asc", "desc"},
		},
}

	h := &LibraryPageHandler{
		domainHandlers:  make(map[types.DomainType]types.DomainHandler),
		processor:       processor,
		organizer:       organizer,
		redisClient:     redisClient,
		logger:          logger,
		cacheManager:    cacheManager,
		metrics:         &LibraryMetrics{},
		requestTimeout:  5 * time.Second,
		baseValidator:   baseValidator,
		validationRules: validationRules,
	}

	return h, nil
}

func (h *LibraryPageHandler) HandleGetLibraryPageData(w http.ResponseWriter, r *http.Request) {
	// 1. Initial Setup
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
			requestID = uuid.New().String()
	}

	// 2. Request Tracking
	requestStart := time.Now()
	defer func() {
			h.logger.Info("completed library page request",
					"requestID", requestID,
					"duration", time.Since(requestStart),
					"cacheHits", atomic.LoadInt64(&h.metrics.CacheHits),
					"cacheMisses", atomic.LoadInt64(&h.metrics.CacheMisses),
					"errors", atomic.LoadInt64(&h.metrics.Errors),
			)
	}()

	// 3. JWT Authentication (Primary Security Check)
	userID, err := jwt.ExtractUserIDFromJWT(r, config.AppConfig.JWTPublicKey)
	if err != nil {
			h.logger.Error("authentication failed",
					"requestID", requestID,
					"error", err,
			)
			h.respondWithError(w, requestID, err, http.StatusUnauthorized)
			return
	}
	h.logger.Debug("user authenticated",
			"requestID", requestID,
			"userID", userID,
	)

	// 4. Context Setup
	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, h.requestTimeout)
	if ctx == nil {
		h.logger.Error("nil context received",
			"requestID", requestID,
		)
		h.respondWithError(w, requestID, fmt.Errorf("nil context received"), http.StatusInternalServerError)
		return
	}
	defer cancel()

	// 5. Start validation timing
	validationStart := time.Now()
	defer func() {
		duration := time.Since(validationStart)
		if ctx.Err() == nil {
				h.metrics.recordValidationDuration(time.Since(validationStart))
		}
		h.metrics.recordValidationDuration(duration)
	}()

	h.logger.Info("starting library page request",
			"requestID", requestID,
			"method", r.Method,
			"path", r.URL.Path,
	)

    // Parse and validate query parameters
    queryParams := LibraryQueryParams{
			Page:  1,  // Default values
			Limit: 10,
		}

		if err := h.parseQueryParams(r, &queryParams); err != nil {
				h.logger.Error("invalid query parameters",
						"requestID", requestID,
						"error", err,
				)
				h.respondWithError(w, requestID, err, http.StatusBadRequest)
				return
		}

	// 6. Context check after time consuming operations
	if ctx.Err() != nil {
    h.logger.Error("request timeout",
        "requestID", requestID,
        "error", ctx.Err(),
    )
    h.respondWithError(w, requestID, ctx.Err(), http.StatusGatewayTimeout)
    return
	}

	// 7. Get domain from query param
	domainType := types.DomainType(r.URL.Query().Get("domain"))
	if domainType == "" {
		h.logger.Debug("no domain specified, using default",
		"requestID", requestID,
		"defaultDomain", BooksDomain,
		)

		domainType = types.BookDomainType // Default to books if not specified
	}
	if err := h.baseValidator.ValidateField("domain", string(domainType)); err != nil {
		h.logger.Error("domain validation failed",
			"requestID", requestID,
			"domain", domainType,
			"error", err,
		)
		atomic.AddInt64(&h.metrics.ValidationTotalErrors, 1)
		h.respondWithError(w, requestID, err, http.StatusBadRequest)
		return
	}

	// 8. Validate domain type
	domainHandler, exists := h.domainHandlers[domainType]
	if !exists {
			h.logger.Error("invalid domain requested",
					"requestID", requestID,
					"domain", domainType,
					"availableDomains", getAvailableDomains(h.domainHandlers),
			)
			h.respondWithError(w, requestID, fmt.Errorf("invalid domain: %s", domainType), http.StatusBadRequest)
			return
	}
	h.logger.Debug("domain handler found",
	"requestID", requestID,
	"domain", domainType,
	)

	// 9. Check cache
	cacheKey := fmt.Sprintf("page:library:%d:domain:%s", userID, domainType)
	h.logger.Debug("checking cache",
			"requestID", requestID,
			"cacheKey", cacheKey,
	)

	var libraryData *types.LibraryPageData
	cacheStart := time.Now()
	cached, err := h.redisClient.Get(ctx, cacheKey)
	cacheDuration := time.Since(cacheStart)
	atomic.AddInt64(&h.metrics.CacheOperationDuration, cacheDuration.Nanoseconds())
	h.logger.Debug("cache operation completed",
    "requestID", requestID,
    "operation", "GET",
		"duration", cacheDuration,
	)
	if err == nil {
		libraryData = types.NewLibraryPageData(h.logger)
		if err := libraryData.UnmarshalBinary([]byte(cached)); err != nil {
			atomic.AddInt64(&h.metrics.Errors, 1)
			h.logger.Error("failed to unmarshal cached data in library page handler",
				"requestID", requestID,
				"error", err,
				"cacheKey", cacheKey,
			)
			// Continue to fetch data from database
		} else {
			// Validate cached data before using
			if err := h.baseValidator.ValidateStruct(ctx, libraryData); err != nil {
				h.logger.Warn("cached data validation failed, fetching fresh data",
					"requestID", requestID,
					"error", err,
				)
			} else {
				h.logger.Debug("cached data validated",
					"requestID", requestID,
					"cacheKey", cacheKey,
				)
				// Cache hit
				h.incrementMetric(&h.metrics.CacheHits)
				h.respondWithJSON(w, requestID, LibraryResponse{
						RequestID: requestID,
						Data:     libraryData,
						Source:   "cache",
				})
				return
			}
		}
	}

	// Add error metrics
	if err != nil {
		if err == goredis.Nil {
			h.logger.Debug("cache miss",
			"requestID", requestID,
			"error", err,
				"cacheKey", cacheKey,
			)
			atomic.AddInt64(&h.metrics.CacheMisses, 1)
		} else {
			h.logger.Error("cache error",
				"requestID", requestID,
				"error", err,
				"cacheKey", cacheKey,
			)
			atomic.AddInt64(&h.metrics.Errors, 1)
		}
	}

	// 10. Get domain data, call various domain handler methods (BookDomainHandler is default)
	if ctx.Err() != nil {
		h.respondWithError(w, requestID, ctx.Err(), http.StatusGatewayTimeout)
		return
	}

	// Ensure domain handler isn't nil before calling methods on it
	if domainHandler == nil {
		h.logger.Error("nil domain handler",
				"requestID", requestID,
				"domain", domainType,
		)
		h.respondWithError(w, requestID, fmt.Errorf("internal server error"), http.StatusInternalServerError)
		return
	}

	items, err := domainHandler.GetLibraryItems(ctx, userID)
	if err != nil {
		h.logger.Error("failed to get library items",
			"requestID", requestID,
			"error", err,
		)
		h.respondWithError(w, requestID, err, http.StatusInternalServerError)
		return
	}

	h.logger.Debug("items retrieved",
	"requestID", requestID,
		"itemCount", len(items),
	)

	// 11. Process + organize data
	processedData, err := h.processor.ProcessLibraryItems(ctx, items)
	if err != nil {
		h.logger.Error("processing failed",
		"requestID", requestID,
		"error", err,
	)
		h.respondWithError(w, requestID, err, http.StatusInternalServerError)
		return
	}

	// Type assertion for processed data
	typedData, ok := processedData.(*types.LibraryPageData)
	if !ok {
			err := fmt.Errorf("invalid data type from processor: expected *types.LibraryPageData, got %T", processedData)
			h.logger.Error("type assertion failed",
					"requestID", requestID,
					"error", err,
			)
			h.respondWithError(w, requestID, err, http.StatusInternalServerError)
			return
	}

	h.logger.Debug("organizing processed data",
	"requestID", requestID,
	)
	organizedData := types.NewLibraryPageData(h.logger)

  // Perform thread-safe deep copy
  copyStart := time.Now()
  if err := organizedData.DeepCopy(typedData); err != nil {
			h.logger.Error("deep copy failed",
					"requestID", requestID,
					"error", err,
					"duration", time.Since(copyStart),
			)
			h.respondWithError(w, requestID, err, http.StatusInternalServerError)
			return
	}
	h.logger.Debug("deep copy completed",
			"requestID", requestID,
			"duration", time.Since(copyStart),
	)

	pageData := LibraryResponse{
		RequestID: requestID,
		Data:      organizedData,
		Source:    "database",
	}

	// 12. Cache response
	// todo: replace constant ttl with proper cacheManager config
	h.logger.Debug("caching response",
		"requestID", requestID,
		"cacheKey", cacheKey,
		"ttl", h.redisClient.GetConfig().CacheConfig.DefaultTTL,
	)
	cacheStart = time.Now()
	err = h.redisClient.Set(ctx, cacheKey, organizedData, h.redisClient.GetConfig().CacheConfig.DefaultTTL)
	cacheDuration = time.Since(cacheStart)
	atomic.AddInt64(&h.metrics.CacheOperationDuration, cacheDuration.Nanoseconds())

	if err != nil {
			atomic.AddInt64(&h.metrics.Errors, 1)
			h.logger.Error("failed to cache response",
					"requestID", requestID,
					"error", err,
					"cacheKey", cacheKey,
					"duration", cacheDuration,
			)
	}

	// 13. Return response
	h.logger.Info("sending response",
	"requestID", requestID,
	"source", "database",
	)
	h.respondWithJSON(w, requestID, pageData)
}

func (h *LibraryPageHandler) RegisterDomain(domain types.DomainHandler) error {
	if domain == nil {
		return fmt.Errorf("domain handler cannot be nil")
	}

	domainType := domain.GetType()
	h.logger.Info("registering domain handler",
			"type", domainType,
	)

	h.mu.Lock()
	h.domainHandlers[domainType] = domain
	h.mu.Unlock()

	h.logger.Info("domain handler registered",
			"type", domainType,
			"availableDomains", getAvailableDomains(h.domainHandlers),
	)
	return nil
}

func (h *LibraryPageHandler) Cleanup() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	cleanupStart := time.Now()
	defer func() {
		if h.baseValidator != nil {
			h.baseValidator.Cleanup()
		}
		h.logger.Info("cleanup completed",
				"duration", time.Since(cleanupStart),
		)
	}()
	h.logger.Info("library handler cleanup",
			"cacheHits", atomic.LoadInt64(&h.metrics.CacheHits),
			"cacheMisses", atomic.LoadInt64(&h.metrics.CacheMisses),
			"errors", atomic.LoadInt64(&h.metrics.Errors),
			"cacheOperationDuration", time.Duration(atomic.LoadInt64(&h.metrics.CacheOperationDuration)),
			"validationTotal", atomic.LoadInt64(&h.metrics.ValidationTotalAttempts),
			"validationErrors", atomic.LoadInt64(&h.metrics.ValidationTotalErrors),
			"validationAvgDuration", float64(atomic.LoadInt64(&h.metrics.ValidationTotalDuration)) / float64(atomic.LoadInt64(&h.metrics.ValidationTotalAttempts)),
			"validationMaxDuration", time.Duration(atomic.LoadInt64(&h.metrics.ValidationMaxDuration)),
			"validationMinDuration", time.Duration(atomic.LoadInt64(&h.metrics.ValidationMinDuration)),
	)

	// Clear domain handler maps
	h.domainHandlers = make(map[types.DomainType]types.DomainHandler)
	return nil
}

// Helper functions
// Error response following your existing pattern
func (h *LibraryPageHandler) respondWithError(w http.ResponseWriter, requestID string, err error, status int) {
	// Guard clause
	if w == nil {
		h.logger.Error("nil response writer in error response",
			"requestID", requestID,
			"error",err,
		)
		return
	}

	h.logger.Error("handler error",
		"requestID", requestID,
		"error", err,
		"status", status,
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
	})
}

func (h *LibraryPageHandler) respondWithJSON(w http.ResponseWriter, requestID string, data interface{}) {
	// Guard clause
	if w == nil {
		h.logger.Error("nil response writer",
			"requestID", requestID,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
			h.respondWithError(w, requestID, err, http.StatusInternalServerError)
			return
	}
}

// Helper function to get available domains for logging
func getAvailableDomains(domains map[types.DomainType]types.DomainHandler) []string {
	available := make([]string, 0, len(domains))
	for domain := range domains {
			available = append(available, string(domain))
	}
	return available
}

func (m *LibraryMetrics) recordValidationDuration(duration time.Duration) {
	durationNanos := duration.Nanoseconds()

	// Update total duration
	atomic.AddInt64(&m.ValidationTotalDuration, durationNanos)

	// Update total count
	atomic.AddInt64(&m.ValidationTotalAttempts, 1)

	// Update max duration (atomic compare and swap)
	for {
			current := atomic.LoadInt64(&m.ValidationMaxDuration)
			if durationNanos <= current {
					break
			}
			if atomic.CompareAndSwapInt64(&m.ValidationMaxDuration, current, durationNanos) {
					break
			}
	}

	// Update min duration (atomic compare and swap)
	for {
			current := atomic.LoadInt64(&m.ValidationMinDuration)
			if current == 0 || durationNanos < current {
					if atomic.CompareAndSwapInt64(&m.ValidationMinDuration, current, durationNanos) {
							break
					}
			} else {
					break
			}
	}
}

func (h *LibraryPageHandler) incrementMetric(metric *int64) {
	if h == nil || h.metrics == nil || metric == nil {
    return
	}
	atomic.AddInt64(metric, 1)
}

func (h *LibraryPageHandler) parseQueryParams(r *http.Request, params *LibraryQueryParams) error {
	query := r.URL.Query()

	// Parse page
	if page := query.Get("page"); page != "" {
			val, err := strconv.Atoi(page)
			if err != nil {
					return fmt.Errorf("invalid page parameter: %w", err)
			}
			params.Page = val
	}

	// Parse limit
	if limit := query.Get("limit"); limit != "" {
			val, err := strconv.Atoi(limit)
			if err != nil {
					return fmt.Errorf("invalid limit parameter: %w", err)
			}
			params.Limit = val
	}

	// Parse sort
	if sort := query.Get("sort"); sort != "" {
			params.Sort = sort
	}

	// Validate using existing rules
	if validationErrors := h.baseValidator.ValidateStruct(r.Context(), params); len(validationErrors) > 0 {
			// Return the first validation error as the error message
			return fmt.Errorf("validation failed: %s", validationErrors[0].Error())
	}

	return nil
}