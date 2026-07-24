package character

import (
	"fmt"

	"github.com/philoserf/traveller/dice"
)

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

const (
	physicalAgingOnset = 34
	mentalAgingOnset   = 66
	agingInterval      = 4
)

// agingCheckpoints returns the ages, ascending, at which an Aging Check is
// due for finalAge years lived: physicalAgingOnset through finalAge
// inclusive, every agingInterval years. Empty if finalAge is below onset.
func agingCheckpoints(finalAge int) []int {
	var ages []int

	for age := physicalAgingOnset; age <= finalAge; age += agingInterval {
		ages = append(ages, age)
	}

	return ages
}

// agingPositionsAt returns which characteristics are due an Aging Check at
// age: Str/Dex/End (C1-C3) once Physical Aging has begun, plus Int (C4)
// once Mental Aging has also begun. Both onsets share phase 2 mod 4 (34
// mod 4 == 66 mod 4 == 2), so from age 66 onward every checkpoint in
// agingCheckpoints checks all four at once.
func agingPositionsAt(age int) []Position {
	positions := []Position{C1, C2, C3}
	if age >= mentalAgingOnset {
		positions = append(positions, C4)
	}

	return positions
}

// agingSeverity classifies how many characteristics were newly reduced to
// 0 within a single Aging checkpoint, per Book 1 p.89's own escalating
// consequences.
type agingSeverity int

// agingSeverity values, from least to most severe.
const (
	agingSeverityNone agingSeverity = iota
	agingSeverityMajor
	agingSeverityExtreme
)

// classifyAgingBatch is agingSeverity's own dice-free decision. Book 1
// only names outcomes for exactly one/two/three simultaneous zeros; four
// (only possible at age 66+, when Str/Dex/End/Int can all be checked in
// the same checkpoint) isn't separately defined, so this treats "three or
// more" as the book's own most severe named tier rather than inventing an
// undefined fourth one.
func classifyAgingBatch(zeroedCount int) agingSeverity {
	switch {
	case zeroedCount >= 3:
		return agingSeverityExtreme
	case zeroedCount == 2:
		return agingSeverityMajor
	default:
		return agingSeverityNone
	}
}

// agingExtremeIllnessIsFatal reports whether an agingSeverityExtreme batch
// kills the character, per Book 1 p.89's own "The second time three
// characteristics are reduced to 0, the character dies." priorExtremeCount
// is how many such batches already occurred earlier in this same
// simulation.
func agingExtremeIllnessIsFatal(priorExtremeCount int) bool {
	return priorExtremeCount >= 1
}

// ResolveAging simulates every Aging checkpoint from Physical Aging's own
// onset through finalAge against upp, per Book 1 p.89. Returns the
// resulting UPP, whether the character survived (false only when a second
// agingSeverityExtreme batch occurs — "the character dies"), and a
// human-readable note for every major/extreme illness or the fatal event,
// in chronological order.
//
// A characteristic already at 0 entering a checkpoint (from some
// non-Aging cause, e.g. a Scout Risk wound) is left at 0 rather than
// underflowing or counting as newly zeroed — Aging's own reset-to-1 rule
// exists for values Aging itself just reduced to 0, not as a
// general-purpose floor for characteristics zeroed by other mechanics.
//
// On death, the characteristics zeroed in the fatal checkpoint are left
// at 0 in the returned UPP rather than reset to 1 — the reset-to-1 rule
// is for a character who survives the illness, which a dead one, by
// definition, does not.
func ResolveAging(r *dice.Roller, upp UPP, finalAge int) (UPP, bool, []string) {
	var notes []string

	extremeCount := 0

	for _, age := range agingCheckpoints(finalAge) {
		lifeStage := LifeStageForAge(age)

		var zeroed []Position

		for _, p := range agingPositionsAt(age) {
			if upp.Characteristics[p] == 0 || !rollAgingCheck(r, lifeStage) {
				continue
			}

			upp.Characteristics[p]--

			if upp.Characteristics[p] == 0 {
				zeroed = append(zeroed, p)
			}
		}

		switch classifyAgingBatch(len(zeroed)) {
		case agingSeverityMajor:
			notes = append(notes, fmt.Sprintf(
				"Age %d: major illness (two characteristics reduced to 0) — four weeks recuperation", age))
		case agingSeverityExtreme:
			if agingExtremeIllnessIsFatal(extremeCount) {
				notes = append(notes, fmt.Sprintf(
					"Age %d: died of natural causes (%d characteristics reduced to 0 for a second time)",
					age,
					len(zeroed),
				))

				return upp, false, notes
			}

			extremeCount++

			notes = append(notes, fmt.Sprintf(
				"Age %d: extremely major illness (%d characteristics reduced to 0) — four months recuperation",
				age,
				len(zeroed),
			))
		case agingSeverityNone:
			// No reduction reached 0 this checkpoint — nothing to record.
		}

		for _, p := range zeroed {
			upp.Characteristics[p] = 1
		}
	}

	return upp, true, notes
}
