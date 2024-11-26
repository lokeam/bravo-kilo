package library

import (
	"log/slog"
	"net/http"

	"github.com/lokeam/bravo-kilo/internal/shared/services"
)


type LibraryHandler struct {
	libraryService            *LibraryService
	validationService         *services.ValidationService
	logger                    *slog.Logger
}

// RESPONSIBILITIES:

// 1. HTTP Request Setup
func (h *LibraryHandler) HandleGetLibraryPageData(w http.ResponseWriter, r *http.Request) {
    /*
    Responsibilities:
    - Extract request ID (from header or generate new)
    - Basic request logging
    - Call auth middleware for userID
    - Parse query params into LibraryQueryParams
    - Call service layer
    - Handle response writing
    */

	// Setup
	ctx := r.Context()
	requestID := h.getOrGenerateRequestID(r)

	// Auth
	userID, err := h.authenticateRequest(r)
	if err != nil {
		h.respondWithError(w, requestID, err)
		return
	}

	// Validate and parse params
	params, err := h.validationService.ValidateLibraryRequest(ctx, r.URL.Query())
	if err != nil {
		h.respondWithError(w, requestID, err)
		return
	}


	// Get library data
	libraryData, err := h.libraryService.GetLibraryData(ctx, userID, params)
	if err != nil {
		h.respondWithError(w, requestID, err)
		return
	}

	// send response back to frontend
	h.respondWithJSON(w, http.StatusOK, libraryData)
}

// Helpers

func (h *LibraryHandler) respondWithJSON(w http.ResponseWriter, data any) {
	/*
	Responsibilities:
	- Set content type to JSON
	- Write status code
	- Write JSON response
	*/
}

func (h *LibraryHandler) respondWithError(w http.ResponseWriter, err error, status int) {
	/*
	Responsibilities:
	- Log error
	- Set error status
	- Write error response
	*/
}
