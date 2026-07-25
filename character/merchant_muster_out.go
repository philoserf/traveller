package character

import "github.com/philoserf/traveller/dice"

// merchantMusterOutMoney/merchantMusterOutBenefits are Book 1 p.80's own
// "D MUSTER OUT" table, re-verified directly against the page image
// during a code-review pass (a finder flagged these as suspiciously
// identical to rogueMusterOutMoney/rogueMusterOutBenefits,
// character/rogue_muster_out.go) — confirmed as a genuine coincidence,
// not a copy-paste error: Book 1 p.80's own table is byte-for-byte the
// same as p.84's own Rogue table, verified row by row against the page
// image, not assumed from the earlier transcription.
var merchantMusterOutMoney = [12]string{
	"Cr40,000", "StarPass", "StarPass", "High Passage",
	"High Passage", "StarPass", "Cr25,000", "Cr30,000",
	"Cr35,000", "Cr40,000", "Cr50,000", "Cr90,000",
}

var merchantMusterOutBenefits = [12]string{
	"Str +1", "C5 +1", "Wafer Jack", "C2 +1",
	"C3 +1", "Life Insurance", "Ship Share", "Knighthood",
	"Ship Share", "Ship Share", "Ship Share", "Knighthood",
}

// ResolveMerchantMusterOut mirrors ResolveMarineMusterOut's own body
// (marine_muster_out.go) exactly — dm := len(career.Terms), +tier if
// Officer — taking the final (isOfficer, tier) directly from
// ResolveMerchantCareer's own return instead of recomputing via
// rankState (Merchant's own variable start config isn't a fit for it).
//
// A code-review pass noted this roll-loop/column-split body (D6+dm ->
// musterOutRow -> Uniform(2) column pick) is now duplicated verbatim
// across all eight existing MusterOut functions, past the "generalize
// on 2nd instance" threshold this codebase otherwise follows —
// deliberately not extracted here: doing so would mean touching all
// eight existing career files as a side effect of the Merchant slice,
// well outside this slice's own scope. Left as a candidate for its own
// dedicated cleanup phase.
func ResolveMerchantMusterOut(r *dice.Roller, career Career, isOfficerAtEnd bool, tierAtEnd int) MusteringOut {
	var out MusteringOut

	dm := len(career.Terms)
	if isOfficerAtEnd {
		dm += tierAtEnd
	}

	for range scoutMusterOutRollCount(career) {
		row := musterOutRow(r.D6()+dm, len(merchantMusterOutMoney))

		if r.Uniform(2) == 1 {
			out.Money = append(out.Money, merchantMusterOutMoney[row])
		} else {
			out.Benefits = append(out.Benefits, merchantMusterOutBenefits[row])
		}
	}

	return out
}
