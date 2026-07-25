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
// (step E, p.57), one roll per term served — see resolveRankMusterOut
// (career_muster_out.go) for the shared body, common to Marine, Soldier,
// and Spacer.
func ResolveSpacerMusterOut(r *dice.Roller, career Career) MusteringOut {
	return resolveRankMusterOut(
		r, career,
		spacerMusterOutMoney[:], spacerMusterOutBenefits[:],
		len(spacerEnlistedRankNames), len(spacerOfficerRankNames),
	)
}
