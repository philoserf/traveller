package character

import "github.com/philoserf/traveller/dice"

// citizenLifePositions is Citizen Life's own Controlling Characteristic
// set (Book 1 p.78: "Citizen Life C1 C2 C3 C4") — one more than Scout's
// own C1 C2 C3 (career_loop.go's scoutRiskRewardPositions).
var citizenLifePositions = []Position{C1, C2, C3, C4}

// continueCitizenOutcome is continueCitizen's own dice-free decision,
// split out the same way continueScoutOutcome is. roll==2 always
// succeeds (Book 1's generic "Mandatory Continue," p.65-66) — kept as an
// explicit branch even though it's provably unreachable here (roll==2 is
// already covered by succeedsAgainst's own roll<=10 check below), the
// same deliberate choice continueScoutOutcome's own doc comment already
// explains: implementing the literal rule, not just its currently-observable
// effect. Otherwise defers to succeedsAgainst against p.78's fixed
// "Continue 10-" target — unlike Scout's characteristic-based Continue
// Int, this never varies by UPP.
func continueCitizenOutcome(roll int) bool {
	if roll == 2 {
		return true
	}

	return succeedsAgainst(roll, 10, 0)
}

func continueCitizen(r *dice.Roller) bool {
	return continueCitizenOutcome(r.TwoD6())
}

// citizenLifeSuccessCount counts CitizenLifeSucceeded terms — derived
// from terms rather than a separately-tracked counter, so it can't drift
// out of sync with the Terms it's counting.
func citizenLifeSuccessCount(terms []Term) int {
	n := 0

	for _, t := range terms {
		if t.CitizenLifeSucceeded {
			n++
		}
	}

	return n
}

// ResolveCitizenCareer resolves a full multi-term Citizen career (Book 1
// p.78) by looping ResolveCitizenTerm. BeginCitizen is called for
// symmetry with ResolveScoutCareer's own Begin-then-loop shape even
// though its own "To Begin Auto" rule (BeginCitizen's own doc comment)
// makes the zero-terms branch below provably unreachable today — kept
// visible rather than silently assumed, the same reasoning BeginCitizen
// itself exists as a real function for at all.
//
// Each term: pick this term's Controlling Characteristic (nextCC,
// rotating through citizenLifePositions per p.65), resolve the term,
// append it, then stop if Continue fails; otherwise loop, capped
// defensively at maxCareerTerms. No UPP is threaded or returned: Citizen
// Life never reduces a characteristic (no wound mechanic at all, unlike
// Scout's Risk & Reward), and Table C's own Personal-column boosts are
// recorded only as SkillLevel entries, never mechanically applied to
// UPP.Characteristics — the same treatment Scout's identical Personal
// grants already get, not a new gap. upp therefore never changes across
// a Citizen career; ResolveCitizenCareer only reads it, for nextCC's own
// highestOf tie-breaking.
func ResolveCitizenCareer(r *dice.Roller, upp UPP) Career {
	career := Career{Name: CitizenCareerName}

	if !BeginCitizen() {
		return career
	}

	usedThisCycle := make(map[Position]bool, len(citizenLifePositions))

	for range maxCareerTerms {
		ccPos := nextCC(upp, citizenLifePositions, usedThisCycle)

		term, jobSkill, hobbySkill := ResolveCitizenTerm(
			r, upp, ccPos, citizenLifeSuccessCount(career.Terms), career.JobSkill, career.HobbySkill)
		career.Terms = append(career.Terms, term)
		career.JobSkill, career.HobbySkill = jobSkill, hobbySkill

		if !continueCitizen(r) {
			break
		}
	}

	return career
}
