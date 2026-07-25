package character

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

func TestContinueSoldierOutcomeBoundaries(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		roll, end int
		want      bool
	}{
		{"natural 2 always succeeds, even against a target it would otherwise beat easily", 2, 1, true},
		{"exactly at End succeeds", 7, 7, true},
		{"one above End fails", 8, 7, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := continueSoldierOutcome(c.roll, c.end); got != c.want {
				t.Errorf("continueSoldierOutcome(%d, %d) = %v, want %v", c.roll, c.end, got, c.want)
			}
		})
	}
}

func TestResolveSoldierCareerNeverQualifiedReturnsZeroTermsCareer(t *testing.T) {
	t.Parallel()

	upp := UPP{}
	r := dice.New(rand.NewPCG(1, 1))

	career, _ := ResolveSoldierCareer(r, upp)

	if career.Name != SoldierCareerName {
		t.Errorf("career.Name = %q, want %q", career.Name, SoldierCareerName)
	}

	if !career.HasRank {
		t.Error("career.HasRank = false, want true")
	}

	if len(career.Terms) != 0 {
		t.Errorf("career.Terms = %v, want empty (Begin against Str=0 always fails)", career.Terms)
	}
}

// TestResolveSoldierCareerRespectsMaxTermsCap uses a provably immortal
// fixture: Str=End=Int=20 (all three of Soldier's own Risk & Reward
// positions), high enough that even the worst-case combined Mod
// (Branch's own max 2 + Operations' own max 3 = 5) still leaves a target
// of 15 — above 2D6's own maximum of 12.
func TestResolveSoldierCareerRespectsMaxTermsCap(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{20, 0, 20, 20, 8, 0}}

	for _, seed := range []uint64{1, 2, 3} {
		r := dice.New(rand.NewPCG(seed, seed))

		career, _ := ResolveSoldierCareer(r, upp)
		if len(career.Terms) != maxCareerTerms {
			t.Errorf("seed %d: len(career.Terms) = %d, want %d", seed, len(career.Terms), maxCareerTerms)
		}
	}
}

// TestResolveSoldierCareerCCRotation reuses the immortal fixture: C1=C3=C4
// tied at 20, so highestOf's first-wins-on-tie makes the rotation fully
// predictable — the pool cycles C1, C3, C4 in that order (the order
// soldierRiskRewardPositions itself lists them).
func TestResolveSoldierCareerCCRotation(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{20, 0, 20, 20, 8, 0}}
	r := dice.New(rand.NewPCG(7, 7))

	career, _ := ResolveSoldierCareer(r, upp)

	want := []Position{C1, C3, C4, C1, C3, C4, C1, C3, C4, C1, C3, C4, C1, C3}
	if len(career.Terms) != len(want) {
		t.Fatalf("len(career.Terms) = %d, want %d", len(career.Terms), len(want))
	}

	for i, w := range want {
		if got := career.Terms[i].ControllingCharacteristic; got != w {
			t.Errorf("term %d: ControllingCharacteristic = %v, want %v", i+1, got, w)
		}
	}
}

// TestResolveSoldierCareerPersistsCharacteristicReduction mirrors
// TestResolveMarineCareerPersistsCharacteristicReduction: confirms
// ResolveSoldierCareer threads ResolveSoldierTerm's own returned UPP
// into the next term, not a stale copy.
func TestResolveSoldierCareerPersistsCharacteristicReduction(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{6, 0, 6, 6, 8, 0}}
	r := dice.New(rand.NewPCG(23, 29))

	career, finalUPP := ResolveSoldierCareer(r, upp)

	if len(career.Terms) == 0 {
		t.Fatal("career.Terms is empty, want at least one term")
	}

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

// TestResolveSoldierCareerStopsOnDisabled mirrors
// TestResolveMarineCareerStopsOnDisabled — the regression test for the
// missing-Disabled-check class of bug.
func TestResolveSoldierCareerStopsOnDisabled(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{5, 0, 5, 12, 8, 0}}
	r := dice.New(rand.NewPCG(3, 3))

	sawDisabled := false

	for range 500 {
		career, _ := ResolveSoldierCareer(r, upp)

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

// TestResolveSoldierCareerStopsOnDeath mirrors
// TestResolveMarineCareerStopsOnDeath: End=Int=0 makes Risk against
// either fatal (resolveRisk's own clamp reduces to exactly 0), so
// whenever the rotation lands on C3 or C4, that term is fatal. The
// structural invariant under test is that Dead is always the career's
// own last term.
func TestResolveSoldierCareerStopsOnDeath(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 0, 0, 0, 8, 0}}
	r := dice.New(rand.NewPCG(41, 43))

	sawDeath := false

	for range 500 {
		career, _ := ResolveSoldierCareer(r, upp)

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
		t.Error("Dead never occurred across 500 trials with End=Int=0 — fixture assumption broke")
	}
}
