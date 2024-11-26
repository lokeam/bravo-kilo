package validator

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"sync"
	"time"

	"net/url"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/lokeam/bravo-kilo/internal/shared/core"
	"github.com/lokeam/bravo-kilo/internal/shared/types"
)

type ValidationDomain = core.DomainType
type QueryParamType = types.QueryParamType
type ValidationContext = types.ValidationContext
type ValidationError = types.ValidationError
type ValidationResponse = types.ValidationResponse
type QueryValidationRules = types.QueryValidationRules

var _ types.Validator = (*BaseValidator)(nil)

const (
	DefaultMaxQueryParamLength = 100
	MaxPatternCompileTimeout   = 100 * time.Millisecond
	MaxPatternCacheSize = 1000
)


type BaseValidator struct {
	validate    *validator.Validate
	logger      *slog.Logger
	domain      ValidationDomain
	logFields   map[string]interface{}
	patterns    map[string]*regexp.Regexp
	patternsMu  sync.RWMutex
	commonPatterns struct {
		email *regexp.Regexp
		url   *regexp.Regexp
	}
}

// Constructor
func NewBaseValidator(logger *slog.Logger, domain ValidationDomain) (*BaseValidator, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	bv := &BaseValidator{
		validate:   validator.New(),
		logger:     logger,
		domain:     domain,
		logFields:  make(map[string]interface{}),
		patterns:   make(map[string]*regexp.Regexp),
	}

	// Pre compile common patterns
	bv.commonPatterns.email = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if bv.commonPatterns.email == nil {
			return nil, fmt.Errorf("failed to compile email regex")
	}

	bv.commonPatterns.url = regexp.MustCompile(`^https?:\/\/(?:www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b(?:[-a-zA-Z0-9()@:%_\+.~#?&\/=]*)$`)
	if bv.commonPatterns.url == nil {
			return nil, fmt.Errorf("failed to compile URL regex")
	}

	return bv, nil
}

// Create new validation context
func (bv *BaseValidator) CreateValidationContext(requestID string, userID int) *ValidationContext {
	return &ValidationContext{
			RequestID:  requestID,
			Domain:     bv.domain,
			UserID:     userID,
			Timestamp:  time.Now(),
			TraceID:    uuid.New().String(), // Requires github.com/google/uuid
	}
}

// Extract validationContext from context
func (bv *BaseValidator) GetValidationContext(ctx context.Context) (*ValidationContext, error) {
	valCtx, ok := ctx.Value(ValidationContextKey).(*ValidationContext)
	if !ok || valCtx == nil {
		return nil, fmt.Errorf("validation context not found or invalid")
	}

	return valCtx, nil
}

// Add validation context
func (bv *BaseValidator) WithContext(ctx context.Context, valCtx *ValidationContext) context.Context {
	return context.WithValue(ctx, ValidationContextKey, valCtx)
}

// Add error context

// Perform base validation on any struct
func (bv *BaseValidator) ValidateStruct(ctx context.Context, data interface{}) []ValidationError {
	if err := bv.validate.Struct(data); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			return bv.formatValidationErrors(validationErrors)
		}

		// Log unexpected error
		bv.logger.Error("unexpected validation error", "error", err)
		return []ValidationError{{
			Message: "an unexpected validation error occurred",
		}}
	}

	return nil
}

// Perform validation on a single field
func (bv *BaseValidator) ValidateField(field, value string) error {
	// Check against known domain types for domain validation
	if field == "domain" {
		switch value {
		case string(core.BookDomainType), string(core.GameDomainType):
			return nil
		default:
			return fmt.Errorf("invalid domain type")
		}
	}

	// Todo: add more field validations
	return fmt.Errorf("field validation failed, unknown field: %s", field)
}

// Creates successful response
func (bv *BaseValidator) CreateSuccessResponse(ctx context.Context, data interface{}) ValidationResponse {
	valCtx, ok := ctx.Value(ValidationContextKey).(*ValidationContext)
	if !ok || valCtx == nil {
		bv.logger.Error("validation context not found or invalid")

		// Send response with generated request ID
		return ValidationResponse{
			RequestID: uuid.New().String(),
			Success:   true,
			Data:     data,
			Timestamp: time.Now(),
		}
	}

	return ValidationResponse{
		RequestID:   valCtx.RequestID,
		Success:     true,
		Data:        data,
		Timestamp:   time.Now(),
	}
}

// Create error response for MULTIPLE validation errors
func (bv *BaseValidator) BuildValidationErrorResponse(ctx context.Context, errors []ValidationError) ValidationResponse{
	var requestID string
	if valCtx, ok := ctx.Value(ValidationContextKey).(*ValidationContext); ok && valCtx != nil {
			requestID = valCtx.RequestID
	} else {
			requestID = uuid.New().String()
			bv.logger.Warn("validation context not found, using generated ID")
	}

	return ValidationResponse{
		RequestID: requestID,
		Success:   false,
		Errors:    errors,
		Timestamp: time.Now(),
}
}

