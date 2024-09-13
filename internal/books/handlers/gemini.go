package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/lokeam/bravo-kilo/internal/shared/utils"
	"google.golang.org/api/option"
)

// Todo: replace this with a Redis cache in production
var cache = make(map[string]string)

func cacheGet(key string) (string, bool) {
	value, exists := cache[key]
	return value, exists
}

func cacheSet(key string, value string) {
	cache[key] = value
}

// HandleGetGeminiBookSummary processes the Google Gemini request
func (h *Handlers) HandleGetGeminiBookSummary(response http.ResponseWriter, request *http.Request) {
	// Extract user ID from JWT
	h.logger.Info("Handling Google Gemini request")
	_, err := utils.ExtractUserIDFromJWT(request)
	if err != nil {
		h.logger.Error("Error extracting user ID", "error", err)
		return
	}

	prompt := request.URL.Query().Get("prompt")
	if prompt == "" {
		http.Error(response, "Query parameter required in request", http.StatusBadRequest)
		return
	}

	// Check if response has been cached
	if cachedResponse, found := cacheGet(prompt); found {
		response.Header().Set("Content-Type", "application/json")
		response.Write([]byte(cachedResponse))
		return
	}

	// Initialize Model
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

	// Cache the response
	jsonResponse, err := json.Marshal(formattedResponse)
	if err == nil {
		cacheSet(prompt, string(jsonResponse))
	}

	// Set response headers and return the formatted response
	response.Header().Set("Content-Type", "application/json")
	if _, err := response.Write(jsonResponse); err != nil {
		http.Error(response, "Error writing response", http.StatusInternalServerError)
		return
	}
}
