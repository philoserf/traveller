package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// TestContinueEntertainerIsAPlainRoll mirrors
// TestContinueScholarHasNoNaturalRollException's own reasoning: no
// special-case exception is documented anywhere for Entertainer's own
// Continue, so a natural 2 against a target of 0 must still fail.
func TestContinueEntertainerIsAPlainRoll(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(21, 21)) // seed 21's own first TwoD6() = 2, confirmed this session

	if got := continueEntertainer(r, 0); got {
		t.Error("continueEntertainer(fame=0) with a natural 2 = true, want false (no natural-roll exception)")
	}
}

// TestResolveEntertainerCareerNeverQualifiedStillSetsFame is the
// regression test for this slice's own central judgment call: Fame is
// rolled "Before Begin" in Book 1's own ordering, so it's returned even
// when BeginEntertainer's own roll fails — seed 6 against an all-8 UPP
// was confirmed by direct inspection to fail Begin while still
// producing Fame 7.
func TestResolveEntertainerCareerNeverQualifiedStillSetsFame(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}
	r := dice.New(rand.NewPCG(6, 6))

	career, fame := ResolveEntertainerCareer(r, upp)

	if career.Name != EntertainerCareerName {
		t.Errorf("career.Name = %q, want %q", career.Name, EntertainerCareerName)
	}

	if career.HasRank {
		t.Error("career.HasRank = true, want false (Entertainer has no Rank concept)")
	}

	if len(career.Terms) != 0 {
		t.Errorf("career.Terms = %v, want empty (BeginEntertainer's own roll fails)", career.Terms)
	}

	if career.Specialty == "" {
		t.Error("career.Specialty is empty, want a Specialty to have been chosen before Begin was even rolled")
	}

	if fame != 7 {
		t.Errorf("fame = %d, want 7 (the initial 2D roll, set regardless of Begin's own outcome)", fame)
	}
}

// TestResolveEntertainerCareerRespectsMaxTermsCap mirrors every other
// career's own cap test — seed 22 against an all-8 UPP was confirmed by
// direct inspection to reach the full 14-term cap.
func TestResolveEntertainerCareerRespectsMaxTermsCap(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}
	r := dice.New(rand.NewPCG(22, 22))

	career, _ := ResolveEntertainerCareer(r, upp)

	if len(career.Terms) != maxCareerTerms {
		t.Errorf("len(career.Terms) = %d, want %d (maxCareerTerms, seed 22 confirmed to reach the cap)",
			len(career.Terms), maxCareerTerms)
	}
}
