package world

import (
	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// lawCeiling is the highest Law value: the letter "J" (T5 skips "I", so J's
// numeric value in the ehex alphabet is 18, not the naively-expected 19).
// Computed via Parse rather than hardcoded so the discrepancy can't bite.
var lawCeiling = func() ehex.Value {
	v, err := ehex.Parse("J")
	if err != nil {
		panic(err) // the literal "J" is always valid; a panic here means ehex itself is broken
	}

	return v
}()

// clampEhex clamps v to [lo,hi] and converts it to ehex.Value. lo and hi
// must themselves be valid ehex digits (0..ehex.Max) — the switch below is
// what proves the conversion below never overflows uint8, though gosec
// can't verify that itself (it doesn't reason about a callee's own bounds
// checks, only literal patterns in the same function), hence the nolint.
//
//nolint:gosec // v is provably in [lo,hi] by the switch below; callers pass ehex-valid lo/hi
func clampEhex(v, lo, hi int) ehex.Value {
	switch {
	case v < lo:
		return ehex.Value(lo)
	case v > hi:
		return ehex.Value(hi)
	default:
		return ehex.Value(v)
	}
}

func rollStarport(r *dice.Roller) Starport {
	switch v := r.TwoD6(); {
	case v <= 4:
		return StarportA
	case v <= 6:
		return StarportB
	case v <= 8:
		return StarportC
	case v == 9:
		return StarportD
	case v <= 11:
		return StarportE
	default:
		return StarportNone
	}
}

// rollSize: 2D6-2, range 0-10. A raw 10 rerolls as 1D6+9 (range 10-15),
// extending the world into the A-F "large world" band.
func rollSize(r *dice.Roller) ehex.Value {
	v := r.TwoD6() - 2
	if v == 10 {
		v = r.D6() + 9
	}

	return clampEhex(v, 0, int(ehex.Max)) // mathematically 0-15 already; defensive
}

// rollAtmosphere: Flux+Size, forced to 0 if Size=0, clamped to 0..15(F).
func rollAtmosphere(r *dice.Roller, size ehex.Value) ehex.Value {
	if size == 0 {
		return 0
	}

	v := r.Flux() + int(size)

	return clampEhex(v, 0, 15) // 15 = F
}

// rollHydrographics: Flux+Atmosphere-4(if Atm out of 2..9), forced to 0 if
// Size<2, clamped to 0..10(A).
func rollHydrographics(r *dice.Roller, size, atm ehex.Value) ehex.Value {
	if size < 2 {
		return 0
	}

	mod := 0
	if atm < 2 || atm > 9 {
		mod = -4
	}

	v := r.Flux() + int(atm) + mod

	return clampEhex(v, 0, 10) // 10 = A
}

// rollPopulation: 2D6-2, independent of prior fields. A raw 10 rerolls as
// 2D6+3 (range 5-15), extending into the very-populous band.
func rollPopulation(r *dice.Roller) ehex.Value {
	v := r.TwoD6() - 2
	if v == 10 {
		v = r.TwoD6() + 3
	}

	return clampEhex(v, 0, int(ehex.Max)) // mathematically 0-15 already; defensive
}

// rollGovernment: Flux+Population, clamped to 0..15(F). The rulebook states
// only the ceiling; the floor of 0 is a design choice (Government feeds
// Law, which needs a sane input), not literal rulebook text.
func rollGovernment(r *dice.Roller, pop ehex.Value) ehex.Value {
	v := r.Flux() + int(pop)

	return clampEhex(v, 0, 15) // 15 = F
}

// rollLaw: Flux+Government, clamped to 0..lawCeiling(J). As with
// Government, the floor of 0 is a design choice, not literal rulebook text.
func rollLaw(r *dice.Roller, gov ehex.Value) ehex.Value {
	v := r.Flux() + int(gov)

	return clampEhex(v, 0, int(lawCeiling))
}

func starportTechLevelMod(s Starport) int {
	switch s {
	case StarportA:
		return 6
	case StarportB:
		return 4
	case StarportC:
		return 2
	case StarportNone:
		return -4
	default: // StarportD, StarportE
		return 0
	}
}

func sizeTechLevelMod(s ehex.Value) int {
	switch {
	case s <= 1:
		return 2
	case s <= 4:
		return 1
	default:
		return 0
	}
}

func atmosphereTechLevelMod(a ehex.Value) int {
	switch {
	case a <= 3:
		return 1
	case a <= 9:
		return 0
	default: // 10-15 (A-F)
		return 1
	}
}

func hydrographicsTechLevelMod(h ehex.Value) int {
	switch h {
	case 9:
		return 1
	case 10: // A
		return 2
	default:
		return 0
	}
}

func populationTechLevelMod(p ehex.Value) int {
	switch {
	case p == 0:
		return 0
	case p <= 5:
		return 1
	case p <= 8:
		return 0
	case p == 9:
		return 2
	default: // 10-15 (A-F)
		return 4
	}
}

func governmentTechLevelMod(g ehex.Value) int {
	switch g {
	case 0, 5:
		return 1
	case 13: // D
		return -2
	default:
		return 0
	}
}

// techLevelModifier sums the Tech Level dice modifier table over an
// otherwise-complete UWP (Starport through Government). Kept dice-free so
// the modifier table is unit-testable against fixed fixtures.
func techLevelModifier(u UWP) int {
	return starportTechLevelMod(u.Starport) +
		sizeTechLevelMod(u.Size) +
		atmosphereTechLevelMod(u.Atmosphere) +
		hydrographicsTechLevelMod(u.Hydrographics) +
		populationTechLevelMod(u.Population) +
		governmentTechLevelMod(u.Government)
}

// rollTechLevel: 1D6 + techLevelModifier, rolled last since the modifier
// depends on every other field. The rulebook states no floor or ceiling,
// but the modifier table's worst case (Starport X, Government D, minimum
// roll) computes to -5 — converting that directly to ehex.Value (uint8)
// would silently wrap to an invalid digit, so the floor of 0 here is a
// correctness guard, not a rules interpretation. The ceiling at ehex.Max is
// defensive insurance against future modifier-table changes; the realistic
// max (~22) never reaches it today.
func rollTechLevel(r *dice.Roller, u UWP) ehex.Value {
	v := r.D6() + techLevelModifier(u)

	return clampEhex(v, 0, int(ehex.Max))
}

// navalBaseTarget returns the 2D target number for a Naval Base at the
// given Starport, or false if that Starport can't have one at all.
func navalBaseTarget(s Starport) (int, bool) {
	switch s {
	case StarportA:
		return 6, true
	case StarportB:
		return 5, true
	default:
		return 0, false
	}
}

// scoutBaseTarget returns the 2D target number for a Scout Base at the
// given Starport, or false if that Starport can't have one at all.
func scoutBaseTarget(s Starport) (int, bool) {
	switch s {
	case StarportA:
		return 4, true
	case StarportB:
		return 5, true
	case StarportC:
		return 6, true
	case StarportD:
		return 7, true
	default:
		return 0, false
	}
}

// rollBases rolls Naval and Scout base presence, independent 2D checks
// gated by Starport grade. The rulebook's own worked example (Regina,
// Starport A, rolling 5 for Naval and 3 for Scout) reports both as
// present despite the printed table requiring 6+ and 4+ respectively — an
// inconsistency in the source, not a transcription error here. This
// implementation trusts the table (matching known Naval/Scout base
// mechanics across Traveller editions) over the example.
//
// Naval Depot and Way Station are excluded: both are Starport-A-only
// density/frequency placements ("1 per 1000 worlds", "1 per 50 parsecs on
// a trade route") rather than per-world dice rolls. Military, Scientific,
// Diplomatic, and Cultural bases are excluded too: the rulebook calls them
// out as referee-assigned exceptions with no given mechanic at all.
func rollBases(r *dice.Roller, starport Starport) []Base {
	var bases []Base

	if target, ok := navalBaseTarget(starport); ok && r.TwoD6() >= target {
		bases = append(bases, NavalBase)
	}

	if target, ok := scoutBaseTarget(starport); ok && r.TwoD6() >= target {
		bases = append(bases, ScoutBase)
	}

	return bases
}

// rollPBG rolls a system's PBG: Population digit, Belts, and Gas Giants.
// PopulationDigit is a flavor detail distinct from the UWP Population
// field — a uniform 1-9 roll when Population>0, or 0 when Population=0 —
// not meant to reflect the actual order-of-magnitude population value.
// Belts (1D-3, floor 0, range 0-3) and Gas Giants (2D/2-2, floor 0, range
// 0-4) describe the whole system, not just this world, but PBG lives on
// World per this package's existing type (see world/extensions.go).
func rollPBG(r *dice.Roller, population ehex.Value) PBG {
	var populationDigit ehex.Value
	if population > 0 {
		populationDigit = clampEhex(r.Uniform(9), 0, 9)
	}

	return PBG{
		PopulationDigit: populationDigit,
		Belts:           clampEhex(r.D6()-3, 0, 3),
		GasGiants:       clampEhex(r.TwoD6()/2-2, 0, 4),
	}
}

// GenerateUWP rolls a complete UWP in the order T5 requires: each field may
// depend only on fields already rolled, ending with TechLevel, which
// depends on all the others.
func GenerateUWP(r *dice.Roller) UWP {
	var u UWP

	u.Starport = rollStarport(r)
	u.Size = rollSize(r)
	u.Atmosphere = rollAtmosphere(r, u.Size)
	u.Hydrographics = rollHydrographics(r, u.Size, u.Atmosphere)
	u.Population = rollPopulation(r)
	u.Government = rollGovernment(r, u.Population)
	u.Law = rollLaw(r, u.Government)
	u.TechLevel = rollTechLevel(r, u)

	return u
}

// Generate produces a new World: a rolled UWP, its UWP-derivable trade
// codes, TravelZone, Bases, PBG, and the Importance/Economic/Cultural
// extensions. Name, Sector, Hex, Worlds, and Notes are left zero-valued —
// none of them are generated yet (see DeriveTradeCodes and the world
// package's generation docs for what's deliberately out of scope, and
// why). Nobility and Allegiance are permanently out of scope for
// generation, not just "not yet": both are referee/campaign-assigned in
// T5, with no dice mechanic given for either.
func Generate(r *dice.Roller) World {
	uwp := GenerateUWP(r)
	tradeCodes := DeriveTradeCodes(uwp)
	travelZone := computeTravelZone(uwp)
	bases := rollBases(r, uwp.Starport)
	pbg := rollPBG(r, uwp.Population)
	importance := computeImportance(uwp, tradeCodes, bases)

	return World{
		UWP:        uwp,
		TradeCodes: tradeCodes,
		TravelZone: travelZone,
		Bases:      bases,
		PBG:        pbg,
		Importance: importance,
		Economic:   rollEconomic(r, uwp, importance, pbg),
		Cultural:   rollCultural(r, uwp, importance),
	}
}
