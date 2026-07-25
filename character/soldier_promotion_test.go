package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
)

func TestSoldierRankTablesMatchBook1P82(t *testing.T) {
	t.Parallel()

	wantEnlisted := [6]string{
		"Private", "Corporal", "Sergeant",
		"Staff Sergeant", "Master Sergeant", "Sergeant Major",
	}
	if soldierEnlistedRankNames != wantEnlisted {
		t.Errorf("soldierEnlistedRankNames =\n%v\nwant\n%v", soldierEnlistedRankNames, wantEnlisted)
	}

	wantOfficer := [7]string{
		"2nd Lieutenant", "1st Lieutenant", "Captain", "Major",
		"Lt Colonel", "Colonel", "General",
	}
	if soldierOfficerRankNames != wantOfficer {
		t.Errorf("soldierOfficerRankNames =\n%v\nwant\n%v", soldierOfficerRankNames, wantOfficer)
	}
}

func TestSoldierRankName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		isOfficer bool
		tier      int
		want      string
	}{
		{false, 1, "S1 Private"},
		{false, 6, "S6 Sergeant Major"},
		{true, 1, "O1 2nd Lieutenant"},
		{true, 4, "O4 Major"},
		{true, 7, "O7 General"},
	}

	for _, c := range cases {
		if got := soldierRankName(c.isOfficer, c.tier); got != c.want {
			t.Errorf("soldierRankName(%v, %d) = %q, want %q", c.isOfficer, c.tier, got, c.want)
		}
	}
}

// TestSoldierRankAutomaticSkill covers every defined and undefined
// (isOfficer, tier) combination across both tracks' own full tier
// range — genuinely distinct data from Marine's own table (S1 Private
// -> Fighter and O6 Colonel -> Leader have no Marine equivalent).
func TestSoldierRankAutomaticSkill(t *testing.T) {
	t.Parallel()

	cases := []struct {
		isOfficer bool
		tier      int
		want      string // "" means no automatic skill
	}{
		{false, 1, "Fighter"},
		{false, 2, ""},
		{false, 3, "Heavy Weapons"},
		{false, 4, "Leader"},
		{false, 5, ""},
		{false, 6, ""},
		{true, 1, "Leader"},
		{true, 2, ""},
		{true, 3, ""},
		{true, 4, "Tactics"},
		{true, 5, ""},
		{true, 6, "Leader"},
		{true, 7, ""},
	}

	for _, c := range cases {
		skill, ok := soldierRankAutomaticSkill(c.isOfficer, c.tier)

		if c.want == "" {
			if ok {
				t.Errorf("soldierRankAutomaticSkill(%v, %d) = (%v, true), want false", c.isOfficer, c.tier, skill)
			}

			continue
		}

		if !ok || skill.Name != c.want {
			t.Errorf(
				"soldierRankAutomaticSkill(%v, %d) = (%v, %v), want (%q, true)",
				c.isOfficer,
				c.tier,
				skill,
				ok,
				c.want,
			)
		}
	}
}

func TestRollSoldierCommissionRate(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))

	const trials = 20000

	fired := 0

	for range trials {
		if rollSoldierCommission(r, 7) {
			fired++
		}
	}

	gotPct := 100 * float64(fired) / trials
	if wantPct := 100.0 * 21 / 36; gotPct < wantPct-1 || gotPct > wantPct+1 {
		t.Errorf("rollSoldierCommission(c3=7) fired %.2f%% of %d trials, want ~%.2f%%", gotPct, trials, wantPct)
	}
}

// TestRollSoldierOfficerPromotionRate and
// TestRollSoldierEnlistedPromotionRate confirm a nonzero medalMod
// measurably shifts the pass rate versus mod=0.
func TestRollSoldierOfficerPromotionRate(t *testing.T) {
	t.Parallel()

	const trials = 20000

	r1 := dice.New(rand.NewPCG(1, 1))

	noModFired := 0

	for range trials {
		if rollSoldierOfficerPromotion(r1, 7, 0) {
			noModFired++
		}
	}

	r2 := dice.New(rand.NewPCG(2, 2))

	withModFired := 0

	for range trials {
		if rollSoldierOfficerPromotion(r2, 7, 3) {
			withModFired++
		}
	}

	if withModFired <= noModFired {
		t.Errorf("rollSoldierOfficerPromotion with medalMod=3 fired %d/%d, want more than mod=0's own %d/%d",
			withModFired, trials, noModFired, trials)
	}
}

func TestRollSoldierEnlistedPromotionRate(t *testing.T) {
	t.Parallel()

	const trials = 20000

	r1 := dice.New(rand.NewPCG(1, 1))

	noModFired := 0

	for range trials {
		if rollSoldierEnlistedPromotion(r1, 7, 0) {
			noModFired++
		}
	}

	r2 := dice.New(rand.NewPCG(2, 2))

	withModFired := 0

	for range trials {
		if rollSoldierEnlistedPromotion(r2, 7, 3) {
			withModFired++
		}
	}

	if withModFired <= noModFired {
		t.Errorf("rollSoldierEnlistedPromotion with medalMod=3 fired %d/%d, want more than mod=0's own %d/%d",
			withModFired, trials, noModFired, trials)
	}
}
