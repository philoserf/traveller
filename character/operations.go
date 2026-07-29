package character

import (
	"slices"

	"github.com/philoserf/traveller/dice"
)

// Book 1 p.65's own "Spacer-Soldier-Marine Term Skills" rule:
//
//	"In Spacer, Soldier, or Marine careers, Term skills (but not
//	commission, promotion, or other skill eligibilities) may be taken on
//	a column of the Skills table corresponding to an Operations result
//	received in the Term."
//
// with p.65's own preceding line, which is why personalSkillColumn is
// always in the eligible set: "Column 1-Personal Skills may always be
// rolled."
//
// The parenthetical is as load-bearing as the rule. p.65's own worked
// example has Eneri roll Naval Operations (4, 1, 5) for Patrol, Battle
// and Mission, then say he "can consult the columns for 1 Personal, 3
// Battle, 4 Patrol/Strike, 5 Siege, 6 Mission" — five columns for four
// Term skills plus one promotion skill. Siege is in that list precisely
// because it is not a Term skill drawing it: the promotion skill the
// rule exempts can come from any column at all.
const personalSkillColumn = 0

// operationsColumns maps a career's own Operations result names to the
// column of its Skills table that result makes eligible.
//
// The three careers' column layouts are almost but not quite parallel
// (Spacer's own column 2 is Shore Duty where Soldier has Base and Marine
// has Garrison), so each career carries its own map rather than sharing
// one keyed by position.
//
// Two properties are worth naming because they look like bugs and are
// not. Spacer's Patrol and Strike both map to column 3, which the table
// prints as the single heading "4 Patrol/Strike". And no Operations
// result in any of the three careers maps to Occupation/Siege (column 4)
// or Technical (column 6) — those columns are reachable only through the
// commission, promotion and automatic-skill eligibilities the rule
// exempts, which is exactly what the p.65 example shows happening.
//
// "ANM School" is absent from all three deliberately: the Operations
// tables resolve it as Education rather than as an assignment, so it
// makes no skill column eligible.
var (
	// Marine columns: 0 Personal, 1 Garrison, 2 Combat, 3 Peacekeeper,
	// 4 Occupation, 5 Mission, 6 Technical (Book 1 p.86).
	marineOperationsColumns = map[string]int{
		"Garrison": 1, "Combat": 2, "Peace Keeper": 3, "Mission": 5,
	}
	// Soldier columns: 0 Personal, 1 Base, 2 Combat, 3 Peacekeeper,
	// 4 Occupation, 5 Mission, 6 Technical (Book 1 p.82).
	soldierOperationsColumns = map[string]int{
		"Base": 1, "Combat": 2, "Peace Keeper": 3, "Mission": 5,
	}
	// Spacer columns: 0 Personal, 1 Shore Duty, 2 Battle,
	// 3 Patrol/Strike, 4 Siege, 5 Mission, 6 Technical (Book 1 p.81).
	spacerOperationsColumns = map[string]int{
		"Shore Duty": 1, "Battle": 2, "Patrol": 3, "Strike": 3, "Siege": 4, "Mission": 5,
	}
)

// operationsRollsPerTerm is Book 1's own "Rolls 4 times per Term for
// Operations", stated identically on Spacer's, Soldier's and Marine's
// own pages.
const operationsRollsPerTerm = 4

// rollOperations resolves one term's Operations: rolls 1D against the
// career's own table rolls times, returning every assignment received
// and the single highest associated Mod. Every caller passes
// operationsRollsPerTerm except a Flight-School-shortened first term
// (#113), which rolls one fewer — p.61's own worked example: "Because
// Flight School took a year, this first Term is reduced to three years:
// he rolls 1D three times for Naval Operations."
//
// Both halves matter, and this codebase previously kept only the second.
// The Mod is what Risk & Reward uses — p.65's own example takes
// "assignments in Patrol, Battle, and Mission. The highest associated
// Mod is 3" — while the assignments themselves decide which Skills
// columns the Term's skills may be drawn from. Discarding them meant
// term skills could come from an Operation the character never received.
//
// Duplicates are kept rather than deduplicated: they are a real outcome
// (p.65's own second example rolls "3, 3, 4, 4" for "Siege, Siege,
// Patrol, Patrol") and callers that only care about eligibility dedupe
// themselves via eligibleSkillColumns.
// Returns every assignment received, the name of the highest-Mod one
// (the term's headline duty, recorded on Term.Assignment) and that Mod.
// Ties go to the first rolled.
func rollOperations(r *dice.Roller, dm int, names []string, mods []int, rolls int) ([]string, string, int) {
	received := make([]string, 0, rolls)
	best := 0

	for i := range rolls {
		row := musterOutRow(r.D6()+dm, len(names))
		received = append(received, names[row])

		if i == 0 || mods[row] > mods[best] {
			best = row
		}
	}

	return received, names[best], mods[best]
}

// operationMod is the Mod an Operations result name carries. Several
// rows of each table share a name at different Mods (Marine's own Combat
// appears at 2, 2 and 3), so this reports the highest any row with that
// name carries — the most that name could have contributed.
func operationMod(name string, names []string, mods []int) int {
	best, found := 0, false

	for i, n := range names {
		if n != name {
			continue
		}

		if !found || mods[i] > best {
			best, found = mods[i], true
		}
	}

	return best
}

// eligibleSkillColumns is the set of Skills-table columns a Term's own
// skills may be drawn from: Personal, which p.65 says "may always be
// rolled", plus one column per distinct Operations result received.
//
// Returned sorted and deduplicated so the column drawn from is a
// function of which Operations were received, not of the order they were
// rolled in — two terms receiving the same set of assignments must offer
// the same choices.
//
// A result the career maps to no column (ANM School) contributes
// nothing, so a term of nothing but ANM School leaves only Personal —
// which is the rule working, not a gap.
func eligibleSkillColumns(received []string, columnsFor map[string]int) []int {
	columns := []int{personalSkillColumn}

	for _, name := range received {
		column, ok := columnsFor[name]
		if !ok {
			continue
		}

		if !slices.Contains(columns, column) {
			columns = append(columns, column)
		}
	}

	slices.Sort(columns)

	return columns
}

// rollSkillsFromColumns is rollSkillsFromTable restricted to columns —
// the Term-skill draw for the three Armed Forces careers. Same
// "unresolvable entry grants nothing" convention as the unrestricted
// form (the book's own wording is that such a benefit is "lost"), so the
// returned slice can be shorter than count.
//
// Draws two dice per skill, one for the column and one for the row, the
// same shape as the unrestricted form. Total dice per skill still varies
// with the cell landed on — "One Trade" and "Any Skill" resolve
// themselves with further rolls — so switching a career onto this
// function shifts the stream, it does not preserve it.
func rollSkillsFromColumns(r *dice.Roller, table [7][6]string, columns []int, count int) []SkillLevel {
	if len(columns) == 0 {
		return nil
	}

	var skills []SkillLevel

	for range count {
		column := columns[r.Uniform(len(columns))-1]

		if skill, ok := resolveSkillCell(r, table, column, r.Uniform(6)-1); ok {
			skills = append(skills, skill)
		}
	}

	return skills
}
