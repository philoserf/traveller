package character

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
)

func TestSoldierMusterOutTablesMatchBook1P82(t *testing.T) {
	t.Parallel()

	wantMoney := [10]string{
		"Low Passage", "Middle Passage", "High Passage", "StarPass",
		"Cr30,000", "Cr40,000", "Cr50,000", "Retirement x2",
		"Retirement x2", "Cr60,000",
	}
	if soldierMusterOutMoney != wantMoney {
		t.Errorf("soldierMusterOutMoney =\n%v\nwant\n%v", soldierMusterOutMoney, wantMoney)
	}

	wantBenefits := [10]string{
		"Forbidden Knowledge", "Str +1", "Wafer Jack", "C5 +1",
		"Int +1", "C2 +1", "C3 +1", "Life Insurance",
		"TAS Fellow Membership", "Knighthood",
	}
	if soldierMusterOutBenefits != wantBenefits {
		t.Errorf("soldierMusterOutBenefits =\n%v\nwant\n%v", soldierMusterOutBenefits, wantBenefits)
	}
}

func TestResolveSoldierMusterOutRollCountMatchesTerms(t *testing.T) {
	t.Parallel()

	careers := []Career{
		{},
		{Terms: make([]Term, 3)},
		{Terms: append(make([]Term, 4), Term{RiskResult: Wounded})},
		{Terms: append(make([]Term, 1), Term{RiskResult: Disabled})},
		{Terms: append(make([]Term, 6), Term{RiskResult: Dead})},
	}

	r := dice.New(rand.NewPCG(9, 10))

	for _, career := range careers {
		want := musterOutRollCount(
			career,
			rankBasedCareerFame(career, len(soldierEnlistedRankNames), len(soldierOfficerRankNames)),
		)

		out := ResolveSoldierMusterOut(r, career)
		if got := len(out.Money) + len(out.Benefits); got != want {
			t.Errorf("len(Money)+len(Benefits) = %d, want %d (musterOutRollCount)", got, want)
		}
	}
}

func TestResolveSoldierMusterOutEntriesAreFromTables(t *testing.T) {
	t.Parallel()

	career := Career{Terms: make([]Term, 14)}
	r := dice.New(rand.NewPCG(15, 16))

	for range 200 {
		out := ResolveSoldierMusterOut(r, career)

		for _, m := range out.Money {
			if !slices.Contains(soldierMusterOutMoney[:], m) {
				t.Fatalf("ResolveSoldierMusterOut granted unexpected Money entry %q", m)
			}
		}

		for _, b := range out.Benefits {
			if !slices.Contains(soldierMusterOutBenefits[:], b) {
				t.Fatalf("ResolveSoldierMusterOut granted unexpected Benefits entry %q", b)
			}
		}
	}
}

func TestResolveSoldierMusterOutBothColumnsReachable(t *testing.T) {
	t.Parallel()

	career := Career{Terms: make([]Term, 10)}
	r := dice.New(rand.NewPCG(51, 53))

	var sawMoney, sawBenefits bool

	for range 50 {
		out := ResolveSoldierMusterOut(r, career)
		if len(out.Money) > 0 {
			sawMoney = true
		}

		if len(out.Benefits) > 0 {
			sawBenefits = true
		}

		if sawMoney && sawBenefits {
			return
		}
	}

	t.Fatalf("Money and Benefits didn't both appear across 50 trials of 10 rolls each (sawMoney=%v, sawBenefits=%v)",
		sawMoney, sawBenefits)
}
