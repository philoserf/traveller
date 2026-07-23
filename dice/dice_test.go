package dice_test

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
)

func TestD6Bounds(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 2))

	for range 10000 {
		v := r.D6()
		if v < 1 || v > 6 {
			t.Fatalf("D6() = %d, want 1..6", v)
		}
	}
}

func TestTwoD6Bounds(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(3, 4))

	for range 10000 {
		v := r.TwoD6()
		if v < 2 || v > 12 {
			t.Fatalf("TwoD6() = %d, want 2..12", v)
		}
	}
}

func TestFluxBounds(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(5, 6))

	for range 10000 {
		v := r.Flux()
		if v < -5 || v > 5 {
			t.Fatalf("Flux() = %d, want -5..5", v)
		}
	}
}

func TestUniformBounds(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(7, 8))

	for range 10000 {
		v := r.Uniform(9)
		if v < 1 || v > 9 {
			t.Fatalf("Uniform(9) = %d, want 1..9", v)
		}
	}
}

func TestDeterminism(t *testing.T) {
	t.Parallel()

	r1 := dice.New(rand.NewPCG(42, 42))
	r2 := dice.New(rand.NewPCG(42, 42))

	for i := range 100 {
		a, b := r1.TwoD6(), r2.TwoD6()
		if a != b {
			t.Fatalf("roll %d: r1.TwoD6()=%d, r2.TwoD6()=%d, want equal for identical seeds", i, a, b)
		}
	}
}

func TestResolveSeedExplicitNonZeroUnchanged(t *testing.T) {
	t.Parallel()

	seed := int64(12345)
	if got := dice.ResolveSeed(&seed); got != 12345 {
		t.Errorf("ResolveSeed(&12345) = %d, want 12345 (unchanged)", got)
	}
}

func TestResolveSeedExplicitZeroUnchanged(t *testing.T) {
	t.Parallel()

	// The whole point of the *int64 signature: an explicit 0 must be
	// honored, not silently treated the same as "no seed provided."
	seed := int64(0)
	if got := dice.ResolveSeed(&seed); got != 0 {
		t.Errorf("ResolveSeed(&0) = %d, want 0 (explicit zero honored)", got)
	}
}

func TestResolveSeedNilDerivesNonZero(t *testing.T) {
	t.Parallel()

	got := dice.ResolveSeed(nil)
	if got == 0 {
		t.Error("ResolveSeed(nil) = 0, want a non-zero time-derived seed")
	}
}

func TestRollerFromSeedDeterminism(t *testing.T) {
	t.Parallel()

	r1 := dice.RollerFromSeed(999)
	r2 := dice.RollerFromSeed(999)

	if r1.TwoD6() != r2.TwoD6() {
		t.Error("RollerFromSeed(999) produced different first rolls across two calls, want identical")
	}
}
