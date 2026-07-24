package character

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/world"
)

// TestHomeworldSkillByTradeCodeMapping pins every one of
// homeworldSkillByTradeCode's 19 entries against Book 1 p.56's
// "Homeworld and Birthworld Skills" table, transcribed directly from the
// page image — not just a sample, so a transcription slip on any one
// entry (e.g. "JOT" instead of the actual skill name "Jack of all
// Trades," an earlier version of this table's own mistake) fails here
// instead of only showing up on generated character sheets.
func TestHomeworldSkillByTradeCodeMapping(t *testing.T) {
	t.Parallel()

	cases := map[world.TradeCode]string{
		world.Agricultural:    "Animals",
		world.AsteroidBelt:    "Zero-G",
		world.Dangerous:       "Fighter",
		world.Desert:          "Survival",
		world.Fluid:           "Hostile Environment",
		world.Garden:          "Trader",
		world.Hellworld:       "Hostile Environment",
		world.HighPopulation:  "Streetwise",
		world.IceCapped:       "Vacc Suit",
		world.LowPopulation:   "Flyer",
		world.NonAgricultural: "Survey",
		world.NonIndustrial:   "Driver",
		world.Ocean:           "Hi-G",
		world.PreAgricultural: "Trader",
		world.PreIndustrial:   "Jack of all Trades",
		world.Poor:            "Steward",
		world.PreRich:         "Craftsman",
		world.Vacuum:          "Vacc Suit",
		world.WaterWorld:      "Seafarer",
	}

	if got, want := len(homeworldSkillByTradeCode), len(cases); got != want {
		t.Fatalf("homeworldSkillByTradeCode has %d entries, this test pins %d — keep them in sync", got, want)
	}

	for tc, want := range cases {
		if got := homeworldSkillByTradeCode[tc]; got != want {
			t.Errorf("homeworldSkillByTradeCode[%s] = %q, want %q", tc, got, want)
		}
	}
}

// TestHomeworldSkillByTradeCodeOmitsNoSkillCodes confirms the "(no
// skill)" codes (Book 1 p.56) are genuinely absent from the map, not
// mapped to an empty string.
func TestHomeworldSkillByTradeCodeOmitsNoSkillCodes(t *testing.T) {
	t.Parallel()

	for _, tc := range []world.TradeCode{
		world.DataRepository, world.AncientSite, world.Barren, world.Dieback,
		world.Forbidden, world.Locked, world.MilitaryRule, world.PreHigh,
		world.PrisonExileCamp, world.Puzzle, world.Reserve,
		world.Satellite, world.PenalColony, world.Colony,
	} {
		if _, ok := homeworldSkillByTradeCode[tc]; ok {
			t.Errorf("homeworldSkillByTradeCode[%s] present, want absent (no-skill code)", tc)
		}
	}
}

// TestHomeworldSkillByTradeCodeOmitsUnreachableCodes confirms the codes
// world.Generate can never actually produce (orbit-relative, sector-
// context, referee-assigned, or non-mainworld-only — see
// homeworldSkillByTradeCode's own doc comment) are absent from the map,
// distinct from TestHomeworldSkillByTradeCodeOmitsNoSkillCodes' genuine
// "(no skill)" codes: these DO have a Book 1 skill, it's just
// unreachable given this generator's homeworld source.
func TestHomeworldSkillByTradeCodeOmitsUnreachableCodes(t *testing.T) {
	t.Parallel()

	for _, tc := range []world.TradeCode{
		world.Cold, world.Farming, world.Frozen, world.Hot, world.Tropic, world.Tundra, world.TwilightZone,
		world.SubsectorCapital, world.SectorCapital, world.Capital, world.Mining,
	} {
		if _, ok := homeworldSkillByTradeCode[tc]; ok {
			t.Errorf("homeworldSkillByTradeCode[%s] present, want absent (unreachable via world.Generate)", tc)
		}
	}
}

func TestHomeworldSkillForTradeCodeRiResolvesToOneArt(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 2))

	for range 100 {
		skill, ok := homeworldSkillForTradeCode(r, world.Rich)
		if !ok {
			t.Fatal("homeworldSkillForTradeCode(Ri) reported ok=false, want true")
		}

		if !slices.Contains(oneArtChoices, skill.Name) {
			t.Errorf("homeworldSkillForTradeCode(Ri) = %q, want one of %v", skill.Name, oneArtChoices)
		}
	}
}

