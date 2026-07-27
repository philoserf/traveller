package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/ehex"

	"github.com/philoserf/traveller/dice"
)

// TestQREBSPointScaleMatchesTheBook pins Book 1's own allocation scale —
// "for the ranges -5 to +5, -5 = 1 point; +5 = 11 points" — and the
// consequence that makes the rest of the rule cohere: five dimensions at
// maximum cost exactly 55, which is the Perfect Masterpiece threshold.
//
// That equality is the strongest available check that the offset is
// right. An offset one either way makes the maximum 50 or 60 and severs
// the connection to "A Perfect Masterpiece has 55 or more Master
// Points", which is what the excess-allocation sentence exists to handle.
func TestQREBSPointScaleMatchesTheBook(t *testing.T) {
	t.Parallel()

	if cost := qrebsMin + qrebsPointOffset; cost != 1 {
		t.Errorf("a dimension at %d costs %d points, want 1", qrebsMin, cost)
	}

	if cost := qrebsMax + qrebsPointOffset; cost != 11 {
		t.Errorf("a dimension at %d costs %d points, want 11", qrebsMax, cost)
	}

	if qrebsMinTotal != 5 {
		t.Errorf("all dimensions at minimum cost %d, want 5", qrebsMinTotal)
	}

	if qrebsMaxTotal != craftsmanPerfectMasterPoints {
		t.Errorf("all dimensions at maximum cost %d, want %d (the Perfect Masterpiece threshold)",
			qrebsMaxTotal, craftsmanPerfectMasterPoints)
	}
}

// TestAllocateQREBSSpendsExactlyItsPoints is the allocation invariant:
// the five dimensions' costs must sum to the Master Points allocated,
// no more and no less, for every reachable point total.
func TestAllocateQREBSSpendsExactlyItsPoints(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(31, 31))

	for points := qrebsMinTotal; points <= qrebsMaxTotal; points++ {
		for range 40 {
			q := allocateQREBS(r, points)

			spent := 0
			for _, v := range q.Values() {
				spent += v + qrebsPointOffset
			}

			if spent != points {
				t.Fatalf("%d Master Points allocated as %+v, costing %d", points, q, spent)
			}
		}
	}
}

// TestAllocateQREBSStaysInRange keeps every dimension inside -5..+5 for
// any total that does not exceed the maximum. Only an over-maximum
// allocation may exceed +5, and only via the book's own excess rule.
func TestAllocateQREBSStaysInRange(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(37, 37))

	for points := qrebsMinTotal; points <= qrebsMaxTotal; points++ {
		for range 40 {
			for i, v := range allocateQREBS(r, points).Values() {
				if v < qrebsMin || v > qrebsMax {
					t.Fatalf("%d points: dimension %d = %d, outside %d..%d",
						points, i, v, qrebsMin, qrebsMax)
				}
			}
		}
	}
}

// TestAllocateQREBSAtMaximumSetsEveryDimensionToMax is the boundary the
// excess rule hinges on: at exactly 55 points there is one possible
// allocation, every dimension at +5.
func TestAllocateQREBSAtMaximumSetsEveryDimensionToMax(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(41, 41))

	for range 100 {
		for i, v := range allocateQREBS(r, qrebsMaxTotal).Values() {
			if v != qrebsMax {
				t.Fatalf("at %d points, dimension %d = %d, want %d", qrebsMaxTotal, i, v, qrebsMax)
			}
		}
	}
}

// TestAllocateQREBSSpreadsExcessEqually is Book 1's own "If all QREBS
// values are set at the Maximum, excess Master Points can be allocated
// equally in excess of +5" — every dimension above the maximum, and all
// of them equal.
func TestAllocateQREBSSpreadsExcessEqually(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(43, 43))

	for _, excess := range []int{5, 10, 25, 50} {
		q := allocateQREBS(r, qrebsMaxTotal+excess)
		values := q.Values()

		want := qrebsMax + excess/qrebsDimensions
		for i, v := range values {
			if v != want {
				t.Errorf("%d excess points: dimension %d = %d, want %d (spread equally)",
					excess, i, v, want)
			}
		}

		if values[0] <= qrebsMax {
			t.Errorf("%d excess points left every dimension at or below the maximum", excess)
		}
	}
}

