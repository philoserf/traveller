package character

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
)

func TestMerchantMusterOutTablesMatchBook1P80(t *testing.T) {
	t.Parallel()

	wantMoney := [12]string{
		"Cr40,000", "StarPass", "StarPass", "High Passage",
		"High Passage", "StarPass", "Cr25,000", "Cr30,000",
		"Cr35,000", "Cr40,000", "Cr50,000", "Cr90,000",
	}
	if merchantMusterOutMoney != wantMoney {
		t.Errorf("merchantMusterOutMoney =\n%v\nwant\n%v", merchantMusterOutMoney, wantMoney)
	}

	wantBenefits := [12]string{
		"Str +1", "C5 +1", "Wafer Jack", "C2 +1",
		"C3 +1", "Life Insurance", "Ship Share", "Knighthood",
		"Ship Share", "Ship Share", "Ship Share", "Knighthood",
	}
	if merchantMusterOutBenefits != wantBenefits {
		t.Errorf("merchantMusterOutBenefits =\n%v\nwant\n%v", merchantMusterOutBenefits, wantBenefits)
	}
}

func TestResolveMerchantMusterOutRollCountMatchesTerms(t *testing.T) {
	t.Parallel()

	careers := []Career{
		{},
		{Terms: make([]Term, 1)},
		{Terms: make([]Term, 5)},
		{Terms: make([]Term, 14)},
	}

	r := dice.New(rand.NewPCG(9, 10))

	for _, career := range careers {
		out := ResolveMerchantMusterOut(r, career, true, 3)
		if got := len(out.Money) + len(out.Benefits); got != len(career.Terms) {
			t.Errorf("total entries = %d, want %d (len(career.Terms))", got, len(career.Terms))
		}
	}
}

func TestResolveMerchantMusterOutEntriesAreFromTables(t *testing.T) {
	t.Parallel()

	career := Career{Terms: make([]Term, 14)}
	r := dice.New(rand.NewPCG(15, 16))

	for range 200 {
		out := ResolveMerchantMusterOut(r, career, true, 3)

		for _, m := range out.Money {
			if !slices.Contains(merchantMusterOutMoney[:], m) {
				t.Fatalf("ResolveMerchantMusterOut granted unexpected Money entry %q", m)
			}
		}

		for _, b := range out.Benefits {
			if !slices.Contains(merchantMusterOutBenefits[:], b) {
				t.Fatalf("ResolveMerchantMusterOut granted unexpected Benefits entry %q", b)
			}
		}
	}
}

// TestResolveMerchantMusterOutBothColumnsReachable guards the
// Uniform(2) column split, mirroring
// TestResolveScholarMusterOutBothColumnsReachable's own never-fired-in-
// N-trials style.
func TestResolveMerchantMusterOutBothColumnsReachable(t *testing.T) {
	t.Parallel()

	career := Career{Terms: make([]Term, 10)}
	r := dice.New(rand.NewPCG(51, 53))

	var sawMoney, sawBenefits bool

	for range 50 {
		out := ResolveMerchantMusterOut(r, career, false, 0)
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
