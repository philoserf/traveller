package character

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

func TestBuildSpacerCharacterNeverQualified(t *testing.T) {
	t.Parallel()

	upp := UPP{}
	homeworldSkills := []SkillLevel{{Name: "Vacc Suit", Level: 1, Kind: Skill}}
	r := dice.New(rand.NewPCG(1, 1))

	c, ok := buildSpacerCharacter(r, upp, "A000000-0", homeworldSkills)

	if ok {
		t.Error("ok = true, want false (never qualified)")
	}

	if len(c.Careers) != 1 || len(c.Careers[0].Terms) != 0 {
		t.Errorf("Careers = %+v, want one Career with zero Terms", c.Careers)
	}

	if c.WoundBadges != 0 {
		t.Errorf("WoundBadges = %d, want 0", c.WoundBadges)
	}

	if c.Age != 18 {
		t.Errorf("Age = %d, want 18", c.Age)
	}

	if !slices.Equal(c.Skills, homeworldSkills) {
		t.Errorf("Skills = %v, want exactly homeworldSkills %v (no career skills)", c.Skills, homeworldSkills)
	}
}

// TestBuildSpacerCharacterDies mirrors TestBuildSoldierCharacterDies.
// C4=12 guarantees Begin always succeeds (Spacer has no Retry, and its
// own Begin targets Int, not Str); Dex=1 makes death likely whenever the
// C1/C2/C4 rotation lands on C2.
func TestBuildSpacerCharacterDies(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{12, 1, 0, 12, 8, 0}}
	r := dice.New(rand.NewPCG(41, 43))

	sawDeath := false

	for range 200 {
		c, ok := buildSpacerCharacter(r, upp, "hw", nil)

		terms := c.Careers[0].Terms
		if len(terms) == 0 {
			t.Fatal("career has zero terms, want Begin to always succeed (C4=12)")
		}

		wantOK := terms[len(terms)-1].RiskResult != Dead
		if ok != wantOK {
			t.Fatalf("ok = %v, want %v (last term RiskResult = %v)", ok, wantOK, terms[len(terms)-1].RiskResult)
		}

		if !wantOK {
			sawDeath = true
		}
	}

	if !sawDeath {
		t.Fatal("no trial produced a death across 200 trials — fixture can't verify the ok=false path")
	}
}

// TestBuildSpacerCharacterQualified confirms a real career actually
// populates Age/LifeStage/Birthdate/Skills/WoundBadges/Rank. Like
// Soldier's own equivalent test (not Marine's), Spacer's own Commission
// and Rating Promotion both target C2 (Dex) — the same characteristic
// that must also be high for Risk & Reward to survive on all three of
// Spacer's own CC positions (C1 C2 C4) — so this fixture sets
// Str=Dex=Int=Soc=20: Risk always succeeds, Commission succeeds
// immediately (Dex=20), and Officer Promotion (Soc=20) then always
// succeeds too, deterministically reaching O7 Admiral by term 7 and
// staying capped for the remaining 7 terms.
func TestBuildSpacerCharacterQualified(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{20, 20, 0, 20, 8, 20}}
	homeworldSkills := []SkillLevel{{Name: "Vacc Suit", Level: 1, Kind: Skill}}
	r := dice.New(rand.NewPCG(1, 1))

	c, ok := buildSpacerCharacter(r, upp, "hw", homeworldSkills)

	if !ok {
		t.Fatal("ok = false, want true (fixture guarantees survival)")
	}

	if len(c.Careers) != 1 || c.Careers[0].Name != SpacerCareerName {
		t.Fatalf("Careers = %+v, want one Career named %q", c.Careers, SpacerCareerName)
	}

	if !c.Careers[0].HasRank {
		t.Error("Careers[0].HasRank = false, want true")
	}

	if len(c.Careers[0].Terms) == 0 {
		t.Fatal("Careers[0].Terms is empty, want at least one term")
	}

	if c.Age < 18 {
		t.Errorf("Age = %d, want >= 18", c.Age)
	}

	if want := LifeStageForAge(c.Age); c.LifeStage != want {
		t.Errorf("LifeStage = %d, want %d (LifeStageForAge(%d))", c.LifeStage, want, c.Age)
	}

	assertBirthdateFormat(t, c.Birthdate, c.Age)

	if len(c.Skills) < len(homeworldSkills) {
		t.Errorf("len(Skills) = %d, want at least %d (homeworld skills alone)", len(c.Skills), len(homeworldSkills))
	}

	if c.WoundBadges != 0 {
		t.Errorf("WoundBadges = %d, want 0 (fixture guarantees Risk always succeeds)", c.WoundBadges)
	}

	// This seed's own Reward rolls resolve to 2 MCUF (+1 each), 1 MCG
	// (+2), and 1 SEH (+3) — 7 Medal Fame — plus Officer Rank Fame at
	// the O7 cap (tier 7): 7+7=14.
	if c.Fame != 14 {
		t.Errorf("Fame = %d, want 14 (2 MCUF + 1 MCG + 1 SEH Medal Fame, plus O7's own Rank Fame)", c.Fame)
	}

	if len(c.Medals) != 28 {
		t.Errorf("len(Medals) = %d, want 28 (2 medals x 14 terms: flat XS plus a Reward-table medal each)",
			len(c.Medals))
	}

	if c.Rank != "O7 Admiral" {
		t.Errorf("Rank = %q, want %q", c.Rank, "O7 Admiral")
	}
}
