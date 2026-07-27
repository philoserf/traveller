package character

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/philoserf/traveller/dice"
)

func TestCareerChainRegistryCoversExpectedCareers(t *testing.T) {
	t.Parallel()

	want := []string{
		"agent", "citizen", "craftsman", "entertainer", "functionary", "marine", "merchant",
		"noble", "rogue", "scholar", "scout", "soldier", "spacer",
	}

	got := make([]string, 0, len(careerChainRegistry))
	for name := range careerChainRegistry {
		got = append(got, name)
	}

	slices.Sort(got)

	if !slices.Equal(got, want) {
		t.Errorf("careerChainRegistry keys = %v, want %v", got, want)
	}
}

func TestValidateCareerChain(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		careers []string
		wantErr bool
	}{
		{"empty", nil, true},
		{"unknown name", []string{"pirate"}, true},
		{"noble alone", []string{"noble"}, false},
		{"noble terminal", []string{"scout", "noble"}, false},
		{"cannot transfer from noble", []string{"noble", "scout"}, true},
		{"citizen not first", []string{"scout", "citizen"}, true},
		{"citizen first is fine", []string{"citizen", "scout"}, false},
		{"functionary not first", []string{"functionary"}, true},
		{"functionary not first, mid-list", []string{"functionary", "scout"}, true},
		{"functionary later is fine", []string{"scholar", "functionary"}, false},
		{"cannot transfer from functionary", []string{"scholar", "functionary", "scout"}, true},
		{"craftsman not first", []string{"craftsman"}, true},
		{"craftsman not first, mid-list", []string{"craftsman", "scout"}, true},
		{"craftsman later is fine", []string{"citizen", "craftsman"}, false},
		{"adjacent duplicate", []string{"scout", "scout"}, true},
		{"non-adjacent repeat is fine", []string{"scout", "spacer", "scout"}, false},
		{"single valid", []string{"scout"}, false},
		{"valid multi", []string{"scout", "spacer", "merchant"}, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			err := validateCareerChain(c.careers)
			if (err != nil) != c.wantErr {
				t.Errorf("validateCareerChain(%v) error = %v, wantErr %v", c.careers, err, c.wantErr)
			}
		})
	}
}

// TestCareerChainSingleEntryMatchesLegacyGenerator confirms every
// adapter faithfully replicates its own existing buildXCharacter, for
// every seed where the underlying career actually succeeds. Seeds where
// the career never qualifies are deliberately excluded — see
// TestCareerChainFallsBackToCitizenEvenForASingleFailedEntry below for
// why the chain path and the legacy single-career function are NOT
// expected to agree there.
func TestCareerChainSingleEntryMatchesLegacyGenerator(t *testing.T) {
	t.Parallel()

	type legacyCase struct {
		name   string
		legacy func(r *dice.Roller) (Character, bool)
	}

	cases := []legacyCase{
		{"scout", GenerateScoutCharacter},
		{"marine", GenerateMarineCharacter},
		{"soldier", GenerateSoldierCharacter},
		{"spacer", GenerateSpacerCharacter},
		{"agent", GenerateAgentCharacter},
		{"rogue", GenerateRogueCharacter},
		{"scholar", GenerateScholarCharacter},
		{"merchant", GenerateMerchantCharacter},
		{"entertainer", GenerateEntertainerCharacter},
		{"citizen", GenerateCitizenCharacter},
		// Noble joined this battery once buildNobleCharacter started
		// recording Character.Rank, which it had never set: every
		// standalone Noble reported an empty Rank while the identical
		// chain-generated Noble reported the p.88 title its ladder had
		// reached, so the two could not have agreed here before.
		// Craftsman and Functionary are still absent for a different
		// reason — neither has a standalone Generate*Character to compare
		// against at all.
		{"noble", GenerateNobleCharacter},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			for seed := uint64(1); seed <= 50; seed++ {
				want, wantOk := c.legacy(dice.New(rand.NewPCG(seed, seed)))
				if !wantOk {
					continue
				}

				got, gotOk, err := GenerateCareerChainCharacter(dice.New(rand.NewPCG(seed, seed)), []string{c.name}, 0)
				if err != nil {
					t.Fatalf("seed=%d: %v", seed, err)
				}

				if !gotOk || !reflect.DeepEqual(got, want) {
					t.Fatalf("seed=%d:\ngot  %+v\nwant %+v", seed, got, want)
				}
			}
		})
	}
}

