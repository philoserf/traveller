package character

import (
	"strings"

	"github.com/philoserf/traveller/dice"
)

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

// musterOutCommendations counts Commendations awarded across a career's
// terms. Agent is the only career that grants them (its Reward result is
// formatted "<Undercover Career> Commendation-N", agent_generate.go), so
// this reports 0 for every other career without needing to know that.
func musterOutCommendations(career Career) int {
	n := 0

	for _, t := range career.Terms {
		if strings.Contains(t.RewardResult, "Commendation") {
			n++
		}
	}

	return n
}

// musterOutExtraRollMedals counts the medals that earn an extra Mustering
// Out roll. Book 1 p.68 names exactly two — "one additional roll per
// Commendation, MCG, or SEH" — and deliberately not XS or MCUF, which the
// Armed Forces careers award far more freely: every surviving Risk roll
// grants an XS. Counting those would roughly double the rolls of any
// long Armed Forces career.
func musterOutExtraRollMedals(career Career) int {
	n := 0

	for _, t := range career.Terms {
		for _, medal := range t.Medals {
			if medal == "MCG" || medal == "SEH" {
				n++
			}
		}
	}

	return n
}

// musterOutRollCount is Book 1 p.68's own roll budget: "A character is
// allowed one Mustering Out roll for each term served in Career
// Resolution. He is allowed one additional roll per Commendation, MCG, or
// SEH. He is allowed one additional roll if Fame 19+."
//
// Dice-free so the budget is directly testable against a fixed fixture.
// Returns 0 for a career that never qualified (nil Terms) and, per p.69's
// "Dying During Character Generation" ("all efforts in this particular
// character creation process are lost"), for one whose last Term ended in
// Death — a dead character does not reach Mustering Out at all, not
// merely "with no benefits."
//
// The Disability doubling (p.69: "Muster Out at Term End with Double
// Benefits... twice the count of Benefits") applies to the finished
// total, not to the per-term part alone: it doubles the count of
// Benefits, and by then the extra rolls are part of that count.
//
// fame is the character's Fame as known when this career musters out.
func musterOutRollCount(career Career, fame int) int {
	if len(career.Terms) == 0 {
		return 0
	}

	last := career.Terms[len(career.Terms)-1]
	if last.RiskResult == Dead {
		return 0
	}

	rolls := len(career.Terms) + musterOutCommendations(career) + musterOutExtraRollMedals(career)
	if fame >= fameExtraRollThreshold {
		rolls++
	}

	if last.RiskResult == Disabled {
		return rolls * 2
	}

	return rolls
}

// fameExtraRollThreshold is p.68's own "one additional roll if Fame 19+".
const fameExtraRollThreshold = 19

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

	// Armed Forces Fame is Medals + Wound Badges + Officer Rank
	// (rankBasedCareerFame) — the same value the career itself reports,
	// computed here so p.68's Fame-19+ roll can be tested against it.
	fame := rankBasedCareerFame(career, enlistedRankCount, officerRankCount)

	for range musterOutRollCount(career, fame) {
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
// Loops musterOutRollCount(career) times. Each roll independently
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
// musterOutRollCount's own Double-Benefits doubling — p.69 doubles
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

	// The same Discovery Fame that seeds the DM also decides p.68's
	// Fame-19+ extra roll. Evaluated once, before the loop: a "Fame +2"
	// landing mid-sequence raises the DM for later rolls (below) but
	// cannot retroactively grant another roll.
	for range musterOutRollCount(career, fame) {
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
