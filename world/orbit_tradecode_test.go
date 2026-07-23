package world

import (
	"slices"
	"testing"
)

func TestDeriveOrbitTradeCodesDeltaBoundaries(t *testing.T) {
	t.Parallel()

	// A UWP that qualifies for every delta-gated code (Tropic/Tundra's
	// Size/Atm/Hyd predicate, and Frozen's), so each case below isolates
	// the delta boundary itself, not an unmet secondary condition.
	qualifies := UWP{Size: 7, Atmosphere: 6, Hydrographics: 5, Population: 4}

	cases := []struct {
		name   string
		orbit  int
		hz     int
		want   TradeCode
		absent TradeCode
	}{
		{"delta -2: no climate code", 5, 7, "", Hot},
		{"delta -1: Hot+Tropic", 6, 7, Tropic, Cold},
		{"delta 0: no Hot/Cold (mainworld)", 7, 7, "", Hot},
		{"delta +1: Cold+Tundra", 8, 7, Tundra, Hot},
		{"delta +2: Frozen", 9, 7, Frozen, Cold},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := DeriveOrbitTradeCodes(qualifies, c.orbit, c.hz, true)

			if c.want != "" && !slices.Contains(got, c.want) {
				t.Errorf("DeriveOrbitTradeCodes(orbit=%d, hz=%d) = %v, want to contain %s", c.orbit, c.hz, got, c.want)
			}

			if slices.Contains(got, c.absent) {
				t.Errorf(
					"DeriveOrbitTradeCodes(orbit=%d, hz=%d) = %v, want NOT to contain %s",
					c.orbit,
					c.hz,
					got,
					c.absent,
				)
			}
		})
	}
}

func TestDeriveOrbitTradeCodesTwilightZone(t *testing.T) {
	t.Parallel()

	u := UWP{}

	for _, orbit := range []int{0, 1} {
		got := DeriveOrbitTradeCodes(u, orbit, 5, true)
		if !slices.Contains(got, TwilightZone) {
			t.Errorf("DeriveOrbitTradeCodes(orbit=%d) = %v, want to contain TwilightZone", orbit, got)
		}
	}

	got := DeriveOrbitTradeCodes(u, 2, 5, true)
	if slices.Contains(got, TwilightZone) {
		t.Errorf("DeriveOrbitTradeCodes(orbit=2) = %v, want NOT to contain TwilightZone", got)
	}
}

func TestDeriveOrbitTradeCodesFarmingExcludesMainworld(t *testing.T) {
	t.Parallel()

	u := UWP{Atmosphere: 6, Hydrographics: 5, Population: 4}

	if got := DeriveOrbitTradeCodes(u, 5, 5, false); !slices.Contains(got, Farming) {
		t.Errorf("DeriveOrbitTradeCodes(non-mainworld, delta=0) = %v, want to contain Farming", got)
	}

	if got := DeriveOrbitTradeCodes(u, 5, 5, true); slices.Contains(got, Farming) {
		t.Errorf("DeriveOrbitTradeCodes(mainworld, delta=0) = %v, want NOT to contain Farming (\"HZ but not MW\")", got)
	}
}
