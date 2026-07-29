package character

import "github.com/philoserf/traveller/dice"

// continueMarineOutcome is continueMarine's own dice-free decision, the
// same shape continueScoutOutcome already establishes: roll==2 always
// succeeds (Book 1's generic "Mandatory Continue," p.65-66), otherwise
// succeeds against the (possibly Risk-reduced) current Str — Book 1
// p.86's own "Continue C1," characteristic-based like Scout's own
// Continue Int, not a flat target like Citizen's/Noble's.
func continueMarineOutcome(roll, str int) bool {
	if roll == 2 {
		return true
	}

	return roll <= str
}

func continueMarine(r *dice.Roller, upp UPP) bool {
	return continueMarineOutcome(r.TwoD6(), int(upp.Characteristics[C1]))
}

// ResolveMarineCareer resolves a full multi-term Marine career (Book 1
// p.86) via resolveCareerLoop (career_loop.go — shared with
// ResolveScoutCareer once Marine became a second real caller of the
// identical loop shape). Branch is selected once, before the term loop
// — Book 1's own "Officers may not change Branch once selected;
// Enlisted may select a new Branch upon Promotion" — a real
// simplification: this codebase doesn't yet model that Enlisted
// re-selection path, matching Book 1's own "record what's true now,
// defer what's complex" precedent elsewhere.
//
// priorTerms accumulates locally via closure capture rather than
// widening resolveCareerLoop's own shared resolveTerm signature — Scout
// doesn't need a terms-so-far view, so the shared loop stays untouched;
// only Marine's own closure needs it, to give ResolveMarineTerm
// (marine_generate.go) what it needs to derive rank state and cumulative
// Medal Mods for Commission/Promotion rolls.
//
// HasRank is true — Marine is not in Book 1 p.65's own no-rank career
// list — and Term.Rank is now genuinely populated every term
// (marineRankName), not a deferred gap.
func ResolveMarineCareer(r *dice.Roller, upp UPP) (Career, UPP) {
	return resolveMarineCareerWithBudget(r, upp, maxCareerTerms, &agingSimulation{}, false, false)
}

// resolveMarineCareerWithBudget is ResolveMarineCareer's own body, with
// the resolveCareerLoop term cap threaded as a parameter — see
// resolveCareerLoop's own doc comment for why.
//
// commissioned is true for a Service Academy/OTC/NOTC Commission (#113,
// commissionAppliesTo in career_chain.go) — p.61's own "provide graduates
// an Army or Navy Commission (a Naval Academy graduate may choose a
// Marine Commission instead)" substitutes for BeginMarine's own roll, the
// same "no attempt to fail" treatment automatic-entry careers already
// get. flightSchool (only ever true alongside commissioned — see
// career_chain.go) is p.61's Flight School: it forces Branch=Flight for
// the whole career (Marine's own branch table has no such row — a
// documented gap, not a contradiction, so Mod is 0) and shortens term 1
// to operationsRollsPerTerm-1 Operations rolls.
func resolveMarineCareerWithBudget(
	r *dice.Roller, upp UPP, maxTerms int, aging *agingSimulation, commissioned, flightSchool bool,
) (Career, UPP) {
	career := Career{Name: MarineCareerName, HasRank: true}

	if !commissioned && !BeginMarine(r, int(upp.Characteristics[C1])) {
		// Book 1 p.65: a failed Begin roll costs a year. Careers whose
		// Begin is a prerequisite rather than a roll (Noble's Soc B+,
		// Craftsman's held skills, Citizen's automatic entry) charge
		// nothing here — there was no attempt to fail.
		upp = aging.chargeFailedAttempt(r, upp)

		return career, upp
	}

	branch, branchMod := "Flight", 0
	if !flightSchool {
		branch, branchMod = rollMarineBranch(r)
	}

	var priorTerms []Term

	// Book 1 p.61's Command College, which fires once this career
	// reaches Officer 4 and Continues (command_college.go).
	collegeCheck, collegeSkills := armedForcesCommandCollege(MarineCareerName)

	terms, finalUPP := resolveCareerLoop(r, upp, marineRiskRewardPositions,
		func(r *dice.Roller, upp UPP, ccPos Position) (Term, UPP) {
			entryCommissioned := commissioned && len(priorTerms) == 0

			operationsRolls := operationsRollsPerTerm
			if flightSchool && len(priorTerms) == 0 {
				operationsRolls = operationsRollsPerTerm - 1
			}

			term, updatedUPP := ResolveMarineTerm(
				r, upp, ccPos, branch, branchMod, priorTerms, entryCommissioned, operationsRolls)
			term.SkillsAwarded = append(collegeSkills(), term.SkillsAwarded...)
			priorTerms = append(priorTerms, term)

			return term, updatedUPP
		},
		continueMarine,
		maxTerms,
		aging,
		collegeCheck,
	)
	career.Terms = terms

	grantBranchSkillToFirstTerm(r, &career, branch)

	// A plain (uncommissioned) entry gets no
	// grantStartingRankAutoSkillToFirstTerm call: unlike Soldier's own S1
	// Private and Spacer's own R1 Spacehand, M1 Private has no automatic
	// skill at all (marineRankAutomaticSkill's own doc comment) — there
	// is nothing to grant. A Commissioned entry's own Officer1 skill is
	// granted inside ResolveMarineTerm instead (entryCommissioned above),
	// through the same mechanism a term-by-term Commission already uses.
	if !commissioned {
		grantStartingRankAutoSkillToFirstTerm(&career, marineRankAutomaticSkill)
	}

	if flightSchool && len(career.Terms) > 0 {
		career.Terms[0].SkillsAwarded = append(
			career.Terms[0].SkillsAwarded,
			SkillLevel{Name: "Pilot", Level: 3, Kind: Skill},
		)
	}

	return career, finalUPP
}
