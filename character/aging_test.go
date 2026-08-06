package character

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// resolveAgingForTest runs every Aging checkpoint from onset through
// finalAge in one call — a whole-life convenience wrapper around
// agingSimulation.checkpoint that only these tests need; production code
// only ever checkpoints incrementally, interleaved between career terms
// (see resolveCareerLoop, career_loop.go). Returns the resulting UPP,
// whether the character survived, the human-readable notes in
// chronological order, and the age actually reached (finalAge for a
// survivor, or the fatal checkpoint's own age for one who died).
func resolveAgingForTest(r *dice.Roller, upp UPP, finalAge int) (UPP, bool, []string, int) {
	var sim agingSimulation

	for _, age := range agingCheckpoints(finalAge) {
		upp = sim.checkpoint(r, upp, age)
		if !sim.alive() {
			return upp, false, sim.notes, sim.diedAtAge
		}
	}

	return upp, true, sim.notes, finalAge
}

// TestLifeStageForAge pins every boundary age in Book 1 p.89's own "THE
// STAGES OF LIFE" table, full-pinned rather than sampled.
func TestLifeStageForAge(t *testing.T) {
	t.Parallel()

	cases := []struct {
		age  int
		want int
	}{
		{0, 0},
		{1, 0},
		{2, 1},
		{9, 1},
		{10, 2},
		{17, 2},
		{18, 3},
		{25, 3},
		{26, 4},
		{33, 4},
		{34, 5},
		{41, 5},
		{42, 6},
		{49, 6},
		{50, 7},
		{57, 7},
		{58, 8},
		{65, 8},
		{66, 9},
		{71, 9},
		{72, 9},
		{100, 9},
	}

	for _, c := range cases {
		if got := LifeStageForAge(c.age); got != c.want {
			t.Errorf("LifeStageForAge(%d) = %d, want %d", c.age, got, c.want)
		}
	}
}

func TestAgeFromTermsServed(t *testing.T) {
	t.Parallel()

	cases := []struct {
		terms int
		want  int
	}{
		{0, 18},
		{1, 22},
		{maxCareerTerms, 74},
	}

	for _, c := range cases {
		if got := AgeFromTermsServed(c.terms); got != c.want {
			t.Errorf("AgeFromTermsServed(%d) = %d, want %d", c.terms, got, c.want)
		}
	}
}

// TestAgingCheckOutcomeBoundaries is the regression test for using a
// strict "<" boundary, not succeedsAgainst's own "<=" convention.
func TestAgingCheckOutcomeBoundaries(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name            string
		roll, lifeStage int
		want            bool
	}{
		{"one below life stage succeeds (feels the effect)", 4, 5, true},
		{"equal to life stage fails (strict less-than, not <=)", 5, 5, false},
		{"one above life stage fails", 6, 5, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := agingCheckOutcome(c.roll, c.lifeStage); got != c.want {
				t.Errorf("agingCheckOutcome(%d, %d) = %v, want %v", c.roll, c.lifeStage, got, c.want)
			}
		})
	}
}

// TestRollAgingCheckRate confirms rollAgingCheck is correctly wired to
// agingCheckOutcome (roll = TwoD6, target = lifeStage), mirroring
// rollDeepSpaceBonusRate's own statistical-rate test style. At
// lifeStage=5: P(2D6<5) = P(2)+P(3)+P(4) = (1+2+3)/36 = 6/36 ≈ 16.67%.
func TestRollAgingCheckRate(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(11, 13))

	const trials = 20000

	fired := 0

	for range trials {
		if rollAgingCheck(r, 5) {
			fired++
		}
	}

	gotPct := 100 * float64(fired) / trials
	if wantPct := 100.0 * 6 / 36; gotPct < wantPct-1 || gotPct > wantPct+1 {
		t.Errorf("rollAgingCheck(lifeStage=5) fired %.2f%% of %d trials, want ~%.2f%%", gotPct, trials, wantPct)
	}
}

// TestAgingCheckpoints pins every boundary this session's own research
// turned up: no checkpoints below onset, one at onset, and the full
// traditional-lifespan (74) count.
func TestAgingCheckpoints(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		finalAge int
		want     []int
	}{
		{"below onset", 33, nil},
		{"exactly onset", 34, []int{34}},
		{"just short of next checkpoint", 37, []int{34}},
		{"second checkpoint", 38, []int{34, 38}},
		{"traditional lifespan", 74, []int{34, 38, 42, 46, 50, 54, 58, 62, 66, 70, 74}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := agingCheckpoints(c.finalAge); !slices.Equal(got, c.want) {
				t.Errorf("agingCheckpoints(%d) = %v, want %v", c.finalAge, got, c.want)
			}
		})
	}
}

