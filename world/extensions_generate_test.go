package world

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// Regina's UWP is A788899-C: Starport A, Size 7, Atmosphere 8,
// Hydrographics 8, Population 8, Government 9, Law 9, TechLevel C(12).
// Her published extensions are Ix={+4}, Ex=(D7E+4), Cx=[9C6D], PBG=703
// (PopulationDigit 7, Belts 0, GasGiants 3) — the rulebook's own worked
// example, used throughout this file to pin every formula against a
// known-correct source rather than only against each other.
var reginaUWP = UWP{
	Starport:      StarportA,
	Size:          7,
	Atmosphere:    8,
	Hydrographics: 8,
	Population:    8,
	Government:    9,
	Law:           9,
	TechLevel:     12,
}

func TestComputeImportanceMatchesRegina(t *testing.T) {
	t.Parallel()

	// Regina's actual trade codes, used directly rather than via
	// DeriveTradeCodes: An (AncientSite) and Cp (SubsectorCapital) are
	// referee-assigned and never appear in generated output, but this
	// test is pinning computeImportance's formula, not the generator.
	tradeCodes := []TradeCode{Rich, PreAgricultural, PreHigh, AncientSite, SubsectorCapital}

	// The book's own worked example computes Ix=+4 BEFORE Naval/Scout
	// bases are rolled (a procedural quirk in the source) — reproduced
	// here with no bases, matching that moment exactly.
	if got, want := computeImportance(reginaUWP, tradeCodes, nil), Importance(4); got != want {
		t.Errorf("computeImportance(Regina, no bases) = %d, want %d (book's worked example)", got, want)
	}

	// Regina's actual bases are Naval+Scout. This implementation computes
	// Ix after Bases so the bonus is included — deliberately +5, not the
	// book's +4 (see the doc comment on computeImportance).
	withBases := []Base{NavalBase, ScoutBase}
	if got, want := computeImportance(reginaUWP, tradeCodes, withBases), Importance(5); got != want {
		t.Errorf("computeImportance(Regina, with bases) = %d, want %d (includes Naval+Scout bonus)", got, want)
	}
}

// TestComputeImportanceWayStationBonus exercises the WayStation branch
// directly: rollBases (the only real producer of World.Bases in this
// codebase) never emits WayStation — it's excluded as a density/frequency
// placement rather than a per-world roll (see rollBases's doc comment) —
// so Generate() can never reach this branch. computeImportance itself
// still implements the full Ix formula for any caller that does supply a
// WayStation base (e.g. a manually-constructed World), and this pins that
// behavior instead of leaving it untested.
func TestComputeImportanceWayStationBonus(t *testing.T) {
	t.Parallel()

	u := UWP{Starport: StarportC, TechLevel: 5, Population: 8}

	without := computeImportance(u, nil, nil)
	with := computeImportance(u, nil, []Base{WayStation})

	if with != without+1 {
		t.Errorf("computeImportance with WayStation = %d, want %d (without) + 1", with, without)
	}
}

func TestComputeTravelZoneMatchesRegina(t *testing.T) {
	t.Parallel()

	// Gov+Law=18, Population=8: neither Amber trigger fires.
	if got, want := computeTravelZone(reginaUWP), ZoneGreen; got != want {
		t.Errorf("computeTravelZone(Regina) = %v, want %v", got, want)
	}
}

// TestComputeTravelZoneBoundaries pins the exact Gov+Law thresholds from
// Book 3 p.28 ("Z Travel Zones"): Amber at 20, Red at 22, one point below
// each staying in the lower zone. Population is held above 6 throughout so
// only the Government+Law trigger is exercised.
func TestComputeTravelZoneBoundaries(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		government ehex.Value
		law        ehex.Value
		want       TravelZone
	}{
		{"govLaw 19 stays Green", 10, 9, ZoneGreen},
		{"govLaw 20 triggers Amber", 10, 10, ZoneAmber},
		{"govLaw 21 stays Amber", 11, 10, ZoneAmber},
		{"govLaw 22 triggers Red", 11, 11, ZoneRed},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			u := UWP{Population: 9, Government: c.government, Law: c.law}
			if got := computeTravelZone(u); got != c.want {
				t.Errorf("computeTravelZone(Gov=%d, Law=%d) = %v, want %v", c.government, c.law, got, c.want)
			}
		})
	}
}

// TestComputeTravelZoneIgnoresPopulationAlone confirms low Population
// alone (Government+Law both 0) does NOT trigger Amber — Book 3 p.28's
// "Da if pop 0-6" line only picks Amber's own Trade Code label (see
// travelZoneTradeCode), it doesn't set the zone itself.
func TestComputeTravelZoneIgnoresPopulationAlone(t *testing.T) {
	t.Parallel()

	if got, want := computeTravelZone(UWP{Population: 0}), ZoneGreen; got != want {
		t.Errorf("computeTravelZone(Population=0, Gov=0, Law=0) = %v, want %v", got, want)
	}
}

func TestComputeLaborMatchesRegina(t *testing.T) {
	t.Parallel()

	if got, want := computeLabor(8), 7; got != want {
		t.Errorf("computeLabor(8) = %d, want %d", got, want)
	}

	if got, want := computeLabor(0), 0; got != want {
		t.Errorf("computeLabor(0) = %d, want %d (floored, not -1)", got, want)
	}
}

