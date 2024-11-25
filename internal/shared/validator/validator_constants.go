package validator

type ContextKey string

const (
	// Context keys
	ValidationContextKey ContextKey = "validation_context"

	// Common validation limits
	MaxStringLength      = 500
	MaxArrayLength       = 100
	MaxFileSize          = 10 * 1024 * 1024 // 10MB

	// Common validation tags
	TagRequired    = "required"
	TagMin        = "min"
	TagMax        = "max"
	TagEmail      = "email"
	TagURL        = "url"
	TagISBN       = "isbn"

	// Common validation patterns
	EmailPattern         = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	URLPattern      = `^https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{2,256}\.[a-z]{2,6}\b([-a-zA-Z0-9@:%_\+.~#?&//=]*)$`

	// Query parameter validation rules
	MinPage = 1
	MaxPage = 99999
	MinLimit = 1
	MaxLimit = 100
)