// TestAgingPositionsAt is the regression test for Mental Aging's own
// separate onset (66) gating C4 independently of Physical Aging's onset.
func TestAgingPositionsAt(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		age  int
		want []Position
	}{
		{"physical onset", 34, []Position{C1, C2, C3}},
		{"one year short of mental onset", 65, []Position{C1, C2, C3}},
		{"mental onset", 66, []Position{C1, C2, C3, C4}},
		{"well past mental onset", 100, []Position{C1, C2, C3, C4}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := agingPositionsAt(c.age); !slices.Equal(got, c.want) {
				t.Errorf("agingPositionsAt(%d) = %v, want %v", c.age, got, c.want)
			}
		})
	}
}

// TestClassifyAgingBatch includes the undefined four-simultaneous-zeros
// case (only possible at age 66+) as its own regression test for treating
// "three or more" as the book's own most severe named tier.
func TestClassifyAgingBatch(t *testing.T) {
	t.Parallel()

	cases := []struct {
		zeroedCount int
		want        agingSeverity
	}{
		{0, agingSeverityNone},
		{1, agingSeverityNone},
		{2, agingSeverityMajor},
		{3, agingSeverityExtreme},
		{4, agingSeverityExtreme},
	}

	for _, c := range cases {
		if got := classifyAgingBatch(c.zeroedCount); got != c.want {
			t.Errorf("classifyAgingBatch(%d) = %v, want %v", c.zeroedCount, got, c.want)
		}
	}
}

func TestAgingExtremeIllnessIsFatal(t *testing.T) {
	t.Parallel()

	cases := []struct {
		priorExtremeCount int
		want              bool
	}{
		{0, false},
		{1, true},
		{2, true},
	}

	for _, c := range cases {
		if got := agingExtremeIllnessIsFatal(c.priorExtremeCount); got != c.want {
			t.Errorf("agingExtremeIllnessIsFatal(%d) = %v, want %v", c.priorExtremeCount, got, c.want)
		}
	}
}

func TestResolveAgingNoEffectBeforeOnset(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{7, 7, 7, 7, 7, 7}}
	r := dice.New(rand.NewPCG(1, 1))

	got, survived, notes, _ := resolveAgingForTest(r, upp, 33)

	if got != upp {
		t.Errorf("resolveAgingForTest before onset changed upp: got %v, want unchanged %v", got, upp)
	}

	if !survived {
		t.Error("resolveAgingForTest before onset reported survived = false, want true")
	}

	if len(notes) != 0 {
		t.Errorf("resolveAgingForTest before onset returned notes %v, want none", notes)
	}
}

// TestResolveAgingNeverTriggersIllnessWithSufficientBuffer is
// dice-outcome-independent: starting every characteristic at 15, the
// traditional lifespan (74) runs 11 physical checkpoints and 3 mental
// ones, so even if every single Aging Check succeeded, no characteristic
// could drop below 15-11=4 — proving the loop's own bookkeeping (not
// relying on luck) regardless of what the dice actually produce.
func TestResolveAgingNeverTriggersIllnessWithSufficientBuffer(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{15, 15, 15, 15, 15, 15}}

	for _, seed := range []uint64{1, 2, 3, 4, 5} {
		r := dice.New(rand.NewPCG(seed, seed))

		got, survived, notes, _ := resolveAgingForTest(r, upp, 74)
		if !survived {
			t.Errorf("seed %d: survived = false, want true", seed)
		}

		if len(notes) != 0 {
			t.Errorf("seed %d: notes = %v, want none", seed, notes)
		}

		for i, c := range got.Characteristics {
			if c < 4 || c > 15 {
				t.Errorf("seed %d: Characteristics[%d] = %d, want in [4, 15]", seed, i, c)
			}
		}
	}
}

