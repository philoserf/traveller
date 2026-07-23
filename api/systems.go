package api

import (
	"net/http"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/world"
)

// GasGiantResponse is the wire shape of a Gas Giant occupying an orbit.
type GasGiantResponse struct {
	Size    string `json:"size"`
	Bracket string `json:"bracket"`
}

// StarResponse is the wire shape of a single star in a system. Orbit is
// nil for the Primary — it's the system's center, not a numbered orbit
// (see world.Orbit's doc comment on the sentinel this maps from).
type StarResponse struct {
	SpectralType       string `json:"spectralType"`
	SpectralDecimal    int    `json:"spectralDecimal"`
	LuminosityClass    string `json:"luminosityClass"`
	Role               string `json:"role"`
	Orbit              *int   `json:"orbit,omitempty"`
	HabitableZoneOrbit int    `json:"habitableZoneOrbit"`
	HasCompanion       bool   `json:"hasCompanion"`
}

// MainworldResponse is the wire shape of a system's mainworld and its
// placement. It deliberately duplicates WorldResponse's UWP/TradeCodes/
// TravelZone/Bases/PBG/Importance/Economic/Cultural fields rather than
// embedding WorldResponse: WorldResponse's own Seed field doesn't belong
// here — the mainworld and the rest of the system share one seed,
// SystemResponse's own Seed — and there's no other WorldResponse field
// this doesn't already need.
type MainworldResponse struct {
	Orbit      int               `json:"orbit"`
	AU         float64           `json:"au,omitempty"`
	Satellite  bool              `json:"satellite"`
	GasGiant   *GasGiantResponse `json:"gasGiant,omitempty"`
	UWP        string            `json:"uwp"`
	TradeCodes []world.TradeCode `json:"tradeCodes"`
	TravelZone string            `json:"travelZone"`
	Bases      []world.Base      `json:"bases"`
	PBG        string            `json:"pbg"`
	Importance int               `json:"importance"`
	Economic   EconomicResponse  `json:"economic"`
	Cultural   CulturalResponse  `json:"cultural"`
}

// SystemResponse is the wire shape of a generated star system: its stars
// and its mainworld's placement within them.
type SystemResponse struct {
	Seed      int64             `json:"seed"`
	Stars     []StarResponse    `json:"stars"`
	Mainworld MainworldResponse `json:"mainworld"`
}

// handleSystemsRandom godoc
//
//	@Summary		Generate a random star system
//	@Description	Rolls a Traveller5 mainworld and the system around it: stars, habitable zone, and mainworld placement.
//	@Tags			systems
//	@Produce		json
//	@Param			seed	query		int	false	"PRNG seed (omit for a time-derived seed)"
//	@Success		200		{object}	SystemResponse
//	@Failure		400		{object}	errorResponse	"seed is not an integer"
//	@Router			/systems/random [get]
func handleSystemsRandom(w http.ResponseWriter, r *http.Request) {
	seed, present, err := parseSeed(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "seed must be an integer")

		return
	}

	var seedPtr *int64
	if present {
		seedPtr = &seed
	}

	resolved := dice.ResolveSeed(seedPtr)
	roller := dice.RollerFromSeed(resolved)
	mainworld := world.Generate(roller)
	sys := world.GenerateSystem(roller, mainworld)

	writeJSON(w, http.StatusOK, toSystemResponse(resolved, sys))
}

func toSystemResponse(seed int64, sys world.StarSystem) SystemResponse {
	stars := make([]StarResponse, 0, len(sys.Orbits))

	for _, o := range sys.Orbits {
		if o.Star == nil {
			continue
		}

		stars = append(stars, toStarResponse(o))
	}

	mwOrbit := sys.Orbits[sys.MainworldOrbit]

	return SystemResponse{
		Seed:      seed,
		Stars:     stars,
		Mainworld: toMainworldResponse(sys, mwOrbit),
	}
}

func toStarResponse(o world.Orbit) StarResponse {
	star := o.Star

	resp := StarResponse{
		SpectralType:       string(star.SpectralType),
		SpectralDecimal:    star.SpectralDecimal,
		LuminosityClass:    star.LuminosityClass,
		Role:               star.Role.String(),
		HabitableZoneOrbit: star.HabitableZoneOrbit,
		HasCompanion:       star.Companion != nil,
	}

	if o.Number >= 0 {
		n := o.Number
		resp.Orbit = &n
	}

	return resp
}

func toMainworldResponse(sys world.StarSystem, mwOrbit world.Orbit) MainworldResponse {
	mw := mwOrbit.World

	resp := MainworldResponse{
		Orbit:      mwOrbit.Number,
		AU:         mwOrbit.AU,
		Satellite:  mwOrbit.Satellite,
		UWP:        mw.UWP.String(),
		TradeCodes: mw.TradeCodes,
		TravelZone: mw.TravelZone.String(),
		Bases:      mw.Bases,
		PBG:        mw.PBG.String(),
		Importance: int(mw.Importance),
		Economic: EconomicResponse{
			Resources:      mw.Economic.Resources,
			Labor:          mw.Economic.Labor,
			Infrastructure: mw.Economic.Infrastructure,
			Efficiency:     mw.Economic.Efficiency,
		},
		Cultural: CulturalResponse{
			Heterogeneity: mw.Cultural.Heterogeneity,
			Acceptance:    mw.Cultural.Acceptance,
			Strangeness:   mw.Cultural.Strangeness,
			Symbols:       mw.Cultural.Symbols,
		},
	}

	if mwOrbit.Satellite {
		if gg := sys.GasGiantAt(mwOrbit.Number); gg != nil {
			resp.GasGiant = &GasGiantResponse{Size: string(gg.Size), Bracket: gg.Bracket}
		}
	}

	return resp
}
