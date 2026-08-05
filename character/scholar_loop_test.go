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

	// Soc 0 as well as Edu 0: p.76 lets a Scholar waive an adverse
	// "Position" result on a Check Soc, so a failed Begin at Soc 8
	// would simply be waived and the character would qualify after
	// all. A 2D check cannot come in at or below 0.
	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 0, 0}}
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

// TestResolveScholarCareerLoopUsesCurrentEduNotEntryEdu is the
// regression test for #138: the term-resolve and Continue closures
// inside resolveScholarCareerWithBudget used to close over edu as
// captured before the term loop began, instead of deriving it from the
// closure's own upp parameter (which resolveCareerLoop threads forward
// with any Personal-row Edu boost already applied, applyPersonalAwards).
// Since Tenure requires tier == 3 && edu >= 10, a character who enters
// at Edu 9 (tier 1, auto-begin) could never reach Tenure under the stale
// value even after a mid-career Personal-row Edu boost to 10 — Book 1
// p.76's own "Promotion beyond Scholar3 not possible without Tenure"
// then permanently caps them at Assistant Professor. Seed 1 against Edu
// 9 was confirmed by direct comparison against the pre-fix
// implementation: the character gains a Personal Edu boost to 10 mid-
// career and, only with the fix, reaches a term where Tenure is rolled
// for and succeeds.
func TestResolveScholarCareerLoopUsesCurrentEduNotEntryEdu(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 9, 8}}
	r := dice.New(rand.NewPCG(1, 1))

	career, _ := resolveScholarCareerWithBudget(r, upp, maxCareerTerms, &agingSimulation{}, Education{})

	tenureGranted := false

	for _, term := range career.Terms {
		if term.TenureGranted {
			tenureGranted = true
		}
	}

	if !tenureGranted {
		t.Error(
			"no term granted Tenure, want at least one " +
				"(seed 1's mid-career Personal Edu boost to 10 should make tier-3 Tenure reachable)",
		)
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
