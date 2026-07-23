package world

import (
	"slices"
	"testing"
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

// TestExcludedCodesNeverDerived guards against accidentally reintroducing a
// referee-assigned or orbit-dependent code into the trigger table — see the
// exclusion list documented on tradeCodeTriggers.
func TestExcludedCodesNeverDerived(t *testing.T) {
	t.Parallel()

	excluded := []TradeCode{
		Satellite, Locked,
		Frozen, Hot, Cold, Tropic, Tundra, TwilightZone, Farming,
		MilitaryRule, SubsectorCapital, SectorCapital, Capital, Colony, Forbidden, DataRepository, AncientSite,
		Puzzle, Dangerous,
		Mining, PenalColony,
		Dieback,
	}

	for _, trigger := range tradeCodeTriggers {
		if slices.Contains(excluded, trigger.Code) {
			t.Errorf("tradeCodeTriggers contains excluded code %s", trigger.Code)
		}
	}
}
