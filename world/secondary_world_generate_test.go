package world

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

func TestCapPopulation(t *testing.T) {
	t.Parallel()

	if got, want := capPopulation(8, 5), ehex.Value(5); got != want {
		t.Errorf("capPopulation(8, 5) = %d, want %d", got, want)
	}

	if got, want := capPopulation(3, 5), ehex.Value(3); got != want {
		t.Errorf("capPopulation(3, 5) = %d, want %d", got, want)
	}
}

func TestGenerateBigWorldSizeFloor(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))

	for range 1000 {
		u := generateBigWorld(r, 10)
		if u.Size < 7 {
			t.Fatalf("generateBigWorld: Size = %d, want >= 7 (2D+7 floor)", u.Size)
		}
	}
}

func TestGenerateInfernoFixedFields(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(2, 2))

	for range 1000 {
		u := generateInferno(r)

		if u.Starport != StarportNone {
			t.Fatalf("generateInferno: Starport = %v, want StarportNone", u.Starport)
		}

		if u.Size < 7 {
			t.Fatalf("generateInferno: Size = %d, want >= 7 (6+1D floor)", u.Size)
		}

		if u.Atmosphere != 0 || u.Hydrographics != 0 || u.Population != 0 || u.Government != 0 || u.Law != 0 ||
			u.TechLevel != 0 {
			t.Fatalf("generateInferno: %+v, want every field but Starport/Size fixed at 0", u)
		}
	}
}

func TestGenerateRadWorldFixedFields(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(3, 3))

	for range 1000 {
		u := generateRadWorld(r)

		if u.Population != 0 || u.Government != 0 || u.Law != 0 || u.TechLevel != 0 {
			t.Fatalf("generateRadWorld: %+v, want Population/Government/Law/TechLevel fixed at 0", u)
		}
	}
}

func TestGenerateWorldletSizeCeiling(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(4, 4))

	for range 1000 {
		u := generateWorldlet(r, 10)
		if u.Size > 3 {
			t.Fatalf("generateWorldlet: Size = %d, want <= 3 (1D-3, D6 max 6)", u.Size)
		}
	}
}

func TestGeneratePlanetoidWorldFixedFields(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(5, 5))

	for range 1000 {
		u := generatePlanetoidWorld(r, 10)
		if u.Size != 0 || u.Atmosphere != 0 || u.Hydrographics != 0 {
			t.Fatalf("generatePlanetoidWorld: %+v, want Size/Atmosphere/Hydrographics fixed at 0", u)
		}

		if !slices.Contains(DeriveTradeCodes(u), AsteroidBelt) {
			t.Fatalf("generatePlanetoidWorld: DeriveTradeCodes(%+v) doesn't contain AsteroidBelt", u)
		}
	}
}

// TestSecondaryWorldsRespectMaxPopulation runs every population-rolling
// category against a low maxPopulation ceiling and checks it's honored.
func TestSecondaryWorldsRespectMaxPopulation(t *testing.T) {
	t.Parallel()

	const maxPopulation = ehex.Value(2)

	r := dice.New(rand.NewPCG(6, 6))

	generators := map[string]func(*dice.Roller, ehex.Value) UWP{
		"Hospitable": generateHospitableWorld,
		"BigWorld":   generateBigWorld,
		"Worldlet":   generateWorldlet,
		"Iceworld":   generateIceworld,
		"InnerWorld": generateInnerWorld,
		"StormWorld": generateStormWorld,
	}

	for name, gen := range generators {
		for range 500 {
			if u := gen(r, maxPopulation); u.Population > maxPopulation {
				t.Fatalf("%s: Population = %d, want <= maxPopulation (%d)", name, u.Population, maxPopulation)
			}
		}
	}
}

func TestRollSecondaryWorldCategoryBoundary(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(7, 7))

	for range 100 {
		if got := rollSecondaryWorldCategory(r, 0); got != categoryInferno && got != categoryInnerWorld &&
			got != categoryBigWorld && got != categoryStormWorld && got != categoryRadWorld && got != categoryHospitable {
			t.Fatalf("rollSecondaryWorldCategory(delta=0) = %v, want an Inner/HZ-column category", got)
		}

		if got := rollSecondaryWorldCategory(r, 1); got != categoryWorldlet && got != categoryIceworld &&
			got != categoryBigWorld && got != categoryRadWorld {
			t.Fatalf("rollSecondaryWorldCategory(delta=1) = %v, want an Outer-column category", got)
		}
	}
}