// TestCareerChainFallsBackToCitizenEvenForASingleFailedEntry documents a
// deliberate divergence from the legacy single-career functions: Book 1
// p.64's "Begin Citizen Life is Automatic" fallback applies whenever a
// character never manages to begin ANY listed career — including a
// list of exactly one. GenerateScoutCharacter has no such fallback (it
// just reports never-qualified); the chain path is the more RAW-correct
// of the two. cmd/chargen/main.go deliberately keeps routing a single
// -career name through the legacy switch (see this slice's own plan
// Context), so this divergence is invisible there — it only matters to
// callers of GenerateCareerChainCharacter directly.
func TestCareerChainFallsBackToCitizenEvenForASingleFailedEntry(t *testing.T) {
	t.Parallel()

	const seed = 37 // confirmed by direct inspection: Scout fails to Begin at this seed

	_, scoutOk := GenerateScoutCharacter(dice.New(rand.NewPCG(seed, seed)))
	if scoutOk {
		t.Fatalf("seed=%d: expected GenerateScoutCharacter to report never-qualified, got ok=true", seed)
	}

	got, ok, err := GenerateCareerChainCharacter(dice.New(rand.NewPCG(seed, seed)), []string{"scout"}, 0)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Fatal("ok = false, want true (Citizen fallback always succeeds)")
	}

	if len(got.Careers) != 2 || got.Careers[0].Name != "Scout" || len(got.Careers[0].Terms) != 0 {
		t.Fatalf("Careers[0] = %+v, want a zero-term Scout attempt", got.Careers)
	}

	if got.Careers[1].Name != CitizenCareerName || len(got.Careers[1].Terms) == 0 {
		t.Fatalf("Careers[1] = %+v, want a real Citizen career", got.Careers)
	}
}

// TestCareerChainTwoSegmentsAggregateAcrossBoth confirms a genuine
// voluntary transfer (one Scout term, then Spacer Begins and runs) threads
// UPP forward correctly and sums Fame/Cash/WoundBadges/Skills across
// BOTH segments, not just the last — replaying each segment
// independently (against identically-seeded rollers fed the same
// GenerateUPP/GenerateHomeworldSkills prefix the orchestrator itself
// consumes) rather than pinning brittle literal totals.
func TestCareerChainTwoSegmentsAggregateAcrossBoth(t *testing.T) {
	t.Parallel()

	var (
		seed uint64
		got  Character
		ok   bool
		err  error
	)

	for seed = 1; seed <= 500; seed++ {
		got, ok, err = GenerateCareerChainCharacter(dice.New(rand.NewPCG(seed, seed)), []string{"scout", "spacer"}, 0)
		if err == nil && ok && len(got.Careers) == 2 &&
			len(got.Careers[0].Terms) == 1 && len(got.Careers[1].Terms) > 0 {
			break
		}
	}

	if seed > 500 {
		t.Fatal("no seed in [1,500] produced successful Scout-to-Spacer transfer")
	}

	r := dice.New(rand.NewPCG(seed, seed))
	upp := GenerateUPP(r)
	homeworld, homeworldSkills := GenerateHomeworldSkills(r)
	scoutSeg := resolveScoutSegment(r, upp, 1, segmentContext{})
	spacerSeg := resolveSpacerSegment(r, scoutSeg.UPP, maxCareerTerms, segmentContext{})

	wantFame := scoutSeg.Fame + spacerSeg.Fame
	wantCash := scoutSeg.Cash + spacerSeg.Cash
	wantWoundBadges := scoutSeg.WoundBadges + spacerSeg.WoundBadges

	// aggregateSkills (character/skill.go) merges same-name/same-Kind
	// grants across the WHOLE chain, not per segment — a skill repeated
	// between homeworld/Scout/Spacer collapses to one higher-level
	// entry, so the raw concatenated length is only an upper bound, not
	// the expected count.
	rawSkills := make([]SkillLevel, 0, len(homeworldSkills)+len(scoutSeg.Skills)+len(spacerSeg.Skills))

	rawSkills = append(rawSkills, homeworldSkills...)
	rawSkills = append(rawSkills, scoutSeg.Skills...)
	rawSkills = append(rawSkills, spacerSeg.Skills...)
	wantSkills := aggregateSkills(rawSkills)

	if got.Fame != wantFame {
		t.Errorf("Fame = %d, want %d (sum of both segments)", got.Fame, wantFame)
	}

	if got.Cash != wantCash {
		t.Errorf("Cash = %d, want %d (sum of both segments)", got.Cash, wantCash)
	}

	if got.WoundBadges != wantWoundBadges {
		t.Errorf("WoundBadges = %d, want %d (sum of both segments)", got.WoundBadges, wantWoundBadges)
	}

	if !slices.Equal(got.Skills, wantSkills) {
		t.Errorf("Skills = %+v, want %+v (homeworld + both segments, aggregated)", got.Skills, wantSkills)
	}

	if got.Homeworld != homeworld {
		t.Errorf("Homeworld = %q, want %q", got.Homeworld, homeworld)
	}
}

