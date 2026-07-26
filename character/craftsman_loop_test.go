package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
)

func TestResolveCraftsmanCareerWithBudgetNeverQualifiesWithoutPrerequisite(t *testing.T) {
	t.Parallel()

	for seed := uint64(1); seed <= 10; seed++ {
		r := dice.New(rand.NewPCG(seed, seed))

		career, _ := resolveCraftsmanCareerWithBudget(r, uppCraftsman12, maxCareerTerms, segmentContext{})
		if len(career.Terms) != 0 {
			t.Fatalf("seed=%d: len(career.Terms) = %d, want 0 (no prior skills at all)", seed, len(career.Terms))
		}
	}
}

// TestResolveCraftsmanCareerWithBudgetRespectsATighterBudget confirms
// Craftsman's own loop honors the -age-derived budget exactly like
// every other career, using the same high-skill fixture
// TestResolveCraftsmanTermSuccess already confirmed produces a
// guaranteed-successful Masterpiece attempt every term (Continue target
// 2*6=12 is also automatic on 2D6), making a 14-term uncapped run and a
// tighter budget both fully deterministic.
func TestResolveCraftsmanCareerWithBudgetRespectsATighterBudget(t *testing.T) {
	t.Parallel()

	ctx := segmentContext{SkillsSoFar: craftsmanHighSkillFixture}

	const budget = 3

	career, _ := resolveCraftsmanCareerWithBudget(dice.New(rand.NewPCG(1, 1)), uppCraftsman12, budget, ctx)
	if len(career.Terms) != budget {
		t.Errorf("len(career.Terms) = %d, want %d", len(career.Terms), budget)
	}
}

func TestResolveCraftsmanCareerWithBudgetThreadsGrowingHeldSkills(t *testing.T) {
	t.Parallel()

	ctx := segmentContext{SkillsSoFar: craftsmanHighSkillFixture}

	career, finalUPP := resolveCraftsmanCareerWithBudget(
		dice.New(rand.NewPCG(1, 1)),
		uppCraftsman12,
		maxCareerTerms,
		ctx,
	)

	if len(career.Terms) == 0 {
		t.Fatal("len(career.Terms) = 0, want at least 1")
	}

	if career.HasRank {
		t.Error("HasRank = true, want false (Craftsman has no rank, Book 1 p.65)")
	}

	wantUPP := uppCraftsman12
	for _, term := range career.Terms {
		wantUPP = applyPersonalAwards(wantUPP, term.SkillsAwarded)
	}

	if finalUPP != wantUPP {
		t.Errorf("finalUPP = %+v, want Personal awards applied as %+v", finalUPP, wantUPP)
	}
}
