package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// TestLandGrantIncomeMatchesTheBooksWorkedExample is the strongest check
// available on the income formula, because Book 1 p.88 prints both
// figures for a fully specified case:
//
//	"recently knighted Sir Richard of Hefry (Trade Classifications Ni Va)
//	has a Land Grant of one Terrain Hex on Hefry producing an income of
//	Cr20,000 annually, and a companion Land Grant on a minor world (no
//	Trade Classifications) elsewhere in the system producing Cr5,000
//	annually."
//
// A Knight's row is MW 1 / other 1 and Hefry has two TCs, so the two
// halves must come out at exactly Cr20,000 and Cr5,000. This is what
// rules out the plausible misreading that the Cr10,000-per-TC rate
// applies once per grant rather than once per hex.
func TestLandGrantIncomeMatchesTheBooksWorkedExample(t *testing.T) {
	t.Parallel()

	knight, ok := nobleRankForSoc(11)
	if !ok {
		t.Fatal("Soc B confers no rank")
	}

	if knight.Title != "Knight" {
		t.Fatalf("Soc B = %q, want Knight", knight.Title)
	}

	if knight.MainworldHexes != 1 || knight.OtherHexes != 1 {
		t.Fatalf("Knight hexes = MW %d / other %d, want 1 / 1",
			knight.MainworldHexes, knight.OtherHexes)
	}

	// The mainworld hex alone, against Hefry's own Ni Va.
	if got := landGrantAnnualIncome(1, 0, 2); got != 20_000 {
		t.Errorf("one hex on a 2-TC world = Cr%d, want Cr20,000", got)
	}

	// The companion hex alone, on a world with no TCs.
	if got := landGrantAnnualIncome(0, 1, 0); got != 5_000 {
		t.Errorf("one companion hex = Cr%d, want Cr5,000", got)
	}

	if got := landGrantAnnualIncome(1, 1, 2); got != 25_000 {
		t.Errorf("Sir Richard's whole grant = Cr%d, want Cr25,000", got)
	}
}

// TestLandGrantIncomeUsesTheNoTCFloorNotZero pins the branch p.68 states
// separately — "A Land Grant on a World with no TCs generates Cr5,000
// per year" — which a naive tradeCodes*Cr10,000 would report as nothing
// at all.
func TestLandGrantIncomeUsesTheNoTCFloorNotZero(t *testing.T) {
	t.Parallel()

	if got := landGrantHexIncome(0); got != landGrantIncomeNoTradeCodes {
		t.Errorf("landGrantHexIncome(0) = %d, want %d", got, landGrantIncomeNoTradeCodes)
	}

	// p.68's own example: "a world classified as Hi In Va with three TCs
	// provides an income of Cr30,000 per year".
	if got := landGrantHexIncome(3); got != 30_000 {
		t.Errorf("landGrantHexIncome(3) = %d, want Cr30,000", got)
	}

	// The floor must not exceed the one-TC rate, or a worthless world
	// would out-earn a classified one.
	if landGrantIncomeNoTradeCodes > landGrantIncomePerTradeCode {
		t.Error("the no-TC floor pays better than a real Trade Classification")
	}
}

// TestNobleRanksMatchThePrintedTable pins the transcription itself. The
// p.88 rows were recovered by character offset rather than read visually
// — the page interleaves unrelated body text into every row — so the
// values most likely to be wrong are asserted directly rather than only
// through behavior.
func TestNobleRanksMatchThePrintedTable(t *testing.T) {
	t.Parallel()

	want := []struct {
		code      string
		soc       ehex.Value
		title     string
		mw, other int
	}{
		{"A", 10, "Gentleman", 0, 1},
		{"B", 11, "Knight", 1, 1},
		{"c", 12, "Baronet", 2, 2},
		{"C", 12, "Baron", 4, 4},
		{"D", 13, "Marquis", 8, 8},
		{"e", 14, "Viscount", 16, 16},
		{"E", 14, "Count", 32, 32},
		{"f", 15, "Duke", 64, 64},
		{"F", 15, "Duke", 128, 128},
		{"G", 16, "Archduke", 256, 256},
		{"H", 17, "Imperial Family", 256, 256},
		{"H", 17, "Emperor", 256, 256},
	}

	if len(nobleRanks) != len(want) {
		t.Fatalf("nobleRanks has %d rows, want %d", len(nobleRanks), len(want))
	}

	for i, w := range want {
		got := nobleRanks[i]
		if got.Code != w.code || got.Soc != w.soc || got.Title != w.title {
			t.Errorf("row %d = %s/%d/%q, want %s/%d/%q",
				i, got.Code, got.Soc, got.Title, w.code, w.soc, w.title)
		}

		if got.MainworldHexes != w.mw || got.OtherHexes != w.other {
			t.Errorf("row %d (%s) hexes = %d/%d, want %d/%d",
				i, w.title, got.MainworldHexes, got.OtherHexes, w.mw, w.other)
		}
	}
}

