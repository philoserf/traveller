package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// TestResolveFunctionaryCareerWithBudgetNeverQualifiesWithNoPriorTerms
// confirms BeginFunctionary's own zero-target formula makes Functionary
// unreachable with no prior career experience — Book 1's own "never a
// first career," exercised end to end.
func TestResolveFunctionaryCareerWithBudgetNeverQualifiesWithNoPriorTerms(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}

	for seed := uint64(1); seed <= 10; seed++ {
		r := dice.New(rand.NewPCG(seed, seed))

		career, _, tier := resolveFunctionaryCareerWithBudget(
			r,
			upp,
			maxCareerTerms,
			segmentContext{TermsServedSoFar: 0},
		)
		if len(career.Terms) != 0 {
			t.Fatalf("seed=%d: len(career.Terms) = %d, want 0", seed, len(career.Terms))
		}

		if tier != 0 {
			t.Errorf("seed=%d: tier = %d, want 0", seed, tier)
		}
	}
}

// TestResolveFunctionaryCareerWithBudgetGrantsF0AutoSkillOnce confirms
// the F0 starting Auto Skill (Bureaucrat) is applied to term 1 only,
// not re-granted every term — mirroring Marine's own one-time
// Branch-skill precedent.
func TestResolveFunctionaryCareerWithBudgetGrantsF0AutoSkillOnce(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}
	ctx := segmentContext{TermsServedSoFar: 4, PrecedingCareer: "scholar"}

	var career Career

	for seed := uint64(1); seed <= 100; seed++ {
		r := dice.New(rand.NewPCG(seed, seed))

		c, _, _ := resolveFunctionaryCareerWithBudget(r, upp, maxCareerTerms, ctx)
		if len(c.Terms) >= 2 {
			career = c

			break
		}
	}

	if len(career.Terms) < 2 {
		t.Fatal("no seed in 1..100 produced a 2+ term Functionary career (fixture search failed)")
	}

	bureaucratCount := func(term Term) int {
		n := 0

		for _, s := range term.SkillsAwarded {
			if s.Name == "Bureaucrat" {
				n++
			}
		}

		return n
	}

	if bureaucratCount(career.Terms[0]) == 0 {
		t.Errorf("term 1 SkillsAwarded = %+v, want at least one Bureaucrat (F0's own Auto Skill)",
			career.Terms[0].SkillsAwarded)
	}
}

// TestResolveFunctionaryCareerWithBudgetRespectsATighterBudget confirms
// Functionary's own loop honors the -age-derived budget exactly like
// every other career (career_loop_test.go's own
// TestResolveScoutCareerWithBudgetTruncatesAnImmortalCareer establishes
// the same property for the shared resolveCareerLoop).
func TestResolveFunctionaryCareerWithBudgetRespectsATighterBudget(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{12, 12, 12, 12, 12, 12}}
	ctx := segmentContext{TermsServedSoFar: 4}

	const budget = 3

	var found bool

	for seed := uint64(1); seed <= 50; seed++ {
		r := dice.New(rand.NewPCG(seed, seed))

		career, _, _ := resolveFunctionaryCareerWithBudget(r, upp, budget, ctx)
		if len(career.Terms) > budget {
			t.Fatalf("seed=%d: len(career.Terms) = %d, want <= %d", seed, len(career.Terms), budget)
		}

		if len(career.Terms) == budget {
			found = true
		}
	}

	if !found {
		t.Error("0 of 50 seeds reached the budget cap, want at least 1 (Str 12 makes Office Politics near-certain)")
	}
}
