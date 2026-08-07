package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// TestBuildNobleCharacterNeverQualified confirms a below-B Soc produces
// a zero-Terms career with ok=false, mirroring
// TestBuildScoutCharacterNeverQualified — but Age is still computed
// (18, 0 terms served), matching Scout's own never-qualified-still-
// gets-an-Age precedent (finalizeAging always runs, only ResolveAging
// itself is skipped when !ok).
func TestBuildNobleCharacterNeverQualified(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 10}} // Soc 10 < B
	r := dice.New(rand.NewPCG(1, 1))

	c, ok := buildNobleCharacter(r, upp, "hw", nil, nil)

	if ok {
		t.Error("ok = true, want false (Soc below B)")
	}

	if len(c.Careers) != 1 || len(c.Careers[0].Terms) != 0 {
		t.Errorf("Careers = %+v, want one Career with zero Terms", c.Careers)
	}

	if c.Age != 18 {
		t.Errorf("Age = %d, want 18", c.Age)
	}

	if c.WoundBadges != 0 {
		t.Errorf("WoundBadges = %d, want 0", c.WoundBadges)
	}

	// A never-qualified character never became a Noble at all — p.85's
	// "Nobles have a Base Fame..." doesn't apply, matching Scout's own
	// precedent that Fame stays 0 on a never-qualified path.
	if c.Fame != 0 {
		t.Errorf("Fame = %d, want 0 (never qualified, Base Fame shouldn't apply)", c.Fame)
	}
}

// TestBuildNobleCharacterQualified confirms a real career actually
// populates Age/LifeStage/Birthdate/Skills, mirroring the equivalent
// Scout/Citizen build tests. Soc 12, C2-C5 12: BeginNoble succeeds, and
// Intrigue's own target (2D<=12+mod) starts at a guaranteed success.
func TestBuildNobleCharacterQualified(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 12, 12, 12, 12, 12}}
	homeworldSkills := []SkillLevel{{Name: "Vacc Suit", Level: 1, Kind: Skill}}
	r := dice.New(rand.NewPCG(1, 1))

	c, ok := buildNobleCharacter(r, upp, "hw", homeworldSkills, nil)

	if !ok {
		t.Fatal("ok = false, want true (Soc qualifies and the fixture guarantees at least one term)")
	}

	if len(c.Careers) != 1 || c.Careers[0].Name != NobleCareerName {
		t.Fatalf("Careers = %+v, want one Career named %q", c.Careers, NobleCareerName)
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

	// >=, not ==: Mustering Out can additionally grant "Fame +2" on top of
	// Base Fame, and nobleExileFame adds +1 per Exile suffered along the
	// way — this fixture doesn't control for either, so an exact equality
	// would be flaky against real dice; both only ever add, never
	// subtract, so >= nobleBaseFame alone still holds.
	if want := nobleBaseFame(upp.Characteristics[C6]); c.Fame < want {
		t.Errorf("Fame = %d, want >= %d (nobleBaseFame)", c.Fame, want)
	}
}

// TestGenerateNobleCharacterElectsLaterEducation is #164's own
// end-to-end regression test, mirroring the pilot's
// TestGenerateRogueCharacterElectsLaterEducation (rogue_character_generate_test.go).
// Seed 556 confirmed by direct inspection to produce a one-term career
// that is itself a Later Education term (University) — the sharpest
// version of the Rank regression this wiring had to avoid: Return &
// Intrigue's Elevation logic never runs for this character at all, so
// Character.Rank has to come from the ladder position BeginNoble already
// established (Book 1 p.65: "Nobles begin with rank equal to their
// Social Standing"), not from anything term.Elevated touched. Before the
// fix (term.Rank scoped inside the skipped branch), this seed would have
// reported an empty Rank despite legitimately holding a title. Re-seeded
// from the original seed 33 when #165 (step C's own multi-institution
// escalation) shifted what that seed's pre-career Education looked like.
func TestGenerateNobleCharacterElectsLaterEducation(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(556, 556))

	c, ok := GenerateNobleCharacter(r)
	if !ok {
		t.Fatal("ok = false, want true")
	}

	if len(c.Careers[0].Terms) != 1 || !c.Careers[0].Terms[0].LaterEducation {
		t.Fatalf("Terms = %+v, want exactly one Later Education term (fixture assumption broke)", c.Careers[0].Terms)
	}

	term := c.Careers[0].Terms[0]

	if term.LaterEducationSchool == "" {
		t.Error("LaterEducationSchool is empty")
	}

	if len(term.SkillsAwarded) == 0 {
		t.Error("the Later Education term awarded no skills")
	}

	if c.Rank == "" {
		t.Error(
			"Rank is empty, want the ladder title BeginNoble established (a career that never Elevated still holds one)",
		)
	}
}
