package character

import (
	"fmt"
	"slices"

	"github.com/philoserf/traveller/dice"
)

// CraftsmanCareerName is Craftsman's own Career.Name value — exported
// and shared, matching every other career's own CareerName rationale.
const CraftsmanCareerName = "Craftsman"

// craftsmanRiskRewardPositions is the Masterpiece Controlling
// Characteristic set (Book 1 p.75: "Masterpiece C1 C2 C3 C4") — no
// "*Special Case" marker the way Rogue's own fixed-for-career CC has,
// so this rotates per term via the shared nextCC, the same as most
// other careers.
var craftsmanRiskRewardPositions = []Position{C1, C2, C3, C4}

// craftsmanSkillTable is Book 1 p.75's own "CRAFTSMAN SKILLS" table.
// Column 5's own two "New Trade" cells (rows 3-4) are resolved outside
// resolveSkillCell's shared switch — see resolveCraftsmanSkillCell,
// below — since they need the character's own held-skills context that
// function deliberately doesn't have. "C6+1"'s own "lost if C6=Caste"
// footnote never applies here — GenerateUPP only produces Human
// characters (C6 always Social Standing), the same already-documented
// reasoning as every other career's own C6 column.
var craftsmanSkillTable = [7][6]string{
	{"Str", "C2", "C3", "Int", "C5", "C6"},
	{"Major", "Major", "Minor", "Minor", "One Trade", "One Trade"},
	{"Seafarer", "Navigation", "Hostile Environment", "Flyer", "Driver", "Vacc Suit"},
	{"Animals", "Comms", "Computer", "Designer", "Designer", "Designer"},
	{"Liaison", "Comms", "Bureaucrat", "Diplomat", "Leader", "Trader"},
	{"Naval Architect", "One Art", "New Trade", "New Trade", "One Trade", "One Trade"},
	{"Animals", "One Art", "One Science", "Athlete", "Medic", "One Trade"},
}

const (
	craftsmanSkillsPerTerm     = 4
	craftsmanSkillBonusPerHit  = 3 // "Per Success 3+Craftsman-1"
	craftsmanSkillBonusPerMiss = 1 // "Per Failure 1+Craftsman-1"

	craftsmanMinMasterPoints     = 40 // below this, cannot attempt a Masterpiece at all (treated as Failure)
	craftsmanPerfectMasterPoints = 55 // "A Perfect Masterpiece has 55 or more Master Points"
	craftsmanMasterPointsDice    = 9  // "Roll 9D for Master Points" / "Roll: 9D < Master Points"

	craftsmanBaseMasterpieceValue = 150000 // "sold at Cr150,000..."
	craftsmanValuePerPoint        = 10000  // "...plus Cr10,000 per Master Point over 40"
)

// craftsmanQualifyingSkills is Book 1's own "up to FIVE Skills at level
// 6+ (or Knowledges at level-6) (but not languages)" pool for the
// Master Points formula and the Begin prerequisite alike — excludes
// "Language" (a specific-tongue skill name used elsewhere in this
// codebase, the literal thing "not languages" means) and "Craftsman"
// itself (already counted separately, both here and in Master Points).
func craftsmanQualifyingSkills(skills []SkillLevel) []SkillLevel {
	var qualifying []SkillLevel

	for _, s := range skills {
		if s.Level < 6 || s.Name == "Language" || s.Name == "Craftsman" {
			continue
		}

		if s.Kind != Skill && s.Kind != Knowledge {
			continue
		}

		qualifying = append(qualifying, s)
	}

	return qualifying
}

// craftsmanSkillLevel returns the character's own current "Craftsman"
// skill level, or 0 if never held.
func craftsmanSkillLevel(skills []SkillLevel) int {
	for _, s := range skills {
		if s.Name == "Craftsman" && s.Kind == Skill {
			return s.Level
		}
	}

	return 0
}

