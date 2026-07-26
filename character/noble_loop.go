package character

import (
	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/world"
)

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
// careers have no rank"). Term.Rank now carries the p.88 title held
// after each term; the Land Grants that come with those titles are
// discarded here, since this entry point returns only a Career —
// buildNobleCharacter calls the budgeted form directly to keep them.
func ResolveNobleCareer(r *dice.Roller, upp UPP) Career {
	career, _, _ := resolveNobleCareerAndUPPWithBudget(r, upp, maxCareerTerms, &agingSimulation{})

	return career
}

func resolveNobleCareerAndUPPWithBudget(
	r *dice.Roller, upp UPP, maxTerms int, aging *agingSimulation,
) (Career, UPP, []LandGrant) {
	career := Career{Name: NobleCareerName, HasRank: true}

	if !BeginNoble(upp.Characteristics[C6]) {
		return career, upp, nil
	}

	// Book 1 p.65: "Nobles begin with rank equal to their Social
	// Standing." BeginNoble already gates on Soc B+, so a rank always
	// exists here.
	rankIndex, _ := nobleRankIndexForSoc(upp.Characteristics[C6])

	var grants []LandGrant

	usedThisCycle := make(map[Position]bool, len(nobleReturnIntriguePositions))

	for range maxTerms {
		if !aging.alive() {
			break
		}

		ccPos := nextCC(upp, nobleReturnIntriguePositions, usedThisCycle)

		term := ResolveNobleTerm(r, upp, ccPos, career.Terms)

		// Book 1 p.65: Elevation raises the character "to the next higher
		// Noble rank and its associated increase in Social Standing (if
		// any)". Applied here rather than in ResolveNobleTerm because the
		// ladder position has to persist across terms — Soc alone cannot
		// carry it, since three pairs of ranks share a Soc value.
		if term.Elevated {
			next, socRose := nobleRankAfterElevation(rankIndex)
			rankIndex = next
			soc := nobleRanks[rankIndex].Soc
			upp.Characteristics[C6] = soc

			// p.85: "Land Grants. Each increase in Soc during CharGen
			// awards a Land Grant." Keyed on the Soc increase, not on the
			// elevation — a Baronet raised to Baron gains a title at the
			// same Soc and so no new fief.
			if socRose {
				if grant, ok := newNobleLandGrant(soc, world.Generate(r)); ok {
					grants = append(grants, grant)
				}
			}
		}

		term.Rank = nobleRanks[rankIndex].Title
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

	return career, upp, grants
}
