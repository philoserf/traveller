package character

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
)

func TestSpacerMusterOutTablesMatchBook1P81(t *testing.T) {
	t.Parallel()

	wantMoney := [11]string{
		"Low Passage", "Middle Passage", "High Passage", "StarPass",
		"Cr30,000", "Cr40,000", "Cr50,000", "Retirement x2",
		"Retirement x2", "Cr60,000", "Cr70,000",
	}
	if spacerMusterOutMoney != wantMoney {
		t.Errorf("spacerMusterOutMoney =\n%v\nwant\n%v", spacerMusterOutMoney, wantMoney)
	}

	wantBenefits := [11]string{
		"Forbidden Knowledge", "Str +1", "Wafer Jack", "C5 +1",
		"Str +1", "C2 +1", "C3 +1", "Int +1",
		"Ship Share", "Life Insurance", "Knighthood",
	}
	if spacerMusterOutBenefits != wantBenefits {
		t.Errorf("spacerMusterOutBenefits =\n%v\nwant\n%v", spacerMusterOutBenefits, wantBenefits)
	}
}

func TestResolveSpacerMusterOutRollCountMatchesTerms(t *testing.T) {
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
			rankBasedCareerFame(career, len(spacerEnlistedRankNames), len(spacerOfficerRankNames)),
		)

		out := ResolveSpacerMusterOut(r, career)
		if got := len(out.Money) + len(out.Benefits); got != want {
			t.Errorf("len(Money)+len(Benefits) = %d, want %d (musterOutRollCount)", got, want)
		}

		if len(career.Terms) > 0 && career.Terms[len(career.Terms)-1].RiskResult == Dead && out.RetirementPay != 0 {
			t.Errorf("RetirementPay = %d, want 0 for a character whose last term ended in Death (Book 1 p.69)",
				out.RetirementPay)
		}
	}
}

func TestResolveSpacerMusterOutEntriesAreFromTables(t *testing.T) {
	t.Parallel()

	career := Career{Terms: make([]Term, 14)}
	r := dice.New(rand.NewPCG(15, 16))

	for range 200 {
		out := ResolveSpacerMusterOut(r, career)

		for _, m := range out.Money {
			if !slices.Contains(spacerMusterOutMoney[:], m) {
				t.Fatalf("ResolveSpacerMusterOut granted unexpected Money entry %q", m)
			}
		}

		for _, b := range out.Benefits {
			if !slices.Contains(spacerMusterOutBenefits[:], b) {
				t.Fatalf("ResolveSpacerMusterOut granted unexpected Benefits entry %q", b)
			}
		}
	}
}

func TestResolveSpacerMusterOutBothColumnsReachable(t *testing.T) {
	t.Parallel()

	career := Career{Terms: make([]Term, 10)}
	r := dice.New(rand.NewPCG(51, 53))

	var sawMoney, sawBenefits bool

	for range 50 {
		out := ResolveSpacerMusterOut(r, career)
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
