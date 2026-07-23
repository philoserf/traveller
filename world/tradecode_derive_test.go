package world

import (
	"slices"
	"testing"

	"github.com/philoserf/traveller/ehex"
)

func TestDeriveTradeCodes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		uwp  UWP
		want TradeCode
	}{
		{"AsteroidBelt", UWP{Size: 0, Atmosphere: 0, Hydrographics: 0}, AsteroidBelt},
		{"Desert", UWP{Atmosphere: 5, Hydrographics: 0}, Desert},
		{"Fluid", UWP{Atmosphere: 10, Hydrographics: 5}, Fluid},
		{"Garden", UWP{Size: 7, Atmosphere: 6, Hydrographics: 6}, Garden},
		{"Hellworld", UWP{Size: 5, Atmosphere: 4, Hydrographics: 1}, Hellworld},
		{"IceCapped", UWP{Atmosphere: 1, Hydrographics: 5}, IceCapped},
		{"Ocean", UWP{Size: 10, Atmosphere: 5, Hydrographics: 10}, Ocean},
		{"Vacuum", UWP{Atmosphere: 0}, Vacuum},
		{"WaterWorld", UWP{Size: 5, Atmosphere: 5, Hydrographics: 10}, WaterWorld},
		{"Barren", UWP{Population: 0, Government: 0, Law: 0}, Barren},
		{"Dieback", UWP{Population: 0, Government: 0, Law: 0, TechLevel: 5}, Dieback},
		{"LowPopulation", UWP{Population: 2}, LowPopulation},
		{"NonIndustrial", UWP{Population: 5}, NonIndustrial},
		{"PreHigh", UWP{Population: 8}, PreHigh},
		{"HighPopulation", UWP{Population: 10}, HighPopulation},
		{"PreAgricultural", UWP{Atmosphere: 5, Hydrographics: 5, Population: 4}, PreAgricultural},
		{"Agricultural", UWP{Atmosphere: 5, Hydrographics: 5, Population: 6}, Agricultural},
		{"NonAgricultural", UWP{Atmosphere: 1, Hydrographics: 1, Population: 7}, NonAgricultural},
		{"PrisonExileCamp", UWP{Atmosphere: 2, Hydrographics: 3, Population: 4, Law: 7}, PrisonExileCamp},
		{"PreIndustrial", UWP{Atmosphere: 1, Population: 7}, PreIndustrial},
		{"Industrial", UWP{Atmosphere: 1, Population: 10}, Industrial},
		{"Poor", UWP{Atmosphere: 3, Hydrographics: 1}, Poor},
		{"PreRich", UWP{Atmosphere: 6, Population: 5}, PreRich},
		{"Rich", UWP{Atmosphere: 6, Population: 7}, Rich},
		{"Reserve", UWP{Population: 2, Government: 6, Law: 4}, Reserve},
		{"Dangerous", UWP{Population: 3}, Dangerous},
		{"Puzzle", UWP{Population: 9, Government: 10, Law: 10}, Puzzle},       // Gov+Law=20
		{"Forbidden", UWP{Population: 9, Government: 15, Law: 10}, Forbidden}, // Gov+Law=25
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := DeriveTradeCodes(c.uwp)
			if !slices.Contains(got, c.want) {
				t.Errorf("DeriveTradeCodes(%s) = %v, want to contain %s", c.uwp, got, c.want)
			}
		})
	}
}

// TestBarrenVsDiebackMutuallyExclusive pins the specific distinction the
// Dieback trigger exists to draw: Barren (never populated, TL=0) and
// Dieback (evidence of past civilization, TL>=1) are alternatives for the
// same Pop=0/Gov=0/Law=0 condition, matching the rulebook's own Native
// Intelligent Life table, which draws exactly this TL=0-vs-TL=1+ line and
// never applies both labels to one world. Never both, for any TL.
func TestBarrenVsDiebackMutuallyExclusive(t *testing.T) {
	t.Parallel()

	neverPopulated := UWP{Population: 0, Government: 0, Law: 0, TechLevel: 0}
	got := DeriveTradeCodes(neverPopulated)

	if !slices.Contains(got, Barren) {
		t.Errorf("DeriveTradeCodes(TL=0) = %v, want to contain Barren", got)
	}

	if slices.Contains(got, Dieback) {
		t.Errorf("DeriveTradeCodes(TL=0) = %v, want NOT to contain Dieback (no evidence of past civilization)", got)
	}

	for _, tl := range []ehex.Value{1, 5, 15, 20, ehex.Max} {
		diedBack := UWP{Population: 0, Government: 0, Law: 0, TechLevel: tl}
		got = DeriveTradeCodes(diedBack)

		if slices.Contains(got, Barren) {
			t.Errorf(
				"DeriveTradeCodes(TL=%d) = %v, want NOT to contain Barren (evidence of past civilization present)",
				tl,
				got,
			)
		}

		if !slices.Contains(got, Dieback) {
			t.Errorf("DeriveTradeCodes(TL=%d) = %v, want to contain Dieback", tl, got)
		}
	}
}

// TestExcludedCodesNeverDerived guards against accidentally reintroducing a
// referee-assigned or orbit-dependent code — or Forbidden/Puzzle/Dangerous,
// whose real predicate doesn't fit this table's shape (see
// travelZoneTradeCode instead) — as a tradeCodeTriggers row. See the
// exclusion list documented on tradeCodeTriggers.
func TestExcludedCodesNeverDerived(t *testing.T) {
	t.Parallel()

	excluded := []TradeCode{
		Satellite, Locked,
		Frozen, Hot, Cold, Tropic, Tundra, TwilightZone, Farming,
		MilitaryRule, SubsectorCapital, SectorCapital, Capital, Colony, Forbidden, DataRepository, AncientSite,
		Puzzle, Dangerous,
		Mining, PenalColony,
	}

	for _, trigger := range tradeCodeTriggers {
		if slices.Contains(excluded, trigger.Code) {
			t.Errorf("tradeCodeTriggers contains excluded code %s", trigger.Code)
		}
	}
}
