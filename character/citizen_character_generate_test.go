package character

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// TestBuildCitizenCharacterUPPExactlyUnchangedBelowAgingOnset confirms
// Character.UPP is exactly the pre-career UPP (not just "close to it")
// when the character's approximated Age never reaches Physical Aging's
// own onset (34): Citizen Life itself never modifies a characteristic
// (ResolveCitizenCareer's own signature only returns Career, nothing to
// mutate upp with), and finalizeAging's own ResolveAging call is a
// structural no-op below onset regardless of dice (agingCheckpoints
// returns empty). Seed 4 was found by direct search to produce exactly 1
// Citizen term (Age 22): 18 + 4*1 = 22 < 34.
func TestBuildCitizenCharacterUPPExactlyUnchangedBelowAgingOnset(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}
	r := dice.New(rand.NewPCG(4, 4))

	c := buildCitizenCharacter(r, upp, "hw", nil)

	if len(c.Careers[0].Terms) != 1 {
		t.Fatalf("seed 4: len(Terms) = %d, want 1 (fixture assumption broke)", len(c.Careers[0].Terms))
	}

	if c.UPP != upp {
		t.Errorf("UPP = %+v, want unchanged %+v (Age %d is below Physical Aging's own onset of 34)", c.UPP, upp, c.Age)
	}
}

// TestBuildCitizenCharacterUPPBoundedWithAgingBuffer confirms
// finalizeAging is actually wired into buildCitizenCharacter for a
// long-enough career to cross Aging's onset — using the same
// dice-outcome-independent buffer trick as aging_test.go's
// TestResolveAgingNeverTriggersIllnessWithSufficientBuffer: starting
// characteristics high enough (15) that even the maximum possible Aging
// reduction over maxCareerTerms terms can't reach 0, so Notes stays
// empty and every characteristic stays within the provable bound —
// without needing to control real dice outcomes.
func TestBuildCitizenCharacterUPPBoundedWithAgingBuffer(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{15, 15, 15, 15, 15, 0}}

	for _, seed := range []uint64{1, 2, 3} {
		r := dice.New(rand.NewPCG(seed, seed))

		c := buildCitizenCharacter(r, upp, "hw", nil)

		for i, v := range c.UPP.Characteristics[:5] {
			if v < 4 || v > 15 {
				t.Errorf("seed %d: Characteristics[%d] = %d, want in [4, 15]", seed, i, v)
			}
		}

		if c.Notes != "" {
			t.Errorf("seed %d: Notes = %q, want empty", seed, c.Notes)
		}
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
	}
}

// TestBuildCitizenCharacterSetsAgeAndLifeStage confirms Age/LifeStage are
// actually computed now (finalizeAging) from termsServed, mirroring
// TestBuildScoutCharacterSetsAgeAndLifeStage. Age is pinned to an exact,
// independently-known value (18 + 4*termsServed, not re-derived from
// c.Age itself) so a regression that wires the wrong termsServed into
// finalizeAging — e.g. passing 0 instead of len(career.Terms) — would
// still be caught. Seed 5 was found by direct search to produce exactly
// 8 Citizen terms (Age 18 + 4*8 = 50, Life Stage 7 Senior).
func TestBuildCitizenCharacterSetsAgeAndLifeStage(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}
	r := dice.New(rand.NewPCG(5, 5))

	c := buildCitizenCharacter(r, upp, "hw", nil)

	if len(c.Careers[0].Terms) != 8 {
		t.Fatalf("seed 5: len(Terms) = %d, want 8 (fixture assumption broke)", len(c.Careers[0].Terms))
	}

	if c.Age != 50 {
		t.Errorf("Age = %d, want 50 (18 + 4*8 terms served)", c.Age)
	}

	if c.LifeStage != 7 {
		t.Errorf("LifeStage = %d, want 7 (Senior)", c.LifeStage)
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
