package api

import (
	"net/http"
	"strconv"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/world"
)

// EconomicResponse is the wire shape of a world's Economic (Ex) extension.
// A local mirror of world.Economic with JSON tags, rather than adding JSON
// concerns to the domain type directly — same reasoning as WorldResponse
// itself.
type EconomicResponse struct {
	Resources      int `json:"resources"`
	Labor          int `json:"labor"`
	Infrastructure int `json:"infrastructure"`
	Efficiency     int `json:"efficiency"`
}

// CulturalResponse is the wire shape of a world's Cultural (Cx) extension.
type CulturalResponse struct {
	Heterogeneity int `json:"heterogeneity"`
	Acceptance    int `json:"acceptance"`
	Strangeness   int `json:"strangeness"`
	Symbols       int `json:"symbols"`
}

// WorldResponse is the wire shape of a generated world. It deliberately
// mirrors only what world.Generate actually populates today (UWP,
// TradeCodes, Bases, PBG, Importance, Economic, Cultural) rather than the
// full world.World struct — see world/generate.go's doc comment for
// what's not generated yet, and why. PBG is rendered as its 3-character
// string form, same as UWP, rather than its raw ehex.Value struct —
// consistent wire representation, and it sidesteps deciding how
// ehex.Value itself should marshal to JSON.
type WorldResponse struct {
	Seed       int64             `json:"seed"`
	UWP        string            `json:"uwp"`
	TradeCodes []world.TradeCode `json:"tradeCodes"`
	Bases      []world.Base      `json:"bases"`
	PBG        string            `json:"pbg"`
	Importance int               `json:"importance"`
	Economic   EconomicResponse  `json:"economic"`
	Cultural   CulturalResponse  `json:"cultural"`
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
		Bases:      generated.Bases,
		PBG:        generated.PBG.String(),
		Importance: int(generated.Importance),
		Economic: EconomicResponse{
			Resources:      generated.Economic.Resources,
			Labor:          generated.Economic.Labor,
			Infrastructure: generated.Economic.Infrastructure,
			Efficiency:     generated.Economic.Efficiency,
		},
		Cultural: CulturalResponse{
			Heterogeneity: generated.Cultural.Heterogeneity,
			Acceptance:    generated.Cultural.Acceptance,
			Strangeness:   generated.Cultural.Strangeness,
			Symbols:       generated.Cultural.Symbols,
		},
	})
}
