package api

import "net/http"

type healthzResponse struct {
	Status string `json:"status"`
}

// handleHealthz handles GET /healthz: reports whether the server is up.
// Responds 200 with a healthzResponse.
func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthzResponse{Status: "ok"})
}
