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
// Reward failure/success is recorded as "None"/"XS Exemplary Service"
// (p.86's own Reward-success text) — render's own termOutcomeLine
// already handles any non-"None" RewardResult generically (it just
// prints whatever string is there), so Marine needs no new render
// branch once wired in, unlike Citizen/Noble's own distinct outcome
// shapes.
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

	// "XS Exemplary Service," Book 1 p.86's own Reward-success text
	// ("Success: XS Exemplary Service and consult Medals table") — not
	// Scout's own "Discovery," a distinct outcome name. Medals table
	// consultation itself is deferred (not yet transcribed).
	if resolveReward(r, reducedCC, mod) {
		term.RewardResult = "XS Exemplary Service"
	}

	term.SkillsAwarded = rollSkillsFromTable(r, marineSkillTable, marineSkillsPerTerm)

	return term, upp
}
