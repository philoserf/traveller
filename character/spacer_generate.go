package character

import "github.com/philoserf/traveller/dice"

// SpacerCareerName is Spacer's own Career.Name value — exported and
// shared, matching SoldierCareerName's/MarineCareerName's own rationale.
const SpacerCareerName = "Spacer"

// spacerRiskRewardPositions is Spacer's own Risk & Reward Controlling
// Characteristic set (Book 1 p.81: "Risk & Reward C1 C2 C4") — Str, Dex,
// Int, a third distinct CC set from Marine's {C1,C4} and Soldier's
// {C1,C3,C4}.
var spacerRiskRewardPositions = []Position{C1, C2, C4}

// spacerBranchOfficerNames/spacerBranchOfficerMods/spacerBranchEnlistedNames/
// spacerBranchEnlistedMods are Book 1 p.81's own "NAVAL BRANCH" table (8
// rows) — a genuine second dimension Marine's/Soldier's own single-column
// Branch tables don't have: the same row's own name AND Mod can differ by
// current Officer/Enlisted status (row 6: Enlisted "Gunnery" Mod 1 becomes
// Officer "Flight" Mod 2 the moment Commission succeeds — Book 1 p.65's
// own "for Spacers, Crew becomes Line"), confirmed directly against the
// page image.
var spacerBranchOfficerNames = [8]string{
	"Line", "Line", "Line", "Engineer", "Gunnery", "Flight", "Technical", "Medical",
}
var spacerBranchOfficerMods = [8]int{1, 1, 1, 0, 1, 2, 0, 0}

var spacerBranchEnlistedNames = [8]string{
	"Crew", "Crew", "Engineer", "Engineer", "Gunnery", "Gunnery", "Technical", "Medical",
}
var spacerBranchEnlistedMods = [8]int{1, 1, 0, 0, 1, 1, 0, 0}

// spacerNavalOperationsNames/spacerNavalOperationsMods are Book 1 p.81's
// own "NAVAL OPERATIONS" table (8 rows, not Marine's/Soldier's 9) — no
// per-branch DM column at all, confirmed directly against the page
// image; only the universal "+2 if Edu 10+" applies. Rows 7-8 ("Shore
// Duty" twice) are only reachable with that Edu bonus, since a bare 1D6
// alone maxes at 6.
var spacerNavalOperationsNames = [8]string{
	"Battle", "Strike", "Siege", "Patrol", "Mission", "ANM School", "Shore Duty", "Shore Duty",
}
var spacerNavalOperationsMods = [8]int{2, 2, 0, 1, 3, 0, 0, 0}

// spacerSkillTable is Book 1 p.81's own "C SPACER SKILLS" table,
// transcribed directly from the page image.
var spacerSkillTable = [7][6]string{
	{"Str", "Dex", "End", "Int", "Edu", "Soc"},
	{"Major", "Major", "Minor", "Minor", "One Trade", "One Trade"},
	{"Fighter", "Fleet Tactics", "Pilot", "Starship Skill", "Gunner", "Sensors"},
	{"Astrogator", "Fleet Tactics", "Computer", "Starship Skill", "Gunner", "Sensors"},
	{"Computer", "Strategy", "Counsellor", "Gunner", "Gunner", "Gunner"},
	{"Diplomat", "Admin", "Language", "Starship Skill", "Liaison", "Comms"},
	{"One Art", "One Science", "Athlete", "Medic", "Zero-G", "One Trade"},
}

// spacerSkillsPerTerm is Book 1 p.81's own "Skill Eligibility: Per Term 4".
const spacerSkillsPerTerm = 4

// BeginSpacer reports Book 1 p.81's own "To Begin Int" — roll 2D <= Int,
// the first career in this codebase whose Begin check isn't Str-based.
// No Retry, matching Marine's/Soldier's own no-Retry treatment.
func BeginSpacer(r *dice.Roller, intChar int) bool {
	return rollAgainstTarget(r, intChar, 0)
}

// rollSpacerBranchRow resolves "Select Branch" (a genuine open player
// choice, the same convention already established for Marine's/
// Soldier's own Branch pick), returning a row index (0-7) rather than a
// resolved name+Mod — which one applies depends on the character's own
// current Officer/Enlisted status every term, not just once at Begin.
func rollSpacerBranchRow(r *dice.Roller) int {
	return r.Uniform(len(spacerBranchEnlistedNames)) - 1
}

// spacerBranchNameAndMod resolves row's own current name and Mod for
// isOfficer's own track.
func spacerBranchNameAndMod(row int, isOfficer bool) (string, int) {
	if isOfficer {
		return spacerBranchOfficerNames[row], spacerBranchOfficerMods[row]
	}

	return spacerBranchEnlistedNames[row], spacerBranchEnlistedMods[row]
}

