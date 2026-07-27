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

// rollMarineOperations rolls 4 times (p.86: "Roll 4 times per Term for
// Operations; select the highest Mod from the four"), returning the
// winning operation's own name and Mod. Delegates to the shared
// rollOperations/operationsEduDM (career_generate.go) — Marine's own
// former per-loop body, generalized once Soldier became a second real
// caller of the identical shape.
func rollMarineOperations(r *dice.Roller, branch string, edu int) ([]string, string, int) {
	dm := marineOperationsBranchDM[branch] + operationsEduDM(edu)

	return rollOperations(r, dm, marineOperationsNames[:], marineOperationsMods[:])
}

// ResolveMarineTerm resolves one 4-year Marine term. branch/branchMod
// are fixed for the whole career (selected once, see
// ResolveMarineCareer); Operations rolls fresh every term. Risk &
// Reward's combined Mod (branchMod + this term's own Operations Mod,
// Book 1 p.65's universal "-Mod under Risk, +Mod under Reward"
// convention) feeds resolveRisk/resolveReward with the same
// shared-mod-value convention ResolveScoutTerm already establishes.
// priorTerms is every term already resolved this career (oldest first,
// not including this one) — needed to derive the current rank state
// (marineRankState) and the cumulative Medal Mod (marineMedalModTotal)
// Promotion rolls consume; see character/marine_promotion.go.
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
//
// Rank state (isOfficer/tier) is derived from priorTerms up front, so
// term.Rank reflects the rank actually held entering this term even on
// an early return (RiskResult == Dead) — a code-review-caught bug in an
// earlier draft computed rank state only after the Dead check, leaving
// a fallen Officer's own final term.Rank empty.
//
// Commission/Promotion: a still-Enlisted character rolls Commission
// first; if it succeeds, Enlisted Promotion is not separately rolled
// that same term (the character is no longer Enlisted by the time that
// roll would apply) — a documented judgment call resolving a genuine
// ambiguity in the Master Checklist's own flat per-term step listing.
// An Officer only ever rolls Officer Promotion. Neither roll is even
// attempted once a track is already at its own maximum tier (M6/O7) —
// another code-review-caught bug let Promotion keep "succeeding" past
// the cap, granting an unearned +1 skill every remaining term. Both
// Promotion rolls use medalMod (cumulative Medal Mods, including this
// term's own just-earned medals — Book 1 p.66's own worked example adds
// a medal earned in the same term to that term's own Promotion roll)
// but not Wound Badges, despite the box's own "+Medals and WB Mods"
// footnote — see this slice's own plan-file Context for the full
// reasoning. Commission/Promotion success each grant +1 skill this term
// (p.86's own "Skill Eligibility: Commission 1 / Promotion 1") plus any
// Automatic Skill by Rank (marineRankAutomaticSkill) the newly-reached
// rank carries.
func ResolveMarineTerm(
	r *dice.Roller, upp UPP, ccPos Position, branch string, branchMod int, priorTerms []Term,
) (Term, UPP) {
	operations, opName, opMod := rollMarineOperations(r, branch, int(upp.Characteristics[C5]))

	cc := upp.Characteristics[ccPos]
	mod := -(branchMod + opMod)

	isOfficer, tier := rankState(priorTerms, len(marineEnlistedRankNames), len(marineOfficerRankNames))

	term := Term{
		Length:                    4,
		ControllingCharacteristic: ccPos,
		Branch:                    branch,
		Assignment:                opName,
		Operations:                operations,
		RewardResult:              "None",
		Rank:                      marineRankName(isOfficer, tier),
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
		if tier < len(marineOfficerRankNames) && rollMarineOfficerPromotion(r, int(upp.Characteristics[C4]), medalMod) {
			term.Promoted = true
			tier++
		}
	case rollMarineCommission(r, int(upp.Characteristics[C3])):
		term.Commissioned = true
		isOfficer = true
		tier = 1
	case tier < len(marineEnlistedRankNames) && rollMarineEnlistedPromotion(r, int(upp.Characteristics[C1]), medalMod):
		term.Promoted = true
		tier++
	}

	term.Rank = marineRankName(isOfficer, tier)

	// Term skills only. The extra skill a Commission or Promotion
	// grants is counted separately below, because p.65 exempts it from
	// the Operations-column restriction.
	skillCount := marineSkillsPerTerm
	exemptSkills := 0

	if term.Commissioned || term.Promoted {
		exemptSkills++

		if skill, ok := marineRankAutomaticSkill(isOfficer, tier); ok {
			term.SkillsAwarded = append(term.SkillsAwarded, skill)
		}
	}

	// Book 1 p.65: Term skills "may be taken on a column of the Skills
	// table corresponding to an Operations result received in the Term",
	// plus Personal, which "may always be rolled".
	columns := eligibleSkillColumns(operations, marineOperationsColumns)

	term.SkillsAwarded = append(term.SkillsAwarded, armedForcesTermSkills(
		r, upp, MarineCareerName, marineSkillTable, operations, columns, skillCount, exemptSkills)...)

	return term, upp
}
