package validator

type ValidationErrorCode string

const (
	ErrInvalidFormat     ValidationErrorCode = "INVALID_FORMAT"
	ErrMaxExceeded       ValidationErrorCode = "MAX_EXCEEDED"
	ErrRequired          ValidationErrorCode = "REQUIRED"
	ErrInvalidContent    ValidationErrorCode = "INVALID_CONTENT"
	ErrUnauthorized      ValidationErrorCode = "UNAUTHORIZED"
	ErrNotFound          ValidationErrorCode = "NOT_FOUND"
	ErrDatabaseError     ValidationErrorCode = "DATABASE_ERROR"
	ErrCacheError        ValidationErrorCode = "CACHE_ERROR"
	ErrInvalidISBN       ValidationErrorCode = "INVALID_ISBN"
	ErrInvalidTitle      ValidationErrorCode = "INVALID_TITLE"
	ErrInvalidAuthors    ValidationErrorCode = "INVALID_AUTHORS"
	ErrInvalidGenre      ValidationErrorCode = "INVALID_GENRE"
	ErrInvalidLanguage   ValidationErrorCode = "INVALID_LANGUAGE"
	ErrInvalidPrice      ValidationErrorCode = "INVALID_PRICE"
	ErrInvalidPublisher  ValidationErrorCode = "INVALID_PUBLISHER"
	ErrInvalidYear       ValidationErrorCode = "INVALID_YEAR"

	// Query validation errors
	ErrQueryRequired      ValidationErrorCode = "QUERY_REQUIRED"
	ErrQueryMaxLength     ValidationErrorCode = "QUERY_MAX_LENGTH"
	ErrQueryMinLength     ValidationErrorCode = "QUERY_MIN_LENGTH"
  ErrQueryPattern       ValidationErrorCode = "QUERY_PATTERN"
	ErrQueryInvalidValue  ValidationErrorCode = "QUERY_INVALID_VALUE"
	ErrQueryInvalidFormat ValidationErrorCode = "QUERY_INVALID_FORMAT"
)

type ValidationError struct {
	Field   string                 `json:"field"`
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Value     interface{}          // legacy code. Double check if this is still needed
	Context map[string]interface{} `json:"context,omitempty"`
}

// Constructor
func NewValidationError(field string, code ValidationErrorCode, message string) ValidationError {
	return ValidationError{
		Field:    field,
		Code:     string(code),
		Message:  message,
		Context:  make(map[string]interface{}),
	}
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