package character

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

func TestContinueSpacerOutcomeBoundaries(t *testing.T) {
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

			if got := continueSpacerOutcome(c.roll, c.str); got != c.want {
				t.Errorf("continueSpacerOutcome(%d, %d) = %v, want %v", c.roll, c.str, got, c.want)
			}
		})
	}
}

func TestResolveSpacerCareerNeverQualifiedReturnsZeroTermsCareer(t *testing.T) {
	t.Parallel()

	upp := UPP{}
	r := dice.New(rand.NewPCG(1, 1))

	career, _ := ResolveSpacerCareer(r, upp)

	if career.Name != SpacerCareerName {
		t.Errorf("career.Name = %q, want %q", career.Name, SpacerCareerName)
	}

	if !career.HasRank {
		t.Error("career.HasRank = false, want true")
	}

	if len(career.Terms) != 0 {
		t.Errorf("career.Terms = %v, want empty (Begin against Int=0 always fails)", career.Terms)
	}
}

// TestResolveSpacerCareerWithBudgetCommissionBypassesBegin mirrors
// TestResolveMarineCareerWithBudgetCommissionBypassesBegin (#113): a
// Commission substitutes for BeginSpacer's own roll, entering term 1
// already Commissioned at Officer1 and granting Spacer's own O1 auto
// skill ("Astrogator," spacer_promotion.go) — the same skill p.61's own
// worked example has Eneri receive on his NOTC Commission.
func TestResolveSpacerCareerWithBudgetCommissionBypassesBegin(t *testing.T) {
	t.Parallel()

	upp := UPP{}
	r := dice.New(rand.NewPCG(1, 1))

	career, _ := resolveSpacerCareerWithBudget(r, upp, maxCareerTerms, &agingSimulation{}, true)

	if len(career.Terms) == 0 {
		t.Fatal("career.Terms is empty, want a Commission to bypass BeginSpacer")
	}

	if !career.Terms[0].Commissioned {
		t.Error("career.Terms[0].Commissioned = false, want true")
	}

	want := SkillLevel{Name: "Astrogator", Level: 1, Kind: Skill}
	if !slices.Contains(career.Terms[0].SkillsAwarded, want) {
		t.Errorf("career.Terms[0].SkillsAwarded = %+v, want to contain %+v (Spacer Officer1 auto skill)",
			career.Terms[0].SkillsAwarded, want)
	}
}

// TestResolveSpacerCareerRespectsMaxTermsCap uses a provably immortal
// fixture: Str=Dex=Int=20 (all three of Spacer's own Risk & Reward
// positions), high enough that even the worst-case combined Mod
// (Branch's own max 2 + Operations' own max 3 = 5) still leaves a
// target of 15 — above 2D6's own maximum of 12.
func TestResolveSpacerCareerRespectsMaxTermsCap(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{20, 20, 0, 20, 8, 0}}

	for _, seed := range []uint64{1, 2, 3} {
		r := dice.New(rand.NewPCG(seed, seed))

		career, _ := ResolveSpacerCareer(r, upp)
		if len(career.Terms) != maxCareerTerms {
			t.Errorf("seed %d: len(career.Terms) = %d, want %d", seed, len(career.Terms), maxCareerTerms)
		}
	}
}

// TestResolveSpacerCareerGrantsStartingRankAutoSkill is the regression
// for #53: R1 Spacehand's own "Fighter" Auto Skill must land on term 1
// even though a fresh career's first term is never itself a
// Commissioned/Promoted event.
func TestResolveSpacerCareerGrantsStartingRankAutoSkill(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{20, 20, 0, 20, 8, 0}}
	r := dice.New(rand.NewPCG(1, 1))

	career, _ := ResolveSpacerCareer(r, upp)

	if len(career.Terms) == 0 {
		t.Fatal("career.Terms is empty (fixture assumption broke)")
	}

	if !slices.ContainsFunc(career.Terms[0].SkillsAwarded, func(s SkillLevel) bool { return s.Name == "Fighter" }) {
		t.Errorf("term 1 SkillsAwarded = %+v, want a Fighter entry (R1 Spacehand's own starting Auto Skill)",
			career.Terms[0].SkillsAwarded)
	}
}

// TestResolveSpacerCareerCCRotation reuses the immortal fixture: C1=C2=C4
// tied at 20, so highestOf's first-wins-on-tie makes the rotation fully
// predictable — the pool cycles C1, C2, C4 in that order (the order
// spacerRiskRewardPositions itself lists them).
func TestResolveSpacerCareerCCRotation(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{20, 20, 0, 20, 8, 0}}
	r := dice.New(rand.NewPCG(7, 7))

	career, _ := ResolveSpacerCareer(r, upp)

	assertCCRotationCycles(t, career.Terms, spacerRiskRewardPositions)
}

// TestResolveSpacerCareerPersistsCharacteristicReduction mirrors
// TestResolveSoldierCareerPersistsCharacteristicReduction.
func TestResolveSpacerCareerPersistsCharacteristicReduction(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{6, 6, 0, 6, 8, 0}}
	r := dice.New(rand.NewPCG(23, 29))

	career, finalUPP := ResolveSpacerCareer(r, upp)

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

// TestResolveSpacerCareerStopsOnDisabled mirrors
// TestResolveSoldierCareerStopsOnDisabled.
func TestResolveSpacerCareerStopsOnDisabled(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{5, 5, 0, 12, 8, 0}}
	r := dice.New(rand.NewPCG(3, 3))

	sawDisabled := false

	for range 500 {
		career, _ := ResolveSpacerCareer(r, upp)

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

// TestResolveSpacerCareerStopsOnDeath mirrors
// TestResolveSoldierCareerStopsOnDeath, with one adjustment: Spacer's
// own Begin check targets Int (C4), not Str — Int must stay high enough
// to guarantee Begin succeeds (and to plausibly survive Risk when
// rotated there too), so Dex=0 alone carries the guaranteed-fatal role
// instead of the Soldier fixture's own two zeroed positions.
func TestResolveSpacerCareerStopsOnDeath(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 0, 0, 12, 8, 0}}
	r := dice.New(rand.NewPCG(41, 43))

	sawDeath := false

	for range 500 {
		career, _ := ResolveSpacerCareer(r, upp)

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
		t.Error("Dead never occurred across 500 trials with Dex=Int=0 — fixture assumption broke")
	}
}
