package home

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

type HomeHandler struct {
	authExecutor *operations.OperationExecutor[int]
	homeService *HomeService
	validationService *services.ValidationService
	logger *slog.Logger
}

func NewHomeHandler(
	logger *slog.Logger,
	validationService *services.ValidationService,
	homeService *HomeService,
) *HomeHandler {
	if logger == nil {
		panic("logger is required")
	}
	if validationService == nil {
		panic("validation service is required")
	}
	if homeService == nil {
		panic("home service is required")
	}

	return &HomeHandler{
		authExecutor: operations.NewOperationExecutor[int]("auth_request", 5*time.Second, logger),
		homeService: homeService,
		validationService: validationService,
		logger: logger,
	}
}

func (h *HomeHandler) HandleGetHomePageData(w http.ResponseWriter, r *http.Request) {
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

	// Validate / parse params
	params, err := h.validationService.ValidatePageRequest(ctx, r.URL.Query())
	if err != nil {
		httputils.RespondWithError(w, h.logger, requestID, err)
		return
	}

	// Get home page data
	homeData, err := h.homeService.GetHomeData(ctx, userID, params)
	if err != nil {
		httputils.RespondWithError(w, h.logger, requestID, err)
		return
	}

	// Send response back to frontend
	if err := httputils.RespondWithJSON(w, h.logger, http.StatusOK, homeData); err != nil {
		h.logger.Error("failed to send response",
			"error", err,
		)
	}

	return
}
