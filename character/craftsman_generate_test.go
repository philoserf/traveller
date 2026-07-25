package character

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

func TestCraftsmanSkillTableMatchesBook1P75(t *testing.T) {
	t.Parallel()

	want := [7][6]string{
		{"Str", "C2", "C3", "Int", "C5", "C6"},
		{"Major", "Major", "Minor", "Minor", "One Trade", "One Trade"},
		{"Seafarer", "Navigation", "Hostile Environment", "Flyer", "Driver", "Vacc Suit"},
		{"Animals", "Comms", "Computer", "Designer", "Designer", "Designer"},
		{"Liaison", "Comms", "Bureaucrat", "Diplomat", "Leader", "Trader"},
		{"Naval Architect", "One Art", "New Trade", "New Trade", "One Trade", "One Trade"},
		{"Animals", "One Art", "One Science", "Athlete", "Medic", "One Trade"},
	}

	if craftsmanSkillTable != want {
		t.Errorf("craftsmanSkillTable = %v, want %v", craftsmanSkillTable, want)
	}
}

func TestCraftsmanQualifyingSkillsExcludesLanguageAndCraftsmanItself(t *testing.T) {
	t.Parallel()

	skills := []SkillLevel{
		{Name: "Craftsman", Level: 8, Kind: Skill},
		{Name: "Language", Level: 8, Kind: Skill},
		{Name: "Pilot", Level: 8, Kind: Skill},
		{Name: "Astrogation", Level: 5, Kind: Skill}, // below 6, excluded
		{Name: "Ancient History", Level: 6, Kind: Knowledge},
		{Name: "Str", Level: 8, Kind: Personal}, // wrong Kind, excluded
	}

	got := craftsmanQualifyingSkills(skills)
	if len(got) != 2 {
		t.Fatalf("craftsmanQualifyingSkills(...) = %+v, want exactly 2 entries (Pilot, Ancient History)", got)
	}

	names := map[string]bool{got[0].Name: true, got[1].Name: true}
	if !names["Pilot"] || !names["Ancient History"] {
		t.Errorf("craftsmanQualifyingSkills(...) = %+v, want Pilot and Ancient History", got)
	}
}

func TestCraftsmanSkillLevelNeverHeld(t *testing.T) {
	t.Parallel()

	if got := craftsmanSkillLevel(nil); got != 0 {
		t.Errorf("craftsmanSkillLevel(nil) = %d, want 0", got)
	}
}

func TestBeginCraftsman(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		skills []SkillLevel
		want   bool
	}{
		{"no skills at all", nil, false},
		{
			"craftsman held but only one qualifying skill",
			[]SkillLevel{
				{Name: "Craftsman", Level: 1, Kind: Skill},
				{Name: "Pilot", Level: 8, Kind: Skill},
			},
			false,
		},
		{
			"two qualifying skills but craftsman never held",
			[]SkillLevel{
				{Name: "Pilot", Level: 8, Kind: Skill},
				{Name: "Admin", Level: 6, Kind: Skill},
			},
			false,
		},
		{
			"meets both prerequisites",
			[]SkillLevel{
				{Name: "Craftsman", Level: 1, Kind: Skill},
				{Name: "Pilot", Level: 8, Kind: Skill},
				{Name: "Admin", Level: 6, Kind: Skill},
			},
			true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := BeginCraftsman(c.skills); got != c.want {
				t.Errorf("BeginCraftsman(%+v) = %v, want %v", c.skills, got, c.want)
			}
		})
	}
}

