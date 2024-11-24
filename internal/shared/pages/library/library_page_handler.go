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
	"github.com/lokeam/bravo-kilo/internal/shared/pages/library/domains"
	"github.com/lokeam/bravo-kilo/internal/shared/processor/bookprocessor"
	"github.com/lokeam/bravo-kilo/internal/shared/redis"
	"github.com/lokeam/bravo-kilo/internal/shared/types"
	"github.com/lokeam/bravo-kilo/internal/shared/validator"
	goredis "github.com/redis/go-redis/v9"
)

const (
	BooksDomain types.DomainType = "books"
	GamesDomain types.DomainType = "games"

	PageRuleKey  validator.ValidationRuleKey = "page"
	LimitRuleKey validator.ValidationRuleKey = "limit"
	SortRuleKey  validator.ValidationRuleKey = "sort"

	defaultContextTimeout = 30 * time.Second  // Increased from 10s

	// Operation budgets (percentages of total timeout)
	cacheOperationBudget    = 0.2  // 20% of total time
	dbOperationBudget      = 0.3  // 30% of total time
	processingBudget       = 0.3  // 30% of total time
	organizingBudget       = 0.2  // 20% of total time
)

type LibraryPageHandler struct {
	mu                 sync.RWMutex
	domainHandlers     map[types.DomainType]types.DomainHandler
	redisClient        *redis.RedisClient
	logger             *slog.Logger
	cacheManager       *cache.CacheManager
	metrics            *LibraryMetrics
	organizer          *organizer.BookOrganizer
	bookProcessor      *bookprocessor.BookProcessor

	// Validator
	baseValidator      *validator.BaseValidator
	validationRules    validator.QueryValidationRules
}

type LibraryResponse struct {
	RequestID  string                  `json:"requestId"`
	Data       *types.LibraryPageData   `json:"data"`
	Source     string                  `json:"source"` // either "cache" || "database"
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

// ADD: New helper type for operation tracking
type operationDetails struct {
	name      string
	budget    float64
	startTime time.Time
}

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

	// Init organizer + processor
	bookProcessor, err := bookprocessor.NewBookProcessor(logger)
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
		PageRuleKey: {
				Required:  true,
				Type:     validator.QueryTypeInt,
				MinLength: 1,
				MaxLength: 5,
		},
		LimitRuleKey: {
				Required:  true,
				Type:     validator.QueryTypeInt,
				MinLength: 1,
				MaxLength: 3,
		},
		SortRuleKey: {
				Required:      false,
				Type:         validator.QueryTypeString,
				AllowedValues: []string{"asc", "desc"},
		},
	}

	bookDomainHandler := domains.NewBookDomainHandler(bookHandlers, logger)
	return &LibraryPageHandler{
		domainHandlers: map[types.DomainType]types.DomainHandler{
			BooksDomain: bookDomainHandler,
		},
		logger:          logger,
		redisClient:     redisClient,
		organizer:       organizer,
		bookProcessor:   bookProcessor,
		cacheManager:    cacheManager,
		metrics:         &LibraryMetrics{},
		baseValidator:   baseValidator,
		validationRules: validationRules,
	}, nil
}

