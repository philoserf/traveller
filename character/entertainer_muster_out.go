package character

import "github.com/philoserf/traveller/dice"

// entertainerMusterOutMoney/entertainerMusterOutBenefits are Book 1
// p.77's own "D MUSTER OUT BENEFITS" table (13 rows).
var entertainerMusterOutMoney = [13]string{
	"Low Passage", "Low Passage", "Mid Passage", "High Passage",
	"Cr15,000", "StarPass", "Cr25,000", "Cr30,000",
	"Cr35,000", "Cr40,000", "Cr50,000", "Cr400,000", "Cr500,000",
}

var entertainerMusterOutBenefits = [13]string{
	"C5 +1", "C5 +1", "Wafer Jack", "Edu +1",
	"Str +1", "C2 +1", "C3 +1", "Int +1",
	"Fame +1", "Ship Share", "TAS Fellow", "Knighthood", "TAS Life",
}

// ResolveEntertainerMusterOut takes fame (unlike every other career's
// two-argument MusterOut function) because its own DM is "+Fame/3
// +Terms" — Fame isn't derivable from Career.Terms alone the way a rank
// tier is, so it's threaded in directly from ResolveEntertainerCareer's
// own second return value. Reuses scoutMusterOutRollCount
// (career_muster_out.go) directly for the Double-on-Disabled/
// Dead-zeroing roll count — already career-agnostic.
func ResolveEntertainerMusterOut(r *dice.Roller, career Career, fame int) MusteringOut {
	var out MusteringOut

	dm := fame/3 + len(career.Terms)

	for range scoutMusterOutRollCount(career) {
		row := musterOutRow(r.D6()+dm, len(entertainerMusterOutMoney))

		if r.Uniform(2) == 1 {
			out.Money = append(out.Money, entertainerMusterOutMoney[row])
		} else {
			out.Benefits = append(out.Benefits, entertainerMusterOutBenefits[row])
		}
	}

	return out
}
