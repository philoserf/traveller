package character

import "github.com/philoserf/traveller/dice"

// rogueMusterOutMoney/rogueMusterOutBenefits are Book 1 p.84's own "D
// MUSTER OUT" table (12 rows — matching Scout's own row count, not
// Marine's/Soldier's/Spacer's 10/10/11).
var rogueMusterOutMoney = [12]string{
	"Cr40,000", "StarPass", "StarPass", "High Passage",
	"High Passage", "StarPass", "Cr25,000", "Cr30,000",
	"Cr35,000", "Cr40,000", "Cr50,000", "Cr90,000",
}

var rogueMusterOutBenefits = [12]string{
	"Str +1", "C5 +1", "Wafer Jack", "C2 +1",
	"C3 +1", "Life Insurance", "Ship Share", "Knighthood",
	"Ship Share", "Ship Share", "Ship Share", "Knighthood",
}

// ResolveRogueMusterOut resolves Book 1 p.84's own Mustering Out table
// (step E, p.57), one roll per term served — mirrors
// ResolveNobleMusterOut's own simpler shape (character/noble_muster_out.go:
// len(career.Terms) directly), not Scout's/Marine's own
// musterOutRollCount, since Rogue has no Dead/Disabled wrinkle at
// all (no death or disability concept exists for this career). DM is
// p.84's own "+Total Terms" on both columns.
func ResolveRogueMusterOut(r *dice.Roller, career Career) MusteringOut {
	var out MusteringOut

	dm := len(career.Terms)

	for range len(career.Terms) {
		row := musterOutRow(r.D6()+dm, len(rogueMusterOutMoney))

		if r.Uniform(2) == 1 {
			out.Money = append(out.Money, rogueMusterOutMoney[row])
		} else {
			out.Benefits = append(out.Benefits, rogueMusterOutBenefits[row])
		}
	}

	return out
}
