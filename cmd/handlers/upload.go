package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"bravo-kilo/cmd/middleware"
	"bravo-kilo/internal/utils"
)

// File handling
func (h *Handlers) UploadCSV(response http.ResponseWriter, request *http.Request) {
	// Check auth
	userID, ok := middleware.GetUserID(request.Context())
	if !ok {
		http.Error(response, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse multipart form, cap size@10MB
	err := request.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(response, "File too large", http.StatusBadRequest)
		return
	}

	// Get file from form data
	file, fileHeader, err := request.FormFile("file")
	if err != nil {
		http.Error(response, "Invalid file upload", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate Content-Type header
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType != "text/csv" {
		http.Error(response, "Invalid content type", http.StatusBadRequest)
		return
	}

	// Validate file type using magic numbers
	buf := make([]byte, 512)
	if _, err := file.Read(buf); err != nil {
		http.Error(response, "Error reading file", http.StatusInternalServerError)
		return
	}
	if !isCSV(buf) {
		http.Error(response, "Invalid file type", http.StatusBadRequest)
		return
	}

	// Reset file reader
	file.Seek(0, 0)

	// Sanitize and store file
	safeFileName := fmt.Sprintf("%d_%s", userID, sanitizeFileName(fileHeader.Filename))

	// Only upload to secure directory
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		http.Error(response, "Upload directory is not configured", http.StatusInternalServerError)
		return
	}

	destination := filepath.Join(uploadDir, safeFileName)

	// Ensure destination directory exists
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		http.Error(response, "Unable to create directory", http.StatusInternalServerError)
		return
	}

	// Ensure file is not saved outside the intended directory
	if !strings.HasPrefix(destination, uploadDir) {
		http.Error(response, "Invalid file path", http.StatusBadGateway)
		return
	}

	// Create or truncate file at path if its not already there
	outFile, err := os.Create(destination)
	if err != nil {
		http.Error(response, "Unable to save file", http.StatusInternalServerError)
		return
	}
	defer outFile.Close()

	// Copy content
	if _, err := io.Copy(outFile, file); err != nil {
		http.Error(response, "Unable to save file", http.StatusInternalServerError)
		return
	}

	// Log upload event
	h.logger.Info("File uploaded successfully", "userID", userID, "filename", safeFileName)

	// Call ParseAndProcessCSV, update request body to pass file path
	request.Body = io.NopCloser(strings.NewReader(destination))
	h.ParseAndProcessCSV(destination, response)
}


func isCSV(data []byte) bool {
	return http.DetectContentType(data) == "text/csv"
}

func sanitizeFileName(filename string) string {
	// Strip dir path
	filename = filepath.Base(filename)

	// Swap unsafe chars w/ underscore
	sanitized := utils.SanitizeChars(filename)

	// Make sure file ends w/ .csv
	if !strings.HasSuffix(sanitized, ".csv") {
		sanitized += ".csv"
	}

	return sanitized
}

// ParseAndProcessCSV
func (h *Handlers) ParseAndProcessCSV(filePath string, response http.ResponseWriter) {

	// Open the uploaded file
	file, err := os.Open(filePath)
	if err != nil {
		h.logger.Error("Unable to open file", "error", err)
		http.Error(response, "Unable to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Create csv reader
	reader := csv.NewReader(file)
	maxLength := 250

	// Process each record in file
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			h.logger.Error("Error reading file", "error", err)
			http.Error(response, "Error reading file", http.StatusInternalServerError)
		}

		// Sanitize, truncate, validate each field
		for i, field := range record {
			sanitizedField := utils.SanitizeChars(field)
			truncatedField := utils.TruncateField(sanitizedField, maxLength)

			if err := utils.ValidateFieldLength(truncatedField, maxLength); err != nil {
				h.logger.Error("Error in record %d: %v", i, err)
				http.Error(response, fmt.Sprintf("Error in record %d: %v", i, err), http.StatusBadRequest)
				return
			}
			record[i] = truncatedField
		}

		// Todo: create handler to validate + insert record into db
	}

	// Delete file after successful processing
	if err := os.Remove(filePath); err != nil {
		h.logger.Error("Failed to delete after processing", "filePath", filePath)
	}

	h.logger.Info("File processed and deleted successfully", "filePath", filePath)
	response.WriteHeader(http.StatusOK)
	json.NewEncoder(response).Encode(map[string]string{"message": "File processed successfully"})
}

func (h *Handlers) handleParsingError(filePath string, response http.ResponseWriter, err error) {
	// Log the error and send a response to the client
	h.logger.Error("Error during CSV processing", "error", err)

	// Attempt to delete the file
	if removeErr := os.Remove(filePath); removeErr != nil {
		h.logger.Error("Failed to delete file after error", "filePath", filePath, "error", removeErr)
	}

	http.Error(response, err.Error(), http.StatusBadRequest)
}
