package render_test

import (
	"strings"
	"testing"

	"github.com/philoserf/traveller/render"
	"github.com/philoserf/traveller/world"
)

// regina is the rulebook's own worked example, reused here (as elsewhere
// in this codebase) as a known-good fixture rather than an arbitrary one.
var regina = world.World{
	UWP: world.UWP{
		Starport: world.StarportA, Size: 7, Atmosphere: 8, Hydrographics: 8,
		Population: 8, Government: 9, Law: 9, TechLevel: 12,
	},
	TradeCodes: []world.TradeCode{world.Rich, world.PreAgricultural, world.PreHigh},
	Bases:      []world.Base{world.NavalBase, world.ScoutBase},
	PBG:        world.PBG{PopulationDigit: 7, Belts: 0, GasGiants: 3},
	Importance: 4,
	Economic:   world.Economic{Resources: 13, Labor: 7, Infrastructure: 14, Efficiency: 4},
	Cultural:   world.Cultural{Heterogeneity: 9, Acceptance: 12, Strangeness: 6, Symbols: 13},
}

func TestWorldContainsAllFields(t *testing.T) {
	t.Parallel()

	out := render.World(regina)

	want := []string{
		"A788899-C",      // UWP
		"Ri", "Pa", "Ph", // Trade Codes
		"N", "S", // Bases
		"703",           // PBG
		"+4",            // Importance
		"13", "7", "14", // Economic Resources/Labor/Infrastructure
		"9", "12", "6", // Cultural Heterogeneity/Acceptance/Strangeness
	}

	for _, w := range want {
		if !strings.Contains(out, w) {
			t.Errorf("render.World(regina) missing %q in output:\n%s", w, out)
		}
	}
}

func TestWorldTitleFallsBackToUWP(t *testing.T) {
	t.Parallel()

	out := render.World(regina) // regina.Name is unset, matching Generate()'s real output
	if !strings.HasPrefix(out, "# A788899-C\n") {
		t.Errorf("render.World with no Name should title with the UWP code, got:\n%s", out)
	}

	named := regina
	named.Name = "Regina"

	out = render.World(named)
	if !strings.HasPrefix(out, "# Regina\n") {
		t.Errorf("render.World with a Name set should use it as the title, got:\n%s", out)
	}
}

func TestWorldOmitsEmptyBasesAndTradeCodes(t *testing.T) {
	t.Parallel()

	bare := world.World{UWP: world.UWP{Starport: world.StarportNone}}
	out := render.World(bare)

	if !strings.Contains(out, "**Trade Codes:** None") {
		t.Errorf("render.World with no trade codes should show \"None\", got:\n%s", out)
	}

	if !strings.Contains(out, "**Bases:** None") {
		t.Errorf("render.World with no bases should show \"None\", got:\n%s", out)
	}
}

func TestWorldOmitsUngeneratedTravelZone(t *testing.T) {
	t.Parallel()

	out := render.World(regina) // regina.TravelZone is unset, matching Generate()'s real output
	if strings.Contains(out, "Travel Zone") {
		t.Errorf("render.World should omit Travel Zone when it's not set, got:\n%s", out)
	}

	zoned := regina
	zoned.TravelZone = world.ZoneGreen

	out = render.World(zoned)
	if !strings.Contains(out, "**Travel Zone:** Green") {
		t.Errorf("render.World should show a set Travel Zone, got:\n%s", out)
	}
}
