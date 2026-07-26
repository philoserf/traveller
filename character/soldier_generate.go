package character

import "github.com/philoserf/traveller/dice"

// SoldierCareerName is Soldier's own Career.Name value — exported and
// shared, matching MarineCareerName's/CitizenCareerName's own rationale.
const SoldierCareerName = "Soldier"

// soldierRiskRewardPositions is Soldier's own Risk & Reward Controlling
// Characteristic set (Book 1 p.82: "Risk & Reward C1 C3 C4") — three
// characteristics (Str, End, Int), confirmed directly against Soldier's
// own dedicated box. The Master Checklist's own summary (p.73, "Str C2
// C3 Int") lists a fourth (C2); Soldier's own box is the more specific,
// authoritative source and wins, the same precedent already applied
// twice for Marine (Branch selection, Officer Promotion target).
var soldierRiskRewardPositions = []Position{C1, C3, C4}

// soldierBranchNames/soldierBranchMods are Book 1 p.82's own "ARMY
// BRANCH" table (8 rows) — genuinely distinct from Marine's own table:
// Protected appears twice (rows 5-6) instead of Marine's Protected/
// Commando pair, confirmed directly against the page image, not assumed
// identical to Marine's.
var soldierBranchNames = [8]string{
	"Infantry", "Infantry", "Artillery", "Cavalry",
	"Protected", "Protected", "Technical", "Medical",
}
var soldierBranchMods = [8]int{1, 1, 1, 1, 2, 2, 0, 0}

// soldierOperationsNames/soldierOperationsMods are Book 1 p.82's own
// "ARMY OPERATIONS" table's Operation/Mod columns (9 rows) — row 9 is
// "Base", not Marine's own "Garrison".
var soldierOperationsNames = [9]string{
	"Combat", "Combat", "Peace Keeper", "Mission", "ANM School",
	"Combat", "Peace Keeper", "Mission", "Base",
}
var soldierOperationsMods = [9]int{2, 2, 1, 2, 0, 3, 1, 2, 0}

// soldierOperationsBranchDM is Operations' own separate "DM By Branch"
// column (6 unique entries for Soldier's own 6 unique branch names).
// A first direct read of this cell in the page image mis-paired two
// entries (its own small font made "Infantry +1" easy to miss and
// "Artillery" looked like it appeared twice) — resolved by cross-
// checking the `.txt` extraction's own independent OCR pass, which
// aligned unambiguously with all six branch names appearing exactly
// once, DMs 0/1/3/4/5/6.
var soldierOperationsBranchDM = map[string]int{
	"Protected": 0, "Infantry": 1, "Cavalry": 3,
	"Medical": 4, "Artillery": 5, "Technical": 6,
}

// soldierSkillTable is Book 1 p.82's own "C SOLDIER SKILLS" table,
// transcribed directly from the page image.
var soldierSkillTable = [7][6]string{
	{"Str", "Dex", "End", "Int", "Edu", "Soc"},
	{"Major", "Major", "Minor", "Minor", "One Trade", "One Trade"},
	{"Fighter", "Vacc Suit", "Fighter", "Stealth", "Leader", "Tactics"},
	{"Admin", "Fighter", "Hostile Environment", "Animals", "Liaison", "Navigation"},
	{"Fighter", "Vacc Suit", "Driver", "Stealth", "Heavy Weapons", "Sensors"},
	{"Soldier Skill", "Liaison", "Language", "Soldier Skill", "Computer", "Tactics"},
	{"One Art", "One Science", "Explosives", "Medic", "Seafarer", "One Trade"},
}

// soldierSkillsPerTerm is Book 1 p.82's own "Skill Eligibility: Per Term 4".
const soldierSkillsPerTerm = 4

// BeginSoldier reports Book 1 p.82's own "To Begin Str" — roll 2D <=
// Str. No Retry, matching Marine's own no-Retry treatment (the Master
// Checklist shows no Retry line for either).
func BeginSoldier(r *dice.Roller, str int) bool {
	return rollAgainstTarget(r, str, 0)
}

// rollSoldierBranch resolves "Select Branch" — a genuine open player
// choice (Book 1's own Master Checklist "Select," not "Roll," the same
// verb already resolved this way for Scout's own Duty pick and Marine's
// own Branch pick).
func rollSoldierBranch(r *dice.Roller) (string, int) {
	i := r.Uniform(len(soldierBranchNames)) - 1

	return soldierBranchNames[i], soldierBranchMods[i]
}

