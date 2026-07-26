package character

import "github.com/philoserf/traveller/dice"

// continueRogue mirrors continueMarineOutcome's own natural-2 override,
// but Rogue's own Continue check additionally uses the same natural-12-
// always-fails exception as rogueSucceeds instead of plain
// succeedsAgainst — Book 1's "Continue CC*" shares its own "*Mod +Terms.
// But, 12 is always automatic failure" footnote with Begin/Risk/Reward.
func continueRogue(r *dice.Roller, cc, mod int) bool {
	roll := r.TwoD6()

	switch roll {
	case 2:
		return true
	case 12:
		return false
	default:
		return succeedsAgainst(roll, cc, mod)
	}
}

// ResolveRogueCareer resolves a full multi-term Rogue career (Book 1
// p.84). The fixed CC is selected once via rollRogueCC and passed to
// resolveCareerLoop as a single-element positions slice — nextCC
// (career_loop.go) already returns the same element every call when
// len(positions)==1 (usedThisCycle fills to length 1 and clears every
// call), so "fixed for the whole career, not rotating" falls out of the
// existing shared loop for free, no new generalization needed.
//
// resolveCareerLoop's own RiskResult.Survived()/Disabled stop checks key
// off Term.RiskResult, which Rogue's own terms never set (stays at its
// zero value, Unharmed) — both checks pass through as "keep going" for
// every Rogue term. Confirmed by inspection, not just assumed: Rogue has
// no death/disability concept at all, so this is the correct behavior,
// not a coincidental side effect.
func ResolveRogueCareer(r *dice.Roller, upp UPP) Career {
	career, _ := resolveRogueCareerAndUPPWithBudget(r, upp, maxCareerTerms)

	return career
}

// resolveRogueCareerWithBudget is ResolveRogueCareer's own body, with
// the resolveCareerLoop term cap threaded as a parameter — see
// resolveCareerLoop's own doc comment for why.
func resolveRogueCareerAndUPPWithBudget(r *dice.Roller, upp UPP, maxTerms int) (Career, UPP) {
	career := Career{Name: RogueCareerName}

	ccPos := rollRogueCC(r)
	cc := int(upp.Characteristics[ccPos])

	if !BeginRogue(r, cc) {
		return career, upp
	}

	// termCount tracks "+Terms" via local closure state, not
	// career.Terms — resolveCareerLoop builds its own local slice and
	// only assigns it to career.Terms after the whole loop returns, so
	// reading career.Terms from inside either closure would always see
	// zero terms. The same reason Marine/Soldier/Spacer accumulate their
	// own priorTerms locally instead of reading back through Career.
	termCount := 0

	terms, finalUPP := resolveCareerLoop(r, upp, []Position{ccPos},
		func(r *dice.Roller, upp UPP, ccPos Position) (Term, UPP) {
			term := ResolveRogueTerm(r, int(upp.Characteristics[ccPos]), termCount) // "+Terms"
			term.ControllingCharacteristic = ccPos
			termCount++

			return term, upp // Rogue never modifies its own CC
		},
		func(r *dice.Roller, currentUPP UPP) bool {
			return continueRogue(r, int(currentUPP.Characteristics[ccPos]), termCount)
		},
		maxTerms,
	)
	career.Terms = terms

	return career, finalUPP
}