// TestResolveAgingProducesIllnessAndDeathOverManyTrials is the
// statistical proof that the illness/death paths actually fire:
// Str/Dex/End/Int all start at 1, so any single successful Aging Check
// immediately zeros that characteristic, and finalAge=110 runs well past
// 71 (Life Stage clamps at 9, its own high success rate) with every
// checkpoint from 66 on checking all four at once.
func TestResolveAgingProducesIllnessAndDeathOverManyTrials(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{1, 1, 1, 1, 12, 12}}

	const trials = 100

	died, illOnly := 0, 0

	for seed := uint64(1); seed <= trials; seed++ {
		r := dice.New(rand.NewPCG(seed, seed))

		got, survived, notes, _ := resolveAgingForTest(r, upp, 110)

		if !survived {
			died++

			if !slices.Contains(got.Characteristics[:], ehex.Value(0)) {
				t.Errorf("seed %d: died with no characteristic left at 0: %v", seed, got.Characteristics)
			}
		} else if len(notes) > 0 {
			illOnly++
		}
	}

	if died == 0 {
		t.Errorf("0 of %d trials died from Aging, want at least 1 (fatal path never exercised)", trials)
	}

	if died+illOnly == 0 {
		t.Errorf("0 of %d trials produced any illness or death notes, want at least 1", trials)
	}
}

// TestResolveAgingReportsReachedAge is #60's regression for the
// Age-coherence half: a survivor reaches finalAge exactly, while one who
// dies reaches only the fatal checkpoint's own age — never finalAge.
// Without this a sheet could report an Age older than the age in its own
// death note.
func TestResolveAgingReportsReachedAge(t *testing.T) {
	t.Parallel()

	t.Run("survivor reaches finalAge", func(t *testing.T) {
		t.Parallel()

		upp := UPP{Characteristics: [6]ehex.Value{15, 15, 15, 15, 15, 15}}

		_, survived, _, reachedAge := resolveAgingForTest(dice.New(rand.NewPCG(1, 1)), upp, 74)
		if !survived {
			t.Fatal("survived = false, want true (all-15 UPP has ample aging buffer)")
		}

		if reachedAge != 74 {
			t.Errorf("reachedAge = %d, want 74 (a survivor reaches the age asked for)", reachedAge)
		}
	})

	t.Run("the dead reach only the fatal checkpoint", func(t *testing.T) {
		t.Parallel()

		// finalAge far past the likely death so the cap is actually
		// observable — at 74 this fixture dies on the final checkpoint,
		// where a capped and an uncapped answer are indistinguishable.
		upp := UPP{Characteristics: [6]ehex.Value{2, 2, 2, 2, 2, 2}}

		_, survived, notes, reachedAge := resolveAgingForTest(dice.New(rand.NewPCG(1, 1)), upp, 110)
		if survived {
			t.Fatal("survived = true, want false (fixture assumption broke)")
		}

		if reachedAge >= 110 {
			t.Errorf("reachedAge = %d, want < 110 (death must cap the age reached)", reachedAge)
		}

		want := fmt.Sprintf("Age %d: died", reachedAge)
		if last := notes[len(notes)-1]; !strings.HasPrefix(last, want) {
			t.Errorf("final note = %q, want it to start with %q (reachedAge must match the death note)", last, want)
		}
	})
}

// TestAgingDeathStopsServiceAndMusterOut is the regression for PR #69's
// second review round: a character Aging has already killed must not
// serve another term, and must not muster out. Both were reachable
// because a career's own "did this kill them?" state is separate from
// the chain-wide Aging simulation — the loops checked Aging only after
// resolving a term, so the next career in a chain got one free term,
// and Mustering Out ran unconditionally afterward. Book 1 p.69's rule
// that a dead character never reaches Mustering Out was already encoded
// for career death (musterOutRollCount); this extends it to p.89.
func TestAgingDeathStopsServiceAndMusterOut(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{12, 12, 12, 12, 12, 12}}

	// A simulation already killed by an earlier career in the same chain.
	dead := func() *agingSimulation {
		return &agingSimulation{
			termsServed: 5,
			diedAtAge:   38,
			notes:       []string{"Age 38: died of natural causes (3 characteristics reduced to 0 for a second time)"},
		}
	}

	loops := map[string]func(*agingSimulation) int{
		"shared resolveCareerLoop (Scout)": func(a *agingSimulation) int {
			c, _ := resolveScoutCareerWithBudget(dice.New(rand.NewPCG(1, 1)), upp, maxCareerTerms, a)

			return len(c.Terms)
		},
		"hand-rolled Citizen": func(a *agingSimulation) int {
			c, _ := resolveCitizenCareerAndUPPWithBudget(dice.New(rand.NewPCG(1, 1)), upp, maxCareerTerms, a, nil)

			return len(c.Terms)
		},
		"hand-rolled Noble": func(a *agingSimulation) int {
			c, _, _ := resolveNobleCareerAndUPPWithBudget(dice.New(rand.NewPCG(1, 1)), upp, maxCareerTerms, a, nil)

			return len(c.Terms)
		},
		"hand-rolled Entertainer": func(a *agingSimulation) int {
			c, _, _ := resolveEntertainerCareerAndUPPWithBudget(
				dice.New(rand.NewPCG(1, 1)),
				upp,
				maxCareerTerms,
				a,
				nil,
			)

			return len(c.Terms)
		},
	}

	for name, serve := range loops {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := serve(dead()); got != 0 {
				t.Errorf("terms served = %d, want 0 (the character was already dead on entry)", got)
			}
		})
	}

	t.Run("segment grants no Mustering Out", func(t *testing.T) {
		t.Parallel()

		seg := resolveScoutSegment(dice.New(rand.NewPCG(1, 1)), upp, maxCareerTerms, segmentContext{Aging: dead()})

		if n := len(seg.Career.MusteringOut.Benefits) + len(seg.Career.MusteringOut.Money); n != 0 {
			t.Errorf("Mustering Out entries = %d, want 0 (the dead don't muster out)", n)
		}

		if seg.Cash != 0 {
			t.Errorf("Cash = %d, want 0", seg.Cash)
		}
	})
}

