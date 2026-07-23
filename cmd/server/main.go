// Command server runs the traveller HTTP API.
//
//	@title			Traveller API
//	@version		0.1
//	@description	API-first Traveller5 (T5) rules engine: world/character/starship generation.
//	@BasePath		/
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/philoserf/traveller/api"
)

func main() {
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           api.NewMux(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("server: listening on %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
