package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
)

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
