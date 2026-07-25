package character

import "github.com/philoserf/traveller/dice"

// GenerateSoldierCharacter generates a full Human Soldier Character end
// to end: a UPP, a homeworld and its background skills, a full
// multi-term Soldier career, and Soldier's own Mustering Out benefits.
//
// Returns (Character, bool) mirroring GenerateMarineCharacter's own
// signature exactly, for the same reasons: BeginSoldier can fail (a
// real "never qualified" outcome) and RiskResult can be Dead (a real
// death outcome).
func GenerateSoldierCharacter(r *dice.Roller) (Character, bool) {
	upp := GenerateUPP(r)
	homeworld, homeworldSkills := GenerateHomeworldSkills(r)

	return buildSoldierCharacter(r, upp, homeworld, homeworldSkills)
}

// buildSoldierCharacter assembles a Character from an already-rolled
// upp, homeworld, and homeworldSkills, mirroring buildMarineCharacter's
// own split for testability. Delegates to buildRiskCareerCharacter
// (character_generate.go) — Soldier shares Marine's own shape exactly.
func buildSoldierCharacter(r *dice.Roller, upp UPP, homeworld string, homeworldSkills []SkillLevel) (Character, bool) {
	return buildRiskCareerCharacter(
		r, upp, homeworld, homeworldSkills, ResolveSoldierCareer, ResolveSoldierMusterOut, soldierCareerFame)
}

// soldierCareerFame mirrors marineCareerFame's own corrected formula
// exactly (character/marine_character_generate.go) — Medal Fame + Wound
// Badge Fame + Officer Rank Fame (=Rank, the numeric tier) — since
// Soldier shares the same "Army/Marine/Navy: Officer Rank*" Fame bracket
// (Book 1 p.91) Marine does. See that function's own doc comment for
// the full reasoning.
func soldierCareerFame(career Career) int {
	fame := 0

	for _, t := range career.Terms {
		for _, medal := range t.Medals {
			fame += medalFame[medal]
		}
	}

	fame += scoutWoundBadges(career) // Wound Badge Fame, x1 each, p.91

	if isOfficer, tier := rankState(
		career.Terms,
		len(soldierEnlistedRankNames),
		len(soldierOfficerRankNames),
	); isOfficer {
		fame += tier // Officer Rank Fame, read as =Rank (the numeric tier)
	}

	return fame
}
