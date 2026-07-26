package character

import (
	"math/rand/v2"
	"slices"
	"strings"
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

func TestScoutDiscoveryFame(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		terms []Term
		want  int
	}{
		{"no terms", nil, 0},
		{"no discoveries", []Term{{RewardResult: "None"}, {RewardResult: ""}}, 0},
		{"one discovery", []Term{{RewardResult: "Discovery"}}, 4},
		{
			"mixed",
			[]Term{
				{RewardResult: "Discovery"},
				{RewardResult: "None"},
				{RewardResult: "Discovery"},
				{RewardResult: ""},
			},
			8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := scoutDiscoveryFame(Career{Terms: tt.terms}); got != tt.want {
				t.Errorf("scoutDiscoveryFame() = %d, want %d", got, tt.want)
			}
		})
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

// TestBuildScoutCharacterDies confirms ok exactly matches the two rules
// that can each independently kill a character, across many trials of a
// fixture where death is a frequent outcome, not just statistically
// probable — the equivalence is checked every trial, not sampled. Those
// rules are a Dead last term (Book 1 p.69, Career Resolution) and an
// Aging death (p.89, finalizeAging); ok is false when either fires.
// This fixture's own Str/Dex/End of 1 makes the second one common, not
// hypothetical: Aging reduces exactly those three, so there's almost no
// margin before a checkpoint zeroes them.
func TestBuildScoutCharacterDies(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{1, 1, 1, 12, 12, 0}}
	r := dice.New(rand.NewPCG(41, 43))

	sawCareerDeath, sawAgingDeath := false, false

	for range 200 {
		c, ok := buildScoutCharacter(r, upp, "hw", nil)

		terms := c.Careers[0].Terms
		if len(terms) == 0 {
			t.Fatal("career has zero terms, want Begin to always succeed via Retry (Edu=12)")
		}

		diedInCareer := terms[len(terms)-1].RiskResult == Dead
		diedOfAging := strings.Contains(c.Notes, "died of natural causes")

		wantOK := !diedInCareer && !diedOfAging
		if ok != wantOK {
			t.Fatalf("ok = %v, want %v (last term RiskResult = %v, Notes = %q)",
				ok, wantOK, terms[len(terms)-1].RiskResult, c.Notes)
		}

		sawCareerDeath = sawCareerDeath || diedInCareer
		sawAgingDeath = sawAgingDeath || diedOfAging
	}

	if !sawCareerDeath {
		t.Error("no trial produced a career death across 200 trials — fixture can't verify that ok=false path")
	}

	if !sawAgingDeath {
		t.Error("no trial produced an Aging death across 200 trials — fixture can't verify that ok=false path")
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

	if got := len(c.Careers[0].Terms); got != maxCareerTerms {
		t.Errorf("len(Terms) = %d, want %d", got, maxCareerTerms)
	}

	if c.WoundBadges != 0 {
		t.Errorf("WoundBadges = %d, want 0 (Risk can never fail against target 12)", c.WoundBadges)
	}
}

// TestBuildScoutCharacterUPPCarriesForwardReduction confirms
// buildScoutCharacter feeds finalizeAging the career-resolution-and-
// Mustering-Out-updated UPP, not the stale pre-career one — by replaying
// the exact same pipeline (ResolveScoutCareer -> ResolveScoutMusterOut ->
// ApplyMusteringOut -> finalizeAging) against an independent same-seed
// roller and comparing. A regression here (e.g. passing the original upp
// instead of ResolveScoutCareer's own returned UPP into ApplyMusteringOut
// or finalizeAging) is exactly the class of copy-paste bug Phase F's own
// aliasing-bug code review finding was about.
func TestBuildScoutCharacterUPPCarriesForwardReduction(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{6, 6, 6, 12, 12, 0}}

	r1 := dice.New(rand.NewPCG(23, 29))

	var wantAging agingSimulation

	wantCareer, updatedUPP := resolveScoutCareerWithBudget(r1, upp, maxCareerTerms, &wantAging)
	wantCareer.MusteringOut = ResolveScoutMusterOut(r1, wantCareer)
	wantUPP, _ := ApplyMusteringOut(wantCareer, updatedUPP)
	wantOK := len(wantCareer.Terms) > 0 && wantCareer.Terms[len(wantCareer.Terms)-1].RiskResult != Dead
	wantAge, wantLifeStage, wantNotes, _ := finalizeAging(&wantAging, wantOK)

	r2 := dice.New(rand.NewPCG(23, 29))
	c, _ := buildScoutCharacter(r2, upp, "hw", nil)

	if c.UPP != wantUPP {
		t.Errorf("Character.UPP = %+v, want %+v (career resolution's own updated UPP, aged forward)", c.UPP, wantUPP)
	}

	if c.Age != wantAge || c.LifeStage != wantLifeStage || c.Notes != wantNotes {
		t.Errorf("got Age=%d LifeStage=%d Notes=%q, want Age=%d LifeStage=%d Notes=%q",
			c.Age, c.LifeStage, c.Notes, wantAge, wantLifeStage, wantNotes)
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
	// Seed 1 confirmed by direct inspection to re-grant "Vacc Suit"
	// during the (immortal, 14-term) career itself, so aggregateSkills
	// merging is actually exercised — pinning the exact merged Level
	// (not just ">= 1") means a regression that stops calling
	// aggregateSkills (character_generate.go's own buildRiskCareerCharacter)
	// would leave Skills[0] at Level 1 and fail this assertion, rather
	// than passing trivially.
	upp := UPP{Characteristics: [6]ehex.Value{12, 12, 12, 12, 12, 0}}
	r2 := dice.New(rand.NewPCG(1, 1))

	qualified, _ := buildScoutCharacter(r2, upp, "hw", homeworldSkills)

	want := SkillLevel{Name: "Vacc Suit", Level: 3, Kind: Skill}
	if len(qualified.Skills) == 0 || qualified.Skills[0] != want {
		t.Errorf("qualified Skills[0] = %+v, want %+v (merged with a later in-career grant)", qualified.Skills[0], want)
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
// and the remaining "left at zero value" contract, so a future change
// can't silently half-populate one of these without a test forcing a
// doc-comment update. Fame/Cash are no longer in this list — they're
// genuinely computed now (ApplyMusteringOut), just not by this seed's
// fixture; see TestBuildScoutCharacterAppliesMusteringOutFameAndCash.
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
	case c.Equipment != nil:
		t.Errorf("Equipment = %v, want nil", c.Equipment)
	case c.Name != "":
		t.Errorf("Name = %q, want empty", c.Name)
	}
}

// TestBuildScoutCharacterSetsAgeAndLifeStage confirms Age/LifeStage are
// actually computed now (finalizeAging), including for a never-qualified
// attempt (0 terms served, still a meaningful "age 18, never got in" fact).
func TestBuildScoutCharacterSetsAgeAndLifeStage(t *testing.T) {
	t.Parallel()

	upp := UPP{}
	r := dice.New(rand.NewPCG(1, 1))

	c, ok := buildScoutCharacter(r, upp, "hw", nil)
	if ok {
		t.Fatalf("buildScoutCharacter with zero UPP unexpectedly qualified")
	}

	// 0 terms served, but not age 18: Scout is the one career with a
	// Retry, so a character who qualifies for neither has failed two
	// rolls, and Book 1 p.65 charges a year for each.
	if c.Age != 20 {
		t.Errorf("Age = %d, want 20 (18 plus a year each for the failed Begin and Retry)", c.Age)
	}

	if c.LifeStage != 3 {
		t.Errorf("LifeStage = %d, want 3 (Young Adult)", c.LifeStage)
	}

	if c.Notes != "" {
		t.Errorf("Notes = %q, want empty (Aging not simulated for a never-qualified attempt)", c.Notes)
	}
}

// TestBuildScoutCharacterSetsBirthdate confirms Birthdate is actually
// computed now (GenerateBirthdate), including for a never-qualified
// attempt (Age is still set, so a Birthdate can still be computed).
func TestBuildScoutCharacterSetsBirthdate(t *testing.T) {
	t.Parallel()

	upp := UPP{}
	r := dice.New(rand.NewPCG(1, 1))

	c, _ := buildScoutCharacter(r, upp, "hw", nil)

	assertBirthdateFormat(t, c.Birthdate, c.Age)
}

// TestBuildRiskCareerCharacterGatesCareerFameOnOK is the regression test
// for a code-review-caught bug: careerFame was being added unconditionally,
// so a character who earned Fame in an early term (e.g. a Scout Discovery)
// but then died in a later one retained that Fame despite ok=false — unlike
// buildNobleCharacter's own analogous Fame terms, which are gated on ok.
// Uses buildRiskCareerCharacter directly with a stub resolveCareer so the
// Discovery-then-Death sequence is pinned exactly, not seed-hunted.
func TestBuildRiskCareerCharacterGatesCareerFameOnOK(t *testing.T) {
	t.Parallel()

	deadCareer := Career{
		Terms: []Term{
			{RewardResult: "Discovery", RiskResult: Unharmed},
			{RiskResult: Dead},
		},
	}

	resolveCareer := func(_ *dice.Roller, upp UPP, _ *agingSimulation) (Career, UPP) { return deadCareer, upp }
	resolveMusterOut := func(_ *dice.Roller, _ Career) MusteringOut { return MusteringOut{} }

	r := dice.New(rand.NewPCG(1, 1))

	c, ok := buildRiskCareerCharacter(r, UPP{}, "hw", nil, resolveCareer, resolveMusterOut, scoutDiscoveryFame)

	if ok {
		t.Fatal("ok = true, want false (fixture's last term is Dead)")
	}

	if c.Fame != 0 {
		t.Errorf("Fame = %d, want 0 (careerFame must not apply to a character who died mid-career)", c.Fame)
	}
}

// TestBuildScoutCharacterAppliesMusteringOutFameAndCash confirms
// ApplyMusteringOut is actually wired into buildScoutCharacter — seed 46
// with this fixture was found by direct search to produce both
// Fame-bearing and Cash-bearing Mustering Out rolls. This seed's own
// career also produces two Discoveries (scoutDiscoveryFame's own +4
// each), so the total is 12, not Mustering Out's own 4 alone — the
// regression test for careerFame actually being summed in, not just
// bonuses.Fame.
func TestBuildScoutCharacterAppliesMusteringOutFameAndCash(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}
	r := dice.New(rand.NewPCG(46, 46))

	c, ok := buildScoutCharacter(r, upp, "hw", nil)
	if !ok {
		t.Fatalf("seed 46: buildScoutCharacter unexpectedly failed (fixture assumption broke)")
	}

	if c.Fame != 12 {
		t.Errorf("Fame = %d, want 12 (8 from two Discoveries + 4 from Mustering Out)", c.Fame)
	}

	if c.Cash != 140000 {
		t.Errorf("Cash = %d, want 140000", c.Cash)
	}
}

