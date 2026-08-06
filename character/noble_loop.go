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
	return mandatoryContinueOutcome(roll, 7)
}

func continueNoble(r *dice.Roller) bool {
	return continueNobleOutcome(r.TwoD6())
}

// applyNobleElevation handles Book 1 p.65's own Elevation consequence:
// "raised... to the next higher Noble rank and its associated increase
// in Social Standing (if any)". Applied here rather than in
// ResolveNobleTerm because the ladder position has to persist across
// terms — Soc alone cannot carry it, since three pairs of ranks share a
// Soc value. Returns the land grant p.85 awards for the Soc increase
// itself ("Each increase in Soc during CharGen awards a Land Grant"),
// or nil — keyed on the Soc rising, not on the elevation, since a
// Baronet raised to Baron gains a title at the same Soc and so no new
// fief. Extracted out of resolveNobleCareerAndUPPWithBudget's own loop
// body to keep it under the cognitive-complexity budget once the Later
// Education branch (#164) sat beside it.
func applyNobleElevation(r *dice.Roller, upp UPP, rankIndex int) (int, UPP, *LandGrant) {
	next, socRose := nobleRankAfterElevation(rankIndex)
	rankIndex = next
	soc := nobleRanks[rankIndex].Soc
	upp.Characteristics[C6] = soc

	if !socRose {
		return rankIndex, upp, nil
	}

	if grant, ok := newNobleLandGrant(soc, world.Generate(r)); ok {
		return rankIndex, upp, &grant
	}

	return rankIndex, upp, nil
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
	career, _, _ := resolveNobleCareerAndUPPWithBudget(r, upp, maxCareerTerms, &agingSimulation{}, nil)

	return career
}

// education is #164's own wiring — nil for callers with no Education
// context, mirroring resolveRogueCareerAndUPPWithBudget's own education
// parameter (rogue_loop.go). Noble has no resolveCareerLoop to hand this
// to, so the laterEducationHook check is inlined below instead.
func resolveNobleCareerAndUPPWithBudget(
	r *dice.Roller, upp UPP, maxTerms int, aging *agingSimulation, education *Education,
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

	hook := laterEducationHook(education)

	usedThisCycle := make(map[Position]bool, len(nobleReturnIntriguePositions))

	for range maxTerms {
		if !aging.alive() {
			break
		}

		var (
			term    Term
			elected bool
		)

		if hook != nil {
			term, upp, elected = hook(r, upp)
		}

		if !elected {
			ccPos := nextCC(upp, nobleReturnIntriguePositions, usedThisCycle)
			term = ResolveNobleTerm(r, upp, ccPos, career.Terms)

			// Never true for a Later Education term (elected above) —
			// Return & Intrigue didn't run, so there's nothing to elevate
			// on.
			if term.Elevated {
				var grant *LandGrant

				rankIndex, upp, grant = applyNobleElevation(r, upp, rankIndex)
				if grant != nil {
					grants = append(grants, *grant)
				}
			}
		}

		// Set unconditionally, including on a Later Education term: the
		// character still holds whatever rank the ladder last reached, and
		// lastTermRank (character_generate.go) reads only the final term's
		// Rank — a career ending on a Later Education term would otherwise
		// report an empty Character.Rank despite legitimately holding a
		// title.
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
