package types


type ValidationConfig struct {
	MaxBookTitleLength     int       `json:"maxBookTitleLength"`
	MinTitleLength         int       `json:"minBookTitleLength"`
	MaxCategoryLength      int       `json:"maxBooksPerCategory"`
	CategoryPattern        string    `json:"categoryPattern"`
}

func DefaultValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		MaxBookTitleLength:     255,
		MinTitleLength:         1,
		MaxCategoryLength:      1000,
		CategoryPattern:        `^[a-zA-Z0-9\s\-_]+$`,
	}
}