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
	return buildRiskCareerCharacter(
		r, upp, homeworld, homeworldSkills, ResolveMarineCareer, ResolveMarineMusterOut, marineCareerFame)
}

// marineCareerFame is Book 1 p.91's own Armed Forces Fame bracket
// (Army/Marine/Navy). Since Phase V, Officer Rank is a real, reachable
// outcome (rankState/ResolveMarineTerm), so the bracket's own "Officer
// Rank*" line (*Armed Forces Enlisted = no Fame) genuinely applies now —
// read as "=Rank" (the numeric tier, O1=1 .. O7=7), matching the same
// "=Rank" formula shape p.91 already uses for Scholar and Merchant; no
// explicit multiplier is printed for Armed Forces the way Medals/Noble
// each get one, so this is a documented judgment call, not a
// directly-quoted formula. The bracket's nested Medal/Wound-Badge Fame
// values (Exemplary Service XS x0, Wound Badge WB x1, MCUF x1, MCG x2,
// SEH x3) are read as scoped only to the "Rank*" line by the asterisk's
// own placement, not to the whole bracket — Medals and Wound Badges are
// earned with no Officer gate anywhere in the Risk & Reward mechanic, so
// their Fame applies regardless of rank. See the plan-file history for
// the full reasoning. The shared formula body lives in
// rankBasedCareerFame (career_rank.go), common to Marine, Soldier, and
// Spacer.
func marineCareerFame(career Career) int {
	return rankBasedCareerFame(career, len(marineEnlistedRankNames), len(marineOfficerRankNames))
}
