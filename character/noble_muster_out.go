package character

import (
	"fmt"

	"github.com/philoserf/traveller/dice"
)

// nobleMusterOutMoney is Book 1 p.85's own "D MUSTER OUT" table, Money
// column (1D-indexed, row 1 at index 0). "StarPass" keeps the book's own
// literal token (its own "*Upgraded to High Passage as a courtesy"
// footnote is narrative color a GM applies at the table, not a
// mechanical value change — noted here, not separately modeled).
var nobleMusterOutMoney = [12]string{
	"StarPass", "StarPass", "StarPass", "StarPass",
	"Cr100,000", "Cr200,000", "Cr300,000", "Cr400,000",
	"Cr500,000", "Cr600,000", "Cr700,000", "Cr800,000",
}

// nobleMusterOutBenefits is Book 1 p.85's own "D MUSTER OUT" table,
// Benefits column. "Forbidden Knowledge"/"Life Insurance" spelled out
// from the box's own "Forbidden"/"Life Insur", matching the exact
// expansions already established for Scout's identical abbreviations
// (career_muster_out.go's scoutMusterOutBenefits). "TAS Life
// Membership" is Noble's own distinct entitlement tier — the book
// prints "TAS Life" here, not Scout/Citizen's own "TAS Fellow" — verified
// directly against the page image, not a transcription slip. "Cx +1"/
// "Str +1" keep the box's own characteristic-boost shorthand, matching
// every skill table's existing convention.
var nobleMusterOutBenefits = [12]string{
	"Forbidden Knowledge", "Str +1", "Wafer Jack", "C5 +1",
	"Directorship", "C2 +1", "C3 +1", "Int +1",
	"Ship Share", "Life Insurance", "TAS Life Membership", "Directorship",
}

// nobleMusterOutPower is Book 1 p.85's own "Power" column — a third
// Mustering Out column with no analog in Scout's or Citizen's own
// two-column tables. Rows 1-6 carry a fixed Proxy rating; rows 7-12 show
// "Proxy (2D)" — an embedded sub-roll, resolved by
// rollNobleMusterOutPower into concrete text, not stored as a literal
// formula string.
var nobleMusterOutPower = [12]string{
	"Proxy (1)", "Proxy (2)", "Proxy (3)", "Proxy (4)", "Proxy (5)", "Proxy (6)",
	"Proxy (2D)", "Proxy (2D)", "Proxy (2D)", "Proxy (2D)", "Proxy (2D)", "Proxy (2D)",
}

// rollNobleMusterOutPower resolves row's own Power entry into concrete
// text: rows 1-6 return their fixed rating directly; rows 7-12 roll 2D6
// to fill in "Proxy (2D)"'s own embedded sub-roll — the same
// store-the-resolved-value convention every other MusteringOut entry
// already follows (e.g. "CrN,NNN" is stored fully resolved, never as a
// formula).
func rollNobleMusterOutPower(r *dice.Roller, row int) string {
	if nobleMusterOutPower[row] != "Proxy (2D)" {
		return nobleMusterOutPower[row]
	}

	return fmt.Sprintf("Proxy (%d)", r.TwoD6())
}

// ResolveNobleMusterOut resolves Book 1 p.85's own Noble Mustering Out
// table (step E, p.57 — a distinct step following Career Resolution,
// the same convention ResolveScoutMusterOut/ResolveCitizenMusterOut
// already establish). One roll per term served (p.68's own "One Per
// Term" rule), no Dead/Disabled wrinkle — Return & Intrigue has neither
// concept, matching Citizen's own simpler precedent over Scout's more
// involved musterOutRollCount.
//
// Each roll shares one row (musterOutRow, dm = len(career.Terms), p.85's
// own "DM +Total Terms") across all three columns, then independently
// picks Money, Benefits, or Power via a uniform 3-way pick — the book
// gives no mechanic for the choice, the same "genuine open player
// choice, no book-given tiebreaker" reasoning ResolveScoutMusterOut's
// own 2-way column pick already establishes, generalized here from 2
// columns to 3. The book's own table doesn't repeat "DM +Total Terms"
// under Power specifically — read as printing economy, not a genuinely
// independent roll, since nothing else in the box suggests otherwise.
// Power's results go into MusteringOut.Entitlements — an existing field
// no other career has ever populated.
func ResolveNobleMusterOut(r *dice.Roller, career Career) MusteringOut {
	var out MusteringOut

	dm := len(career.Terms)

	for range len(career.Terms) {
		row := musterOutRow(r.D6()+dm, 12)

		switch r.Uniform(3) {
		case 1:
			out.Money = append(out.Money, nobleMusterOutMoney[row])
		case 2:
			out.Benefits = append(out.Benefits, nobleMusterOutBenefits[row])
		default:
			out.Entitlements = append(out.Entitlements, rollNobleMusterOutPower(r, row))
		}
	}

	return out
}
