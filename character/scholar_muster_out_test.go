package character

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

func TestScholarMusterOutTablesMatchBook1P76(t *testing.T) {
	t.Parallel()

	wantMoney := [11]string{
		"Low Passage", "Low Passage", "Mid Passage", "High Passage",
		"Cr15,000", "StarPass", "Cr25,000", "Cr30,000",
		"Cr35,000", "Cr40,000", "Cr50,000",
	}
	if scholarMusterOutMoney != wantMoney {
		t.Errorf("scholarMusterOutMoney =\n%v\nwant\n%v", scholarMusterOutMoney, wantMoney)
	}

	wantBenefits := [11]string{
		"C5 +1", "C5 +1", "Wafer Jack", "Edu +1",
		"Str +1", "C2 +1", "C3 +1", "Int +1",
		"Fame +1", "Ship Share", "TAS Fellow Membership",
	}
	if scholarMusterOutBenefits != wantBenefits {
		t.Errorf("scholarMusterOutBenefits =\n%v\nwant\n%v", scholarMusterOutBenefits, wantBenefits)
	}
}

func TestResolveScholarMusterOutRollCountMatchesTerms(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}

	careers := []Career{
		{},
		{Terms: make([]Term, 1)},
		{Terms: make([]Term, 5)},
		{Terms: make([]Term, 14)},
	}

	r := dice.New(rand.NewPCG(9, 10))

	for _, career := range careers {
		out := ResolveScholarMusterOut(r, career, upp)
		if got := len(out.Money) + len(out.Benefits); got != len(career.Terms) {
			t.Errorf("total entries = %d, want %d (len(career.Terms))", got, len(career.Terms))
		}
	}
}

func TestResolveScholarMusterOutEntriesAreFromTables(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}
	career := Career{Terms: make([]Term, 14)}
	r := dice.New(rand.NewPCG(15, 16))

	for range 200 {
		out := ResolveScholarMusterOut(r, career, upp)

		for _, m := range out.Money {
			if !slices.Contains(scholarMusterOutMoney[:], m) {
				t.Fatalf("ResolveScholarMusterOut granted unexpected Money entry %q", m)
			}
		}

		for _, b := range out.Benefits {
			if !slices.Contains(scholarMusterOutBenefits[:], b) {
				t.Fatalf("ResolveScholarMusterOut granted unexpected Benefits entry %q", b)
			}
		}
	}
}

// TestResolveScholarMusterOutBothColumnsReachable guards the Uniform(2)
// column split, mirroring TestResolveRogueMusterOutBothColumnsReachable's
// own never-fired-in-N-trials style.
func TestResolveScholarMusterOutBothColumnsReachable(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}
	career := Career{Terms: make([]Term, 10)}
	r := dice.New(rand.NewPCG(51, 53))

	var sawMoney, sawBenefits bool

	for range 50 {
		out := ResolveScholarMusterOut(r, career, upp)
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
