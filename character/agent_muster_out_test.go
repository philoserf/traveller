package character

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
)

func TestAgentMusterOutTablesMatchBook1P83(t *testing.T) {
	t.Parallel()

	wantMoney := [12]string{
		"Low Passage", "Low Passage", "Mid Passage", "High Passage",
		"Cr15,000", "StarPass", "Cr25,000", "Cr30,000",
		"Cr35,000", "Cr40,000", "Cr80,000", "Cr90,000",
	}
	if agentMusterOutMoney != wantMoney {
		t.Errorf("agentMusterOutMoney =\n%v\nwant\n%v", agentMusterOutMoney, wantMoney)
	}

	wantBenefits := [12]string{
		"Ship Share", "Forbidden Knowledge", "Wafer Jack", "C5 +1",
		"Str +1", "C2 +1", "C3 +1", "Ship Share",
		"Life Insurance", "TAS Fellow Membership", "Fame +2", "Knighthood",
	}
	if agentMusterOutBenefits != wantBenefits {
		t.Errorf("agentMusterOutBenefits =\n%v\nwant\n%v", agentMusterOutBenefits, wantBenefits)
	}
}

func TestResolveAgentMusterOutRollCountMatchesTerms(t *testing.T) {
	t.Parallel()

	careers := []Career{
		{},
		{Terms: make([]Term, 1)},
		{Terms: make([]Term, 5)},
		{Terms: make([]Term, 14)},
	}

	r := dice.New(rand.NewPCG(9, 10))

	for _, career := range careers {
		out := ResolveAgentMusterOut(r, career)
		if got := len(out.Money) + len(out.Benefits); got != len(career.Terms) {
			t.Errorf("total entries = %d, want %d (len(career.Terms))", got, len(career.Terms))
		}
	}
}

func TestResolveAgentMusterOutEntriesAreFromTables(t *testing.T) {
	t.Parallel()

	career := Career{Terms: make([]Term, 14)}
	r := dice.New(rand.NewPCG(15, 16))

	for range 200 {
		out := ResolveAgentMusterOut(r, career)

		for _, m := range out.Money {
			if !slices.Contains(agentMusterOutMoney[:], m) {
				t.Fatalf("ResolveAgentMusterOut granted unexpected Money entry %q", m)
			}
		}

		for _, b := range out.Benefits {
			if !slices.Contains(agentMusterOutBenefits[:], b) {
				t.Fatalf("ResolveAgentMusterOut granted unexpected Benefits entry %q", b)
			}
		}
	}
}

// TestResolveAgentMusterOutBothColumnsReachable guards the Uniform(2)
// column split, mirroring
// TestResolveMerchantMusterOutBothColumnsReachable's own never-fired-
// in-N-trials style.
func TestResolveAgentMusterOutBothColumnsReachable(t *testing.T) {
	t.Parallel()

	career := Career{Terms: make([]Term, 10)}
	r := dice.New(rand.NewPCG(51, 53))

	var sawMoney, sawBenefits bool

	for range 50 {
		out := ResolveAgentMusterOut(r, career)
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
