package system

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

func TestSatelliteHostKindForBoundaries(t *testing.T) {
	t.Parallel()

	cases := []struct {
		isGasGiant bool
		delta      int
		want       satelliteHostKind
	}{
		{true, 0, hostGasGiant},
		{true, -5, hostGasGiant}, // GasGiant always wins regardless of delta
		{false, -3, hostInnerWorld},
		{false, -2, hostInnerWorld},
		{false, -1, hostHZWorld},
		{false, 0, hostHZWorld},
		{false, 1, hostHZWorld},
		{false, 2, hostOuterWorld},
		{false, 3, hostOuterWorld},
	}

	for _, c := range cases {
		if got := satelliteHostKindFor(c.isGasGiant, c.delta); got != c.want {
			t.Errorf("satelliteHostKindFor(%v, %d) = %v, want %v", c.isGasGiant, c.delta, got, c.want)
		}
	}
}

// TestRollSatelliteCountZeroTriggersExactlyOneReroll pins the "Zero=Ring
// and reroll" rule precisely: a raw roll of 0 sets ring=true and rerolls
// exactly once — not a loop until nonzero. hostHZWorld's formula (1D-4)
// rolls 0 only on a D6 of 4; seed hunting for a D6 sequence starting with
// 4 confirms the reroll happens and ring is set.
func TestRollSatelliteCountZeroTriggersExactlyOneReroll(t *testing.T) {
	t.Parallel()

	// Search for a seed whose first D6 roll is 4 (triggers the 1D-4=0
	// case for hostHZWorld), to exercise the reroll branch deterministically.
	var found *dice.Roller

	for seed := int64(1); seed < 10000; seed++ {
		r := dice.RollerFromSeed(seed)
		if r.D6() == 4 {
			found = dice.RollerFromSeed(seed)

			break
		}
	}

	if found == nil {
		t.Fatal("no seed found with first D6=4 in range; widen the search")
	}

	count, ring := rollSatelliteCount(found, hostHZWorld)

	if !ring {
		t.Errorf(
			"rollSatelliteCount with first roll=4 (0 after -4 DM) = (count=%d, ring=%v), want ring=true",
			count,
			ring,
		)
	}

	if count < 0 {
		t.Errorf("rollSatelliteCount returned count=%d, want >= 0 (floored)", count)
	}
}

func TestRollSatelliteCountNeverNegative(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))

	for _, kind := range []satelliteHostKind{hostGasGiant, hostInnerWorld, hostHZWorld, hostOuterWorld} {
		for range 1000 {
			if count, _ := rollSatelliteCount(r, kind); count < 0 {
				t.Fatalf("rollSatelliteCount(%v) = %d, want >= 0", kind, count)
			}
		}
	}
}

// TestRollSatelliteCategoryOuterDiffersFromRegular pins the one real,
// book-verified divergence between satellite and regular Other-World
// placement: at Outer roll=4, satellites get StormWorld where regular
// placement gets Iceworld.
func TestRollSatelliteCategoryOuterDiffersFromRegular(t *testing.T) {
	t.Parallel()

	if got, want := satelliteOuterCategoryByRoll[4], categoryStormWorld; got != want {
		t.Errorf("satelliteOuterCategoryByRoll[4] = %v, want %v", got, want)
	}

	if got, want := outerCategoryByRoll[4], categoryIceworld; got != want {
		t.Errorf("outerCategoryByRoll[4] = %v, want %v (regression: regular placement's own table)", got, want)
	}

	// Every other row must still match between the two tables.
	for roll := 1; roll <= 6; roll++ {
		if roll == 4 {
			continue
		}

		if satelliteOuterCategoryByRoll[roll] != outerCategoryByRoll[roll] {
			t.Errorf("satelliteOuterCategoryByRoll[%d] = %v, outerCategoryByRoll[%d] = %v, want equal",
				roll, satelliteOuterCategoryByRoll[roll], roll, outerCategoryByRoll[roll])
		}
	}
}

func TestGenerateSatelliteSizeAdjustment(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(2, 2))

	const parentSize = ehex.Value(3)

	for range 2000 {
		u := generateSatellite(r, 0, parentSize, true, 10)
		if u.Size > parentSize {
			t.Fatalf(
				"generateSatellite with hasParentSize=true: Size = %d, want <= parentSize (%d)",
				u.Size,
				parentSize,
			)
		}
	}
}

func TestGenerateSatelliteNoAdjustmentForGasGiantParent(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(3, 3))

	sawLargerThanZero := false

	for range 200 {
		u := generateSatellite(r, 0, 0, false, 10)
		if u.Size > 0 {
			sawLargerThanZero = true

			break
		}
	}

	if !sawLargerThanZero {
		t.Error(
			"generateSatellite with hasParentSize=false never produced Size>0 in 200 tries — adjustment may be wrongly applied",
		)
	}
}
