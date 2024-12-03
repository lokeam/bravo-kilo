package core

import "errors"

var (
	ErrValidation     = errors.New("validation error")
	ErrAuthentication = errors.New("authentication error")
)

type ErrorResponse struct {
	Error     string `json:"error"`
	RequestID string `json:"requestId"`
}