// TestFailedBeginAttemptsCostAYear is #62's regression. Book 1 p.65
// charges one year per failed Begin or Retry, which generation used to
// discard entirely: Age was 18 + 4*terms, so every failed entry attempt
// vanished from the chronology and, with it, from Birthdate, Life Stage,
// Aging checkpoints and -age budgeting.
//
// Driven through the public generators rather than the counter directly,
// since the counter being right matters only if it reaches Age.
func TestFailedBeginAttemptsCostAYear(t *testing.T) {
	t.Parallel()

	// A zero UPP fails every roll-based Begin there is.
	zero := UPP{}

	t.Run("one failed roll costs one year", func(t *testing.T) {
		t.Parallel()

		c, ok := buildMarineCharacter(dice.New(rand.NewPCG(1, 1)), zero, "hw", nil, false, false)
		if ok || len(c.Careers[0].Terms) != 0 {
			t.Fatal("fixture qualified, want a failed Begin")
		}

		if c.Age != 19 {
			t.Errorf("Age = %d, want 19 (18 + one failed Begin)", c.Age)
		}
	})

	t.Run("Scout's Retry is a second failed attempt", func(t *testing.T) {
		t.Parallel()

		c, ok := buildScoutCharacter(dice.New(rand.NewPCG(1, 1)), zero, "hw", nil)
		if ok || len(c.Careers[0].Terms) != 0 {
			t.Fatal("fixture qualified, want both Begin and Retry to fail")
		}

		if c.Age != 20 {
			t.Errorf("Age = %d, want 20 — Scout alone has a Retry, so failing to qualify costs two years",
				c.Age)
		}
	})

	t.Run("a prerequisite that isn't a roll costs nothing", func(t *testing.T) {
		t.Parallel()

		// Noble's Begin is Soc B+ — a threshold, not an attempt. There is
		// no roll to fail, so p.65's per-attempt year never applies.
		c, ok := buildNobleCharacter(dice.New(rand.NewPCG(1, 1)), zero, "hw", nil, nil)
		if ok || len(c.Careers[0].Terms) != 0 {
			t.Fatal("fixture qualified, want Soc below B")
		}

		if c.Age != 18 {
			t.Errorf("Age = %d, want 18 — failing a prerequisite is not a failed attempt", c.Age)
		}
	})
}

// TestFailedBeginYearsAccumulateAcrossAChain confirms the years are
// counted over a whole life rather than per career: each career in a
// chain whose Begin roll fails adds its own year, and the total reaches
// Age. A per-career counter would reset at every transfer and lose them.
func TestFailedBeginYearsAccumulateAcrossAChain(t *testing.T) {
	t.Parallel()

	// Precondition: at least one listed career fails its Begin roll, so
	// there is a charged year to accumulate. Searched rather than pinned
	// to a seed — the old fixture relied on seed 1 happening to fail all
	// three Begins, which is only true until the dice stream moves.
	chain := []string{"marine", "spacer", "soldier"}

	seed := seedFor(t, "a chain where at least one career fails to Begin", func(seed uint64) bool {
		c, _, err := GenerateCareerChainCharacter(dice.New(rand.NewPCG(seed, seed)), chain, 0)
		if err != nil {
			return false
		}

		for _, career := range c.Careers {
			if len(career.Terms) == 0 {
				return true
			}
		}

		return false
	})

	got, _, err := GenerateCareerChainCharacter(dice.New(rand.NewPCG(seed, seed)), chain, 0)
	if err != nil {
		t.Fatal(err)
	}

	failed := 0

	for _, career := range got.Careers {
		if len(career.Terms) == 0 {
			failed++
		}
	}

	termsServed := 0
	for _, career := range got.Careers {
		termsServed += len(career.Terms)
	}

	if want := AgeFromTermsServed(termsServed) + failed; got.Age != want {
		t.Errorf("Age = %d, want %d (%d terms plus %d failed Begins)",
			got.Age, want, termsServed, failed)
	}
}

