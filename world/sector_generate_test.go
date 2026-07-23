package world

import (
	"math/rand/v2"
	"reflect"
	"testing"

	"github.com/philoserf/traveller/dice"
)

func TestHexLocationBoundaries(t *testing.T) {
	t.Parallel()

	cases := []struct {
		col, row int
		want     string
	}{
		{1, 1, "0101"},
		{32, 40, "3240"},
		{8, 10, "0810"},
	}

	for _, c := range cases {
		if got := hexLocation(c.col, c.row); got != c.want {
			t.Errorf("hexLocation(%d, %d) = %q, want %q", c.col, c.row, got, c.want)
		}
	}
}

// TestRollSystemPresentBoundaries runs rollSystemPresent many times per
// Density and confirms the true-rate lands within a few points of Book 3
// p.13's own documented percentage for that density — the practical way
// to verify a "roll N dice, <= target" implementation matches a named
// probability without hand-deriving each one.
func TestRollSystemPresentBoundaries(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))

	presentCounts := map[Density]int{}

	const trials = 20000

	for range trials {
		for d := range densityTable {
			if rollSystemPresent(r, d) {
				presentCounts[d]++
			}
		}
	}

	wantPct := map[Density]float64{
		DensityExtraGalactic: 1,
		DensityRift:          3,
		DensitySparse:        17,
		DensityScattered:     33,
		DensityStandard:      50,
		DensityDense:         66,
		DensityCluster:       83,
		// Core's true rate is 2D6<=11 = 35/36 = 97.2%, not the book's own
		// (internally inconsistent, verified against the page image)
		// quoted 91% — see densityTable's doc comment.
		DensityCore: 97.2,
	}

	for d, want := range wantPct {
		got := 100 * float64(presentCounts[d]) / trials
		if got < want-3 || got > want+3 {
			t.Errorf("%s: rollSystemPresent true %.1f%% of %d trials, want ~%.0f%% (Book 3 p.13)", d, got, trials, want)
		}
	}
}

func TestGenerateSectorHexCount(t *testing.T) {
	t.Parallel()

	r := dice.RollerFromSeed(1)
	sec := GenerateSector(r, "Test", DensityStandard)

	if len(sec.Hexes) != sectorWidth*sectorHeight {
		t.Fatalf("GenerateSector: len(Hexes) = %d, want %d", len(sec.Hexes), sectorWidth*sectorHeight)
	}
}

// TestGenerateSectorStampsSectorAndHex confirms every populated hex's
// StarSystem (and its mainworld World) carries the sector's own Name and
// that hex's own Location — the previously-unused Sector/Hex fields both
// types already had.
func TestGenerateSectorStampsSectorAndHex(t *testing.T) {
	t.Parallel()

	r := dice.RollerFromSeed(1)
	sec := GenerateSector(r, "Spinward Marches", DensityStandard)

	checked := 0

	for _, hex := range sec.Hexes {
		if hex.System == nil {
			continue
		}

		checked++

		if hex.System.Sector != "Spinward Marches" || hex.System.Hex != hex.Location {
			t.Fatalf("hex %s: StarSystem.Sector/.Hex = %q/%q, want %q/%q",
				hex.Location, hex.System.Sector, hex.System.Hex, "Spinward Marches", hex.Location)
		}

		mw := hex.System.Orbits[hex.System.MainworldOrbit].World
		if mw.Sector != "Spinward Marches" || mw.Hex != hex.Location {
			t.Fatalf("hex %s: mainworld World.Sector/.Hex = %q/%q, want %q/%q",
				hex.Location, mw.Sector, mw.Hex, "Spinward Marches", hex.Location)
		}
	}

	if checked == 0 {
		t.Fatal("no populated hexes found — test can't verify anything, try a different seed")
	}
}

func TestGenerateSectorDeterminism(t *testing.T) {
	t.Parallel()

	r1 := dice.New(rand.NewPCG(55, 55))
	r2 := dice.New(rand.NewPCG(55, 55))

	sec1 := GenerateSector(r1, "Test", DensityStandard)
	sec2 := GenerateSector(r2, "Test", DensityStandard)

	if !reflect.DeepEqual(sec1, sec2) {
		t.Error("identical seeds produced different sectors")
	}
}

// TestGenerateSectorDensityAffectsPopulation confirms Core (91%) produces
// visibly more populated hexes than Rift (3%) for the same seed sequence
// — a coarse sanity check that density actually drives presence, without
// re-deriving the exact statistical bounds TestRollSystemPresentBoundaries
// already covers.
func TestGenerateSectorDensityAffectsPopulation(t *testing.T) {
	t.Parallel()

	countPopulated := func(sec Sector) int {
		n := 0

		for _, hex := range sec.Hexes {
			if hex.System != nil {
				n++
			}
		}

		return n
	}

	rift := countPopulated(GenerateSector(dice.RollerFromSeed(2), "Test", DensityRift))
	core := countPopulated(GenerateSector(dice.RollerFromSeed(2), "Test", DensityCore))

	if rift >= core {
		t.Errorf("Rift produced %d populated hexes, Core produced %d — want Rift well below Core", rift, core)
	}
}
