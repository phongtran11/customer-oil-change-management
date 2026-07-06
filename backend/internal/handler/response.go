package handler

import (
	"encoding/json"
	"net/http"
)

// JSON writes a JSON-encoded response with the given HTTP status code.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// At this point the header has been sent, so we can only log.
		_ = err
	}
}

// Error writes a JSON error response: {"error": "message"}.
func Error(w http.ResponseWriter, status int, msg string) {
	JSON(w, status, map[string]string{"error": msg})
}
