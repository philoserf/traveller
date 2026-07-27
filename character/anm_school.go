package character

import "github.com/philoserf/traveller/dice"

// Book 1 p.59's ANM School row, read against the header's own column
// boundaries:
//
//	ANM School | assigned | auto | 1 year | C2 or C3 | 1x | Knowledge-2 from School=ANM
//
// so: assigned rather than chosen, admission automatic, one year, a
// single Pass/Fail check against the better of Dex and End, and on
// success one knowledge at level 2 drawn from the School=ANM columns of
// p.60's matrix.
//
// "ANM" is Army/Navy/Marine, and p.61's legend gives each its own School
// column — A Army School, N Naval School, M Marine School. This is the
// column Marine finally uses: Command College sends Marine officers to
// the Naval Academy (p.61 commissions them there), but ANM School is a
// School rather than an Academy, so a Marine attends the Marine one.
//
// Until now an "ANM School" Operations result mapped to no skill column
// and so contributed nothing at all — the Operations tables' own aside,
// "Resolve ANM School as Education", had no Education to resolve
// against. It does now.
//
// One limitation is recorded rather than papered over. p.60 marks
// knowledge-only entries in bold ("Bold= Knowledge-Only skill."), and
// boldness does not survive into the text extract or the bounding-box
// output, so the pool here is every skill the column flags rather than
// only the knowledges among them. The level and the source column are
// right; the Knowledge/Skill split within the pool is not yet
// recoverable, and the grant is recorded as a Knowledge because that is
// what p.59's Provides column says it awards.

// anmSchoolOperation is the Operations-table result that sends a
// character to school. It has to match the three careers' own tables
// exactly (marine/soldier/spacer_generate.go).
const anmSchoolOperation = "ANM School"

// anmSchoolSkillLevel is the "-2" in p.59's "Knowledge-2".
const anmSchoolSkillLevel = 2

// anmSchoolFor is the School column each Armed Force attends.
func anmSchoolFor(careerName string) (schoolSet, bool) {
	switch careerName {
	case SoldierCareerName:
		return schoolMilitaryAcademy, true // A, Army School
	case SpacerCareerName:
		return schoolNavalAcademy, true // N, Naval School
	case MarineCareerName:
		return schoolMarine, true // M, Marine School
	default:
		return 0, false
	}
}

// resolveANMSchool awards p.59's "Knowledge-2 from School=ANM" for a
// term whose Operations sent the character to school, or nothing if the
// Pass/Fail check failed.
//
// Draws no dice unless the term actually received the assignment, so
// every other term's stream is untouched. Repeats are resolved
// separately: a term can receive ANM School more than once (the tables
// print it once per career but Operations are rolled several times a
// term), and p.59 gives one Pass/Fail check per attendance.
func resolveANMSchool(r *dice.Roller, upp UPP, careerName string, received []string) []SkillLevel {
	school, ok := anmSchoolFor(careerName)
	if !ok {
		return nil
	}

	pool := skillsForSchool(school)
	if len(pool) == 0 {
		return nil
	}

	target := int(upp.Characteristics[highestOf(upp, C2, C3)])

	var awarded []SkillLevel

	for _, result := range received {
		if result != anmSchoolOperation {
			continue
		}

		if !succeedsAgainst(r.TwoD6(), target, 0) {
			continue
		}

		awarded = append(awarded, SkillLevel{
			Name:  pool[r.Uniform(len(pool))-1],
			Level: anmSchoolSkillLevel,
			Kind:  Knowledge,
		})
	}

	return awarded
}

// anmSchoolAwards counts the p.59 "Knowledge-2" grants an ANM School
// assignment added to a term.
//
// Identified by shape rather than tracked separately: every other skill
// grant in this package is a flat level 1 (skillLevel1), so a level-2
// Knowledge on an Armed Forces term can only have come from school. The
// three careers' own per-term skill-count assertions use this to
// subtract the school year, since Operations decide whether a term has
// one and the count is therefore not fixed.
func anmSchoolAwards(term Term) int {
	n := 0

	for _, s := range term.SkillsAwarded {
		if s.Kind == Knowledge && s.Level == anmSchoolSkillLevel {
			n++
		}
	}

	return n
}

// armedForcesTermSkills is the skill block Marine, Soldier and Spacer
// each resolve identically, extracted once ANM School made a third
// verbatim copy of it — the same "generalize on the second instance"
// discipline rollSkillsFromColumns and musterOutRow already followed.
//
// Three grants in Book 1's own order: p.59's ANM School award, if this
// term's Operations sent the character to one; then p.65's term skills,
// restricted to the columns the term's Operations made eligible; then the
// Commission or Promotion skill, which p.65 exempts from that restriction
// and which therefore draws from the whole table — the exemption that
// lets p.65's own worked example reach a column its Operations never
// granted.
func armedForcesTermSkills(
	r *dice.Roller,
	upp UPP,
	careerName string,
	table [7][6]string,
	operations []string,
	columns []int,
	skillCount, exemptSkills int,
) []SkillLevel {
	skills := resolveANMSchool(r, upp, careerName, operations)
	skills = append(skills, rollSkillsFromColumns(r, table, columns, skillCount)...)

	return append(skills, rollSkillsFromTable(r, table, exemptSkills)...)
}
