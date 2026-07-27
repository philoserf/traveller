package character

import "github.com/philoserf/traveller/dice"

// GenerateSpacerCharacter generates a full Human Spacer Character end to
// end: a UPP, a homeworld and its background skills, a full multi-term
// Spacer career, and Spacer's own Mustering Out benefits.
//
// Returns (Character, bool) mirroring GenerateSoldierCharacter's own
// signature exactly, for the same reasons: BeginSpacer can fail (a real
// "never qualified" outcome) and RiskResult can be Dead (a real death
// outcome).
func GenerateSpacerCharacter(r *dice.Roller) (Character, bool) {
	upp, homeworld, homeworldSkills, education := generateStart(r)

	c, ok := buildSpacerCharacter(r, upp, homeworld, homeworldSkills)
	c = applyEducation(c, education)

	return c, ok
}

// buildSpacerCharacter assembles a Character from an already-rolled upp,
// homeworld, and homeworldSkills, mirroring buildSoldierCharacter's own
// split for testability. Delegates to buildRiskCareerCharacter
// (character_generate.go) — Spacer shares Marine's/Soldier's own shape
// exactly.
func buildSpacerCharacter(r *dice.Roller, upp UPP, homeworld string, homeworldSkills []SkillLevel) (Character, bool) {
	return buildRiskCareerCharacter(
		r, upp, homeworld, homeworldSkills, func(r *dice.Roller, upp UPP, aging *agingSimulation) (Career, UPP) {
			return resolveSpacerCareerWithBudget(r, upp, maxCareerTerms, aging)
		}, ResolveSpacerMusterOut, spacerCareerFameAwards)
}

// spacerCareerFameAwards mirrors soldierCareerFameAwards's own formula exactly
// (character/soldier_character_generate.go) — Medal Fame + Wound Badge
// Fame + Officer Rank Fame (=Rank, the numeric tier) — since Spacer
// shares the same "Army/Marine/Navy: Officer Rank*" Fame bracket (Book 1
// p.91) Marine and Soldier do. See marineCareerFameAwards's own doc comment
// for the full reasoning. The shared formula body lives in
// rankBasedCareerFame (career_rank.go).
func spacerCareerFameAwards(career Career) []int {
	return rankBasedCareerFameAwards(career, len(spacerEnlistedRankNames), len(spacerOfficerRankNames))
}