// Create an error response for a SINGLE validation error
func (bv *BaseValidator) BuildSingleErrorResponse(ctx context.Context, err ValidationError) ValidationResponse {
	return bv.BuildValidationErrorResponse(ctx, []ValidationError{err})
}

// Validates URL query parameters against define rules
func (bv *BaseValidator) ValidateQueryParams(ctx context.Context, query url.Values, rules QueryValidationRules) []ValidationError {

	var errors []ValidationError

	for param, rule := range rules {
			value := query.Get(param.String())

			// Required check
			if rule.Required && value == "" {
					errors = append(errors, NewValidationError(
							param.String(),
							ErrQueryRequired,
							fmt.Sprintf("parameter '%s' is required", param),
					))
					continue
			}

			// Skip empty optional params
			if value == "" {
					continue
			}

			// Length check
			if err := bv.validateQueryParamLength(param, value, rule.MinLength, rule.MaxLength); err != nil {
					errors = append(errors, *err)
			}

			// Type validation
			if err := bv.validateQueryParamType(param, value, rule.Type); err != nil {
					errors = append(errors, *err)
			}

			// Allowed values check
			if err := bv.validateQueryAllowedValues(param, value, rule.AllowedValues); err != nil {
					errors = append(errors, *err)
			}

			// Regex pattern check
			if rule.Pattern != "" {
				if err := bv.validateRegexQueryPattern(param, value, rule.Pattern); err != nil {
					errors = append(errors, *err)
				}
			}
	}

	return errors
}

// Returns common validation rules for domain
func (bv *BaseValidator) GetDefaultQueryRules() QueryValidationRules {
	return QueryValidationRules{
		"domain": {
			Required: true,
			MaxLength: 50,
			AllowedValues: []string{
				string(core.BookDomainType),
				string(core.GameDomainType),
			},
		},
		"email":
		{
			Required: true,
			Type: types.QueryTypeEmail,
			MaxLength: 255,
		},
		"date": {
			Required: false,
			Type: types.QueryTypeDate,
		},
		"id": {
			Required: true,
			Type: types.QueryTypeUUID,
		},
		"username": {
			Required: true,
			MinLength: 3,
			MaxLength: 50,
			Pattern:   `^[a-zA-Z0-9_-]+$`,
		},
	}
}

// Tidy up
func (bv *BaseValidator) Cleanup() {
	bv.logFields = make(map[string]interface{})
	bv.patternsMu.Lock()
	defer bv.patternsMu.Unlock()

	// Clear pattern cache
	for k := range bv.patterns {
		delete(bv.patterns, k)
	}

	// Clear existing fields
	bv.logFields = make(map[string]interface{})
}



// Helper fns:
// Convert validator errors to custom format
func (bv *BaseValidator) formatValidationErrors(errors validator.ValidationErrors) []ValidationError {
	var validationErrors []ValidationError

	for _, err := range errors {
		validationErrors = append(validationErrors, ValidationError{
			Field:    err.Field(),
			Code:     err.Tag(),
			Message:  fmt.Sprintf("validation failed on '%s' with tag '%s'", err.Field(), err.Tag()),
			Context: map[string]interface{}{
				"value": err.Value(),
				"param": err.Param(),
			},
		})
	}

	return validationErrors
}

// Query param validator helpers
func (bv *BaseValidator) validateQueryParamLength(param types.ValidationRuleKey, value string, minLen,maxLen int) *ValidationError {
	paramStr := param.String()
	if maxLen == 0 {
			maxLen = DefaultMaxQueryParamLength
	}
	if minLen > 0 && len(value) < minLen {
		verr := NewValidationError(
			paramStr,
			ErrQueryMinLength,
			fmt.Sprintf("parameter '%s' is shorter than the minimum length of %d", param, minLen),
		)
		verrPtr := &verr
		verrPtr.WithContext("min_length", minLen)
		verrPtr.WithContext("actual_length", len(value))
		return verrPtr
	}
	if len(value) > maxLen {
			verr := NewValidationError(
				paramStr,
					ErrQueryMaxLength,
					fmt.Sprintf("parameter '%s' exceeds maximum length of %d", param, maxLen),
			)
			// Create a pointer to the error and chain context
			verrPtr := &verr
			verrPtr.WithContext("max_length", maxLen)
			verrPtr.WithContext("actual_length", len(value))
			return verrPtr
	}

	return nil
}