// TestCraftsmanMasterPointsPicksFiveHighest confirms "up to FIVE" caps
// at the five highest-level qualifying skills, not the first five
// encountered or all of them.
func TestCraftsmanMasterPointsPicksFiveHighest(t *testing.T) {
	t.Parallel()

	skills := []SkillLevel{
		{Name: "Craftsman", Level: 2, Kind: Skill},
		{Name: "A", Level: 6, Kind: Skill},
		{Name: "B", Level: 7, Kind: Skill},
		{Name: "C", Level: 8, Kind: Skill},
		{Name: "D", Level: 9, Kind: Skill},
		{Name: "E", Level: 10, Kind: Skill},
		{Name: "F", Level: 11, Kind: Skill}, // sixth qualifying skill, excluded by the "up to five" cap
	}

	// cc=1 + craftsman=2 + top five (11+10+9+8+7=45) = 48.
	if got := craftsmanMasterPoints(1, skills); got != 48 {
		t.Errorf("craftsmanMasterPoints(1, ...) = %d, want 48 (top 5 of 6 qualifying skills, not all 6)", got)
	}
}

func TestCraftsmanMasterpieceValue(t *testing.T) {
	t.Parallel()

	cases := []struct {
		points int
		want   int
	}{
		{40, 150000},
		{50, 250000},
		{54, 290000}, // one point below Perfect
		{55, 600000}, // Perfect: doubled, not 300,000
		{63, 760000},
	}

	for _, c := range cases {
		if got := craftsmanMasterpieceValue(c.points); got != c.want {
			t.Errorf("craftsmanMasterpieceValue(%d) = %d, want %d", c.points, got, c.want)
		}
	}
}

func TestContinueCraftsmanBoundaries(t *testing.T) {
	t.Parallel()

	// A natural roll below/at/above the 2*level target, tested via the
	// same rollAgainstTarget helper every other career's own plain-roll
	// Continue check already relies on — confirmed indirectly by
	// checking a level of 0 (target 0) always fails and a level high
	// enough to guarantee the maximum 2D6 roll always succeeds.
	if continueCraftsman(dice.New(rand.NewPCG(1, 1)), 0) {
		t.Error("continueCraftsman(level=0) = true, want false (target 0 is unreachable on 2D6)")
	}

	for seed := uint64(1); seed <= 20; seed++ {
		if !continueCraftsman(dice.New(rand.NewPCG(seed, seed)), 6) {
			t.Errorf("seed=%d: continueCraftsman(level=6) = false, want true (target 12 is automatic)", seed)
		}
	}
}

func TestResolveCraftsmanSkillCellNewTradeExcludesHeld(t *testing.T) {
	t.Parallel()

	// column 5, row 2 is "New Trade" (0-indexed).
	held := make([]SkillLevel, 0, len(theTradeChoices)-1)
	for _, name := range theTradeChoices[:len(theTradeChoices)-1] {
		held = append(held, skillLevel1(name, Skill))
	}

	r := dice.New(rand.NewPCG(1, 1))

	skill, ok := resolveCraftsmanSkillCell(r, 5, 2, held)
	if !ok {
		t.Fatal("ok = false, want true (one trade still unheld)")
	}

	if skill.Name != theTradeChoices[len(theTradeChoices)-1] {
		t.Errorf(
			"skill.Name = %q, want the one remaining unheld trade %q",
			skill.Name,
			theTradeChoices[len(theTradeChoices)-1],
		)
	}
}

func TestResolveCraftsmanSkillCellNewTradeUnresolvableWhenAllHeld(t *testing.T) {
	t.Parallel()

	held := make([]SkillLevel, 0, len(theTradeChoices))
	for _, name := range theTradeChoices {
		held = append(held, skillLevel1(name, Skill))
	}

	r := dice.New(rand.NewPCG(1, 1))

	if _, ok := resolveCraftsmanSkillCell(r, 5, 2, held); ok {
		t.Error("ok = true, want false (every trade already held, benefit lost)")
	}
}

