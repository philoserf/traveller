package character

import "github.com/philoserf/traveller/dice"

// MarineCareerName is Marine's own Career.Name value — exported and
// shared (ResolveMarineCareer's own Career{Name: MarineCareerName}
// literal below, a future render dispatch) as a single source of truth,
// matching CitizenCareerName's/NobleCareerName's own rationale.
const MarineCareerName = "Marine"

// marineRiskRewardPositions is Marine Risk & Reward's own Controlling
// Characteristic set (Book 1 p.86: "Risk & Reward C1 C4").
var marineRiskRewardPositions = []Position{C1, C4}

// marineBranchNames/marineBranchMods are Book 1 p.86's own "MARINE
// BRANCH" table (8 rows). Not a 1D roll despite the "1D" column header —
// Book 1's own Master Checklist (p.72) confirms "Select Branch" is a
// player choice ("Select," not "Roll," the same verb the checklist uses
// for Scout's own already-implemented "Select Courier Duty"), resolved
// here via the same uniform-random-choice convention rollScoutDuty
// already establishes for a genuine open choice with no book-given
// tiebreaker. Infantry appearing twice gives it a natural 2-in-8 weight,
// matching the table's own row layout as printed.
var marineBranchNames = [8]string{
	"Infantry", "Infantry", "Artillery", "Cavalry",
	"Protected", "Commando", "Technical", "Medical",
}
var marineBranchMods = [8]int{1, 1, 1, 1, 2, 2, 0, 0}

// marineOperationsNames/marineOperationsMods are Book 1 p.86's own
// "MARINE OPERATIONS" table's Operation/Mod columns (9 rows). "ANM
// School" is resolved as Education per the table's own aside ("Resolve
// ANM School as Education") — noted, not separately modeled; this
// codebase has no Education generation yet, the same deferral already
// made everywhere else Education comes up.
var marineOperationsNames = [9]string{
	"Combat", "Combat", "Peace Keeper", "Mission", "ANM School",
	"Combat", "Peace Keeper", "Mission", "Garrison",
}
var marineOperationsMods = [9]int{2, 2, 1, 2, 0, 3, 1, 2, 0}

// marineOperationsBranchDM is Operations' own separate "DM By Branch"
// column, keyed by branch NAME rather than row — both Infantry rows in
// marineBranchNames share this DM, distinct from Branch's own Mod
// (marineBranchMods) used for Risk & Reward.
var marineOperationsBranchDM = map[string]int{
	"Commando": 0, "Protected": 1, "Infantry": 2, "Cavalry": 3,
	"Medical": 4, "Artillery": 5, "Technical": 6,
}

// marineSkillTable is Book 1 p.86's own "C MARINE SKILLS" table,
// re-verified directly against the page image this slice (row 4's own
// Garrison entry is "Gambler," not "Minor *" as an earlier, memory-based
// recollection from earlier in this session had it — caught by
// re-reading rather than trusting a prior transcription).
var marineSkillTable = [7][6]string{
	{"Str", "Dex", "End", "Int", "Edu", "Soc"},
	{"One Trade", "Major", "Minor", "Gambler", "Athlete", "One Trade"},
	{"Fighter", "Fighter", "Soldier Skill", "Stealth", "Leader", "Tactics"},
	{"Vacc Suit", "Fighter", "Soldier Skill", "Hostile Environment", "Stealth", "Tactics"},
	{"Fighter", "Flyer", "Fighter", "Stealth", "Leader", "Heavy Weapons"},
	{"Soldier Skill", "Survival", "Language", "Gunner", "Leader", "Fighter"},
	{"One Art", "One Science", "Explosives", "Medic", "Seafarer", "One Trade"},
}

// marineSkillsPerTerm is Book 1 p.86's own "Skill Eligibility: Per Term 4".
const marineSkillsPerTerm = 4

// marineMedalCodes is Book 1 p.70's own "IMPERIAL MEDALS" table, keyed
// by the raw unmodified Reward roll (2-12; index 0 = roll 2). Roll 13
// (SEH With Diamonds, the +1 Officer bonus) is unreachable in this
// codebase yet — no character is ever assigned an Officer Rank
// (Promotion/Commission deferred), the same "+Officer" omission already
// documented for Mustering Out's own DM.
var marineMedalCodes = [11]string{
	"XS", "XS", "XS", "XS", "XS", "XS", "XS", // rolls 2-8
	"MCUF", "MCUF", // rolls 9-10
	"MCG", // roll 11
	"SEH", // roll 12
}

