package character

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// TestCommandCollegeDrawsFromTheRightAcademy is p.61's "two skill levels
// from the appropriate Military or Naval Academy", and the ruling that
// Marine takes the Naval one — the pool a career draws from is the whole
// mechanical consequence of that reading, so it is asserted directly
// rather than through generated output.
func TestCommandCollegeDrawsFromTheRightAcademy(t *testing.T) {
	t.Parallel()

	cases := []struct {
		career string
		want   schoolSet
	}{
		{SoldierCareerName, schoolMilitaryAcademy},
		{MarineCareerName, schoolNavalAcademy},
		{SpacerCareerName, schoolNavalAcademy},
	}

	for _, c := range cases {
		got, ok := commandCollegeSchool(c.career)
		if !ok || got != c.want {
			t.Errorf("%s draws from %05b (ok=%v), want %05b", c.career, got, ok, c.want)
		}
	}

	// Every other career has no Command College at all, and must not
	// silently draw from one.
	for _, career := range []string{"Scout", NobleCareerName, MerchantCareerName} {
		if _, ok := commandCollegeSchool(career); ok {
			t.Errorf("%s was given a Command College", career)
		}
	}
}

// TestCommandCollegeAwardsTwoSkillsOnSuccess covers p.59's "2x Skill-1"
// Provides column and p.61's terminal failure in one place, using
// characteristics that force each outcome: a 2D6 check can never exceed
// 20 and can never come in at or below 1.
func TestCommandCollegeAwardsTwoSkillsOnSuccess(t *testing.T) {
	t.Parallel()

	t.Run("success awards two skills and keeps the career", func(t *testing.T) {
		t.Parallel()

		upp := UPP{Characteristics: [6]ehex.Value{0, 0, 0, 20, 20, 0}}

		skills, mayContinue := resolveCommandCollege(dice.New(rand.NewPCG(3, 3)), upp, MarineCareerName)
		if !mayContinue {
			t.Fatal("mayContinue = false on a check that cannot fail")
		}

		if len(skills) != commandCollegeSkillGrants {
			t.Fatalf("awarded %d skills, want %d", len(skills), commandCollegeSkillGrants)
		}

		naval := skillsForSchool(schoolNavalAcademy, false)

		for _, s := range skills {
			if s.Level != 1 {
				t.Errorf("%s awarded at level %d, want 1", s.Name, s.Level)
			}

			if !slices.Contains(naval, s.Name) {
				t.Errorf("%q is not in the Naval Academy pool", s.Name)
			}
		}
	})

	t.Run("failure ends the career", func(t *testing.T) {
		t.Parallel()

		upp := UPP{Characteristics: [6]ehex.Value{0, 0, 0, 1, 1, 0}}

		skills, mayContinue := resolveCommandCollege(dice.New(rand.NewPCG(3, 3)), upp, MarineCareerName)
		if mayContinue {
			t.Error("mayContinue = true on a check that cannot succeed")
		}

		if len(skills) != 0 {
			t.Errorf("awarded %d skills on a failure, want none", len(skills))
		}
	})
}

// TestCommandCollegeChecksTheBetterOfIntAndEdu is p.59's "Int or Edu"
// and "Int or C5" being the same open choice — resolved to the higher of
// the two, the treatment every other "X or Y" check in this package
// gets. A character strong in only one of them must still pass.
func TestCommandCollegeChecksTheBetterOfIntAndEdu(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name string
		upp  UPP
	}{
		{"Int carries it", UPP{Characteristics: [6]ehex.Value{0, 0, 0, 20, 2, 0}}},
		{"Edu carries it", UPP{Characteristics: [6]ehex.Value{0, 0, 0, 2, 20, 0}}},
	} {
		if _, ok := resolveCommandCollege(dice.New(rand.NewPCG(9, 9)), c.upp, SoldierCareerName); !ok {
			t.Errorf("%s: Command College failed despite a characteristic of 20", c.name)
		}
	}
}

// TestCommandCollegeFiresOnceAtO4 is p.61's trigger — "promoted to
// Officer4" — and the once-only property that keeps a character who
// serves five more terms as an officer from attending five more times.
func TestCommandCollegeFiresOnceAtO4(t *testing.T) {
	t.Parallel()

	// rankState (career_rank.go) walks Commissioned/Promoted flags rather
	// than reading Rank strings, so the tier has to be built the way a
	// real career reaches it: commission to O1, then promote.
	officerTerms := func(promotions int) []Term {
		terms := make([]Term, 0, 1+promotions)
		terms = append(terms, Term{Commissioned: true})

		for range promotions {
			terms = append(terms, Term{Promoted: true})
		}

		return terms
	}

	enlisted, officers := len(marineEnlistedRankNames), len(marineOfficerRankNames)

	if _, tier := rankState(officerTerms(3), enlisted, officers); tier != commandCollegeOfficerTier {
		t.Fatalf("fixture reached O%d, want O%d — the trigger below would be testing nothing",
			tier, commandCollegeOfficerTier)
	}

	if commandCollegeDue(officerTerms(2), enlisted, officers, 0) {
		t.Error("Command College fired at O3, before p.61's own O4 trigger")
	}

	if !commandCollegeDue(officerTerms(3), enlisted, officers, 0) {
		t.Error("Command College did not fire at O4")
	}

	if commandCollegeDue(officerTerms(3), enlisted, officers, 1) {
		t.Error("Command College fired a second time")
	}

	// An Enlisted character promoted just as far must not attend: p.61
	// restricts Command College to officers.
	enlistedTerms := []Term{{Promoted: true}, {Promoted: true}, {Promoted: true}}
	if commandCollegeDue(enlistedTerms, enlisted, officers, 0) {
		t.Error("Command College fired for an Enlisted character at M4")
	}
}

// TestCommandCollegeCostsNoDiceBelowO4 holds the reproducibility line:
// the hook runs after every successful Continue in three careers, so a
// character who never reaches O4 must draw nothing for it.
func TestCommandCollegeCostsNoDiceBelowO4(t *testing.T) {
	t.Parallel()

	hook, drain := armedForcesCommandCollege(MarineCareerName)

	upp := UPP{Characteristics: [6]ehex.Value{0, 0, 0, 20, 20, 0}}
	terms := []Term{{Rank: marineRankName(true, 3)}}

	spent := dice.New(rand.NewPCG(4, 4))
	if !hook(spent, upp, terms) {
		t.Fatal("the hook ended a career that never reached O4")
	}

	if got := drain(); len(got) != 0 {
		t.Errorf("drained %d skills below O4, want none", len(got))
	}

	untouched := dice.New(rand.NewPCG(4, 4))
	if spent.TwoD6() != untouched.TwoD6() {
		t.Error("the Command College hook consumed dice for a character below O4")
	}
}
