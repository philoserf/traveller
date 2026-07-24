package system

import (
	"slices"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
	"github.com/philoserf/traveller/world"
)

// primaryOrbitNumber is the sentinel Orbit.Number for the Primary star
// itself — "Primary = at center of System" (Book 3 p.28), not a numbered
// planetary orbit around anything. Distinct from every real orbit number
// (0-20 per orbitAUTable), which are all non-negative.
const primaryOrbitNumber = -1

// GenerateSystem builds a StarSystem around an already-generated
// mainworld (from Generate): rolls the Primary star (and, independently,
// whether a Close/Near/Far star and any Companions exist), computes each
// star's HZ orbit and precluded-orbit ceiling (for oversized stars —
// precludedOrbitHost), places the mainworld (as a Planet, as a Satellite
// of a freshly rolled Gas Giant, or — if the mainworld is itself an
// Asteroid Belt — via the Belt placement roll instead of HZ+Var), then
// rolls and places every other Gas Giant, Belt, and secondary world
// across all stars (placeGasGiants/placeBelts/placeOtherWorlds), each
// with its own satellites and Rings, and merges the newly-derivable
// orbit-dependent trade codes (DeriveOrbitTradeCodes) into every placed
// world's TradeCodes.
func GenerateSystem(r *dice.Roller, mainworld world.World) StarSystem {
	primaryFlux := r.Flux()
	primary := rollStar(r, primaryFlux, true)
	primary.Role = Primary
	primary.Companion = attachCompanion(r, primaryFlux, Primary)

	orbits := []Orbit{{Number: primaryOrbitNumber, Star: &primary}}

	if r.Flux() >= starPresenceFlux {
		orbits = append(orbits, rollAndPlaceStar(r, primaryFlux, Close, r.D6()-1))
	}

	if r.Flux() >= starPresenceFlux {
		orbits = append(orbits, rollAndPlaceStar(r, primaryFlux, Near, 5+r.D6()))
	}

	if r.Flux() >= starPresenceFlux {
		orbits = append(orbits, rollAndPlaceStar(r, primaryFlux, Far, 11+r.D6()))
	}

	orbits, mainworldOrbitIndex := placeMainworld(r, orbits, primary, mainworld)
	satelliteOfGasGiant := orbits[mainworldOrbitIndex].Satellite
	mw := *orbits[mainworldOrbitIndex].World

	// Book 3 p.29's "W Worlds" formula (Total Worlds = MW + GG + Belts +
	// 2D) needs Gas Giant / Belt counts as inputs, not outputs — they're
	// mw.PBG.GasGiants/.Belts, already rolled by Generate (rollPBG's own
	// doc comment: "describe the whole system, not just this world").
	// Phase 1 generated them but never consulted them for placement.
	gasGiantsToPlace := int(mw.PBG.GasGiants)

	initialSGGCount := 0

	if satelliteOfGasGiant {
		// The satellite-hosting Gas Giant already placed above counts
		// against this total, per P1's own sequence ("Place Mainworld"
		// — including its satellite GG — immediately followed by "Place
		// Gas Giants" for the rest). PBG.GasGiants can still be 0 even
		// though a satellite mainworld always gets one placed anyway (a
		// real edge case — see placeMainworld's doc comment) — clamped
		// here so gasGiantsToPlace never goes negative.
		gasGiantsToPlace = max(gasGiantsToPlace-1, 0)

		// The GG-vs-SGG "every second SGG converts to an IG" counter
		// (placeGasGiants) needs to know about this one too, or it
		// mis-numbers every SGG rolled after it.
		if gg := orbits[mainworldOrbitIndex-1].GasGiant; gg != nil && gg.Bracket == "SGG" {
			initialSGGCount = 1
		}
	}

	maxPopulation := ehex.Value(0)
	if mw.UWP.Population > 0 {
		maxPopulation = mw.UWP.Population - 1 // Book 3 p.29: "Subject to: Max Pop= MW Pop - 1"
	}

	hosts := availableHosts(orbits)
	placeGasGiants(r, &orbits, hosts, gasGiantsToPlace, initialSGGCount)
	placeBelts(r, &orbits, hosts, int(mw.PBG.Belts), maxPopulation)
	placeOtherWorlds(r, &orbits, hosts, r.TwoD6(), maxPopulation)

	// Every top-level body — the mainworld, its host Gas Giant if any,
	// and everything just placed above — gets its own satellite roll
	// (Book 3 p.29: "For Each World in the System" + Gas Giants) — except
	// Asteroid Belts, which the satellite-count table (Gas Giants/Inners/
	// Hospitables/Outers) has no row for at all. Snapshotting the current
	// top-level bodies first, rather than ranging over orbits live, is
	// what keeps newly-appended satellites from being mistaken for more
	// top-level bodies to recurse into — satellites don't get their own
	// satellites.
	topLevel := make([]Orbit, 0, len(orbits))

	for _, o := range orbits {
		if o.Satellite {
			continue
		}

		if o.GasGiant != nil || (o.World != nil && !slices.Contains(o.World.TradeCodes, world.AsteroidBelt)) {
			topLevel = append(topLevel, o)
		}
	}

	for _, parent := range topLevel {
		generateSatellitesForBody(r, &orbits, parent, parent.HostHZOrbit, maxPopulation)
	}

	return StarSystem{
		Orbits:         orbits,
		MainworldOrbit: mainworldOrbitIndex,
	}
}

