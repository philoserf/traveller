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
	return resolveSoldierCareerWithBudget(r, upp, maxCareerTerms, &agingSimulation{}, false, false)
}

// resolveSoldierCareerWithBudget is ResolveSoldierCareer's own body,
// with the resolveCareerLoop term cap threaded as a parameter — see
// resolveCareerLoop's own doc comment for why.
//
// commissioned is true for a Service Academy/OTC Commission (#113,
// commissionAppliesTo in career_chain.go) — p.61's "Success confers a
// Commission (OTC= Army Officer1...)" substitutes for BeginSoldier's own
// roll. flightSchool (only ever true alongside commissioned) is p.61's
// Flight School — see resolveMarineCareerWithBudget's own doc comment
// for the full shape; Soldier's own branch table has no Flight row
// either, same documented gap, Mod 0.
func resolveSoldierCareerWithBudget(
	r *dice.Roller, upp UPP, maxTerms int, aging *agingSimulation, commissioned, flightSchool bool,
) (Career, UPP) {
	career := Career{Name: SoldierCareerName, HasRank: true}

	if !commissioned && !BeginSoldier(r, int(upp.Characteristics[C1])) {
		// Book 1 p.65: a failed Begin roll costs a year. Careers whose
		// Begin is a prerequisite rather than a roll (Noble's Soc B+,
		// Craftsman's held skills, Citizen's automatic entry) charge
		// nothing here — there was no attempt to fail.
		upp = aging.chargeFailedAttempt(r, upp)

		return career, upp
	}

	branch, branchMod := "Flight", 0
	if !flightSchool {
		branch, branchMod = rollSoldierBranch(r)
	}

	var priorTerms []Term

	// Book 1 p.61's Command College, which fires once this career
	// reaches Officer 4 and Continues (command_college.go).
	collegeCheck, collegeSkills := armedForcesCommandCollege(SoldierCareerName)

	terms, finalUPP := resolveCareerLoop(r, upp, soldierRiskRewardPositions,
		func(r *dice.Roller, upp UPP, ccPos Position) (Term, UPP) {
			entryCommissioned := commissioned && len(priorTerms) == 0

			operationsRolls := operationsRollsPerTerm
			if flightSchool && len(priorTerms) == 0 {
				operationsRolls = operationsRollsPerTerm - 1
			}

			term, updatedUPP := ResolveSoldierTerm(
				r, upp, ccPos, branch, branchMod, priorTerms, entryCommissioned, operationsRolls)
			term.SkillsAwarded = append(collegeSkills(), term.SkillsAwarded...)
			priorTerms = append(priorTerms, term)

			return term, updatedUPP
		},
		continueSoldier,
		maxTerms,
		aging,
		collegeCheck,
	)
	career.Terms = terms

	grantBranchSkillToFirstTerm(r, &career, branch)

	// A Commissioned entry's own Officer1 skill is granted inside
	// ResolveSoldierTerm instead (entryCommissioned above) — calling
	// this too would double-grant.
	if !commissioned {
		grantStartingRankAutoSkillToFirstTerm(&career, soldierRankAutomaticSkill)
	}

	if flightSchool && len(career.Terms) > 0 {
		career.Terms[0].SkillsAwarded = append(
			career.Terms[0].SkillsAwarded,
			SkillLevel{Name: "Pilot", Level: 3, Kind: Skill},
		)
	}

	return career, finalUPP
}
