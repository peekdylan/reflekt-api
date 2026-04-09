package api

import (
	"encoding/json"
	"log"
	"net/http"
)

// respondWithJSON writes a JSON response with the given status code and payload.
// This is a helper used by all handlers to keep response formatting consistent.
func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	// Marshal the payload into JSON bytes
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal JSON response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}

// respondWithError writes a JSON error response with the given status code and message.
// This keeps error responses consistent across the entire API.
func respondWithError(w http.ResponseWriter, status int, message string) {
	// Error responses always follow this structure so clients can rely on it
	respondWithJSON(w, status, map[string]string{"error": message})
}
