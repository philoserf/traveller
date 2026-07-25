package character

import "github.com/philoserf/traveller/dice"

// GenerateMarineCharacter generates a full Human Marine Character end to
// end: a UPP, a homeworld and its background skills, a full multi-term
// Marine career, and Marine's own Mustering Out benefits.
//
// Returns (Character, bool) mirroring GenerateScoutCharacter's own
// signature, not GenerateCitizenCharacter's/GenerateNobleCharacter's own
// shape: Marine has both a real "never qualified" outcome (BeginMarine
// can fail) and a real death outcome (RiskResult == Dead), the same two
// reasons Scout's own ok needs both checks.
func GenerateMarineCharacter(r *dice.Roller) (Character, bool) {
	upp := GenerateUPP(r)
	homeworld, homeworldSkills := GenerateHomeworldSkills(r)

	return buildMarineCharacter(r, upp, homeworld, homeworldSkills)
}

// buildMarineCharacter assembles a Character from an already-rolled upp,
// homeworld, and homeworldSkills, mirroring buildScoutCharacter's own
// split for testability. Delegates to buildRiskCareerCharacter
// (character_generate.go) — Marine shares Scout's own shape exactly
// (byte-identical, per golangci-lint's own dupl check), including reuse
// of scoutWoundBadges and scoutMusterOutRollCount (already generic,
// nothing Scout-specific in either body).
func buildMarineCharacter(r *dice.Roller, upp UPP, homeworld string, homeworldSkills []SkillLevel) (Character, bool) {
	return buildRiskCareerCharacter(r, upp, homeworld, homeworldSkills, ResolveMarineCareer, ResolveMarineMusterOut)
}
