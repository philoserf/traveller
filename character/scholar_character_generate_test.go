package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// TestBuildScholarCharacterNeverQualified mirrors
// TestBuildRogueCharacterNeverQualified's own shape: Edu 0 guarantees
// BeginScholar's own roll fails.
func TestBuildScholarCharacterNeverQualified(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 0, 8}}
	r := dice.New(rand.NewPCG(1, 1))

	c, ok := buildScholarCharacter(r, upp, "hw", nil)

	if ok {
		t.Error("ok = true, want false (BeginScholar's own roll fails against Edu 0)")
	}

	if len(c.Careers) != 1 || len(c.Careers[0].Terms) != 0 {
		t.Errorf("Careers = %+v, want one Career with zero Terms", c.Careers)
	}

	if c.Age != 19 {
		t.Errorf("Age = %d, want 19 (18 plus the one year the failed Begin roll cost, Book 1 p.65)", c.Age)
	}

	if c.Fame != 0 {
		t.Errorf("Fame = %d, want 0 (never qualified)", c.Fame)
	}
}

// TestBuildScholarCharacterQualified pins seed 9 against Str/Dex/End/
// Int/Soc 8, Edu 9: one term, Research succeeds (Unharmed), Publication
// succeeds (not Award-Winning), Promoted true (tier 1 Lecturer -> 2
// Instructor). Fame = 4 (scholarCareerFame: tier 2 + pubs 1 = 3, plus a
// Mustering Out "Fame +1" Benefit) — confirmed by direct inspection
// before being pinned, not assumed from the formula alone.
func TestBuildScholarCharacterQualified(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 9, 8}}
	homeworldSkills := []SkillLevel{{Name: "Vacc Suit", Level: 1, Kind: Skill}}
	r := dice.New(rand.NewPCG(9, 9))

	c, ok := buildScholarCharacter(r, upp, "hw", homeworldSkills)

	if !ok {
		t.Fatal("ok = false, want true")
	}

	if len(c.Careers) != 1 || c.Careers[0].Name != ScholarCareerName {
		t.Fatalf("Careers = %+v, want one Career named %q", c.Careers, ScholarCareerName)
	}

	if !c.Careers[0].HasRank {
		t.Error("Careers[0].HasRank = false, want true")
	}

	if len(c.Careers[0].Terms) != 1 {
		t.Fatalf("len(Careers[0].Terms) = %d, want 1", len(c.Careers[0].Terms))
	}

	term := c.Careers[0].Terms[0]
	if term.RiskResult != Unharmed {
		t.Errorf("Terms[0].RiskResult = %v, want Unharmed", term.RiskResult)
	}

	if !term.PublicationSucceeded || term.AwardWinning {
		t.Errorf("Terms[0].PublicationSucceeded = %v, AwardWinning = %v, want true, false",
			term.PublicationSucceeded, term.AwardWinning)
	}

	if !term.Promoted {
		t.Error("Terms[0].Promoted = false, want true")
	}

	if c.Rank != "Instructor" {
		t.Errorf("Rank = %q, want %q", c.Rank, "Instructor")
	}

	if c.Fame != 4 {
		t.Errorf("Fame = %d, want 4", c.Fame)
	}

	if len(c.Skills) < len(homeworldSkills) {
		t.Errorf("len(Skills) = %d, want at least %d (homeworld skills alone)", len(c.Skills), len(homeworldSkills))
	}

	assertBirthdateFormat(t, c.Birthdate, c.Age)
}

// TestBuildScholarCharacterDiedMidCareer pins seed 1 against a very low
// Str/Dex/End/Int (2) UPP: the character dies on term 2's own Risk roll
// (RiskResult == Dead), so ok is false despite two real Terms having
// been resolved — mirrors buildRiskCareerCharacter's own Dead-check
// convention, confirmed here since buildScholarCharacter reimplements
// it directly (bespoke, not delegating to buildRiskCareerCharacter).
func TestBuildScholarCharacterDiedMidCareer(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{2, 2, 2, 2, 10, 8}}
	r := dice.New(rand.NewPCG(1, 1))

	c, ok := buildScholarCharacter(r, upp, "hw", nil)

	if ok {
		t.Error("ok = true, want false (the character dies mid-career)")
	}

	if len(c.Careers[0].Terms) == 0 {
		t.Fatal("Careers[0].Terms is empty, want at least one term before death")
	}

	last := c.Careers[0].Terms[len(c.Careers[0].Terms)-1]
	if last.RiskResult != Dead {
		t.Errorf("last term's RiskResult = %v, want Dead", last.RiskResult)
	}

	if c.Fame != 0 {
		t.Errorf("Fame = %d, want 0 (a died character doesn't retain intrinsic career Fame)", c.Fame)
	}

	if c.Cash != 0 {
		t.Errorf("Cash = %d, want 0 (Dead zeroes Mustering Out, per scoutMusterOutRollCount)", c.Cash)
	}
}

// TestGenerateScholarCharacterProducesAHumanCharacter is a smoke test
// confirming the full public entry point wires GenerateUPP/
// GenerateHomeworldSkills into buildScholarCharacter, mirroring every
// other career's own GenerateXCharacter smoke test.
func TestGenerateScholarCharacterProducesAHumanCharacter(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))

	c, _ := GenerateScholarCharacter(r)

	if c.Species != "Human" {
		t.Errorf("Species = %q, want %q", c.Species, "Human")
	}

	if len(c.Careers) != 1 || c.Careers[0].Name != ScholarCareerName {
		t.Errorf("Careers = %+v, want one Career named %q", c.Careers, ScholarCareerName)
	}
}
