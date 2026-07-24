package character

import "github.com/philoserf/traveller/dice"

// scoutMusterOutMoney is Book 1 p.79's Scout "D MUSTER OUT" table, Money
// column (1D-indexed, row 1 at index 0). Names are Book 1's own spelled-out
// prose forms (p.68), not the box's abbreviated column text — e.g. "Middle
// Passage" not "Mid Psg" — matching homeworldSkillByTradeCode's own
// doc-comment precedent (p.56's "Hostile Env" -> "Hostile Environment").
var scoutMusterOutMoney = [12]string{
	"Low Passage", "Middle Passage", "High Passage", "StarPass",
	"Cr30,000", "Cr40,000", "Cr50,000", "Cr60,000",
	"Cr60,000", "Cr60,000", "Cr70,000", "Cr80,000",
}

// scoutMusterOutBenefits is Book 1 p.79's Scout "D MUSTER OUT" table,
// Benefits column (1D-indexed, row 1 at index 0). "Cx +1"/"Str +1" keep the
// box's own characteristic-boost shorthand, matching scoutSkillTable's
// existing convention; other entries use Book 1's spelled-out prose forms
// (p.67-68) — "Forbidden Knowledge" not "Forbidden K", "Life Insurance" not
// "Life Insur", "TAS Fellow Membership" not "TAS Fellow" (p.68: "A TAS
// Fellow Membership is a temporary membership in the Travellers' Aid
// Society").
var scoutMusterOutBenefits = [12]string{
	"Ship Share", "Forbidden Knowledge", "Wafer Jack", "C5 +1",
	"Str +1", "C2 +1", "C3 +1", "Ship Share",
	"Life Insurance", "TAS Fellow Membership", "Fame +2", "Knighthood",
}

// scoutMusterOutRow maps a 1D+DM Mustering Out roll to a 0-based row index
// into scoutMusterOutMoney/scoutMusterOutBenefits, implementing Book 1
// p.68's "If the roll is greater than the maximum value on the table, use
// the maximum value instead." Split out dice-free, same rationale as
// scoutRiskOutcome/continueScoutOutcome: the clamp boundary is directly
// testable against a fixed roll instead of a real D6 draw.
func scoutMusterOutRow(roll int) int {
	return min(roll, 12) - 1
}

// rollScoutMusterOutRow rolls 1D and applies dm, per Book 1 p.79's "DM
// +Terms +Fame/2" (Fame/2 omitted — see ResolveScoutMusterOut's own doc
// comment).
func rollScoutMusterOutRow(r *dice.Roller, dm int) int {
	return scoutMusterOutRow(r.D6() + dm)
}

// scoutMusterOutRollCount is Book 1 p.68's "One Per Term" rule (one
// Mustering Out roll per Term served), applied to a Career, dice-free so
// it's directly testable against a fixed fixture. Returns 0 for a career
// that never qualified (nil Terms) and, per p.69's "Dying During Character
// Generation" ("all efforts in this particular character creation process
// are lost"), for a career whose last Term ended in Death — a dead
// character does not reach Mustering Out at all, not merely "with no
// benefits." Otherwise returns len(career.Terms), doubled if the last Term
// left the character Disabled (p.69's Disability Muster Out: "Muster Out at
// Term End with Double Benefits... twice the count of Benefits" — the roll
// *count* doubles, not each individual benefit).
//
// Commendation/MCG/SEH and Fame-19+ extra-roll sources (p.68) are omitted:
// nothing in this codebase ever populates Character.Commendations or
// Character.Fame, so both are permanently 0/empty and these bonuses can
// never actually fire.
func scoutMusterOutRollCount(career Career) int {
	if len(career.Terms) == 0 {
		return 0
	}

	last := career.Terms[len(career.Terms)-1]
	if last.RiskResult == Dead {
		return 0
	}

	terms := len(career.Terms)
	if last.RiskResult == Disabled {
		return terms * 2
	}

	return terms
}

// ResolveScoutMusterOut resolves Book 1 p.79's Scout Mustering Out table
// (step E, p.57 — a distinct step following Career Resolution; see this
// file's own package context for why it stays a separate function from
// ResolveScoutCareer). Not called from ResolveScoutCareer; a caller invokes
// this explicitly once a Career is final.
//
// Loops scoutMusterOutRollCount(career) times. Each roll independently
// picks the Money or Benefits column via a uniform random pick (p.68:
// "Character may select either the Money column or Benefits column for
// each roll" — a genuine open player choice with no book-given mechanic and
// no objectively-better column, resolved the same way rollScoutDuty
// resolves Courier-vs-Explorer Duty), then rolls rollScoutMusterOutRow with
// dm = len(career.Terms) (p.79's "DM +Terms"; the table's own "+Fame/2" is
// omitted — Character.Fame is never populated anywhere in this codebase, so
// Fame/2 is always 0 in practice; a real, separate future gap, not silently
// dropped). DM is unaffected by scoutMusterOutRollCount's own Double-
// Benefits doubling — p.69 doubles the roll count, not the per-roll DM.
//
// Deliberately not implemented, each for a specific documented reason —
// see this package's own chargen plan history for the full rationale:
// Duplicate Benefits rerolling (p.68: optional, "may"), partial DM
// application (p.68: optional, "may"), the p.67 "TAS Life Membership...
// Scout with 3 Discoveries" Automatic (illustrative referee discretion, not
// a crisp trigger), and mechanical application of Characteristic
// Improvement/Cash benefits onto a Character (MusteringOut's own []string
// shape is for recording what was granted, not resolving it).
func ResolveScoutMusterOut(r *dice.Roller, career Career) MusteringOut {
	var out MusteringOut

	dm := len(career.Terms)

	for range scoutMusterOutRollCount(career) {
		row := rollScoutMusterOutRow(r, dm)

		if r.Uniform(2) == 1 {
			out.Money = append(out.Money, scoutMusterOutMoney[row])
		} else {
			out.Benefits = append(out.Benefits, scoutMusterOutBenefits[row])
		}
	}

	return out
}
