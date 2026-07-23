package sector

import "github.com/philoserf/traveller/world"

// capitalCandidate is one populated hex's mainworld, paired with its own
// Hex Location — the unit bestCandidate compares, whether the candidates
// come straight from a Sector's Hexes or (for the sector-wide pass) from
// the 16 subsectors' own already-decided winners.
type capitalCandidate struct {
	world    *world.World
	location string
}

// candidatesFromHexes collects one capitalCandidate per populated,
// inhabited hex in hexes. A system-present hex isn't necessarily
// inhabited — RollPopulation can land on 0 (an uninhabited outpost or
// rock) independently of whether a system was rolled there at all — and
// an uninhabited world can't be a Capital, so Population=0 hexes are
// skipped here rather than in bestCandidate.
func candidatesFromHexes(hexes []Hex) []capitalCandidate {
	var candidates []capitalCandidate

	for _, hex := range hexes {
		if hex.System == nil {
			continue
		}

		mw := hex.System.Orbits[hex.System.MainworldOrbit].World
		if mw.UWP.Population == 0 {
			continue
		}

		candidates = append(candidates, capitalCandidate{mw, hex.Location})
	}

	return candidates
}

// bestCandidate returns the candidate with the highest Importance
// (nil/"" if candidates is empty). Ties break by TradeCodes count (Book
// 3's own "most Trade Classifications" rule), then by Location
// ascending — the book doesn't specify further, so Location keeps this
// fully deterministic per seed.
func bestCandidate(candidates []capitalCandidate) (*world.World, string) {
	var best capitalCandidate

	for _, c := range candidates {
		switch {
		case best.world == nil, c.world.Importance > best.world.Importance:
			best = c
		case c.world.Importance == best.world.Importance:
			switch {
			case len(c.world.TradeCodes) > len(best.world.TradeCodes):
				best = c
			case len(c.world.TradeCodes) == len(best.world.TradeCodes) && c.location < best.location:
				best = c
			}
		}
	}

	return best.world, best.location
}

// assignCapitals scans sec (already fully generated) and mutates each
// winning mainworld's TradeCodes in place: the single most-Important
// inhabited hex in each of the 16 subsectors gets SubsectorCapital, and
// the single most-Important inhabited hex sector-wide gets SectorCapital
// — Book 3's own "Capitals... established by Importance" rule. The
// sector-wide winner is found among the 16 subsector winners rather than
// rescanning all 1280 hexes: whichever subsector contains the sector's
// best world necessarily reports it as that subsector's own winner too
// (same tie-break rule, strict superset of candidates), so re-deriving it
// from all 1280 hexes again would just repeat work already done. All 17
// winners are determined first, over the sector's original TradeCodes,
// before any mutation — so an early subsector-capital mutation can never
// skew a later tie-break count. A world can legitimately hold both codes
// at once (the sector's overall winner is trivially also its own
// subsector's winner).
//
// Not reproducible from HexSeed alone, unlike everything else about a
// generated hex: capital status depends on comparing Importance across
// the WHOLE sector, information a standalone single-hex reroll (e.g. via
// GET /systems/random?seed=<hex seed>) doesn't have access to. This is
// expected, not a reproducibility bug — see
// TestGenerateSectorHexIsIndependentlyReproducible, which only compares
// UWP for exactly this reason.
func assignCapitals(sec Sector) {
	type winner struct {
		world *world.World
		code  world.TradeCode
	}

	var winners []winner

	var subsectorWinners []capitalCandidate

	for letter := byte('A'); letter <= 'P'; letter++ {
		best, location := bestCandidate(candidatesFromHexes(sec.Subsector(letter)))
		if best == nil {
			continue
		}

		winners = append(winners, winner{best, world.SubsectorCapital})
		subsectorWinners = append(subsectorWinners, capitalCandidate{best, location})
	}

	if best, _ := bestCandidate(subsectorWinners); best != nil {
		winners = append(winners, winner{best, world.SectorCapital})
	}

	for _, w := range winners {
		w.world.TradeCodes = append(w.world.TradeCodes, w.code)
	}
}
