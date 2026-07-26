package character

import "github.com/philoserf/traveller/dice"

// rollRestrictedD6 rolls 1D6 and rerolls until the result is at most
// limit — Book 1's own "(reroll if >3)" written as the literal reroll
// loop it describes rather than as an equivalent Uniform(3).
//
// The two are not interchangeable here even though their distributions
// match: this codebase's determinism contract is the dice stream, so a
// seed that reproduces a character must consume the same number of D6
// draws the book's own procedure would. Both of Book 1's A/B/C tables
// (p.78's Citizen Skills and Knowledges, p.83's Agent Undercover
// Assignment) spell the reroll out in identical words, which is what
// justifies extracting it rather than writing it twice.
func rollRestrictedD6(r *dice.Roller, limit int) int {
	for {
		if roll := r.D6(); roll <= limit {
			return roll
		}
	}
}

// agentUndercoverAssignment is one cell of Book 1 p.83's own "AGENT
// UNDERCOVER ASSIGNMENT" table.
type agentUndercoverAssignment struct {
	// Service is the table's own left-column label. It is not always a
	// career name: the table names military services ("Army Enlisted",
	// "Navy Officer"), which Book 1's own p.64 career list implements
	// as the Soldier and Spacer careers.
	Service string
	// Career is the career whose skill table backs this assignment, per
	// the p.83 box's own "Select (not Roll) one skill from the skill
	// tables of that Career."
	Career string
	// Titles are the row's own C-column rank titles. Book 1 rolls "top
	// row C (reroll if >3) if required" — a row printing fewer than
	// three titles does not require it, which is how the table marks
	// Scout's own single "World Discover" cell and the four rows that
	// print no title at all (both Citizen rows, and Functionary).
	Titles []string
}

// agentUndercoverTableAColumns is the "reroll if >3" bound on both the A
// and C rolls: the table has three A blocks and three C columns.
const agentUndercoverTableAColumns = 3

// agentUndercoverTable is Book 1 p.83's own 18-row "AGENT UNDERCOVER
// ASSIGNMENT" table, indexed [A-1][B-1] for A 1-3 and B 1-6.
//
// Transcribed from the page directly; the .txt extraction interleaves
// the "C / 1 2 3" column header into the first data row, so Army
// Enlisted's own Private/Corporal/Sergeant appears scrambled there.
//
// Two of the book's own spellings are kept as printed rather than
// corrected. "Coronel" (Marine Officer C3) is not an OCR artifact: the
// book prints it three times, all in Marine contexts, against "Colonel"
// in Soldier's own rank table — a distinction this codebase already
// honors in marineOfficerRankNames. "World Discover" is the whole
// printed cell.
//
// One is corrected: "Pilor" (Merchant, A3 B2 C1) appears exactly once in
// the entire book against many "Pilot"s, and the Merchant skill tables
// name Pilot, so it is a typo rather than a term of art.
//
// Note what the table does *not* contain: there is no Rogue row and no
// Craftsman row, though both are careers on p.64. An Undercover
// Assignment therefore cannot be either one — the previous
// implementation's uniform pick over this codebase's own implemented
// careers included Rogue, which no cell of this table can produce.
var agentUndercoverTable = [3][6]agentUndercoverAssignment{
	{
		{"Army Enlisted", SoldierCareerName, []string{"Private", "Corporal", "Sergeant"}},
		{"Army Officer", SoldierCareerName, []string{"First Lieutenant", "Captain", "Major"}},
		{"Marine Enlisted", MarineCareerName, []string{"Corporal", "Sergeant", "Master Sergeant"}},
		{"Marine Officer", MarineCareerName, []string{"Captain", "Force Commander", "Coronel"}},
		{"Navy Enlisted", SpacerCareerName, []string{"Able Spacer", "Petty Officer First", "Master Chief"}},
		{"Navy Officer", SpacerCareerName, []string{"Lieutenant", "Commander", "Captain"}},
	},
	{
		{ScholarCareerName, ScholarCareerName, []string{"Amateur", "Lecturer", "Instructor"}},
		{ScholarCareerName, ScholarCareerName, []string{"Asst Professor", "Professor", "Dist Professor"}},
		{EntertainerCareerName, EntertainerCareerName, []string{"Musician", "Actor", "Dancer"}},
		{EntertainerCareerName, EntertainerCareerName, []string{"Artist", "Author", "Chef"}},
		{"Citizen Job", CitizenCareerName, nil},
		{"Citizen Hobby", CitizenCareerName, nil},
	},
	{
		{MerchantCareerName, MerchantCareerName, []string{"Engineer", "Astrogator", "Steward"}},
		{MerchantCareerName, MerchantCareerName, []string{"Pilot", "Freightmaster", "Counsellor"}},
		{scoutCareerName, scoutCareerName, []string{"Courier", "Courier", "Courier"}},
		{scoutCareerName, scoutCareerName, []string{"World Discover"}},
		{NobleCareerName, NobleCareerName, []string{"Knight", "Baron", "Marquis"}},
		{FunctionaryCareerName, FunctionaryCareerName, nil},
	},
}

