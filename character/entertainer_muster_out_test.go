package character

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
)

func TestEntertainerMusterOutTablesMatchBook1P77(t *testing.T) {
	t.Parallel()

	wantMoney := [13]string{
		"Low Passage", "Low Passage", "Mid Passage", "High Passage",
		"Cr15,000", "StarPass", "Cr25,000", "Cr30,000",
		"Cr35,000", "Cr40,000", "Cr50,000", "Cr400,000", "Cr500,000",
	}
	if entertainerMusterOutMoney != wantMoney {
		t.Errorf("entertainerMusterOutMoney =\n%v\nwant\n%v", entertainerMusterOutMoney, wantMoney)
	}

	wantBenefits := [13]string{
		"C5 +1", "C5 +1", "Wafer Jack", "Edu +1",
		"Str +1", "C2 +1", "C3 +1", "Int +1",
		"Fame +1", "Ship Share", "TAS Fellow", "Knighthood", "TAS Life",
	}
	if entertainerMusterOutBenefits != wantBenefits {
		t.Errorf("entertainerMusterOutBenefits =\n%v\nwant\n%v", entertainerMusterOutBenefits, wantBenefits)
	}
}

func TestResolveEntertainerMusterOutRollCountMatchesTerms(t *testing.T) {
	t.Parallel()

	careers := []Career{
		{},
		{Terms: make([]Term, 1)},
		{Terms: make([]Term, 5)},
		{Terms: make([]Term, 14)},
	}

	r := dice.New(rand.NewPCG(9, 10))

	for _, career := range careers {
		out := ResolveEntertainerMusterOut(r, career, 10)
		if got := len(out.Money) + len(out.Benefits); got != len(career.Terms) {
			t.Errorf("total entries = %d, want %d (len(career.Terms))", got, len(career.Terms))
		}
	}
}

func TestResolveEntertainerMusterOutEntriesAreFromTables(t *testing.T) {
	t.Parallel()

	career := Career{Terms: make([]Term, 14)}
	r := dice.New(rand.NewPCG(15, 16))

	for range 200 {
		out := ResolveEntertainerMusterOut(r, career, 10)

		for _, m := range out.Money {
			if !slices.Contains(entertainerMusterOutMoney[:], m) {
				t.Fatalf("ResolveEntertainerMusterOut granted unexpected Money entry %q", m)
			}
		}

		for _, b := range out.Benefits {
			if !slices.Contains(entertainerMusterOutBenefits[:], b) {
				t.Fatalf("ResolveEntertainerMusterOut granted unexpected Benefits entry %q", b)
			}
		}
	}
}

// TestResolveEntertainerMusterOutBothColumnsReachable guards the
// Uniform(2) column split, mirroring
// TestResolveScholarMusterOutBothColumnsReachable's own never-fired-in-
// N-trials style.
func TestResolveEntertainerMusterOutBothColumnsReachable(t *testing.T) {
	t.Parallel()

	career := Career{Terms: make([]Term, 10)}
	r := dice.New(rand.NewPCG(51, 53))

	var sawMoney, sawBenefits bool

	for range 50 {
		out := ResolveEntertainerMusterOut(r, career, 10)
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
