package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
)

// countNonLaterEducationTerms is termsServed's own definition, computed
// independently (a plain loop, not a call to the function under test),
// so the sweeps below check termsServed against its own contract rather
// than against len(terms) — which stopped being equivalent once Rogue
// (#113 item 5's pilot) started producing real LaterEducation terms.
func countNonLaterEducationTerms(terms []Term) int {
	n := 0

	for _, t := range terms {
		if !t.LaterEducation {
			n++
		}
	}

	return n
}

// TestTermsServedExcludesLaterEducationAcrossGeneratedCharacters is the
// termsServed sweep (#113 item 5, stage 1) generalized past its own
// original zero-behavior-change premise: it no longer assumes no
// generator ever sets LaterEducation (Rogue now does), only that
// termsServed always agrees with counting non-LaterEducation terms
// directly, for every career on every generated character, standalone
// or chained.
func TestTermsServedExcludesLaterEducationAcrossGeneratedCharacters(t *testing.T) {
	t.Parallel()

	generators := []func(r *dice.Roller) (Character, bool){
		GenerateAgentCharacter,
		GenerateCitizenCharacter,
		GenerateScoutCharacter,
		GenerateEntertainerCharacter,
		GenerateMarineCharacter,
		GenerateNobleCharacter,
		GenerateSoldierCharacter,
		GenerateMerchantCharacter,
		GenerateScholarCharacter,
		GenerateRogueCharacter,
		GenerateSpacerCharacter,
	}

	checked, sawLaterEducation := 0, false

	for _, generate := range generators {
		for seed := uint64(1); seed <= 200; seed++ {
			c, _ := generate(dice.New(rand.NewPCG(seed, seed)))

			for _, career := range c.Careers {
				if got, want := termsServed(career.Terms), countNonLaterEducationTerms(career.Terms); got != want {
					t.Errorf("%s: termsServed(Terms) = %d, want %d", career.Name, got, want)
				}

				for _, term := range career.Terms {
					if term.LaterEducation {
						sawLaterEducation = true
					}
				}

				checked++
			}
		}
	}

	if checked == 0 {
		t.Fatal("no careers checked — generators list is empty or produced nothing")
	}

	if !sawLaterEducation {
		t.Error("no LaterEducation term appeared across 200 Rogue seeds — the pilot wiring may not be firing")
	}
}

// TestTermsServedExcludesLaterEducationAcrossCareerChains is the same
// proof for career_chain.go's own accumulator, since chained careers
// are assembled differently than standalone ones.
func TestTermsServedExcludesLaterEducationAcrossCareerChains(t *testing.T) {
	t.Parallel()

	chains := [][]string{
		{"scout"},
		{"citizen", "craftsman"},
		{"marine", "functionary"},
		{"scholar", "rogue", "noble"},
		{"spacer", "merchant", "agent"},
	}

	checked := 0

	for _, careerNames := range chains {
		for seed := uint64(1); seed <= 100; seed++ {
			c, _, err := GenerateCareerChainCharacter(dice.New(rand.NewPCG(seed, seed)), careerNames, 0)
			if err != nil {
				t.Fatalf("GenerateCareerChainCharacter(%v) error: %v", careerNames, err)
			}

			for _, career := range c.Careers {
				if got, want := termsServed(career.Terms), countNonLaterEducationTerms(career.Terms); got != want {
					t.Errorf("%v: %s: termsServed(Terms) = %d, want %d", careerNames, career.Name, got, want)
				}

				checked++
			}
		}
	}

	if checked == 0 {
		t.Fatal("no careers checked — chains list is empty or produced nothing")
	}
}
