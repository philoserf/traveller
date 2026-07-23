// Package api implements the traveller HTTP API: route registration,
// handlers, and their JSON wire types. Kept out of package main so it's
// unit-testable with httptest and cmd/server stays a thin composition root.
package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
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
	mux.HandleFunc("GET /systems/random", handleSystemsRandom)

	return jsonErrors(mux)
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

// parseSeed parses the "seed" query parameter from r. present is false if
// the parameter was absent (the caller should resolve a fresh seed via
// dice.ResolveSeed(nil)); err is non-nil if present but not a valid
// integer. Shared by every /random-style handler so seed validation can't
// drift between them.
func parseSeed(r *http.Request) (int64, bool, error) {
	raw := r.URL.Query().Get("seed")
	if raw == "" {
		return 0, false, nil
	}

	seed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, false, fmt.Errorf("parsing seed query param: %w", err)
	}

	return seed, true, nil
}

// jsonErrorWriter substitutes a JSON body for net/http's default
// plain-text 404/405 one — see jsonErrors. Unlike writeJSONError, it's
// unconditional: every WriteHeader call it sees is assumed to be net/http's
// own fallback, never an application handler's, because jsonErrors only
// ever wraps a request that mux.Handler has already confirmed will hit
// that fallback (no application handler runs on this path at all).
type jsonErrorWriter struct {
	http.ResponseWriter

	intercepting bool
}

func (w *jsonErrorWriter) WriteHeader(status int) {
	w.intercepting = true
	writeJSON(w.ResponseWriter, status, errorResponse{Error: http.StatusText(status)})
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

// jsonErrors wraps mux so unmatched paths (404) and wrong-method requests
// (405) — cases where no application handler ever runs — get a JSON
// {"error": "..."} body instead of net/http's default plain text, matching
// every in-handler error's envelope (writeJSONError). mux.Handler decides
// *before* dispatch whether a request will hit ServeMux's own fallback (an
// empty pattern) rather than inspecting the status code after the fact, so
// an application handler that legitimately calls writeJSONError with its
// own 404/405 status is never touched by this wrapper at all. Not a mux
// route — see NewMux's doc comment for why a catch-all route was tried
// and reverted.
func jsonErrors(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, pattern := mux.Handler(r); pattern == "" {
			mux.ServeHTTP(&jsonErrorWriter{ResponseWriter: w}, r)

			return
		}

		mux.ServeHTTP(w, r)
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
