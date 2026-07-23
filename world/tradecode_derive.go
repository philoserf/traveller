package world

import (
	"slices"

	"github.com/philoserf/traveller/ehex"
)

// tradeCodeTrigger is one row of the T5 Trade Classifications table: a
// trade code plus its UWP-digit predicate. A nil field means "any value"
// (the rulebook's "--"); a non-nil field is an OR-set of allowed digits.
// A world matches a trigger when every non-nil field's value is in its set.
type tradeCodeTrigger struct {
	Code          TradeCode
	Size          []ehex.Value
	Atmosphere    []ehex.Value
	Hydrographics []ehex.Value
	Population    []ehex.Value
	Government    []ehex.Value
	Law           []ehex.Value
	TechLevel     []ehex.Value
}

// nonZeroTechLevels is every valid nonzero ehex digit (1..Max). Unlike the
// other trigger fields, which cap at 15 by game rule (Population,
// Government, etc.), TechLevel is open-ended, so Dieback's "TechLevel>=1"
// condition needs the full range rather than a hand-authored literal list —
// hardcoding it to 1..15 would silently stop matching at TL 16 and up.
var nonZeroTechLevels = func() []ehex.Value {
	vs := make([]ehex.Value, 0, int(ehex.Max))
	for v := ehex.Value(1); v <= ehex.Max; v++ {
		vs = append(vs, v)
	}

	return vs
}()

