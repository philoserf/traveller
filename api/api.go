// Package api implements the traveller HTTP API: route registration,
// handlers, and their JSON wire types. Kept out of package main so it's
// unit-testable with httptest and cmd/server stays a thin composition root.
package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// NewMux returns the traveller API's routes, wrapped in jsonErrors so every
// response — including unmatched paths (404) and wrong-method requests
// (405) — uses the same JSON error envelope as in-handler errors
// (writeJSONError).
//
// A catch-all "/" pattern was tried here once and reverted: registering it
// makes every method match at every path from the mux's perspective, which
// silently disables ServeMux's automatic 405 Method Not Allowed detection
// for the routes above — a POST to /healthz became an indistinguishable
// 404 instead. jsonErrors sidesteps that by wrapping the ResponseWriter
// instead: it lets ServeMux decide the status code exactly as before, and
// only swaps in a JSON body for whatever status ServeMux already chose.
func NewMux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("GET /worlds/random", handleWorldsRandom)

	return jsonErrors(mux)
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

// jsonErrorWriter intercepts WriteHeader for 404/405 so a JSON body can
// replace net/http's default plain-text one — see jsonErrors.
type jsonErrorWriter struct {
	http.ResponseWriter

	intercepting bool
}

func (w *jsonErrorWriter) WriteHeader(status int) {
	if status != http.StatusNotFound && status != http.StatusMethodNotAllowed {
		w.ResponseWriter.WriteHeader(status)

		return
	}

	w.intercepting = true
	w.ResponseWriter.Header().Set("Content-Type", "application/json")
	w.ResponseWriter.WriteHeader(status)

	if err := json.NewEncoder(w.ResponseWriter).Encode(errorResponse{Error: http.StatusText(status)}); err != nil {
		log.Printf("api: writing error response: %v", err)
	}
}

// Write discards net/http's own plain-text 404/405 body (from http.Error,
// called right after WriteHeader) once WriteHeader has already written a
// JSON body in its place.
func (w *jsonErrorWriter) Write(b []byte) (int, error) {
	if w.intercepting {
		return len(b), nil
	}

	n, err := w.ResponseWriter.Write(b)
	if err != nil {
		return n, fmt.Errorf("api: writing response body: %w", err)
	}

	return n, nil
}

// jsonErrors wraps h so unmatched paths (404) and wrong-method requests
// (405) get a JSON {"error": "..."} body instead of net/http's default
// plain text, matching every in-handler error's envelope (writeJSONError).
// A ResponseWriter wrapper, not a mux route — see NewMux's doc comment for
// why a catch-all route was tried and reverted.
func jsonErrors(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(&jsonErrorWriter{ResponseWriter: w}, r)
	})
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