func TestHomeworldSkillForTradeCodeInResolvesToOneTrade(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(3, 4))

	for range 100 {
		skill, ok := homeworldSkillForTradeCode(r, world.Industrial)
		if !ok {
			t.Fatal("homeworldSkillForTradeCode(In) reported ok=false, want true")
		}

		if !slices.Contains(theTradeChoices, skill.Name) {
			t.Errorf("homeworldSkillForTradeCode(In) = %q, want one of %v", skill.Name, theTradeChoices)
		}
	}
}

func TestGenerateHomeworldSkillsDeterminism(t *testing.T) {
	t.Parallel()

	r1 := dice.New(rand.NewPCG(9, 9))
	r2 := dice.New(rand.NewPCG(9, 9))

	hw1, skills1 := GenerateHomeworldSkills(r1)
	hw2, skills2 := GenerateHomeworldSkills(r2)

	if hw1 != hw2 {
		t.Fatalf("identical seeds produced different homeworlds: %s vs %s", hw1, hw2)
	}

	if len(skills1) != len(skills2) {
		t.Fatalf("identical seeds produced different skill counts: %v vs %v", skills1, skills2)
	}

	for i := range skills1 {
		if skills1[i] != skills2[i] {
			t.Fatalf("identical seeds produced different skills at %d: %+v vs %+v", i, skills1[i], skills2[i])
		}
	}
}

// TestRollDeepSpaceBonusRate confirms the Zero-G/Vacc Suit bonus (Book 1
// p.58: "roll 2 on 2D") fires at ~1/36 (2.8%). Exercises
// rollDeepSpaceBonus directly rather than GenerateHomeworldSkills: an
// Asteroid Belt (As) or Vacuum (Va) homeworld already grants Zero-G/Vacc
// Suit via the normal Trade Code path, so detecting the bonus by
// scanning GenerateHomeworldSkills' combined output for those skill
// names would over-count coincidental matches.
func TestRollDeepSpaceBonusRate(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(13, 14))

	const trials = 20000

	fired := 0

	for range trials {
		if rollDeepSpaceBonus(r) != nil {
			fired++
		}
	}

	gotPct := 100 * float64(fired) / trials
	if wantPct := 100.0 / 36; gotPct < wantPct-1 || gotPct > wantPct+1 {
		t.Errorf("rollDeepSpaceBonus fired %.2f%% of %d trials, want ~%.2f%% (2D6=2)", gotPct, trials, wantPct)
	}
}

// TestRollDeepSpaceBonusGrantsSkills confirms a firing roll grants
// exactly Zero-G-1 and Vacc Suit-1.
func TestRollDeepSpaceBonusGrantsSkills(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(13, 14))

	for range 20000 {
		skills := rollDeepSpaceBonus(r)
		if skills == nil {
			continue
		}

		want := []SkillLevel{{Name: "Zero-G", Level: 1, Kind: Skill}, {Name: "Vacc Suit", Level: 1, Kind: Skill}}
		if !slices.Equal(skills, want) {
			t.Fatalf("rollDeepSpaceBonus() = %+v, want %+v", skills, want)
		}

		return
	}

	t.Fatal("rollDeepSpaceBonus never fired in 20000 trials — test can't verify anything")
}

// TestGenerateHomeworldSkillsOnlyGrantsReachableSkills runs many real
// generated homeworlds through GenerateHomeworldSkills and confirms
// every returned skill name is one this generator can actually produce
// — homeworldSkillByTradeCode's values, plus the two Choose-One lists,
// plus the deep-space-birth bonus names. This is the integration-level
// check TestHomeworldSkillByTradeCodeMapping and
// TestHomeworldSkillByTradeCodeOmitsUnreachableCodes can't provide on
// their own: it exercises GenerateHomeworldSkills against real
// world.Generate output rather than only probing the map directly, so a
// future change that reintroduces an unreachable code (or a typo in a
// skill name) fails here even if the map-level tests still pass.
func TestGenerateHomeworldSkillsOnlyGrantsReachableSkills(t *testing.T) {
	t.Parallel()

	reachable := map[string]bool{"Zero-G": true, "Vacc Suit": true}
	for _, name := range homeworldSkillByTradeCode {
		reachable[name] = true
	}

	for _, name := range oneArtChoices {
		reachable[name] = true
	}

	for _, name := range theTradeChoices {
		reachable[name] = true
	}

	r := dice.New(rand.NewPCG(21, 22))

	for range 2000 {
		_, skills := GenerateHomeworldSkills(r)

		for _, s := range skills {
			if !reachable[s.Name] {
				t.Fatalf("GenerateHomeworldSkills granted unexpected skill %q", s.Name)
			}
		}
	}
}
