package character

import "github.com/philoserf/traveller/dice"

// continueAgent targets Str, Mod +Terms (same-term-inclusive, the
// established Eneri Dinsha precedent) — no natural-roll exception
// documented, matching Scholar's/Entertainer's/Merchant's own
// plain-roll Continue.
func continueAgent(r *dice.Roller, str, mod int) bool {
	return rollAgainstTarget(r, str, mod)
}

// ResolveAgentCareer resolves a full multi-term Agent career (Book 1
// p.83).
func ResolveAgentCareer(r *dice.Roller, upp UPP) (Career, UPP) {
	return resolveAgentCareerWithBudget(r, upp, maxCareerTerms, &agingSimulation{})
}

// resolveAgentCareerWithBudget is ResolveAgentCareer's own body, with
// the resolveCareerLoop term cap threaded as a parameter — see
// resolveCareerLoop's own doc comment for why.
func resolveAgentCareerWithBudget(r *dice.Roller, upp UPP, maxTerms int, aging *agingSimulation) (Career, UPP) {
	career := Career{Name: AgentCareerName}

	if !BeginAgent(r, int(upp.Characteristics[C3])) {
		// Book 1 p.65: a failed Begin roll costs a year. Careers whose
		// Begin is a prerequisite rather than a roll (Noble's Soc B+,
		// Craftsman's held skills, Citizen's automatic entry) charge
		// nothing here — there was no attempt to fail.
		upp = aging.chargeFailedAttempt(r, upp)

		return career, upp
	}

	var priorTerms []Term

	terms, finalUPP := resolveCareerLoop(r, upp, agentRiskRewardPositions,
		func(r *dice.Roller, upp UPP, ccPos Position) (Term, UPP) {
			term, updatedUPP := ResolveAgentTerm(r, upp, ccPos)
			priorTerms = append(priorTerms, term)

			return term, updatedUPP
		},
		func(r *dice.Roller, upp UPP) bool {
			return continueAgent(r, int(upp.Characteristics[C1]), len(priorTerms))
		},
		maxTerms,
		aging,
	)
	career.Terms = terms

	return career, finalUPP
}
