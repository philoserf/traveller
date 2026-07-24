package character

import "github.com/philoserf/traveller/dice"

// LifeStageForAge returns the Life Stage (0-9) for age years, per Book 1
// p.89's own "THE STAGES OF LIFE" table. Ages beyond 71 (Retirement's
// own upper bound) clamp to Life Stage 9 — the table doesn't define
// anything further, and a character surviving repeated Aging Checks
// that long is already an edge case the book doesn't dwell on.
func LifeStageForAge(age int) int {
	switch {
	case age <= 1:
		return 0
	case age <= 9:
		return 1
	case age <= 17:
		return 2
	case age <= 25:
		return 3
	case age <= 33:
		return 4
	case age <= 41:
		return 5
	case age <= 49:
		return 6
	case age <= 57:
		return 7
	case age <= 65:
		return 8
	default:
		return 9
	}
}

// AgeFromTermsServed approximates a character's age at the end of Career
// Resolution: Young Adult (18, Book 1 p.89's own "typical start of
// adventuring") plus 4 years per term served. An approximation, not
// exact — the same imprecision already flagged in
// character_generate.go's own doc comment: whether Begin succeeded on
// the first roll or needed a 1-year Retry isn't surfaced anywhere in
// this codebase, so this can undercount by up to a year.
func AgeFromTermsServed(termsServed int) int {
	return 18 + 4*termsServed
}

// agingCheckOutcome is rollAgingCheck's own dice-free decision. Book 1
// p.89's own "To Feel Age Effects (The Aging Check): 2D < Life Stage" —
// strictly less-than, not the <=-succeeds convention every other roll in
// this codebase uses (Risk, Reward, Begin, Continue, Citizen Life all
// succeed on <=) — verified directly against the primary source given
// how much this diverges from the established pattern; do not reuse
// succeedsAgainst here. Success means the character DOES feel the aging
// effect (Book 1's own "A character wants to FAIL this action").
func agingCheckOutcome(roll, lifeStage int) bool {
	return roll < lifeStage
}

func rollAgingCheck(r *dice.Roller, lifeStage int) bool {
	return agingCheckOutcome(r.TwoD6(), lifeStage)
}
