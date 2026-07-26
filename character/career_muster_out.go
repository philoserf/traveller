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

// musterOutRow maps a 1D+DM Mustering Out roll to a 0-based row index,
// implementing Book 1 p.68's "If the roll is greater than the maximum
// value on the table, use the maximum value instead" — shared by every
// career's own Mustering Out table (maxRow is that table's own row count;
// Scout's own table has 12 rows, Citizen's 11), the second concrete
// instance of the identical clamp rule that justifies generalizing it
// out of scoutMusterOutRow's own original, Scout-only body. Split out
// dice-free, same rationale as riskOutcome/continueScoutOutcome:
// the clamp boundary is directly testable against a fixed roll instead
// of a real D6 draw.
func musterOutRow(roll, maxRow int) int {
	return min(roll, maxRow) - 1
}

func scoutMusterOutRow(roll int) int {
	return musterOutRow(roll, 12)
}

// rollScoutMusterOutRow rolls 1D and applies dm, per Book 1 p.79's "DM
// +Terms +Fame/2" — see ResolveScoutMusterOut's own doc comment for
// where that Fame comes from.
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
// nothing in this codebase ever populates Character.Commendations, so
// that half is permanently empty and can never fire. Fame is now
// genuinely accumulated within a single Mustering Out resolution (see
// ResolveScoutMusterOut's own doc comment), but reaching 19 within one
// career would need roughly ten separate "Fame +2" rolls — implausible
// enough in practice that the Fame-19+ extra-roll bonus stays deferred
// too, not silently dropped.
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

// resolveRankMusterOut is Marine's/Soldier's/Spacer's own shared Mustering
// Out body (Book 1 p.86/p.82/p.81, step E, p.57) — confirmed byte-identical
// across all three careers except which money/benefits tables and rank-name
// tables get passed in, extracted per this codebase's own "generalize on
// 2nd instance" discipline once a third verbatim match (Spacer) appeared.
//
// dm is each career's own "DM +Terms +Officer Rank" — Officer Rank read as
// the numeric tier (O1=1 .. O7=7), the same "=Rank" judgment call each
// career's own CareerFame Officer Rank term already makes (no explicit
// multiplier is printed for this DM either). money and benefits must be the
// same length (both indexed by the same rolled row).
func resolveRankMusterOut(
	r *dice.Roller,
	career Career,
	money, benefits []string,
	enlistedRankCount, officerRankCount int,
) MusteringOut {
	var out MusteringOut

	dm := len(career.Terms)
	if isOfficer, tier := rankState(career.Terms, enlistedRankCount, officerRankCount); isOfficer {
		dm += tier
	}

	for range scoutMusterOutRollCount(career) {
		row := musterOutRow(r.D6()+dm, len(money))

		if r.Uniform(2) == 1 {
			out.Money = append(out.Money, money[row])
		} else {
			out.Benefits = append(out.Benefits, benefits[row])
		}
	}

	return out
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
// dm = len(career.Terms) + fame/2, per p.79's own "DM +Terms +Fame/2".
//
// fame is a running local accumulator seeded from the Fame this career
// already earned before Mustering Out began — scoutDiscoveryFame's own
// "Discoveries x4" (p.91). It used to start at zero, on the reasoning
// that Character.Fame doesn't exist yet at this point in the pipeline
// (ApplyMusteringOut derives it from this function's own returned
// MusteringOut afterward). True, but beside the point: Discovery Fame is
// a pure function of the finished Career sitting right here, and a
// Scout who made discoveries is famous for them whether or not the
// Character value has been assembled yet. Starting at zero applied
// "+Fame/2" to a Fame the character demonstrably did not have.
//
// Every time a landed Benefits entry itself grants Fame ("Fame +2"),
// it's added to that accumulator immediately (via musterOutFameBonus,
// character/muster_out_apply.go) so a later roll in the same Mustering
// Out sequence sees the correct, already-elevated DM — not the DM of a
// stale, separately-computed final Fame value.
//
// One consequence worth naming: Scout's own table has 12 rows and this
// DM is uncapped, so a long or discovery-rich career saturates it —
// every roll clamps to row 12 via p.68's own "if the roll is greater
// than the maximum value on the table, use the maximum value instead."
// That was already true from +Terms alone for a 12-term Scout; seeding
// Discovery Fame widens the band of characters it applies to (measured:
// Knighthood on 55% of Benefits rolls before, 66% after, across 3,660
// generated Scouts). The saturation is the book's own clamp rule
// operating on the book's own DM, not an artifact of this change. DM is unaffected by
// scoutMusterOutRollCount's own Double-Benefits doubling — p.69 doubles
// the roll count, not the per-roll DM.
//
// Deliberately not implemented, each for a specific documented reason —
// see this package's own chargen plan history for the full rationale:
// Duplicate Benefits rerolling (p.68: optional, "may"), partial DM
// application (p.68: optional, "may"), and the p.67 "TAS Life
// Membership... Scout with 3 Discoveries" Automatic (illustrative referee
// discretion, not a crisp trigger). Mechanical application of
// Characteristic Improvement/Fame/Cash benefits onto a Character is a
// separate function, ApplyMusteringOut (character/muster_out_apply.go) —
// this function's own returned MusteringOut still just records what was
// granted; resolving it is a caller's job, same as before.
func ResolveScoutMusterOut(r *dice.Roller, career Career) MusteringOut {
	var out MusteringOut

	terms := len(career.Terms)
	fame := scoutDiscoveryFame(career)

	for range scoutMusterOutRollCount(career) {
		row := rollScoutMusterOutRow(r, terms+fame/2)

		if r.Uniform(2) == 1 {
			out.Money = append(out.Money, scoutMusterOutMoney[row])
		} else {
			entry := scoutMusterOutBenefits[row]
			out.Benefits = append(out.Benefits, entry)

			if bonus, ok := musterOutFameBonus(entry); ok {
				fame += bonus
			}
		}
	}

	return out
}
