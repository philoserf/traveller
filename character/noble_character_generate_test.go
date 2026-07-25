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

	c, ok := buildNobleCharacter(r, upp, "hw", nil)

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

	c, ok := buildNobleCharacter(r, upp, "hw", homeworldSkills)

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
}
