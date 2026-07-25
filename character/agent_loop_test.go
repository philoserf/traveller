package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// TestContinueAgentIsAPlainRoll mirrors
// TestContinueMerchantIsAPlainRoll's own reasoning: no special-case
// exception is documented anywhere for Agent's own Continue, so a
// natural 2 against a target of 0 must still fail.
func TestContinueAgentIsAPlainRoll(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(21, 21)) // seed 21's own first TwoD6() = 2, confirmed this session

	if got := continueAgent(r, 0, 0); got {
		t.Error("continueAgent(str=0, mod=0) with a natural 2 = true, want false (no natural-roll exception)")
	}
}

// TestResolveAgentCareerNeverQualifiedReturnsZeroTermsCareer mirrors
// every other career's own never-qualified test — seed 4 against an
// all-8 UPP was confirmed by direct inspection to fail Begin.
func TestResolveAgentCareerNeverQualifiedReturnsZeroTermsCareer(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(4, 4))

	career, _ := ResolveAgentCareer(r, uppAgent88)

	if career.Name != AgentCareerName {
		t.Errorf("career.Name = %q, want %q", career.Name, AgentCareerName)
	}

	if career.HasRank {
		t.Error("career.HasRank = true, want false (Agent has no Rank concept)")
	}

	if len(career.Terms) != 0 {
		t.Errorf("career.Terms = %v, want empty (BeginAgent's own roll fails)", career.Terms)
	}
}

// TestResolveAgentCareerRespectsMaxTermsCap mirrors every other
// career's own cap test — seed 1 against an all-12 UPP was confirmed by
// direct inspection to reach the full 14-term cap.
func TestResolveAgentCareerRespectsMaxTermsCap(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{12, 12, 12, 12, 12, 12}}
	r := dice.New(rand.NewPCG(1, 1))

	career, _ := ResolveAgentCareer(r, upp)

	if len(career.Terms) != maxCareerTerms {
		t.Errorf("len(career.Terms) = %d, want %d (maxCareerTerms, seed 1 confirmed to reach the cap)",
			len(career.Terms), maxCareerTerms)
	}
}
