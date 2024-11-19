package repository

import (
	"fmt"
	"log/slog"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

// Consts for validation limits
const (
	MaxNestingDepth       = 5      // Max allowed nesting of attributes
	MaxOperations         = 1000   // Max number of operations in a RichText
	MaxOperationLength    = 50000  // Max length of a single operation
	MaxHexColorLength    = 7      // Including #
	MinHexColorLength    = 4      // Including #
	MaxHeaderLevel       = 3
	MinHeaderLevel       = 1
)

const (
	ErrInvalidAttribute = "INVALID_ATTRIBUTE"
	ErrInvalidValue     = "INVALID_VALUE"
	ErrNestedAttribute  = "INVALID_NESTED_ATTR"
	ErrMaxDepthExceeded = "MAX_DEPTH_EXCEEDED"
	ErrMaxOpsExceeded   = "MAX_OPS_EXCEEDED"
	ErrMalformedOp      = "MALFORMED_OPERATION"
	ErrInvalidUTF8      = "INVALID_UTF8"
)

var (
	// Cache common values
	namedColors = map[string]bool{
			"red":     true,
			"green":   true,
			"blue":    true,
			"yellow":  true,
			"purple":  true,
			"orange":  true,
			"pink":    true,
			"brown":   true,
			"black":   true,
			"white":   true,
	}

	// Cache allowed attributes
	allowedAttributes = map[string]bool{
			"bold":       true,
			"italic":     true,
			"underline":  true,
			"strike":     true,
			"header":     true,
			"color":      true,
			"background": true,
			"blockquote": true,
			"list":       true,
			"indent":     true,
	}

	// Precompiled regex for hex color validation
	hexColorRegex = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}){1,2}$`)

	validAttributeValues = map[string]func(interface{}) bool{
		"header":     validateHeader,
		"list":       validateList,
		"indent":     validateIndent,
		"color":      validateColor,
		"background": validateColor,
	}
)

// RichText represents a Quill Delta format text content with operations and attributes.
// Provides methods for validation, sanitization, and content checking.
type RichText struct {
	Ops      []DeltaOp `json:"ops"`
	logger   *slog.Logger
	metrics  *RichTextMetrics
}

type RichTextError struct {
	Code    string
	Message string
	Attr    string
	Value   interface{}
}

// Add after the existing const blocks
type RichTextMetrics struct {
	mu              sync.RWMutex
	Duration        time.Duration
	OperationCount  int
	AttributeCount  map[string]int
	MaxDepthReached int
	ErrorCount      map[string]int
}

func (e *RichTextError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}


// IsRichTextEmpty checks if the RichText content is empty or contains only whitespace.
// Returns true if the RichText is nil, has no operations, or contains only whitespace.
func (rt *RichText) IsRichTextEmpty() bool {
	if rt == nil || len(rt.Ops) == 0 {
			return true
	}

	// Check each operation for content
	for _, op := range rt.Ops {
			switch v := op.Insert.(type) {
			case string:
					if len(strings.TrimSpace(v)) > 0 {
							return false
					}
			default:
					// Non-string insert (like images) means content exists
					return false
			}
	}
	return true
}

// CheckRichTextLength calculates the total length of the RichText content.
// Non-string inserts (like images) count as 1 character.
func (rt *RichText) CheckRichTextLength() int {
	if rt == nil {
			return 0
	}

	length := 0
	for _, op := range rt.Ops {
			switch v := op.Insert.(type) {
			case string:
					length += len(v)
			default:
					// Count non-string inserts (like images) as 1 character
					length += 1
			}
	}
	return length
}

func (rt *RichText) SanitizeContent() {
	if rt == nil {
		return
	}

	// Preallocate slice w/ length
	sanitizedOps := make([]DeltaOp, 0, len(rt.Ops))

	for _, op := range rt.Ops {
		if str, ok := op.Insert.(string); ok {
			// Remove control chars except newline + tab
			cleaned := strings.Map(func(r rune) rune {
				if r < 32 && r != '\n' && r != '\t' {
					return -1
				}
				return r
			}, str)

			if cleaned != "" {
				sanitizedOps = append(sanitizedOps, DeltaOp{
					Insert: cleaned,
					Attributes: op.Attributes,
				})
			}
		} else {
			// Non strings are passed through
			sanitizedOps = append(sanitizedOps, op)
		}
	}

	rt.Ops = sanitizedOps

}

// ValidateAttributes performs comprehensive validation of RichText attributes.
// It checks for allowed attributes, valid values, and proper nesting depth.
// Returns a RichTextError if validation fails.
func (rt *RichText) ValidateAttributes() error {
	startTime := time.Now()
	metrics := &RichTextMetrics{
    mu:             sync.RWMutex{},
    AttributeCount: make(map[string]int, len(allowedAttributes)),
    ErrorCount:     make(map[string]int, len(allowedAttributes)),
    OperationCount: len(rt.Ops),
    MaxDepthReached: 0,
	}

	defer func() {
			metrics.Duration = time.Since(startTime)
			if rt.logger != nil {
					rt.logger.Info("rich text validation completed",
							"duration_ms", metrics.Duration.Milliseconds(),
							"operations", metrics.OperationCount,
							"max_depth", metrics.MaxDepthReached,
							"errors", metrics.ErrorCount,
					)
			}
			rt.metrics = metrics
	}()

	// Guard clause - validate structure
	if err := rt.ValidateStructure(); err != nil {
		return err
	}

	metrics.OperationCount = len(rt.Ops)

	var validateNestedOp func(map[string]interface{}, []string, int) error
	validateNestedOp = func(attrs map[string]interface{}, path []string, depth int) error {
			metrics.updateMaxDepth(depth)

			for attr, value := range attrs {
					attrPath := append(path, attr)
					pathStr := strings.Join(attrPath, ".")

					metrics.incrementAttribute(attr)

					// Check if attribute is allowed
					if !allowedAttributes[attr] {
							err := &RichTextError{
									Code:    ErrInvalidAttribute,
									Message: fmt.Sprintf("invalid attribute '%s' at path '%s' (allowed: %v)",
											attr, pathStr, getAllowedAttributesList()),
									Attr:    attr,
									Value:   value,
							}
							metrics.incrementError(ErrInvalidAttribute)
							logValidationError(rt.logger, err, "invalid attribute")
							return err
					}

					if validator, exists := validAttributeValues[attr]; exists {
							if !validator(value) {
									err := &RichTextError{
											Code:    ErrInvalidValue,
											Message: fmt.Sprintf("invalid value for attribute '%s' at path '%s'", attr, pathStr),
											Attr:    attr,
											Value:   value,
									}
									metrics.incrementError(ErrInvalidValue)
									logValidationError(rt.logger, err, "invalid attribute value")
									return err
							}
					}

					// Check for nested attributes
					if nested, ok := value.(map[string]interface{}); ok {
							if err := validateNestedOp(nested, attrPath, depth+1); err != nil {
									return err
							}
					}
			}
			return nil
	}

	// Primary validation loop
	for i, op := range rt.Ops {
		if op.Attributes != nil {
				if err := validateNestedOp(op.Attributes, []string{fmt.Sprintf("op[%d]", i)}, 0); err != nil {
						return err
				}
		}
}

return nil
}

func (rt *RichText) ValidateStructure() error {
	if rt == nil {
		return &RichTextError{
			Code:     ErrMalformedOp,
			Message:  "RichText cannot be nil",
		}
	}

	// Check for number of operations
	if len(rt.Ops) > MaxOperations {
		return &RichTextError{
			Code:     ErrMaxOpsExceeded,
			Message:  fmt.Sprintf("RichText cannot have more than %d operations", MaxOperations),
			Value:    len(rt.Ops),
		}
	}

	// Validate each operation
    // Validate each operation
    for i, op := range rt.Ops {
			// Check for nil operation
			if op.Insert == nil {
					return &RichTextError{
							Code:    ErrMalformedOp,
							Message: fmt.Sprintf("operation %d has nil insert", i),
					}
			}

			// Validate string content
			if str, ok := op.Insert.(string); ok {
					// Check UTF-8 validity
					if !utf8.ValidString(str) {
							return &RichTextError{
									Code:    ErrInvalidUTF8,
									Message: fmt.Sprintf("operation %d contains invalid UTF-8", i),
							}
					}

					// Check operation length
					if len(str) > MaxOperationLength {
							return &RichTextError{
									Code:    ErrMalformedOp,
									Message: fmt.Sprintf("operation %d exceeds maximum length of %d", i, MaxOperationLength),
									Value:   len(str),
							}
					}
			}

			// Validate attributes nesting depth
			if op.Attributes != nil {
					if err := validateNestingDepth(op.Attributes, 0); err != nil {
							return err
					}
			}
	}

	return nil
}

// Helper functions

// Check the nesting depth of RichTextattributes
func validateNestingDepth(attrs map[string]interface{}, depth int) error {
	if depth > MaxNestingDepth {
		return &RichTextError{
				Code:    ErrMaxDepthExceeded,
				Message: fmt.Sprintf("attribute nesting exceeds maximum depth of %d", MaxNestingDepth),
				Value:   depth,
		}
	}

	for key, value := range attrs {
			if nested, ok := value.(map[string]interface{}); ok {
					if err := validateNestingDepth(nested, depth+1); err != nil {
							return &RichTextError{
									Code:    err.(*RichTextError).Code,
									Message: fmt.Sprintf("in attribute '%s': %s", key, err.(*RichTextError).Message),
									Attr:    key,
									Value:   err.(*RichTextError).Value,
							}
					}
			}
	}

	return nil
}

// getAllowedAttributesList returns a sorted list of allowed attributes
func getAllowedAttributesList() []string {
	attrs := make([]string, 0, len(allowedAttributes))
	for attr := range allowedAttributes {
			attrs = append(attrs, attr)
	}
	sort.Strings(attrs) // Import "sort" if not already imported
	return attrs
}

func isValidHexColor(color string) bool {
	if color == "" {
		return true
	}

	return hexColorRegex.MatchString(color)
}

func validateHeader(v interface{}) bool {
  switch val := v.(type) {
	case float64:
		return val >= MinHeaderLevel && val <= MaxHeaderLevel
	case bool:
		return !val
	default:
		return false
	}
}

func validateList(v interface{}) bool {
	if listValue, ok := v.(string); ok {
		return listValue == "ordered" || listValue == "bullet"
	}

	return false
}

func validateIndent(v interface{}) bool {
	if indentValue, ok := v.(string); ok {
		return indentValue == "+1" || indentValue == "-1"
	}

	return false
}

func validateColor(v interface{}) bool {
	if colorValue, ok := v.(string); ok {
		if colorValue == "" {
			return true
		}
		if strings.HasPrefix(colorValue, "#") {
			return isValidHexColor(colorValue)
		}
		return namedColors[strings.ToLower(colorValue)]
	}

	return false
}

// Add metrics methods
func (m *RichTextMetrics) incrementError(code string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ErrorCount[code]++
}

func (m *RichTextMetrics) incrementAttribute(attr string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.AttributeCount[attr]++
}

func (m *RichTextMetrics) updateMaxDepth(depth int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if depth > m.MaxDepthReached {
			m.MaxDepthReached = depth
	}
}

func logValidationError(logger *slog.Logger, err *RichTextError, msg string) {
	if logger == nil {
			return
	}

	logger.Error(msg,
			"code", err.Code,
			"message", err.Message,
			"attribute", err.Attr,
			"value", err.Value,
	)
}