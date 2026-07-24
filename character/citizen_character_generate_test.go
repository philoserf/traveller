package character

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

func TestBuildCitizenCharacterUPPIsUnchanged(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{6, 7, 8, 9, 0, 0}}
	r := dice.New(rand.NewPCG(1, 1))

	c := buildCitizenCharacter(r, upp, "hw", nil)

	if c.UPP != upp {
		t.Errorf("UPP = %+v, want unchanged %+v (Citizen Life never modifies a characteristic)", c.UPP, upp)
	}
}

func TestBuildCitizenCharacterHomeworldEqualsBirthworld(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(3, 4))

	c := buildCitizenCharacter(r, UPP{}, "some-uwp", nil)

	if c.Homeworld != "some-uwp" || c.Birthworld != "some-uwp" {
		t.Errorf("Homeworld = %q, Birthworld = %q, want both %q", c.Homeworld, c.Birthworld, "some-uwp")
	}
}

func TestBuildCitizenCharacterSkillsIncludeHomeworldSkills(t *testing.T) {
	t.Parallel()

	homeworldSkills := []SkillLevel{{Name: "Vacc Suit", Level: 1, Kind: Skill}}

	r := dice.New(rand.NewPCG(7, 7))
	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}

	c := buildCitizenCharacter(r, upp, "hw", homeworldSkills)
	if len(c.Skills) < len(homeworldSkills) || !slices.Equal(c.Skills[:len(homeworldSkills)], homeworldSkills) {
		t.Errorf("Skills = %v, want to start with homeworldSkills %v", c.Skills, homeworldSkills)
	}
}

// TestBuildCitizenCharacterDoesNotAliasHomeworldSkills is the same
// regression test Phase F's own review added for
// TestBuildScoutCharacterDoesNotAliasHomeworldSkills — replicated here
// since buildCitizenCharacter repeats the identical clone pattern.
func TestBuildCitizenCharacterDoesNotAliasHomeworldSkills(t *testing.T) {
	t.Parallel()

	shared := make([]SkillLevel, 1, 64) // deliberate spare capacity
	shared[0] = SkillLevel{Name: "Vacc Suit", Level: 1, Kind: Skill}

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}

	r1 := dice.New(rand.NewPCG(7, 7))
	c1 := buildCitizenCharacter(r1, upp, "hw", shared)

	before := slices.Clone(c1.Skills)

	r2 := dice.New(rand.NewPCG(9, 9))
	_ = buildCitizenCharacter(r2, upp, "hw", shared)

	if !slices.Equal(c1.Skills, before) {
		t.Errorf(
			"c1.Skills changed after a second buildCitizenCharacter call reusing the same homeworldSkills slice: %v -> %v",
			before,
			c1.Skills,
		)
	}
}

// TestBuildCitizenCharacterFixedZeroValueFields pins Species/GeneticProfile,
// WoundBadges, and the full "left at zero value" contract, mirroring
// TestBuildScoutCharacterFixedZeroValueFields.
func TestBuildCitizenCharacterFixedZeroValueFields(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}
	r := dice.New(rand.NewPCG(5, 5))

	c := buildCitizenCharacter(r, upp, "hw", nil)

	if c.Species != "Human" {
		t.Errorf("Species = %q, want %q", c.Species, "Human")
	}

	if c.GeneticProfile != "SDEIES" {
		t.Errorf("GeneticProfile = %q, want %q", c.GeneticProfile, "SDEIES")
	}

	switch {
	case c.WoundBadges != 0:
		t.Errorf("WoundBadges = %d, want 0 (Citizen Life has no wound mechanic)", c.WoundBadges)
	case c.Rank != "":
		t.Errorf("Rank = %q, want empty", c.Rank)
	case c.Medals != nil:
		t.Errorf("Medals = %v, want nil", c.Medals)
	case c.Commendations != nil:
		t.Errorf("Commendations = %v, want nil", c.Commendations)
	case c.Fame != 0:
		t.Errorf("Fame = %d, want 0", c.Fame)
	case c.Cash != 0:
		t.Errorf("Cash = %d, want 0", c.Cash)
	case c.Equipment != nil:
		t.Errorf("Equipment = %v, want nil", c.Equipment)
	case c.Name != "":
		t.Errorf("Name = %q, want empty", c.Name)
	case c.Birthdate != "":
		t.Errorf("Birthdate = %q, want empty", c.Birthdate)
	case c.Age != 0:
		t.Errorf("Age = %d, want 0", c.Age)
	case c.LifeStage != 0:
		t.Errorf("LifeStage = %d, want 0", c.LifeStage)
	case c.Notes != "":
		t.Errorf("Notes = %q, want empty", c.Notes)
	}
}

func TestGenerateCitizenCharacterDeterminism(t *testing.T) {
	t.Parallel()

	r1 := dice.New(rand.NewPCG(99, 99))
	r2 := dice.New(rand.NewPCG(99, 99))

	c1 := GenerateCitizenCharacter(r1)
	c2 := GenerateCitizenCharacter(r2)

	if c1.Species != c2.Species || c1.GeneticProfile != c2.GeneticProfile {
		t.Fatalf("identical seeds produced different Species/GeneticProfile: %+v vs %+v", c1, c2)
	}

	if c1.UPP != c2.UPP {
		t.Fatalf("identical seeds produced different UPP: %+v vs %+v", c1.UPP, c2.UPP)
	}

	if c1.Homeworld != c2.Homeworld || c1.Birthworld != c2.Birthworld {
		t.Fatalf("identical seeds produced different Homeworld/Birthworld: %+v vs %+v", c1, c2)
	}

	if c1.Careers[0].JobSkill != c2.Careers[0].JobSkill || c1.Careers[0].HobbySkill != c2.Careers[0].HobbySkill {
		t.Fatalf("identical seeds produced different Job/Hobby: %+v vs %+v", c1.Careers[0], c2.Careers[0])
	}

	if len(c1.Careers[0].Terms) != len(c2.Careers[0].Terms) {
		t.Fatalf("identical seeds produced different term counts: %d vs %d",
			len(c1.Careers[0].Terms), len(c2.Careers[0].Terms))
	}

	if !slices.Equal(c1.Skills, c2.Skills) {
		t.Fatalf("identical seeds produced different Skills: %v vs %v", c1.Skills, c2.Skills)
	}
}

// TestGenerateCitizenCharacterManySeedsInvariants is a lightweight smoke
// test over the wrapper wiring across many real seeds: no panics, and
// the one invariant this slice's own design relies on — a Citizen
// attempt can never fail to produce at least one term — holds
// everywhere.
func TestGenerateCitizenCharacterManySeedsInvariants(t *testing.T) {
	t.Parallel()

	for seed := range uint64(500) {
		r := dice.New(rand.NewPCG(seed+1, seed+1))

		c := GenerateCitizenCharacter(r)

		terms := c.Careers[0].Terms
		if len(terms) < 1 {
			t.Fatalf("seed %d: len(Terms) = %d, want >= 1 (Begin is Auto)", seed, len(terms))
		}

		if c.WoundBadges != 0 {
			t.Fatalf("seed %d: WoundBadges = %d, want 0", seed, c.WoundBadges)
		}
	}
}
