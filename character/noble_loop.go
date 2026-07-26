package character

import "github.com/philoserf/traveller/dice"

// continueNobleOutcome is continueNoble's own dice-free decision. roll==2
// always succeeds (Book 1's generic "Mandatory Continue," p.65-66 — the
// same override continueScoutOutcome/continueCitizenOutcome both already
// implement). Otherwise defers to succeedsAgainst against p.85's own
// fixed "Continue 7" target.
func continueNobleOutcome(roll int) bool {
	if roll == 2 {
		return true
	}

	return succeedsAgainst(roll, 7, 0)
}

func continueNoble(r *dice.Roller) bool {
	return continueNobleOutcome(r.TwoD6())
}

// ResolveNobleCareer resolves a full multi-term Noble career (Book 1
// p.85) by looping ResolveNobleTerm, mirroring ResolveCitizenCareer's own
// shape (no UPP threading — Return & Intrigue never modifies a
// characteristic, no wound/death mechanic at all, matching Citizen's own
// precedent). Returns a zero-Terms Career when BeginNoble fails (Soc <
// B) — the same "not zero-valued Character, empty Career" precedent
// ResolveScoutCareer/ResolveCitizenCareer already establish. HasRank is
// true regardless — Noble is not in Book 1 p.65's own no-rank career
// list ("the Citizen, Entertainer, Craftsman, Scout, Agent, and Rogue
// careers have no rank") — even though this slice doesn't compute the
// actual rank name (Term.Rank stays unpopulated; see this slice's own
// plan-file Context for why that's deferred).
func ResolveNobleCareer(r *dice.Roller, upp UPP) Career {
	career, _ := resolveNobleCareerAndUPPWithBudget(r, upp, maxCareerTerms, &agingSimulation{})

	return career
}

func resolveNobleCareerAndUPPWithBudget(r *dice.Roller, upp UPP, maxTerms int, aging *agingSimulation) (Career, UPP) {
	career := Career{Name: NobleCareerName, HasRank: true}

	if !BeginNoble(upp.Characteristics[C6]) {
		return career, upp
	}

	usedThisCycle := make(map[Position]bool, len(nobleReturnIntriguePositions))

	for range maxTerms {
		ccPos := nextCC(upp, nobleReturnIntriguePositions, usedThisCycle)

		term := ResolveNobleTerm(r, upp, ccPos, career.Terms)
		upp = applyPersonalAwards(upp, term.SkillsAwarded)
		career.Terms = append(career.Terms, term)

		upp = aging.advanceTerm(r, upp)
		if !aging.alive() {
			break
		}

		if len(career.Terms) == maxTerms {
			break
		}

		if !continueNoble(r) {
			break
		}
	}

	return career, upp
}
