package sector

import (
	"slices"
	"testing"

	"github.com/philoserf/traveller/system"
	"github.com/philoserf/traveller/world"
)

// fixtureHex builds a populated, inhabited Hex at location whose
// mainworld has the given Importance and exactly tradeCodeCount
// TradeCodes (content doesn't matter, only the count, for tie-break
// purposes). Population is fixed at 1 (any nonzero value) unless a test
// specifically needs to exercise the uninhabited-world exclusion.
func fixtureHex(location string, importance world.Importance, tradeCodeCount int) Hex {
	return fixtureHexPop(location, importance, tradeCodeCount, 1)
}

func fixtureHexPop(location string, importance world.Importance, tradeCodeCount, population int) Hex {
	codes := make([]world.TradeCode, tradeCodeCount)
	for i := range codes {
		codes[i] = world.TradeCode("X")
	}

	mw := &world.World{
		Importance: importance,
		TradeCodes: codes,
		UWP:        world.UWP{Population: world.ClampEhex(population, 0, 15)},
	}

	return Hex{
		Location: location,
		System: &system.StarSystem{
			Orbits:         []system.Orbit{{World: mw}},
			MainworldOrbit: 0,
		},
	}
}

func TestBestCandidatePicksHighestImportance(t *testing.T) {
	t.Parallel()

	hexes := []Hex{
		fixtureHex("0101", 1, 0),
		fixtureHex("0102", 4, 0),
		fixtureHex("0103", 2, 0),
	}

	best, location := bestCandidate(candidatesFromHexes(hexes))
	if best == nil || location != "0102" {
		t.Fatalf("bestCandidate = (%v, %q), want the Importance=4 world at 0102", best, location)
	}
}

func TestBestCandidateTiesBreakByTradeCodeCount(t *testing.T) {
	t.Parallel()

	hexes := []Hex{
		fixtureHex("0101", 3, 1),
		fixtureHex("0102", 3, 4),
	}

	best, location := bestCandidate(candidatesFromHexes(hexes))
	if best == nil || location != "0102" {
		t.Fatalf("bestCandidate = (%v, %q), want the 4-TradeCode world at 0102 (tie-break)", best, location)
	}
}

func TestBestCandidateTiesBreakByLocation(t *testing.T) {
	t.Parallel()

	hexes := []Hex{
		fixtureHex("0205", 3, 2),
		fixtureHex("0103", 3, 2),
	}

	best, location := bestCandidate(candidatesFromHexes(hexes))
	if best == nil || location != "0103" {
		t.Fatalf("bestCandidate = (%v, %q), want the lexicographically-earlier 0103 (fully-tied)", best, location)
	}
}

func TestBestCandidateNoPopulatedHexes(t *testing.T) {
	t.Parallel()

	hexes := []Hex{{Location: "0101"}, {Location: "0102"}}

	if best, location := bestCandidate(candidatesFromHexes(hexes)); best != nil || location != "" {
		t.Fatalf("bestCandidate(all empty) = (%v, %q), want (nil, \"\")", best, location)
	}
}

// TestCandidatesFromHexesExcludesUninhabitedWorlds confirms a
// system-present hex whose mainworld rolled Population=0 (an
// uninhabited outpost or rock — system presence and population are
// independent rolls) is never eligible to become a Capital, even with
// the highest Importance in the set.
func TestCandidatesFromHexesExcludesUninhabitedWorlds(t *testing.T) {
	t.Parallel()

	hexes := []Hex{
		fixtureHexPop("0101", 9, 5, 0), // highest Importance, but uninhabited
		fixtureHex("0102", 1, 0),
	}

	best, location := bestCandidate(candidatesFromHexes(hexes))
	if best == nil || location != "0102" {
		t.Fatalf("bestCandidate = (%v, %q), want the inhabited (if lower-Importance) world at 0102", best, location)
	}
}

// TestAssignCapitalsMarksSubsectorAndSectorWinners runs a real
// GenerateSector and confirms exactly one hex sector-wide carries
// SectorCapital, and every non-empty subsector carries at most one
// SubsectorCapital — the integration-level check that assignCapitals,
// called from GenerateSector itself, actually ran.
func TestAssignCapitalsMarksSubsectorAndSectorWinners(t *testing.T) {
	t.Parallel()

	sec := GenerateSector(1, "Test", DensityStandard)

	sectorCapitals := 0

	for _, hex := range sec.Hexes {
		if hex.System == nil {
			continue
		}

		if slices.Contains(hex.System.Orbits[hex.System.MainworldOrbit].World.TradeCodes, world.SectorCapital) {
			sectorCapitals++
		}
	}

	if sectorCapitals != 1 {
		t.Errorf("sector-wide SectorCapital count = %d, want exactly 1", sectorCapitals)
	}

	for letter := byte('A'); letter <= 'P'; letter++ {
		subsectorCapitals := 0

		for _, hex := range sec.Subsector(letter) {
			if hex.System == nil {
				continue
			}

			if slices.Contains(hex.System.Orbits[hex.System.MainworldOrbit].World.TradeCodes, world.SubsectorCapital) {
				subsectorCapitals++
			}
		}

		if subsectorCapitals > 1 {
			t.Errorf("subsector %c: SubsectorCapital count = %d, want at most 1", letter, subsectorCapitals)
		}
	}
}
