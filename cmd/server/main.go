// Command server runs the traveller HTTP API — an API-first Traveller5
// (T5) rules engine: world/character/starship generation. See package
// api for the routes and api/*.go's handler doc comments for what each
// one does, or the README's API section for an endpoint table.
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
