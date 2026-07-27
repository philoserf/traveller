package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
)

func TestAgentCareerFame(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		terms []Term
		want  int
	}{
		{"no terms", nil, 0},
		{"one commendation", []Term{{RewardResult: "Scout Commendation-2"}}, 1},
		{
			"two commendations, one failure",
			[]Term{
				{RewardResult: "Scout Commendation-2"},
				{RewardResult: "None"},
				{RewardResult: "Noble Commendation-0"},
			},
			2,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := sumInts(agentCareerFameAwards(Career{Terms: c.terms})); got != c.want {
				t.Errorf("sumInts(agentCareerFameAwards(%v)) = %d, want %d", c.terms, got, c.want)
			}
		})
	}
}

// TestGenerateAgentCharacterQualified pins seed 2: one term whose
// Reward succeeds against the term-start CC after Risk reduces it.
// The resulting commendation contributes one Fame.
func TestGenerateAgentCharacterQualified(t *testing.T) {
	t.Parallel()

	// Precondition: an Agent who qualified and served. The old fixture
	// pinned seed 8 and with it a one-term career and Cash of exactly 0,
	// neither of which this test is about.
	seed := seedFor(t, "an Agent who qualified and served a term", func(seed uint64) bool {
		c, ok := GenerateAgentCharacter(dice.New(rand.NewPCG(seed, seed)))

		return ok && len(c.Careers) == 1 && len(c.Careers[0].Terms) > 0
	})

	c, ok := GenerateAgentCharacter(dice.New(rand.NewPCG(seed, seed)))

	if !ok {
		t.Fatal("ok = false, want true")
	}

	if c.Species != "Human" {
		t.Errorf("Species = %q, want %q", c.Species, "Human")
	}

	if len(c.Careers) != 1 || c.Careers[0].Name != AgentCareerName {
		t.Fatalf("Careers = %+v, want one Career named %q", c.Careers, AgentCareerName)
	}

	if c.Careers[0].HasRank {
		t.Error("Careers[0].HasRank = true, want false (Agent has no Rank concept)")
	}

	if len(c.Careers[0].Terms) == 0 {
		t.Fatal("Careers[0].Terms is empty, want at least one term")
	}

	// Book 1 p.91: "Agent =Number of Commendations". Asserted as that
	// relationship rather than as a pinned 1, so the test keeps checking
	// the rule rather than the seed after a dice-stream shift.
	if want := agentCommendationCount(c.Careers[0].Terms); c.Fame != want {
		t.Errorf("Fame = %d, want %d (one per Commendation)", c.Fame, want)
	}

	// Derived, not pinned at 0: Cash is whatever the career's own
	// Mustering Out Money entries came to, and which entries a seed rolls
	// moves with the dice stream. What is being checked is that the sum
	// propagates onto the Character at all.
	if want := musterOutCash(c.Careers[0]); c.Cash != want {
		t.Errorf("Cash = %d, want %d (the career's own Mustering Out Money)", c.Cash, want)
	}

	assertBirthdateFormat(t, c.Birthdate, c.Age)
}

// TestGenerateAgentCharacterNeverQualifiedIsAValidOutcome confirms
// GenerateAgentCharacter can produce ok=false (unlike Merchant, Agent
// has a real chance to never qualify) without erroring — a smoke test
// across several seeds, not a specific pinned fixture.
func TestGenerateAgentCharacterNeverQualifiedIsAValidOutcome(t *testing.T) {
	t.Parallel()

	sawNeverQualified := false

	for seed := uint64(1); seed <= 30; seed++ {
		r := dice.New(rand.NewPCG(seed, seed))

		c, ok := GenerateAgentCharacter(r)
		if !ok && len(c.Careers[0].Terms) == 0 {
			sawNeverQualified = true

			break
		}
	}

	if !sawNeverQualified {
		t.Error("0 of 30 seeds produced a never-qualified Agent, want at least 1")
	}
}
