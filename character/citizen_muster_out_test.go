package character

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
)

// TestCitizenMusterOutRowClamp confirms rollCitizenMusterOutRow's own
// max (11, one fewer than Scout's 12) is wired correctly into the shared
// musterOutRow helper — musterOutRow's own general clamp behavior is
// already covered by TestScoutMusterOutRow.
func TestCitizenMusterOutRowClamp(t *testing.T) {
	t.Parallel()

	cases := []struct {
		roll int
		want int
	}{
		{1, 0},
		{6, 5},
		{11, 10},
		{12, 10},
		{20, 10},
		{99, 10},
	}

	for _, c := range cases {
		if got := musterOutRow(c.roll, 11); got != c.want {
			t.Errorf("musterOutRow(%d, 11) = %d, want %d", c.roll, got, c.want)
		}
	}
}

// TestCitizenMusterOutTablesMatchBook1P78 pins both 11-entry Citizen
// Mustering Out columns in full, transcribed directly from the page
// image — not just a sample, matching
// TestScoutMusterOutTablesMatchBook1P79's own "no partial pins"
// convention.
func TestCitizenMusterOutTablesMatchBook1P78(t *testing.T) {
	t.Parallel()

	wantMoney := [11]string{
		"Low Passage", "Low Passage", "Middle Passage", "High Passage",
		"Cr15,000", "StarPass", "Cr25,000", "Cr30,000", "Cr35,000", "Cr40,000", "Cr50,000",
	}
	if citizenMusterOutMoney != wantMoney {
		t.Errorf("citizenMusterOutMoney =\n%v\nwant\n%v", citizenMusterOutMoney, wantMoney)
	}

	wantBenefits := [11]string{
		"Str +1", "Str +1", "Wafer Jack", "C5 +1",
		"Str +1", "C2 +1", "C3 +1", "Int +1", "Soc +1", "TAS Fellow Membership", "Ship Share",
	}
	if citizenMusterOutBenefits != wantBenefits {
		t.Errorf("citizenMusterOutBenefits =\n%v\nwant\n%v", citizenMusterOutBenefits, wantBenefits)
	}
}

func TestResolveCitizenMusterOutRollCountMatchesTerms(t *testing.T) {
	t.Parallel()

	careers := []Career{
		{},
		{Terms: make([]Term, 3)},
		{Terms: make([]Term, 7)},
		{Terms: make([]Term, 14)},
	}

	r := dice.New(rand.NewPCG(1, 2))

	for _, career := range careers {
		out := ResolveCitizenMusterOut(r, career)
		if got, want := len(out.Money)+len(out.Benefits), len(career.Terms); got != want {
			t.Errorf("len(Money)+len(Benefits) = %d, want %d (len(Terms))", got, want)
		}
	}
}

// TestResolveCitizenMusterOutEntriesAreFromTables confirms every returned
// entry is one of the 11 literal table values, mirroring
// TestResolveScoutMusterOutEntriesAreFromTables's reachability-check
// style.
func TestResolveCitizenMusterOutEntriesAreFromTables(t *testing.T) {
	t.Parallel()

	career := Career{Terms: make([]Term, 14)}
	r := dice.New(rand.NewPCG(3, 4))

	for range 200 {
		out := ResolveCitizenMusterOut(r, career)

		for _, m := range out.Money {
			if !slices.Contains(citizenMusterOutMoney[:], m) {
				t.Fatalf("ResolveCitizenMusterOut granted unexpected Money entry %q", m)
			}
		}

		for _, b := range out.Benefits {
			if !slices.Contains(citizenMusterOutBenefits[:], b) {
				t.Fatalf("ResolveCitizenMusterOut granted unexpected Benefits entry %q", b)
			}
		}
	}
}

// TestResolveCitizenMusterOutBothColumnsReachable guards the Uniform(2)
// column split, mirroring rollDeepSpaceBonus's own
// never-fired-in-N-trials failure-mode test.
func TestResolveCitizenMusterOutBothColumnsReachable(t *testing.T) {
	t.Parallel()

	career := Career{Terms: make([]Term, 10)}
	r := dice.New(rand.NewPCG(5, 6))

	var sawMoney, sawBenefits bool

	for range 50 {
		out := ResolveCitizenMusterOut(r, career)
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

func TestResolveCitizenMusterOutDeterminism(t *testing.T) {
	t.Parallel()

	career := Career{Terms: make([]Term, 5)}

	r1 := dice.New(rand.NewPCG(9, 10))
	r2 := dice.New(rand.NewPCG(9, 10))

	out1 := ResolveCitizenMusterOut(r1, career)
	out2 := ResolveCitizenMusterOut(r2, career)

	if !slices.Equal(out1.Money, out2.Money) {
		t.Fatalf("identical seeds produced different Money: %v vs %v", out1.Money, out2.Money)
	}

	if !slices.Equal(out1.Benefits, out2.Benefits) {
		t.Fatalf("identical seeds produced different Benefits: %v vs %v", out1.Benefits, out2.Benefits)
	}
}
