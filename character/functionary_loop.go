package character

import "github.com/philoserf/traveller/dice"

// resolveFunctionaryCareerWithBudget resolves a full multi-term
// Functionary career (Book 1 p.87) via resolveCareerLoop. Office
// Politics has no separate Continue roll at all (see
// ResolveFunctionaryTerm's own doc comment) — its own continueCareer
// callback always returns true, since the Disabled reuse already
// unconditionally stops resolveCareerLoop whenever Office Politics'
// Risk fails. Office Politics never modifies a characteristic either
// (no Mods, no reduction), so upp passes through the term resolver
// unchanged — the same shape Rogue's own "Rogue never modifies its own
// CC" already establishes.
//
// Unlike every chainable career reachable as a first entry, this one
// takes ctx directly (not threaded through a shared helper like
// resolveRiskCareerSegment) because BeginFunctionary and the F6 Rank
// title both need ctx.TermsServedSoFar/ctx.PrecedingCareer — inputs no
// other career's own Begin or Rank logic needs.
//
// Returns the final tier alongside (Career, UPP) — ResolveFunctionaryMusterOut
// needs it (Automatic: Directorship if Rank F6+), the same "widen the
// return for a real second need" precedent Merchant's/Scholar's own
// (isOfficer, tier)/tier returns already established.
func resolveFunctionaryCareerWithBudget(
	r *dice.Roller,
	upp UPP,
	maxTerms int,
	ctx segmentContext,
) (Career, UPP, int) {
	career := Career{Name: FunctionaryCareerName, HasRank: true}

	if !BeginFunctionary(r, ctx.TermsServedSoFar) {
		return career, upp, 0
	}

	tier, nthUnderSecretary := 0, 0

	terms, finalUPP := resolveCareerLoop(r, upp, functionaryRiskRewardPositions,
		func(r *dice.Roller, upp UPP, ccPos Position) (Term, UPP) {
			term, newTier, newNth := ResolveFunctionaryTerm(r, upp, ccPos, tier, nthUnderSecretary, ctx.PrecedingCareer)
			tier, nthUnderSecretary = newTier, newNth

			return term, upp
		},
		func(_ *dice.Roller, _ UPP) bool { return true },
		maxTerms,
	)
	career.Terms = terms

	// F0's own Auto Skill (Bureaucrat) is never "promoted into" — every
	// Functionary starts there — so it's applied once to term 1 after
	// the loop completes, mirroring Marine's own one-time Branch-skill
	// grant (marine_loop.go).
	if len(career.Terms) > 0 {
		career.Terms[0].SkillsAwarded = append(
			career.Terms[0].SkillsAwarded, skillLevel1(functionaryTierAutoSkills[0], Skill))
	}

	return career, finalUPP, tier
}
