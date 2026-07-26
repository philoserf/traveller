package character

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
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

// TestMusterOutRollCount pins Book 1 p.68's whole roll budget: "one
// Mustering Out roll for each term served in Career Resolution. He is
// allowed one additional roll per Commendation, MCG, or SEH. He is
// allowed one additional roll if Fame 19+."
//
// The medal cases matter most. p.68 names MCG and SEH and stops there;
// XS and MCUF are far commoner (an XS lands on every surviving Risk
// roll), so counting them would roughly double a long Armed Forces
// career's benefits.
func TestMusterOutRollCount(t *testing.T) {
	t.Parallel()

	medalTerm := func(medals ...string) Term { return Term{Medals: medals} }

	cases := []struct {
		name   string
		career Career
		fame   int
		want   int
	}{
		{"never qualified", Career{}, 0, 0},
		{"3 terms, Unharmed ending", Career{Terms: make([]Term, 3)}, 0, 3},
		{
			"5 terms, Wounded ending",
			Career{Terms: append(make([]Term, 4), Term{RiskResult: Wounded})},
			0, 5,
		},
		{
			"7 terms, Dead ending",
			Career{Terms: append(make([]Term, 6), Term{RiskResult: Dead})},
			0, 0,
		},
		{
			"one roll per Commendation on top of terms",
			Career{Terms: []Term{{RewardResult: "Noble Commendation-3"}, {RewardResult: "None"}}},
			0, 3,
		},
		{"MCG earns an extra roll", Career{Terms: []Term{medalTerm("MCG")}}, 0, 2},
		{"SEH earns an extra roll", Career{Terms: []Term{medalTerm("SEH")}}, 0, 2},
		{
			"XS and MCUF do not — p.68 names only MCG and SEH",
			Career{Terms: []Term{medalTerm("XS", "MCUF"), medalTerm("XS")}},
			0, 2,
		},
		{
			"several qualifying medals each earn one",
			Career{Terms: []Term{medalTerm("XS", "MCG"), medalTerm("SEH", "MCUF")}},
			0, 4,
		},
		{"Fame 18 earns nothing", Career{Terms: make([]Term, 2)}, 18, 2},
		{"Fame 19 earns one extra roll", Career{Terms: make([]Term, 2)}, 19, 3},
		{"Fame well past 19 still earns exactly one", Career{Terms: make([]Term, 2)}, 40, 3},
		{
			"Disability doubles the finished total, extras included",
			Career{Terms: []Term{medalTerm("MCG"), {RiskResult: Disabled}}},
			19, 8, // (2 terms + 1 medal + 1 fame) * 2
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := musterOutRollCount(c.career, c.fame); got != c.want {
				t.Errorf("musterOutRollCount(%s) = %d, want %d", c.name, got, c.want)
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
		want := musterOutRollCount(career, scoutDiscoveryFame(career))

		out := ResolveScoutMusterOut(r, career)
		if got := len(out.Money) + len(out.Benefits); got != want {
			t.Errorf("len(Money)+len(Benefits) = %d, want %d (musterOutRollCount)", got, want)
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

// TestResolveScoutMusterOutAccumulatesFameIntoDM is the regression test
// for Book 1 p.79's own "DM +Terms +Fame/2": seed 49 with this fixture
// was found by direct search to start with an unsaturated DM (6 terms
// + 4 Discovery Fame / 2 = 8, below the table's own 12 rows) and then
// grant "Fame +2" twice at Benefits indices 0 and 1,
// so every roll after each one must use an elevated dm
// (terms+fame/2, not a static terms) — a full pin of the exact resulting
// sequence, not just "it doesn't crash," so a regression that reverts to
// the old static-dm formula (which would roll differently once fame > 0)
// changes this test's expected output.
func TestResolveScoutMusterOutAccumulatesFameIntoDM(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{9, 9, 9, 12, 12, 0}}
	r := dice.New(rand.NewPCG(2579, 2579))

	career, _ := ResolveScoutCareer(r, upp)
	if len(career.Terms) != 6 {
		t.Fatalf("seed 2579: len(Terms) = %d, want 6 (fixture assumption broke)", len(career.Terms))
	}

	out := ResolveScoutMusterOut(r, career)

	wantBenefits := []string{
		"Fame +2", "Fame +2", "Knighthood",
		"Knighthood", "Knighthood", "Knighthood",
	}
	if !slices.Equal(out.Benefits, wantBenefits) {
		t.Fatalf("Benefits = %v, want %v", out.Benefits, wantBenefits)
	}

	wantMoney := []string{"Cr80,000", "Cr70,000", "Cr80,000", "Cr80,000", "Cr80,000", "Cr80,000"}
	if !slices.Equal(out.Money, wantMoney) {
		t.Fatalf("Money = %v, want %v", out.Money, wantMoney)
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

// TestResolveScoutMusterOutSeedsFameDMFromDiscoveries is #58's
// regression, stated as the rule rather than as a pinned sequence: Book
// 1 p.79's "DM +Terms +Fame/2" counts the Fame the character already
// earned, and for a Scout that is Discovery Fame (p.91, "Discoveries
// x4"). It used to start at zero, so the DM ignored discoveries the
// career had demonstrably made.
//
// Checked by construction rather than by seed-hunting: two careers with
// identical term counts, one with discoveries and one without, rolled
// from the same seed. Identical terms mean an identical +Terms DM, so
// any divergence in the rolled rows can only come from the Fame half.
func TestResolveScoutMusterOutSeedsFameDMFromDiscoveries(t *testing.T) {
	t.Parallel()

	const terms = 4

	withDiscoveries := Career{Name: "Scout", Terms: make([]Term, terms)}
	for i := range withDiscoveries.Terms {
		withDiscoveries.Terms[i].RewardResult = "Discovery"
	}

	without := Career{Name: "Scout", Terms: make([]Term, terms)}
	for i := range without.Terms {
		without.Terms[i].RewardResult = "None"
	}

	if scoutDiscoveryFame(without) != 0 {
		t.Fatal("control career has Discovery Fame, want none")
	}

	if got := scoutDiscoveryFame(withDiscoveries); got != terms*4 {
		t.Fatalf("scoutDiscoveryFame = %d, want %d", got, terms*4)
	}

	rich := ResolveScoutMusterOut(dice.New(rand.NewPCG(7, 7)), withDiscoveries)
	plain := ResolveScoutMusterOut(dice.New(rand.NewPCG(7, 7)), without)

	// Same seed, same term count — only the Fame half of the DM differs,
	// so identical output would mean discoveries never reached it.
	if slices.Equal(rich.Benefits, plain.Benefits) && slices.Equal(rich.Money, plain.Money) {
		t.Errorf("a Scout with %d Discoveries mustered out identically to one with none "+
			"(benefits %v, money %v) — Discovery Fame is not reaching the DM",
			terms, rich.Benefits, rich.Money)
	}
}
