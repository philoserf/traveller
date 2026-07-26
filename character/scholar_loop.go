package character

import "github.com/philoserf/traveller/dice"

// continueScholar has no natural-roll exception (unlike Marine's
// natural-2 or Rogue's natural-12) — confirmed directly against the box
// and the Master Checklist, both of which describe a plain roll.
func continueScholar(r *dice.Roller, edu, mod int) bool {
	return rollAgainstTarget(r, edu, mod)
}

// ResolveScholarCareer resolves a full multi-term Scholar career (Book 1
// p.76). priorTerms accumulates via closure capture (the same pattern
// established for Rogue's own termCount, career.go's own Terms not
// being available until resolveCareerLoop returns) since
// scholarPublicationsTotal needs terms-so-far for both the term
// resolver's own Promotion/Tenure Mod and the Continue check's own Mod.
func ResolveScholarCareer(r *dice.Roller, upp UPP) (Career, UPP) {
	return resolveScholarCareerWithBudget(r, upp, maxCareerTerms, &agingSimulation{})
}

// resolveScholarCareerWithBudget is ResolveScholarCareer's own body,
// with the resolveCareerLoop term cap threaded as a parameter — see
// resolveCareerLoop's own doc comment for why.
func resolveScholarCareerWithBudget(r *dice.Roller, upp UPP, maxTerms int, aging *agingSimulation) (Career, UPP) {
	career := Career{Name: ScholarCareerName, HasRank: true}

	edu := int(upp.Characteristics[C5])

	ok, tier := BeginScholar(r, edu)
	if !ok {
		// Reaching here means Edu was below 8 and the roll was taken and
		// lost — Book 1 p.65's own one-year cost. An Edu 8+ Scholar
		// enters automatically without rolling, so never pays it.
		upp = aging.chargeFailedAttempts(r, upp, 1)

		return career, upp
	}

	var priorTerms []Term

	terms, finalUPP := resolveCareerLoop(r, upp, scholarRiskRewardPositions,
		func(r *dice.Roller, upp UPP, ccPos Position) (Term, UPP) {
			var (
				term       Term
				updatedUPP UPP
			)

			term, updatedUPP, tier = ResolveScholarTerm(r, upp, ccPos, edu, tier, priorTerms)
			priorTerms = append(priorTerms, term)

			return term, updatedUPP
		},
		func(r *dice.Roller, _ UPP) bool {
			return continueScholar(r, edu, scholarPublicationsTotal(priorTerms))
		},
		maxTerms,
		aging,
	)
	career.Terms = terms

	return career, finalUPP
}
