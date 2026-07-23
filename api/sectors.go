package api

import (
	"net/http"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/sector"
)

// HexResponse is the wire shape of one located hex in a Sector's grid.
// System is nil for an empty (deep space) hex.
type HexResponse struct {
	Location string          `json:"location"`
	System   *SystemResponse `json:"system,omitempty"`
}

// SectorResponse is the wire shape of a generated sector: its name and
// every Hex in its 32x40 grid (sector.Sector.Hexes' own documented order),
// optionally filtered down to a single Subsector — see handleSectorsRandom.
type SectorResponse struct {
	Seed  int64         `json:"seed"`
	Name  string        `json:"name"`
	Hexes []HexResponse `json:"hexes"`
}

// handleSectorsRandom handles GET /sectors/random: rolls a full
// Traveller5 sector, a 32x40 hex grid where each hex is either empty or
// holds a complete generated star system. Optional query params: seed
// (int, omit for a time-derived seed), name (string, default "Unnamed"),
// density (string, e.g. Standard (default) or Core — see sector.Density),
// subsector (single letter A-P, filters the response to that 80-hex
// block only). Responds 200 with a SectorResponse, or 400 with an
// errorResponse if seed/density/subsector is invalid.
func handleSectorsRandom(w http.ResponseWriter, r *http.Request) {
	seed, present, err := parseSeed(r)
	if err != nil {
		writeJSONError(w, "seed must be an integer")

		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		name = "Unnamed"
	}

	density := sector.DensityStandard

	if raw := r.URL.Query().Get("density"); raw != "" {
		var ok bool

		density, ok = sector.ParseDensity(raw)
		if !ok {
			writeJSONError(
				w,
				"density must be one of: Extra Galactic, Rift, Sparse, Scattered, Standard, Dense, Cluster, Core",
			)

			return
		}
	}

	subsectorLetter := r.URL.Query().Get("subsector")
	if subsectorLetter != "" && (len(subsectorLetter) != 1 || !sector.ValidSubsectorLetter(subsectorLetter[0])) {
		writeJSONError(w, "subsector must be a single letter A-P")

		return
	}

	var seedPtr *int64
	if present {
		seedPtr = &seed
	}

	resolved := dice.ResolveSeed(seedPtr)
	sec := sector.GenerateSector(resolved, name, density)

	hexes := sec.Hexes
	if subsectorLetter != "" {
		hexes = sec.Subsector(subsectorLetter[0])
	}

	writeJSON(w, http.StatusOK, toSectorResponse(resolved, sec.Name, hexes))
}

// toSectorResponse builds the wire shape for hexes, generated under
// sectorSeed. Each populated hex's own SystemResponse.Seed is
// sector.HexSeed(sectorSeed, hex.Location) — the same per-hex seed
// GenerateSector itself derived to roll that hex — not sectorSeed
// directly, so GET /systems/random?seed=<that value> actually reproduces
// that specific hex's system (GenerateSector gives every hex its own
// independent Roller rather than sharing one sequential stream across
// the whole grid).
func toSectorResponse(sectorSeed int64, name string, hexes []sector.Hex) SectorResponse {
	resp := SectorResponse{Seed: sectorSeed, Name: name, Hexes: make([]HexResponse, 0, len(hexes))}

	for _, hex := range hexes {
		hr := HexResponse{Location: hex.Location}

		if hex.System != nil {
			sys := toSystemResponse(sector.HexSeed(sectorSeed, hex.Location), *hex.System)
			hr.System = &sys
		}

		resp.Hexes = append(resp.Hexes, hr)
	}

	return resp
}
