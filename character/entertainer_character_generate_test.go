package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// TestBuildEntertainerCharacterTalentExhaustedIsStillAlive is the
// regression test for a code-review-caught bug: an earlier version
// gated ok on the last term's own RiskResult != Dead, the same as every
// physically-mortal career — but Entertainer's own Dead means Talent
// completely spent, not physical death, so that check silently skipped
// ResolveAging for a perfectly alive character (age would stay frozen at
// 18, no aging Notes ever generated) and misreported WoundBadges via
// scoutWoundBadges (which counts Talent setbacks as physical wounds).
// Seed 39 against an all-8 UPP was confirmed by direct inspection to end
// its one term Dead (Talent exhausted) while still producing ok=true,
// Age > 18 (aging ran), and WoundBadges == 0.
func TestBuildEntertainerCharacterTalentExhaustedIsStillAlive(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}
	r := dice.New(rand.NewPCG(39, 39))

	c, ok := buildEntertainerCharacter(r, upp, "hw", nil)

	if !ok {
		t.Fatal("ok = false, want true (Talent exhaustion is not physical death)")
	}

	last := c.Careers[0].Terms[len(c.Careers[0].Terms)-1]
	if last.RiskResult != Dead {
		t.Fatalf("last term's RiskResult = %v, want Dead (fixture assumption broke)", last.RiskResult)
	}

	if c.Age <= 18 {
		t.Errorf("Age = %d, want > 18 (ResolveAging must still run for a Talent-exhausted-but-alive character)", c.Age)
	}

	if c.WoundBadges != 0 {
		t.Errorf(
			"WoundBadges = %d, want 0 (Entertainer's own Wounded/Disabled means a Talent setback, not a physical wound)",
			c.WoundBadges,
		)
	}
}

// TestBuildEntertainerCharacterNeverQualifiedStillSetsFame is the
// end-to-end regression test for the same judgment call
// TestResolveEntertainerCareerNeverQualifiedStillSetsFame covers at the
// career level: Character.Fame is set even when ok is false, since the
// initial roll happens "Before Begin." Seed 6 against an all-8 UPP was
// confirmed by direct inspection to fail Begin while still producing
// Fame 7.
func TestBuildEntertainerCharacterNeverQualifiedStillSetsFame(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}
	r := dice.New(rand.NewPCG(6, 6))

	c, ok := buildEntertainerCharacter(r, upp, "hw", nil)

	if ok {
		t.Error("ok = true, want false (BeginEntertainer's own roll fails)")
	}

	if len(c.Careers) != 1 || len(c.Careers[0].Terms) != 0 {
		t.Errorf("Careers = %+v, want one Career with zero Terms", c.Careers)
	}

	if c.Age != 19 {
		t.Errorf("Age = %d, want 19 (18 plus the one year the failed Begin roll cost, Book 1 p.65)", c.Age)
	}

	if c.Fame != 7 {
		t.Errorf("Fame = %d, want 7 (the initial 2D roll, set even though never qualified)", c.Fame)
	}

	if c.Cash != 0 {
		t.Errorf("Cash = %d, want 0", c.Cash)
	}
}

// TestBuildEntertainerCharacterQualified pins seed 3 against an all-8
// UPP: one term, Fame doesn't increase (FameAfterTerm 1), Risk succeeds
// (Unharmed), Reward fails — Fame = 1 (no Mustering Out "Fame +N" entry
// this fixture), Cash = 25,000 from a Mustering Out "Cr" Money entry —
// confirmed by direct inspection before being pinned.
func TestBuildEntertainerCharacterQualified(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}
	homeworldSkills := []SkillLevel{{Name: "Vacc Suit", Level: 1, Kind: Skill}}
	r := dice.New(rand.NewPCG(3, 3))

	c, ok := buildEntertainerCharacter(r, upp, "hw", homeworldSkills)

	if !ok {
		t.Fatal("ok = false, want true")
	}

	if len(c.Careers) != 1 || c.Careers[0].Name != EntertainerCareerName {
		t.Fatalf("Careers = %+v, want one Career named %q", c.Careers, EntertainerCareerName)
	}

	if c.Careers[0].HasRank {
		t.Error("Careers[0].HasRank = true, want false (Entertainer has no Rank concept)")
	}

	if c.Careers[0].Specialty == "" {
		t.Error("Careers[0].Specialty is empty, want a chosen Specialty")
	}

	if len(c.Careers[0].Terms) != 1 {
		t.Fatalf("len(Careers[0].Terms) = %d, want 1", len(c.Careers[0].Terms))
	}

	if c.Fame != 1 {
		t.Errorf("Fame = %d, want 1", c.Fame)
	}

	if c.Cash != 25000 {
		t.Errorf("Cash = %d, want 25000", c.Cash)
	}

	if len(c.Skills) < len(homeworldSkills) {
		t.Errorf("len(Skills) = %d, want at least %d (homeworld skills alone)", len(c.Skills), len(homeworldSkills))
	}

	assertBirthdateFormat(t, c.Birthdate, c.Age)
}

// TestGenerateEntertainerCharacterProducesAHumanCharacter is a smoke
// test confirming the full public entry point wires GenerateUPP/
// GenerateHomeworldSkills into buildEntertainerCharacter, mirroring
// every other career's own GenerateXCharacter smoke test.
func TestGenerateEntertainerCharacterProducesAHumanCharacter(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))

	c, _ := GenerateEntertainerCharacter(r)

	if c.Species != "Human" {
		t.Errorf("Species = %q, want %q", c.Species, "Human")
	}

	if len(c.Careers) != 1 || c.Careers[0].Name != EntertainerCareerName {
		t.Errorf("Careers = %+v, want one Career named %q", c.Careers, EntertainerCareerName)
	}
}
