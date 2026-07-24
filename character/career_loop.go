package character

import "github.com/philoserf/traveller/dice"

// scoutRiskRewardPositions is Scout's own Risk & Reward Controlling
// Characteristic set (Book 1 p.79: "Risk & Reward C1 C2 C3") — C4 Int is
// deliberately excluded even though the generic per-career summary table
// elsewhere (p.64's "5 Scout C1 C2 C3 C4" row) lists a fourth entry; Scout's
// own dedicated box (p.79) is the authoritative, career-specific source and
// reserves C4 Int exclusively for the Continue check (p.79's separate
// "Continue Int" line).
var scoutRiskRewardPositions = []Position{C1, C2, C3}

// nextScoutCC picks the Risk & Reward Controlling Characteristic for one
// term and records it in usedThisCycle, implementing Book 1 p.65's generic
// rule: the player picks any not-yet-used Characteristic from the career's
// own set; it "cannot be used again until all of the others in the sequence
// have been used." Scout's CC therefore rotates through C1/C2/C3 rather than
// staying fixed for the whole career — that fixed-for-career behavior is
// Rogue's own explicitly-marked "*Special Case" (p.83: a Rogue's Controlling
// Characteristic "is then used throughout his career, not just in the
// current Term"), a carve-out that would be redundant if every career
// already worked that way. Which still-available characteristic to pick is,
// like BeginScout's own Begin-target choice, an open choice Book 1 leaves to
// the player; resolved here via highestOf (same rationale as BeginScout's
// own doc comment: a genuinely better option exists, so a uniform random
// pick would be a worse default). usedThisCycle resets once all three have
// been used, starting a fresh three-term cycle.
func nextScoutCC(upp UPP, usedThisCycle map[Position]bool) Position {
	if len(usedThisCycle) == len(scoutRiskRewardPositions) {
		clear(usedThisCycle)
	}

	var available []Position

	for _, pos := range scoutRiskRewardPositions {
		if !usedThisCycle[pos] {
			available = append(available, pos)
		}
	}

	pos := highestOf(upp, available...)
	usedThisCycle[pos] = true

	return pos
}

// continueScoutOutcome is continueScout's own dice-free decision, split out
// the same way succeedsAgainst/scoutRiskOutcome are so the boundary is
// directly testable against a fixed roll instead of a real 2D6 draw. roll==2
// always succeeds (Book 1's "Mandatory Continue"); otherwise roll<=intChar.
func continueScoutOutcome(roll, intChar int) bool {
	if roll == 2 {
		return true
	}

	return roll <= intChar
}

// continueScout resolves Book 1's end-of-Term Continue check (p.65-66: "the
// Character must successfully roll (2D) to Continue (or less) in the
// career... If the Continue roll is 2 exactly, the character is required to
// Continue") against Scout's own p.79 Continue target, Int (C4) — unlike
// Agent's "Continue Str*" or Rogue's "Continue CC*" (each footnoted "*Mod
// +Terms"), Scout's own box has no asterisk and no +Terms Mod: the target is
// bare Int every term. Note: GenerateUPP only ever produces characteristics
// in [2,12] (a 2D6 roll) and Scout's own Risk & Reward set
// (scoutRiskRewardPositions) never includes C4, so Int can never fall below 2
// for any character this codebase produces — the roll==2 mandatory-continue
// branch is implemented because it's the literal rule, but is currently
// unreachable as an actual behavior change (2<=Int already covers it).
func continueScout(r *dice.Roller, upp UPP) bool {
	return continueScoutOutcome(r.TwoD6(), int(upp.Characteristics[C4]))
}

// maxScoutTerms is a defensive engineering cap, not a Book 1 rule —
// grounded in Book 1 p.89's traditional human lifespan (74 years) minus the
// traditional adventuring start (Young Adult, 18), divided into 4-year
// terms: (74-18)/4 = 14. A career hitting this cap is not a normal outcome
// Book 1 describes — it guards against an immortal-Courier-Duty Scout
// looping indefinitely, since Aging (which would otherwise end this
// naturally) isn't modeled yet.
const maxScoutTerms = 14

// ResolveScoutCareer resolves a full multi-term Scout career (Book 1 p.79)
// by looping ResolveScoutTerm. BeginScout's one-time "To Begin" check runs
// first; if it fails, this returns Career{Name: "Scout"} with a nil Terms
// slice and upp unchanged — a real, legitimate "never qualified" outcome
// per Book 1's own "If both Begin and Retry fail, this career may not be
// used," not an error. Otherwise, each term: pick this term's Controlling
// Characteristic (nextScoutCC, rotating per p.65), resolve the term
// (ResolveScoutTerm, which also returns an updated UPP carrying forward any
// p.65 "permanently reduced" characteristic value), append the Term,
// replace the working UPP with the updated one, then stop if the character
// didn't survive (RiskResult.Survived(), false only on Dead), was left
// Disabled (p.65: "the Character is disabled and must Muster Out at the
// end of the Term" — mandatory regardless of the Continue roll; the actual
// Double Benefits mustering-out payout itself is Muster-Out's own future
// work, deferred alongside it below), or failed to Continue; otherwise
// loop, capped defensively at maxScoutTerms.
//
// Returns the final UPP alongside the Career — not just the Career — so a
// caller building a full Character can persist any permanent reduction
// (RiskResult Wounded or Disabled) onto Character.UPP; returning only the
// Career would silently lose it one call frame up, reintroducing the exact
// p.65 persistence gap this function exists to fix.
//
// BeginScout's own returned Position is not threaded into the first term's
// ccPos — nextScoutCC's first-ever call (against an empty usedThisCycle)
// independently computes the identical value, both being
// highestOf(upp, C1, C2, C3) over the same untouched characteristic set — so
// term 1's CC naturally coincides with BeginScout's own pick.
//
// Deferred, same treatment as Aging/Muster-Out: Book 1 p.79's own "reduce
// San= -1 for each TWO Terms served" Sanity rule — meaningless to apply
// since GenerateUPP never generates Sanity in the first place (p.57:
// "created only as needed").
func ResolveScoutCareer(r *dice.Roller, upp UPP) (Career, UPP) {
	career := Career{Name: "Scout"}

	if _, began := BeginScout(r, upp); !began {
		return career, upp
	}

	usedThisCycle := make(map[Position]bool, len(scoutRiskRewardPositions))

	for range maxScoutTerms {
		ccPos := nextScoutCC(upp, usedThisCycle)

		term, updatedUPP, survived := ResolveScoutTerm(r, upp, ccPos)
		career.Terms = append(career.Terms, term)
		upp = updatedUPP

		if !survived || term.RiskResult == Disabled || !continueScout(r, upp) {
			break
		}
	}

	return career, upp
}
