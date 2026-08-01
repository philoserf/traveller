package character

import (
	"github.com/philoserf/traveller/dice"
)

// spacerEnlistedRankNames/spacerOfficerRankNames are Book 1 p.81's own
// "Enlisted Naval Ranks"/"Officer Naval Ranks" tables (R1-R6, O1-O7) —
// genuinely distinct names from Marine's/Soldier's own tables.
var spacerEnlistedRankNames = [6]string{
	"Spacehand", "Able Spacer", "Petty Officer Second",
	"Petty Officer First", "Chief Petty Officer", "Master Chief Petty Officer",
}

var spacerOfficerRankNames = [7]string{
	"Ensign", "Sublieutenant", "Lieutenant", "Lt Commander",
	"Commander", "Captain", "Admiral",
}

// *Command College triggers in Year 1 of the next Term upon reaching O4
// Lt Commander (if Continue). Modeled — see command_college.go. A Spacer
// draws his two skills from the Naval Academy column, as does a Marine
// (p.61 commissions Marine officers through the Naval Academy).

// spacerRankName formats a rank state as Book 1's own "R1 Spacehand"/"O4
// Lt Commander" notation — "R" (Rating) for Enlisted, matching the
// table's own "Enlisted Naval Ranks" column header convention, not "S"
// like Soldier's own.
func spacerRankName(isOfficer bool, tier int) string {
	return rankName(isOfficer, tier, "R", spacerEnlistedRankNames[:], spacerOfficerRankNames[:])
}

// spacerEnlistedRankAutoSkills/spacerOfficerRankAutoSkills back
// spacerRankAutomaticSkill — a map-based lookup rather than
// marineRankAutomaticSkill's/soldierRankAutomaticSkill's own switch
// statement, since Spacer's own 7 defined tiers (one more than either)
// pushed a flat switch over golangci-lint's own cyclomatic complexity
// limit.
var spacerEnlistedRankAutoSkills = map[int]string{1: "Fighter", 4: "Gunner", 5: "Sensors"}

var spacerOfficerRankAutoSkills = map[int]string{1: "Astrogator", 3: "Engineer", 4: "Pilot", 6: "Leader"}

// spacerRankAutomaticSkill is Book 1 p.81's own "Automatic Skills by
// Rank" — genuinely distinct data from Marine's/Soldier's own tables,
// same one-time-on-transition semantics as marineRankAutomaticSkill/
// soldierRankAutomaticSkill.
func spacerRankAutomaticSkill(isOfficer bool, tier int) (SkillLevel, bool) {
	return rankAutoSkillFromTables(spacerEnlistedRankAutoSkills, spacerOfficerRankAutoSkills, isOfficer, tier)
}

// rollSpacerCommission/rollSpacerOfficerPromotion/rollSpacerRatingPromotion
// are Book 1 p.81's own "Officer Commission C2" (no Mod — no asterisk on
// this line), "Officer Promotion Soc*", and "Rating Promotion C2*"
// (Spacer's own name for Enlisted Promotion — "*+Medals and WB Mods";
// this codebase applies Medal Mods only, not Wound Badges, matching
// Marine's/Soldier's own treatment). Officer Commission and Rating
// Promotion share C2 (Dex) as their target, the same "Commission and
// Enlisted-Promotion share a target" shape Soldier's own C3 already
// established.
func rollSpacerCommission(r *dice.Roller, c2 int) bool {
	return rollAgainstTarget(r, c2, 0)
}

func rollSpacerOfficerPromotion(r *dice.Roller, soc, medalMod int) bool {
	return rollAgainstTarget(r, soc, medalMod)
}

func rollSpacerRatingPromotion(r *dice.Roller, c2, medalMod int) bool {
	return rollAgainstTarget(r, c2, medalMod)
}
