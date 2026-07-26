package character

import "github.com/philoserf/traveller/dice"

// craftsmanMusterOutMoney/craftsmanMusterOutBenefits are Book 1 p.75's
// own 12-row Muster Out table, transcribed directly from the page
// image. Abbreviations spelled out to their canonical names already
// established by every other career's own muster-out table: "Forbidden
// K" -> "Forbidden Knowledge", "Low/Mid/High Psg" -> "Low/Mid/High
// Passage" (Agent's own precedent), "TAS Fellow" -> "TAS Fellow
// Membership" (career_muster_out.go's own doc comment), "TAS Life" ->
// "TAS Life Membership" (Noble's own precedent — a distinct benefit
// from "TAS Fellow," not the same one abbreviated differently).
var craftsmanMusterOutMoney = [12]string{
	"Low Passage", "Low Passage", "Mid Passage", "High Passage",
	"Cr15,000", "StarPass", "Cr25,000", "Cr30,000",
	"Cr35,000", "Cr40,000", "Cr50,000", "Cr60,000",
}

var craftsmanMusterOutBenefits = [12]string{
	"Forbidden Knowledge", "Forbidden Knowledge", "Wafer Jack", "C5 +1",
	"Str +1", "C2 +1", "C3 +1", "Int +1",
	"Ship Share", "TAS Fellow Membership", "Director", "TAS Life Membership",
}

// ResolveCraftsmanMusterOut resolves Book 1 p.75's own Muster Out
// table. DM = Terms only — no "+Officer Rank" boilerplate component on
// this career's own box (unlike Functionary's). Safely reuses
// musterOutRollCount directly: Craftsman's own RiskResult is
// always Unharmed (ResolveCraftsmanTerm's own doc comment — no Risk &
// Reward at all), so the shared helper's own "double on Disabled"
// branch can never spuriously fire here the way it would have for
// Functionary.
func ResolveCraftsmanMusterOut(r *dice.Roller, career Career) MusteringOut {
	var out MusteringOut

	dm := len(career.Terms)

	for range musterOutRollCount(career, craftsmanCareerFame(career.Terms)) {
		appendMusterOutRoll(r, &out, dm, craftsmanMusterOutMoney[:], craftsmanMusterOutBenefits[:])
	}

	return out
}
