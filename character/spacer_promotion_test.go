package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
)

func TestSpacerRankTablesMatchBook1P81(t *testing.T) {
	t.Parallel()

	wantEnlisted := [6]string{
		"Spacehand", "Able Spacer", "Petty Officer Second",
		"Petty Officer First", "Chief Petty Officer", "Master Chief Petty Officer",
	}
	if spacerEnlistedRankNames != wantEnlisted {
		t.Errorf("spacerEnlistedRankNames =\n%v\nwant\n%v", spacerEnlistedRankNames, wantEnlisted)
	}

	wantOfficer := [7]string{
		"Ensign", "Sublieutenant", "Lieutenant", "Lt Commander",
		"Commander", "Captain", "Admiral",
	}
	if spacerOfficerRankNames != wantOfficer {
		t.Errorf("spacerOfficerRankNames =\n%v\nwant\n%v", spacerOfficerRankNames, wantOfficer)
	}
}

func TestSpacerRankName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		isOfficer bool
		tier      int
		want      string
	}{
		{false, 1, "R1 Spacehand"},
		{false, 6, "R6 Master Chief Petty Officer"},
		{true, 1, "O1 Ensign"},
		{true, 4, "O4 Lt Commander"},
		{true, 7, "O7 Admiral"},
	}

	for _, c := range cases {
		if got := spacerRankName(c.isOfficer, c.tier); got != c.want {
			t.Errorf("spacerRankName(%v, %d) = %q, want %q", c.isOfficer, c.tier, got, c.want)
		}
	}
}

// TestSpacerRankAutomaticSkill covers every defined and undefined
// (isOfficer, tier) combination across both tracks' own full tier
// range — genuinely distinct data from Marine's/Soldier's own tables.
func TestSpacerRankAutomaticSkill(t *testing.T) {
	t.Parallel()

	cases := []struct {
		isOfficer bool
		tier      int
		want      string // "" means no automatic skill
	}{
		{false, 1, "Fighter"},
		{false, 2, ""},
		{false, 3, ""},
		{false, 4, "Gunner"},
		{false, 5, "Sensors"},
		{false, 6, ""},
		{true, 1, "Astrogator"},
		{true, 2, ""},
		{true, 3, "Engineer"},
		{true, 4, "Pilot"},
		{true, 5, ""},
		{true, 6, "Leader"},
		{true, 7, ""},
	}

	for _, c := range cases {
		skill, ok := spacerRankAutomaticSkill(c.isOfficer, c.tier)

		if c.want == "" {
			if ok {
				t.Errorf("spacerRankAutomaticSkill(%v, %d) = (%v, true), want false", c.isOfficer, c.tier, skill)
			}

			continue
		}

		if !ok || skill.Name != c.want {
			t.Errorf(
				"spacerRankAutomaticSkill(%v, %d) = (%v, %v), want (%q, true)",
				c.isOfficer,
				c.tier,
				skill,
				ok,
				c.want,
			)
		}
	}
}

func TestRollSpacerCommissionRate(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))

	const trials = 20000

	fired := 0

	for range trials {
		if rollSpacerCommission(r, 7) {
			fired++
		}
	}

	gotPct := 100 * float64(fired) / trials
	if wantPct := 100.0 * 21 / 36; gotPct < wantPct-1 || gotPct > wantPct+1 {
		t.Errorf("rollSpacerCommission(c2=7) fired %.2f%% of %d trials, want ~%.2f%%", gotPct, trials, wantPct)
	}
}

// TestRollSpacerOfficerPromotionRate and TestRollSpacerRatingPromotionRate
// confirm a nonzero medalMod measurably shifts the pass rate versus
// mod=0.
func TestRollSpacerOfficerPromotionRate(t *testing.T) {
	t.Parallel()

	const trials = 20000

	r1 := dice.New(rand.NewPCG(1, 1))

	noModFired := 0

	for range trials {
		if rollSpacerOfficerPromotion(r1, 7, 0) {
			noModFired++
		}
	}

	r2 := dice.New(rand.NewPCG(2, 2))

	withModFired := 0

	for range trials {
		if rollSpacerOfficerPromotion(r2, 7, 3) {
			withModFired++
		}
	}

	if withModFired <= noModFired {
		t.Errorf("rollSpacerOfficerPromotion with medalMod=3 fired %d/%d, want more than mod=0's own %d/%d",
			withModFired, trials, noModFired, trials)
	}
}

func TestRollSpacerRatingPromotionRate(t *testing.T) {
	t.Parallel()

	const trials = 20000

	r1 := dice.New(rand.NewPCG(1, 1))

	noModFired := 0

	for range trials {
		if rollSpacerRatingPromotion(r1, 7, 0) {
			noModFired++
		}
	}

	r2 := dice.New(rand.NewPCG(2, 2))

	withModFired := 0

	for range trials {
		if rollSpacerRatingPromotion(r2, 7, 3) {
			withModFired++
		}
	}

	if withModFired <= noModFired {
		t.Errorf("rollSpacerRatingPromotion with medalMod=3 fired %d/%d, want more than mod=0's own %d/%d",
			withModFired, trials, noModFired, trials)
	}
}
