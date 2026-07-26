package character

import (
	"slices"

	"github.com/philoserf/traveller/dice"
)

// resolveCraftsmanCareerWithBudget resolves a full multi-term Craftsman
// career (Book 1 p.75) via resolveCareerLoop. Craftsman has no injury
// reduction, but the shared loop applies Personal characteristic awards.
//
// Takes ctx directly (not threaded through a shared helper like
// resolveRiskCareerSegment) because BeginCraftsman needs
// ctx.SkillsSoFar — an input no other risk-shaped career's own Begin
// needs.
func resolveCraftsmanCareerWithBudget(
	r *dice.Roller,
	upp UPP,
	maxTerms int,
	ctx segmentContext,
	aging *agingSimulation,
) (Career, UPP) {
	career := Career{Name: CraftsmanCareerName}

	if !BeginCraftsman(ctx.SkillsSoFar) {
		return career, upp
	}

	heldSkills := slices.Clone(ctx.SkillsSoFar)

	terms, finalUPP := resolveCareerLoop(r, upp, craftsmanRiskRewardPositions,
		func(r *dice.Roller, upp UPP, ccPos Position) (Term, UPP) {
			term, newHeld := ResolveCraftsmanTerm(r, upp, ccPos, heldSkills)
			heldSkills = newHeld

			return term, upp
		},
		func(r *dice.Roller, _ UPP) bool {
			return continueCraftsman(r, craftsmanSkillLevel(heldSkills))
		},
		maxTerms,
		aging,
	)
	career.Terms = terms

	return career, finalUPP
}
