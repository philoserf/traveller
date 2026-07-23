// Package api implements the traveller HTTP API: route registration,
// handlers, and their JSON wire types. Kept out of package main so it's
// unit-testable with httptest and cmd/server stays a thin composition root.
package api

import (
	"encoding/json"
	"log"
	"net/http"
)

// NewMux returns the traveller API's routes.
//
// Unmatched paths and wrong-method requests fall through to net/http's
// built-in plain-text 404/405 handling rather than a JSON error envelope.
// A catch-all "/" pattern was tried here and reverted: registering it makes
// every method match at every path from the mux's perspective, which
// silently disables ServeMux's automatic 405 Method Not Allowed detection
// for the routes above — a POST to /healthz became an indistinguishable
// 404 instead. Making 404 *and* 405 both return JSON without that
// regression needs a ResponseWriter-wrapping middleware that inspects the
// mux's already-decided status code, not a route registration — deferred
// until there's a second real consumer to justify it.
func NewMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("GET /worlds/random", handleWorldsRandom)

	return mux
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
