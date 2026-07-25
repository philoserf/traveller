package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
)

func TestMarineRankTablesMatchBook1P86(t *testing.T) {
	t.Parallel()

	wantEnlisted := [6]string{
		"Private", "Lance Corporal", "Sergeant",
		"Staff Sergeant", "Master Sergeant", "Sergeant Major",
	}
	if marineEnlistedRankNames != wantEnlisted {
		t.Errorf("marineEnlistedRankNames =\n%v\nwant\n%v", marineEnlistedRankNames, wantEnlisted)
	}

	wantOfficer := [7]string{
		"2nd Lieutenant", "1st Lieutenant", "Captain", "Force Commander",
		"Lt Coronel", "Coronel", "Brigadier",
	}
	if marineOfficerRankNames != wantOfficer {
		t.Errorf("marineOfficerRankNames =\n%v\nwant\n%v", marineOfficerRankNames, wantOfficer)
	}
}

func TestMarineRankName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		isOfficer bool
		tier      int
		want      string
	}{
		{false, 1, "M1 Private"},
		{false, 6, "M6 Sergeant Major"},
		{true, 1, "O1 2nd Lieutenant"},
		{true, 4, "O4 Force Commander"},
		{true, 7, "O7 Brigadier"},
	}

	for _, c := range cases {
		if got := marineRankName(c.isOfficer, c.tier); got != c.want {
			t.Errorf("marineRankName(%v, %d) = %q, want %q", c.isOfficer, c.tier, got, c.want)
		}
	}
}

// TestMarineRankAutomaticSkill covers every defined and undefined
// (isOfficer, tier) combination across both tracks' own full tier range.
func TestMarineRankAutomaticSkill(t *testing.T) {
	t.Parallel()

	cases := []struct {
		isOfficer bool
		tier      int
		want      string // "" means no automatic skill
	}{
		{false, 1, ""},
		{false, 2, ""},
		{false, 3, "Heavy Weapons"},
		{false, 4, "Tactics"},
		{false, 5, "Leader"},
		{false, 6, ""},
		{true, 1, "Leader"},
		{true, 2, ""},
		{true, 3, ""},
		{true, 4, "Tactics"},
		{true, 5, ""},
		{true, 6, ""},
		{true, 7, ""},
	}

	for _, c := range cases {
		skill, ok := marineRankAutomaticSkill(c.isOfficer, c.tier)

		if c.want == "" {
			if ok {
				t.Errorf("marineRankAutomaticSkill(%v, %d) = (%v, true), want false", c.isOfficer, c.tier, skill)
			}

			continue
		}

		if !ok || skill.Name != c.want {
			t.Errorf(
				"marineRankAutomaticSkill(%v, %d) = (%v, %v), want (%q, true)",
				c.isOfficer,
				c.tier,
				skill,
				ok,
				c.want,
			)
		}
	}
}

func TestRollMarineCommissionRate(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))

	const trials = 20000

	fired := 0

	for range trials {
		if rollMarineCommission(r, 7) {
			fired++
		}
	}

	gotPct := 100 * float64(fired) / trials
	if wantPct := 100.0 * 21 / 36; gotPct < wantPct-1 || gotPct > wantPct+1 {
		t.Errorf("rollMarineCommission(c3=7) fired %.2f%% of %d trials, want ~%.2f%%", gotPct, trials, wantPct)
	}
}

// TestRollMarineOfficerPromotionRate and TestRollMarineEnlistedPromotionRate
// confirm a nonzero medalMod measurably shifts the pass rate versus mod=0.
func TestRollMarineOfficerPromotionRate(t *testing.T) {
	t.Parallel()

	const trials = 20000

	r1 := dice.New(rand.NewPCG(1, 1))

	noModFired := 0

	for range trials {
		if rollMarineOfficerPromotion(r1, 7, 0) {
			noModFired++
		}
	}

	r2 := dice.New(rand.NewPCG(2, 2))

	withModFired := 0

	for range trials {
		if rollMarineOfficerPromotion(r2, 7, 3) {
			withModFired++
		}
	}

	if withModFired <= noModFired {
		t.Errorf("rollMarineOfficerPromotion with medalMod=3 fired %d/%d, want more than mod=0's own %d/%d",
			withModFired, trials, noModFired, trials)
	}
}

func TestRollMarineEnlistedPromotionRate(t *testing.T) {
	t.Parallel()

	const trials = 20000

	r1 := dice.New(rand.NewPCG(1, 1))

	noModFired := 0

	for range trials {
		if rollMarineEnlistedPromotion(r1, 7, 0) {
			noModFired++
		}
	}

	r2 := dice.New(rand.NewPCG(2, 2))

	withModFired := 0

	for range trials {
		if rollMarineEnlistedPromotion(r2, 7, 3) {
			withModFired++
		}
	}

	if withModFired <= noModFired {
		t.Errorf("rollMarineEnlistedPromotion with medalMod=3 fired %d/%d, want more than mod=0's own %d/%d",
			withModFired, trials, noModFired, trials)
	}
}
