// Provides validation framework
// Inclues base validation functionality as well as domain-specific validation + standardized error handling
package validator

import (
	"time"
)

type ValidationDomain string

const (
	// Validation domains
	BookDomain ValidationDomain = "book"

	// Query param types
	QueryTypeString  QueryParamType = "string"
	QueryTypeInt     QueryParamType = "int"
	QueryTypeBool    QueryParamType = "bool"
	QueryTypeEnum    QueryParamType = "enum"
	QueryTypeDate    QueryParamType = "date"
	QueryTypeUUID    QueryParamType = "uuid"
	QueryTypeEmail   QueryParamType = "email"
	QueryTypeURL     QueryParamType = "url"
)

// Hold request specific validation data
type ValidationContext struct {
	RequestID     string
	Domain        ValidationDomain
	UserID        int
	Timestamp     time.Time
	TraceID       string
	Timeout       time.Duration
}

// Standardize validator response format
type ValidationResponse struct {
	RequestID        string              `json:"requestId"`
	Success          bool                `json:"success"`
	Errors           []ValidationError   `json:"errors,omitempty"`
	Data             interface{}         `json:"data,omitempty"`
	Timestamp        time.Time           `json:"timestamp"`
}

type QueryValidationRule struct {
	Required       bool
	MaxLength      int
	MinLength      int
	AllowedValues  []string
	Type           QueryParamType
	Pattern        string // Regex pattern for validation
}

type QueryValidationRules map[string]QueryValidationRule
type QueryParamType string