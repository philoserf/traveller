package character

import (
	"slices"

	"github.com/philoserf/traveller/dice"
)

// resolveCraftsmanCareerWithBudget resolves a full multi-term Craftsman
// career (Book 1 p.75) via resolveCareerLoop. Craftsman never modifies a
// characteristic (no Risk & Reward at all), so upp passes through the
// term resolver unchanged — the same shape Rogue's own "Rogue never
// modifies its own CC" already establishes.
//
// Takes ctx directly (not threaded through a shared helper like
// resolveRiskCareerSegment) because BeginCraftsman needs
// ctx.SkillsSoFar — an input no other risk-shaped career's own Begin
// needs.
func resolveCraftsmanCareerWithBudget(r *dice.Roller, upp UPP, maxTerms int, ctx segmentContext) (Career, UPP) {
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
	)
	career.Terms = terms

	return career, finalUPP
}
