package character

import "testing"

// seedSearchLimit is how far seedFor will look.
//
// Set from the rarest outcome any fixture here needs: an end-to-end
// Aging death, which requires three characteristics reduced to 0 twice
// over (Book 1 p.89), and so requires a career that keeps restoring them
// between checkpoints. Measured at roughly three per 100,000 generated
// Scouts. Across several artificial dice-stream offsets the first
// qualifying seed landed anywhere from 4,972 to 71,984, and a 20,000
// limit missed it outright at one of them — so the limit is well clear
// of the worst case observed rather than of the typical one.
//
// A full sweep costs about 1.2 seconds; searches stop at the first hit,
// so the usual cost is far less.
const seedSearchLimit = 200_000

// seedFor returns the first seed from 1 upward whose generated outcome
// satisfies want, and fails the test if none does.
//
// Fixtures that need a rare generated outcome — an Aging death, a Begin
// that never qualifies, a career long enough for an -age target to cut
// short — used to hard-code a seed found by hand, each with a comment
// like "confirmed by direct inspection". That works exactly until
// something moves the dice stream, at which point the seed produces an
// ordinary character and the test fails with "fixture assumption broke"
// while asserting nothing about the behaviour it was written for. Eight
// of them broke at once against a single extra die roll.
//
// Searching at run time costs a few thousand cheap generations and never
// needs re-deriving.
//
// The predicate must capture only the *precondition* — the situation the
// test exists to exercise — and never the behaviour being asserted.
// Folding an assertion into the predicate makes the search find a seed
// that satisfies it and the test then verify it, which is vacuous.
func seedFor(t *testing.T, what string, want func(seed uint64) bool) uint64 {
	t.Helper()

	for seed := uint64(1); seed <= seedSearchLimit; seed++ {
		if want(seed) {
			return seed
		}
	}

	t.Fatalf("no seed in 1..%d produced %s", seedSearchLimit, what)

	return 0
}

// termsForAge is the inverse of AgeFromTermsServed: how many whole terms
// fit inside an -age target. Derived rather than written out, so a test
// asserting "an age target of 30 allows three terms" cannot drift from
// the function that actually converts one into the other.
func termsForAge(age int) int {
	terms := 0
	for AgeFromTermsServed(terms+1) <= age {
		terms++
	}

	return terms
}

// musterOutCash totals a career's own Mustering Out Money column, so a
// test can assert that Character.Cash reflects it without pinning the
// figure a particular seed happened to roll.
func musterOutCash(career Career) int {
	total := 0

	for _, entry := range career.MusteringOut.Money {
		if amount, ok := musterOutCashAmount(entry); ok {
			total += amount
		}
	}

	return total
}

// firstRepeatedSkill returns the name of a skill some character was
// granted more than once across all their careers' terms, or "" if none
// was. Personal-column awards are excluded: those are characteristic
// boosts, which aggregateSkills deliberately never merges into Skills.
//
// In its own function so the seed predicate that uses it stays a single
// readable expression rather than three nested loops.
func firstRepeatedSkill(c Character) string {
	counts := map[string]int{}

	for _, career := range c.Careers {
		for _, term := range career.Terms {
			for _, sk := range term.SkillsAwarded {
				if sk.Kind != Skill {
					continue
				}

				counts[sk.Name]++
				if counts[sk.Name] > 1 {
					return sk.Name
				}
			}
		}
	}

	return ""
}