// TestAllocateQREBSVariesBetweenMasterpieces confirms the allocation is
// actually resolved through the dice rather than being a fixed split —
// two Masterpieces of the same point total should not always agree.
func TestAllocateQREBSVariesBetweenMasterpieces(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(47, 47))

	first := allocateQREBS(r, 45)
	varied := false

	for range 50 {
		if allocateQREBS(r, 45) != first {
			varied = true

			break
		}
	}

	if !varied {
		t.Error("50 allocations of 45 points were all identical — the split is not being rolled")
	}
}

// TestMasterpieceValueMatchesThePrintedTable checks the value formula
// against Book 1 p.75's own printed Masterpiece Value table, which is
// what settles the book's own internal disagreement: p.75 says
// "Cr10,000 per Master Point over 40" and the QREBS chapter says "over
// 39". Only "over 40" reproduces these rows.
func TestMasterpieceValueMatchesThePrintedTable(t *testing.T) {
	t.Parallel()

	cases := map[int]int{
		45: 200_000,
		46: 210_000,
		47: 220_000,
		48: 230_000,
		49: 240_000,
		50: 250_000,
		51: 260_000,
		52: 270_000,
		53: 280_000,
		54: 290_000,
		// The doubling at the Perfect threshold: the plain formula gives
		// Cr300,000 here, and the table prints Cr600,000.
		55: 600_000,
		56: 620_000,
		57: 640_000,
		58: 660_000,
		59: 680_000,
		60: 700_000,
		61: 720_000,
		62: 740_000,
		63: 760_000,
	}

	for points, want := range cases {
		if got := craftsmanMasterpieceValue(points); got != want {
			t.Errorf("craftsmanMasterpieceValue(%d) = %d, want %d", points, got, want)
		}
	}

	// At the floor, the base value and nothing more — the rule pays only
	// "per Master Point over 40".
	if got := craftsmanMasterpieceValue(craftsmanMinMasterPoints); got != craftsmanBaseMasterpieceValue {
		t.Errorf("at the floor, value = %d, want %d", got, craftsmanBaseMasterpieceValue)
	}

	// Below the floor a Masterpiece still cannot sell for less than it
	// cost to make.
	if got := craftsmanMasterpieceValue(10); got != craftsmanBaseMasterpieceValue {
		t.Errorf("below the floor, value = %d, want %d (never less than cost)",
			got, craftsmanBaseMasterpieceValue)
	}
}

// TestVintageValueIsSimpleInterest pins Book 1's own "increases in value
// about 1% per year (simple interest)" — simple, so 100 years is +100%
// and not the ~170% compounding would give.
func TestVintageValueIsSimpleInterest(t *testing.T) {
	t.Parallel()

	m := Masterpiece{BaseValue: 200_000, CreatedAtAge: 30}

	cases := map[int]int{
		30:  200_000, // the year it was made
		31:  202_000, // +1%
		40:  220_000, // +10%
		80:  300_000, // +50%
		130: 400_000, // +100% — double, which compounding would overshoot
	}

	for age, want := range cases {
		if got := m.VintageValue(age); got != want {
			t.Errorf("VintageValue(age %d) = %d, want %d", age, got, want)
		}
	}

	// A value read before the Masterpiece existed must not depreciate it.
	for _, age := range []int{0, 10, 29} {
		if got := m.VintageValue(age); got != m.BaseValue {
			t.Errorf("VintageValue(age %d) = %d, want the base %d", age, got, m.BaseValue)
		}
	}
}

// TestVintageValueAtSaleAppliesFluxAroundTheVintageValue covers the
// half of the rule that is explicitly a sale-time roll: "subject to Flux
// (in percent) when sold". Flux is signed, so sales must land on both
// sides of the vintage value — a one-sided result would mean the sign is
// being dropped.
func TestVintageValueAtSaleAppliesFluxAroundTheVintageValue(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(53, 53))
	m := Masterpiece{BaseValue: 200_000, CreatedAtAge: 30}

	const ageNow = 80

	vintage := m.VintageValue(ageNow)

	above, below, equal := 0, 0, 0

	for range 2000 {
		switch got := m.VintageValueAtSale(r, ageNow); {
		case got > vintage:
			above++
		case got < vintage:
			below++
		default:
			equal++
		}

		// Flux is -5..+5, so a sale can never move more than 5% either way.
		if got := m.VintageValueAtSale(r, ageNow); got < vintage*95/100 || got > vintage*105/100 {
			t.Fatalf("sale value %d is outside +/-5%% of the vintage value %d", got, vintage)
		}
	}

	if above == 0 || below == 0 {
		t.Errorf("sales were one-sided (%d above, %d below, %d equal) — Flux's sign is being lost",
			above, below, equal)
	}
}