// rollSoldierOperations rolls 4 times (p.82: "Rolls 4 times per Term for
// Operations; select the highest Mod of the four"), delegating to the
// shared rollOperations/operationsEduDM (career_generate.go).
func rollSoldierOperations(r *dice.Roller, branch string, edu int) ([]string, string, int) {
	dm := soldierOperationsBranchDM[branch] + operationsEduDM(edu)

	return rollOperations(r, dm, soldierOperationsNames[:], soldierOperationsMods[:])
}

// ResolveSoldierTerm resolves one 4-year Soldier term — mirrors
// ResolveMarineTerm's own structure exactly (marine_generate.go), with
// Soldier's own distinct target characteristics: Officer Promotion
// against Soc (C6, not Marine's Int) and Enlisted Promotion against C3
// (not Marine's C1), both confirmed directly against Soldier's own
// dedicated box ("Officer Promotion Soc*", "Enlisted Promotion C3*").
// Officer Commission targets C3, same as Marine.
func ResolveSoldierTerm(
	r *dice.Roller, upp UPP, ccPos Position, branch string, branchMod int, priorTerms []Term,
) (Term, UPP) {
	operations, opName, opMod := rollSoldierOperations(r, branch, int(upp.Characteristics[C5]))

	cc := upp.Characteristics[ccPos]
	mod := -(branchMod + opMod)

	isOfficer, tier := rankState(priorTerms, len(soldierEnlistedRankNames), len(soldierOfficerRankNames))

	term := Term{
		Length:                    4,
		ControllingCharacteristic: ccPos,
		Branch:                    branch,
		Assignment:                opName,
		Operations:                operations,
		RewardResult:              "None",
		Rank:                      soldierRankName(isOfficer, tier),
	}

	risk, reducedCC := resolveRisk(r, cc, mod)
	term.RiskResult = risk
	upp.Characteristics[ccPos] = reducedCC

	if risk == Dead {
		return term, upp
	}

	if risk == Unharmed {
		term.Medals = append(term.Medals, "XS")
	}

	if ok, rewardRoll := resolveReward(r, cc, mod); ok {
		if isOfficer {
			rewardRoll++ // Book 1 p.70's own "If Officer, increase +1"
		}

		medal := medalFromReward(rewardRoll)
		term.RewardResult = medalNames[medal]
		term.Medals = append(term.Medals, medal)
	}

	medalMod := medalModTotal(priorTerms) + medalModSum(term.Medals)

	switch {
	case isOfficer:
		if tier < len(soldierOfficerRankNames) &&
			rollSoldierOfficerPromotion(r, int(upp.Characteristics[C6]), medalMod) {
			term.Promoted = true
			tier++
		}
	case rollSoldierCommission(r, int(upp.Characteristics[C3])):
		term.Commissioned = true
		isOfficer = true
		tier = 1
	case tier < len(soldierEnlistedRankNames) && rollSoldierEnlistedPromotion(r, int(upp.Characteristics[C3]), medalMod):
		term.Promoted = true
		tier++
	}

	term.Rank = soldierRankName(isOfficer, tier)

	// Term skills only. The extra skill a Commission or Promotion
	// grants is counted separately below, because p.65 exempts it from
	// the Operations-column restriction.
	skillCount := soldierSkillsPerTerm
	exemptSkills := 0

	if term.Commissioned || term.Promoted {
		exemptSkills++

		if skill, ok := soldierRankAutomaticSkill(isOfficer, tier); ok {
			term.SkillsAwarded = append(term.SkillsAwarded, skill)
		}
	}

	// Book 1 p.65: Term skills "may be taken on a column of the Skills
	// table corresponding to an Operations result received in the Term",
	// plus Personal, which "may always be rolled".
	columns := eligibleSkillColumns(operations, soldierOperationsColumns)
	term.SkillsAwarded = append(term.SkillsAwarded,
		rollSkillsFromColumns(r, soldierSkillTable, columns, skillCount)...)

	// The Commission/Promotion skill is one of the eligibilities p.65
	// exempts, so it draws from the whole table — which is how p.65's own
	// worked example reaches a column its Operations never granted.
	term.SkillsAwarded = append(term.SkillsAwarded,
		rollSkillsFromTable(r, soldierSkillTable, exemptSkills)...)

	return term, upp
}
