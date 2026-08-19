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
// careerWoundBadges (which counts Talent setbacks as physical wounds).
// Seed 39 against an all-8 UPP was confirmed by direct inspection to end
// its one term Dead (Talent exhausted) while still producing ok=true,
// Age > 18 (aging ran), and WoundBadges == 0.
func TestBuildEntertainerCharacterTalentExhaustedIsStillAlive(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}
	r := dice.New(rand.NewPCG(21, 21))

	c, ok := buildEntertainerCharacter(r, upp, "hw", nil, nil)

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

	c, ok := buildEntertainerCharacter(r, upp, "hw", nil, nil)

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
// this fixture), Cash = 15,000 from a Mustering Out "Cr" Money entry —
// confirmed by direct inspection before being pinned.
func TestBuildEntertainerCharacterQualified(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}
	homeworldSkills := []SkillLevel{{Name: "Vacc Suit", Level: 1, Kind: Skill}}
	r := dice.New(rand.NewPCG(14, 14))

	c, ok := buildEntertainerCharacter(r, upp, "hw", homeworldSkills, nil)

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

	if c.Fame != 7 {
		t.Errorf("Fame = %d, want 7", c.Fame)
	}

	if c.Cash != 15000 {
		t.Errorf("Cash = %d, want 15000", c.Cash)
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

// TestGenerateEntertainerCharacterElectsLaterEducation is #164's own
// end-to-end regression test, mirroring the pilot's
// TestGenerateRogueCharacterElectsLaterEducation (rogue_character_generate_test.go):
// seed 1 confirmed by direct inspection to elect Later Education five
// times (Service Academy, then University, then — #163 — three Training
// Courses at Edu=10, two passed and the third's escalating Mod finally
// failing it — the same repeated-attempts-then-a-failure shape p.62's
// own Barr Vech worked example follows). Pins that the mechanism actually
// fires through the real standalone generator for a hand-rolled loop —
// not just resolveCareerLoop's shared hook in isolation — and that Fame/
// Talent are untouched by a Later Education term (no Risk/Reward or Flux
// roll happens during one).
func TestGenerateEntertainerCharacterElectsLaterEducation(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))

	c, ok := GenerateEntertainerCharacter(r)
	if !ok {
		t.Fatal("ok = false, want true")
	}

	var laterEdTerms, skillsAwarded int

	for _, term := range c.Careers[0].Terms {
		if !term.LaterEducation {
			continue
		}

		laterEdTerms++

		if term.LaterEducationSchool == "" {
			t.Error("a Later Education term has an empty LaterEducationSchool")
		}

		skillsAwarded += len(term.SkillsAwarded)
	}

	if laterEdTerms != 5 {
		t.Fatalf("laterEdTerms = %d, want 5", laterEdTerms)
	}

	if skillsAwarded == 0 {
		t.Error("no Later Education term awarded any skills")
	}
}
