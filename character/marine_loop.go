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
// Enlisted may select a new Branch upon Promotion," and Promotion is
// deferred to a follow-up slice, so no re-selection path exists yet.
//
// HasRank is true — Marine is not in Book 1 p.65's own no-rank career
// list — even though Term.Rank itself stays unpopulated until Officer/
// Enlisted Promotion is implemented: a real, documented, temporary gap,
// not "correctly zero" the way Scout's own no-rank careers are.
func ResolveMarineCareer(r *dice.Roller, upp UPP) (Career, UPP) {
	career := Career{Name: MarineCareerName, HasRank: true}

	if !BeginMarine(r, int(upp.Characteristics[C1])) {
		return career, upp
	}

	branch, branchMod := rollMarineBranch(r)

	terms, finalUPP := resolveCareerLoop(r, upp, marineRiskRewardPositions,
		func(r *dice.Roller, upp UPP, ccPos Position) (Term, UPP) {
			return ResolveMarineTerm(r, upp, ccPos, branch, branchMod)
		},
		continueMarine,
	)
	career.Terms = terms

	return career, finalUPP
}
