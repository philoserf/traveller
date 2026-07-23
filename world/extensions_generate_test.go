package world

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
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
}