// TestVintageValueAtSaleNeverGoesNegative pins the clamp. Unreachable
// with real Flux against a positive value, but recorded so the guard
// cannot be removed as dead code.
func TestVintageValueAtSaleNeverGoesNegative(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(59, 59))

	for _, base := range []int{0, 1, 100} {
		m := Masterpiece{BaseValue: base, CreatedAtAge: 30}
		for range 500 {
			if got := m.VintageValueAtSale(r, 90); got < 0 {
				t.Fatalf("base %d sold for %d", base, got)
			}
		}
	}
}

// TestCraftsmanTermsRecordStructuredMasterpieces is the end-to-end
// check: a successful creation leaves a structured record whose QREBS
// spend, Perfect flag and value all agree with its Master Points, and a
// failed one leaves none — p.75's "the Piece (not Masterpiece) is flawed
// and worthless".
func TestCraftsmanTermsRecordStructuredMasterpieces(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(61, 61))
	upp := UPP{Characteristics: [6]ehex.Value{10, 10, 10, 10, 10, 10}}

	// Master Points are CC + Craftsman level + the five highest
	// qualifying skills (level 6+, excluding Language and Craftsman
	// itself). Ten plus six plus five sixes is 46 — above the 40 floor
	// that lets a creation be attempted at all, and short of the 55 that
	// would make every attempt Perfect.
	qualifying := []string{"Artist", "Sculptor", "Jeweler", "Author", "Chef"}

	held := make([]SkillLevel, 0, len(qualifying)+1)
	held = append(held, SkillLevel{Name: "Craftsman", Level: 6, Kind: Skill})

	for _, name := range qualifying {
		held = append(held, SkillLevel{Name: name, Level: 6, Kind: Skill})
	}

	if got := craftsmanMasterPoints(10, held); got < craftsmanMinMasterPoints {
		t.Fatalf("fixture gives %d Master Points, below the %d floor — no creation can be attempted",
			got, craftsmanMinMasterPoints)
	}

	created, failed := 0, 0

	for range 300 {
		term, _ := ResolveCraftsmanTerm(r, upp, C1, held, 30)

		if term.Masterpiece == nil {
			failed++

			if term.RewardResult != "None" {
				t.Errorf("no Masterpiece recorded but RewardResult = %q", term.RewardResult)
			}

			continue
		}

		created++
		m := term.Masterpiece

		spent := 0
		for _, v := range m.QREBS.Values() {
			spent += v + qrebsPointOffset
		}

		if spent != m.MasterPoints {
			t.Errorf("QREBS spend %d != Master Points %d", spent, m.MasterPoints)
		}

		if m.Perfect != (m.MasterPoints >= craftsmanPerfectMasterPoints) {
			t.Errorf("Perfect = %v at %d Master Points", m.Perfect, m.MasterPoints)
		}

		if m.Perfect != term.Perfect {
			t.Errorf("Masterpiece.Perfect = %v but Term.Perfect = %v", m.Perfect, term.Perfect)
		}

		if want := craftsmanMasterpieceValue(m.MasterPoints); m.BaseValue != want {
			t.Errorf("BaseValue = %d, want %d", m.BaseValue, want)
		}

		if m.CreatedAtAge != 30 {
			t.Errorf("CreatedAtAge = %d, want 30", m.CreatedAtAge)
		}
	}

	if created == 0 {
		t.Fatal("no Masterpiece was created in 300 terms — the test never exercised its subject")
	}

	if failed == 0 {
		t.Log("note: no failed creation in 300 terms, so the nil-record path went unchecked")
	}
}

