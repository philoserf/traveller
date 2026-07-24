package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
)

func TestGenerateUPPDeterminism(t *testing.T) {
	t.Parallel()

	r1 := dice.New(rand.NewPCG(99, 99))
	r2 := dice.New(rand.NewPCG(99, 99))

	u1 := GenerateUPP(r1)
	u2 := GenerateUPP(r2)

	if u1 != u2 {
		t.Fatalf("identical seeds produced different UPPs: %s vs %s", u1, u2)
	}
}

// TestGenerateUPPCharacteristicsInRange confirms every one of the six
// Characteristics lands in 2D6's true range (2-12) over many trials.
func TestGenerateUPPCharacteristicsInRange(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 2))

	for range 1000 {
		u := GenerateUPP(r)

		for i, c := range u.Characteristics {
			if c < 2 || c > 12 {
				t.Fatalf("Characteristics[%d] = %d, want 2-12 (2D6 range)", i, c)
			}
		}
	}
}

// TestGenerateUPPDefersObscureCharacteristics confirms Sanity and
// Psionics are left untouched — Book 1 p.57: both are "created only as
// needed," not part of standard character generation.
func TestGenerateUPPDefersObscureCharacteristics(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(5, 6))
	u := GenerateUPP(r)

	if u.Sanity != 0 {
		t.Errorf("Sanity = %d, want 0 (deferred)", u.Sanity)
	}

	if u.Psionics != 0 {
		t.Errorf("Psionics = %d, want 0 (deferred)", u.Psionics)
	}
}