// TestCareerChainFallsBackToCitizenWhenEveryListedCareerFails confirms
// that when every explicitly listed career fails to Begin, Citizen is
// used as the guaranteed fallback (Book 1 p.64).
func TestCareerChainFallsBackToCitizenWhenEveryListedCareerFails(t *testing.T) {
	t.Parallel()

	const seed = 340 // confirmed by direct inspection: both Scout and Spacer fail to Begin

	got, ok, err := GenerateCareerChainCharacter(dice.New(rand.NewPCG(seed, seed)), []string{"scout", "spacer"}, 0)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Fatal("ok = false, want true (Citizen fallback always succeeds)")
	}

	if len(got.Careers) != 3 {
		t.Fatalf("Careers = %+v, want 3 entries (failed Scout, failed Spacer, Citizen)", got.Careers)
	}

	if len(got.Careers[0].Terms) != 0 || len(got.Careers[1].Terms) != 0 {
		t.Fatalf("Careers[0:2] = %+v, want both to be zero-term failed attempts", got.Careers[:2])
	}

	if got.Careers[2].Name != CitizenCareerName || len(got.Careers[2].Terms) == 0 {
		t.Fatalf("Careers[2] = %+v, want a real Citizen career", got.Careers[2])
	}
}

// TestCareerChainTerminalFirstSegmentEndsTheWholeAttempt covers outcomes
// which end Career Resolution before a requested transfer can occur.
func TestCareerChainTerminalFirstSegmentEndsTheWholeAttempt(t *testing.T) {
	t.Parallel()

	for _, result := range []RiskResult{Disabled, Dead} {
		seg := careerSegment{Career: Career{Terms: []Term{{RiskResult: result}}}}
		if !segmentEndsCareerResolution(seg) {
			t.Errorf("segmentEndsCareerResolution(%v) = false, want true", result)
		}
	}
}

// TestCareerChainAgeTargetCutsOffMidCareer confirms a nonzero ageTarget
// stops a still-running career from attempting a term that would push
// past it — seed 3 confirmed by direct inspection to run Scout 4 terms
// (age 34) uncapped; a target of 30 (18+4*3) must produce exactly 3.
func TestCareerChainAgeTargetCutsOffMidCareer(t *testing.T) {
	t.Parallel()

	const seed = 3

	uncapped, ok, err := GenerateCareerChainCharacter(dice.New(rand.NewPCG(seed, seed)), []string{"scout"}, 0)
	if err != nil {
		t.Fatal(err)
	}

	if !ok || len(uncapped.Careers) != 1 || len(uncapped.Careers[0].Terms) != 7 {
		t.Fatalf("uncapped = %+v, want a single 7-term Scout career (fixture assumption broken)", uncapped.Careers)
	}

	capped, ok, err := GenerateCareerChainCharacter(dice.New(rand.NewPCG(seed, seed)), []string{"scout"}, 30)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Fatal("ok = false, want true")
	}

	if len(capped.Careers) != 1 || len(capped.Careers[0].Terms) != 3 {
		t.Fatalf("capped.Careers = %+v, want a single 3-term Scout career", capped.Careers)
	}

	if capped.Age != 30 {
		t.Errorf("Age = %d, want 30", capped.Age)
	}

	if !reflect.DeepEqual(capped.Careers[0].Terms, uncapped.Careers[0].Terms[:3]) {
		t.Fatalf("capped terms diverge from the first 3 terms of the uncapped run")
	}
}

// TestCareerChainAgeTargetStopsBeforeFurtherProgression confirms that
// two available term slots are shared across an explicit transfer.
// The ordered list elects to leave Scout after one term, so the second
// four-year slot belongs to Spacer rather than a second Scout term.
func TestCareerChainAgeTargetStopsBeforeFurtherProgression(t *testing.T) {
	t.Parallel()

	const seed = 1

	got, ok, err := GenerateCareerChainCharacter(dice.New(rand.NewPCG(seed, seed)), []string{"scout", "spacer"}, 26)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Fatal("ok = false, want true")
	}

	if len(got.Careers) != 2 || len(got.Careers[0].Terms) != 1 || len(got.Careers[1].Terms) != 1 {
		t.Fatalf("Careers = %+v, want one Scout term followed by one Spacer term", got.Careers)
	}

	if got.Age != 26 {
		t.Errorf("Age = %d, want 26", got.Age)
	}
}

