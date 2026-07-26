package character

import "github.com/philoserf/traveller/dice"

// scholarMusterOutMoney/scholarMusterOutBenefits are Book 1 p.76's own
// "D MUSTER OUT" table (11 rows — matching Soldier's own row count, not
// Scout's/Rogue's 12).
var scholarMusterOutMoney = [11]string{
	"Low Passage", "Low Passage", "Mid Passage", "High Passage",
	"Cr15,000", "StarPass", "Cr25,000", "Cr30,000",
	"Cr35,000", "Cr40,000", "Cr50,000",
}

var scholarMusterOutBenefits = [11]string{
	"C5 +1", "C5 +1", "Wafer Jack", "Edu +1",
	"Str +1", "C2 +1", "C3 +1", "Int +1",
	"Fame +1", "Ship Share", "TAS Fellow Membership",
}

// ResolveScholarMusterOut takes upp (unlike every other career's
// MusterOut function) because its own DM needs Edu to know Scholar's
// starting rank tier (0 or 1) before scholarRankTier can derive the
// final tier — this is why Scholar gets its own bespoke
// buildScholarCharacter instead of buildRiskCareerCharacter's shared
// (r, career)-only MusterOut signature. DM is Book 1's own "+Scholar
// Rank +Terms". Reuses musterOutRollCount (career_muster_out.go)
// directly for the Double-on-Disabled/Dead-zeroing roll count — already
// career-agnostic.
func ResolveScholarMusterOut(r *dice.Roller, career Career, upp UPP) MusteringOut {
	var out MusteringOut

	startTier := scholarStartTier(int(upp.Characteristics[C5]))

	dm := scholarRankTier(career.Terms, startTier) + len(career.Terms)

	for range musterOutRollCount(career, resolveFameStacks(scholarSegmentFameAwards(upp, career.Terms))) {
		appendMusterOutRoll(r, &out, dm, scholarMusterOutMoney[:], scholarMusterOutBenefits[:])
	}

	// p.70: "A tenured Professor receives a pension of Cr10,000 per
	// year." Tenure gates it, not rank (hasTenure, scholar_generate.go).
	if hasTenure(career.Terms) {
		applyPension(&out, professorPensionRate)
	}

	return out
}
