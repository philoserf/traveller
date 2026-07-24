package character

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
)

func TestScoutMusterOutRow(t *testing.T) {
	t.Parallel()

	cases := []struct {
		roll int
		want int
	}{
		{1, 0},
		{6, 5},
		{12, 11},
		{13, 11},
		{20, 11},
		{99, 11},
	}

	for _, c := range cases {
		if got := scoutMusterOutRow(c.roll); got != c.want {
			t.Errorf("scoutMusterOutRow(%d) = %d, want %d", c.roll, got, c.want)
		}
	}
}

// TestScoutMusterOutTablesMatchBook1P79 pins both 12-entry Scout Mustering
// Out columns in full, transcribed directly from the page image — not just
// a sample, matching TestHomeworldSkillByTradeCodeMapping's own
// "no partial pins" convention.
func TestScoutMusterOutTablesMatchBook1P79(t *testing.T) {
	t.Parallel()

	wantMoney := [12]string{
		"Low Passage", "Middle Passage", "High Passage", "StarPass",
		"Cr30,000", "Cr40,000", "Cr50,000", "Cr60,000",
		"Cr60,000", "Cr60,000", "Cr70,000", "Cr80,000",
	}
	if scoutMusterOutMoney != wantMoney {
		t.Errorf("scoutMusterOutMoney =\n%v\nwant\n%v", scoutMusterOutMoney, wantMoney)
	}

	wantBenefits := [12]string{
		"Ship Share", "Forbidden Knowledge", "Wafer Jack", "C5 +1",
		"Str +1", "C2 +1", "C3 +1", "Ship Share",
		"Life Insurance", "TAS Fellow Membership", "Fame +2", "Knighthood",
	}
	if scoutMusterOutBenefits != wantBenefits {
		t.Errorf("scoutMusterOutBenefits =\n%v\nwant\n%v", scoutMusterOutBenefits, wantBenefits)
	}
}

func TestScoutMusterOutRollCount(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		career Career
		want   int
	}{
		{"never qualified", Career{}, 0},
		{"3 terms, Unharmed ending", Career{Terms: make([]Term, 3)}, 3},
		{
			"5 terms, Wounded ending",
			Career{Terms: append(make([]Term, 4), Term{RiskResult: Wounded})},
			5,
		},
		{
			"2 terms, Disabled ending doubles the roll count",
			Career{Terms: append(make([]Term, 1), Term{RiskResult: Disabled})},
			4,
		},
		{
			"7 terms, Dead ending",
			Career{Terms: append(make([]Term, 6), Term{RiskResult: Dead})},
			0,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := scoutMusterOutRollCount(c.career); got != c.want {
				t.Errorf("scoutMusterOutRollCount(%s) = %d, want %d", c.name, got, c.want)
			}
		})
	}
}

func TestRollScoutMusterOutRow(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(5, 6))

	for _, dm := range []int{0, 14} {
		for range 500 {
			row := rollScoutMusterOutRow(r, dm)
			if row < 0 || row > 11 {
				t.Fatalf("rollScoutMusterOutRow(dm=%d) = %d, want in [0,11]", dm, row)
			}
		}
	}
}

func TestResolveScoutMusterOutRollCountMatchesTerms(t *testing.T) {
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
		want := scoutMusterOutRollCount(career)

		out := ResolveScoutMusterOut(r, career)
		if got := len(out.Money) + len(out.Benefits); got != want {
			t.Errorf("len(Money)+len(Benefits) = %d, want %d (scoutMusterOutRollCount)", got, want)
		}
	}
}

// TestResolveScoutMusterOutEntriesAreFromTables confirms every returned
// entry is one of the 12 literal table values, mirroring
// TestGenerateHomeworldSkillsOnlyGrantsReachableSkills's reachability-check
// style.
func TestResolveScoutMusterOutEntriesAreFromTables(t *testing.T) {
	t.Parallel()

	career := Career{Terms: make([]Term, 14)}
	r := dice.New(rand.NewPCG(15, 16))

	for range 200 {
		out := ResolveScoutMusterOut(r, career)

		for _, m := range out.Money {
			if !slices.Contains(scoutMusterOutMoney[:], m) {
				t.Fatalf("ResolveScoutMusterOut granted unexpected Money entry %q", m)
			}
		}

		for _, b := range out.Benefits {
			if !slices.Contains(scoutMusterOutBenefits[:], b) {
				t.Fatalf("ResolveScoutMusterOut granted unexpected Benefits entry %q", b)
			}
		}
	}
}

// TestResolveScoutMusterOutBothColumnsReachable guards the Uniform(2)
// column split, mirroring rollDeepSpaceBonus's own
// never-fired-in-N-trials failure-mode test.
func TestResolveScoutMusterOutBothColumnsReachable(t *testing.T) {
	t.Parallel()

	career := Career{Terms: make([]Term, 10)}
	r := dice.New(rand.NewPCG(51, 53))

	var sawMoney, sawBenefits bool

	for range 50 {
		out := ResolveScoutMusterOut(r, career)
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

func TestResolveScoutMusterOutZeroForDeadEnding(t *testing.T) {
	t.Parallel()

	career := Career{Terms: append(make([]Term, 3), Term{RiskResult: Dead})}

	for _, seed := range []uint64{1, 2, 3} {
		r := dice.New(rand.NewPCG(seed, seed))

		out := ResolveScoutMusterOut(r, career)
		if len(out.Money) != 0 || len(out.Benefits) != 0 {
			t.Errorf("seed %d: ResolveScoutMusterOut(dead career) = %+v, want zero-value", seed, out)
		}
	}
}

func TestResolveScoutMusterOutZeroForNeverQualified(t *testing.T) {
	t.Parallel()

	career := Career{}

	for _, seed := range []uint64{1, 2, 3} {
		r := dice.New(rand.NewPCG(seed, seed))

		out := ResolveScoutMusterOut(r, career)
		if len(out.Money) != 0 || len(out.Benefits) != 0 {
			t.Errorf("seed %d: ResolveScoutMusterOut(never-qualified career) = %+v, want zero-value", seed, out)
		}
	}
}

func TestResolveScoutMusterOutDeterminism(t *testing.T) {
	t.Parallel()

	career := Career{Terms: append(make([]Term, 3), Term{RiskResult: Wounded})}

	r1 := dice.New(rand.NewPCG(77, 78))
	r2 := dice.New(rand.NewPCG(77, 78))

	out1 := ResolveScoutMusterOut(r1, career)
	out2 := ResolveScoutMusterOut(r2, career)

	if !slices.Equal(out1.Money, out2.Money) {
		t.Fatalf("identical seeds produced different Money: %v vs %v", out1.Money, out2.Money)
	}

	if !slices.Equal(out1.Benefits, out2.Benefits) {
		t.Fatalf("identical seeds produced different Benefits: %v vs %v", out1.Benefits, out2.Benefits)
	}
}