// TestCareerChainAgeTargetAtOrBelow18ProducesNoCareers confirms a
// target that allows zero terms degrades cleanly: ok=true, no Careers
// at all (not even a zero-term "attempted" entry, and no Citizen
// fallback — the fallback itself is gated by the same budget), only
// homeworld skills.
func TestCareerChainAgeTargetAtOrBelow18ProducesNoCareers(t *testing.T) {
	t.Parallel()

	got, ok, err := GenerateCareerChainCharacter(dice.New(rand.NewPCG(1, 1)), []string{"scout"}, 18)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Fatal("ok = false, want true")
	}

	if len(got.Careers) != 0 {
		t.Fatalf("Careers = %+v, want none", got.Careers)
	}

	if got.Age != 18 {
		t.Errorf("Age = %d, want 18", got.Age)
	}

	r := dice.New(rand.NewPCG(1, 1))
	GenerateUPP(r) // replay the same prefix GenerateCareerChainCharacter itself consumes
	_, homeworldSkills := GenerateHomeworldSkills(r)

	if want := aggregateSkills(homeworldSkills); !slices.Equal(got.Skills, want) {
		t.Errorf("Skills = %+v, want %+v (homeworld skills only, aggregated)", got.Skills, want)
	}
}

// TestCareerChainAgeTargetSoLargeItNeverBindsMatchesUnbounded confirms
// ageTarget==0 truly is a no-op relative to a target so large it can
// never be reached — an -age value a caller might pick "just to be
// safe" shouldn't change anything.
func TestCareerChainAgeTargetSoLargeItNeverBindsMatchesUnbounded(t *testing.T) {
	t.Parallel()

	const seed = 15 // confirmed by direct inspection: Scout runs the full 14-term cap

	unbounded, ok1, err := GenerateCareerChainCharacter(dice.New(rand.NewPCG(seed, seed)), []string{"scout"}, 0)
	if err != nil {
		t.Fatal(err)
	}

	generous, ok2, err := GenerateCareerChainCharacter(dice.New(rand.NewPCG(seed, seed)), []string{"scout"}, 1000)
	if err != nil {
		t.Fatal(err)
	}

	if ok1 != ok2 || !reflect.DeepEqual(unbounded, generous) {
		t.Fatalf("ageTarget=1000 diverged from ageTarget=0:\nunbounded=%+v ok=%v\ngenerous=%+v ok=%v",
			unbounded, ok1, generous, ok2)
	}
}

// TestCareerChainNobleAgeTargetCutsOffMidCareer is #49's regression: a
// standalone Noble career must honor an -age budget exactly like every
// other career, even though cmd/chargen used to reject -age for Noble
// outright. Seed 43 (found by direct search) runs Noble to 4 terms
// (age 34) unbounded; capping at age 30 (18+4*3, an exact term
// boundary) must truncate to the first 3 of those same terms, not
// generate a divergent career.
func TestCareerChainNobleAgeTargetCutsOffMidCareer(t *testing.T) {
	t.Parallel()

	const seed = 43

	uncapped, ok, err := GenerateCareerChainCharacter(dice.New(rand.NewPCG(seed, seed)), []string{"noble"}, 0)
	if err != nil {
		t.Fatal(err)
	}

	if !ok || len(uncapped.Careers) != 1 || len(uncapped.Careers[0].Terms) != 4 {
		t.Fatalf("uncapped = %+v, want a single 4-term Noble career (fixture assumption broken)", uncapped.Careers)
	}

	capped, ok, err := GenerateCareerChainCharacter(dice.New(rand.NewPCG(seed, seed)), []string{"noble"}, 30)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Fatal("ok = false, want true")
	}

	if len(capped.Careers) != 1 || len(capped.Careers[0].Terms) != 3 {
		t.Fatalf("capped.Careers = %+v, want a single 3-term Noble career", capped.Careers)
	}

	if capped.Age != 30 {
		t.Errorf("Age = %d, want 30", capped.Age)
	}

	if !reflect.DeepEqual(capped.Careers[0].Terms, uncapped.Careers[0].Terms[:3]) {
		t.Fatalf("capped terms diverge from the first 3 terms of the uncapped run")
	}
}

// TestCareerChainNobleAgeTargetAtOrBelow18ProducesNoCareers mirrors
// TestCareerChainAgeTargetAtOrBelow18ProducesNoCareers for Noble
// specifically: a target that allows zero terms degrades cleanly
// regardless of Noble's own Begin odds for this seed.
func TestCareerChainNobleAgeTargetAtOrBelow18ProducesNoCareers(t *testing.T) {
	t.Parallel()

	got, ok, err := GenerateCareerChainCharacter(dice.New(rand.NewPCG(1, 1)), []string{"noble"}, 18)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Fatal("ok = false, want true")
	}

	if len(got.Careers) != 0 {
		t.Fatalf("Careers = %+v, want none", got.Careers)
	}

	if got.Age != 18 {
		t.Errorf("Age = %d, want 18", got.Age)
	}
}

