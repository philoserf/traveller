package api

import (
	"net/http"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/system"
	"github.com/philoserf/traveller/world"
)

// GasGiantResponse is the wire shape of a Gas Giant occupying an orbit.
type GasGiantResponse struct {
	Size    string `json:"size"`
	Bracket string `json:"bracket"`
}

// StarResponse is the wire shape of a single star in a system. Orbit is
// nil for the Primary — it's the system's center, not a numbered orbit
// (see system.Orbit's doc comment on the sentinel this maps from).
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
	Orbit     int     `json:"orbit"`
	AU        float64 `json:"au,omitempty"`
	Satellite bool    `json:"satellite"`
	// Close is meaningful only when Satellite is true — Close (tidally
	// locked) vs Far, mirroring system.Orbit.Close.
	Close      bool              `json:"close"`
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

// SatelliteResponse is the wire shape of a satellite orbiting a Gas Giant
// or a placed World. IsMainworld is true when this satellite is the
// system's own mainworld (a mainworld can itself be a satellite of a Gas
// Giant — see MainworldResponse.Satellite).
type SatelliteResponse struct {
	Close       bool              `json:"close"`
	IsMainworld bool              `json:"isMainworld,omitempty"`
	UWP         string            `json:"uwp"`
	TradeCodes  []world.TradeCode `json:"tradeCodes"`
}

// BodyResponse is the wire shape of a non-star body placed in the system:
// either a Gas Giant, or a placed World with its own UWP/TradeCodes
// (GasGiant is nil in that case), plus any Satellites of its own. Ring is
// whichever of GasGiant.Ring/World.Ring applies. IsMainworld is true when
// this body is the system's own mainworld (never true for a Gas Giant —
// the mainworld is always a World).
type BodyResponse struct {
	Orbit       int                 `json:"orbit"`
	Ring        bool                `json:"ring,omitempty"`
	IsMainworld bool                `json:"isMainworld,omitempty"`
	GasGiant    *GasGiantResponse   `json:"gasGiant,omitempty"`
	UWP         string              `json:"uwp,omitempty"`
	TradeCodes  []world.TradeCode   `json:"tradeCodes,omitempty"`
	Satellites  []SatelliteResponse `json:"satellites,omitempty"`
}

// StarGroupResponse is one star and every non-satellite body it hosts
// (sorted by orbit number) — the shared orbit-numbering across a
// multi-star system means a body's Orbit alone doesn't say which star
// placed it, so bodies are nested under their hosting star instead of
// carrying a separate host reference.
type StarGroupResponse struct {
	Star   StarResponse   `json:"star"`
	Bodies []BodyResponse `json:"bodies"`
}

// SystemResponse is the wire shape of a generated star system: its
// mainworld's placement, and every star with the bodies it hosts (Gas
// Giants, Belts, secondary worlds, and their satellites — see
// system.GenerateSystem for what's placed and why).
type SystemResponse struct {
	Seed       int64               `json:"seed"`
	StarGroups []StarGroupResponse `json:"starGroups"`
	Mainworld  MainworldResponse   `json:"mainworld"`
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
		writeJSONError(w, "seed must be an integer")

		return
	}

	var seedPtr *int64
	if present {
		seedPtr = &seed
	}

	resolved := dice.ResolveSeed(seedPtr)
	roller := dice.RollerFromSeed(resolved)
	mainworld := world.Generate(roller)
	sys := system.GenerateSystem(roller, mainworld)

	writeJSON(w, http.StatusOK, toSystemResponse(resolved, sys))
}

func toSystemResponse(seed int64, sys system.StarSystem) SystemResponse {
	starOrbits, bodiesByRole, satellitesOf := sys.SystemBodies()

	starGroups := make([]StarGroupResponse, 0, len(starOrbits))

	for _, so := range starOrbits {
		bodies := bodiesByRole[so.Star.Role]
		bodyResponses := make([]BodyResponse, 0, len(bodies))

		for _, o := range bodies {
			bodyResponses = append(bodyResponses, toBodyResponse(sys, o, satellitesOf[o.Number]))
		}

		starGroups = append(starGroups, StarGroupResponse{Star: toStarResponse(so), Bodies: bodyResponses})
	}

	return SystemResponse{
		Seed:       seed,
		StarGroups: starGroups,
		Mainworld:  toMainworldResponse(sys, sys.Orbits[sys.MainworldOrbit]),
	}
}

// toBodyResponse builds the wire shape for a single non-star,
// non-Satellite Orbit entry — a Gas Giant, or a placed World — with
// satellites (already collected by Number) nested under it. sys.IsMainworld
// marks whichever entry (this body, or one of its satellites) is the
// system's own mainworld.
func toBodyResponse(sys system.StarSystem, o system.Orbit, satellites []system.Orbit) BodyResponse {
	resp := BodyResponse{Orbit: o.Number, IsMainworld: sys.IsMainworld(o)}

	if o.GasGiant != nil {
		resp.Ring = o.GasGiant.Ring
		resp.GasGiant = &GasGiantResponse{Size: string(o.GasGiant.Size), Bracket: o.GasGiant.Bracket}
	} else {
		resp.Ring = o.World.Ring
		resp.UWP = o.World.UWP.String()
		resp.TradeCodes = o.World.TradeCodes
	}

	for _, sat := range satellites {
		resp.Satellites = append(resp.Satellites, SatelliteResponse{
			Close:       sat.Close,
			IsMainworld: sys.IsMainworld(sat),
			UWP:         sat.World.UWP.String(),
			TradeCodes:  sat.World.TradeCodes,
		})
	}

	return resp
}

func toStarResponse(o system.Orbit) StarResponse {
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

func toMainworldResponse(sys system.StarSystem, mwOrbit system.Orbit) MainworldResponse {
	mw := mwOrbit.World

	resp := MainworldResponse{
		Orbit:      mwOrbit.Number,
		AU:         mwOrbit.AU,
		Satellite:  mwOrbit.Satellite,
		Close:      mwOrbit.Close,
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
