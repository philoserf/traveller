package character

import "github.com/philoserf/traveller/dice"

// marineMusterOutMoney/marineMusterOutBenefits are Book 1 p.86's own
// "D MUSTER OUT" table (10 rows — fewer than Scout's 12 or Citizen's
// 11, confirmed directly against the page image). "Retirement x2"
// (rows 8-9, the box's own "Retire x2") is p.70's own "Retirement x 2"
// glossary entry, spelled out; recorded as a literal narrative string
// like every other benefit here, not mechanically resolved into
// MusteringOut.RetirementPay.
var marineMusterOutMoney = [10]string{
	"Low Passage", "Middle Passage", "High Passage", "StarPass",
	"Cr30,000", "Cr40,000", "Cr50,000", "Retirement x2",
	"Retirement x2", "Cr60,000",
}

var marineMusterOutBenefits = [10]string{
	"Forbidden Knowledge", "Str +1", "Wafer Jack", "C5 +1",
	"Int +1", "C2 +1", "Life Insurance", "Ship Share",
	"Directorate", "Knighthood",
}

// ResolveMarineMusterOut resolves Book 1 p.86's own Mustering Out table
// (step E, p.57), one roll per term served — see resolveRankMusterOut
// (career_muster_out.go) for the shared body, common to Marine, Soldier,
// and Spacer.
func ResolveMarineMusterOut(r *dice.Roller, career Career) MusteringOut {
	return resolveRankMusterOut(
		r, career,
		marineMusterOutMoney[:], marineMusterOutBenefits[:],
		len(marineEnlistedRankNames), len(marineOfficerRankNames),
	)
}
