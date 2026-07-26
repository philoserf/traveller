package character

import (
	"fmt"

	"github.com/philoserf/traveller/dice"
)

// functionaryMusterOutMoney/functionaryMusterOutBenefits are Book 1
// p.87's own 11-row Muster Out table — one row longer than the more
// common 10/12-row tables elsewhere, transcribed directly from the page
// image. Abbreviations spelled out to their canonical names already
// established by every other career's own muster-out table: "Forbidden
// K" -> "Forbidden Knowledge", "Life Insur" -> "Life Insurance", "TAS
// Fellow" -> "TAS Fellow Membership" (career_muster_out.go's own doc
// comment), "High Psg" -> "High Passage" (Agent's own precedent).
var functionaryMusterOutMoney = [11]string{
	"Cr5,000", "Cr10,000", "Cr15,000", "Cr20,000",
	"High Passage", "StarPass", "Cr25,000", "Cr30,000",
	"Cr35,000", "Pension x2", "Pension x2",
}

var functionaryMusterOutBenefits = [11]string{
	"Forbidden Knowledge", "Str +1", "Wafer Jack", "Str +1",
	"C2 +1", "C3 +1", "Int +1", "Life Insurance",
	"TAS Fellow Membership", "Knighthood", "Directorship",
}

// functionaryMusterOutRollCount mirrors musterOutRollCount's own
// shape (career_muster_out.go) but does NOT double the roll count on a
// Disabled last term. Functionary reuses RiskResult.Disabled to mean
// "Office Politics failed, career ends" (ResolveFunctionaryTerm's own
// doc comment), not the universal p.65 "reduced 4+" physical disability
// the Double Benefits rule is actually about — Office Politics has no
// equivalent at all (no Mods, no characteristic reduction), so doubling
// here would be a real bug, not a faithful reuse of the shared helper.
func functionaryMusterOutRollCount(career Career) int {
	return len(career.Terms)
}

// ResolveFunctionaryMusterOut resolves Book 1 p.87's own Muster Out
// table. finalTier is Functionary's own final Rank tier (0-8), needed
// for "Automatic: Directorship if Rank F6+" — resolveFunctionaryCareerWithBudget
// widens its own return for exactly this, the same "widen the return
// for a real second need" precedent Merchant's/Scholar's own tier
// returns already established. DM = Terms; the box's own "+Officer
// Rank" DM component doesn't map to anything Functionary tracks (no
// Officer/Enlisted distinction at any tier) and is treated as a no-op —
// the same boilerplate annotation appears verbatim on the Armed Forces
// careers' own muster-out tables.
func ResolveFunctionaryMusterOut(r *dice.Roller, career Career, finalTier int) MusteringOut {
	var out MusteringOut

	dm := len(career.Terms)

	for range functionaryMusterOutRollCount(career) {
		appendMusterOutRoll(r, &out, dm, functionaryMusterOutMoney[:], functionaryMusterOutBenefits[:])
	}

	if len(career.Terms) > 0 {
		out.Automatics = append(out.Automatics, fmt.Sprintf("Gold Watch (Cr%d)", 100*len(career.Terms)))

		if finalTier >= 6 {
			out.Automatics = append(out.Automatics, "Directorship")
		}
	}

	return out
}
