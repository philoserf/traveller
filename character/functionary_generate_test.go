package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

func TestFunctionarySkillTableMatchesBook1P87(t *testing.T) {
	t.Parallel()

	want := [7][6]string{
		{"Str", "C2", "C3", "Int", "C5", "C6"},
		{"Major", "Major", "Minor", "Minor", "One Trade", "One Trade"},
		{"High-G", "Vacc Suit", "Driver", "Flyer", "Navigation", "Seafarer"},
		{"One Trade", "One Art", "One Science", "Any Skill", "Bureaucrat", "Leader"},
		{"Advocate", "Broker", "Trader", "Teacher", "One Trade", "Driver"},
		{"Advocate", "Comms", "Language", "Admin", "Bureaucrat", "Comms"},
		{"One Art", "One Science", "Athlete", "Designer", "Seafarer", "One Trade"},
	}

	if functionarySkillTable != want {
		t.Errorf("functionarySkillTable = %v, want %v", functionarySkillTable, want)
	}
}

// TestResolveSkillCellTreatsAnySkillAsCitizenTableE is the regression
// test for the shared career_generate.go fix this slice needed: "Any
// Skill" now draws from citizenTableE via rollCitizenTableEName, per
// Book 1 p.87's own footnote ("from Citizen Life Skills and
// Knowledges"), rather than falling through to the default case and
// being granted as a literal skill named "Any Skill".
func TestResolveSkillCellTreatsAnySkillAsCitizenTableE(t *testing.T) {
	t.Parallel()

	sawResolved, sawUnresolved := false, false

	for seed := uint64(1); seed <= 200; seed++ {
		r := dice.New(rand.NewPCG(seed, seed))

		skill, ok := resolveSkillCell(r, functionarySkillTable, 3, 3) // column 4 "General", row 4 "Any Skill"
		if ok {
			sawResolved = true

			if skill.Name == "Any Skill" {
				t.Fatalf("seed=%d: resolveSkillCell granted a literal \"Any Skill\" skill", seed)
			}
		} else {
			sawUnresolved = true
		}
	}

	if !sawResolved {
		t.Error("0 of 200 seeds resolved \"Any Skill\" to a real skill, want at least 1")
	}

	if !sawUnresolved {
		t.Error("0 of 200 seeds hit citizenTableE's own \"No Skill\" entry, want at least 1 (in 200 trials)")
	}
}

func TestFunctionaryRankNamesMatchBook1P87(t *testing.T) {
	t.Parallel()

	want := [9]string{
		"Clerk", "Supervisor", "Senior Supervisor", "Manager", "Senior Manager",
		"Assistant Director", "Director", "Nth UnderSecretary", "Secretary",
	}

	if functionaryRankNames != want {
		t.Errorf("functionaryRankNames = %v, want %v", functionaryRankNames, want)
	}
}

func TestFunctionaryRankNameF6VariesByPrecedingCareer(t *testing.T) {
	t.Parallel()

	cases := []struct {
		precedingCareer string
		want            string
	}{
		{"scholar", "F6 College President"},
		{"entertainer", "F6 Association Director"},
		{"merchant", "F6 Starport Warden"},
		{"rogue", "F6 Bank President"},
		{"scout", "F6 Director"},
		{"", "F6 Director"},
	}

	for _, c := range cases {
		t.Run(c.precedingCareer, func(t *testing.T) {
			t.Parallel()

			if got := functionaryRankName(6, c.precedingCareer, 0); got != c.want {
				t.Errorf("functionaryRankName(6, %q, 0) = %q, want %q", c.precedingCareer, got, c.want)
			}
		})
	}
}

func TestFunctionaryRankNameF7UsesOrdinal(t *testing.T) {
	t.Parallel()

	if got := functionaryRankName(7, "", 3); got != "F7 3rd UnderSecretary" {
		t.Errorf("functionaryRankName(7, \"\", 3) = %q, want %q", got, "F7 3rd UnderSecretary")
	}
}

