package character

import "github.com/philoserf/traveller/dice"

var agentMusterOutMoney = [12]string{
	"Low Passage", "Low Passage", "Mid Passage", "High Passage",
	"Cr15,000", "StarPass", "Cr25,000", "Cr30,000",
	"Cr35,000", "Cr40,000", "Cr80,000", "Cr90,000",
}

var agentMusterOutBenefits = [12]string{
	"Ship Share", "Forbidden Knowledge", "Wafer Jack", "C5 +1",
	"Str +1", "C2 +1", "C3 +1", "Ship Share",
	"Life Insurance", "TAS Fellow Membership", "Fame +2", "Knighthood",
}

// ResolveAgentMusterOut is Book 1 p.83's own "D MUSTER OUT" table — DM
// is "+Terms +Commends", both derivable from career.Terms alone.
func ResolveAgentMusterOut(r *dice.Roller, career Career) MusteringOut {
	var out MusteringOut

	dm := servedTermCount(career.Terms) + agentCommendationCount(career.Terms)

	for range musterOutRollCount(career, resolveFameStacks(agentCareerFameAwards(career))) {
		appendMusterOutRoll(r, &out, dm, agentMusterOutMoney[:], agentMusterOutBenefits[:])
	}

	return out
}
