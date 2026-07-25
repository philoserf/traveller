package character

import "github.com/philoserf/traveller/dice"

// continueMerchant targets Str, no stated Mod or natural-roll exception
// (matching Scholar's/Entertainer's own plain-roll Continue).
func continueMerchant(r *dice.Roller, str int) bool {
	return rollAgainstTarget(r, str, 0)
}

// ResolveMerchantCareer resolves a full multi-term Merchant career
// (Book 1 p.80). BeginMerchant never fails (the only career this
// session with no "never qualified" outcome), so there is no early
// exit — the loop always runs at least once. Returns (Career, UPP,
// isOfficer, tier) — the final UPP (carrying forward any Risk-reduced
// characteristic from a survived Wounded/Disabled term, the same
// UPP-threading discipline established for Scout/Marine/Scholar; a
// code-review pass caught an earlier version of this function
// discarding resolveCareerLoop's own returned UPP entirely) and the
// final rank state, needed by ResolveMerchantMusterOut's own DM and
// merchantCareerFame. (isOfficer, tier) are threaded directly through
// the term-resolution closure (the same closure-accumulation shape
// established for Rogue's own termCount and Scholar's own tier) rather
// than recomputed via a Marine-family-style rankState — Merchant's own
// variable starting state isn't a fit for that shared function (see
// this slice's own plan-file Context).
func ResolveMerchantCareer(r *dice.Roller, upp UPP) (Career, UPP, bool, int) {
	return resolveMerchantCareerWithBudget(r, upp, maxCareerTerms)
}

// resolveMerchantCareerWithBudget is ResolveMerchantCareer's own body,
// with the resolveCareerLoop term cap threaded as a parameter — see
// resolveCareerLoop's own doc comment for why.
func resolveMerchantCareerWithBudget(r *dice.Roller, upp UPP, maxTerms int) (Career, UPP, bool, int) {
	career := Career{Name: MerchantCareerName, HasRank: true}

	isOfficer, tier := BeginMerchant(r, upp)

	var priorTerms []Term

	terms, finalUPP := resolveCareerLoop(r, upp, merchantRiskRewardPositions,
		func(r *dice.Roller, upp UPP, ccPos Position) (Term, UPP) {
			var (
				term       Term
				updatedUPP UPP
			)

			term, updatedUPP, isOfficer, tier = ResolveMerchantTerm(r, upp, ccPos, isOfficer, tier, priorTerms)
			priorTerms = append(priorTerms, term)

			return term, updatedUPP
		},
		func(r *dice.Roller, upp UPP) bool {
			return continueMerchant(r, int(upp.Characteristics[C1]))
		},
		maxTerms,
	)
	career.Terms = terms

	return career, finalUPP, isOfficer, tier
}
