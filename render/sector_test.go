package render_test

import (
	"strings"
	"testing"

	"github.com/philoserf/traveller/render"
	"github.com/philoserf/traveller/sector"
	"github.com/philoserf/traveller/system"
)

// fixtureSector builds a minimal Sector with one empty hex (0101) and one
// populated, single-star hex (0102) hosting fixtureMainworld() — the
// shared one-empty-one-populated fixture both TestSectorRendersEmptyAndPopulatedHexes
// and TestSectorCompactOmitsEmptyHexes exercise.
func fixtureSector() sector.Sector {
	primary := system.Star{
		SpectralType: system.SpectralG, SpectralDecimal: 2, LuminosityClass: "V",
		Role: system.Primary, HabitableZoneOrbit: 3,
	}

	sys := system.StarSystem{
		Orbits: []system.Orbit{
			{Number: -1, Star: &primary},
			{Number: 3, HostRole: system.Primary, World: fixtureMainworld()},
		},
		MainworldOrbit: 1,
	}

	return sector.Sector{
		Name: "Test",
		Hexes: []sector.Hex{
			{Location: "0101"},
			{Location: "0102", System: &sys},
		},
	}
}

// TestSectorRendersEmptyAndPopulatedHexes confirms the title, the
// empty-hex marker line, and a populated hex's locator line immediately
// preceding its full System(...) output.
func TestSectorRendersEmptyAndPopulatedHexes(t *testing.T) {
	t.Parallel()

	out := render.Sector(fixtureSector())

	if !strings.Contains(out, "# Test Sector") {
		t.Errorf("output missing sector title:\n%s", out)
	}

	if !strings.Contains(out, "**Hex 0101:** empty") {
		t.Errorf("output missing empty-hex marker:\n%s", out)
	}

	hexIdx := strings.Index(out, "**Hex 0102**")
	if hexIdx == -1 {
		t.Fatalf("output missing populated-hex locator line:\n%s", out)
	}

	systemIdx := strings.Index(out, "# A788899-C System")
	if systemIdx == -1 {
		t.Fatalf("output missing the populated hex's full system render:\n%s", out)
	}

	if systemIdx < hexIdx {
		t.Errorf("system render appears before its own hex locator line, want it nested after:\n%s", out)
	}
}

// TestSectorCompactOmitsEmptyHexes confirms SectorCompact emits a table
// row only for the populated hex — the empty hex is omitted entirely,
// and there's none of System(...)'s nested star/orbit detail.
func TestSectorCompactOmitsEmptyHexes(t *testing.T) {
	t.Parallel()

	out := render.SectorCompact(fixtureSector())

	if !strings.Contains(out, "# Test Sector (compact)") {
		t.Errorf("output missing sector title:\n%s", out)
	}

	if strings.Contains(out, "0101") {
		t.Errorf("empty hex 0101 should be omitted entirely:\n%s", out)
	}

	if !strings.Contains(out, "| 0102 | A788899-C |") {
		t.Errorf("output missing populated-hex UWP:\n%s", out)
	}

	if strings.Contains(out, "###") {
		t.Errorf("compact output should have no nested star detail:\n%s", out)
	}
}