func (bv *BaseValidator) validateQueryParamType(param types.ValidationRuleKey, value string, paramType QueryParamType) *ValidationError {
	paramStr := param.String()
	switch paramType {
	case types.QueryTypeInt:
			if _, err := strconv.Atoi(value); err != nil {
					verr := NewValidationError(
						paramStr,
							ErrQueryInvalidFormat,
							fmt.Sprintf("parameter '%s' must be an integer", param),
					)
					verrPtr := &verr
					verrPtr.WithContext("actual_value", value)
					return verrPtr
			}
	case types.QueryTypeBool:
			if value != "true" && value != "false" {
					verr := NewValidationError(
						paramStr,
							ErrQueryInvalidFormat,
							fmt.Sprintf("parameter '%s' must be 'true' or 'false'", param),
					)
					verrPtr := &verr
					verrPtr.WithContext("actual_value", value)
					return verrPtr
			}
	case types.QueryTypeDate:
		if _, err := time.Parse("2006-01-02", value); err != nil {
			verr := NewValidationError(
				paramStr,
					ErrQueryInvalidFormat,
					fmt.Sprintf("parameter '%s' must be a valid date in the format YYYY-MM-DD", param),
			)
			verrPtr := &verr
			verrPtr.WithContext("actual_value", value)
			return verrPtr
		}
	case types.QueryTypeUUID:
		if _, err := uuid.Parse(value); err != nil {
			verr := NewValidationError(
				paramStr,
				ErrQueryInvalidFormat,
				fmt.Sprintf("parameter '%s' must be a valid UUID", param),
			)
			verrPtr := &verr
			verrPtr.WithContext("actual_value", value)
			return verrPtr
		}
	case types.QueryTypeEmail:
		if !bv.commonPatterns.email.MatchString(value) {
			verr := NewValidationError(
				paramStr,
				ErrQueryInvalidFormat,
				fmt.Sprintf("parameter '%s' must be a valid email address", param),
			)
			verrPtr := &verr
			verrPtr.WithContext("actual_value", value)
			return verrPtr
		}
	case types.QueryTypeURL:
		if _, err := url.Parse(value); err != nil {
			verr := NewValidationError(
				paramStr,
				ErrQueryInvalidFormat,
				fmt.Sprintf("parameter '%s' must be a valid URL", param),
			)
			verrPtr := &verr
			verrPtr.WithContext("actual_value", value)
			return verrPtr
		}
	}
	return nil
}

func (bv *BaseValidator) validateQueryAllowedValues(param types.ValidationRuleKey, value string, allowedValues []string) *ValidationError {
	paramStr := param.String()
	if len(allowedValues) == 0 {
			return nil
	}

	for _, allowed := range allowedValues {
			if value == allowed {
					return nil
			}
	}

	verr := NewValidationError(
		paramStr,
			ErrQueryInvalidValue,
			fmt.Sprintf("parameter '%s' must be one of: %v", param, allowedValues),
	)
	verrPtr := &verr
	verrPtr.WithContext("allowed_values", allowedValues)
	verrPtr.WithContext("actual_value", value)
	return verrPtr
}

func (bv *BaseValidator) validateRegexQueryPattern(param types.ValidationRuleKey, value, pattern string) *ValidationError {
	paramStr := param.String()
	if pattern == "" {
		return nil
	}

	// Get validation context from parent if available
	var ctx context.Context
	var cancel context.CancelFunc

	// Set timeout for pattern compilation
	timeout := MaxPatternCompileTimeout
	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Get compiled pattern with timeout protection
	regex, err := bv.getCompiledPattern(ctx, pattern)
	if err != nil {
			bv.logger.Error("pattern validation failed",
					"param", param,
					"pattern", pattern,
					"error", err,
			)
			verr := NewValidationError(
				paramStr,
					ErrQueryPattern,
					"invalid pattern configuration",
			)
			return &verr
	}

	if !regex.MatchString(value) {
			verr := NewValidationError(
				paramStr,
					ErrQueryInvalidFormat,
					fmt.Sprintf("parameter '%s' does not match required pattern", param),
			)
			verrPtr := &verr
			verrPtr.WithContext("actual_value", value)
			return verrPtr
	}

	return nil
}


// Make pattern compilation thread safe
func (bv *BaseValidator) getCompiledPattern(ctx context.Context, pattern string) (*regexp.Regexp, error) {
	// Fast path - check if pattern is already compiled
	bv.patternsMu.RLock()
	if regex, exists := bv.patterns[pattern]; exists {
			bv.patternsMu.RUnlock()
			return regex, nil
	}
	bv.patternsMu.RUnlock()

	// Check cache size before compilation
	bv.patternsMu.RLock()
	if len(bv.patterns) >= MaxPatternCacheSize {
			bv.patternsMu.RUnlock()
			bv.logger.Warn("pattern cache full, compiling without caching",
					"pattern", pattern,
					"cache_size", len(bv.patterns),
			)
			// Compile without caching if cache is full
			return regexp.Compile(pattern)
	}
	bv.patternsMu.RUnlock()

	// Slow path - compile + cache pattern
	bv.patternsMu.Lock()
	defer bv.patternsMu.Unlock()

	// Double-check after acquiring write lock
	if regex, exists := bv.patterns[pattern]; exists {
			return regex, nil
	}

	// Compile pattern with timeout protection
	done := make(chan struct{})
	var regex *regexp.Regexp
	var compileErr error

	go func() {
			regex, compileErr = regexp.Compile(pattern)
			close(done)
	}()

	select {
	case <-done:
			if compileErr != nil {
					return nil, fmt.Errorf("pattern compilation failed: %w", compileErr)
			}
			// Cache successful compilation
			bv.patterns[pattern] = regex
			return regex, nil
	case <-ctx.Done():
			return nil, fmt.Errorf("pattern compilation timed out after %v", MaxPatternCompileTimeout)
	}
}

