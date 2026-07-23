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
}

// tradeCodeTriggers is the set of trade codes derivable purely from a
// world's UWP digits, transcribed from the T5 Book 3 Trade Classifications
// table.
//
// Deliberately excluded, and not represented here:
//   - Structural, not a UWP-digit predicate: Satellite (Sa), Locked (Lk).
//   - Depends on orbit/Habitable-Zone position, which a standalone UWP
//     doesn't carry: Frozen (Fr), Hot (Ho), Cold (Co), Tropic (Tr),
//     Tundra (Tu), TwilightZone (Tz), Farming (Fa).
//   - Referee-assigned in the rulebook's own text: MilitaryRule (Mr),
//     SubsectorCapital (Cp), SectorCapital (Cs), Capital (Cx), Colony
//     (Cy), Forbidden (Fo), DataRepository (Ab), AncientSite (An).
//   - Puzzle (Pz) and Dangerous (Da) look population-triggered at a
//     glance (Da: Pop 0-6, Pz: Pop 7-F — together every population
//     digit), but the rulebook states they're sub-labels applied only to
//     worlds a referee has already classified Travel Zone Amber ("assigns
//     to worlds a basic warning level based on experience", Book 3 line
//     877). Treating them as auto-derived would tag every generated world
//     Amber, which is not what the rulebook intends.
//   - Explicitly non-mainworld ("Not MW") in the rulebook: Mining (Mi),
//     PenalColony (Pe).
//   - Ambiguous: Dieback (Di) extracts with an identical Pop=0/Gov=0/Law=0
//     trigger to Barren (Ba) — likely narratively distinct (an ex-Barren
//     world) rather than digit-distinguishable. Applying both would be
//     wrong; arbitrarily suppressing one would be a guess. Left out.
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

	{Code: Barren, Population: []ehex.Value{0}, Government: []ehex.Value{0}, Law: []ehex.Value{0}},
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
		matchesAny(u.Law, t.Law)
}

// DeriveTradeCodes returns every trade code derivable purely from u's UWP
// digits. See tradeCodeTriggers' doc comment for what's deliberately
// excluded, and why.
func DeriveTradeCodes(u UWP) []TradeCode {
	var codes []TradeCode

	for _, t := range tradeCodeTriggers {
		if t.matches(u) {
			codes = append(codes, t.Code)
		}
	}

	return codes
}
