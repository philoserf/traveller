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

			if got := agentCareerFame(Career{Terms: c.terms}); got != c.want {
				t.Errorf("agentCareerFame(%v) = %d, want %d", c.terms, got, c.want)
			}
		})
	}
}

// TestGenerateAgentCharacterQualified pins seed 2: one term, Risk fails
// Disabled (RiskResult 2), Reward fails (RewardResult "None", so
// agentCareerFame's own Commendation count is 0) — confirmed by direct
// inspection (character/zzdebug_test.go trace, not kept in the repo):
// a Disabled last term doubles scoutMusterOutRollCount to 2 rolls for a
// single term, landing one Money entry ("Cr25,000", Cash 25000) and one
// Benefits entry ("C3 +1").
func TestGenerateAgentCharacterQualified(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(2, 2))

	c, ok := GenerateAgentCharacter(r)

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

	if len(c.Careers[0].Terms) != 1 {
		t.Fatalf("len(Careers[0].Terms) = %d, want 1", len(c.Careers[0].Terms))
	}

	if c.Fame != 0 {
		t.Errorf("Fame = %d, want 0", c.Fame)
	}

	if c.Cash != 25000 {
		t.Errorf("Cash = %d, want 25000", c.Cash)
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
