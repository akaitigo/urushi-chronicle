// Package handler provides HTTP handlers for the urushi-chronicle REST API.
package handler

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents an API error response body.
type ErrorResponse struct {
	Error  string   `json:"error"`
	Errors []string `json:"errors,omitempty"`
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}

// writeValidationErrors writes a 400 response with validation errors.
func writeValidationErrors(w http.ResponseWriter, errors []string) {
	writeJSON(w, http.StatusBadRequest, ErrorResponse{
		Error:  "validation error",
		Errors: errors,
	})
}
