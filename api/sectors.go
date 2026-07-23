package api

import (
	"net/http"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/world"
)

// HexResponse is the wire shape of one located hex in a Sector's grid.
// System is nil for an empty (deep space) hex.
type HexResponse struct {
	Location string          `json:"location"`
	System   *SystemResponse `json:"system,omitempty"`
}

// SectorResponse is the wire shape of a generated sector: its name and
// every Hex in its 32x40 grid (world.Sector.Hexes' own documented order),
// optionally filtered down to a single Subsector — see handleSectorsRandom.
type SectorResponse struct {
	Name  string        `json:"name"`
	Hexes []HexResponse `json:"hexes"`
}

// handleSectorsRandom godoc
//
//	@Summary		Generate a random sector
//	@Description	Rolls a full Traveller5 sector: a 32x40 hex grid, each hex
//	@Description	either empty or holding a complete generated star system.
//	@Tags			sectors
//	@Produce		json
//	@Param			seed		query		int		false	"PRNG seed (omit for a time-derived seed)"
//	@Param			name		query		string	false	"Sector name (default: Unnamed)"
//	@Param			density		query		string	false	"Density name, e.g. Standard (default) or Core — see world.Density"
//	@Param			subsector	query		string	false	"Single letter A-P — filter the response to that 80-hex block only"
//	@Success		200			{object}	SectorResponse
//	@Failure		400			{object}	errorResponse	"seed/density/subsector invalid"
//	@Router			/sectors/random [get]
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

	density := world.DensityStandard

	if raw := r.URL.Query().Get("density"); raw != "" {
		var ok bool

		density, ok = world.ParseDensity(raw)
		if !ok {
			writeJSONError(
				w,
				"density must be one of: Extra Galactic, Rift, Sparse, Scattered, Standard, Dense, Cluster, Core",
			)

			return
		}
	}

	subsector := r.URL.Query().Get("subsector")
	if subsector != "" && (len(subsector) != 1 || !world.ValidSubsectorLetter(subsector[0])) {
		writeJSONError(w, "subsector must be a single letter A-P")

		return
	}

	var seedPtr *int64
	if present {
		seedPtr = &seed
	}

	resolved := dice.ResolveSeed(seedPtr)
	sector := world.GenerateSector(resolved, name, density)

	hexes := sector.Hexes
	if subsector != "" {
		hexes = sector.Subsector(subsector[0])
	}

	writeJSON(w, http.StatusOK, toSectorResponse(resolved, sector.Name, hexes))
}

// toSectorResponse builds the wire shape for hexes, generated under
// sectorSeed. Each populated hex's own SystemResponse.Seed is
// world.HexSeed(sectorSeed, hex.Location) — the same per-hex seed
// GenerateSector itself derived to roll that hex — not sectorSeed
// directly, so GET /systems/random?seed=<that value> actually reproduces
// that specific hex's system (GenerateSector gives every hex its own
// independent Roller rather than sharing one sequential stream across
// the whole grid).
func toSectorResponse(sectorSeed int64, name string, hexes []world.Hex) SectorResponse {
	resp := SectorResponse{Name: name, Hexes: make([]HexResponse, 0, len(hexes))}

	for _, hex := range hexes {
		hr := HexResponse{Location: hex.Location}

		if hex.System != nil {
			sys := toSystemResponse(world.HexSeed(sectorSeed, hex.Location), *hex.System)
			hr.System = &sys
		}

		resp.Hexes = append(resp.Hexes, hr)
	}

	return resp
}
