// Package api implements the traveller HTTP API: route registration,
// handlers, and their JSON wire types. Kept out of package main so it's
// unit-testable with httptest and cmd/server stays a thin composition root.
package api

import (
	"encoding/json"
	"log"
	"net/http"
)

// NewMux returns the traveller API's routes. Unmatched paths get a JSON 404
// via handleNotFound, same as every other error response — the "/" pattern
// only ever fires when nothing more specific matched.
func NewMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("GET /worlds/random", handleWorldsRandom)
	mux.HandleFunc("/", handleNotFound)

	return mux
}

func handleNotFound(w http.ResponseWriter, _ *http.Request) {
	writeJSONError(w, http.StatusNotFound, "not found")
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

// writeJSON writes status and encodes v as the response body. v is always
// one of this package's own response types (WorldResponse, healthzResponse,
// errorResponse), so it's always marshalable — the any signature is just
// for reuse across handlers.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		// Headers and possibly some body bytes are already written, so
		// there's nothing left to do but log — the client just gets a
		// truncated/malformed response.
		log.Printf("api: writing response: %v", err)
	}
}