// TestAgingDeathDuringAFailedBeginStopsGeneration covers PR #75's own
// review findings: a failed Begin costs a year, that year can cross an
// Aging checkpoint, and the checkpoint can be fatal — so a segment that
// served no terms at all can still have killed the character.
//
// Both paths this guards were reachable only through a zero-term
// segment, which is exactly the shape the chain's "nothing happened,
// carry on" continuation was written to skip past.
func TestAgingDeathDuringAFailedBeginStopsGeneration(t *testing.T) {
	t.Parallel()

	// One year short of a checkpoint (34) and one extreme illness in
	// already, so the next fatal batch needs only this one year to land.
	almostGone := func() *agingSimulation {
		return &agingSimulation{
			termsServed:  3, // age 30
			failedYears:  3, // age 33 — a single failed Begin reaches 34
			extremeCount: 1, // a second extreme batch is fatal
		}
	}

	frail := UPP{Characteristics: [6]ehex.Value{1, 1, 1, 1, 0, 0}}

	t.Run("the year is charged and can kill", func(t *testing.T) {
		t.Parallel()

		var killed bool

		for seed := range uint64(200) {
			aging := almostGone()
			_, _ = resolveMarineCareerWithBudget(
				dice.New(rand.NewPCG(seed+1, seed+1)),
				frail,
				maxCareerTerms,
				aging,
				false,
				false,
			)

			if !aging.alive() {
				killed = true

				if aging.age() != 34 {
					t.Errorf("died at %d, want the age-34 checkpoint the failed Begin's year reached", aging.age())
				}

				break
			}
		}

		if !killed {
			t.Fatal("no seed died during a failed Begin — the fixture can't reach the path under test")
		}
	})

	t.Run("a dead character starts no further career", func(t *testing.T) {
		t.Parallel()

		dead := &agingSimulation{termsServed: 5, diedAtAge: 38, notes: []string{"Age 38: died of natural causes (x)"}}

		career, _ := resolveMarineCareerWithBudget(
			dice.New(rand.NewPCG(1, 1)),
			frail,
			maxCareerTerms,
			dead,
			false,
			false,
		)
		if len(career.Terms) != 0 {
			t.Errorf("terms served = %d, want 0", len(career.Terms))
		}
	})
}

// TestScoutRetryIsRolledAfterTheFailedBeginYear covers the second review
// finding: BeginScout and RetryScout used to roll back to back inside
// one call, with both years charged afterward. That let the Retry be
// taken by a character the intervening year had already killed.
//
// Asserted through the roller rather than the outcome: if the Retry were
// still rolled after a fatal checkpoint it would consume dice, so a
// dead-on-arrival simulation must leave the sequence untouched.
func TestScoutRetryIsRolledAfterTheFailedBeginYear(t *testing.T) {
	t.Parallel()

	frail := UPP{Characteristics: [6]ehex.Value{1, 1, 1, 1, 12, 0}}

	// Already dead: neither attempt may be rolled at all.
	dead := &agingSimulation{termsServed: 5, diedAtAge: 38, notes: []string{"Age 38: died of natural causes (x)"}}

	career, _ := resolveScoutCareerWithBudget(dice.New(rand.NewPCG(1, 1)), frail, maxCareerTerms, dead)
	if len(career.Terms) != 0 {
		t.Errorf("terms served = %d, want 0 (the character was dead before the career began)", len(career.Terms))
	}

	// Edu 12 means RetryScout always succeeds, so a Scout who reaches it
	// alive always qualifies — proving the split didn't drop the Retry.
	alive := &agingSimulation{}

	qualified := false

	for seed := range uint64(50) {
		c, _ := resolveScoutCareerWithBudget(dice.New(rand.NewPCG(seed+1, seed+1)), frail, maxCareerTerms, alive)
		if len(c.Terms) > 0 {
			qualified = true

			break
		}

		alive = &agingSimulation{}
	}

	if !qualified {
		t.Error("no Scout qualified despite Edu 12 — the Retry is no longer being rolled")
	}
}