// tradeCodeTriggers is the set of trade codes derivable purely from a
// world's UWP digits, transcribed from the T5 Book 3 Trade Classifications
// table.
//
// Deliberately excluded, and not represented here:
//   - Structural, not a UWP-digit predicate: Satellite (Sa), Locked (Lk).
//   - Referee-assigned in the rulebook's own text: MilitaryRule (Mr),
//     SubsectorCapital (Cp), SectorCapital (Cs), Capital (Cx), Colony
//     (Cy), DataRepository (Ab), AncientSite (An).
//   - Explicitly non-mainworld ("Not MW") in the rulebook: Mining (Mi),
//     PenalColony (Pe).
//
// Frozen (Fr), Hot (Ho), Cold (Co), Tropic (Tr), Tundra (Tu),
// TwilightZone (Tz), and Farming (Fa) are absent for the same reason as
// Forbidden/Puzzle/Dangerous below: they depend on orbit/Habitable-Zone
// position, which a standalone UWP doesn't carry — see
// DeriveOrbitTradeCodes (world/orbit_tradecode.go) instead, called once
// GenerateSystem has placed a world in an orbit.
//
// Forbidden (Fo), Puzzle (Pz), and Dangerous (Da) are never trigger-table
// rows either, but for a different reason: they ARE derivable (Book 3
// p.28's "Z Travel Zones" step gives a concrete Population/Government/Law
// rule — see computeTravelZone and travelZoneTradeCode), just not via
// this table. Their real predicate is a population threshold OR'd with a
// two-field sum threshold, which doesn't fit tradeCodeTrigger's
// per-field AND-of-sets shape. An earlier version of this comment claimed
// all three were purely referee-assigned, citing the rulebook's flavor
// text (Book 3 line 877) — that was written without having found the
// p.28 generation-step table, which gives an actual mechanic.
//
// Dieback (Di) was excluded for a while: its Pop/Gov/Law columns are
// identical to Barren's (Ba) in the printed table, with only a trailing
// "(000-T)" annotation Barren lacks. Read directly against the PDF page
// (not just the pdftotext extraction, ruling out a column-misalignment
// artifact), "(000-T)" visually completes the UWP format string's
// "...PGL-T" tail with T left as a variable rather than pinned to 0 —
// and the book's own Native Intelligent Life table draws exactly this
// distinction elsewhere for Pop=0 worlds: "Extinct Natives" (TL=0) vs.
// "Catastrophic Extinct Natives" (TL=1+, ruins/evidence of past
// civilization) are mutually exclusive categories, since a world has
// exactly one TL. Neither source states the equivalence outright, but
// both point the same way, so Barren and Dieback share the Pop=0/Gov=0/
// Law=0 condition and split on TechLevel (Barren: 0, Dieback: 1+) —
// never both, matching the NIL table's own mutual exclusivity — rather
// than a competing guess at which of two contradictory rulebook
// statements to trust (contrast the Extensions ambiguities this project
// leaves unresolved, where the rulebook's own text disagrees with
// itself and there's no way to pick a reading without guessing).
var tradeCodeTriggers = []tradeCodeTrigger{
	{Code: AsteroidBelt, Size: []ehex.Value{0}, Atmosphere: []ehex.Value{0}, Hydrographics: []ehex.Value{0}},
	{Code: Desert, Atmosphere: []ehex.Value{2, 3, 4, 5, 6, 7, 8, 9}, Hydrographics: []ehex.Value{0}},
	{
		Code: Fluid, Atmosphere: []ehex.Value{10, 11, 12}, // A, B, C
		Hydrographics: []ehex.Value{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	},
	{
		Code:          Garden,
		Size:          []ehex.Value{6, 7, 8},
		Atmosphere:    []ehex.Value{5, 6, 8},
		Hydrographics: []ehex.Value{5, 6, 7},
	},
	{
		Code: Hellworld, Size: []ehex.Value{3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		Atmosphere: []ehex.Value{2, 4, 7, 9, 10, 11, 12}, Hydrographics: []ehex.Value{0, 1, 2},
	},
	{Code: IceCapped, Atmosphere: []ehex.Value{0, 1}, Hydrographics: []ehex.Value{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
	{
		Code: Ocean, Size: []ehex.Value{10, 11, 12, 13, 14, 15}, // A-F
		Atmosphere: []ehex.Value{3, 4, 5, 6, 7, 8, 9, 13, 14, 15}, Hydrographics: []ehex.Value{10}, // D, E, F / A
	},
	{Code: Vacuum, Atmosphere: []ehex.Value{0}},
	{
		Code: WaterWorld, Size: []ehex.Value{3, 4, 5, 6, 7, 8, 9},
		Atmosphere: []ehex.Value{3, 4, 5, 6, 7, 8, 9, 13, 14, 15}, Hydrographics: []ehex.Value{10},
	},

	{
		Code: Barren, Population: []ehex.Value{0}, Government: []ehex.Value{0}, Law: []ehex.Value{0},
		TechLevel: []ehex.Value{0},
	},
	{
		Code: Dieback, Population: []ehex.Value{0}, Government: []ehex.Value{0}, Law: []ehex.Value{0},
		TechLevel: nonZeroTechLevels,
	},
	{Code: LowPopulation, Population: []ehex.Value{1, 2, 3}},
	{Code: NonIndustrial, Population: []ehex.Value{4, 5, 6}},
	{Code: PreHigh, Population: []ehex.Value{8}},
	{Code: HighPopulation, Population: []ehex.Value{9, 10, 11, 12, 13, 14, 15}}, // 9, A-F
	{
		Code: PreAgricultural, Atmosphere: []ehex.Value{4, 5, 6, 7, 8, 9},
		Hydrographics: []ehex.Value{4, 5, 6, 7, 8}, Population: []ehex.Value{4, 8},
	},
	{
		Code: Agricultural, Atmosphere: []ehex.Value{4, 5, 6, 7, 8, 9},
		Hydrographics: []ehex.Value{4, 5, 6, 7, 8}, Population: []ehex.Value{5, 6, 7},
	},
	{
		Code: NonAgricultural, Atmosphere: []ehex.Value{0, 1, 2, 3}, Hydrographics: []ehex.Value{0, 1, 2, 3},
		Population: []ehex.Value{6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
	},

	{
		Code: PrisonExileCamp, Atmosphere: []ehex.Value{2, 3, 10, 11}, // 2, 3, A, B
		Hydrographics: []ehex.Value{1, 2, 3, 4, 5}, Population: []ehex.Value{3, 4, 5, 6}, Law: []ehex.Value{6, 7, 8, 9},
	},
	{Code: PreIndustrial, Atmosphere: []ehex.Value{0, 1, 2, 4, 7, 9}, Population: []ehex.Value{7, 8}},
	{
		Code: Industrial, Atmosphere: []ehex.Value{0, 1, 2, 4, 7, 9, 10, 11, 12}, // ..., A, B, C
		Population: []ehex.Value{9, 10, 11, 12, 13, 14, 15},
	},
	{Code: Poor, Atmosphere: []ehex.Value{2, 3, 4, 5}, Hydrographics: []ehex.Value{0, 1, 2, 3}},
	{Code: PreRich, Atmosphere: []ehex.Value{6, 8}, Population: []ehex.Value{5, 9}},
	{Code: Rich, Atmosphere: []ehex.Value{6, 8}, Population: []ehex.Value{6, 7, 8}},

	{Code: Reserve, Population: []ehex.Value{0, 1, 2, 3, 4}, Government: []ehex.Value{6}, Law: []ehex.Value{0, 4, 5}},
}

func matchesAny(v ehex.Value, allowed []ehex.Value) bool {
	return allowed == nil || slices.Contains(allowed, v)
}

func (t tradeCodeTrigger) matches(u UWP) bool {
	return matchesAny(u.Size, t.Size) &&
		matchesAny(u.Atmosphere, t.Atmosphere) &&
		matchesAny(u.Hydrographics, t.Hydrographics) &&
		matchesAny(u.Population, t.Population) &&
		matchesAny(u.Government, t.Government) &&
		matchesAny(u.Law, t.Law) &&
		matchesAny(u.TechLevel, t.TechLevel)
}

// travelZoneTradeCode returns the trade code sub-label for a precomputed
// Travel Zone, and whether one applies (Green has none). Takes the zone
// (and the population that helped compute it) rather than a UWP, so a
// caller that's already computed the zone doesn't redo that work — see
// deriveTradeCodesForZone. This is a separate mechanism from
// tradeCodeTriggers: Da/Pz/Fo's real predicate (a population threshold
// OR'd with a two-field sum threshold) isn't expressible as
// tradeCodeTrigger's per-field AND-of-sets shape.
func travelZoneTradeCode(zone TravelZone, population ehex.Value) (TradeCode, bool) {
	switch zone {
	case ZoneRed:
		return Forbidden, true
	case ZoneAmber:
		if int(population) <= 6 {
			return Dangerous, true
		}

		return Puzzle, true
	default:
		return "", false
	}
}

// deriveTradeCodesForZone is DeriveTradeCodes' implementation, taking an
// already-computed TravelZone as input so Generate can share one
// computeTravelZone call across both TradeCodes and World.TravelZone
// instead of computing it twice per generated world.
func deriveTradeCodesForZone(u UWP, zone TravelZone) []TradeCode {
	var codes []TradeCode

	for _, t := range tradeCodeTriggers {
		if t.matches(u) {
			codes = append(codes, t.Code)
		}
	}

	if code, ok := travelZoneTradeCode(zone, u.Population); ok {
		codes = append(codes, code)
	}

	return codes
}

// DeriveTradeCodes returns every trade code derivable purely from u's UWP
// digits. See tradeCodeTriggers' doc comment for what's deliberately
// excluded, and why.
func DeriveTradeCodes(u UWP) []TradeCode {
	return deriveTradeCodesForZone(u, computeTravelZone(u))
}