// TestRollCraftsmanSkillsMergesRepeatedGrantsMidCareer is the
// regression test for the bug code review caught before this shipped:
// rollCraftsmanSkills used to append a repeated grant as a separate
// raw entry instead of merging it via aggregateSkills, so a skill split
// across two sub-6 grants (e.g. Designer-5 then another Designer-1)
// stayed excluded from craftsmanQualifyingSkills for the rest of the
// career — Craftsman is the first career whose own mid-career logic
// reads heldSkills back before the whole chain's final assembly ever
// aggregates it.
func TestRollCraftsmanSkillsMergesRepeatedGrantsMidCareer(t *testing.T) {
	t.Parallel()

	held := []SkillLevel{
		{Name: "Designer", Level: 5, Kind: Skill}, // one grant below the level-6 threshold
	}

	// Seed 4 confirmed by direct inspection to grant "Designer" again
	// (column 3 "General" has three Designer cells) within 20 rolls,
	// merging to exactly level 6.
	_, newHeld := rollCraftsmanSkills(dice.New(rand.NewPCG(4, 4)), 20, held)

	var designerEntries int

	var designerLevel int

	for _, s := range newHeld {
		if s.Name == "Designer" {
			designerEntries++
			designerLevel = s.Level
		}
	}

	if designerEntries != 1 {
		t.Fatalf("found %d separate \"Designer\" entries in heldSkills, want exactly 1 merged entry", designerEntries)
	}

	if designerLevel < 6 {
		t.Fatalf(
			"merged Designer Level = %d, want >= 6 (fixture assumption broken: no repeat grant happened)",
			designerLevel,
		)
	}

	qualifying := craftsmanQualifyingSkills(newHeld)

	found := false

	for _, s := range qualifying {
		if s.Name == "Designer" {
			found = true
		}
	}

	if !found {
		t.Errorf(
			"craftsmanQualifyingSkills(newHeld) = %+v, want Designer to qualify (merged to level %d)",
			qualifying,
			designerLevel,
		)
	}
}

// The ResolveCraftsmanTerm fixtures below use a fixed, high-skill
// heldSkills list and an all-12 UPP, seed-hunted and verified by direct
// inspection (masterPoints=42, confirmed via craftsmanMasterPoints
// directly, well clear of the 40 minimum) before being pinned here.
var craftsmanHighSkillFixture = []SkillLevel{
	{Name: "Craftsman", Level: 6, Kind: Skill},
	{Name: "Naval Architect", Level: 8, Kind: Skill},
	{Name: "Designer", Level: 8, Kind: Skill},
	{Name: "Bureaucrat", Level: 8, Kind: Skill},
}

var uppCraftsman12 = UPP{Characteristics: [6]ehex.Value{12, 12, 12, 12, 12, 12}}

func TestResolveCraftsmanTermSuccess(t *testing.T) {
	t.Parallel()

	if mp := craftsmanMasterPoints(12, craftsmanHighSkillFixture); mp != 42 {
		t.Fatalf("craftsmanMasterPoints(12, fixture) = %d, want 42 (fixture assumption broken)", mp)
	}

	r := dice.New(rand.NewPCG(1, 1))

	term, _ := ResolveCraftsmanTerm(r, uppCraftsman12, C1, craftsmanHighSkillFixture)

	if term.RiskResult != Unharmed {
		t.Errorf("RiskResult = %v, want Unharmed (Craftsman never rolls Risk & Reward)", term.RiskResult)
	}

	if term.Perfect {
		t.Error("Perfect = true, want false (42 points, below the 55 Perfect threshold)")
	}

	if term.RewardResult != "Masterpiece (Cr170000)" {
		t.Errorf("RewardResult = %q, want %q", term.RewardResult, "Masterpiece (Cr170000)")
	}

	// An upper bound, not an exact count: some rolls can land on an
	// unresolvable cell (Major/Minor/One Science/etc.) and grant
	// nothing — rollSkillsFromTable's own documented "unresolvable =
	// lost" convention, not a bug.
	wantMax := craftsmanSkillsPerTerm + craftsmanSkillBonusPerHit + craftsmanSkillLevel(craftsmanHighSkillFixture)
	if len(term.SkillsAwarded) == 0 || len(term.SkillsAwarded) > wantMax {
		t.Errorf(
			"len(SkillsAwarded) = %d, want 1..%d (4 base + 3 success bonus + Craftsman level 6, some rolls may be lost)",
			len(term.SkillsAwarded),
			wantMax,
		)
	}
}

