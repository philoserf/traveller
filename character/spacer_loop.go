package character

import "github.com/philoserf/traveller/dice"

// continueSpacerOutcome is continueSpacer's own dice-free decision,
// mirroring continueMarineOutcome's own shape: roll==2 always succeeds
// (Book 1's generic "Mandatory Continue," p.65-66), otherwise succeeds
// against the (possibly Risk-reduced) current Str (C1) — Book 1 p.81's
// own "Continue C1," matching Marine's own target, unlike Soldier's C3.
func continueSpacerOutcome(roll, str int) bool {
	return mandatoryContinueOutcome(roll, str)
}

func continueSpacer(r *dice.Roller, upp UPP) bool {
	return continueSpacerOutcome(r.TwoD6(), int(upp.Characteristics[C1]))
}

// ResolveSpacerCareer resolves a full multi-term Spacer career (Book 1
// p.81) — mirrors ResolveSoldierCareer's own structure (soldier_loop.go),
// threading a branchRow (not a resolved branch/branchMod pair, since
// Spacer's own Branch name/Mod depends on current Officer/Enlisted
// status every term — see spacer_generate.go's own ResolveSpacerTerm).
// This entry point is never Commissioned (it always passes commissioned=
// false to resolveSpacerCareerWithBudget), so term 1 is definitionally
// still Enlisted and the one-time branch-tied automatic skill at career
// start correctly resolves from the Enlisted name — see
// resolveSpacerCareerWithBudget's own doc comment for the Commissioned
// case, where that premise doesn't hold.
func ResolveSpacerCareer(r *dice.Roller, upp UPP) (Career, UPP) {
	return resolveSpacerCareerWithBudget(r, upp, maxCareerTerms, &agingSimulation{}, false, false)
}

// spacerFlightBranchRow is spacerBranchOfficerNames[5] == "Flight" —
// Flight School (#113) selects this row directly instead of rolling
// one, so a Commissioned entry's own forced isOfficer=true
// (ResolveSpacerTerm) resolves it to "Flight" through the ordinary
// spacerBranchNameAndMod lookup, the same table Command College's own
// Naval Academy skill pool already treats as authoritative.
const spacerFlightBranchRow = 5

// resolveSpacerCareerWithBudget is ResolveSpacerCareer's own body, with
// the resolveCareerLoop term cap threaded as a parameter — see
// resolveCareerLoop's own doc comment for why.
//
// commissioned is true for a Service Academy/NOTC Commission (#113,
// commissionAppliesTo in career_chain.go) — p.61's "Success confers a
// Commission (... NOTC= Navy Officer1 or Marine Officer1)" substitutes
// for BeginSpacer's own roll. p.61's own worked example is exactly this
// path: Eneri is "commissioned O1 Ensign in the Imperial Navy" via NOTC.
// flightSchool (only ever true alongside commissioned) is p.61's Flight
// School — see resolveMarineCareerWithBudget's own doc comment for the
// full shape; unlike Marine/Soldier, Spacer's own table already has a
// Flight row, so this needs no Mod-0 gap-filling.
//
// The one-time branch-tied automatic skill at career start (#139) uses
// spacerBranchNameAndMod(branchRow, commissioned) rather than always the
// Enlisted name: term 1's own isOfficer state is exactly commissioned
// (ResolveSpacerTerm's own entryCommissioned is commissioned &&
// len(priorTerms)==0, and priorTerms is empty for term 1). A Service
// Academy/NOTC Commission or Flight School entry forces isOfficer=true
// before term 1 resolves (spacer_generate.go), so the granted skill must
// come from the same Officer-side name term 1's own Branch actually
// resolved to.
func resolveSpacerCareerWithBudget(
	r *dice.Roller, upp UPP, maxTerms int, aging *agingSimulation, commissioned, flightSchool bool,
) (Career, UPP) {
	career := Career{Name: SpacerCareerName, HasRank: true}

	if !commissioned && !BeginSpacer(r, int(upp.Characteristics[C4])) {
		// Book 1 p.65: a failed Begin roll costs a year. Careers whose
		// Begin is a prerequisite rather than a roll (Noble's Soc B+,
		// Craftsman's held skills, Citizen's automatic entry) charge
		// nothing here — there was no attempt to fail.
		upp = aging.chargeFailedAttempt(r, upp)

		return career, upp
	}

	branchRow := spacerFlightBranchRow
	if !flightSchool {
		branchRow = rollSpacerBranchRow(r)
	}

	var priorTerms []Term

	// Book 1 p.61's Command College, which fires once this career
	// reaches Officer 4 and Continues (command_college.go).
	collegeCheck, collegeSkills := armedForcesCommandCollege(SpacerCareerName)

	terms, finalUPP := resolveCareerLoop(r, upp, spacerRiskRewardPositions,
		func(r *dice.Roller, upp UPP, ccPos Position) (Term, UPP) {
			entryCommissioned := commissioned && len(priorTerms) == 0

			operationsRolls := operationsRollsPerTerm
			if flightSchool && len(priorTerms) == 0 {
				operationsRolls = operationsRollsPerTerm - 1
			}

			term, updatedUPP := ResolveSpacerTerm(
				r,
				upp,
				ccPos,
				branchRow,
				priorTerms,
				entryCommissioned,
				operationsRolls,
			)
			term.SkillsAwarded = append(collegeSkills(), term.SkillsAwarded...)
			priorTerms = append(priorTerms, term)

			return term, updatedUPP
		},
		continueSpacer,
		maxTerms,
		aging,
		collegeCheck,
		nil,
	)
	career.Terms = terms

	term1Branch, _ := spacerBranchNameAndMod(branchRow, commissioned)
	grantBranchSkillToFirstTerm(r, &career, term1Branch)

	// A Commissioned entry's own Officer1 skill is granted inside
	// ResolveSpacerTerm instead (entryCommissioned above) — calling this
	// too would double-grant.
	if !commissioned {
		grantStartingRankAutoSkillToFirstTerm(&career, spacerRankAutomaticSkill)
	}

	if flightSchool && len(career.Terms) > 0 {
		career.Terms[0].SkillsAwarded = append(
			career.Terms[0].SkillsAwarded,
			SkillLevel{Name: "Pilot", Level: 3, Kind: Skill},
		)
	}

	return career, finalUPP
}
