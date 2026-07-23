package world

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

func TestClampEhex(t *testing.T) {
	t.Parallel()

	cases := []struct {
		v, lo, hi int
		want      ehex.Value
	}{
		{v: -5, lo: 0, hi: 10, want: 0},
		{v: 15, lo: 0, hi: 10, want: 10},
		{v: 5, lo: 0, hi: 10, want: 5},
	}

	for _, c := range cases {
		if got := clampEhex(c.v, c.lo, c.hi); got != c.want {
			t.Errorf("clampEhex(%d, %d, %d) = %d, want %d", c.v, c.lo, c.hi, got, c.want)
		}
	}
}

func TestTechLevelModifierWorstCase(t *testing.T) {
	t.Parallel()

	// Starport X (-4), Size/Atm/Hyd/Pop in their zero-modifier bands,
	// Government D=13 (-2): total -6, matching rollTechLevel's documented
	// worst case (D6 min 1, so TechLevel would compute to -5 before the
	// floor clamp).
	u := UWP{
		Starport: StarportNone, Size: 5, Atmosphere: 5, Hydrographics: 0,
		Population: 0, Government: 13,
	}

	if got, want := techLevelModifier(u), -6; got != want {
		t.Errorf("techLevelModifier(worst case) = %d, want %d", got, want)
	}
}

func TestTechLevelModifierBestCase(t *testing.T) {
	t.Parallel()

	// Starport A (+6), Size 0 (+2), Atm 0 (+1), Hyd A=10 (+2), Pop A=10 (+4),
	// Government 0 (+1): total +16.
	u := UWP{
		Starport: StarportA, Size: 0, Atmosphere: 0, Hydrographics: 10,
		Population: 10, Government: 0,
	}

	if got, want := techLevelModifier(u), 16; got != want {
		t.Errorf("techLevelModifier(best case) = %d, want %d", got, want)
	}
}

func TestRollTechLevelNeverInvalid(t *testing.T) {
	t.Parallel()

	// Worst-case modifiers (-6) plus the minimum possible D6 roll (1) would
	// compute to -5 without the floor clamp — assert the clamp holds.
	u := UWP{
		Starport: StarportNone, Size: 5, Atmosphere: 5, Hydrographics: 0,
		Population: 0, Government: 13,
	}

	r := dice.New(rand.NewPCG(1, 1))

	for range 1000 {
		tl := rollTechLevel(r, u)
		if !tl.Valid() {
			t.Fatalf("rollTechLevel produced invalid ehex.Value %d", tl)
		}
	}
}

func TestGenerateUWPInvariants(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(7, 8))

	for range 20000 {
		u := GenerateUWP(r)

		if u.Size == 0 && u.Atmosphere != 0 {
			t.Fatalf("Size=0 but Atmosphere=%s (want forced 0): %s", u.Atmosphere, u)
		}

		if u.Size < 2 && u.Hydrographics != 0 {
			t.Fatalf("Size=%s (<2) but Hydrographics=%s (want forced 0): %s", u.Size, u.Hydrographics, u)
		}

		fields := map[string]ehex.Value{
			"Size": u.Size, "Atmosphere": u.Atmosphere, "Hydrographics": u.Hydrographics,
			"Population": u.Population, "Government": u.Government, "Law": u.Law, "TechLevel": u.TechLevel,
		}
		for name, v := range fields {
			if !v.Valid() {
				t.Fatalf("%s=%d is not a valid ehex digit: %s", name, v, u)
			}
		}

		if u.Law > lawCeiling {
			t.Fatalf("Law=%s exceeds ceiling %s: %s", u.Law, lawCeiling, u)
		}
	}
}

// TestGenerateWithSizeUsesGivenSizeRoll confirms generateWithSize
// actually threads its sizeRoll parameter through to the resulting
// World's UWP — the mainworld BigWorld fallback (world/system_generate.go)
// depends on generateWithSize(r, rollBigWorldSize) producing a BigWorld-
// range Size (Book 3 p.29: "Siz= 2D+7", "any with Siz=B+ is BW"), not
// silently falling back to the standard rollSize formula.
func TestGenerateWithSizeUsesGivenSizeRoll(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(3, 4))

	for range 1000 {
		w := generateWithSize(r, rollBigWorldSize)
		if w.UWP.Size < 9 {
			t.Fatalf(
				"generateWithSize(r, rollBigWorldSize): Size = %s, want >= 9 (2D+7 floor: TwoD6 min 2, +7)",
				w.UWP.Size,
			)
		}
	}
}