// TestResolveCraftsmanTermBelowMinimumMasterPointsNeverRolls confirms a
// Masterpiece attempt below 40 Master Points is an automatic Failure —
// no 9D roll happens at all, and the Failure skill rate still applies.
func TestResolveCraftsmanTermBelowMinimumMasterPointsNeverRolls(t *testing.T) {
	t.Parallel()

	lowSkill := []SkillLevel{
		{Name: "Craftsman", Level: 1, Kind: Skill},
		{Name: "Pilot", Level: 6, Kind: Skill},
		{Name: "Admin", Level: 6, Kind: Skill},
	}
	upp := UPP{Characteristics: [6]ehex.Value{2, 2, 2, 2, 2, 2}}

	if mp := craftsmanMasterPoints(2, lowSkill); mp >= craftsmanMinMasterPoints {
		t.Fatalf(
			"craftsmanMasterPoints(2, lowSkill) = %d, want < %d (fixture assumption broken)",
			mp,
			craftsmanMinMasterPoints,
		)
	}

	r := dice.New(rand.NewPCG(1, 1))

	term, _ := ResolveCraftsmanTerm(r, upp, C1, lowSkill)

	if term.RewardResult != "None" {
		t.Errorf(
			"RewardResult = %q, want %q (below-minimum Master Points is an automatic Failure)",
			term.RewardResult,
			"None",
		)
	}

	if term.Perfect {
		t.Error("Perfect = true, want false")
	}

	// An upper bound, not an exact count — see TestResolveCraftsmanTermSuccess's
	// own comment on rollSkillsFromTable's "unresolvable = lost" rule.
	wantMax := craftsmanSkillsPerTerm + craftsmanSkillBonusPerMiss + craftsmanSkillLevel(lowSkill)
	if len(term.SkillsAwarded) == 0 || len(term.SkillsAwarded) > wantMax {
		t.Errorf(
			"len(SkillsAwarded) = %d, want 1..%d (4 base + 1 failure bonus + Craftsman level 1, some rolls may be lost)",
			len(term.SkillsAwarded),
			wantMax,
		)
	}
}

// TestResolveCraftsmanTermPerfectMasterpiece confirms a 55+-point
// success sets Perfect and doubles the recorded Masterpiece Value.
func TestResolveCraftsmanTermPerfectMasterpiece(t *testing.T) {
	t.Parallel()

	perfectSkill := []SkillLevel{
		{Name: "Craftsman", Level: 10, Kind: Skill},
		{Name: "Naval Architect", Level: 10, Kind: Skill},
		{Name: "Designer", Level: 10, Kind: Skill},
		{Name: "Bureaucrat", Level: 10, Kind: Skill},
		{Name: "Liaison", Level: 10, Kind: Skill},
		{Name: "Trader", Level: 10, Kind: Skill},
	}
	upp := UPP{Characteristics: [6]ehex.Value{12, 12, 12, 12, 12, 12}}

	mp := craftsmanMasterPoints(12, perfectSkill)
	if mp < craftsmanPerfectMasterPoints {
		t.Fatalf(
			"craftsmanMasterPoints(12, perfectSkill) = %d, want >= %d (fixture assumption broken)",
			mp,
			craftsmanPerfectMasterPoints,
		)
	}

	r := dice.New(rand.NewPCG(1, 1))

	term, _ := ResolveCraftsmanTerm(r, upp, C1, perfectSkill)

	if !term.Perfect {
		t.Fatalf("Perfect = false, want true (masterPoints=%d)", mp)
	}

	wantValue := craftsmanMasterpieceValue(mp)

	want := fmt.Sprintf("Perfect Masterpiece (Cr%d)", wantValue)
	if term.RewardResult != want {
		t.Errorf("RewardResult = %q, want %q", term.RewardResult, want)
	}
}
