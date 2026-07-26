package character

import (
	"fmt"

	"github.com/philoserf/traveller/dice"
)

// MerchantCareerName is Merchant's own Career.Name value — exported and
// shared, matching every other career's own CareerName rationale.
const MerchantCareerName = "Merchant"

var merchantRiskRewardPositions = []Position{C1, C2, C3, C4}

// merchantEnlistedRankNames/merchantOfficerRankNames are Book 1 p.80's
// own "Table Of Merchant Ranks" — RX Temp is tier 0 (Book 1's own "Temp
// is casual or untrained," a distinct floor below R0 Spacehand), R0-R2
// tiers 1-3.
var merchantEnlistedRankNames = [4]string{"Temp", "Spacehand", "Steward Apprentice", "Drive Helper"}

var merchantOfficerRankNames = [6]string{
	"Fourth Officer", "Third Officer", "Second Officer",
	"First Officer", "Captain", "Senior Captain",
}

// merchantEnlistedAutoSkills/merchantOfficerAutoSkills are the Rank
// table's own "Auto Skill" column — a one-time grant the moment a
// character first reaches a named tier (Marine/Soldier/Spacer's own
// established "upon reaching rank X" framing), keyed by tier. RX/R0 and
// M5/M6 grant nothing (absent from the map).
var merchantEnlistedAutoSkills = map[int]string{2: "Steward", 3: "Engineer"} // R1, R2

var merchantOfficerAutoSkills = map[int]string{1: "Steward", 2: "Engineer", 3: "Astrogator", 4: "Pilot"} // M1-M4

// merchantSkillTable is Book 1 p.80's own "MERCHANT SKILLS" table. Its
// own "JOT" cell (column 6, row 5) is normalized to "Jack of all
// Trades", the same transcription call already made for Rogue's own
// table.
var merchantSkillTable = [7][6]string{
	{"Str", "C2", "C3", "Int", "C5", "C6"},
	{"Major", "Major", "Minor", "Minor", "One Trade", "One Trade"},
	{"Astrogator", "Pilot", "Medic", "Sensors", "Steward", "Gunner"},
	{"Broker", "Trader", "Diplomat", "Admin", "Steward", "Trader"},
	{"Broker", "Trader", "Diplomat", "Advocate", "Steward", "Comms"},
	{"Broker", "Admin", "Language", "Starship Skill", "Jack of all Trades", "Vacc Suit"},
	{"One Art", "One Science", "Computer", "Comms", "Medic", "One Trade"},
}

const merchantSkillsPerTerm = 4

// BeginMerchant resolves Book 1 p.80's own three-tier cascading Begin —
// the only career this session with no failure state at all. Returns
// (isOfficer, tier): a successful Int roll enters as M1 Fourth Officer
// (tier 1, Officer); failing that, a successful Dex roll enters as R0
// Spacehand (tier 1, Enlisted); failing both, automatic entry as RX Temp
// (tier 0, Enlisted).
func BeginMerchant(r *dice.Roller, upp UPP) (bool, int) {
	if rollAgainstTarget(r, int(upp.Characteristics[C4]), 0) {
		return true, 1
	}

	if rollAgainstTarget(r, int(upp.Characteristics[C2]), 0) {
		return false, 1
	}

	return false, 0
}

// merchantPromotionMod is Book 1 p.80's own single "*Mod +3 if Int 8+"
// footnote, shared by both Officer Promotion and Rating Promotion.
func merchantPromotionMod(intChar int) int {
	if intChar >= 8 {
		return 3
	}

	return 0
}

func merchantRankName(isOfficer bool, tier int) string {
	if isOfficer {
		return fmt.Sprintf("M%d %s", tier, merchantOfficerRankNames[tier-1])
	}

	names := [4]string{"RX", "R0", "R1", "R2"}

	return fmt.Sprintf("%s %s", names[tier], merchantEnlistedRankNames[tier])
}

func merchantRankAutoSkill(isOfficer bool, tier int) (SkillLevel, bool) {
	return rankAutoSkillFromTables(merchantEnlistedAutoSkills, merchantOfficerAutoSkills, isOfficer, tier)
}

// merchantRewardCount sums how many prior terms' own Reward succeeded —
// the "receipt number" Escalating Ship Shares needs. Mirrors
// termOutcomeLine's own generic "RewardResult != ” && != 'None'" check
// (render/character.go) so the two stay consistent by construction.
func merchantRewardCount(terms []Term) int {
	n := 0

	for _, t := range terms {
		if t.RewardResult != "" && t.RewardResult != "None" {
			n++
		}
	}

	return n
}

// ResolveMerchantTerm resolves one 4-year Merchant term (Book 1 p.80).
// isOfficer/tier are the rank state entering this term. Risk reuses the
// exact universal p.65 mechanic (resolveRisk/riskOutcome); Reward is
// rolled whenever Risk isn't Dead, and success grants Escalating Ship
// Shares equal to the running count of successful Rewards so far,
// recorded as a formatted string in the generic RewardResult field
// ("2 Ship Shares") rather than a new typed field.
func ResolveMerchantTerm(
	r *dice.Roller,
	upp UPP,
	ccPos Position,
	isOfficer bool,
	tier int,
	priorTerms []Term,
) (Term, UPP, bool, int) {
	term := Term{Length: 4, ControllingCharacteristic: ccPos}

	cc := upp.Characteristics[ccPos]
	riskResult, reducedCC := resolveRisk(r, cc, 0)
	term.RiskResult = riskResult
	upp.Characteristics[ccPos] = reducedCC

	if riskResult != Dead {
		if rewardOK, _ := resolveReward(r, cc, 0); rewardOK {
			shares := merchantRewardCount(priorTerms) + 1
			term.RewardResult = fmt.Sprintf("%d Ship Share", shares)

			if shares != 1 {
				term.RewardResult += "s"
			}
		} else {
			term.RewardResult = "None"
		}
	}

	intChar := int(upp.Characteristics[C4])
	mod := merchantPromotionMod(intChar)

	switch {
	case isOfficer:
		target := 2 * (len(priorTerms) + 1)
		if rollAgainstTarget(r, target, mod) {
			term.Promoted = true
			tier = min(tier+1, len(merchantOfficerRankNames))
		}
	case rollAgainstTarget(r, intChar, 0):
		term.Commissioned = true
		isOfficer = true
		tier = 1
	default:
		if rollAgainstTarget(r, int(upp.Characteristics[C2]), mod) {
			term.Promoted = true
			tier = min(tier+1, len(merchantEnlistedRankNames)-1)
		}
	}

	term.Rank = merchantRankName(isOfficer, tier)

	skillCount := merchantSkillsPerTerm
	if term.Commissioned || term.Promoted {
		skillCount++

		if skill, ok := merchantRankAutoSkill(isOfficer, tier); ok {
			term.SkillsAwarded = append(term.SkillsAwarded, skill)
		}
	}

	term.SkillsAwarded = append(term.SkillsAwarded, rollSkillsFromTable(r, merchantSkillTable, skillCount)...)

	return term, upp, isOfficer, tier
}
