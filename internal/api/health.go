package api

import "net/http"

// HandlerHealth responds with a simple JSON status message.
// Used to verify the API is running and reachable.
func (cfg *APIConfig) HandlerHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok"}`))
}
