package character

import "github.com/philoserf/traveller/dice"

// spacerMusterOutMoney is Book 1 p.81's own "D MUSTER OUT" table's Money
// column (11 rows — not Marine's/Soldier's own 10).
var spacerMusterOutMoney = [11]string{
	"Low Passage", "Middle Passage", "High Passage", "StarPass",
	"Cr30,000", "Cr40,000", "Cr50,000", "Retirement x2",
	"Retirement x2", "Cr60,000", "Cr70,000",
}

// spacerMusterOutBenefits is Book 1 p.81's own Benefits column. Row 2
// and row 5 both print "Str +1" — a genuine duplicate confirmed across
// two independent page-image reads plus the .txt OCR extraction, not a
// transcription error introduced here.
var spacerMusterOutBenefits = [11]string{
	"Forbidden Knowledge", "Str +1", "Wafer Jack", "C5 +1",
	"Str +1", "C2 +1", "C3 +1", "Int +1",
	"Ship Share", "Life Insurance", "Knighthood",
}

// ResolveSpacerMusterOut resolves Book 1 p.81's own Mustering Out table
// (step E, p.57), one roll per term served — mirrors
// ResolveSoldierMusterOut's own body exactly (soldier_muster_out.go),
// substituting Spacer's own 11-row tables and rank-tier counts.
func ResolveSpacerMusterOut(r *dice.Roller, career Career) MusteringOut {
	var out MusteringOut

	dm := len(career.Terms)
	if isOfficer, tier := rankState(
		career.Terms,
		len(spacerEnlistedRankNames),
		len(spacerOfficerRankNames),
	); isOfficer {
		dm += tier
	}

	for range scoutMusterOutRollCount(career) {
		row := musterOutRow(r.D6()+dm, len(spacerMusterOutMoney))

		if r.Uniform(2) == 1 {
			out.Money = append(out.Money, spacerMusterOutMoney[row])
		} else {
			out.Benefits = append(out.Benefits, spacerMusterOutBenefits[row])
		}
	}

	return out
}
