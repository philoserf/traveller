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
// (step E, p.57), one roll per term served. Reuses
// scoutMusterOutRollCount directly rather than a new
// marineMusterOutRollCount — its own doc comment already frames the
// Disabled-doubling/Dead-zeroing rule as Book 1 p.68's universal one,
// not Scout-specific, and it only reads career.Terms, nothing
// Scout-specific in its body; Marine is a second real, verbatim-
// matching caller, the same "generalize on 2nd instance" discipline
// already applied to resolveRisk/resolveReward/riskOutcome/
// resolveCareerLoop in Phase T1.
//
// dm is p.86's own "DM +Terms +Officer Rank" — Officer Rank is a real,
// reachable outcome since Phase V (marine_promotion.go's Commission/
// Promotion mechanics), read as the numeric tier (O1=1 .. O7=7), the
// same "=Rank" judgment call marineCareerFame's own Officer Rank Fame
// term already makes (no explicit multiplier is printed for this DM
// either).
func ResolveMarineMusterOut(r *dice.Roller, career Career) MusteringOut {
	var out MusteringOut

	dm := len(career.Terms)
	if isOfficer, tier := marineRankState(career.Terms); isOfficer {
		dm += tier
	}

	for range scoutMusterOutRollCount(career) {
		row := musterOutRow(r.D6()+dm, len(marineMusterOutMoney))

		if r.Uniform(2) == 1 {
			out.Money = append(out.Money, marineMusterOutMoney[row])
		} else {
			out.Benefits = append(out.Benefits, marineMusterOutBenefits[row])
		}
	}

	return out
}
