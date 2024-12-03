package library

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/lokeam/bravo-kilo/internal/shared/core"
	"github.com/lokeam/bravo-kilo/internal/shared/httputils"
	"github.com/lokeam/bravo-kilo/internal/shared/operations"
	"github.com/lokeam/bravo-kilo/internal/shared/services"
)

type ErrorResponse struct {
	Error     string `json:"error"`
	RequestID string `json:"requestId"`
}

type LibraryHandler struct {
	authExecutor              *operations.OperationExecutor[int]
	libraryService            *LibraryService
	validationService         *services.ValidationService
	logger                    *slog.Logger
}

func NewLibraryHandler(
	libraryService *LibraryService,
	validationService *services.ValidationService,
	logger *slog.Logger,
) *LibraryHandler {
	if libraryService == nil {
		panic("library service is required")
	}
	if validationService == nil {
		panic("validation service is required")
	}
	if logger == nil {
		panic("logger is required")
	}
	return &LibraryHandler{
		authExecutor:        operations.NewOperationExecutor[int]("auth_request", 5*time.Second, logger),
		libraryService:      libraryService,
		validationService:   validationService,
		logger:              logger,
	}
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
	requestID, ok := ctx.Value(core.RequestIDKey).(string)
	if !ok {
		h.logger.Error("request ID not found in context")
		httputils.RespondWithError(w, h.logger, "", fmt.Errorf("internal server error"))
		return
	}

	// Auth
	userID, err := httputils.AuthenticateRequest(ctx, r, h.authExecutor)
	if err != nil {
		httputils.RespondWithError(w, h.logger, requestID, err)
		return
	}

	// Validate and parse params
	params, err := h.validationService.ValidatePageRequest(ctx, r.URL.Query())
	if err != nil {
		httputils.RespondWithError(w, h.logger, requestID, err)
		return
	}


	// Get library data
	libraryData, err := h.libraryService.GetLibraryData(ctx, userID, params)
	if err != nil {
		httputils.RespondWithError(w, h.logger, requestID, err)
		return
	}

	// Send response back to frontend
	if err := httputils.RespondWithJSON(w, h.logger, http.StatusOK, libraryData); err != nil {
		h.logger.Error("failed to send response",
			"error", err,
		)
	}

	return
}
