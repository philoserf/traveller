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
	return resolveMarineCareerWithBudget(r, upp, maxCareerTerms)
}

// resolveMarineCareerWithBudget is ResolveMarineCareer's own body, with
// the resolveCareerLoop term cap threaded as a parameter — see
// resolveCareerLoop's own doc comment for why.
func resolveMarineCareerWithBudget(r *dice.Roller, upp UPP, maxTerms int) (Career, UPP) {
	career := Career{Name: MarineCareerName, HasRank: true}

	if !BeginMarine(r, int(upp.Characteristics[C1])) {
		return career, upp
	}

	branch, branchMod := rollMarineBranch(r)

	var priorTerms []Term

	terms, finalUPP := resolveCareerLoop(r, upp, marineRiskRewardPositions,
		func(r *dice.Roller, upp UPP, ccPos Position) (Term, UPP) {
			term, updatedUPP := ResolveMarineTerm(r, upp, ccPos, branch, branchMod, priorTerms)
			priorTerms = append(priorTerms, term)

			return term, updatedUPP
		},
		continueMarine,
		maxTerms,
	)
	career.Terms = terms

	grantBranchSkillToFirstTerm(r, &career, branch)

	// No grantStartingRankAutoSkillToFirstTerm call: unlike Soldier's own
	// S1 Private and Spacer's own R1 Spacehand, M1 Private has no
	// automatic skill at all (marineRankAutomaticSkill's own doc
	// comment) — there is nothing to grant.

	return career, finalUPP
}
