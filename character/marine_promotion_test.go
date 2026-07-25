package character

import (
	"math/rand/v2"
	"slices"
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

func TestMarineMedalModTableMatchesBook1P70(t *testing.T) {
	t.Parallel()

	want := map[string]int{"XS": 1, "MCUF": 2, "MCG": 3, "SEH": 4, "SEHD": 5}
	if len(marineMedalMod) != len(want) {
		t.Fatalf("marineMedalMod has %d entries, want %d", len(marineMedalMod), len(want))
	}

	for code, wantMod := range want {
		if got := marineMedalMod[code]; got != wantMod {
			t.Errorf("marineMedalMod[%q] = %d, want %d", code, got, wantMod)
		}
	}
}

func TestMarineMedalModSum(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		medals []string
		want   int
	}{
		{"no medals", nil, 0},
		{"one XS", []string{"XS"}, 1},
		{"mixed", []string{"XS", "MCUF", "MCG", "SEH"}, 1 + 2 + 3 + 4},
	}

	for _, c := range cases {
		if got := marineMedalModSum(c.medals); got != c.want {
			t.Errorf("%s: marineMedalModSum(%v) = %d, want %d", c.name, c.medals, got, c.want)
		}
	}
}

func TestMarineMedalModTotal(t *testing.T) {
	t.Parallel()

	terms := []Term{
		{Medals: []string{"XS"}},
		{Medals: []string{"MCUF", "XS"}},
		{},
	}

	if got, want := marineMedalModTotal(terms), 1+(2+1); got != want {
		t.Errorf("marineMedalModTotal(%+v) = %d, want %d", terms, got, want)
	}
}

func TestMarineRankState(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		terms       []Term
		wantOfficer bool
		wantTier    int
	}{
		{"no terms", nil, false, 1},
		{"one Commissioned term", []Term{{Commissioned: true}}, true, 1},
		{
			"Commissioned then two Promoted",
			[]Term{{Commissioned: true}, {Promoted: true}, {Promoted: true}},
			true, 3,
		},
		{
			"six Enlisted Promoted terms cap at 6",
			[]Term{
				{Promoted: true},
				{Promoted: true},
				{Promoted: true},
				{Promoted: true},
				{Promoted: true},
				{Promoted: true},
			},
			false,
			6,
		},
		{
			"a seventh Enlisted Promoted term stays capped at 6",
			[]Term{
				{Promoted: true},
				{Promoted: true},
				{Promoted: true},
				{Promoted: true},
				{Promoted: true},
				{Promoted: true},
				{Promoted: true},
			},
			false, 6,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			gotOfficer, gotTier := marineRankState(c.terms)
			if gotOfficer != c.wantOfficer || gotTier != c.wantTier {
				t.Errorf("marineRankState(%+v) = (%v, %d), want (%v, %d)",
					c.terms, gotOfficer, gotTier, c.wantOfficer, c.wantTier)
			}
		})
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

func TestMarineBranchAutomaticSkill(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))

	skill, ok := marineBranchAutomaticSkill(r, "Medical")
	if !ok || skill.Name != "Medic" {
		t.Errorf("marineBranchAutomaticSkill(Medical) = (%v, %v), want (Medic, true)", skill, ok)
	}

	skill, ok = marineBranchAutomaticSkill(r, "Technical")
	if !ok || !slices.Contains(theTradeChoices, skill.Name) {
		t.Errorf("marineBranchAutomaticSkill(Technical) = (%v, %v), want (one of %v, true)", skill, ok, theTradeChoices)
	}

	if _, ok := marineBranchAutomaticSkill(r, "Infantry"); ok {
		t.Error("marineBranchAutomaticSkill(Infantry) = true, want false")
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
