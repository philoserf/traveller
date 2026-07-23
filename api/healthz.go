package api

import "net/http"

type healthzResponse struct {
	Status string `json:"status"`
}

// handleHealthz godoc
//
//	@Summary		Health check
//	@Description	Reports whether the server is up.
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	healthzResponse
//	@Router			/healthz [get]
func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthzResponse{Status: "ok"})
}
