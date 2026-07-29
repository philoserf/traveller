package character

import (
	"slices"

	"github.com/philoserf/traveller/dice"
)

// GenerateSoldierCharacter generates a full Human Soldier Character end
// to end: a UPP, a homeworld and its background skills, a full
// multi-term Soldier career, and Soldier's own Mustering Out benefits.
//
// Returns (Character, bool) mirroring GenerateMarineCharacter's own
// signature exactly, for the same reasons: BeginSoldier can fail (a
// real "never qualified" outcome) and RiskResult can be Dead (a real
// death outcome).
func GenerateSoldierCharacter(r *dice.Roller) (Character, bool) {
	upp, homeworld, homeworldSkills, education := generateStart(r)

	commissioned := slices.Contains(education.CommissionCareers, SoldierCareerName)
	flightSchool := commissioned && education.Honors

	c, ok := buildSoldierCharacter(r, upp, homeworld, homeworldSkills, commissioned, flightSchool)
	c = applyEducation(c, education)

	return c, ok
}

// buildSoldierCharacter assembles a Character from an already-rolled
// upp, homeworld, and homeworldSkills, mirroring buildMarineCharacter's
// own split for testability. Delegates to buildRiskCareerCharacter
// (character_generate.go) — Soldier shares Marine's own shape exactly.
//
// commissioned/flightSchool are #113's Service Academy/OTC Commission
// and Flight School — see resolveSoldierCareerWithBudget's own doc
// comment (soldier_loop.go).
func buildSoldierCharacter(
	r *dice.Roller, upp UPP, homeworld string, homeworldSkills []SkillLevel, commissioned, flightSchool bool,
) (Character, bool) {
	return buildRiskCareerCharacter(
		r, upp, homeworld, homeworldSkills, func(r *dice.Roller, upp UPP, aging *agingSimulation) (Career, UPP) {
			return resolveSoldierCareerWithBudget(r, upp, maxCareerTerms, aging, commissioned, flightSchool)
		}, ResolveSoldierMusterOut, soldierCareerFameAwards)
}

// soldierCareerFameAwards mirrors marineCareerFameAwards's own corrected formula
// exactly (character/marine_character_generate.go) — Medal Fame + Wound
// Badge Fame + Officer Rank Fame (=Rank, the numeric tier) — since
// Soldier shares the same "Army/Marine/Navy: Officer Rank*" Fame bracket
// (Book 1 p.91) Marine does. See that function's own doc comment for
// the full reasoning. The shared formula body lives in
// rankBasedCareerFame (career_rank.go).
func soldierCareerFameAwards(career Career) []int {
	return rankBasedCareerFameAwards(career, len(soldierEnlistedRankNames), len(soldierOfficerRankNames))
}