// TestCareerChainNobleAgeTargetSoLargeItNeverBindsMatchesUnbounded
// mirrors TestCareerChainAgeTargetSoLargeItNeverBindsMatchesUnbounded
// for Noble: an -age value with enough headroom to never actually bind
// must produce identical results to no target at all, the same
// unbounded-parity property every other career already has.
func TestCareerChainNobleAgeTargetSoLargeItNeverBindsMatchesUnbounded(t *testing.T) {
	t.Parallel()

	const seed = 43 // confirmed above: a real, multi-term, successfully-begun Noble career

	unbounded, ok1, err := GenerateCareerChainCharacter(dice.New(rand.NewPCG(seed, seed)), []string{"noble"}, 0)
	if err != nil {
		t.Fatal(err)
	}

	generous, ok2, err := GenerateCareerChainCharacter(dice.New(rand.NewPCG(seed, seed)), []string{"noble"}, 1000)
	if err != nil {
		t.Fatal(err)
	}

	if ok1 != ok2 || !reflect.DeepEqual(unbounded, generous) {
		t.Fatalf("ageTarget=1000 diverged from ageTarget=0:\nunbounded=%+v ok=%v\ngenerous=%+v ok=%v",
			unbounded, ok1, generous, ok2)
	}
}

// TestCareerChainNobleFailedBeginStillHonorsAgeBudget confirms a failed
// Noble Begin (Soc < B) combines correctly with an -age target: the
// zero-term Noble entry still counts as "nothing succeeded" for the
// Citizen-fallback rule, and Citizen itself still respects the same
// budget rather than running unbounded once substituted in. Seed 1
// (found by direct search) fails Noble's own Begin (Soc < 11) and falls
// back to Citizen, which then runs to fill the remaining age-42 budget.
func TestCareerChainNobleFailedBeginStillHonorsAgeBudget(t *testing.T) {
	t.Parallel()

	got, ok, err := GenerateCareerChainCharacter(dice.New(rand.NewPCG(1, 1)), []string{"noble"}, 42)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Fatal("ok = false, want true")
	}

	if len(got.Careers) != 2 || got.Careers[0].Name != NobleCareerName || len(got.Careers[0].Terms) != 0 {
		t.Fatalf("Careers[0] = %+v, want a zero-term Noble entry (fixture assumption broke: Begin succeeded?)",
			got.Careers)
	}

	if got.Careers[1].Name != CitizenCareerName {
		t.Fatalf("Careers[1].Name = %q, want %q (Citizen fallback)", got.Careers[1].Name, CitizenCareerName)
	}

	if got.Age != 42 {
		t.Errorf("Age = %d, want 42 (Citizen fallback still respects the same budget)", got.Age)
	}
}

// TestChainRankPersistsAfterALaterRanklessCareer confirms Book 1 p.66's
// Reserves rule: a character's displayed Rank reflects the last
// Armed-Forces rank ever held, even after a later career with no rank
// concept at all (Scout never sets Term.Rank).
func TestChainRankPersistsAfterALaterRanklessCareer(t *testing.T) {
	t.Parallel()

	careers := []Career{
		{Name: MarineCareerName, Terms: []Term{{Rank: "M3 Sergeant"}}},
		{Name: "Scout", Terms: []Term{{}, {}}},
	}

	if got := chainRank(careers); got != "M3 Sergeant" {
		t.Errorf("chainRank(...) = %q, want %q", got, "M3 Sergeant")
	}
}

func TestChainRankEmptyWhenNoSegmentEverHeldARank(t *testing.T) {
	t.Parallel()

	careers := []Career{
		{Name: "Scout", Terms: []Term{{}}},
		{Name: RogueCareerName, Terms: []Term{{}}},
	}

	if got := chainRank(careers); got != "" {
		t.Errorf("chainRank(...) = %q, want empty", got)
	}
}

// A Functionary destination is allowed after a single prior term and,
// once entered, is necessarily the final career.
func TestCareerChainTransfersToTerminalFunctionaryAfterOneTerm(t *testing.T) {
	t.Parallel()

	for seed := uint64(1); seed <= 1000; seed++ {
		got, ok, err := GenerateCareerChainCharacter(
			dice.New(rand.NewPCG(seed, seed)), []string{"scholar", "functionary"}, 0)
		if err == nil && ok && len(got.Careers) == 2 && len(got.Careers[1].Terms) > 0 {
			if len(got.Careers[0].Terms) != 1 {
				t.Fatalf("Scholar terms = %d, want 1 before voluntary transfer", len(got.Careers[0].Terms))
			}

			return
		}
	}

	t.Fatal("no seed in [1,1000] produced a successful Functionary transfer")
}

