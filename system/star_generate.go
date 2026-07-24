package system

import "github.com/philoserf/traveller/dice"

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

	size := rollLuminosityClass(sizeFlux, t, decimal)

	star := Star{SpectralType: t, SpectralDecimal: decimal, LuminosityClass: size}

	// ok is false only for a genuine SpectralDegenerate t (no O-M row in
	// habitableZoneTable at all — see habitableZoneOrbit's own doc
	// comment); HabitableZoneOrbit is left at its zero value for that
	// case rather than guessing at a fallback the book doesn't give.
	if hz, ok := habitableZoneOrbit(t, size); ok {
		star.HabitableZoneOrbit = hz
	}

	return star
}

// attachCompanion rolls a Flux; if it meets starPresenceFlux, rolls and
// returns a Companion star sharing role, per Table 1's "Flux for
// Companions for each Star present." Shared by every star-rolling site
// (Primary, and each Close/Near/Far via rollAndPlaceStar) so this rule
// has exactly one implementation.
func attachCompanion(r *dice.Roller, primaryFlux int, role StellarRole) *Star {
	if r.Flux() < starPresenceFlux {
		return nil
	}

	companion := rollStar(r, primaryFlux, false)
	companion.Role = role

	return &companion
}

// rollAndPlaceStar rolls a Close/Near/Far star (and, per attachCompanion,
// its own optional Companion), returning it as an Orbit at orbitNumber.
func rollAndPlaceStar(r *dice.Roller, primaryFlux int, role StellarRole, orbitNumber int) Orbit {
	star := rollStar(r, primaryFlux, false)
	star.Role = role
	star.Companion = attachCompanion(r, primaryFlux, role)

	return Orbit{Number: orbitNumber, AU: orbitAU(orbitNumber), Star: &star}
}
