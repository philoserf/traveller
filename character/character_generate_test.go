package character

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

func TestScoutWoundBadges(t *testing.T) {
	t.Parallel()

	career := Career{
		Terms: []Term{
			{RiskResult: Unharmed},
			{RiskResult: Wounded},
			{RiskResult: Disabled},
			{RiskResult: Wounded},
		},
	}

	if got, want := scoutWoundBadges(career), 3; got != want {
		t.Errorf("scoutWoundBadges() = %d, want %d", got, want)
	}
}

func TestAllSkillsFromTerms(t *testing.T) {
	t.Parallel()

	terms := []Term{
		{SkillsAwarded: []SkillLevel{{Name: "A", Level: 1, Kind: Skill}}},
		{SkillsAwarded: []SkillLevel{{Name: "B", Level: 1, Kind: Skill}, {Name: "C", Level: 1, Kind: Skill}}},
	}

	want := []SkillLevel{
		{Name: "A", Level: 1, Kind: Skill},
		{Name: "B", Level: 1, Kind: Skill},
		{Name: "C", Level: 1, Kind: Skill},
	}

	if got := allSkillsFromTerms(terms); !slices.Equal(got, want) {
		t.Errorf("allSkillsFromTerms() = %v, want %v", got, want)
	}

	if got := allSkillsFromTerms(nil); got != nil {
		t.Errorf("allSkillsFromTerms(nil) = %v, want nil", got)
	}
}

func TestBuildScoutCharacterNeverQualified(t *testing.T) {
	t.Parallel()

	upp := UPP{}
	homeworldSkills := []SkillLevel{{Name: "Vacc Suit", Level: 1, Kind: Skill}}
	r := dice.New(rand.NewPCG(1, 1))

	c, ok := buildScoutCharacter(r, upp, "A000000-0", homeworldSkills)

	if ok {
		t.Error("ok = true, want false (never qualified)")
	}

	if len(c.Careers) != 1 || len(c.Careers[0].Terms) != 0 {
		t.Errorf("Careers = %+v, want one Career with zero Terms", c.Careers)
	}

	if c.WoundBadges != 0 {
		t.Errorf("WoundBadges = %d, want 0", c.WoundBadges)
	}

	if c.UPP != upp {
		t.Errorf("UPP = %+v, want unchanged %+v", c.UPP, upp)
	}

	if !slices.Equal(c.Skills, homeworldSkills) {
		t.Errorf("Skills = %v, want exactly homeworldSkills %v (no career skills)", c.Skills, homeworldSkills)
	}
}

// TestBuildScoutCharacterDies confirms ok exactly matches the
// terms-derived rule (last term Dead => false) across many trials of a
// fixture where death is a frequent outcome, not just statistically
// probable — the equivalence is checked every trial, not sampled.
func TestBuildScoutCharacterDies(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{1, 1, 1, 12, 12, 0}}
	r := dice.New(rand.NewPCG(41, 43))

	sawDeath := false

	for range 200 {
		c, ok := buildScoutCharacter(r, upp, "hw", nil)

		terms := c.Careers[0].Terms
		if len(terms) == 0 {
			t.Fatal("career has zero terms, want Begin to always succeed via Retry (Edu=12)")
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

func TestBuildScoutCharacterFullTermSurvivor(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{12, 12, 12, 12, 12, 0}}
	r := dice.New(rand.NewPCG(7, 7))

	c, ok := buildScoutCharacter(r, upp, "hw", nil)

	if !ok {
		t.Error("ok = false, want true (immortal fixture)")
	}

	if got := len(c.Careers[0].Terms); got != maxScoutTerms {
		t.Errorf("len(Terms) = %d, want %d", got, maxScoutTerms)
	}

	if c.WoundBadges != 0 {
		t.Errorf("WoundBadges = %d, want 0 (Risk can never fail against target 12)", c.WoundBadges)
	}
}

// TestBuildScoutCharacterUPPCarriesForwardReduction confirms
// Character.UPP is the career-resolution-updated UPP, not the pre-career
// one, by comparing against a same-seed direct ResolveScoutCareer call.
func TestBuildScoutCharacterUPPCarriesForwardReduction(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{6, 6, 6, 12, 12, 0}}

	r1 := dice.New(rand.NewPCG(23, 29))
	r2 := dice.New(rand.NewPCG(23, 29))

	_, wantUPP := ResolveScoutCareer(r1, upp)
	c, _ := buildScoutCharacter(r2, upp, "hw", nil)

	if c.UPP != wantUPP {
		t.Errorf("Character.UPP = %+v, want %+v (ResolveScoutCareer's own updated UPP)", c.UPP, wantUPP)
	}
}

func TestBuildScoutCharacterHomeworldEqualsBirthworld(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(3, 4))

	c, _ := buildScoutCharacter(r, UPP{}, "some-uwp", nil)

	if c.Homeworld != "some-uwp" || c.Birthworld != "some-uwp" {
		t.Errorf("Homeworld = %q, Birthworld = %q, want both %q", c.Homeworld, c.Birthworld, "some-uwp")
	}
}

