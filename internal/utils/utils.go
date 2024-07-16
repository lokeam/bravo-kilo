package utils

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

func ReadJSON(response http.ResponseWriter, request *http.Request, data interface{}) error {
	// Set size limit to 1mb
	maxBytes := 1048576
	request.Body = http.MaxBytesReader(response, request.Body, int64(maxBytes))

	decoder := json.NewDecoder(request.Body)
	err := decoder.Decode(data)
	if err != nil {
		return err
	}

	// Make sure we only have a single json value
	err = decoder.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("error: request body must have a single json value")
	}

	return nil
}

func WriteJSON(response http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			response.Header()[key] = value
		}
	}

	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)
	_,err = response.Write(out)
	if err != nil {
		return err
	}

	return nil
}

// func ErrorJSON(response http.ResponseWriter, err error, status ...int) {
// 	statusCode := http.StatusBadRequest

// 	if len(status) > 0 {
// 		statusCode = status[0]
// 	}

// 	var payload jsonResponse
// 	payload.Error = true
// 	payload.Message = err.Error()

// 	app.writeJSON(response, statusCode, payload)
// }

// Helper function to split full name into first and last names
func SplitFullName(fullName string) (string, string) {
	parts := strings.Fields(fullName)
	if len(parts) == 0 {
			return "", ""
	}
	firstName := parts[0]
	lastName := ""
	if len(parts) > 1 {
			lastName = strings.Join(parts[1:], " ")
	}
	return firstName, lastName
}