func TestResourcesFromRollMatchesRegina(t *testing.T) {
	t.Parallel()

	// Regina: TL=12 (>=8), PBG GasGiants=3, Belts=0, published Resources=D(13).
	if got, want := resourcesFromRoll(10, 12, PBG{GasGiants: 3}), 13; got != want {
		t.Errorf("resourcesFromRoll(10, TL=12, GG=3) = %d, want %d", got, want)
	}

	if got, want := resourcesFromRoll(10, 5, PBG{GasGiants: 3}), 10; got != want {
		t.Errorf("resourcesFromRoll(10, TL=5, GG=3) = %d, want %d (TL<8: no Gas Giant/Belt bonus)", got, want)
	}
}

func TestInfrastructureFromRollMatchesRegina(t *testing.T) {
	t.Parallel()

	// Regina: Population=8 (2D bracket), Ix=+4, published Infrastructure=E(14).
	if got, want := infrastructureFromRoll(10, 4), 14; got != want {
		t.Errorf("infrastructureFromRoll(10, ix=4) = %d, want %d", got, want)
	}

	if got, want := infrastructureFromRoll(1, -3), 0; got != want {
		t.Errorf("infrastructureFromRoll(1, ix=-3) = %d, want %d (floored, not negative)", got, want)
	}
}

func TestCulturalComponentsMatchRegina(t *testing.T) {
	t.Parallel()

	// Regina: Population=8, Ix=+4 (the book's own pre-bases value, since
	// Cx uses the same Ix the book already computed), published
	// Cx=[9C6D]: Heterogeneity=9, Acceptance=C(12), Strangeness=6, Symbols=D(13).
	// All four need Flux=1 to reproduce Regina's numbers exactly.
	const flux = 1

	if got, want := computeHeterogeneity(8, flux), 9; got != want {
		t.Errorf("computeHeterogeneity(8, flux=1) = %d, want %d", got, want)
	}

	if got, want := computeAcceptance(8, 4), 12; got != want {
		t.Errorf("computeAcceptance(8, ix=4) = %d, want %d", got, want)
	}

	if got, want := computeStrangeness(flux), 6; got != want {
		t.Errorf("computeStrangeness(flux=1) = %d, want %d", got, want)
	}

	if got, want := computeSymbols(12, flux), 13; got != want {
		t.Errorf("computeSymbols(TL=12, flux=1) = %d, want %d", got, want)
	}
}

func TestCulturalFloorsAtOne(t *testing.T) {
	t.Parallel()

	if got := computeHeterogeneity(1, -5); got < 1 {
		t.Errorf("computeHeterogeneity(1, flux=-5) = %d, want >=1 (floored)", got)
	}

	if got := computeAcceptance(1, -5); got < 1 {
		t.Errorf("computeAcceptance(1, ix=-5) = %d, want >=1 (floored)", got)
	}
}

func TestRollEconomicPopulationZeroInfrastructure(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))
	u := UWP{Population: 0, TechLevel: 5}

	for range 1000 {
		ex := rollEconomic(r, u, 5, PBG{}) // even a large Ix must not leak through
		if ex.Infrastructure != 0 {
			t.Fatalf("rollEconomic(Population=0).Infrastructure = %d, want 0 regardless of Ix", ex.Infrastructure)
		}
	}
}

func TestRollCulturalPopulationZero(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(2, 2))
	u := UWP{Population: 0, TechLevel: 5}

	for range 1000 {
		cx := rollCultural(r, u, 5)
		if cx != (Cultural{}) {
			t.Fatalf("rollCultural(Population=0) = %+v, want all-zero (overrides the floor-of-1 rule)", cx)
		}
	}
}

func isValidTravelZone(z TravelZone) bool {
	return z == ZoneGreen || z == ZoneAmber || z == ZoneRed
}

func TestGenerateExtensionsInvariants(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(15, 16))

	for range 20000 {
		w := Generate(r)

		if w.Economic.Labor < 0 || w.Economic.Resources < 0 || w.Economic.Infrastructure < 0 {
			t.Fatalf("Economic has a negative field that should be floored at 0: %+v", w.Economic)
		}

		if w.Economic.Efficiency < -5 || w.Economic.Efficiency > 5 {
			t.Fatalf("Economic.Efficiency = %d, want -5..5 (Flux)", w.Economic.Efficiency)
		}

		if !isValidTravelZone(w.TravelZone) {
			t.Fatalf("TravelZone = %v, want one of Green/Amber/Red", w.TravelZone)
		}

		if w.UWP.Population == 0 {
			if w.Cultural != (Cultural{}) {
				t.Fatalf("Population=0 but Cultural is non-zero: %+v", w.Cultural)
			}

			continue
		}

		if w.Cultural.Heterogeneity < 1 || w.Cultural.Acceptance < 1 ||
			w.Cultural.Strangeness < 1 || w.Cultural.Symbols < 1 {
			t.Fatalf("Population>0 but a Cultural field is below its floor of 1: %+v", w.Cultural)
		}
	}
}

func TestGenerateDeterminismIncludesExtensions(t *testing.T) {
	t.Parallel()

	r1 := dice.New(rand.NewPCG(99, 99))
	r2 := dice.New(rand.NewPCG(99, 99))

	w1 := Generate(r1)
	w2 := Generate(r2)

	if w1.Importance != w2.Importance {
		t.Errorf("identical seeds produced different Importance: %d vs %d", w1.Importance, w2.Importance)
	}

	if w1.Economic != w2.Economic {
		t.Errorf("identical seeds produced different Economic: %+v vs %+v", w1.Economic, w2.Economic)
	}

	if w1.Cultural != w2.Cultural {
		t.Errorf("identical seeds produced different Cultural: %+v vs %+v", w1.Cultural, w2.Cultural)
	}

	if w1.TravelZone != w2.TravelZone {
		t.Errorf("identical seeds produced different TravelZone: %v vs %v", w1.TravelZone, w2.TravelZone)
	}
}