// BeginCraftsman is a pure prerequisite check, not a dice roll — Book 1
// p.75's own "To Begin: Automatic* (*if TWO skill-6 and Craftsman-1)."
// "Craftsman" is already a real, grantable skill elsewhere in this
// codebase (homeworld_generate.go's own Pre-Rich/Industrial Trade
// grants, citizen_generate.go's own Table E) — this career depends on a
// character having already accumulated it, and two other skills at
// level 6+, through prior careers or a long enough homeworld/Citizen
// life, not on anything new this function grants.
func BeginCraftsman(skills []SkillLevel) bool {
	return craftsmanSkillLevel(skills) >= 1 && len(craftsmanQualifyingSkills(skills)) >= 2
}

// craftsmanMasterPoints implements Book 1's own Master Points formula:
// Controlling Characteristic + Craftsman skill level + the sum of the
// five HIGHEST qualifying skills' own levels (more than five can
// qualify; "up to five" caps it — with no player choice modeled here,
// the best five are used, the same "resolve an open choice toward the
// better outcome" precedent BeginScout's own doc comment already
// establishes).
func craftsmanMasterPoints(cc int, skills []SkillLevel) int {
	qualifying := craftsmanQualifyingSkills(skills)
	slices.SortFunc(qualifying, func(a, b SkillLevel) int { return b.Level - a.Level })

	total := cc + craftsmanSkillLevel(skills)

	for i, s := range qualifying {
		if i >= 5 {
			break
		}

		total += s.Level
	}

	return total
}

// craftsmanMasterpieceValue reverse-derives Book 1's own printed
// Masterpiece Value table: 40 points -> Cr150,000, +Cr10,000 per point
// over 40 — confirmed against the table's own jump at 55 points
// (Cr290,000 at 54, Cr600,000 at 55, not the Cr300,000 the plain
// formula would give): "A Perfect Masterpiece... sells for Double,"
// applied as a final doubling once points reach 55, not a separate
// curve.
func craftsmanMasterpieceValue(masterPoints int) int {
	value := craftsmanBaseMasterpieceValue + craftsmanValuePerPoint*(masterPoints-craftsmanMinMasterPoints)

	if masterPoints >= craftsmanPerfectMasterPoints {
		value *= 2
	}

	return value
}

// continueCraftsman implements Book 1's own "Continue Craftsman x2" —
// no natural-roll exception is documented on this career's own box
// (unlike Marine's natural-2 or Rogue's natural-12), matching Scholar's/
// Entertainer's/Merchant's own plain-roll convention.
func continueCraftsman(r *dice.Roller, craftsmanLevel int) bool {
	return rollAgainstTarget(r, 2*craftsmanLevel, 0)
}

// resolveCraftsmanSkillCell handles the skill table's own two "New
// Trade" cells (Book 1's own footnote: "Any Trade not already held; if
// all held, this benefit is lost") before delegating everything else to
// the shared resolveSkillCell — "New Trade" needs the character's own
// held-skills context, which resolveSkillCell deliberately doesn't
// have (every other case it handles is context-free).
func resolveCraftsmanSkillCell(r *dice.Roller, column, row int, heldSkills []SkillLevel) (SkillLevel, bool) {
	if craftsmanSkillTable[column][row] != "New Trade" {
		return resolveSkillCell(r, craftsmanSkillTable, column, row)
	}

	var available []string

	for _, choice := range theTradeChoices {
		held := false

		for _, s := range heldSkills {
			if s.Name == choice && s.Kind == Skill {
				held = true

				break
			}
		}

		if !held {
			available = append(available, choice)
		}
	}

	if len(available) == 0 {
		return SkillLevel{}, false
	}

	return skillLevel1(rollChoice(r, available), Skill), true
}

