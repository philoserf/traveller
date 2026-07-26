package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// TestResolveScholarCareerNeverQualifiedReturnsZeroTermsCareer mirrors
// TestResolveRogueCareerNeverQualifiedReturnsZeroTermsCareer's own
// shape: Edu 0 guarantees BeginScholar's own roll fails.
func TestResolveScholarCareerNeverQualifiedReturnsZeroTermsCareer(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 0, 8}}
	r := dice.New(rand.NewPCG(1, 1))

	career, _ := ResolveScholarCareer(r, upp)

	if career.Name != ScholarCareerName {
		t.Errorf("career.Name = %q, want %q", career.Name, ScholarCareerName)
	}

	if !career.HasRank {
		t.Error("career.HasRank = false, want true (Scholar has a real Rank progression)")
	}

	if len(career.Terms) != 0 {
		t.Errorf("career.Terms = %v, want empty (BeginScholar's own roll fails against Edu 0)", career.Terms)
	}
}

// TestResolveScholarCareerRespectsMaxTermsCap mirrors
// TestResolveRogueCareerFixedCCNeverRotates's own seed-confirmed-by-
// inspection style: seed 1 against an all-12 UPP was confirmed directly
// to reach the full 14-term cap.
func TestResolveScholarCareerRespectsMaxTermsCap(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{12, 12, 12, 12, 12, 12}}
	r := dice.New(rand.NewPCG(1, 1))

	career, _ := ResolveScholarCareer(r, upp)

	if len(career.Terms) != maxCareerTerms {
		t.Errorf("len(career.Terms) = %d, want %d (maxCareerTerms, seed 1 confirmed to reach the cap)",
			len(career.Terms), maxCareerTerms)
	}
}

// TestResolveScholarCareerPersistsRiskReductionOntoReturnedUPP is the
// regression test for a code-review-caught bug: ResolveScholarCareer
// used to return only Career, silently discarding resolveCareerLoop's
// own final UPP — so a Wounded/Disabled term's characteristic reduction
// (applied inside ResolveScholarTerm via
// upp.Characteristics[ccPos] = reducedCC) never reached the character's
// final UPP or Mustering Out. Seed 1 against an all-8 UPP was confirmed
// by direct inspection to reduce C3 (End) from 8 to 7 by career's end.
func TestResolveScholarCareerPersistsRiskReductionOntoReturnedUPP(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}
	for seed := uint64(1); seed <= 1000; seed++ {
		career, finalUPP := ResolveScholarCareer(dice.New(rand.NewPCG(seed, seed)), upp)

		for _, term := range career.Terms {
			pos := term.ControllingCharacteristic
			if (term.RiskResult == Wounded || term.RiskResult == Disabled) &&
				finalUPP.Characteristics[pos] < upp.Characteristics[pos] {
				return
			}
		}
	}

	t.Fatal("no seed in [1,1000] retained a net Wounded/Disabled reduction after Personal improvements")
}

// TestContinueScholarHasNoNaturalRollException confirms continueScholar
// is a plain rollAgainstTarget wrapper — unlike Marine's own natural-2
// or Rogue's own natural-12 override, Scholar's own Continue has no
// exception documented anywhere (the box and the Master Checklist both
// describe a plain roll) — a natural 2 against a target of 0 must still
// fail here, the opposite of continueMarineOutcome's own behavior.
func TestContinueScholarHasNoNaturalRollException(t *testing.T) {
	t.Parallel()

	r := dice.New(
		rand.NewPCG(21, 21),
	) // seed 21's own first TwoD6() = 2, confirmed directly (Rogue's own plan-file research)

	if got := continueScholar(r, 0, 0); got {
		t.Error(
			"continueScholar(edu=0, mod=0) with a natural 2 = true, want false (no natural-roll exception for Scholar)",
		)
	}
}
