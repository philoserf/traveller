package world

import (
	"slices"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// importanceBonusTradeCodes are the trade codes that each add +1 to Ix.
// Package-level rather than a literal inside computeImportance so it isn't
// reallocated on every call — computeImportance runs once per Generate(),
// which runs once per /worlds/random request.
var importanceBonusTradeCodes = []TradeCode{Agricultural, PreAgricultural, HighPopulation, Industrial, Rich}

// computeImportance computes a world's Importance (Ix): a sum of modifiers
// over its already-generated UWP, TradeCodes, and Bases. Dice-free — the
// rulebook's own Ix formula has no roll in it, just a sum of conditions —
// so it's directly testable against fixed fixtures.
//
// The formula table states the per-trade-code bonus as "Ag, Hi, In, Ri",
// but the rulebook's own worked "Regina" example applies it to
// Pre-Agricultural (Pa) instead of Agricultural (Ag) as literally written.
// Modeled here as "Ag" covering both Agricultural and Pre-Agricultural,
// matching the example rather than a strict literal-only reading.
//
// Computed after Bases so the Naval+Scout and Way Station bonuses are
// included — the book's own worked example computes Ix in an earlier step
// than it rolls Bases, so its stated Ix=+4 for Regina does NOT include
// the bonus for her actual Naval+Scout bases. That's a procedural quirk
// in the source, not a formula difference; this implementation's ordering
// is a deliberate correction, not a replication of the quirk (see
// TestComputeImportanceMatchesRegina, which pins both readings).
func computeImportance(u UWP, tradeCodes []TradeCode, bases []Base) Importance {
	ix := 0

	switch u.Starport {
	case StarportA, StarportB:
		ix++
	case StarportD, StarportE, StarportNone:
		ix--
	case StarportC:
		// no modifier
	}

	if u.TechLevel >= 10 { // A
		ix++
	}

	if u.TechLevel >= 16 { // G
		ix++
	}

	if u.TechLevel <= 8 {
		ix--
	}

	for _, tc := range tradeCodes {
		if slices.Contains(importanceBonusTradeCodes, tc) {
			ix++
		}
	}

	if u.Population <= 6 {
		ix--
	}

	if slices.Contains(bases, NavalBase) && slices.Contains(bases, ScoutBase) {
		ix++
	}

	if slices.Contains(bases, WayStation) {
		ix++
	}

	return Importance(ix)
}

// computeLabor: Population-1, floored at 0 (Population=0 would otherwise
// go negative).
func computeLabor(population ehex.Value) int {
	return max(int(population)-1, 0)
}

// resourcesFromRoll combines a 2D6 roll with the TechLevel>=8 bonus (the
// system's Gas Giant and Planetoid Belt counts, from PBG). Split from the
// roll itself so the formula is unit-testable against a fixed roll value,
// matching techLevelModifier's pattern elsewhere in this package.
func resourcesFromRoll(roll int, techLevel ehex.Value, pbg PBG) int {
	if techLevel >= 8 {
		roll += int(pbg.GasGiants) + int(pbg.Belts)
	}

	return max(roll, 0)
}

// infrastructureFromRoll combines a population-gated die roll (0, 1D6, or
// 2D6 — chosen by the caller per the Population bracket) with Importance.
func infrastructureFromRoll(roll int, ix Importance) int {
	return max(roll+int(ix), 0)
}

// rollEconomic rolls a world's Economic (Ex) extension: Resources, Labor,
// Infrastructure, Efficiency. Requires Importance (already computed) and
// PBG (for the Resources bonus).
func rollEconomic(r *dice.Roller, u UWP, ix Importance, pbg PBG) Economic {
	resources := resourcesFromRoll(r.TwoD6(), u.TechLevel, pbg)

	var infrastructure int

	switch {
	case u.Population == 0:
		infrastructure = 0
	case u.Population <= 3:
		infrastructure = infrastructureFromRoll(0, ix)
	case u.Population <= 6:
		infrastructure = infrastructureFromRoll(r.D6(), ix)
	default:
		infrastructure = infrastructureFromRoll(r.TwoD6(), ix)
	}

	return Economic{
		Resources:      resources,
		Labor:          computeLabor(u.Population),
		Infrastructure: infrastructure,
		Efficiency:     r.Flux(),
	}
}

// computeHeterogeneity, computeAcceptance, computeStrangeness, and
// computeSymbols are Cultural's four components, split from their dice
// rolls for the same reason as the Economic helpers above. All floor at 1,
// per the rulebook — rollCultural handles the Population=0 override (all
// four values 0 instead) separately, since it applies before this floor.
//
// Strangeness is defined as Flux+5, but the rulebook's own worked example
// instead computes it as "2D-2". This isn't a discrepancy: Flux+5 and
// 2D6-2 are mathematically identical probability distributions (both
// reduce to 2D6-7+5 in law, since a fair die's complement 7-roll is
// itself uniform on 1-6) — verified computationally, not just argued.
// Implemented as Flux+5 to match the formula table's phrasing directly.
func computeHeterogeneity(population ehex.Value, flux int) int {
	return max(int(population)+flux, 1)
}

func computeAcceptance(population ehex.Value, ix Importance) int {
	return max(int(population)+int(ix), 1)
}

func computeStrangeness(flux int) int {
	return max(flux+5, 1)
}

func computeSymbols(techLevel ehex.Value, flux int) int {
	return max(flux+int(techLevel), 1)
}

// rollCultural rolls a world's Cultural (Cx) extension: Heterogeneity,
// Acceptance, Strangeness, Symbols. Requires Importance (already
// computed). "Population digit" in this formula means the UWP Population
// field itself, not the separately-rolled PBG.PopulationDigit — confirmed
// against the rulebook's own worked Regina example, which uses Regina's
// UWP Population (8), not her PBG.PopulationDigit (7).
func rollCultural(r *dice.Roller, u UWP, ix Importance) Cultural {
	if u.Population == 0 {
		return Cultural{}
	}

	return Cultural{
		Heterogeneity: computeHeterogeneity(u.Population, r.Flux()),
		Acceptance:    computeAcceptance(u.Population, ix),
		Strangeness:   computeStrangeness(r.Flux()),
		Symbols:       computeSymbols(u.TechLevel, r.Flux()),
	}
}
