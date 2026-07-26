package character

import "github.com/philoserf/traveller/dice"

// agentCareerFame is Book 1 p.91's own "Agent =Number of Commendations"
// Fame row.
func agentCareerFame(career Career) int {
	return agentCommendationCount(career.Terms)
}

// GenerateAgentCharacter generates a full Human Agent Character end to
// end: a UPP, a homeworld and its background skills, a full multi-term
// Agent career, and Agent's own Mustering Out benefits. Reuses
// buildRiskCareerCharacter directly (character_generate.go) — Agent's
// own Fame/MusterOut both need only career.Terms, with no extra
// parameters the shared signature has no room for, unlike Scholar's own
// Edu or Merchant's own isOfficer/tier.
func GenerateAgentCharacter(r *dice.Roller) (Character, bool) {
	upp := GenerateUPP(r)
	homeworld, homeworldSkills := GenerateHomeworldSkills(r)

	return buildRiskCareerCharacter(r, upp, homeworld, homeworldSkills,
		func(r *dice.Roller, upp UPP, aging *agingSimulation) (Career, UPP) {
			return resolveAgentCareerWithBudget(r, upp, maxCareerTerms, aging)
		}, ResolveAgentMusterOut, agentCareerFame)
}
