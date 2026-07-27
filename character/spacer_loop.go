package character

import "github.com/philoserf/traveller/dice"

// continueSpacerOutcome is continueSpacer's own dice-free decision,
// mirroring continueMarineOutcome's own shape: roll==2 always succeeds
// (Book 1's generic "Mandatory Continue," p.65-66), otherwise succeeds
// against the (possibly Risk-reduced) current Str (C1) — Book 1 p.81's
// own "Continue C1," matching Marine's own target, unlike Soldier's C3.
func continueSpacerOutcome(roll, str int) bool {
	if roll == 2 {
		return true
	}

	return roll <= str
}

func continueSpacer(r *dice.Roller, upp UPP) bool {
	return continueSpacerOutcome(r.TwoD6(), int(upp.Characteristics[C1]))
}

// ResolveSpacerCareer resolves a full multi-term Spacer career (Book 1
// p.81) — mirrors ResolveSoldierCareer's own structure (soldier_loop.go),
// threading a branchRow (not a resolved branch/branchMod pair, since
// Spacer's own Branch name/Mod depends on current Officer/Enlisted
// status every term — see spacer_generate.go's own ResolveSpacerTerm).
// The one-time branch-tied automatic skill at career start uses the
// row's own Enlisted name (spacerBranchEnlistedNames[branchRow]) — the
// character is definitionally still Enlisted at that point.
func ResolveSpacerCareer(r *dice.Roller, upp UPP) (Career, UPP) {
	return resolveSpacerCareerWithBudget(r, upp, maxCareerTerms, &agingSimulation{})
}

// resolveSpacerCareerWithBudget is ResolveSpacerCareer's own body, with
// the resolveCareerLoop term cap threaded as a parameter — see
// resolveCareerLoop's own doc comment for why.
func resolveSpacerCareerWithBudget(r *dice.Roller, upp UPP, maxTerms int, aging *agingSimulation) (Career, UPP) {
	career := Career{Name: SpacerCareerName, HasRank: true}

	if !BeginSpacer(r, int(upp.Characteristics[C4])) {
		// Book 1 p.65: a failed Begin roll costs a year. Careers whose
		// Begin is a prerequisite rather than a roll (Noble's Soc B+,
		// Craftsman's held skills, Citizen's automatic entry) charge
		// nothing here — there was no attempt to fail.
		upp = aging.chargeFailedAttempt(r, upp)

		return career, upp
	}

	branchRow := rollSpacerBranchRow(r)

	var priorTerms []Term

	// Book 1 p.61's Command College, which fires once this career
	// reaches Officer 4 and Continues (command_college.go).
	collegeCheck, collegeSkills := armedForcesCommandCollege(SpacerCareerName)

	terms, finalUPP := resolveCareerLoop(r, upp, spacerRiskRewardPositions,
		func(r *dice.Roller, upp UPP, ccPos Position) (Term, UPP) {
			term, updatedUPP := ResolveSpacerTerm(r, upp, ccPos, branchRow, priorTerms)
			term.SkillsAwarded = append(collegeSkills(), term.SkillsAwarded...)
			priorTerms = append(priorTerms, term)

			return term, updatedUPP
		},
		continueSpacer,
		maxTerms,
		aging,
		collegeCheck,
	)
	career.Terms = terms

	grantBranchSkillToFirstTerm(r, &career, spacerBranchEnlistedNames[branchRow])
	grantStartingRankAutoSkillToFirstTerm(&career, spacerRankAutomaticSkill)

	return career, finalUPP
}