// TestThreeRanksSharePairsOfSocValues is the property that makes
// elevation a table walk rather than a Soc increment. If these ever
// stopped colliding, nobleRankAfterElevation's whole reason for existing
// would be gone — and, more dangerously, the reverse: silently
// "fixing" the collision would inflate every elevated Noble's Soc.
func TestThreeRanksSharePairsOfSocValues(t *testing.T) {
	t.Parallel()

	bySoc := map[ehex.Value][]string{}
	for _, rank := range nobleRanks {
		bySoc[rank.Soc] = append(bySoc[rank.Soc], rank.Title)
	}

	for soc, titles := range map[ehex.Value]int{12: 2, 14: 2, 15: 2} {
		if got := len(bySoc[soc]); got != titles {
			t.Errorf("Soc %d has %d ranks (%v), want %d", soc, got, bySoc[soc], titles)
		}
	}
}

// TestElevationWalksTheLadderWithoutSkippingSameSocRanks is Book 1
// p.65's "Elevated to the next higher Noble rank and its associated
// increase in Social Standing (if any)". The "(if any)" is the whole
// point: Baronet to Baron is a real elevation that raises no
// characteristic, and an implementation that incremented Soc instead
// would skip Baron, Count and Greater Duke entirely.
func TestElevationWalksTheLadderWithoutSkippingSameSocRanks(t *testing.T) {
	t.Parallel()

	start, ok := nobleRankIndexForSoc(11)
	if !ok {
		t.Fatal("Soc B confers no rank")
	}

	var (
		titles   []string
		socRoses []bool
	)

	// Walked by index rather than by comparing titles: two adjacent rows
	// legitimately share the title "Duke", so a title comparison cannot
	// tell a real step from the top of the ladder.
	//
	// Stops two rows from the end, not one: the last row is Emperor,
	// which nobody is elevated into, so Imperial Family is the ladder's
	// real top.
	for index := start; index < len(nobleRanks)-2; index++ {
		next, rose := nobleRankAfterElevation(index)
		titles = append(titles, nobleRanks[next].Title)
		socRoses = append(socRoses, rose)

		if next != index+1 {
			t.Errorf("elevation from row %d landed on %d, want %d", index, next, index+1)
		}

		if rose != (nobleRanks[next].Soc > nobleRanks[index].Soc) {
			t.Errorf("elevation from %s to %s reported socRose=%v",
				nobleRanks[index].Title, nobleRanks[next].Title, rose)
		}
	}

	wantTitles := []string{
		"Baronet",
		"Baron",
		"Marquis",
		"Viscount",
		"Count",
		"Duke",
		"Duke",
		"Archduke",
		"Imperial Family",
	}
	for i, want := range wantTitles {
		if i >= len(titles) || titles[i] != want {
			t.Fatalf("elevation %d = %v, want %q (full path %v)", i+1, titles, want, wantTitles)
		}
	}

	// Baronet->Baron, Viscount->Count and Lesser->Greater Duke all stay
	// at the same Soc, so exactly three of the nine steps raise nothing.
	rose := 0

	for _, r := range socRoses {
		if r {
			rose++
		}
	}

	if rose != 6 {
		t.Errorf("%d of %d elevations raised Soc, want 6 (three pairs share a Soc)", rose, len(socRoses))
	}

	// Elevation from the top must be a no-op, not a walk into Emperor.
	top := len(nobleRanks) - 2
	if next, rose := nobleRankAfterElevation(top); next != top || rose {
		t.Errorf("elevation from %s returned row %d (rose=%v), want it to stay put",
			nobleRanks[top].Title, next, rose)
	}
}

// TestNobleTitleForSocBoundaries covers the ends: below Soc A there is
// no noble rank at all, and past the table's top the lookup clamps
// rather than failing — Mustering Out's own Soc awards and Knighthood
// can carry a character past any row.
func TestNobleTitleForSocBoundaries(t *testing.T) {
	t.Parallel()

	for raw := range 10 {
		soc := ehex.Value(raw)
		if got := NobleTitleForSoc(soc); got != "" {
			t.Errorf("NobleTitleForSoc(%d) = %q, want no title below Soc A", soc, got)
		}
	}

	cases := map[ehex.Value]string{10: "Gentleman", 11: "Knight", 12: "Baronet", 13: "Marquis", 16: "Archduke"}
	for soc, want := range cases {
		if got := NobleTitleForSoc(soc); got != want {
			t.Errorf("NobleTitleForSoc(%d) = %q, want %q", soc, got, want)
		}
	}

	// Emperor is in the table but is never what a Soc lookup returns.
	for soc := ehex.Value(17); soc <= ehex.Max; soc++ {
		if got := NobleTitleForSoc(soc); got == "Emperor" {
			t.Errorf("NobleTitleForSoc(%d) = Emperor, which no character is elevated into", soc)
		}
	}
}