// TestAMasterpieceIsActuallyReachable is #95 made executable.
//
// That issue was filed because QREBS allocation and Vintage appreciation
// were fully implemented, unit-tested, and had never once fired in
// generated output — no Craftsman in 6,000 chains reached the 40 Master
// Points p.75 requires, because no Craftsman ran at all. #110 fixed the
// cause (career chains were capping every non-final career at one term,
// so nobody arrived with the skills BeginCraftsman demands).
//
// A unit test on allocateQREBS cannot catch that class of defect: the
// code was correct and simply unreachable. This walks the whole path
// instead — generate characters until one creates a Masterpiece, then
// check the record it produced is complete.
//
// It is deliberately expensive and deliberately broad. A Masterpiece is
// rare by the rules' own design: a Craftsman must already hold
// Craftsman-1 and two skills at level 6+ merely to Begin, and then reach
// 40 Master Points from his Controlling Characteristic, his Craftsman
// skill and his best five qualifying skills.
func TestAMasterpieceIsActuallyReachable(t *testing.T) {
	t.Parallel()

	// Wide on purpose. A Masterpiece is rare by the rules' own design —
	// measured at roughly one per 10,000 chains, six in 60,000 with the
	// first at seed 6,694 — so a budget close to the rate will fail on
	// any dice-stream shift that moves the first hit, which is what
	// happened at 6,000 the first time this ran against a changed stream.
	// The same lesson as seedSearchLimit: clear the worst case, not the
	// typical one.
	const chainSeeds = 100_000

	holder, craftsmen, attempted := searchForAMasterpiece(chainSeeds)

	if craftsmen == 0 {
		t.Fatalf("no Craftsman served a term in %d chains — the career is unreachable again, which is #95's "+
			"original complaint and #110's own regression", chainSeeds)
	}

	if len(holder.Masterpieces) == 0 {
		t.Fatalf("%d Craftsmen served %d terms across %d chains without creating a Masterpiece; QREBS and "+
			"Vintage are unreachable in generated output again (#95). Expected roughly one per 10,000 chains",
			craftsmen, attempted, chainSeeds)
	}

	found := holder.Masterpieces[0]

	// The record has to be complete, not merely present — every field
	// below is one the issue said had never been exercised.
	if found.MasterPoints < craftsmanMinMasterPoints {
		t.Errorf("MasterPoints = %d, want at least %d (p.75's own gate)",
			found.MasterPoints, craftsmanMinMasterPoints)
	}

	// The allocation ran and spent the Master Points, rather than leaving
	// every dimension at its floor: qrebsMinTotal is what the five
	// minimums cost, so a Masterpiece past the 40-point gate must have
	// had points left to distribute.
	spent := 0

	for _, v := range found.QREBS.Values() {
		if v < qrebsMin {
			t.Errorf("QREBS %+v has a dimension below the %d floor", found.QREBS, qrebsMin)
		}

		spent += v + qrebsPointOffset
	}

	if spent <= qrebsMinTotal {
		t.Errorf("QREBS %+v spent %d points, want more than the %d the five minimums cost — "+
			"the allocation never ran", found.QREBS, spent, qrebsMinTotal)
	}

	if found.BaseValue <= 0 {
		t.Errorf("BaseValue = %d, want a sale price", found.BaseValue)
	}

	if found.CreatedAtAge <= 0 {
		t.Error("CreatedAtAge is unset, so Vintage appreciation has no time reference")
	}

	// Vintage appreciates with age, so a Masterpiece made before the
	// character finished aging must now be worth more than it was.
	if holder.Age > found.CreatedAtAge && holder.MasterpieceValue() <= found.BaseValue {
		t.Errorf("MasterpieceValue = %d at age %d, want more than the Cr%d it sold for at %d "+
			"(Vintage appreciation never applied)",
			holder.MasterpieceValue(), holder.Age, found.BaseValue, found.CreatedAtAge)
	}
}

// searchForAMasterpiece walks generated citizen,craftsman chains until
// one produces a Masterpiece, returning that character alongside how
// many Craftsmen served and how many terms they served between them —
// the two numbers that tell a "never created one" failure apart from a
// "never entered the career" one.
func searchForAMasterpiece(seeds uint64) (Character, int, int) {
	var (
		holder               Character
		craftsmen, attempted int
	)

	for seed := uint64(1); seed <= seeds; seed++ {
		c, ok, err := GenerateCareerChainCharacter(
			dice.New(rand.NewPCG(seed, seed)), []string{"citizen", "craftsman"}, 0)
		if err != nil || !ok {
			continue
		}

		for _, career := range c.Careers {
			if career.Name == CraftsmanCareerName && len(career.Terms) > 0 {
				craftsmen++
				attempted += len(career.Terms)
			}
		}

		if len(c.Masterpieces) > 0 && len(holder.Masterpieces) == 0 {
			holder = c
		}
	}

	return holder, craftsmen, attempted
}
