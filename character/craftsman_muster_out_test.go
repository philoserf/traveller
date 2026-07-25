package character

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
)

func TestCraftsmanMusterOutTablesMatchBook1P75(t *testing.T) {
	t.Parallel()

	wantMoney := [12]string{
		"Low Passage", "Low Passage", "Mid Passage", "High Passage",
		"Cr15,000", "StarPass", "Cr25,000", "Cr30,000",
		"Cr35,000", "Cr40,000", "Cr50,000", "Cr60,000",
	}
	if craftsmanMusterOutMoney != wantMoney {
		t.Errorf("craftsmanMusterOutMoney =\n%v\nwant\n%v", craftsmanMusterOutMoney, wantMoney)
	}

	wantBenefits := [12]string{
		"Forbidden Knowledge", "Forbidden Knowledge", "Wafer Jack", "C5 +1",
		"Str +1", "C2 +1", "C3 +1", "Int +1",
		"Ship Share", "TAS Fellow Membership", "Director", "TAS Life Membership",
	}
	if craftsmanMusterOutBenefits != wantBenefits {
		t.Errorf("craftsmanMusterOutBenefits =\n%v\nwant\n%v", craftsmanMusterOutBenefits, wantBenefits)
	}
}

func TestResolveCraftsmanMusterOutRollCountMatchesTerms(t *testing.T) {
	t.Parallel()

	careers := []Career{
		{},
		{Terms: make([]Term, 1)},
		{Terms: make([]Term, 5)},
		{Terms: make([]Term, 14)},
	}

	r := dice.New(rand.NewPCG(9, 10))

	for _, career := range careers {
		out := ResolveCraftsmanMusterOut(r, career)
		if got := len(out.Money) + len(out.Benefits); got != len(career.Terms) {
			t.Errorf("total entries = %d, want %d (len(career.Terms))", got, len(career.Terms))
		}
	}
}

func TestResolveCraftsmanMusterOutEntriesAreFromTables(t *testing.T) {
	t.Parallel()

	career := Career{Terms: make([]Term, 14)}
	r := dice.New(rand.NewPCG(15, 16))

	for range 200 {
		out := ResolveCraftsmanMusterOut(r, career)

		for _, m := range out.Money {
			if !slices.Contains(craftsmanMusterOutMoney[:], m) {
				t.Fatalf("ResolveCraftsmanMusterOut granted unexpected Money entry %q", m)
			}
		}

		for _, b := range out.Benefits {
			if !slices.Contains(craftsmanMusterOutBenefits[:], b) {
				t.Fatalf("ResolveCraftsmanMusterOut granted unexpected Benefits entry %q", b)
			}
		}
	}
}