// TestNobleTitleIsDerivedFromSocNotFromCareer confirms the accessor
// applies to characters who were never Nobles — p.68's Knighthood raises
// Soc to B from any career's Mustering Out, and that is a Knight.
func TestNobleTitleIsDerivedFromSocNotFromCareer(t *testing.T) {
	t.Parallel()

	var c Character

	c.UPP.Characteristics[C6] = 11
	if got := c.NobleTitle(); got != "Knight" {
		t.Errorf("NobleTitle() = %q at Soc B, want Knight", got)
	}

	c.UPP.Characteristics[C6] = 9
	if got := c.NobleTitle(); got != "" {
		t.Errorf("NobleTitle() = %q at Soc 9, want no title", got)
	}
}

// TestScoutDiscoveryGrantsOnePerDiscovery is p.79's own Reward Success —
// the Scout "receives a Land Grant" — counted against the Discoveries
// that earned them, and against nothing else.
func TestScoutDiscoveryGrantsOnePerDiscovery(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(5, 5))

	for _, discoveries := range []int{0, 1, 2, 5} {
		career := Career{Name: "Scout"}
		for range discoveries {
			career.Terms = append(career.Terms, Term{RewardResult: "Discovery"})
		}

		career.Terms = append(career.Terms, Term{RewardResult: "None"}, Term{RewardResult: ""})

		grants := scoutDiscoveryLandGrants(r, career)
		if len(grants) != discoveries {
			t.Errorf("%d Discoveries produced %d grants, want %d", discoveries, len(grants), discoveries)
		}

		for _, grant := range grants {
			if grant.Source != "Discovery" {
				t.Errorf("grant source = %q, want Discovery", grant.Source)
			}

			if grant.AnnualIncome != landGrantAnnualIncome(
				grant.MainworldHexes,
				grant.OtherHexes,
				len(grant.TradeCodes),
			) {
				t.Errorf("grant income %d disagrees with its own hexes and TCs", grant.AnnualIncome)
			}
		}
	}
}

// TestLandGrantIncomeIsCumulative is p.88's "Land Grants Are Cumulative.
// Each title confers its own Land Grant: a Knight raised to Baronet
// receives it in addition to his Knighthood".
func TestLandGrantIncomeIsCumulative(t *testing.T) {
	t.Parallel()

	grants := []LandGrant{
		{AnnualIncome: 25_000},
		{AnnualIncome: 5_000},
		{AnnualIncome: 65_000},
	}

	c := Character{LandGrants: grants}
	if got := c.LandGrantIncome(); got != 95_000 {
		t.Errorf("LandGrantIncome() = %d, want 95,000", got)
	}

	if got := (Character{}).LandGrantIncome(); got != 0 {
		t.Errorf("a character with no grants has income %d, want 0", got)
	}
}

// TestGeneratedNobleGainsTitleAndGrantsTogether is the end-to-end check
// that the two halves of this change agree: a Noble who was elevated
// holds a title above Knight, and holds a Land Grant for each Soc
// increase that got them there.
func TestGeneratedNobleGainsTitleAndGrantsTogether(t *testing.T) {
	t.Parallel()

	sawElevated := false

	for seed := uint64(1); seed <= 400; seed++ {
		c, ok := GenerateNobleCharacter(dice.New(rand.NewPCG(seed, seed)))
		if !ok || len(c.Careers) == 0 {
			continue
		}

		elevations := 0

		for _, term := range c.Careers[0].Terms {
			if term.Elevated {
				elevations++
			}
		}

		if elevations == 0 {
			continue
		}

		sawElevated = true

		// Every grant a Noble holds came from a Soc increase, and Soc can
		// only rise once per elevation — so grants can never outnumber
		// elevations, though they can fall short of them when an
		// elevation crosses a same-Soc pair.
		if len(c.LandGrants) > elevations {
			t.Errorf("seed %d: %d Land Grants from %d elevations", seed, len(c.LandGrants), elevations)
		}

		if c.NobleTitle() == "" {
			t.Errorf("seed %d: an elevated Noble has no title (Soc %v)", seed, c.UPP.Characteristics[C6])
		}

		for _, grant := range c.LandGrants {
			if grant.Source == "Discovery" {
				t.Errorf("seed %d: a Noble holds a Discovery grant", seed)
			}
		}
	}

	if !sawElevated {
		t.Fatal("no elevated Noble in 400 seeds — the test never exercised its own subject")
	}
}