// placeMainworld places mainworld into orbits (which already holds the
// system's stars): as a Planet, as a Satellite of a freshly rolled Gas
// Giant, as a BigWorld (if Table 2C says Satellite but the system's own
// rolled PBG.GasGiants is 0 — Book 3 p.24's "If Satellite and No Giants,
// place a BigWorld in MW Orbit"), or — if mainworld is an Asteroid Belt —
// via the Belt placement roll instead of HZ+Var. Merges the
// newly-derivable orbit-dependent trade codes (DeriveOrbitTradeCodes)
// into the mainworld's own copy. Returns the updated orbits and the
// index of the mainworld's own Orbit entry within it — when that entry's
// Satellite is true, the immediately preceding orbits entry is its host
// Gas Giant (see the two-Orbit-append below).
func placeMainworld(r *dice.Roller, orbits []Orbit, primary Star, mainworld world.World) ([]Orbit, int) {
	hzOrbit := primary.HabitableZoneOrbit
	mw := mainworld

	var (
		orbitNumber int
		kind        = mainworldPlanet
	)

	if slices.Contains(mw.TradeCodes, world.AsteroidBelt) {
		// "If the Mainworld is an Asteroid Belt, it is placed using the
		// Belt Column of the Basic Placement Chart without regard to
		// Habitable Zone" — skips Table 2B's HZ+Var roll entirely.
		orbitNumber = hzOrbit + rollBeltOffset(r)
	} else {
		dm := 0

		switch primary.SpectralType { //nolint:exhaustive // only M/O/B carry a DM (Table 2B); everything else is +0
		case SpectralM:
			dm = 2
		case SpectralO, SpectralB:
			dm = -2
		}

		orbitNumber = hzOrbit + mainworldHZVar(r.Flux()+dm)
		kind = rollMainworldPlacementKind(r.Flux())
	}

	// "If Satellite and No Giants, place a BigWorld in MW Orbit" (Book 3
	// p.24) — the system's own rolled Gas Giant count can be 0 even when
	// Table 2C's own roll says this mainworld orbits one; when both are
	// true, regenerate the mainworld as a BigWorld and place it as an
	// ordinary planet instead of manufacturing a Gas Giant that would
	// contradict PBG. Only the fields GenerateWithSize itself computes
	// are overwritten — Name/Sector/Hex/Nobility/Allegiance/Worlds/Notes/
	// Ring (left zero by Generate, but not necessarily by an arbitrary
	// caller of GenerateSystem) are preserved rather than wholesale-
	// replaced. An Asteroid Belt mainworld can never reach here — kind
	// only gets set to a satellite kind in the non-belt branch above.
	if kind != mainworldPlanet && mw.PBG.GasGiants == 0 {
		bigWorld := world.GenerateWithSize(r, rollBigWorldSize)
		mw.UWP = bigWorld.UWP
		mw.TradeCodes = bigWorld.TradeCodes
		mw.TravelZone = bigWorld.TravelZone
		mw.Bases = bigWorld.Bases
		mw.PBG = bigWorld.PBG
		mw.Importance = bigWorld.Importance
		mw.Economic = bigWorld.Economic
		mw.Cultural = bigWorld.Cultural
		kind = mainworldPlanet
	}

	// HZVar/the Belt roll can both go negative enough to land below orbit
	// 0 (e.g. an M-type primary with hzOrbit=0 and a negative HZVar) —
	// floored here since orbit 0 is the innermost real orbit, and a
	// negative number would otherwise collide with primaryOrbitNumber's
	// own sentinel value and fall outside orbitAUTable's range.
	orbitNumber = max(orbitNumber, 0)

	// The computed number can independently coincide with a Close/Near/Far
	// star's own orbit, or fall inside the Primary's own photosphere
	// (both are computed separately, with nothing ruling out a match) —
	// nudge via the same collision/preclusion handling placeInOrbit gives
	// every other placement. If nothing in range is free (practically
	// impossible), keep the original number rather than leave the
	// mainworld unplaced.
	primaryHost := starHost{
		role: Primary, hzOrbit: hzOrbit, minOrbit: precludedOrbitHost(primary), maxOrbit: primaryMaxOrbit,
	}
	if n, ok := placeInOrbit(orbits, primaryHost, orbitNumber); ok {
		orbitNumber = n
	}

	mw.TradeCodes = append(mw.TradeCodes, world.DeriveOrbitTradeCodes(mw.UWP, orbitNumber, hzOrbit, true)...)

	if kind != mainworldPlanet {
		gg := rollGasGiant(r)
		orbits = append(
			orbits,
			Orbit{
				Number:      orbitNumber,
				AU:          orbitAU(orbitNumber),
				HostHZOrbit: hzOrbit,
				HostRole:    Primary,
				GasGiant:    &gg,
			},
		)
		orbits = append(orbits, Orbit{
			Number: orbitNumber, Satellite: true, Close: kind == mainworldCloseSatellite, World: &mw,
		})
	} else {
		orbits = append(
			orbits,
			Orbit{Number: orbitNumber, AU: orbitAU(orbitNumber), HostHZOrbit: hzOrbit, HostRole: Primary, World: &mw},
		)
	}

	return orbits, len(orbits) - 1
}
