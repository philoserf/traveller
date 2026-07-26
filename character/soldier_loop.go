package character

import "github.com/philoserf/traveller/dice"

// continueSoldierOutcome is continueSoldier's own dice-free decision,
// mirroring continueMarineOutcome's own shape: roll==2 always succeeds
// (Book 1's generic "Mandatory Continue," p.65-66), otherwise succeeds
// against the (possibly Risk-reduced) current End (C3) — Book 1 p.82's
// own "Continue C3," unlike Marine's own "Continue C1" (Str).
func continueSoldierOutcome(roll, end int) bool {
	if roll == 2 {
		return true
	}

	return roll <= end
}

func continueSoldier(r *dice.Roller, upp UPP) bool {
	return continueSoldierOutcome(r.TwoD6(), int(upp.Characteristics[C3]))
}

// ResolveSoldierCareer resolves a full multi-term Soldier career (Book 1
// p.82) — mirrors ResolveMarineCareer's own structure exactly
// (marine_loop.go): Branch selected once before the loop, priorTerms
// accumulated via closure capture, branchAutomaticSkill applied to term
// 1 after the loop completes.
func ResolveSoldierCareer(r *dice.Roller, upp UPP) (Career, UPP) {
	return resolveSoldierCareerWithBudget(r, upp, maxCareerTerms, &agingSimulation{})
}

// resolveSoldierCareerWithBudget is ResolveSoldierCareer's own body,
// with the resolveCareerLoop term cap threaded as a parameter — see
// resolveCareerLoop's own doc comment for why.
func resolveSoldierCareerWithBudget(r *dice.Roller, upp UPP, maxTerms int, aging *agingSimulation) (Career, UPP) {
	career := Career{Name: SoldierCareerName, HasRank: true}

	if !BeginSoldier(r, int(upp.Characteristics[C1])) {
		return career, upp
	}

	branch, branchMod := rollSoldierBranch(r)

	var priorTerms []Term

	terms, finalUPP := resolveCareerLoop(r, upp, soldierRiskRewardPositions,
		func(r *dice.Roller, upp UPP, ccPos Position) (Term, UPP) {
			term, updatedUPP := ResolveSoldierTerm(r, upp, ccPos, branch, branchMod, priorTerms)
			priorTerms = append(priorTerms, term)

			return term, updatedUPP
		},
		continueSoldier,
		maxTerms,
		aging,
	)
	career.Terms = terms

	grantBranchSkillToFirstTerm(r, &career, branch)
	grantStartingRankAutoSkillToFirstTerm(&career, soldierRankAutomaticSkill)

	return career, finalUPP
}