func TestBuildScoutCharacterSkillsIncludeHomeworldSkills(t *testing.T) {
	t.Parallel()

	homeworldSkills := []SkillLevel{{Name: "Vacc Suit", Level: 1, Kind: Skill}}

	// Never-qualified path: homeworld skills only, nothing from a career.
	r1 := dice.New(rand.NewPCG(1, 1))

	never, _ := buildScoutCharacter(r1, UPP{}, "hw", homeworldSkills)
	if !slices.Equal(never.Skills, homeworldSkills) {
		t.Errorf("never-qualified Skills = %v, want %v", never.Skills, homeworldSkills)
	}

	// Qualified path: homeworld skills still lead, career skills follow.
	upp := UPP{Characteristics: [6]ehex.Value{12, 12, 12, 12, 12, 0}}
	r2 := dice.New(rand.NewPCG(7, 7))

	qualified, _ := buildScoutCharacter(r2, upp, "hw", homeworldSkills)
	if len(qualified.Skills) < len(homeworldSkills) ||
		!slices.Equal(qualified.Skills[:len(homeworldSkills)], homeworldSkills) {
		t.Errorf("qualified Skills = %v, want to start with homeworldSkills %v", qualified.Skills, homeworldSkills)
	}
}

// TestBuildScoutCharacterDoesNotAliasHomeworldSkills is a regression test:
// buildScoutCharacter must not append onto the caller's homeworldSkills
// slice in place — doing so can silently corrupt an earlier Character's
// Skills when the same slice (with spare backing-array capacity) is
// reused across multiple calls, exactly the pattern
// TestBuildScoutCharacterSkillsIncludeHomeworldSkills's own shared
// homeworldSkills fixture demonstrates is a natural thing to do.
func TestBuildScoutCharacterDoesNotAliasHomeworldSkills(t *testing.T) {
	t.Parallel()

	shared := make([]SkillLevel, 1, 64) // deliberate spare capacity
	shared[0] = SkillLevel{Name: "Vacc Suit", Level: 1, Kind: Skill}

	upp := UPP{Characteristics: [6]ehex.Value{12, 12, 12, 12, 12, 0}}

	r1 := dice.New(rand.NewPCG(7, 7))
	c1, _ := buildScoutCharacter(r1, upp, "hw", shared)

	before := slices.Clone(c1.Skills)

	r2 := dice.New(rand.NewPCG(9, 9))
	_, _ = buildScoutCharacter(r2, upp, "hw", shared)

	if !slices.Equal(c1.Skills, before) {
		t.Errorf(
			"c1.Skills changed after a second buildScoutCharacter call reusing the same homeworldSkills slice: %v -> %v",
			before,
			c1.Skills,
		)
	}
}

// TestBuildScoutCharacterFixedZeroValueFields pins Species/GeneticProfile
// and the full "left at zero value" contract, so a future change can't
// silently half-populate one of these without a test forcing a
// doc-comment update.
func TestBuildScoutCharacterFixedZeroValueFields(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}
	r := dice.New(rand.NewPCG(5, 5))

	c, _ := buildScoutCharacter(r, upp, "hw", nil)

	if c.Species != "Human" {
		t.Errorf("Species = %q, want %q", c.Species, "Human")
	}

	if c.GeneticProfile != "SDEIES" {
		t.Errorf("GeneticProfile = %q, want %q", c.GeneticProfile, "SDEIES")
	}

	switch {
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

func TestGenerateScoutCharacterDeterminism(t *testing.T) {
	t.Parallel()

	r1 := dice.New(rand.NewPCG(99, 99))
	r2 := dice.New(rand.NewPCG(99, 99))

	c1, ok1 := GenerateScoutCharacter(r1)
	c2, ok2 := GenerateScoutCharacter(r2)

	if ok1 != ok2 {
		t.Fatalf("identical seeds produced different ok: %v vs %v", ok1, ok2)
	}

	if c1.Species != c2.Species || c1.GeneticProfile != c2.GeneticProfile {
		t.Fatalf("identical seeds produced different Species/GeneticProfile: %+v vs %+v", c1, c2)
	}

	if c1.UPP != c2.UPP {
		t.Fatalf("identical seeds produced different UPP: %+v vs %+v", c1.UPP, c2.UPP)
	}

	if c1.Homeworld != c2.Homeworld || c1.Birthworld != c2.Birthworld {
		t.Fatalf("identical seeds produced different Homeworld/Birthworld: %+v vs %+v", c1, c2)
	}

	if c1.WoundBadges != c2.WoundBadges {
		t.Fatalf("identical seeds produced different WoundBadges: %d vs %d", c1.WoundBadges, c2.WoundBadges)
	}

	if len(c1.Careers[0].Terms) != len(c2.Careers[0].Terms) {
		t.Fatalf("identical seeds produced different term counts: %d vs %d",
			len(c1.Careers[0].Terms), len(c2.Careers[0].Terms))
	}

	if !slices.Equal(c1.Skills, c2.Skills) {
		t.Fatalf("identical seeds produced different Skills: %v vs %v", c1.Skills, c2.Skills)
	}
}

// TestGenerateScoutCharacterManySeedsInvariants is a lightweight smoke
// test over the wrapper wiring across many real seeds: no panics, and the
// two cheap-to-check invariants hold everywhere.
func TestGenerateScoutCharacterManySeedsInvariants(t *testing.T) {
	t.Parallel()

	for seed := range uint64(500) {
		r := dice.New(rand.NewPCG(seed+1, seed+1))

		c, ok := GenerateScoutCharacter(r)

		terms := c.Careers[0].Terms

		wantOK := len(terms) > 0 && terms[len(terms)-1].RiskResult != Dead
		if ok != wantOK {
			t.Fatalf("seed %d: ok = %v, want %v", seed, ok, wantOK)
		}

		if c.WoundBadges > len(terms) {
			t.Fatalf("seed %d: WoundBadges = %d, want <= %d", seed, c.WoundBadges, len(terms))
		}
	}
}
