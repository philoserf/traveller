package character

import (
	"math/rand/v2"
	"reflect"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
)

func TestCareerChainRegistryCoversExpectedCareers(t *testing.T) {
	t.Parallel()

	want := []string{
		"agent", "citizen", "entertainer", "marine", "merchant",
		"rogue", "scholar", "scout", "soldier", "spacer",
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
		{"noble rejected", []string{"noble"}, true},
		{"noble rejected mid-list", []string{"scout", "noble"}, true},
		{"citizen not first", []string{"scout", "citizen"}, true},
		{"citizen first is fine", []string{"citizen", "scout"}, false},
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
		{"citizen", func(r *dice.Roller) (Character, bool) { return GenerateCitizenCharacter(r), true }},
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
// two-career chain (Scout ends, Spacer then Begins and runs) threads
// UPP forward correctly and sums Fame/Cash/WoundBadges/Skills across
// BOTH segments, not just the last — replaying each segment
// independently (against identically-seeded rollers fed the same
// GenerateUPP/GenerateHomeworldSkills prefix the orchestrator itself
// consumes) rather than pinning brittle literal totals.
func TestCareerChainTwoSegmentsAggregateAcrossBoth(t *testing.T) {
	t.Parallel()

	const seed = 7 // confirmed by direct inspection: Scout ends after 1 term, Spacer Begins and runs 1 term

	got, ok, err := GenerateCareerChainCharacter(dice.New(rand.NewPCG(seed, seed)), []string{"scout", "spacer"}, 0)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Fatal("ok = false, want true")
	}

	if len(got.Careers) != 2 || got.Careers[0].Name != "Scout" || got.Careers[1].Name != SpacerCareerName {
		t.Fatalf("Careers = %+v, want [Scout, Spacer]", got.Careers)
	}

	if len(got.Careers[0].Terms) == 0 || len(got.Careers[1].Terms) == 0 {
		t.Fatalf("Careers = %+v, want both segments to have served at least one term", got.Careers)
	}

	r := dice.New(rand.NewPCG(seed, seed))
	upp := GenerateUPP(r)
	homeworld, homeworldSkills := GenerateHomeworldSkills(r)
	scoutSeg := resolveScoutSegment(r, upp, maxCareerTerms)
	spacerSeg := resolveSpacerSegment(r, scoutSeg.UPP, maxCareerTerms)

	wantFame := scoutSeg.Fame + spacerSeg.Fame
	wantCash := scoutSeg.Cash + spacerSeg.Cash
	wantWoundBadges := scoutSeg.WoundBadges + spacerSeg.WoundBadges
	wantSkills := len(homeworldSkills) + len(scoutSeg.Skills) + len(spacerSeg.Skills)

	if got.Fame != wantFame {
		t.Errorf("Fame = %d, want %d (sum of both segments)", got.Fame, wantFame)
	}

	if got.Cash != wantCash {
		t.Errorf("Cash = %d, want %d (sum of both segments)", got.Cash, wantCash)
	}

	if got.WoundBadges != wantWoundBadges {
		t.Errorf("WoundBadges = %d, want %d (sum of both segments)", got.WoundBadges, wantWoundBadges)
	}

	if len(got.Skills) != wantSkills {
		t.Errorf("len(Skills) = %d, want %d (homeworld + both segments)", len(got.Skills), wantSkills)
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

// TestCareerChainDeadInFirstSegmentEndsTheWholeAttempt confirms Book 1
// p.69's "Dying During Character Generation" voids the whole chain — the
// second listed career is never attempted once the first ends in Dead.
func TestCareerChainDeadInFirstSegmentEndsTheWholeAttempt(t *testing.T) {
	t.Parallel()

	const seed = 26 // confirmed by direct inspection: Scout dies in its 3rd term

	got, ok, err := GenerateCareerChainCharacter(dice.New(rand.NewPCG(seed, seed)), []string{"scout", "spacer"}, 0)
	if err != nil {
		t.Fatal(err)
	}

	if ok {
		t.Fatal("ok = true, want false (Scout died)")
	}

	if len(got.Careers) != 1 || got.Careers[0].Name != "Scout" {
		t.Fatalf("Careers = %+v, want exactly one Scout entry (Spacer never attempted)", got.Careers)
	}

	last := got.Careers[0].Terms[len(got.Careers[0].Terms)-1]
	if last.RiskResult != Dead {
		t.Fatalf("last term RiskResult = %v, want Dead", last.RiskResult)
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

	if !ok || len(uncapped.Careers) != 1 || len(uncapped.Careers[0].Terms) != 4 {
		t.Fatalf("uncapped = %+v, want a single 4-term Scout career (fixture assumption broken)", uncapped.Careers)
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

// TestCareerChainAgeTargetStopsBeforeTheNextListedCareer confirms that
// once the budget is exhausted between two listed careers, the second
// is never even attempted — no zero-term "attempted" entry for it,
// distinguishing "never attempted" (this test) from "attempted and
// failed Begin" (TestCareerChainFallsBackToCitizenWhenEveryListedCareerFails).
// Seed 1 confirmed by direct inspection: Scout naturally runs exactly 2
// terms (age 26) before Spacer would be attempted.
func TestCareerChainAgeTargetStopsBeforeTheNextListedCareer(t *testing.T) {
	t.Parallel()

	const seed = 1

	got, ok, err := GenerateCareerChainCharacter(dice.New(rand.NewPCG(seed, seed)), []string{"scout", "spacer"}, 26)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Fatal("ok = false, want true")
	}

	if len(got.Careers) != 1 || got.Careers[0].Name != "Scout" || len(got.Careers[0].Terms) != 2 {
		t.Fatalf("Careers = %+v, want exactly one 2-term Scout entry (Spacer never attempted)", got.Careers)
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

	if len(got.Skills) != len(homeworldSkills) {
		t.Errorf("len(Skills) = %d, want %d (homeworld skills only)", len(got.Skills), len(homeworldSkills))
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
