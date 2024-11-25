package types

import "time"

type ValidationDomain string
type ValidationRuleKey string
type QueryParamType string

type ValidationContext struct {
	RequestID     string
	Domain        ValidationDomain
	UserID        int
	Timestamp     time.Time
	TraceID       string
	Timeout       time.Duration
}

type ValidationError struct {
	Field   string                 `json:"field"`
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Value     interface{}          // legacy code. Double check if this is still needed
	Context map[string]interface{} `json:"context,omitempty"`
}

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

type QueryValidationRules map[ValidationRuleKey]QueryValidationRule


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

func (v ValidationRuleKey) String() string {
	return string(v)
}

func (ve *ValidationError) WithContext(key string, value interface{}) *ValidationError {
	if ve.Context == nil {
		ve.Context = make(map[string]interface{})
	}

	ve.Context[key] = value
	return ve
}

func IsValidationError(err error) bool {
	_, ok := err.(ValidationError)
	return ok
}

func (ve ValidationError) Error() string {
	return ve.Message
}