// rollSpacerOperations rolls 4 times (p.81: "Spacer rolls 4 times per
// Term for Operations; select the highest Mod from the four"),
// delegating to the shared rollOperations/operationsEduDM
// (career_generate.go) — no branch DM term, unlike Marine's/Soldier's
// own Operations roll.
func rollSpacerOperations(r *dice.Roller, edu, rolls int) ([]string, string, int) {
	return rollOperations(r, operationsEduDM(edu), spacerNavalOperationsNames[:], spacerNavalOperationsMods[:], rolls)
}

// ResolveSpacerTerm resolves one 4-year Spacer term — mirrors
// ResolveSoldierTerm's own structure (soldier_generate.go), with two
// differences: branchRow (not a resolved branch/branchMod pair) is
// threaded in, since which name/Mod applies depends on this term's own
// current Officer/Enlisted status (spacerBranchNameAndMod, resolved
// fresh every term); and Spacer's own distinct target characteristics —
// Officer Promotion against Soc (matching Soldier, not Marine's Int),
// Officer Commission and Rating Promotion both against C2 (Dex, the
// same "Commission and Enlisted-Promotion share a target" shape
// Soldier's own C3 already established, here C2 instead).
//
// commissionedEntry and operationsRolls: see ResolveMarineTerm's own
// doc comment. Because Branch itself is officer/enlisted-dependent here
// (unlike Marine's/Soldier's own fixed-for-the-career Branch),
// commissionedEntry forcing isOfficer=true also fixes Branch resolution
// for term 1, not just Rank — a Commissioned Spacer's own branchRow (5
// for Flight School, see spacerFlightBranchRow) now correctly resolves
// to its Officer-side name (spacerBranchNameAndMod) instead of the
// Enlisted one a still-false isOfficer would have produced.
func ResolveSpacerTerm(
	r *dice.Roller, upp UPP, ccPos Position, branchRow int, priorTerms []Term,
	commissionedEntry bool, operationsRolls int,
) (Term, UPP) {
	operations, opName, opMod := rollSpacerOperations(r, int(upp.Characteristics[C5]), operationsRolls)

	isOfficer, tier := rankState(priorTerms, len(spacerEnlistedRankNames), len(spacerOfficerRankNames))
	if commissionedEntry {
		isOfficer, tier = true, 1
	}

	branch, branchMod := spacerBranchNameAndMod(branchRow, isOfficer)
	cc := upp.Characteristics[ccPos]
	mod := -(branchMod + opMod)
	term := Term{
		Length:                    operationsRolls,
		ControllingCharacteristic: ccPos,
		Branch:                    branch,
		Assignment:                opName,
		Operations:                operations,
		RewardResult:              "None",
		Rank:                      spacerRankName(isOfficer, tier),
		Commissioned:              commissionedEntry,
	}

	// See ResolveMarineTerm's own doc comment (granted ahead of resolveRisk).
	if commissionedEntry {
		grantAutoSkillIfAny(&term, spacerRankAutomaticSkill, true, 1)
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

	if !commissionedEntry {
		switch {
		case isOfficer:
			if tier < len(spacerOfficerRankNames) &&
				rollSpacerOfficerPromotion(r, int(upp.Characteristics[C6]), medalMod) {
				term.Promoted = true
				tier++
			}
		case rollSpacerCommission(r, int(upp.Characteristics[C2])):
			term.Commissioned = true
			isOfficer = true
			tier = 1
		case tier < len(spacerEnlistedRankNames) && rollSpacerRatingPromotion(r, int(upp.Characteristics[C2]), medalMod):
			term.Promoted = true
			tier++
		}
	}

	term.Rank = spacerRankName(isOfficer, tier)

	// Term skills only — the Commission/Promotion skill is exempted below.
	skillCount := spacerSkillsPerTerm
	exemptSkills := armedForcesTermExemptSkills(&term, commissionedEntry, isOfficer, tier, spacerRankAutomaticSkill)

	// Book 1 p.65: Term skills "may be taken on a column of the Skills
	// table corresponding to an Operations result received in the Term",
	// plus Personal, which "may always be rolled".
	columns := eligibleSkillColumns(operations, spacerOperationsColumns)

	term.SkillsAwarded = append(term.SkillsAwarded, armedForcesTermSkills(
		r, upp, SpacerCareerName, spacerSkillTable, operations, columns, skillCount, exemptSkills)...)

	return term, upp
}
