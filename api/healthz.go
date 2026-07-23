package api

import "net/http"

type healthzResponse struct {
	Status string `json:"status"`
}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, healthzResponse{Status: "ok"})
}
