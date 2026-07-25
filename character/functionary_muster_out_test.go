package character

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
)

func TestFunctionaryMusterOutTablesMatchBook1P87(t *testing.T) {
	t.Parallel()

	wantMoney := [11]string{
		"Cr5,000", "Cr10,000", "Cr15,000", "Cr20,000",
		"High Passage", "StarPass", "Cr25,000", "Cr30,000",
		"Cr35,000", "Pension x2", "Pension x2",
	}
	if functionaryMusterOutMoney != wantMoney {
		t.Errorf("functionaryMusterOutMoney =\n%v\nwant\n%v", functionaryMusterOutMoney, wantMoney)
	}

	wantBenefits := [11]string{
		"Forbidden Knowledge", "Str +1", "Wafer Jack", "Str +1",
		"C2 +1", "C3 +1", "Int +1", "Life Insurance",
		"TAS Fellow Membership", "Knighthood", "Directorship",
	}
	if functionaryMusterOutBenefits != wantBenefits {
		t.Errorf("functionaryMusterOutBenefits =\n%v\nwant\n%v", functionaryMusterOutBenefits, wantBenefits)
	}
}

// TestFunctionaryMusterOutRollCountDoesNotDoubleOnDisabled is the
// regression test for the bug this slice caught during its own
// implementation, before code review: functionaryMusterOutRollCount
// must NOT reuse scoutMusterOutRollCount's own "double on Disabled"
// rule — Functionary's own Disabled reuse means "Office Politics
// failed, career ends," not the universal p.65 physical disability the
// Double Benefits rule is actually about.
func TestFunctionaryMusterOutRollCountDoesNotDoubleOnDisabled(t *testing.T) {
	t.Parallel()

	career := Career{Terms: []Term{{RiskResult: Unharmed}, {RiskResult: Disabled}}}

	if got := functionaryMusterOutRollCount(career); got != len(career.Terms) {
		t.Errorf("functionaryMusterOutRollCount(...) = %d, want %d (never doubled)", got, len(career.Terms))
	}
}

func TestResolveFunctionaryMusterOutRollCountMatchesTerms(t *testing.T) {
	t.Parallel()

	careers := []Career{
		{},
		{Terms: make([]Term, 1)},
		{Terms: make([]Term, 5)},
		{Terms: make([]Term, 14)},
	}

	r := dice.New(rand.NewPCG(9, 10))

	for _, career := range careers {
		out := ResolveFunctionaryMusterOut(r, career, 0)
		if got := len(out.Money) + len(out.Benefits); got != len(career.Terms) {
			t.Errorf("total entries = %d, want %d (len(career.Terms))", got, len(career.Terms))
		}
	}
}

func TestResolveFunctionaryMusterOutEntriesAreFromTables(t *testing.T) {
	t.Parallel()

	career := Career{Terms: make([]Term, 14)}
	r := dice.New(rand.NewPCG(15, 16))

	for range 200 {
		out := ResolveFunctionaryMusterOut(r, career, 0)

		for _, m := range out.Money {
			if !slices.Contains(functionaryMusterOutMoney[:], m) {
				t.Fatalf("ResolveFunctionaryMusterOut granted unexpected Money entry %q", m)
			}
		}

		for _, b := range out.Benefits {
			if !slices.Contains(functionaryMusterOutBenefits[:], b) {
				t.Fatalf("ResolveFunctionaryMusterOut granted unexpected Benefits entry %q", b)
			}
		}
	}
}

// TestResolveFunctionaryMusterOutAutomatics confirms both of Book 1
// p.87's own Automatic grants: Gold Watch (always, scaled by Terms) and
// Directorship (only when the final Rank is F6+) — the first real use
// of MusteringOut.Automatics in this codebase.
func TestResolveFunctionaryMusterOutAutomatics(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))

	noTerms := ResolveFunctionaryMusterOut(r, Career{}, 0)
	if len(noTerms.Automatics) != 0 {
		t.Errorf("Automatics = %v, want none for a never-qualified attempt", noTerms.Automatics)
	}

	lowTier := ResolveFunctionaryMusterOut(r, Career{Terms: make([]Term, 3)}, 3)
	if len(lowTier.Automatics) != 1 || lowTier.Automatics[0] != "Gold Watch (Cr300)" {
		t.Errorf("Automatics = %v, want exactly [%q]", lowTier.Automatics, "Gold Watch (Cr300)")
	}

	highTier := ResolveFunctionaryMusterOut(r, Career{Terms: make([]Term, 5)}, 6)
	want := []string{"Gold Watch (Cr500)", "Directorship"}

	if !slices.Equal(highTier.Automatics, want) {
		t.Errorf("Automatics = %v, want %v", highTier.Automatics, want)
	}
}
