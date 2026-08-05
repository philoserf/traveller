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
	return resolveScholarCareerWithBudget(r, upp, maxCareerTerms, &agingSimulation{}, Education{})
}

// resolveScholarCareerWithBudget is ResolveScholarCareer's own body,
// with the resolveCareerLoop term cap threaded as a parameter — see
// resolveCareerLoop's own doc comment for why.
func resolveScholarCareerWithBudget(
	r *dice.Roller, upp UPP, maxTerms int, aging *agingSimulation, education Education,
) (Career, UPP) {
	career := Career{Name: ScholarCareerName, HasRank: true}

	edu := int(upp.Characteristics[C5])

	// p.76's Waivers are per-career: "Mod minus previous waivers
	// (successful or not)", counted across this Scholar career rather
	// than across the character's whole life. Education keeps its own
	// separate count for the same reason.
	waivers := 0

	ok, tier := BeginScholar(r, edu)
	if !ok {
		// "Position" is the first of p.76's six waivable events.
		ok = tryWaiver(r, upp, &waivers)
	}

	if !ok {
		// Reaching here means Edu was below 8 and the roll was taken and
		// lost — Book 1 p.65's own one-year cost. An Edu 8+ Scholar
		// enters automatically without rolling, so never pays it.
		upp = aging.chargeFailedAttempt(r, upp)

		return career, upp
	}

	// Settled after the Begin, not before it. "Every Scholar has a Major
	// and a Minor" (p.76) is about Scholars, and a character who never
	// entered the career is not one — so a failed Position must cost no
	// dice here, the same discipline every conditional roll in this
	// package follows. Costs none either for a graduate, whose subjects
	// come from his degree.
	career.Major, career.Minor = scholarMajorMinor(r, education)

	var priorTerms []Term

	terms, finalUPP := resolveCareerLoop(r, upp, scholarRiskRewardPositions,
		func(r *dice.Roller, upp UPP, ccPos Position) (Term, UPP) {
			var (
				term       Term
				updatedUPP UPP
			)

			term, updatedUPP, tier = ResolveScholarTerm(r, upp, ccPos, int(upp.Characteristics[C5]), tier, priorTerms,
				career.Major, &waivers)
			priorTerms = append(priorTerms, term)

			return term, updatedUPP
		},
		func(r *dice.Roller, upp UPP) bool {
			// Continue is the last of p.76's six waivable events. Edu comes
			// from this closure's own upp, not the outer edu captured
			// before the loop — resolveCareerLoop threads upp forward with
			// any Personal-row Edu boost already applied
			// (applyPersonalAwards, career_loop.go), and both Tenure/
			// Promotion eligibility (ResolveScholarTerm) and this Continue
			// check must see that current value, not the entry-time one.
			// Only BeginScholar's entry gate above wants the stale,
			// pre-career edu.
			return scholarWaivableRoll(r, upp, &waivers,
				continueScholar(r, int(upp.Characteristics[C5]), scholarPublicationsTotal(priorTerms)))
		},
		maxTerms,
		aging,
		nil,
	)
	career.Terms = terms

	return career, finalUPP
}