// rollCraftsmanSkills grants count skills, threading heldSkills forward
// as each one resolves — "New Trade" (above) needs to see skills
// granted earlier in the SAME roll batch, not just what the character
// held before this term began. Each new grant is folded into
// heldSkills via aggregateSkills (character/skill.go), not a raw
// append: Craftsman is the first career whose own mid-career logic
// (craftsmanQualifyingSkills' own level>=6 check, both here and in the
// next term's Master Points) reads heldSkills back before the whole
// chain's own final assembly ever aggregates it — a raw append would
// leave a skill split across two sub-6 grants silently excluded for
// the rest of the career, understating Master Points every term until
// the segment ends.
func rollCraftsmanSkills(r *dice.Roller, count int, heldSkills []SkillLevel) ([]SkillLevel, []SkillLevel) {
	var granted []SkillLevel

	for range count {
		skill, ok := resolveCraftsmanSkillCell(r, r.Uniform(7)-1, r.Uniform(6)-1, heldSkills)
		if !ok {
			continue
		}

		granted = append(granted, skill)
		heldSkills = aggregateSkills(append(heldSkills, skill))
	}

	return granted, heldSkills
}

// ResolveCraftsmanTerm resolves one 4-year Craftsman term (Book 1
// p.75). Craftsman "does not roll Risk and Reward" at all — RiskResult
// stays at its zero value, Unharmed, the same "no death mechanic,
// RiskResult never meaningfully set" precedent Rogue's own terms
// already establish, so resolveCareerLoop's Risk-based stop checks
// never fire for this career; continueCraftsman (the career's own
// "Continue Craftsman x2") is the sole real stop condition.
//
// If Master Points falls below 40, no 9D roll happens at all — Book 1's
// own "If the Craftsman cannot show at least 40 Master Points, he
// cannot attempt a Masterpiece (treat as Failure)" — but skills are
// still granted per the Failure rate, matching every other career's own
// "a failed attempt still teaches something" convention.
func ResolveCraftsmanTerm(r *dice.Roller, upp UPP, ccPos Position, heldSkills []SkillLevel) (Term, []SkillLevel) {
	term := Term{Length: 4, ControllingCharacteristic: ccPos}

	cc := int(upp.Characteristics[ccPos])
	masterPoints := craftsmanMasterPoints(cc, heldSkills)

	succeeded := false
	if masterPoints >= craftsmanMinMasterPoints {
		succeeded = r.SumD6(craftsmanMasterPointsDice) < masterPoints
	}

	craftsmanLevel := craftsmanSkillLevel(heldSkills)
	skillCount := craftsmanSkillsPerTerm

	if succeeded {
		perfect := masterPoints >= craftsmanPerfectMasterPoints
		term.Perfect = perfect

		value := craftsmanMasterpieceValue(masterPoints)
		if perfect {
			term.RewardResult = fmt.Sprintf("Perfect Masterpiece (Cr%d)", value)
		} else {
			term.RewardResult = fmt.Sprintf("Masterpiece (Cr%d)", value)
		}

		skillCount += craftsmanSkillBonusPerHit + craftsmanLevel
	} else {
		term.RewardResult = "None"
		skillCount += craftsmanSkillBonusPerMiss + craftsmanLevel
	}

	granted, newHeld := rollCraftsmanSkills(r, skillCount, heldSkills)
	term.SkillsAwarded = granted

	return term, newHeld
}

// craftsmanCareerFame is Book 1 p.91's own two-row, additive Fame
// entry: "Craftsman = Masterpieces x3" and "Craftsman = Perfect
// Masterpieces x5" — a Perfect Masterpiece is a subset of Masterpieces,
// not a separate exclusive category (unlike Rogue's own Successful-xor-
// Failed Scheme shape), so it contributes to BOTH rows: 3+5=8 total,
// versus 3 for a non-perfect one.
func craftsmanCareerFame(terms []Term) int {
	fame := 0

	for _, t := range terms {
		if t.RewardResult == "" || t.RewardResult == "None" {
			continue
		}

		fame += 3

		if t.Perfect {
			fame += 5
		}
	}

	return fame
}
