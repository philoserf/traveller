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

// marineMedalMod is Book 1 p.70's own Medals table Mod column — distinct
// from marineMedalFame (marine_generate.go), which is Fame, not a
// Promotion Mod.
var marineMedalMod = map[string]int{"XS": 1, "MCUF": 2, "MCG": 3, "SEH": 4, "SEHD": 5}

// marineMedalModSum sums medals' own Mod values.
func marineMedalModSum(medals []string) int {
	total := 0
	for _, m := range medals {
		total += marineMedalMod[m]
	}

	return total
}

// marineMedalModTotal sums every medal Mod earned across terms — Book 1
// p.66's own fully-worked Eneri Dinsha example computes Promotion's own
// target as "Soc plus Medal Mods," cumulative across the whole career to
// that point, not just the current term's own medals.
func marineMedalModTotal(terms []Term) int {
	total := 0
	for _, t := range terms {
		total += marineMedalModSum(t.Medals)
	}

	return total
}

// marineRankState derives the current rank track and tier from prior
// terms — derived, not separately tracked, matching this codebase's
// existing "derive from Terms" discipline (nobleIntrigueCounts,
// citizenLifeSuccessCount). Every Marine begins Enlisted at tier 1 (M1
// Private, Book 1 p.65's own "begin with enlisted rank... Marine1") —
// the zero-terms case. A Commissioned term resets to Officer tier 1
// (O1); a Promoted term advances the current track's own tier by 1,
// capped at the track's own maximum (6 Enlisted, 7 Officer).
func marineRankState(terms []Term) (bool, int) {
	isOfficer := false
	tier := 1

	for _, t := range terms {
		switch {
		case t.Commissioned:
			isOfficer = true
			tier = 1
		case t.Promoted:
			if isOfficer {
				tier = min(tier+1, len(marineOfficerRankNames))
			} else {
				tier = min(tier+1, len(marineEnlistedRankNames))
			}
		}
	}

	return isOfficer, tier
}

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

// marineBranchAutomaticSkill is Book 1 p.86's own branch-tied automatic
// skill ("if Medical Branch=Medic-1. If Technical Branch=any Trade"),
// granted once at career start (Branch is selected once per career, not
// per term) — reuses theTradeChoices/rollChoice (homeworld_generate.go)
// for Technical's own open "any Trade" pick, the same resolution already
// used for "One Trade" table entries (career_generate.go).
func marineBranchAutomaticSkill(r *dice.Roller, branch string) (SkillLevel, bool) {
	switch branch {
	case "Medical":
		return skillLevel1("Medic", Skill), true
	case "Technical":
		return skillLevel1(rollChoice(r, theTradeChoices), Skill), true
	default:
		return SkillLevel{}, false
	}
}

// rollMarineCommission/rollMarineOfficerPromotion/rollMarineEnlistedPromotion
// are Book 1 p.86's own "Officer Commission C3" (no Mod — no asterisk on
// this line), "Officer Promotion Int*", and "Enlisted Promotion C1*"
// ("*+Medals and WB Mods" — this codebase applies Medal Mods only, not
// Wound Badges; see this slice's own plan-file Context for why). Officer
// Promotion targets Int per Marine's own dedicated box, not Soc as the
// more general Master Checklist states — also explained in the Context.
func rollMarineCommission(r *dice.Roller, c3 int) bool {
	return rollAgainstTarget(r, c3, 0)
}

func rollMarineOfficerPromotion(r *dice.Roller, intChar, medalMod int) bool {
	return rollAgainstTarget(r, intChar, medalMod)
}

func rollMarineEnlistedPromotion(r *dice.Roller, c1, medalMod int) bool {
	return rollAgainstTarget(r, c1, medalMod)
}
