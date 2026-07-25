package character

import (
	"fmt"

	"github.com/philoserf/traveller/dice"
)

// marineEnlistedRankNames/marineOfficerRankNames are Book 1 p.86's own
// "Enlisted Marine Ranks"/"Officer Marine Ranks" tables (M1-M6, O1-O7).
var marineEnlistedRankNames = [6]string{
	"Private", "Lance Corporal", "Sergeant",
	"Staff Sergeant", "Master Sergeant", "Sergeant Major",
}

var marineOfficerRankNames = [7]string{
	"2nd Lieutenant", "1st Lieutenant", "Captain", "Force Commander",
	"Lt Coronel", "Coronel", "Brigadier",
}

// *Command College triggers in Year 1 of the next Term upon reaching O4
// Force Commander (if Continue) — not modeled; the book doesn't elaborate
// what Command College actually does mechanically beyond being attended,
// and implementing it would need a real structural change (an inserted
// partial term), the same kind of complexity that's kept Education
// deferred all session.

// marineRankName formats a rank state as Book 1's own "M1 Private"/"O4
// Force Commander" notation.
func marineRankName(isOfficer bool, tier int) string {
	if isOfficer {
		return fmt.Sprintf("O%d %s", tier, marineOfficerRankNames[tier-1])
	}

	return fmt.Sprintf("M%d %s", tier, marineEnlistedRankNames[tier-1])
}

// marineRankAutomaticSkill is Book 1 p.86's own "Automatic Skills by
// Rank" — a one-time grant the moment a character first reaches a named
// rank tier (the same "upon reaching rank X" one-time-event framing
// Book 1's general Rank/Position/Promotion section uses for Merchant's
// own analogous automatic skill, p.65), not a repeated per-term grant
// for as long as the rank is held — callers must gate this on the term
// that actually caused the transition (Commissioned or Promoted true),
// not just "current tier matches".
func marineRankAutomaticSkill(isOfficer bool, tier int) (SkillLevel, bool) {
	switch {
	case isOfficer && tier == 1:
		return skillLevel1("Leader", Skill), true
	case isOfficer && tier == 4:
		return skillLevel1("Tactics", Skill), true
	case !isOfficer && tier == 3:
		return skillLevel1("Heavy Weapons", Skill), true
	case !isOfficer && tier == 4:
		return skillLevel1("Tactics", Skill), true
	case !isOfficer && tier == 5:
		return skillLevel1("Leader", Skill), true
	default:
		return SkillLevel{}, false
	}
}

// rollMarineCommission/rollMarineOfficerPromotion/rollMarineEnlistedPromotion
// are Book 1 p.86's own "Officer Commission C3" (no Mod — no asterisk on
// this line), "Officer Promotion Int*", and "Enlisted Promotion C1*"
// ("*+Medals and WB Mods" — this codebase applies Medal Mods only, not
// Wound Badges; see the plan-file history for why). Officer Promotion
// targets Int per Marine's own dedicated box, not Soc as the more
// general Master Checklist states — also explained there. Soldier's own
// equivalents (soldier_promotion.go) confirmed Soc IS the correct target
// for Soldier and Spacer — Marine is the real outlier, not a book-wide
// error.
func rollMarineCommission(r *dice.Roller, c3 int) bool {
	return rollAgainstTarget(r, c3, 0)
}

func rollMarineOfficerPromotion(r *dice.Roller, intChar, medalMod int) bool {
	return rollAgainstTarget(r, intChar, medalMod)
}

func rollMarineEnlistedPromotion(r *dice.Roller, c1, medalMod int) bool {
	return rollAgainstTarget(r, c1, medalMod)
}
