package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/lokeam/bravo-kilo/config"
	"github.com/lokeam/bravo-kilo/internal/shared/jwt"
	"google.golang.org/api/option"
)

// HandleGetGeminiBookSummary processes the Google Gemini request
func (h *BookHandlers) HandleGetGeminiBookSummary(response http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	// Extract user ID from JWT
	h.logger.Info("Handling Google Gemini request")
	_, err := jwt.ExtractUserIDFromJWT(request, config.AppConfig.JWTPublicKey)
	if err != nil {
		h.logger.Error("Error extracting user ID", "error", err)
		return
	}

	prompt := request.URL.Query().Get("prompt")
	if prompt == "" {
		http.Error(response, "Query parameter required in request", http.StatusBadRequest)
		return
	}

    // Check Redis cache with proper error handling
    cachedResponse, found, err := h.bookCacheService.GetCachedGeminiResponse(ctx, prompt)
    if err != nil {
        h.logger.Error("Cache retrieval error", "error", err)
        // Continue execution to get fresh data instead of failing
    } else if found {
        h.logger.Debug("Cache hit for Gemini response", "prompt", prompt)
        response.Header().Set("Content-Type", "application/json")
        if _, err := response.Write([]byte(cachedResponse)); err != nil {
            h.logger.Error("Error writing cached response", "error", err)
            http.Error(response, "Error writing response", http.StatusInternalServerError)
        }
        return
    }

	// Initialize Model
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GOOGLE_GEMINI_API_KEY")))
	if err != nil {
		h.logger.Error("Failed to initialize Google Gemini client", "error", err)
		http.Error(response, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash")
	h.logger.Info("About to make Google Gemini request")

	// Make request to Google Gemini API
	responseData, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		h.logger.Error("Error calling Gemini API", "error", err)
		http.Error(response, "Error calling Gemini API", http.StatusInternalServerError)
		return
	}

	// Extract Parts array from the first candidate's content
	var parts []genai.Part
	if len(responseData.Candidates) > 0 && responseData.Candidates[0].Content != nil {
		parts = responseData.Candidates[0].Content.Parts
	} else {
		http.Error(response, "No valid content received from Gemini API", http.StatusInternalServerError)
		return
	}

	// Prepare the formatted response
	formattedResponse := map[string]interface{}{
		"parts": parts,
	}
	jsonResponse, err := json.Marshal(formattedResponse)
	if err != nil {
			http.Error(response, "Error formatting response", http.StatusInternalServerError)
			return
	}

    // Cache the response with error logging
    if err := h.bookCacheService.SetCachedGeminiResponse(ctx, prompt, string(jsonResponse)); err != nil {
			h.logger.Error("Cache storage error",
					"error", err,
					"prompt", prompt,
					"responseSize", len(jsonResponse),
			)
			// Continue execution as caching failure shouldn't affect the response
		} else {
			h.logger.Debug("Successfully cached Gemini response", "prompt", prompt)
		}

	// Set response headers and return the formatted response
	response.Header().Set("Content-Type", "application/json")
	if _, err := response.Write(jsonResponse); err != nil {
		http.Error(response, "Error writing response", http.StatusInternalServerError)
		return
	}
}
