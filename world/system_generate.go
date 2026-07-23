package world

import (
	"slices"

	"github.com/philoserf/traveller/dice"
)

// primaryOrbitNumber is the sentinel Orbit.Number for the Primary star
// itself — "Primary = at center of System" (Book 3 p.28), not a numbered
// planetary orbit around anything. Distinct from every real orbit number
// (0-20 per orbitAUTable), which are all non-negative.
const primaryOrbitNumber = -1

// rollStar rolls a star's SpectralType, SpectralDecimal, LuminosityClass,
// and HabitableZoneOrbit. isPrimary=true rolls two fresh Flux values (one
// for type, one for size) — Book 3 p.28 Table 2: "Roll Flux for Primary."
// isPrimary=false derives both from primaryFlux (the Primary's own type
// roll) instead: "For all others, Primary Flux + (1D-1)" for type,
// "Primary Flux + (1D+2)" for size — this applies to Close/Near/Far stars
// and to any Companion, not just companions specifically ("for all
// others" reads as everyone but the Primary).
func rollStar(r *dice.Roller, primaryFlux int, isPrimary bool) Star {
	typeFlux := primaryFlux
	sizeFlux := r.Flux()

	if !isPrimary {
		typeFlux = primaryFlux + (r.D6() - 1)
		sizeFlux = primaryFlux + (r.D6() + 2)
	}

	t := rollSpectralType(r, typeFlux)

	var decimal int
	if t != SpectralDegenerate {
		decimal = r.Uniform(10) - 1
	}

	size := rollLuminosityClass(sizeFlux, t)

	star := Star{SpectralType: t, SpectralDecimal: decimal, LuminosityClass: size}

	// Believed unreachable given how rollLuminosityClass and
	// habitableZoneTable interlock (see habitableZoneOrbit's doc
	// comment) — HabitableZoneOrbit is left at its zero value rather
	// than guessing at a fallback for a combination this project has no
	// data for.
	if hz, ok := habitableZoneOrbit(t, size); ok {
		star.HabitableZoneOrbit = hz
	}

	return star
}

// rollAndPlaceStar rolls a Close/Near/Far star (and, per starPresenceFlux,
// its own optional Companion), returning it as an Orbit at orbitNumber.
func rollAndPlaceStar(r *dice.Roller, primaryFlux int, role StellarRole, orbitNumber int) Orbit {
	star := rollStar(r, primaryFlux, false)
	star.Role = role

	if r.Flux() >= starPresenceFlux {
		companion := rollStar(r, primaryFlux, false)
		companion.Role = role
		star.Companion = &companion
	}

	return Orbit{Number: orbitNumber, AU: orbitAU(orbitNumber), Star: &star}
}

// GenerateSystem builds a StarSystem around an already-generated
// mainworld (from Generate): rolls the Primary star (and, independently,
// whether a Close/Near/Far star and any Companions exist), computes the
// Primary's HZ orbit, places the mainworld (as a Planet, as a Satellite
// of a freshly rolled Gas Giant, or — if the mainworld is itself an
// Asteroid Belt — via the Belt placement roll instead of HZ+Var), and
// merges the newly-derivable orbit-dependent trade codes
// (DeriveOrbitTradeCodes) into the mainworld's TradeCodes.
//
// Only the Primary hosts the mainworld. Placing every other body in the
// system (additional gas giants/belts/secondary worlds beyond what
// mainworld placement itself needs, satellites for any of them, rings,
// and precluded-orbit adjustment for oversized stars) is deliberately out
// of scope — see the sysgen plan/issue #3 for why.
func GenerateSystem(r *dice.Roller, mainworld World) StarSystem {
	primaryFlux := r.Flux()
	primary := rollStar(r, primaryFlux, true)
	primary.Role = Primary

	if r.Flux() >= starPresenceFlux {
		companion := rollStar(r, primaryFlux, false)
		companion.Role = Primary
		primary.Companion = &companion
	}

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

	hzOrbit := primary.HabitableZoneOrbit

	mw := mainworld

	var (
		mainworldOrbitNumber int
		satelliteOfGasGiant  bool
	)

	if slices.Contains(mw.TradeCodes, AsteroidBelt) {
		// "If the Mainworld is an Asteroid Belt, it is placed using the
		// Belt Column of the Basic Placement Chart without regard to
		// Habitable Zone" — skips Table 2B's HZ+Var roll entirely.
		mainworldOrbitNumber = hzOrbit + rollBeltOffset(r)
	} else {
		dm := 0

		switch primary.SpectralType { //nolint:exhaustive // only M/O/B carry a DM (Table 2B); everything else is +0
		case SpectralM:
			dm = 2
		case SpectralO, SpectralB:
			dm = -2
		}

		mainworldOrbitNumber = hzOrbit + mainworldHZVar(r.Flux()+dm)

		if kind := rollMainworldPlacementKind(
			r.Flux(),
		); kind == mainworldCloseSatellite ||
			kind == mainworldFarSatellite {
			satelliteOfGasGiant = true
		}
	}

	mw.TradeCodes = append(mw.TradeCodes, DeriveOrbitTradeCodes(mw.UWP, mainworldOrbitNumber, hzOrbit, true)...)

	if satelliteOfGasGiant {
		gg := rollGasGiant(r)
		orbits = append(orbits, Orbit{Number: mainworldOrbitNumber, AU: orbitAU(mainworldOrbitNumber), GasGiant: &gg})
		orbits = append(orbits, Orbit{Number: mainworldOrbitNumber, Satellite: true, World: &mw})
	} else {
		orbits = append(orbits, Orbit{Number: mainworldOrbitNumber, AU: orbitAU(mainworldOrbitNumber), World: &mw})
	}

	return StarSystem{
		Orbits:         orbits,
		MainworldOrbit: len(orbits) - 1,
	}
}
