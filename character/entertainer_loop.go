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
	career, fame, _ := resolveEntertainerCareerAndUPPWithBudget(r, upp, maxCareerTerms, &agingSimulation{})

	return career, fame
}

func resolveEntertainerCareerAndUPPWithBudget(
	r *dice.Roller,
	upp UPP,
	maxTerms int,
	aging *agingSimulation,
) (Career, int, UPP) {
	career := Career{Name: EntertainerCareerName}

	fame := rollEntertainerFameTalent(r)
	talent := fame

	specialty := rollEntertainerSpecialty(r)
	career.Specialty = specialty

	if !BeginEntertainer(r, upp, specialty) {
		return career, fame, upp
	}

	var terms []Term

	for range maxTerms {
		if !aging.alive() {
			break
		}

		var term Term

		term, fame, talent = ResolveEntertainerTerm(r, fame, talent)
		upp = applyPersonalAwards(upp, term.SkillsAwarded)
		terms = append(terms, term)

		// Entertainer's own "Dead" is Talent exhausted, not physical
		// death (see resolveEntertainerSegment), so it still ages.
		upp = aging.advanceTerm(r, upp)
		if !aging.alive() {
			break
		}

		if !term.RiskResult.Survived() || term.RiskResult == Disabled || !continueEntertainer(r, fame) {
			break
		}
	}

	career.Terms = terms

	return career, fame, upp
}