// TestBuildScoutCharacterAgingBufferNeverTriggersIllness reuses
// ResolveAging's own dice-outcome-independent buffer trick
// (aging_test.go's TestResolveAgingNeverTriggersIllnessWithSufficientBuffer):
// characteristics start high enough (15) that even the maximum possible
// Aging reduction over maxCareerTerms terms can't reach 0, proving
// finalizeAging actually calls ResolveAging (UPP still bounded, no
// illness/death notes) without needing to control real dice outcomes.
func TestBuildScoutCharacterAgingBufferNeverTriggersIllness(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{15, 15, 15, 15, 15, 0}}

	for _, seed := range []uint64{1, 2, 3} {
		r := dice.New(rand.NewPCG(seed, seed))

		c, ok := buildScoutCharacter(r, upp, "hw", nil)
		if !ok {
			t.Fatalf("seed %d: buildScoutCharacter with immortal fixture unexpectedly died", seed)
		}

		if c.Notes != "" {
			t.Errorf("seed %d: Notes = %q, want empty", seed, c.Notes)
		}

		for i, v := range c.UPP.Characteristics[:5] {
			// Upper bound is the Human cap, not ehex.Max. This test was
			// relaxed to ehex.Max when Personal awards began raising
			// characteristics; that hid awards carrying a Human past 15.
			if v < 4 || int(v) > HumanCharacteristicMax {
				t.Errorf("seed %d: Characteristics[%d] = %d, want in [4, %d]",
					seed, i, v, HumanCharacteristicMax)
			}
		}
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

// TestBuildScoutCharacterAgingDeathKeepsCareerFame is the regression for
// PR #69's own review finding: Fame earned by completing a career must
// survive a later Aging death. The two answer different questions —
// Fame asks "did Career Resolution succeed?" (Book 1 p.69, whose rule
// voids a mid-career death's own rewards), while ok additionally asks
// "is the character alive at the end of generation?" (p.89 Aging).
// Collapsing both onto one variable silently zeroed this character's
// Fame, and disagreed with the chain generator, which accumulates each
// segment's Fame before Aging is ever simulated.
//
// Seed 4972 also pins the Age/career-timeline coherence the same review
// asked for: 13 terms served, Age 70, and 18+4*13 == 70 exactly.
func TestBuildScoutCharacterAgingDeathKeepsCareerFame(t *testing.T) {
	t.Parallel()

	c, ok := GenerateScoutCharacter(dice.New(rand.NewPCG(4972, 4972)))

	if !strings.Contains(c.Notes, "died of natural causes") {
		t.Fatalf("Notes = %q, want an Aging death (fixture assumption broke)", c.Notes)
	}

	if ok {
		t.Error("ok = true, want false (an Aging death is not a surviving character)")
	}

	terms := c.Careers[0].Terms
	if terms[len(terms)-1].RiskResult == Dead {
		t.Fatal("last term is Dead, want a completed career (fixture assumption broke)")
	}

	if c.Fame == 0 {
		t.Error("Fame = 0, want the completed career's own Fame retained despite the later Aging death")
	}

	if want := AgeFromTermsServed(len(terms)); c.Age != want {
		t.Errorf("Age = %d, want %d — Age and the career timeline must agree, since Aging now runs "+
			"between terms and simply stops the career when it kills someone", c.Age, want)
	}
}
