package character

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

func TestBuildSoldierCharacterNeverQualified(t *testing.T) {
	t.Parallel()

	upp := UPP{}
	homeworldSkills := []SkillLevel{{Name: "Vacc Suit", Level: 1, Kind: Skill}}
	r := dice.New(rand.NewPCG(1, 1))

	c, ok := buildSoldierCharacter(r, upp, "A000000-0", homeworldSkills)

	if ok {
		t.Error("ok = true, want false (never qualified)")
	}

	if len(c.Careers) != 1 || len(c.Careers[0].Terms) != 0 {
		t.Errorf("Careers = %+v, want one Career with zero Terms", c.Careers)
	}

	if c.WoundBadges != 0 {
		t.Errorf("WoundBadges = %d, want 0", c.WoundBadges)
	}

	if c.Age != 19 {
		t.Errorf("Age = %d, want 19 (18 plus the one year the failed Begin roll cost, Book 1 p.65)", c.Age)
	}

	if !slices.Equal(c.Skills, homeworldSkills) {
		t.Errorf("Skills = %v, want exactly homeworldSkills %v (no career skills)", c.Skills, homeworldSkills)
	}
}

// TestBuildSoldierCharacterDies mirrors TestBuildMarineCharacterDies.
// C1=12 guarantees Begin always succeeds (Soldier has no Retry); C4=1
// makes death likely whenever the C1/C3/C4 rotation lands on C4 (Risk
// target = 1-combinedMod, effectively always <=0).
func TestBuildSoldierCharacterDies(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{12, 0, 12, 1, 8, 0}}
	r := dice.New(rand.NewPCG(41, 43))

	sawDeath := false

	for range 200 {
		c, ok := buildSoldierCharacter(r, upp, "hw", nil)

		terms := c.Careers[0].Terms
		if len(terms) == 0 {
			t.Fatal("career has zero terms, want Begin to always succeed (C1=12)")
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

// TestBuildSoldierCharacterQualified confirms a real career actually
// populates Age/LifeStage/Birthdate/Skills/WoundBadges/Rank. Unlike
// Marine's own equivalent test (which keeps the character Enlisted by
// zeroing C3), Soldier's own Commission and Enlisted Promotion both
// target C3 (End) — the same characteristic that must also be high for
// Risk & Reward to survive on all three of Soldier's own CC positions
// (C1 C3 C4) — so End can't be zeroed without also risking death. This
// fixture instead sets Str=End=Int=Soc=20: Commission succeeds
// immediately (End=20), and Officer Promotion (targeting Soc, not Int
// like Marine) then always succeeds too, deterministically reaching O7
// General by term 7 and staying capped for the remaining 7 terms — a
// genuinely different, complementary deterministic path from Marine's
// own "stays Enlisted" fixture.
//
// What this fixture does NOT guarantee is that Risk always succeeds,
// though it used to say so. Fourteen terms carry the character to age 74,
// and Aging erodes those 20s on the way — this seed ends on End 10 — so a
// late-term Risk roll can and does fail. The old "WoundBadges = 0"
// assertion only held because the dice stream happened to avoid it, and
// it broke the moment Command College (command_college.go) added draws in
// the middle of the career. The wound-badge and medal assertions below are
// derived from the resolved career for that reason, not pinned.
func TestBuildSoldierCharacterQualified(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{20, 0, 20, 20, 8, 20}}
	homeworldSkills := []SkillLevel{{Name: "Vacc Suit", Level: 1, Kind: Skill}}
	r := dice.New(rand.NewPCG(1, 1))

	c, ok := buildSoldierCharacter(r, upp, "hw", homeworldSkills)

	if !ok {
		t.Fatal("ok = false, want true (fixture guarantees survival)")
	}

	if len(c.Careers) != 1 || c.Careers[0].Name != SoldierCareerName {
		t.Fatalf("Careers = %+v, want one Career named %q", c.Careers, SoldierCareerName)
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

	if want := scoutWoundBadges(c.Careers[0]); c.WoundBadges != want {
		t.Errorf("WoundBadges = %d, want %d (the career's own wounded terms)", c.WoundBadges, want)
	}

	// Derived rather than pinned. The Medal mix a seed produces shifts
	// whenever the dice stream does, and this assertion is for
	// soldierCareerFameAwards being wired through end to end, not for any
	// particular medal count — the p.91 rule is that Fame is the Medal
	// Fame plus Officer Rank Fame, whatever this seed happened to roll.
	if want := resolveFameStacks(soldierCareerFameAwards(c.Careers[0])); c.Fame < want {
		t.Errorf("Fame = %d, want at least %d (career Fame, before Mustering Out's own additions)",
			c.Fame, want)
	}

	if c.Fame == 0 {
		t.Error("Fame = 0, want some (fixture can't verify Fame propagation at all)")
	}

	// Every surviving term awards the flat XS plus a Reward-table medal;
	// a term whose Risk failed awards only the Reward one. Derived from
	// the terms rather than pinned at 2 x 14, since which terms survive
	// moves with the dice stream.
	if want := len(allMedalsFromTerms(c.Careers[0].Terms)); len(c.Medals) != want {
		t.Errorf("len(Medals) = %d, want %d (the career's own medals)", len(c.Medals), want)
	}

	if len(c.Medals) < len(c.Careers[0].Terms) {
		t.Errorf("len(Medals) = %d, want at least one per term (%d)", len(c.Medals), len(c.Careers[0].Terms))
	}

	// Commission succeeds immediately (End=20), then Officer Promotion
	// (Soc=20) advances every term until the O7 cap, well within the
	// career's own 14-term run.
	if c.Rank != "O7 General" {
		t.Errorf("Rank = %q, want %q", c.Rank, "O7 General")
	}
}