// TestBeginFunctionaryTargetFormula confirms Book 1 p.87's own "To
// Begin Total Terms x3": 0 prior terms is an impossible target (2D6
// minimum is 2), and enough prior terms makes it automatic (2D6 maximum
// is 12).
func TestBeginFunctionaryTargetFormula(t *testing.T) {
	t.Parallel()

	for seed := uint64(1); seed <= 20; seed++ {
		if BeginFunctionary(dice.New(rand.NewPCG(seed, seed)), 0) {
			t.Fatalf("seed=%d: BeginFunctionary(0 prior terms) = true, want false (target 0 is unreachable)", seed)
		}
	}

	for seed := uint64(1); seed <= 20; seed++ {
		if !BeginFunctionary(dice.New(rand.NewPCG(seed, seed)), 4) {
			t.Fatalf("seed=%d: BeginFunctionary(4 prior terms) = false, want true (target 12 is automatic)", seed)
		}
	}
}

var uppFunctionary88 = UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}

// The ResolveFunctionaryTerm fixtures below were seed-hunted against an
// all-8 UPP and verified by direct inspection before being pinned here.

func TestResolveFunctionaryTermRiskSuccessRewardSuccess(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(2, 2))

	term, tier, _ := ResolveFunctionaryTerm(r, uppFunctionary88, C2, 0, 0, "scholar")

	if term.RiskResult != Unharmed {
		t.Fatalf("RiskResult = %v, want Unharmed", term.RiskResult)
	}

	if !term.Promoted || tier != 1 {
		t.Errorf("Promoted = %v, tier = %d, want true, 1", term.Promoted, tier)
	}

	if term.RewardResult != "Promoted to F1 Supervisor" {
		t.Errorf("RewardResult = %q, want %q", term.RewardResult, "Promoted to F1 Supervisor")
	}

	if term.Rank != "F1 Supervisor" {
		t.Errorf("Rank = %q, want %q", term.Rank, "F1 Supervisor")
	}
}

func TestResolveFunctionaryTermRiskSuccessRewardFailure(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))

	term, tier, _ := ResolveFunctionaryTerm(r, uppFunctionary88, C2, 0, 0, "scholar")

	if term.RiskResult != Unharmed {
		t.Fatalf("RiskResult = %v, want Unharmed", term.RiskResult)
	}

	if term.Promoted || tier != 0 {
		t.Errorf("Promoted = %v, tier = %d, want false, 0", term.Promoted, tier)
	}

	if term.RewardResult != "None" {
		t.Errorf("RewardResult = %q, want %q", term.RewardResult, "None")
	}
}

// TestResolveFunctionaryTermRiskFailureEndsCareer confirms Office
// Politics' own Risk failure is recorded as RiskResult Disabled (a
// deliberate reuse, not a physical wound) and that Reward is never
// rolled that term (RewardResult stays at its zero value, "").
func TestResolveFunctionaryTermRiskFailureEndsCareer(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(4, 4))

	term, _, _ := ResolveFunctionaryTerm(r, uppFunctionary88, C2, 0, 0, "scholar")

	if term.RiskResult != Disabled {
		t.Fatalf("RiskResult = %v, want Disabled", term.RiskResult)
	}

	if term.RewardResult != "" {
		t.Errorf("RewardResult = %q, want empty (Reward is never rolled on Risk failure)", term.RewardResult)
	}

	if term.Promoted {
		t.Error("Promoted = true, want false")
	}

	if len(term.SkillsAwarded) == 0 {
		t.Error(
			"SkillsAwarded is empty, want the term's own base skills (a Risk failure doesn't lose the term's benefit)",
		)
	}
}

// TestResolveFunctionaryTermGrantsTierAutoSkillOnPromotion confirms the
// Rank table's own Auto Skill column (F2=Admin) is granted the moment
// tier first reaches 2, in addition to (not instead of) the Skill
// Eligibility box's own "Per Promotion 1" bonus roll.
func TestResolveFunctionaryTermGrantsTierAutoSkillOnPromotion(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(2, 2))

	term, tier, _ := ResolveFunctionaryTerm(r, uppFunctionary88, C2, 1, 0, "")

	if tier != 2 {
		t.Fatalf("tier = %d, want 2 (fixture assumption broken)", tier)
	}

	found := false

	for _, s := range term.SkillsAwarded {
		if s.Name == "Admin" {
			found = true
		}
	}

	if !found {
		t.Errorf("SkillsAwarded = %+v, want an Admin auto skill from reaching F2", term.SkillsAwarded)
	}

	if len(term.SkillsAwarded) != functionarySkillsPerTerm+2 {
		t.Errorf("len(SkillsAwarded) = %d, want %d (4 base + 1 promotion roll + 1 auto skill)",
			len(term.SkillsAwarded), functionarySkillsPerTerm+2)
	}
}