func TestGenerateDeterminism(t *testing.T) {
	t.Parallel()

	r1 := dice.New(rand.NewPCG(99, 99))
	r2 := dice.New(rand.NewPCG(99, 99))

	w1 := Generate(r1)
	w2 := Generate(r2)

	if w1.UWP.String() != w2.UWP.String() {
		t.Fatalf("identical seeds produced different UWPs: %s vs %s", w1.UWP, w2.UWP)
	}

	if len(w1.TradeCodes) != len(w2.TradeCodes) {
		t.Fatalf("identical seeds produced different trade code counts: %v vs %v", w1.TradeCodes, w2.TradeCodes)
	}

	if len(w1.Bases) != len(w2.Bases) {
		t.Fatalf("identical seeds produced different bases: %v vs %v", w1.Bases, w2.Bases)
	}

	if w1.PBG != w2.PBG {
		t.Fatalf("identical seeds produced different PBG: %s vs %s", w1.PBG, w2.PBG)
	}
}

func TestBaseTargetsOnlyAAndBForNaval(t *testing.T) {
	t.Parallel()

	for _, s := range []Starport{StarportC, StarportD, StarportE, StarportNone} {
		if _, ok := navalBaseTarget(s); ok {
			t.Errorf("navalBaseTarget(%s) = ok, want not ok (Naval Base only available at A/B)", s)
		}
	}

	cases := map[Starport]int{StarportA: 6, StarportB: 5}
	for s, want := range cases {
		got, ok := navalBaseTarget(s)
		if !ok || got != want {
			t.Errorf("navalBaseTarget(%s) = (%d, %v), want (%d, true)", s, got, ok, want)
		}
	}
}

func TestBaseTargetsNotAtEOrX(t *testing.T) {
	t.Parallel()

	for _, s := range []Starport{StarportE, StarportNone} {
		if _, ok := scoutBaseTarget(s); ok {
			t.Errorf("scoutBaseTarget(%s) = ok, want not ok (no bases at Starport E/X)", s)
		}
	}

	cases := map[Starport]int{StarportA: 4, StarportB: 5, StarportC: 6, StarportD: 7}
	for s, want := range cases {
		got, ok := scoutBaseTarget(s)
		if !ok || got != want {
			t.Errorf("scoutBaseTarget(%s) = (%d, %v), want (%d, true)", s, got, ok, want)
		}
	}
}

func TestRollBasesKnownCodesOnly(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(11, 12))
	starports := []Starport{StarportA, StarportB, StarportC, StarportD, StarportE, StarportNone}

	for range 2000 {
		for _, s := range starports {
			for _, b := range rollBases(r, s) {
				if b != NavalBase && b != ScoutBase {
					t.Fatalf("rollBases(%s) produced unexpected base %s, want only NavalBase/ScoutBase", s, b)
				}
			}
		}
	}
}

func TestRollBasesNoneAtEOrX(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(21, 22))

	for range 2000 {
		for _, s := range []Starport{StarportE, StarportNone} {
			if bases := rollBases(r, s); len(bases) != 0 {
				t.Fatalf("rollBases(%s) = %v, want none (no bases at Starport E/X)", s, bases)
			}
		}
	}
}

func TestRollBasesNavalOnlyAtAOrB(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(31, 32))

	for range 2000 {
		for _, s := range []Starport{StarportC, StarportD} {
			if slices.Contains(rollBases(r, s), NavalBase) {
				t.Fatalf("rollBases(%s) produced NavalBase, want Naval only at Starport A/B", s)
			}
		}
	}
}

func TestRollPBGInvariants(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(13, 14))

	for range 10000 {
		pbg := rollPBG(r, 0)
		if pbg.PopulationDigit != 0 {
			t.Fatalf("rollPBG(population=0).PopulationDigit = %s, want 0", pbg.PopulationDigit)
		}

		pbg = rollPBG(r, 5)
		if pbg.PopulationDigit < 1 || pbg.PopulationDigit > 9 {
			t.Fatalf("rollPBG(population=5).PopulationDigit = %s, want 1..9", pbg.PopulationDigit)
		}

		if pbg.Belts > 3 {
			t.Fatalf("rollPBG(...).Belts = %s, want 0..3", pbg.Belts)
		}

		if pbg.GasGiants > 4 {
			t.Fatalf("rollPBG(...).GasGiants = %s, want 0..4", pbg.GasGiants)
		}
	}
}
