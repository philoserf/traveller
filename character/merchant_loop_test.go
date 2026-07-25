package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// TestContinueMerchantIsAPlainRoll mirrors
// TestContinueScholarHasNoNaturalRollException's own reasoning: no
// special-case exception is documented anywhere for Merchant's own
// Continue, so a natural 2 against a target of 0 must still fail.
func TestContinueMerchantIsAPlainRoll(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(21, 21)) // seed 21's own first TwoD6() = 2, confirmed this session

	if got := continueMerchant(r, 0); got {
		t.Error("continueMerchant(str=0) with a natural 2 = true, want false (no natural-roll exception)")
	}
}

// TestResolveMerchantCareerAlwaysProducesATerm confirms the central
// finding of this slice: Merchant is the only career this session with
// no "never qualified" outcome at all — BeginMerchant can never fail,
// so ResolveMerchantCareer always produces at least one term, checked
// across a spread of seeds and UPPs, not just one lucky case.
func TestResolveMerchantCareerAlwaysProducesATerm(t *testing.T) {
	t.Parallel()

	upps := []UPP{
		{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}},
		{Characteristics: [6]ehex.Value{1, 1, 1, 1, 1, 1}},
		{Characteristics: [6]ehex.Value{0, 0, 0, 0, 0, 0}},
		{Characteristics: [6]ehex.Value{12, 12, 12, 12, 12, 12}},
	}

	for _, upp := range upps {
		for seed := uint64(1); seed <= 20; seed++ {
			r := dice.New(rand.NewPCG(seed, seed))

			career, _, _, tier := ResolveMerchantCareer(r, upp)
			if len(career.Terms) == 0 {
				t.Fatalf("seed %d, upp %v: career.Terms is empty, want at least one term (BeginMerchant never fails)",
					seed, upp)
			}

			if tier < 0 || tier > 6 {
				t.Fatalf("seed %d, upp %v: tier = %d, want 0-6", seed, upp, tier)
			}
		}
	}
}

// TestResolveMerchantCareerPersistsRiskReductionOntoReturnedUPP is the
// regression test for a code-review-caught bug: ResolveMerchantCareer
// used to return only (Career, isOfficer, tier), silently discarding
// resolveCareerLoop's own final UPP — so a Wounded/Disabled term's real
// physical characteristic reduction (unlike Entertainer's own Talent)
// never reached buildMerchantCharacter's own Mustering Out or final
// UPP. Seed 1 against an all-8 UPP was confirmed by direct inspection
// to reduce C1 (Str) from 8 to 4 within a single term.
func TestResolveMerchantCareerPersistsRiskReductionOntoReturnedUPP(t *testing.T) {
	t.Parallel()

	career, finalUPP, _, _ := ResolveMerchantCareer(dice.New(rand.NewPCG(1, 1)), upp88)

	if len(career.Terms) == 0 {
		t.Fatal("career.Terms is empty, want at least one term")
	}

	sawReduction := false

	for _, term := range career.Terms {
		if term.RiskResult != Wounded && term.RiskResult != Disabled {
			continue
		}

		pos := term.ControllingCharacteristic
		if finalUPP.Characteristics[pos] < upp88.Characteristics[pos] {
			sawReduction = true
		}
	}

	if !sawReduction {
		t.Fatal("no Wounded/Disabled term's own reduction is reflected in the returned finalUPP")
	}
}

// TestResolveMerchantCareerRespectsMaxTermsCap mirrors every other
// career's own cap test, additionally confirming the cap is actually
// reachable (not just never exceeded) across a spread of seeds.
func TestResolveMerchantCareerRespectsMaxTermsCap(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{12, 12, 12, 12, 12, 12}}

	atCap := false

	for seed := uint64(1); seed <= 30; seed++ {
		r := dice.New(rand.NewPCG(seed, seed))

		career, _, isOfficer, _ := ResolveMerchantCareer(r, upp)
		if isOfficer && len(career.Terms) == 0 {
			t.Fatalf("seed %d: isOfficer=true with zero Terms, want at least one term for a real Officer state", seed)
		}

		if len(career.Terms) > maxCareerTerms {
			t.Fatalf("seed %d: len(career.Terms) = %d, want <= %d", seed, len(career.Terms), maxCareerTerms)
		}

		if len(career.Terms) == maxCareerTerms {
			atCap = true
		}
	}

	if !atCap {
		t.Errorf("0 of 30 trials reached maxCareerTerms, want at least 1 (cap never observed as reachable)")
	}
}
