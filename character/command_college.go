package character

import "github.com/philoserf/traveller/dice"

// Book 1 p.61's own Command College:
//
//	"Command College is a special Military School for higher ranking
//	officers. A Character must attend Command College in the first year
//	of the term after he is promoted to Officer4, provided he
//	successfully Continues. A character who fails Command College may
//	not Continue in the service. Success at Command College awards two
//	skill levels from the appropriate Military or Naval Academy."
//
// and its p.59 row in the EDUCATION FOR CHARACTERS table, read by
// character offset against the header's own column boundaries rather
// than visually:
//
//	Command College | assigned | Int or Edu | 1 year | Int or C5 | 1x | 2x Skill-1
//
// giving Pre-Requisite "assigned", To Apply "Int or Edu", Duration one
// year, Pass/Fail "Int or C5", one roll, and two skills at level 1.
// The neighbouring ANM School row has the identical shape with "auto"
// in the To Apply column, which is what confirms Command College's
// "Int or Edu" really is an admission check and not a column shifted in
// from Duration: an "assigned" school can still gate admission.
//
// Two readings are recorded here rather than left implicit.
//
// No partial term. The three deferral comments this replaces
// (marine/soldier/spacer_promotion.go) assumed modelling this needed "a
// real structural change (an inserted partial term)". It does not. A
// term is already four years and p.59's Duration is one year *inside*
// it, so the year Command College consumes is a year the term was going
// to spend anyway; no age is added and no term is split. The Later
// Education rule that does consume a whole term — p.59's "if accepted
// substitutes that process for the entire term" — is about voluntary
// schooling a character elects at the start of a term, not about a
// school he is assigned to.
//
// Int or C5 is not a second distinct characteristic. C5 *is* Education
// for the Humans this codebase generates (characteristic.go), so p.59's
// "Int or C5" and "Int or Edu" name the same pair, and both checks
// resolve against the better of Int and Edu — the same highestOf
// treatment every other open "X or Y" choice in this package gets.

// commandCollegeOfficerTier is p.61's own "promoted to Officer4".
const commandCollegeOfficerTier = 4

// commandCollegeSkillGrants is the "2x Skill-1" in p.59's Provides
// column: two grants, each of one level.
const commandCollegeSkillGrants = 2

// commandCollegeSchool is the Academy a career's officers attend, and
// so the p.60 column their two skills are drawn from.
//
// p.61 offers only two — "two skill levels from the appropriate
// Military or Naval Academy" — while three careers reach O4. Soldier is
// the Army and takes the Military Academy; Spacer is the Navy and takes
// the Naval. Marine takes the Naval Academy too, because that is where
// p.61 says Marine officers are commissioned: "Service Academies ...
// provide graduates an Army or Navy Commission (a Naval Academy
// graduate may choose a Marine Commission instead)."
//
// The alternative reading — giving Marines p.60's own Marine column —
// was rejected because that column is an M in the *School* half of
// p.61's legend (Marine School), not the Academy half, and Command
// College is explicitly an Academy. Marine and Spacer therefore draw the
// same pool, which is the book's own arrangement rather than an
// approximation of it.
func commandCollegeSchool(careerName string) (schoolSet, bool) {
	switch careerName {
	case SoldierCareerName:
		return schoolMilitaryAcademy, true
	case MarineCareerName, SpacerCareerName:
		return schoolNavalAcademy, true
	default:
		return 0, false
	}
}

// resolveCommandCollege runs one character's Command College attendance
// and reports the skills earned and whether he may remain in the
// service.
//
// The admission check and the Pass/Fail check are both rolled, in that
// order, because p.59 lists both for this row and p.59's own process
// description keeps them distinct: "To Apply (for Admission), a
// character must Check one of the stated Characteristics. A failure
// disallows admission and consumes one year", then separately "A
// character must Pass the required number of times". Failing either is
// failing Command College, which p.61 makes terminal: "A character who
// fails Command College may not Continue in the service."
//
// Waivers (p.59) are deliberately not offered here. They are Education's
// own mechanic and are being built with the pre-career Education step;
// wiring half of one in now would put a roll in this dice stream that
// the other half would have to move.
func resolveCommandCollege(r *dice.Roller, upp UPP, careerName string) ([]SkillLevel, bool) {
	school, ok := commandCollegeSchool(careerName)
	if !ok {
		return nil, true
	}

	target := int(upp.Characteristics[highestOf(upp, C4, C5)])

	if !succeedsAgainst(r.TwoD6(), target, 0) {
		return nil, false
	}

	if !succeedsAgainst(r.TwoD6(), target, 0) {
		return nil, false
	}

	pool := skillsForSchool(school, false)
	if len(pool) == 0 {
		return nil, true
	}

	skills := make([]SkillLevel, 0, commandCollegeSkillGrants)
	for range commandCollegeSkillGrants {
		skills = append(skills, skillLevel1(pool[r.Uniform(len(pool))-1], Skill))
	}

	return skills, true
}

// commandCollegeDue reports whether p.61's trigger has just fired: the
// character reached Officer 4 in the term that has just ended and has
// not been before.
//
// Keyed on reaching the tier rather than on the promotion roll itself so
// it cannot fire twice, and so a character who arrives at O4 by any
// route still attends exactly once — the same "one-time event upon
// reaching rank X" framing the automatic rank skills already use
// (marineRankAutomaticSkill).
func commandCollegeDue(terms []Term, enlistedRanks, officerRanks, attended int) bool {
	if attended > 0 {
		return false
	}

	isOfficer, tier := rankState(terms, enlistedRanks, officerRanks)

	return isOfficer && tier >= commandCollegeOfficerTier
}

// armedForcesCommandCollege builds the two halves the three Armed Forces
// careers need to run p.61's Command College: the afterContinue hook
// that decides whether the character may stay in the service, and a
// drain the term resolver calls to collect the skills the school
// awarded.
//
// Two halves rather than one because the roll and the award happen in
// different terms. p.61 puts attendance "in the first year of the term
// after he is promoted to Officer4", so the check belongs at the end of
// the O4 term — where the Continue it is conditioned on is decided — and
// the skills belong to the term that follows, which is the one actually
// spent at the school. Draining into that term keeps the per-term record
// honest instead of back-dating two skills onto a term the character had
// already finished.
func armedForcesCommandCollege(careerName string) (afterContinue, func() []SkillLevel) {
	enlistedRanks, officerRanks := armedForcesRankCounts(careerName)

	var (
		pending  []SkillLevel
		attended int
	)

	hook := func(r *dice.Roller, upp UPP, terms []Term) bool {
		if !commandCollegeDue(terms, enlistedRanks, officerRanks, attended) {
			return true
		}

		attended++

		skills, mayContinue := resolveCommandCollege(r, upp, careerName)
		pending = skills

		return mayContinue
	}

	drain := func() []SkillLevel {
		earned := pending
		pending = nil

		return earned
	}

	return hook, drain
}

// armedForcesRankCounts is how many Enlisted and Officer tiers a career's
// rank ladders have, which rankState needs to know where the top is.
// Looked up here rather than passed in so the three loop call sites
// cannot pair one career's name with another's ladders.
func armedForcesRankCounts(careerName string) (int, int) {
	switch careerName {
	case MarineCareerName:
		return len(marineEnlistedRankNames), len(marineOfficerRankNames)
	case SoldierCareerName:
		return len(soldierEnlistedRankNames), len(soldierOfficerRankNames)
	case SpacerCareerName:
		return len(spacerEnlistedRankNames), len(spacerOfficerRankNames)
	default:
		return 0, 0
	}
}
