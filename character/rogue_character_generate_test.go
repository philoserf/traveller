package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// TestBuildRogueCharacterNeverQualified mirrors
// TestBuildNobleCharacterNeverQualified's own shape: a zero UPP
// guarantees BeginRogue fails regardless of which CC rollRogueCC picks,
// but Age/LifeStage/Birthdate are still computed (finalizeAging always
// runs).
func TestBuildRogueCharacterNeverQualified(t *testing.T) {
	t.Parallel()

	upp := UPP{}
	r := dice.New(rand.NewPCG(1, 1))

	c, ok := buildRogueCharacter(r, upp, "hw", nil, nil)

	if ok {
		t.Error("ok = true, want false (BeginRogue fails against a zero UPP)")
	}

	if len(c.Careers) != 1 || len(c.Careers[0].Terms) != 0 {
		t.Errorf("Careers = %+v, want one Career with zero Terms", c.Careers)
	}

	if c.Age != 19 {
		t.Errorf("Age = %d, want 19 (18 plus the one year the failed Begin roll cost, Book 1 p.65)", c.Age)
	}

	if c.Fame != 0 {
		t.Errorf("Fame = %d, want 0 (never qualified, no Scheme was ever resolved)", c.Fame)
	}

	if c.Cash != 0 {
		t.Errorf("Cash = %d, want 0", c.Cash)
	}
}

// TestBuildRogueCharacterQualifiedOneTermShipShare pins seed 25 against
// an all-8 UPP: one term, Risk succeeds (not Imprisoned), Reward
// succeeds on a Ship-Share-valued Scheme (SchemePayoff stays 0 — a Ship
// Share is a flat grant, not a scaled Payoff), Fame = 2 ("Successful
// Schemes x2", Book 1 p.91), Cash = 25,000 from a Mustering Out "Cr"
// Money entry (not from SchemePayoff, which is 0 here) — confirmed by
// direct inspection before being pinned.
func TestBuildRogueCharacterQualifiedOneTermShipShare(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}
	r := dice.New(rand.NewPCG(25, 25))

	c, ok := buildRogueCharacter(r, upp, "hw", nil, nil)

	if !ok {
		t.Fatal("ok = false, want true")
	}

	if len(c.Careers) != 1 || c.Careers[0].Name != RogueCareerName {
		t.Fatalf("Careers = %+v, want one Career named %q", c.Careers, RogueCareerName)
	}

	if c.Careers[0].HasRank {
		t.Error("Careers[0].HasRank = true, want false (Rogue is in Book 1's no-rank career list)")
	}

	if len(c.Careers[0].Terms) != 1 {
		t.Fatalf("len(Careers[0].Terms) = %d, want 1", len(c.Careers[0].Terms))
	}

	term := c.Careers[0].Terms[0]
	if term.Imprisoned {
		t.Error("Terms[0].Imprisoned = true, want false")
	}

	if !term.SchemeShipShare {
		t.Error("Terms[0].SchemeShipShare = false, want true")
	}

	if c.Fame != 2 {
		t.Errorf("Fame = %d, want 2 (one Successful Scheme x2)", c.Fame)
	}

	if c.Cash != 25000 {
		t.Errorf("Cash = %d, want 25000 (from Mustering Out, not SchemePayoff)", c.Cash)
	}
}

// TestBuildRogueCharacterQualifiedTwoTermsPayoff pins seed 48 against
// the same all-8 UPP: two terms, both successful (neither Imprisoned) —
// term 0 grants a scaled Payoff of 200,000, term 1 is a Ship-Share
// success (Payoff 0) — Fame = 4 (two Successful Schemes x2 each), Cash =
// 200,000 (Mustering Out's own two Benefits entries contribute no Cash)
// — confirmed by direct inspection before being pinned.
func TestBuildRogueCharacterQualifiedTwoTermsPayoff(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}
	r := dice.New(rand.NewPCG(48, 48))

	c, ok := buildRogueCharacter(r, upp, "hw", nil, nil)

	if !ok {
		t.Fatal("ok = false, want true")
	}

	if len(c.Careers[0].Terms) != 2 {
		t.Fatalf("len(Careers[0].Terms) = %d, want 2", len(c.Careers[0].Terms))
	}

	for i, term := range c.Careers[0].Terms {
		if term.Imprisoned {
			t.Errorf("Terms[%d].Imprisoned = true, want false", i)
		}
	}

	if c.Careers[0].Terms[0].SchemePayoff != 200000 {
		t.Errorf("Terms[0].SchemePayoff = %d, want 200000", c.Careers[0].Terms[0].SchemePayoff)
	}

	if !c.Careers[0].Terms[1].SchemeShipShare {
		t.Error("Terms[1].SchemeShipShare = false, want true")
	}

	if c.Fame != 4 {
		t.Errorf("Fame = %d, want 4 (two Successful Schemes x2 each)", c.Fame)
	}

	if c.Cash != 200000 {
		t.Errorf("Cash = %d, want 200000 (Terms[0].SchemePayoff; Mustering Out granted no Cr entries)", c.Cash)
	}
}

// TestGenerateRogueCharacterProducesAHumanCharacter is a smoke test
// confirming the full public entry point wires GenerateUPP/
// GenerateHomeworldSkills into buildRogueCharacter, mirroring every
// other career's own GenerateXCharacter smoke test.
func TestGenerateRogueCharacterProducesAHumanCharacter(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))

	c, _ := GenerateRogueCharacter(r)

	if c.Species != "Human" {
		t.Errorf("Species = %q, want %q", c.Species, "Human")
	}

	if len(c.Careers) != 1 || c.Careers[0].Name != RogueCareerName {
		t.Errorf("Careers = %+v, want one Career named %q", c.Careers, RogueCareerName)
	}
}

// TestGenerateRogueCharacterElectsLaterEducation is the pilot's own
// end-to-end regression test (#113 item 5, stage 3): seed 15 confirmed
// by direct inspection to elect Later Education exactly once, at
// University, graduating with Honors. Pins that the mechanism actually
// fires through the real standalone generator — not just the shared
// resolveCareerLoop hook in isolation — and that the final
// Character.Education reflects the attendance rather than the stale
// pre-career snapshot (the bug TestCareerChainSingleEntryMatchesLegacyGenerator's
// own rogue case caught: careerSegment used to drop a segment's Later
// Education mutation entirely).
func TestGenerateRogueCharacterElectsLaterEducation(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(15, 15))

	c, _ := GenerateRogueCharacter(r)

	laterEdTerms := 0

	for _, term := range c.Careers[0].Terms {
		if term.LaterEducation {
			laterEdTerms++

			if term.LaterEducationSchool != "University" {
				t.Errorf("LaterEducationSchool = %q, want %q", term.LaterEducationSchool, "University")
			}

			if len(term.SkillsAwarded) == 0 {
				t.Error("a Later Education term awarded no skills")
			}
		}
	}

	if laterEdTerms != 1 {
		t.Fatalf("laterEdTerms = %d, want 1", laterEdTerms)
	}

	if c.Education.School != "University" || !c.Education.Graduated || c.Education.Degree != educationDegreeHonors {
		t.Errorf("Education = %+v, want School=University Graduated=true Degree=%q",
			c.Education, educationDegreeHonors)
	}
}