// TestCareerChainFunctionaryBeginCanFailLikeAnyOtherCareer confirms
// Functionary's own Begin failure (too few prior terms) is handled the
// same way as any other listed career's failed Begin — a zero-term
// Career entry, chain otherwise unaffected.
func TestCareerChainFunctionaryBeginCanFailLikeAnyOtherCareer(t *testing.T) {
	t.Parallel()

	const seed = 2 // confirmed by direct inspection: Scholar serves 1 term, Functionary's Begin then fails (target 3)

	got, ok, err := GenerateCareerChainCharacter(
		dice.New(rand.NewPCG(seed, seed)),
		[]string{"scholar", "functionary"},
		0,
	)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Fatal("ok = false, want true")
	}

	if len(got.Careers) != 2 || len(got.Careers[0].Terms) == 0 {
		t.Fatalf("Careers = %+v, want a real Scholar career", got.Careers)
	}

	if got.Careers[1].Name != FunctionaryCareerName || len(got.Careers[1].Terms) != 0 {
		t.Fatalf("Careers[1] = %+v, want a zero-term failed Functionary attempt", got.Careers[1])
	}
}

// TestCareerChainMergesASkillRepeatedAcrossCareers confirms
// aggregateSkills (character/skill.go) applies across the WHOLE chain,
// not just within one career — a skill granted more than once between
// Scholar and Functionary shows up as one merged, higher-level entry,
// not several separate level-1 lines.
func TestCareerChainMergesASkillRepeatedAcrossCareers(t *testing.T) {
	t.Parallel()

	const seed = 7 // confirmed by direct inspection: Bureaucrat is granted 4 times total across this chain

	got, ok, err := GenerateCareerChainCharacter(
		dice.New(rand.NewPCG(seed, seed)),
		[]string{"scholar", "functionary"},
		0,
	)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Fatal("ok = false, want true")
	}

	var bureaucratEntries int

	var mergedLevel int

	for _, s := range got.Skills {
		if s.Name == "Bureaucrat" {
			bureaucratEntries++
			mergedLevel = s.Level
		}
	}

	if bureaucratEntries != 1 {
		t.Fatalf("found %d separate \"Bureaucrat\" entries in Skills, want exactly 1 merged entry", bureaucratEntries)
	}

	if mergedLevel < 1 {
		t.Errorf("merged Bureaucrat Level = %d, want at least 1", mergedLevel)
	}
}

// TestCareerChainCraftsmanBeginFailsWithoutPriorSkills confirms a
// normal chain (no deliberately-inflated skills) essentially never
// meets Craftsman's own prerequisite yet — seed 1 confirmed by direct
// inspection: a 12-term Citizen life still doesn't produce two level-6+
// skills plus Craftsman-1, so the listed Craftsman entry fails to Begin
// like any other career's own failed attempt.
func TestCareerChainCraftsmanBeginFailsWithoutPriorSkills(t *testing.T) {
	t.Parallel()

	const seed = 1

	got, ok, err := GenerateCareerChainCharacter(dice.New(rand.NewPCG(seed, seed)), []string{"citizen", "craftsman"}, 0)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Fatal("ok = false, want true")
	}

	if len(got.Careers) != 2 || got.Careers[0].Name != CitizenCareerName || len(got.Careers[0].Terms) == 0 {
		t.Fatalf("Careers[0] = %+v, want a real Citizen career", got.Careers)
	}

	if got.Careers[1].Name != CraftsmanCareerName || len(got.Careers[1].Terms) != 0 {
		t.Fatalf("Careers[1] = %+v, want a zero-term failed Craftsman attempt", got.Careers[1])
	}
}

// TestCareerChainCraftsmanSegmentProducesEquipmentAndFame exercises
// resolveCraftsmanSegment directly with a crafted ctx guaranteed to
// meet Craftsman's own prerequisite — real chargen essentially never
// reaches these skill levels by chance within a short seed search, but
// the segment adapter's own Equipment/Fame wiring is exactly what this
// test needs to check, independent of how the skills were acquired.
func TestCareerChainCraftsmanSegmentProducesEquipmentAndFame(t *testing.T) {
	t.Parallel()

	ctx := segmentContext{SkillsSoFar: craftsmanHighSkillFixture}

	seg := resolveCraftsmanSegment(dice.New(rand.NewPCG(1, 1)), uppCraftsman12, maxCareerTerms, ctx)

	if !seg.Survived {
		t.Error("Survived = false, want true (Craftsman has no death mechanic)")
	}

	if seg.WoundBadges != 0 {
		t.Errorf("WoundBadges = %d, want 0", seg.WoundBadges)
	}

	if len(seg.Equipment) == 0 {
		t.Fatal("Equipment is empty, want at least one Masterpiece")
	}

	for _, item := range seg.Equipment {
		if !strings.Contains(item, "Masterpiece") {
			t.Errorf("Equipment entry %q doesn't mention Masterpiece", item)
		}
	}

	wantFame := craftsmanCareerFame(seg.Career.Terms)
	if seg.Fame < wantFame {
		t.Errorf(
			"Fame = %d, want at least %d (craftsmanCareerFame alone, before any Mustering Out bonus)",
			seg.Fame,
			wantFame,
		)
	}
}

