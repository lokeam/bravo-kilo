package handlers

import (
	"bravo-kilo/internal/utils"
	"net/http"
)

// HandleExportUserBooks exports a user's books as a CSV file
func (h *Handlers) HandleExportUserBooks(response http.ResponseWriter, request *http.Request) {
	userID, err := utils.ExtractUserIDFromJWT(request)
	if err != nil {
		h.logger.Error("Error extracting userID from JWT", "error", err)
		http.Error(response, "Unauthorized access", http.StatusUnauthorized)
		return
	}

	response.Header().Set("Content-Type", "text/csv")
	response.Header().Set("Content-Disposition", "attachment; filename=books.csv")

	if err := h.models.Book.GenerateBookCSV(userID, response); err != nil {
		h.logger.Error("Error generating CSV for user books", "userID", userID, "error", err)
		http.Error(response, "Error generating CSV", http.StatusInternalServerError)
		return
	}
}