// marineMedalNames are the Medals table's own "Medal Description"
// column, used as ResolveMarineTerm's own RewardResult text — render's
// existing termOutcomeLine already prints any non-"None" RewardResult
// generically, so no render change is needed, the same zero-render-
// change precedent Phase T1 already established for this field.
var marineMedalNames = map[string]string{
	"XS":   "XS Exemplary Service",
	"MCUF": "MCUF Meritorious Conduct Under Fire",
	"MCG":  "MCG Medal for Conspicuous Gallantry",
	"SEH":  "SEH Starburst for Extreme Heroism",
}

// marineMedalFame is Book 1 p.91's own per-medal Fame contribution —
// distinct from marineMedalCodes' own Mod column (p.70), which feeds a
// still-deferred Promotion roll.
var marineMedalFame = map[string]int{"XS": 0, "MCUF": 1, "MCG": 2, "SEH": 3}

// marineMedalFromReward converts a raw, unmodified Reward roll (2-12)
// into its Medals table code.
func marineMedalFromReward(roll int) string {
	return marineMedalCodes[roll-2]
}

// BeginMarine reports Book 1 p.86's own "To Begin C1" — roll 2D <= Str.
// No Retry: unlike Scout's own explicit "Retry vs C5," the Master
// Checklist (p.72) shows no Retry line for Marine.
func BeginMarine(r *dice.Roller, str int) bool {
	return rollAgainstTarget(r, str, 0)
}

// rollMarineBranch resolves "Select Branch" (see marineBranchNames' own
// doc comment for why this is a choice, not a roll).
func rollMarineBranch(r *dice.Roller) (string, int) {
	i := r.Uniform(len(marineBranchNames)) - 1

	return marineBranchNames[i], marineBranchMods[i]
}

// marineOperationsEduDM is Operations' own "DM +2 if Edu 10+".
func marineOperationsEduDM(edu int) int {
	if edu >= 10 {
		return 2
	}

	return 0
}

// rollMarineOperations rolls 4 times (p.86: "Roll 4 times per Term for
// Operations; select the highest Mod from the four"), returning the
// winning operation's own name and Mod.
func rollMarineOperations(r *dice.Roller, branch string, edu int) (string, int) {
	dm := marineOperationsBranchDM[branch] + marineOperationsEduDM(edu)

	// bestIdx starts from the first of the 4 rolls, not a hardcoded
	// index — seeding it with an unrolled row would bias the result
	// toward that row whenever none of the real rolls beat its Mod.
	bestIdx := musterOutRow(r.D6()+dm, len(marineOperationsNames))

	for range 3 {
		row := musterOutRow(r.D6()+dm, len(marineOperationsNames))
		if marineOperationsMods[row] > marineOperationsMods[bestIdx] {
			bestIdx = row
		}
	}

	return marineOperationsNames[bestIdx], marineOperationsMods[bestIdx]
}

// ResolveMarineTerm resolves one 4-year Marine term. branch/branchMod
// are fixed for the whole career (selected once, see
// ResolveMarineCareer); Operations rolls fresh every term. Risk &
// Reward's combined Mod (branchMod + this term's own Operations Mod,
// Book 1 p.65's universal "-Mod under Risk, +Mod under Reward"
// convention) feeds resolveRisk/resolveReward with the same
// shared-mod-value convention ResolveScoutTerm already establishes.
//
// Medals p.86's own box describes two separate, stackable grants per
// term, not one: Risk success alone grants a flat XS Exemplary Service
// Badge ("Success: Receive XS Exemplary Service Badge. Character is
// unharmed."), independent of the Reward roll; Reward success separately
// consults the Medals table (p.70, keyed by the raw unmodified Reward
// roll) for its own, possibly higher-tier medal. RewardResult carries
// the winning medal's own description text — render's own
// termOutcomeLine already handles any non-"None" RewardResult
// generically, so Marine needs no new render branch, unlike Citizen/
// Noble's own distinct outcome shapes.
func ResolveMarineTerm(r *dice.Roller, upp UPP, ccPos Position, branch string, branchMod int) (Term, UPP) {
	opName, opMod := rollMarineOperations(r, branch, int(upp.Characteristics[C5]))

	cc := upp.Characteristics[ccPos]
	mod := -(branchMod + opMod)

	term := Term{
		Length:                    4,
		ControllingCharacteristic: ccPos,
		Branch:                    branch,
		Assignment:                opName,
		RewardResult:              "None",
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

	if ok, rewardRoll := resolveReward(r, reducedCC, mod); ok {
		medal := marineMedalFromReward(rewardRoll)
		term.RewardResult = marineMedalNames[medal]
		term.Medals = append(term.Medals, medal)
	}

	term.SkillsAwarded = rollSkillsFromTable(r, marineSkillTable, marineSkillsPerTerm)

	return term, upp
}
