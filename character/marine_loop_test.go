package character

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

func TestContinueMarineOutcomeBoundaries(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		roll, str int
		want      bool
	}{
		{"natural 2 always succeeds, even against a target it would otherwise beat easily", 2, 1, true},
		{"exactly at Str succeeds", 7, 7, true},
		{"one above Str fails", 8, 7, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := continueMarineOutcome(c.roll, c.str); got != c.want {
				t.Errorf("continueMarineOutcome(%d, %d) = %v, want %v", c.roll, c.str, got, c.want)
			}
		})
	}
}

func TestResolveMarineCareerNeverQualifiedReturnsZeroTermsCareer(t *testing.T) {
	t.Parallel()

	upp := UPP{}
	r := dice.New(rand.NewPCG(1, 1))

	career, _ := ResolveMarineCareer(r, upp)

	if career.Name != MarineCareerName {
		t.Errorf("career.Name = %q, want %q", career.Name, MarineCareerName)
	}

	if !career.HasRank {
		t.Error("career.HasRank = false, want true")
	}

	if len(career.Terms) != 0 {
		t.Errorf("career.Terms = %v, want empty (Begin against Str=0 always fails)", career.Terms)
	}
}

// TestResolveMarineCareerRespectsMaxTermsCap uses a provably immortal
// fixture: Str=Int=20, high enough that even the worst-case combined Mod
// (Branch's own max 2 + Operations' own max 3 = 5, subtracted from the
// Risk target) still leaves a target of 15 — above 2D6's own maximum of
// 12, so Risk (and, at 20, Continue against the unreduced Str) can never
// fail. 12 alone (Scout's own immortal-fixture value) is NOT enough
// here, unlike Scout: Marine's own Risk target is cc-mod, not cc+0, so a
// large enough combined Mod can push it below 12 even starting from 12.
func TestResolveMarineCareerRespectsMaxTermsCap(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{20, 0, 0, 20, 8, 0}}

	for _, seed := range []uint64{1, 2, 3} {
		r := dice.New(rand.NewPCG(seed, seed))

		career, _ := ResolveMarineCareer(r, upp)
		if len(career.Terms) != maxCareerTerms {
			t.Errorf("seed %d: len(career.Terms) = %d, want %d", seed, len(career.Terms), maxCareerTerms)
		}
	}
}

// TestResolveMarineCareerCCRotation reuses the immortal fixture:
// C1=C4=20 tied, so highestOf's first-wins-on-tie makes the rotation
// fully predictable, mirroring TestResolveScoutCareerCCRotationAcrossTerms.
func TestResolveMarineCareerCCRotation(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{20, 0, 0, 20, 8, 0}}
	r := dice.New(rand.NewPCG(7, 7))

	career, _ := ResolveMarineCareer(r, upp)

	assertCCRotationCycles(t, career.Terms, marineRiskRewardPositions)
}

// TestResolveMarineCareerPersistsCharacteristicReduction mirrors
// TestResolveScoutCareerPersistsCharacteristicReduction: confirms
// ResolveMarineCareer threads ResolveMarineTerm's own returned UPP into
// the next term, not a stale copy.
func TestResolveMarineCareerPersistsCharacteristicReduction(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{6, 0, 0, 6, 8, 0}}
	r := dice.New(rand.NewPCG(23, 29))

	career, finalUPP := ResolveMarineCareer(r, upp)

	if len(career.Terms) == 0 {
		t.Fatal("career.Terms is empty, want at least one term")
	}

	// If any term actually reduced its own Controlling Characteristic
	// (RiskResult Wounded or Disabled), finalUPP must reflect the LAST
	// such reduction on that position, not the original UPP's value —
	// proving the loop threads UPP forward rather than discarding it.
	for i, term := range slices.Backward(career.Terms) {
		if term.RiskResult != Wounded && term.RiskResult != Disabled {
			continue
		}

		if finalUPP.Characteristics[term.ControllingCharacteristic] == upp.Characteristics[term.ControllingCharacteristic] {
			t.Errorf("finalUPP.Characteristics[%v] unchanged from original despite a %v result in term %d",
				term.ControllingCharacteristic, term.RiskResult, i+1)
		}

		return
	}
}

// TestResolveMarineCareerStopsOnDisabled is the regression test for the
// exact class of bug Scout's own original ResolveScoutCareer draft had
// (missing the Disabled check, allowing a Disabled Marine to keep
// serving): a low-CC fixture makes Disabled plausible across many
// trials; whenever it occurs, it must be the final term.
func TestResolveMarineCareerStopsOnDisabled(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{5, 0, 0, 12, 8, 0}}
	r := dice.New(rand.NewPCG(3, 3))

	sawDisabled := false

	for range 500 {
		career, _ := ResolveMarineCareer(r, upp)

		for i, term := range career.Terms {
			if term.RiskResult == Disabled && i != len(career.Terms)-1 {
				sawDisabled = true

				break
			}
		}
	}

	if sawDisabled {
		t.Error("a Disabled term was followed by another term — Disabled should mandatorily end the career")
	}
}

// TestResolveMarineCareerStopsOnDeath uses C1=8 (Begin, no Retry for
// Marine, succeeds a plausible fraction of trials: P(2D<=8)=26/36) and
// C4=0 (Risk against a 0 characteristic always fails and always reduces
// to exactly 0, per resolveRisk's own clamp — riskOutcome then always
// classifies that as Dead). Whenever a term's own Controlling
// Characteristic lands on C4, that term is fatal; the structural
// invariant under test is that Dead, wherever it occurs, is always the
// career's own last term — not a specific death rate, which the
// Begin-without-Retry/CC-rotation interaction makes awkward to pin
// exactly.
func TestResolveMarineCareerStopsOnDeath(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 0, 0, 0, 8, 0}}
	r := dice.New(rand.NewPCG(41, 43))

	sawDeath := false

	for range 500 {
		career, _ := ResolveMarineCareer(r, upp)

		for i, term := range career.Terms {
			if term.RiskResult != Dead {
				continue
			}

			sawDeath = true

			if i != len(career.Terms)-1 {
				t.Errorf(
					"a Dead term was followed by another term (term %d of %d) — Dead should mandatorily end the career",
					i+1,
					len(career.Terms),
				)
			}
		}
	}

	if !sawDeath {
		t.Error("Dead never occurred across 500 trials with C4=0 — fixture assumption broke")
	}
}
