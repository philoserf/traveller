package character

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
)

func TestRogueMusterOutTablesMatchBook1P84(t *testing.T) {
	t.Parallel()

	wantMoney := [12]string{
		"Cr40,000", "StarPass", "StarPass", "High Passage",
		"High Passage", "StarPass", "Cr25,000", "Cr30,000",
		"Cr35,000", "Cr40,000", "Cr50,000", "Cr90,000",
	}
	if rogueMusterOutMoney != wantMoney {
		t.Errorf("rogueMusterOutMoney =\n%v\nwant\n%v", rogueMusterOutMoney, wantMoney)
	}

	wantBenefits := [12]string{
		"Str +1", "C5 +1", "Wafer Jack", "C2 +1",
		"C3 +1", "Life Insurance", "Ship Share", "Knighthood",
		"Ship Share", "Ship Share", "Ship Share", "Knighthood",
	}
	if rogueMusterOutBenefits != wantBenefits {
		t.Errorf("rogueMusterOutBenefits =\n%v\nwant\n%v", rogueMusterOutBenefits, wantBenefits)
	}
}

func TestResolveRogueMusterOutRollCountMatchesTerms(t *testing.T) {
	t.Parallel()

	careers := []Career{
		{},
		{Terms: make([]Term, 1)},
		{Terms: make([]Term, 5)},
		{Terms: make([]Term, 14)},
	}

	r := dice.New(rand.NewPCG(9, 10))

	for _, career := range careers {
		out := ResolveRogueMusterOut(r, career)
		if got := len(out.Money) + len(out.Benefits); got != len(career.Terms) {
			t.Errorf("total entries = %d, want %d (len(career.Terms))", got, len(career.Terms))
		}
	}
}

func TestResolveRogueMusterOutEntriesAreFromTables(t *testing.T) {
	t.Parallel()

	career := Career{Terms: make([]Term, 14)}
	r := dice.New(rand.NewPCG(15, 16))

	for range 200 {
		out := ResolveRogueMusterOut(r, career)

		for _, m := range out.Money {
			if !slices.Contains(rogueMusterOutMoney[:], m) {
				t.Fatalf("ResolveRogueMusterOut granted unexpected Money entry %q", m)
			}
		}

		for _, b := range out.Benefits {
			if !slices.Contains(rogueMusterOutBenefits[:], b) {
				t.Fatalf("ResolveRogueMusterOut granted unexpected Benefits entry %q", b)
			}
		}
	}
}

// TestResolveRogueMusterOutBothColumnsReachable guards the Uniform(2)
// column split, mirroring TestResolveScoutMusterOutBothColumnsReachable's
// own never-fired-in-N-trials style (character/career_muster_out_test.go).
func TestResolveRogueMusterOutBothColumnsReachable(t *testing.T) {
	t.Parallel()

	career := Career{Terms: make([]Term, 10)}
	r := dice.New(rand.NewPCG(51, 53))

	var sawMoney, sawBenefits bool

	for range 50 {
		out := ResolveRogueMusterOut(r, career)
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
