package character

import (
	"math/rand/v2"
	"regexp"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
)

func TestNobleMusterOutTablesMatchBook1P85(t *testing.T) {
	t.Parallel()

	wantMoney := [12]string{
		"StarPass", "StarPass", "StarPass", "StarPass",
		"Cr100,000", "Cr200,000", "Cr300,000", "Cr400,000",
		"Cr500,000", "Cr600,000", "Cr700,000", "Cr800,000",
	}
	if nobleMusterOutMoney != wantMoney {
		t.Errorf("nobleMusterOutMoney =\n%v\nwant\n%v", nobleMusterOutMoney, wantMoney)
	}

	wantBenefits := [12]string{
		"Forbidden Knowledge", "Str +1", "Wafer Jack", "C5 +1",
		"Directorship", "C2 +1", "C3 +1", "Int +1",
		"Ship Share", "Life Insurance", "TAS Life Membership", "Directorship",
	}
	if nobleMusterOutBenefits != wantBenefits {
		t.Errorf("nobleMusterOutBenefits =\n%v\nwant\n%v", nobleMusterOutBenefits, wantBenefits)
	}

	wantPower := [12]string{
		"Proxy (1)", "Proxy (2)", "Proxy (3)", "Proxy (4)", "Proxy (5)", "Proxy (6)",
		"Proxy (2D)", "Proxy (2D)", "Proxy (2D)", "Proxy (2D)", "Proxy (2D)", "Proxy (2D)",
	}
	if nobleMusterOutPower != wantPower {
		t.Errorf("nobleMusterOutPower =\n%v\nwant\n%v", nobleMusterOutPower, wantPower)
	}
}

var proxyPattern = regexp.MustCompile(`^Proxy \(\d+\)$`)

func TestRollNobleMusterOutPower(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))

	for row := range 6 {
		if got := rollNobleMusterOutPower(r, row); got != nobleMusterOutPower[row] {
			t.Errorf("row %d: rollNobleMusterOutPower = %q, want fixed %q", row, got, nobleMusterOutPower[row])
		}
	}

	for row := 6; row < 12; row++ {
		got := rollNobleMusterOutPower(r, row)
		if !proxyPattern.MatchString(got) {
			t.Errorf("row %d: rollNobleMusterOutPower = %q, want format %q", row, got, proxyPattern.String())
		}
	}
}

func TestResolveNobleMusterOutRollCountMatchesTerms(t *testing.T) {
	t.Parallel()

	careers := []Career{
		{},
		{Terms: make([]Term, 1)},
		{Terms: make([]Term, 5)},
		{Terms: make([]Term, 14)},
	}

	r := dice.New(rand.NewPCG(9, 10))

	for _, career := range careers {
		out := ResolveNobleMusterOut(r, career)
		if got := len(out.Money) + len(out.Benefits) + len(out.Entitlements); got != len(career.Terms) {
			t.Errorf("total entries = %d, want %d (len(career.Terms))", got, len(career.Terms))
		}
	}
}

// TestResolveNobleMusterOutEntriesAreFromTables confirms every returned
// entry validates against its own source: Money/Benefits against exact
// table membership, Entitlements (Power) against the "Proxy (N)" shape
// since rows 7-12 are dynamically resolved, not literal table members.
func TestResolveNobleMusterOutEntriesAreFromTables(t *testing.T) {
	t.Parallel()

	career := Career{Terms: make([]Term, 14)}
	r := dice.New(rand.NewPCG(15, 16))

	for range 200 {
		out := ResolveNobleMusterOut(r, career)

		for _, m := range out.Money {
			if !slices.Contains(nobleMusterOutMoney[:], m) {
				t.Fatalf("ResolveNobleMusterOut granted unexpected Money entry %q", m)
			}
		}

		for _, b := range out.Benefits {
			if !slices.Contains(nobleMusterOutBenefits[:], b) {
				t.Fatalf("ResolveNobleMusterOut granted unexpected Benefits entry %q", b)
			}
		}

		for _, e := range out.Entitlements {
			if !proxyPattern.MatchString(e) {
				t.Fatalf("ResolveNobleMusterOut granted unexpected Entitlements entry %q", e)
			}
		}
	}
}

// TestResolveNobleMusterOutAllThreeColumnsReachable guards the
// Uniform(3) column split, mirroring
// TestResolveScoutMusterOutBothColumnsReachable's own never-fired-in-N-
// trials style, generalized from 2 columns to 3.
func TestResolveNobleMusterOutAllThreeColumnsReachable(t *testing.T) {
	t.Parallel()

	career := Career{Terms: make([]Term, 10)}
	r := dice.New(rand.NewPCG(51, 53))

	var sawMoney, sawBenefits, sawEntitlements bool

	for range 50 {
		out := ResolveNobleMusterOut(r, career)
		if len(out.Money) > 0 {
			sawMoney = true
		}

		if len(out.Benefits) > 0 {
			sawBenefits = true
		}

		if len(out.Entitlements) > 0 {
			sawEntitlements = true
		}

		if sawMoney && sawBenefits && sawEntitlements {
			return
		}
	}

	t.Fatalf("Money, Benefits, and Entitlements didn't all appear across 50 trials of 10 rolls each "+
		"(sawMoney=%v, sawBenefits=%v, sawEntitlements=%v)", sawMoney, sawBenefits, sawEntitlements)
}
