package character

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

func TestAgentUndercoverCareerNamesMatchImplementedCareers(t *testing.T) {
	t.Parallel()

	want := []string{
		"Scout", "Marine", "Soldier", "Spacer", "Rogue",
		"Scholar", "Entertainer", "Merchant", "Noble",
	}

	if !slices.Equal(agentUndercoverCareerNames, want) {
		t.Errorf("agentUndercoverCareerNames = %v, want %v", agentUndercoverCareerNames, want)
	}

	if len(agentUndercoverSkillTables) != len(want) {
		t.Errorf("agentUndercoverSkillTables has %d entries, want %d", len(agentUndercoverSkillTables), len(want))
	}

	for _, name := range want {
		if _, ok := agentUndercoverSkillTables[name]; !ok {
			t.Errorf("agentUndercoverSkillTables is missing an entry for %q", name)
		}
	}
}

func TestAgentSkillTableMatchesBook1P83(t *testing.T) {
	t.Parallel()

	want := [7][6]string{
		{"Str", "C2", "C3", "Int", "C5", "C6"},
		{"Major", "Major", "Minor", "Minor", "One Trade", "One Trade"},
		{"Zero-G", "Vacc Suit", "Pilot", "Starship Skill", "Gunner", "Sensors"},
		{"Survey", "Survival", "Hostile Environment", "Animals", "Bureaucrat", "Navigation"},
		{"Fighter", "Soldier Skill", "Flyer", "Stealth", "Gunner", "Streetwise"},
		{"Any Knowledge", "Admin", "Language", "Starship Skill", "Comms", "One Trade"},
		{"One Art", "One Science", "Athlete", "Medic", "Seafarer", "One Trade"},
	}

	if agentSkillTable != want {
		t.Errorf("agentSkillTable = %v, want %v", agentSkillTable, want)
	}
}

// TestResolveSkillCellTreatsAnyKnowledgeAsUnresolvable is the regression
// test for the shared career_generate.go fix this slice needed: without
// it, "Any Knowledge" would fall through to resolveSkillCell's own
// default case and be granted as a literal skill named "Any Knowledge".
func TestResolveSkillCellTreatsAnyKnowledgeAsUnresolvable(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))

	_, ok := resolveSkillCell(r, agentSkillTable, 5, 0) // column 6, row 1: "Any Knowledge"
	if ok {
		t.Error("resolveSkillCell(\"Any Knowledge\") ok = true, want false (unresolvable)")
	}
}

// TestBeginAgent covers both outcomes at End 8. Seeds 1/4 were confirmed
// by direct inspection to roll a success/failure respectively.
func TestBeginAgent(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		r := dice.New(rand.NewPCG(1, 1))
		if !BeginAgent(r, 8) {
			t.Error("BeginAgent(c3=8) = false, want true (seed 1 rolls a success)")
		}
	})

	t.Run("failure", func(t *testing.T) {
		t.Parallel()

		r := dice.New(rand.NewPCG(4, 4))
		if BeginAgent(r, 8) {
			t.Error("BeginAgent(c3=8) = true, want false (seed 4 rolls a failure)")
		}
	})
}

func TestAgentCommendationCount(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		terms []Term
		want  int
	}{
		{"no terms", nil, 0},
		{"one commendation", []Term{{RewardResult: "Scout Commendation-2"}}, 1},
		{"one failure", []Term{{RewardResult: "None"}}, 0},
		{"a Dead term with no RewardResult at all", []Term{{RiskResult: Dead}}, 0},
		{
			"mixed",
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

			if got := agentCommendationCount(c.terms); got != c.want {
				t.Errorf("agentCommendationCount(%v) = %d, want %d", c.terms, got, c.want)
			}
		})
	}
}

// TestRollAgentUndercoverCareer confirms it always returns one of the
// nine known career names.
func TestRollAgentUndercoverCareer(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))

	for range 100 {
		name := rollAgentUndercoverCareer(r)
		if !slices.Contains(agentUndercoverCareerNames, name) {
			t.Fatalf("rollAgentUndercoverCareer returned %q, not one of %v", name, agentUndercoverCareerNames)
		}
	}
}

var uppAgent88 = UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}

// The ResolveAgentTerm fixtures below were seed-hunted against an
// all-8 UPP and verified by direct inspection of the resulting Term
// before being pinned here — not assumed from the mechanic alone.

func TestResolveAgentTermRiskSuccessRewardSuccess(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))

	term, _ := ResolveAgentTerm(r, uppAgent88, C1)

	if term.RiskResult != Unharmed {
		t.Fatalf("RiskResult = %v, want Unharmed", term.RiskResult)
	}

	if term.UndercoverCareer != "Noble" {
		t.Errorf("UndercoverCareer = %q, want %q", term.UndercoverCareer, "Noble")
	}

	// N = C - R = 8 - 8 = 0.
	if term.RewardResult != "Noble Commendation-0" {
		t.Errorf("RewardResult = %q, want %q", term.RewardResult, "Noble Commendation-0")
	}
}

func TestResolveAgentTermRiskSuccessRewardFailure(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(2, 2))

	term, _ := ResolveAgentTerm(r, uppAgent88, C1)

	if term.RiskResult != Unharmed {
		t.Fatalf("RiskResult = %v, want Unharmed", term.RiskResult)
	}

	if term.RewardResult != "None" {
		t.Errorf("RewardResult = %q, want %q", term.RewardResult, "None")
	}

	if term.UndercoverCareer != "Marine" {
		t.Errorf("UndercoverCareer = %q, want %q", term.UndercoverCareer, "Marine")
	}
}

// TestResolveAgentTermRiskFailureReducesCharacteristic confirms Agent
// reuses the exact universal Risk mechanic (a real physical
// characteristic reduction, unlike Entertainer's own Talent) — Wound
// Badges are genuinely counted, matching Merchant's own precedent.
func TestResolveAgentTermRiskFailureReducesCharacteristic(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(12, 12))

	term, updatedUPP := ResolveAgentTerm(r, uppAgent88, C1)

	if term.RiskResult != Wounded {
		t.Fatalf("RiskResult = %v, want Wounded", term.RiskResult)
	}

	if updatedUPP.Characteristics[C1] >= uppAgent88.Characteristics[C1] {
		t.Errorf("updatedUPP.Characteristics[C1] = %v, want less than %v (a real reduction)",
			updatedUPP.Characteristics[C1], uppAgent88.Characteristics[C1])
	}

	// N = C - R = 8 - 2 = 6.
	if term.RewardResult != "Soldier Commendation-6" {
		t.Errorf("RewardResult = %q, want %q", term.RewardResult, "Soldier Commendation-6")
	}
}