func (h *LibraryPageHandler) HandleGetLibraryPageData(w http.ResponseWriter, r *http.Request) {
	// 1. Initial Setup
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
			requestID = uuid.New().String()
	}
	h.logger.Info("Starting library page request",
	"method", r.Method,
	"path", r.URL.Path)

	// Start tracking total request duration
	start := time.Now()
	defer h.trackOperationDuration("total_request", start)

	// Parent context with overall timeout
	ctx, cancel := context.WithTimeout(r.Context(), defaultContextTimeout)
	ctx = context.WithValue(ctx, "requestID", requestID)
	defer cancel()

	// Request Tracking
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

	// Pre-auth context check
	preAuthOp := h.beginOperation("pre-authentication", 0.1) // Small budget for auth checks
	if err := h.checkContext(ctx, requestID, preAuthOp); err != nil {
			h.respondWithError(w, requestID, err, http.StatusGatewayTimeout)
			return
	}

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

	// Post-auth context check
	postAuthOp := h.beginOperation("post-authentication", 0.1)
	if err := h.checkContext(ctx, requestID, postAuthOp); err != nil {
			h.respondWithError(w, requestID, err, http.StatusGatewayTimeout)
			return
	}

	h.logger.Debug("user authenticated",
			"requestID", requestID,
			"userID", userID,
	)

	h.logger.Info("starting library page request",
			"requestID", requestID,
			"method", r.Method,
			"path", r.URL.Path,
	)

	// 2. Pre-validation context check with budget
	preValidationOp := h.beginOperation("pre-validation", 0.05) // 5% budget
	if err := h.checkContext(ctx, requestID, preValidationOp); err != nil {
			h.respondWithError(w, requestID, err, http.StatusGatewayTimeout)
			return
	}

	// 3. Parse and validate query parameters
	var params LibraryQueryParams
	err = h.withTimeout(ctx, "query_validation", 0.05, func(opCtx context.Context) error {
			return h.parseQueryParams(r, &params)
	})
	if err != nil {
			h.logger.Error("invalid query parameters",
					"requestID", requestID,
					"error", err,
			)
			h.respondWithError(w, requestID, err, http.StatusBadRequest)
			return
	}

	// 4. Post-validation context check
	postValidationOp := h.beginOperation("post-validation", 0.05)
	if err := h.checkContext(ctx, requestID, postValidationOp); err != nil {
			h.respondWithError(w, requestID, err, http.StatusGatewayTimeout)
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

	// 6. Pre-domain validation context check
	preDomainValidationOp := h.beginOperation("pre-domain-validation", 0.05)
	if err := h.checkContext(ctx, requestID, preDomainValidationOp); err != nil {
			h.respondWithError(w, requestID, err, http.StatusGatewayTimeout)
			return
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

	// 7. Validate domain handler exists
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

	// Post-domain validation context check
	postDomainValidationOp := h.beginOperation("post-domain-validation", 0.05)
	if err := h.checkContext(ctx, requestID, postDomainValidationOp); err != nil {
			h.respondWithError(w, requestID, err, http.StatusGatewayTimeout)
			return
	}

	// 8. Pre-cache context check
	preCacheOp := h.beginOperation("pre-cache", 0.05)
	if err := h.checkContext(ctx, requestID, preCacheOp); err != nil {
			h.respondWithError(w, requestID, err, http.StatusGatewayTimeout)
			return
	}

	// 9. Redis health monitoring
	if !h.redisClient.IsReady() {
		h.logger.Warn("Redis client not ready",
				"requestID", requestID,
				"status", h.redisClient.GetStatus())

		// Get circuit breaker state if available
		if breaker := h.redisClient.GetCircuitBreaker(); breaker != nil {
				h.logger.Warn("Circuit breaker status",
						"requestID", requestID,
						"state", breaker.GetState())
		}
	}

	// 10. Cache operations
	cacheKeyCtx, cacheKeyCancel := context.WithTimeout(ctx, h.cacheTimeout)
	defer cacheKeyCancel()

	cacheKey, err := h.generateCacheKey(cacheKeyCtx, userID, domainType)
	if err != nil {
			if err == context.DeadlineExceeded {
					h.logger.Error("cache key generation timed out",
							"requestID", requestID,
							"error", err,
					)
			} else {
					h.logger.Error("cache key generation failed",
							"requestID", requestID,
							"error", err,
					)
			}
	}

	// 11. Post-cache context check
	postCacheOp := h.beginOperation("post-cache", 0.05)
	if err := h.checkContext(ctx, requestID, postCacheOp); err != nil {
			h.respondWithError(w, requestID, err, http.StatusGatewayTimeout)
			return
	}

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

	// 13. Pre-database operation context check
	preDBOp := h.beginOperation("pre-database", 0.05)
	if err := h.checkContext(ctx, requestID, preDBOp); err != nil {
			h.respondWithError(w, requestID, err, http.StatusGatewayTimeout)
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

	// 14. Database operations with timeout budget
	var libraryData *types.LibraryPageData
	err = h.withTimeout(ctx, "database_operations", dbOperationBudget, func(opCtx context.Context) error {
			data, err := domainHandler.GetLibraryItems(opCtx, userID, params)
			if err != nil {
					h.logger.Error("failed to get library data from database",
							"requestID", requestID,
							"domain", domainType,
							"error", err,
					)
					return fmt.Errorf("database operation failed: %w", err)
			}
			libraryData = data
			return nil
	})
	postDBOp := h.beginOperation("post-database", 0.05)
	if err := h.checkContext(ctx, requestID, postDBOp); err != nil {
			h.respondWithError(w, requestID, err, http.StatusGatewayTimeout)
			return
	}


	// Database operations with specific timeout
	dbCtx, dbCancel := context.WithTimeout(ctx, h.dbTimeout)
	defer dbCancel()

	items, err := domainHandler.GetLibraryItems(dbCtx, userID)
	if err != nil {
		h.logger.Error("failed to get library items",
			"requestID", requestID,
			"error", err,
		)
		h.respondWithError(w, requestID, err, http.StatusInternalServerError)
		return
	}


	h.logger.Debug("Retrieved domain items",
	"requestID", requestID,
	"itemCount", len(items),
	"domain", domainType,
	"hasNilItems", items == nil)

	if err := h.checkContext(ctx, requestID, "pre-processing"); err != nil {
		h.respondWithError(w, requestID, err, http.StatusGatewayTimeout)
		return
	}

	processCtx, processCancel := context.WithTimeout(ctx, h.processTimeout)
	defer processCancel()

	processedData, err := h.bookProcessor.ProcessLibraryItems(processCtx, items)
	if err != nil {
		h.logger.Error("processing failed",
		"requestID", requestID,
		"error", err,
	)
		h.respondWithError(w, requestID, err, http.StatusInternalServerError)
		return
	}

	// Check context after processing
	if err := h.checkContext(ctx, requestID, "post-processing"); err != nil {
		h.respondWithError(w, requestID, err, http.StatusGatewayTimeout)
		return
	}

	h.logger.Debug("Processing completed",
	"requestID", requestID,
	"hasNilProcessedData", processedData == nil)

	// Pre-organization context check
	if err := h.checkContext(ctx, requestID, "pre-organization"); err != nil {
		h.respondWithError(w, requestID, err, http.StatusGatewayTimeout)
		return
	}

	organizeCtx, organizeCancel := context.WithTimeout(ctx, h.organizeTimeout)
	defer organizeCancel()

	organizedData, err := h.organizer.OrganizeForLibrary(organizeCtx, processedData)
	if err != nil {
		h.logger.Error("organizing failed",
			"requestID", requestID,
			"error", err,
		)
		h.respondWithError(w, requestID, err, http.StatusInternalServerError)
		return
	}

	// Post-organization context check
	if err := h.checkContext(ctx, requestID, "post-organization"); err != nil {
		h.respondWithError(w, requestID, err, http.StatusGatewayTimeout)
		return
	}

	h.logger.Debug("Organization completed",
	"requestID", requestID,
	"organizedDataSize", len(organizedData.Books),
	"authorCount", len(organizedData.BooksByAuthors.AllAuthors),
	"authorMappingSize", len(organizedData.BooksByAuthors.ByAuthor),
	"genreCount", len(organizedData.BooksByGenres.AllGenres),
	"genreMappingSize", len(organizedData.BooksByGenres.ByGenre),
	"formatAudioBookCount", len(organizedData.BooksByFormat.AudioBook),
	"formatEBookCount", len(organizedData.BooksByFormat.EBook),
	"formatPhysicalCount", len(organizedData.BooksByFormat.Physical),
	"tagCount", len(organizedData.BooksByTags.AllTags),
	"tagMappingSize", len(organizedData.BooksByTags.ByTag))


	pageData := LibraryResponse{
		RequestID: requestID,
		Data:      organizedData,
		Source:    "database",
	}

	// 12. Cache response
	// todo: replace constant ttl with proper cacheManager config
	cacheStart = time.Now()
	h.logger.Debug("Preparing to cache response",
        "requestID", requestID,
        "cacheKey", cacheKey,
        "ttl", h.redisClient.GetConfig().CacheConfig.DefaultTTL,
        "dataSize", len(organizedData.Books),
        "hasNilData", organizedData == nil)

	// Add context with timeout for cache operation
	cacheCtx, cancelCache := context.WithTimeout(ctx, h.cacheTimeout)
	defer cancelCache()

	err = h.redisClient.Set(cacheCtx, cacheKey, organizedData, h.redisClient.GetConfig().CacheConfig.DefaultTTL)
	cacheDuration = time.Since(cacheStart)
	atomic.AddInt64(&h.metrics.CacheOperationDuration, cacheDuration.Nanoseconds())

	if err != nil {
		if err == context.DeadlineExceeded {
				h.logger.Error("cache operation timed out",
						"requestID", requestID,
						"duration", cacheDuration,
						"cacheKey", cacheKey)
		} else {
				h.logger.Error("failed to cache response",
						"requestID", requestID,
						"error", err,
						"cacheKey", cacheKey,
						"duration", cacheDuration)
		}
		atomic.AddInt64(&h.metrics.Errors, 1)
	} else {
			h.logger.Debug("Successfully cached response",
					"requestID", requestID,
					"duration", cacheDuration,
					"cacheKey", cacheKey)
	}


	// Final context check before response
	if err := h.checkContext(ctx, requestID, "pre-response"); err != nil {
		h.respondWithError(w, requestID, err, http.StatusGatewayTimeout)
		return
	}

	// 13. Return response
	h.logger.Info("Sending response",
        "requestID", requestID,
        "source", pageData.Source,
        "dataSize", len(pageData.Data.Books),
        "responseTime", time.Since(requestStart))
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

func (h *LibraryPageHandler) generateCacheKey (ctx context.Context, userID int, domainType types.DomainType) (string, error) {
	start := time.Now()
	defer h.trackOperationDuration("cache_key_gen", start)

	// Use parent context directly - no need for separate timeout here
	if err := ctx.Err(); err != nil {
			h.logger.Error("context error in generateCacheKey",
					"duration", time.Since(start),
					"error", err,
			)
			return "", err
	}

	key := fmt.Sprintf("library:%d:%s", userID, domainType)

	h.logger.Debug("cache key generated",
			"duration", time.Since(start),
			"key", key,
	)
	return key, nil
}

func (h *LibraryPageHandler) checkContext(ctx context.Context, requestID string, op operationDetails) error {
	if err := ctx.Err(); err != nil {
			h.logger.Error("context cancelled",
					"requestID", requestID,
					"operation", op.name,
					"duration", time.Since(op.startTime),
					"error", err,
			)
			return err
	}
	return nil
}

// Budgeted context timeout helpers
func (h *LibraryPageHandler) beginOperation(name string, budget float64) operationDetails {
	return operationDetails{
			name:      name,
			budget:    budget,
			startTime: time.Now(),
	}
}

func (h *LibraryPageHandler) trackOperationDuration(operation string, start time.Time) {
	duration := time.Since(start)
	h.logger.Debug("operation duration",
			"operation", operation,
			"duration", duration,
	)
}

func (h *LibraryPageHandler) withTimeout(
	parentCtx context.Context,
	operation string,
	budget float64,
	fn func(context.Context) error,
) error {
	timeout := time.Duration(float64(defaultContextTimeout) * budget)
	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()

	start := time.Now()
	defer h.trackOperationDuration(operation, start)

	op := h.beginOperation(operation, budget)

	// Execute operation with timeout context
	err := fn(ctx)
	if err != nil {
			h.logger.Error("operation failed",
					"operation", op.name,
					"duration", time.Since(op.startTime),
					"error", err,
			)
	}

	return err
}