// scoutCareerName is Scout's own Career.Name value. Unexported and
// declared here rather than beside the other CareerName constants
// because Scout predates that convention and spells its own name as a
// literal in career_generate.go; this exists so the Undercover table
// isn't the one place that reintroduces a bare "Scout" string.
const scoutCareerName = "Scout"

// rollAgentUndercoverAssignment resolves Book 1 p.83's own "Roll for
// Undercover Assignment" via the procedure the table itself prints:
// "Roll three dice for a specific Skill or Knowledge: Roll A (reroll if
// >3), then roll B, and finally top row C (reroll if >3) if required."
//
// C is rolled only when the row prints three titles, which is what "if
// required" means — rolling it unconditionally would still pick the
// right assignment but would consume a die the book's own procedure
// does not, desynchronizing every later roll in the term.
//
// Returns the assignment and the rendered title: the Service label
// alone for a row with no titles, otherwise the label plus the selected
// rank ("Navy Officer Commander", "Scout World Discover").
func rollAgentUndercoverAssignment(r *dice.Roller) (agentUndercoverAssignment, string) {
	a := rollRestrictedD6(r, agentUndercoverTableAColumns)
	b := r.D6()

	assignment := agentUndercoverTable[a-1][b-1]

	return assignment, renderUndercoverTitle(assignment, r)
}

// renderUndercoverTitle resolves the "top row C (reroll if >3) if
// required" half of the procedure: the Service label alone for a row
// printing no title, the label plus its single title for Scout's own
// "World Discover" row, and otherwise the label plus whichever of the
// three C-column titles the restricted roll selects.
func renderUndercoverTitle(assignment agentUndercoverAssignment, r *dice.Roller) string {
	switch len(assignment.Titles) {
	case 0:
		return assignment.Service
	case 1:
		return assignment.Service + " " + assignment.Titles[0]
	default:
		return assignment.Service + " " + assignment.Titles[rollRestrictedD6(r, len(assignment.Titles))-1]
	}
}

// agentUndercoverSkillTables maps each career the Undercover table can
// assign to that career's own skill table. Citizen is absent
// deliberately: its two rows say "Roll on Citizen Life Skills", which is
// p.78's own Citizen Skills and Knowledges (Table E, a 3-key A/B/C
// lookup) rather than a flat [7][6] career table, so
// rollAgentUndercoverSkill routes it to rollCitizenTableEName instead.
var agentUndercoverSkillTables = map[string][7][6]string{
	scoutCareerName:       scoutSkillTable,
	MarineCareerName:      marineSkillTable,
	SoldierCareerName:     soldierSkillTable,
	SpacerCareerName:      spacerSkillTable,
	ScholarCareerName:     scholarSkillTable,
	EntertainerCareerName: entertainerSkillTable,
	MerchantCareerName:    merchantSkillTable,
	NobleCareerName:       nobleSkillTable,
	FunctionaryCareerName: functionarySkillTable,
}

// rollAgentUndercoverSkill draws the one Undercover skill from career's
// own table, rerolling until a resolvable cell is drawn rather than
// reusing rollSkillsFromTable's own "unresolvable = lost" convention:
// Book 1's own "Select (not Roll) one skill from the skill tables of
// that Career" is singular and unconditional, with none of the "lost"
// language that governs the count-based grants elsewhere.
//
// The loop always terminates: column 0 of every career table is the
// Personal column, whose cells resolveSkillCell resolves
// unconditionally, so each iteration succeeds with probability at least
// 1/7. Citizen's own Table E has exactly one unresolvable cell
// ("No Skill") out of 108.
func rollAgentUndercoverSkill(r *dice.Roller, career string) SkillLevel {
	for {
		if career == CitizenCareerName {
			if name, ok := rollCitizenTableEName(r); ok {
				return skillLevel1(name, Skill)
			}

			continue
		}

		if skill, ok := rollSkillFromTable(r, agentUndercoverSkillTables[career]); ok {
			return skill
		}
	}
}
