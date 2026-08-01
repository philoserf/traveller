package character

import "github.com/philoserf/traveller/dice"

// citizenLifePositions is Citizen Life's own Controlling Characteristic
// set (Book 1 p.78: "Citizen Life C1 C2 C3 C4") — one more than Scout's
// own C1 C2 C3 (career_loop.go's scoutRiskRewardPositions).
var citizenLifePositions = []Position{C1, C2, C3, C4}

// continueCitizenOutcome is continueCitizen's own dice-free decision,
// split out the same way continueScoutOutcome is. roll==2 always
// succeeds (Book 1's generic "Mandatory Continue," p.65-66) — kept as an
// explicit branch even though it's provably unreachable here (roll==2 is
// already covered by succeedsAgainst's own roll<=10 check below), the
// same deliberate choice continueScoutOutcome's own doc comment already
// explains: implementing the literal rule, not just its currently-observable
// effect. Otherwise defers to succeedsAgainst against p.78's fixed
// "Continue 10-" target — unlike Scout's characteristic-based Continue
// Int, this never varies by UPP.
func continueCitizenOutcome(roll int) bool {
	return mandatoryContinueOutcome(roll, 10)
}

func continueCitizen(r *dice.Roller) bool {
	return continueCitizenOutcome(r.TwoD6())
}

// citizenLifeSuccessCount counts CitizenLifeSucceeded terms — derived
// from terms rather than a separately-tracked counter, so it can't drift
// out of sync with the Terms it's counting.
func citizenLifeSuccessCount(terms []Term) int {
	n := 0

	for _, t := range terms {
		if t.CitizenLifeSucceeded {
			n++
		}
	}

	return n
}

// ResolveCitizenCareer resolves a full multi-term Citizen career (Book 1
// p.78) by looping ResolveCitizenTerm. BeginCitizen is called for
// symmetry with ResolveScoutCareer's own Begin-then-loop shape even
// though its own "To Begin Auto" rule (BeginCitizen's own doc comment)
// makes the zero-terms branch below provably unreachable today — kept
// visible rather than silently assumed, the same reasoning BeginCitizen
// itself exists as a real function for at all.
//
// Each term: pick this term's Controlling Characteristic (nextCC,
// rotating through citizenLifePositions per p.65), resolve the term,
// append it, then stop if Continue fails; otherwise loop, capped
// defensively at maxCareerTerms. Citizen Life has no injury reduction,
// but its Personal-column improvements are applied between terms by the
// internal UPP-returning resolver. The public compatibility wrapper
// returns only Career.
func ResolveCitizenCareer(r *dice.Roller, upp UPP) Career {
	career, _ := resolveCitizenCareerAndUPPWithBudget(r, upp, maxCareerTerms, &agingSimulation{})

	return career
}

// resolveCitizenCareerWithBudget is ResolveCitizenCareer's own body,
// with the term cap threaded as a parameter instead of the implicit
// maxCareerTerms constant — see resolveCareerLoop's own doc comment
// (career_loop.go) for why. Citizen is the guaranteed fallback for the
// multi-career chain (character/career_chain.go), so its own hand-rolled
// loop needs the same -age-target treatment even though it doesn't
// itself call resolveCareerLoop.
func resolveCitizenCareerWithBudget(r *dice.Roller, upp UPP, maxTerms int, aging *agingSimulation) Career {
	career, _ := resolveCitizenCareerAndUPPWithBudget(r, upp, maxTerms, aging)

	return career
}

func resolveCitizenCareerAndUPPWithBudget(r *dice.Roller, upp UPP, maxTerms int, aging *agingSimulation) (Career, UPP) {
	career := Career{Name: CitizenCareerName}

	if !BeginCitizen() {
		return career, upp
	}

	usedThisCycle := make(map[Position]bool, len(citizenLifePositions))

	for range maxTerms {
		if !aging.alive() {
			break
		}

		ccPos := nextCC(upp, citizenLifePositions, usedThisCycle)

		term, jobSkill, hobbySkill := ResolveCitizenTerm(
			r, upp, ccPos, citizenLifeSuccessCount(career.Terms), career.JobSkill, career.HobbySkill)
		upp = applyPersonalAwards(upp, term.SkillsAwarded)
		career.Terms = append(career.Terms, term)
		career.JobSkill, career.HobbySkill = jobSkill, hobbySkill

		upp = aging.advanceTerm(r, upp)
		if !aging.alive() {
			break
		}

		if !continueCitizen(r) {
			break
		}
	}

	return career, upp
}
