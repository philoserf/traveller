package world

import "github.com/philoserf/traveller/ehex"

// DeriveOrbitTradeCodes returns trade codes derivable from a world's
// orbit position relative to its system's Habitable Zone (HZ) orbit
// number — Frozen, Hot, Cold, Tropic, Tundra, TwilightZone, Farming (Book
// 3 p.26) — none of which DeriveTradeCodes can produce from a standalone
// UWP, since a UWP alone carries no orbit data (see its own doc comment).
// isMainworld gates Farming, which the rulebook restricts to "HZ but not
// MW".
func DeriveOrbitTradeCodes(u UWP, orbit, hzOrbit int, isMainworld bool) []TradeCode {
	var codes []TradeCode

	if orbit == 0 || orbit == 1 {
		codes = append(codes, TwilightZone)
	}

	return append(codes, climateTradeCodes(u, orbit-hzOrbit, isMainworld)...)
}

// climateTradeCodes handles Hot/Cold/Tropic/Tundra/Farming/Frozen — the
// delta-from-HZ half of DeriveOrbitTradeCodes, split out to keep either
// function's branching manageable on its own.
func climateTradeCodes(u UWP, delta int, isMainworld bool) []TradeCode {
	var codes []TradeCode

	switch {
	case delta <= -2:
		// no named climate code this close in
	case delta == -1:
		codes = append(codes, Hot)

		if isTropicOrTundra(u) {
			codes = append(codes, Tropic)
		}
	case delta == 0:
		if !isMainworld &&
			matchesAny(u.Atmosphere, []ehex.Value{4, 5, 6, 7, 8, 9}) &&
			matchesAny(u.Hydrographics, []ehex.Value{4, 5, 6, 7, 8}) &&
			matchesAny(u.Population, []ehex.Value{2, 3, 4, 5, 6}) {
			codes = append(codes, Farming)
		}
	case delta == 1:
		codes = append(codes, Cold)

		if isTropicOrTundra(u) {
			codes = append(codes, Tundra)
		}
	default: // delta >= 2
		if matchesAny(u.Size, []ehex.Value{2, 3, 4, 5, 6, 7, 8, 9}) &&
			matchesAny(u.Hydrographics, []ehex.Value{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}) {
			codes = append(codes, Frozen)
		}
	}

	return codes
}

// isTropicOrTundra is Tropic's and Tundra's shared UWP-digit predicate
// (Book 3 p.26): the two trade codes differ only in which side of the HZ
// the world orbits (checked by the caller), not in this condition.
func isTropicOrTundra(u UWP) bool {
	return matchesAny(u.Size, []ehex.Value{6, 7, 8, 9}) &&
		matchesAny(u.Atmosphere, []ehex.Value{4, 5, 6, 7, 8, 9}) &&
		matchesAny(u.Hydrographics, []ehex.Value{3, 4, 5, 6, 7})
}