// TestCareerChainAgingDeathIsNotASuccessfulAttempt is #60's regression
// through a public generation path, and PR #69's own review finding
// alongside it: a character killed by Aging (Book 1 p.89) must report
// ok=false rather than only recording the death in Notes, and their
// career history must not outlive them. Because Aging now runs between
// terms, someone who dies at a checkpoint simply serves no further
// terms — so Age, the death note, and the term count all agree by
// construction. Seeds found by direct search; scout 4972 dies before
// the term cap, so it pins a genuine mid-career stop rather than one
// that merely coincides with running out of terms.
func TestCareerChainAgingDeathIsNotASuccessfulAttempt(t *testing.T) {
	t.Parallel()

	cases := []struct {
		career string
		seed   uint64
	}{
		{"scout", 4972},
		{"citizen", 264},
		{"scholar", 4953},
	}

	for _, c := range cases {
		t.Run(c.career, func(t *testing.T) {
			t.Parallel()

			got, ok, err := GenerateCareerChainCharacter(
				dice.New(rand.NewPCG(c.seed, c.seed)), []string{c.career}, 0)
			if err != nil {
				t.Fatal(err)
			}

			if !strings.Contains(got.Notes, "died of natural causes") {
				t.Fatalf("Notes = %q, want an Aging death (fixture assumption broke)", got.Notes)
			}

			if ok {
				t.Error("ok = true, want false (an Aging death is not a surviving character)")
			}

			if want := fmt.Sprintf("Age %d: died", got.Age); !strings.Contains(got.Notes, want) {
				t.Errorf("Age = %d, but Notes = %q — Age must match the fatal checkpoint", got.Age, got.Notes)
			}

			termsServed := 0
			for _, career := range got.Careers {
				termsServed += len(career.Terms)
			}

			if want := AgeFromTermsServed(termsServed); got.Age != want {
				t.Errorf("Age = %d but %d terms served implies age %d — the sheet must not record "+
					"service beyond the age its own death note gives", got.Age, termsServed, want)
			}

			if got.LifeStage != LifeStageForAge(got.Age) {
				t.Errorf("LifeStage = %d, want %d (must follow Age)", got.LifeStage, LifeStageForAge(got.Age))
			}
		})
	}
}

// TestCareerChainAgingDeathGrantsNoMusterOut is the regression for PR
// #69's third review round: resolveRiskCareerSegment — the shared body
// behind Scout, Marine, Soldier, Spacer and Agent as chain segments —
// called its resolveMusterOut callback unconditionally, so those five
// still collected benefits through the chain path after Aging had
// killed them, even once every other site was guarded.
//
// This exercises the case a dead-on-entry simulation cannot: Aging kills
// *during* the segment, so the career has real terms behind it and
// Mustering Out would genuinely roll against them. A simulation already
// dead when the loop starts returns zero terms, which makes the roll
// count zero anyway and would pass with the guard still missing.
//
// Seed 4972 is deterministic and shared with
// TestBuildScoutCharacterAgingDeathKeepsCareerFame: an -age target high
// enough never to bind routes the same character through the chain path
// instead of the single-career one, so the two must agree.
func TestCareerChainAgingDeathGrantsNoMusterOut(t *testing.T) {
	t.Parallel()

	got, ok, err := GenerateCareerChainCharacter(
		dice.New(rand.NewPCG(4972, 4972)), []string{"scout"}, 1000)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got.Notes, "died of natural causes") {
		t.Fatalf("Notes = %q, want an Aging death (fixture assumption broke)", got.Notes)
	}

	if ok {
		t.Error("ok = true, want false")
	}

	career := got.Careers[0]
	if len(career.Terms) == 0 {
		t.Fatal("career has no terms — this fixture must die partway through a real career, " +
			"not before it starts, or it cannot exercise the guard at all")
	}

	mo := career.MusteringOut
	if n := len(mo.Benefits) + len(mo.Money) + len(mo.Automatics) + len(mo.Entitlements); n != 0 {
		t.Errorf("Mustering Out entries = %d, want 0 (a dead character never reaches Mustering Out): %+v", n, mo)
	}

	if got.Cash != 0 {
		t.Errorf("Cash = %d, want 0", got.Cash)
	}
}

