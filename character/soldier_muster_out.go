package character

import "github.com/philoserf/traveller/dice"

// soldierMusterOutMoney is Book 1 p.82's own "D MUSTER OUT" table's
// Money column (10 rows) — confirmed identical to Marine's own Money
// column, not assumed so.
var soldierMusterOutMoney = [10]string{
	"Low Passage", "Middle Passage", "High Passage", "StarPass",
	"Cr30,000", "Cr40,000", "Cr50,000", "Retirement x2",
	"Retirement x2", "Cr60,000",
}

// soldierMusterOutBenefits is Book 1 p.82's own Benefits column —
// genuinely distinct from Marine's own from row 7 onward (C3 +1, Life
// Insurance, TAS Fellow Membership, Knighthood vs Marine's Life
// Insurance, Ship Share, Directorate, Knighthood), confirmed directly
// against the page image.
var soldierMusterOutBenefits = [10]string{
	"Forbidden Knowledge", "Str +1", "Wafer Jack", "C5 +1",
	"Int +1", "C2 +1", "C3 +1", "Life Insurance",
	"TAS Fellow Membership", "Knighthood",
}

// ResolveSoldierMusterOut resolves Book 1 p.82's own Mustering Out table
// (step E, p.57), one roll per term served — see resolveRankMusterOut
// (career_muster_out.go) for the shared body, common to Marine, Soldier,
// and Spacer.
func ResolveSoldierMusterOut(r *dice.Roller, career Career) MusteringOut {
	return resolveRankMusterOut(
		r, career,
		soldierMusterOutMoney[:], soldierMusterOutBenefits[:],
		len(soldierEnlistedRankNames), len(soldierOfficerRankNames),
	)
}
