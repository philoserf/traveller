package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
)

// TestTermsServedMatchesLenTermsUntilLaterEducationExists is the
// zero-behavior-change proof for the termsServed sweep (#113 item 5,
// stage 1): no generator sets Term.LaterEducation yet, so
// termsServed(terms) must equal len(terms) for every career on every
// generated character, standalone or chained. Once Later Education
// actually elects a term (a later stage of #113), this test's own
// premise breaks and it must be narrowed to non-education terms —
// that's expected, not a regression, and the comment should say so
// then.
func TestTermsServedMatchesLenTermsUntilLaterEducationExists(t *testing.T) {
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

	checked := 0

	for _, generate := range generators {
		for seed := uint64(1); seed <= 200; seed++ {
			c, _ := generate(dice.New(rand.NewPCG(seed, seed)))

			for _, career := range c.Careers {
				if got, want := termsServed(career.Terms), len(career.Terms); got != want {
					t.Errorf("%s: termsServed(Terms) = %d, want %d (len)", career.Name, got, want)
				}

				checked++
			}
		}
	}

	if checked == 0 {
		t.Fatal("no careers checked — generators list is empty or produced nothing")
	}
}

// TestTermsServedMatchesLenTermsAcrossCareerChains is the same proof
// for career_chain.go's own accumulator, since chained careers are
// assembled differently than standalone ones.
func TestTermsServedMatchesLenTermsAcrossCareerChains(t *testing.T) {
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
				if got, want := termsServed(career.Terms), len(career.Terms); got != want {
					t.Errorf("%v: %s: termsServed(Terms) = %d, want %d (len)", careerNames, career.Name, got, want)
				}

				checked++
			}
		}
	}

	if checked == 0 {
		t.Fatal("no careers checked — chains list is empty or produced nothing")
	}
}
