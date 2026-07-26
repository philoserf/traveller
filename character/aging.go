package character

import (
	"fmt"
	"strings"

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
// agingSeverityExtreme batch occurs — "the character dies"), a
// human-readable note for every major/extreme illness or the fatal event,
// in chronological order, and the age the character actually reached.
//
// reachedAge is finalAge for a survivor, or the fatal checkpoint's own
// age for one who died — never finalAge in that case. Aging runs over
// the same span Career Resolution already covered (finalizeAging passes
// the career-end age), so without this a character could be reported as
// having died at 70 while their Age field still read 74: older than
// their own recorded death. Callers use it to cap Age/LifeStage so the
// two can't contradict each other.
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
func ResolveAging(r *dice.Roller, upp UPP, finalAge int) (UPP, bool, []string, int) {
	var sim agingSimulation

	for _, age := range agingCheckpoints(finalAge) {
		upp = sim.checkpoint(r, upp, age)
		if !sim.alive() {
			return upp, false, sim.notes, sim.diedAtAge
		}
	}

	return upp, true, sim.notes, finalAge
}

// agingSimulation carries Aging state across the checkpoints of one
// character's life, so they can be applied chronologically — interleaved
// between career terms by resolveCareerLoop (career_loop.go) — rather
// than in a single pass after Career Resolution has already finished.
//
// Ordering is the whole point, not an implementation detail. Aging
// checkpoints fall every agingInterval years from physicalAgingOnset,
// and career terms are the same length starting at 18, so every
// checkpoint lands exactly on a term boundary (age 34 is the end of term
// 4). Running them in place means a characteristic Aging reduced is
// already reduced for the next term's own Risk/Reward/Continue rolls,
// and a character killed at a checkpoint simply serves no further terms
// — instead of a sheet that records fourteen terms of service by someone
// who died partway through them.
// termsServed lives here, rather than being derived per career, because
// Aging spans a whole life: in a multi-career chain (career_chain.go)
// the same simulation is threaded through every segment in turn, so the
// checkpoint after a Merchant's first term depends on however many Scout
// terms preceded it. A per-career count would restart the clock at each
// transfer and never reach a checkpoint at all.
type agingSimulation struct {
	termsServed  int
	extremeCount int
	notes        []string
	diedAtAge    int // 0 while alive; the fatal checkpoint's own age once not
}

// alive reports whether no checkpoint has killed the character yet.
func (s *agingSimulation) alive() bool {
	return s.diedAtAge == 0
}

// age is the character's age given every term recorded so far, or the
// age they died at if a checkpoint killed them.
func (s *agingSimulation) age() int {
	if !s.alive() {
		return s.diedAtAge
	}

	return AgeFromTermsServed(s.termsServed)
}

// advanceTerm records one completed four-year term and runs whatever
// Aging checkpoint falls at its end — the single call a career loop
// makes per term. Returns the resulting UPP so the next term's own rolls
// see any reduction this checkpoint just applied.
func (s *agingSimulation) advanceTerm(r *dice.Roller, upp UPP) UPP {
	s.termsServed++

	return s.checkpoint(r, upp, AgeFromTermsServed(s.termsServed))
}

// recordFatalTerm counts a term the character died during (Book 1 p.69,
// a Risk roll to zero) without running an Aging checkpoint for it. The
// term still counts toward Age — they served it, right up until it
// killed them — but the dead don't then also grow four years older.
// Skipping the count entirely would report an Age four years short of
// the career the sheet actually lists.
func (s *agingSimulation) recordFatalTerm() {
	s.termsServed++
}

// checkpointDueAt reports whether an Aging checkpoint falls exactly at
// age — from physicalAgingOnset onward, every agingInterval years.
func checkpointDueAt(age int) bool {
	return age >= physicalAgingOnset && (age-physicalAgingOnset)%agingInterval == 0
}

// checkpoint runs the single Aging checkpoint due at age against upp,
// recording any illness or death, and returns the resulting UPP. A no-op
// (upp returned unchanged) when no checkpoint is due at age or the
// character is already dead, so callers can call it after every term
// without first working out whether this particular term ends on one.
func (s *agingSimulation) checkpoint(r *dice.Roller, upp UPP, age int) UPP {
	if !s.alive() || !checkpointDueAt(age) {
		return upp
	}

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
		s.notes = append(s.notes, fmt.Sprintf(
			"Age %d: major illness (two characteristics reduced to 0) — four weeks recuperation", age))
	case agingSeverityExtreme:
		if agingExtremeIllnessIsFatal(s.extremeCount) {
			s.notes = append(s.notes, fmt.Sprintf(
				"Age %d: died of natural causes (%d characteristics reduced to 0 for a second time)",
				age,
				len(zeroed),
			))
			s.diedAtAge = age

			// The fatal checkpoint's own zeroed characteristics stay at 0
			// rather than resetting to 1 — that reset is for surviving an
			// illness, which a dead character by definition did not.
			return upp
		}

		s.extremeCount++

		s.notes = append(s.notes, fmt.Sprintf(
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

	return upp
}

// finalizeAging reports the Age, LifeStage, Notes and overall survival
// of a character whose Aging has *already* been simulated, checkpoint by
// checkpoint, by the career loops (career_loop.go's own resolveCareerLoop
// and the three hand-rolled equivalents). It rolls no dice of its own.
//
// It used to: an earlier version ran the entire Aging simulation here,
// in one pass after Career Resolution had already finished. That could
// only ever produce an inconsistent character, because the two passes
// described the same span of years independently — a Scholar might serve
// fourteen terms to age 74 in one pass and die at 70 in the other, and
// the sheet had to pick which to believe. Simulating Aging inside the
// term loop removes the contradiction at its source: a character who
// dies at a checkpoint simply serves no further terms, so Age, the death
// note, and the career history can't disagree.
//
// survivedCareer is false for a character who died during Career
// Resolution itself (Book 1 p.69's "Dying During Character Generation");
// aging.alive() is false for one killed by an Aging checkpoint (p.89).
// The returned bool is the conjunction — either kills.
func finalizeAging(aging *agingSimulation, survivedCareer bool) (int, int, string, bool) {
	age := aging.age()

	return age, LifeStageForAge(age), strings.Join(aging.notes, "; "), survivedCareer && aging.alive()
}