// TestCareerChainStopsWhenAFailedBeginIsFatal covers PR #75's review
// findings at the chain level. The chain used to treat a zero-term
// segment as "nothing happened" and continue, without asking whether
// the year that failed Begin cost had just killed the character; the
// Citizen fallback had the same gap.
//
// The living path is asserted end to end, since guarding the fallback on
// aging.alive() is exactly the change that could have broken it. The
// fatal path is asserted at segment level, because GenerateCareerChainCharacter
// rolls its own UPP and simulation and gives no way to start one dead.
func TestCareerChainStopsWhenAFailedBeginIsFatal(t *testing.T) {
	t.Parallel()

	t.Run("a living character still reaches the Citizen fallback", func(t *testing.T) {
		t.Parallel()

		// Noble needs Soc B+, so most seeds fail to Begin it — and Noble's
		// Begin is a prerequisite, charging no year, which keeps the
		// character comfortably alive for the fallback to matter.
		for seed := range uint64(50) {
			got, ok, err := GenerateCareerChainCharacter(
				dice.New(rand.NewPCG(seed+1, seed+1)), []string{"noble"}, 0)
			if err != nil {
				t.Fatal(err)
			}

			if len(got.Careers) != 2 || got.Careers[0].Name != NobleCareerName {
				continue // this seed qualified for Noble; try another
			}

			if !ok {
				t.Fatalf("seed %d: ok = false for a living character, Notes = %q", seed+1, got.Notes)
			}

			if got.Careers[1].Name != CitizenCareerName {
				t.Fatalf("seed %d: Careers = %v, want the Citizen fallback after a failed Noble Begin",
					seed+1, careerNamesOf(got))
			}

			return
		}

		t.Fatal("no seed in 1..50 failed Noble's Begin — the fallback path went unexercised")
	})

	t.Run("a dead character begins no career and gets no fallback", func(t *testing.T) {
		t.Parallel()

		dead := func() *agingSimulation {
			return &agingSimulation{
				termsServed: 5,
				diedAtAge:   38,
				notes:       []string{"Age 38: died of natural causes (x)"},
			}
		}

		for name, resolve := range map[string]careerSegmentResolver{
			"marine":  resolveMarineSegment,
			"citizen": resolveCitizenSegment, // the fallback career itself
		} {
			seg := resolve(dice.New(rand.NewPCG(1, 1)), UPP{}, maxCareerTerms, segmentContext{Aging: dead()})
			if len(seg.Career.Terms) != 0 {
				t.Errorf("%s: a dead character served %d terms, want 0", name, len(seg.Career.Terms))
			}
		}
	})
}

func careerNamesOf(c Character) []string {
	names := make([]string, 0, len(c.Careers))
	for _, career := range c.Careers {
		names = append(names, career.Name)
	}

	return names
}

// TestFunctionaryPensionReplacesCitizenPension pins Book 1 p.70's one
// interaction between Entitlements: "a Functionary receives Cr15,000 per
// year (which replaces a Citizen's pension, if any)."
//
// Everything else stacks — p.70 says so outright ("A character may
// receive duplicate Entitlements... both Military and Professor's
// retirement pay") — so the test also pins what must survive alongside.
func TestFunctionaryPensionReplacesCitizenPension(t *testing.T) {
	t.Parallel()

	careers := []Career{
		{Name: CitizenCareerName, MusteringOut: MusteringOut{Pension: citizenPensionRate}},
		{Name: ScholarCareerName, MusteringOut: MusteringOut{Pension: professorPensionRate}},
		{Name: MarineCareerName, MusteringOut: MusteringOut{RetirementPay: 8000}},
		{Name: FunctionaryCareerName, MusteringOut: MusteringOut{Pension: functionaryPensionRate}},
	}

	resolveCitizenPensionReplacement(careers)

	if got := careers[0].MusteringOut.Pension; got != 0 {
		t.Errorf("Citizen pension = %d, want 0 (replaced by the Functionary's)", got)
	}

	if got := careers[1].MusteringOut.Pension; got != professorPensionRate {
		t.Errorf("Professor pension = %d, want %d — only the Citizen's is replaced", got, professorPensionRate)
	}

	if got := careers[2].MusteringOut.RetirementPay; got != 8000 {
		t.Errorf("Retirement Pay = %d, want 8000 — Entitlements otherwise stack", got)
	}

	if got := careers[3].MusteringOut.Pension; got != functionaryPensionRate {
		t.Errorf("Functionary pension = %d, want %d", got, functionaryPensionRate)
	}
}

// TestCitizenPensionSurvivesWithoutAFunctionary is the converse: nothing
// to replace it, so it stands.
func TestCitizenPensionSurvivesWithoutAFunctionary(t *testing.T) {
	t.Parallel()

	careers := []Career{{Name: CitizenCareerName, MusteringOut: MusteringOut{Pension: citizenPensionRate}}}

	resolveCitizenPensionReplacement(careers)

	if got := careers[0].MusteringOut.Pension; got != citizenPensionRate {
		t.Errorf("Citizen pension = %d, want %d", got, citizenPensionRate)
	}
}
