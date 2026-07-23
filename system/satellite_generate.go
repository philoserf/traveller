package system

import (
	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
	"github.com/philoserf/traveller/world"
)

// satelliteHostKind is which of Book 3 p.29's four "S Number of
// Satellites" formulas applies to a body: fixed for a Gas Giant, or by
// its own delta-from-HZ band for a World. A 3-way band (Inner/HZ/Outer),
// distinct from the 2-way band (delta<=0 vs delta>0)
// rollSecondaryWorldCategory uses for a body's own type — the count
// formula's three labels don't collapse to two the way the type table's
// columns do.
type satelliteHostKind int

// satelliteHostKind values.
const (
	hostGasGiant satelliteHostKind = iota
	hostInnerWorld
	hostHZWorld
	hostOuterWorld
)

// satelliteHostKindFor resolves the band a body falls into: fixed for a
// Gas Giant, otherwise by delta ("Inner= inside HZ-1" is delta<=-2,
// "HZ= HZ-1, HZ, HZ+1" is delta in -1..1, "Outer= beyond HZ+1" is
// delta>=2 — Book 3 p.29).
func satelliteHostKindFor(isGasGiant bool, delta int) satelliteHostKind {
	switch {
	case isGasGiant:
		return hostGasGiant
	case delta <= -2:
		return hostInnerWorld
	case delta >= 2:
		return hostOuterWorld
	default:
		return hostHZWorld
	}
}

// rollSatelliteCount rolls Book 3 p.29's "S" formula for kind:
//
//	Gas Giants=  1D-1     Inners=  1D-5
//	Hospitables= 1D-4     Outers=  1D-3
//
// "Zero=Ring and reroll" — a raw roll of exactly 0 sets ring=true and
// rerolls once more (the rule says "reroll," not "reroll until
// nonzero"); "Less than 0=none" — a negative raw roll (on either the
// first roll or the single reroll) floors to 0 satellites.
func rollSatelliteCount(r *dice.Roller, kind satelliteHostKind) (int, bool) {
	roll := func() int {
		switch kind {
		case hostGasGiant:
			return r.D6() - 1
		case hostInnerWorld:
			return r.D6() - 5
		case hostOuterWorld:
			return r.D6() - 3
		default: // hostHZWorld
			return r.D6() - 4
		}
	}

	n := roll()

	ring := false
	if n == 0 {
		ring = true
		n = roll()
	}

	return max(n, 0), ring
}

// satelliteOuterCategoryByRoll is Book 3 p.29's "Outer Satellites"
// column — identical to outerCategoryByRoll (system_tables.go) except
// row 4: categoryStormWorld here, categoryIceworld there. A real,
// book-verified difference between satellite and regular Other-World
// placement, not a transcription error.
var satelliteOuterCategoryByRoll = map[int]secondaryWorldCategory{
	1: categoryWorldlet, 2: categoryIceworld, 3: categoryBigWorld,
	4: categoryStormWorld, 5: categoryRadWorld, 6: categoryIceworld,
}

// rollSatelliteCategory rolls a satellite's category. Reuses the same
// Inner/HZ column (innerHZCategoryByRoll) rollSecondaryWorldCategory
// does — that column is identical for satellites and regular
// placement — but its own Outer column (satelliteOuterCategoryByRoll)
// for delta>0.
func rollSatelliteCategory(r *dice.Roller, delta int) secondaryWorldCategory {
	if delta <= 0 {
		return innerHZCategoryByRoll[r.D6()]
	}

	return satelliteOuterCategoryByRoll[r.D6()]
}

// generateSatellite rolls one satellite's UWP and applies Book 3 p.21's
// size rule: "A satellite is always smaller than its parent; if its size
// is generated as larger than the parent, adjust to fit." Only applied
// when hasParentSize is true (the parent is a World — its Size is
// directly comparable to the satellite's own, unlike a GasGiant parent's
// Size, which is a letter on a different physical scale entirely) and
// parentSize > 0 — for a Size-0 parent there's no strictly-smaller ehex
// digit to adjust down to, so the roll is left as-is rather than being
// silently forced to 0 (which would misrepresent a clamp as the "sizes
// are equal, the result is a double planet" case — left as-is, no
// adjustment and no dedicated flag — when it's really just "couldn't
// adjust").
func generateSatellite(
	r *dice.Roller,
	delta int,
	parentSize ehex.Value,
	hasParentSize bool,
	maxPopulation ehex.Value,
) world.UWP {
	category := rollSatelliteCategory(r, delta)
	u := generateSecondaryWorldUWP(r, category, maxPopulation)

	if hasParentSize && parentSize > 0 && u.Size > parentSize {
		u.Size = world.ClampEhex(int(parentSize)-1, 0, int(ehex.Max))
	}

	return u
}

// generateSatellitesForBody rolls and appends satellites (and a Ring, if
// rolled) for one top-level body — a World or a GasGiant — as new Orbit
// entries sharing parent.Number with Satellite:true. Each satellite
// independently rolls Close (2D<=7, tidally locked) or Far (2D>=8) —
// Book 3 p.21/24; this project doesn't model the book's letter-named
// sub-orbit slots (Ay/Bee/.../Zee) as real data, the same deliberate
// simplification made for the mainworld's own satellite case in #3.
func generateSatellitesForBody(r *dice.Roller, orbits *[]Orbit, parent Orbit, hzOrbit int, maxPopulation ehex.Value) {
	delta := parent.Number - hzOrbit
	isGasGiant := parent.GasGiant != nil
	kind := satelliteHostKindFor(isGasGiant, delta)

	count, ring := rollSatelliteCount(r, kind)

	if ring {
		if isGasGiant {
			parent.GasGiant.Ring = true
		} else {
			parent.World.Ring = true
		}
	}

	var parentSize ehex.Value

	hasParentSize := !isGasGiant
	if hasParentSize {
		parentSize = parent.World.UWP.Size
	}

	for range count {
		u := generateSatellite(r, delta, parentSize, hasParentSize, maxPopulation)
		w := worldWithTradeCodes(u, parent.Number, hzOrbit)

		*orbits = append(*orbits, Orbit{
			Number:    parent.Number,
			Satellite: true,
			Close:     r.TwoD6() <= 7,
			World:     &w,
		})
	}
}
