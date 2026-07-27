package character

import (
	"fmt"

	"github.com/philoserf/traveller/dice"
)

// soldierEnlistedRankNames/soldierOfficerRankNames are Book 1 p.82's own
// "Enlisted Soldier Ranks"/"Officer Soldier Ranks" tables (S1-S6, O1-O7)
// — genuinely distinct names from Marine's own tables (confirmed
// directly, not assumed identical).
var soldierEnlistedRankNames = [6]string{
	"Private", "Corporal", "Sergeant",
	"Staff Sergeant", "Master Sergeant", "Sergeant Major",
}

var soldierOfficerRankNames = [7]string{
	"2nd Lieutenant", "1st Lieutenant", "Captain", "Major",
	"Lt Colonel", "Colonel", "General",
}

// *Command College triggers in Year 1 of the next Term upon reaching O4
// Major (if Continue). Modeled — see command_college.go. A Soldier draws
// his two skills from the Military Academy column, being the Army.

// soldierRankName formats a rank state as Book 1's own "S1 Private"/"O4
// Major" notation.
func soldierRankName(isOfficer bool, tier int) string {
	if isOfficer {
		return fmt.Sprintf("O%d %s", tier, soldierOfficerRankNames[tier-1])
	}

	return fmt.Sprintf("S%d %s", tier, soldierEnlistedRankNames[tier-1])
}

// soldierRankAutomaticSkill is Book 1 p.82's own "Automatic Skills by
// Rank" — genuinely distinct data from Marine's own table (confirmed
// directly): Soldier has an entry Marine doesn't (S1 Private -> Fighter)
// and its own O6 Colonel -> Leader, which Marine's table has no
// equivalent for. Same one-time-on-transition semantics as
// marineRankAutomaticSkill — see that function's own doc comment.
func soldierRankAutomaticSkill(isOfficer bool, tier int) (SkillLevel, bool) {
	switch {
	case !isOfficer && tier == 1:
		return skillLevel1("Fighter", Skill), true
	case !isOfficer && tier == 3:
		return skillLevel1("Heavy Weapons", Skill), true
	case !isOfficer && tier == 4:
		return skillLevel1("Leader", Skill), true
	case isOfficer && tier == 1:
		return skillLevel1("Leader", Skill), true
	case isOfficer && tier == 4:
		return skillLevel1("Tactics", Skill), true
	case isOfficer && tier == 6:
		return skillLevel1("Leader", Skill), true
	default:
		return SkillLevel{}, false
	}
}

// rollSoldierCommission/rollSoldierOfficerPromotion/rollSoldierEnlistedPromotion
// are Book 1 p.82's own "Officer Commission C3" (no Mod — no asterisk on
// this line), "Officer Promotion Soc*", and "Enlisted Promotion C3*"
// ("*+Medals and WB Mods" — this codebase applies Medal Mods only, not
// Wound Badges; see the plan-file history for why). Officer Promotion
// targets Soc, confirmed directly against Soldier's own dedicated box —
// agreeing with the Master Checklist this time (Marine was the real
// outlier there, not a book-wide error).
func rollSoldierCommission(r *dice.Roller, c3 int) bool {
	return rollAgainstTarget(r, c3, 0)
}

func rollSoldierOfficerPromotion(r *dice.Roller, soc, medalMod int) bool {
	return rollAgainstTarget(r, soc, medalMod)
}

func rollSoldierEnlistedPromotion(r *dice.Roller, c3, medalMod int) bool {
	return rollAgainstTarget(r, c3, medalMod)
}
