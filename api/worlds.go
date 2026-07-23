package api

import (
	"net/http"
	"strconv"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/world"
)

// WorldResponse is the wire shape of a generated world. It deliberately
// mirrors only what world.Generate actually populates today (UWP and
// TradeCodes) rather than the full world.World struct — see world/generate.go's
// doc comment for what's not generated yet, and why.
type WorldResponse struct {
	Seed       int64             `json:"seed"`
	UWP        string            `json:"uwp"`
	TradeCodes []world.TradeCode `json:"tradeCodes"`
}

func handleWorldsRandom(w http.ResponseWriter, r *http.Request) {
	var seed *int64

	if raw := r.URL.Query().Get("seed"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "seed must be an integer")

			return
		}

		seed = &parsed
	}

	resolved := dice.ResolveSeed(seed)
	generated := world.Generate(dice.RollerFromSeed(resolved))

	writeJSON(w, http.StatusOK, WorldResponse{
		Seed:       resolved,
		UWP:        generated.UWP.String(),
		TradeCodes: generated.TradeCodes,
	})
}
