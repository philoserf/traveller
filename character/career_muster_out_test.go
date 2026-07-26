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

// TestRerollDuplicateBenefit pins Book 1 p.69's duplicate rule:
//
//	"A result that duplicates a previous (unwanted or unusable) benefit
//	may be rerolled until a different benefit is received, for example:
//	Wafer Jack, TAS Member, Knighthood."
func TestRerollDuplicateBenefit(t *testing.T) {
	t.Parallel()

	// A table with one unique benefit and one that stacks, so a reroll
	// always has somewhere to land.
	table := []string{"Knighthood", "Ship Share"}

	t.Run("a first Knighthood is kept", func(t *testing.T) {
		t.Parallel()

		got := rerollDuplicateBenefit(dice.New(rand.NewPCG(1, 1)), nil, 0, table, "Knighthood")
		if got != "Knighthood" {
			t.Errorf("got %q, want it kept — nothing duplicates it yet", got)
		}
	})

	t.Run("a second Knighthood is rerolled away", func(t *testing.T) {
		t.Parallel()

		got := rerollDuplicateBenefit(
			dice.New(rand.NewPCG(1, 1)), []string{"Knighthood"}, 0, table, "Knighthood")
		if got == "Knighthood" {
			t.Error("got Knighthood again, want it rerolled to a different benefit")
		}
	})

	t.Run("benefits that stack are never rerolled", func(t *testing.T) {
		t.Parallel()

		// Ship Shares accumulate toward ownership, so a repeat is wanted.
		got := rerollDuplicateBenefit(
			dice.New(rand.NewPCG(1, 1)), []string{"Ship Share", "Ship Share"}, 0, table, "Ship Share")
		if got != "Ship Share" {
			t.Errorf("got %q, want Ship Share kept — repeats of it are useful", got)
		}
	})

	t.Run("an unreachable alternative terminates instead of looping", func(t *testing.T) {
		t.Parallel()

		// The whole table is one unique benefit, which is what a
		// saturating DM effectively produces: p.68 clamps any roll past
		// the last row to that row. "Until a different benefit" can never
		// be satisfied, so the duplicate is kept rather than hanging.
		only := []string{"Knighthood"}

		got := rerollDuplicateBenefit(
			dice.New(rand.NewPCG(1, 1)), []string{"Knighthood"}, 0, only, "Knighthood")
		if got != "Knighthood" {
			t.Errorf("got %q, want the duplicate kept once no alternative exists", got)
		}
	})
}

// TestUniqueMusterOutBenefitsMatchTheBooksExamples guards the set itself.
// p.69 names exactly three, and the distinction is load-bearing: marking
// a stacking benefit unique would silently reroll away Ship Shares and
// characteristic awards a character is entitled to keep.
func TestUniqueMusterOutBenefitsMatchTheBooksExamples(t *testing.T) {
	t.Parallel()

	for _, unique := range []string{"Wafer Jack", "TAS Fellow Membership", "Knighthood"} {
		if !uniqueMusterOutBenefits[unique] {
			t.Errorf("%q is one of p.69's own examples but is not treated as unique", unique)
		}
	}

	for _, stacks := range []string{
		"Ship Share", "Fame +2", "Str +1", "C2 +1", "Forbidden Knowledge", "Life Insurance",
	} {
		if uniqueMusterOutBenefits[stacks] {
			t.Errorf("%q is treated as unique, but repeats of it are useful and p.69 does not name it", stacks)
		}
	}
}

// TestEntitlementMultiple pins Book 1 p.68's stacking rule: "Each
// doubling is of the original Pension: the first x2 doubles the Pension,
// the second x2 triples the pension, the third x2 quadruples the
// original Pension." Additive multiples of the base, not compounding —
// three awards pay four times, not eight.
func TestEntitlementMultiple(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		money []string
		want  int
	}{
		{"none", []string{"Cr30,000"}, 1},
		{"one doubles", []string{"Pension x2"}, 2},
		{"two triples", []string{"Pension x2", "Cr30,000", "Pension x2"}, 3},
		{"three quadruples", []string{"Pension x2", "Pension x2", "Pension x2"}, 4},
		{"the other token is ignored", []string{"Retirement x2"}, 1},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := entitlementMultiple(c.money, pensionDoubling); got != c.want {
				t.Errorf("entitlementMultiple(%v) = %d, want %d", c.money, got, c.want)
			}
		})
	}
}

// TestApplyRetirementPay pins p.70's Armed Forces Retirement Pay: at
// least 4 terms of active duty, Cr2,000 per term enlisted or Cr3,000 as
// an Officer, scaled by any "Retirement x2" awards.
func TestApplyRetirementPay(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		terms     int
		isOfficer bool
		money     []string
		want      int
	}{
		{"three terms is short of the minimum", 3, false, nil, 0},
		{"four terms enlisted", 4, false, nil, 4 * enlistedRetirementRate},
		{"four terms as an Officer pays the higher rate", 4, true, nil, 4 * officerRetirementRate},
		{"ten terms enlisted", 10, false, nil, 10 * enlistedRetirementRate},
		{
			"a Retirement x2 award doubles it",
			4, false,
			[]string{"Retirement x2"},
			2 * 4 * enlistedRetirementRate,
		},
		{
			"two of them triple it",
			4, true,
			[]string{"Retirement x2", "Retirement x2"},
			3 * 4 * officerRetirementRate,
		},
		{
			"a Pension x2 award does not touch it",
			4, false,
			[]string{"Pension x2"},
			4 * enlistedRetirementRate,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			out := MusteringOut{Money: c.money}
			applyRetirementPay(&out, Career{Terms: make([]Term, c.terms)}, c.isOfficer)

			if out.RetirementPay != c.want {
				t.Errorf("RetirementPay = %d, want %d", out.RetirementPay, c.want)
			}
		})
	}
}
