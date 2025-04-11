package api

import (
	"encoding/json"
	"log"
	"net/http"

	"time-guard-bot/internal/models"
)

// Sends a JSON response
func sendJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Sends a JSON error response
func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(models.ErrorResponse{Error: message}); err != nil {
		log.Printf("Error encoding JSON error response: %v", err)
	}
}
