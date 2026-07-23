package world

import (
	"math/rand/v2"
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
}
