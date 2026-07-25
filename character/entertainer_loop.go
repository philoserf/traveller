package character

import "github.com/philoserf/traveller/dice"

// continueEntertainer targets Fame — no separate Mod or natural-roll
// exception is documented anywhere on the page, the simplest Continue
// check of any career so far (matching Scholar's own plain-roll shape).
func continueEntertainer(r *dice.Roller, fame int) bool {
	return rollAgainstTarget(r, fame, 0)
}

// ResolveEntertainerCareer resolves a full multi-term Entertainer career
// (Book 1 p.77). Returns the Career and the character's own current Fame
// — the initial 2D roll if BeginEntertainer never even qualifies (Fame
// is rolled "Before Begin," independent of it), or the last term's own
// FameAfterTerm otherwise.
//
// Does not call resolveCareerLoop: confirmed directly against nextCC's
// own body (career_loop.go) that it calls highestOf(upp, available...),
// which indexes available[0] unconditionally — a nil/empty positions
// slice (Entertainer has no Controlling Characteristic pool at all)
// would panic there. This hand-rolled loop mirrors resolveCareerLoop's
// own body exactly (the maxCareerTerms cap, the same
// !RiskResult.Survived() || Disabled || !continueCareer stop condition)
// minus the CC-rotation this career doesn't have.
func ResolveEntertainerCareer(r *dice.Roller, upp UPP) (Career, int) {
	return resolveEntertainerCareerWithBudget(r, upp, maxCareerTerms)
}

// resolveEntertainerCareerWithBudget is ResolveEntertainerCareer's own
// body, with the term cap threaded as a parameter instead of the
// implicit maxCareerTerms constant — see resolveCareerLoop's own doc
// comment (career_loop.go) for why. Entertainer is a registered,
// chainable career (character/career_chain.go), so its own hand-rolled
// loop needs the same -age-target treatment even though it doesn't
// itself call resolveCareerLoop.
func resolveEntertainerCareerWithBudget(r *dice.Roller, upp UPP, maxTerms int) (Career, int) {
	career := Career{Name: EntertainerCareerName}

	fame := rollEntertainerFameTalent(r)
	talent := fame

	specialty := rollEntertainerSpecialty(r)
	career.Specialty = specialty

	if !BeginEntertainer(r, upp, specialty) {
		return career, fame
	}

	var terms []Term

	for range maxTerms {
		var term Term

		term, fame, talent = ResolveEntertainerTerm(r, fame, talent)
		terms = append(terms, term)

		if !term.RiskResult.Survived() || term.RiskResult == Disabled || !continueEntertainer(r, fame) {
			break
		}
	}

	career.Terms = terms

	return career, fame
}
