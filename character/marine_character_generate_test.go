package character

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

func TestBuildMarineCharacterNeverQualified(t *testing.T) {
	t.Parallel()

	upp := UPP{}
	homeworldSkills := []SkillLevel{{Name: "Vacc Suit", Level: 1, Kind: Skill}}
	r := dice.New(rand.NewPCG(1, 1))

	c, ok := buildMarineCharacter(r, upp, "A000000-0", homeworldSkills)

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

// TestBuildMarineCharacterDies confirms ok exactly matches the last
// term's own RiskResult != Dead — mirroring TestBuildScoutCharacterDies.
// C1=12 guarantees Begin always succeeds (Marine has no Retry, unlike
// Scout, so Begin needs to be reliable on its own); C4=1 makes death
// likely whenever the C1/C4 rotation lands on C4 (Risk target =
// 1-combinedMod, effectively always <=0).
func TestBuildMarineCharacterDies(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{12, 0, 0, 1, 8, 0}}
	r := dice.New(rand.NewPCG(41, 43))

	sawDeath := false

	for range 200 {
		c, ok := buildMarineCharacter(r, upp, "hw", nil)

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

// TestBuildMarineCharacterQualified confirms a real career actually
// populates Age/LifeStage/Birthdate/Skills/WoundBadges, mirroring the
// equivalent Scout/Citizen/Noble build tests. C1=C4=20 is high enough
// that even the worst-case combined Mod (Branch's own max 2 +
// Operations' own max 3 = 5) still leaves a target above 2D6's own
// maximum, guaranteeing survival and Continue (see
// TestResolveMarineCareerRespectsMaxTermsCap's own identical reasoning).
func TestBuildMarineCharacterQualified(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{20, 0, 0, 20, 8, 0}}
	homeworldSkills := []SkillLevel{{Name: "Vacc Suit", Level: 1, Kind: Skill}}
	r := dice.New(rand.NewPCG(1, 1))

	c, ok := buildMarineCharacter(r, upp, "hw", homeworldSkills)

	if !ok {
		t.Fatal("ok = false, want true (fixture guarantees survival)")
	}

	if len(c.Careers) != 1 || c.Careers[0].Name != MarineCareerName {
		t.Fatalf("Careers = %+v, want one Career named %q", c.Careers, MarineCareerName)
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
