package character

import (
	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// Book 1 p.72's CharGen step C, "PRE-CAREER EDUCATION (AND TRAINING)",
// which sits between B Determine A Homeworld and D Select Career.
//
// The governing description, p.59:
//
//	"Education is a multi-step process. If Pre-Requisites are met, the
//	character Applies for Admission. If successful, the character rolls
//	for Pass/Fail for each year of the process. Pass awards one of the
//	available skills; Failure terminates the process (but Waiver may
//	result in reinstatement, although no skill is received). Finally, a
//	character who Graduates (who Passes or who has Failure Waived)
//	receives Graduation benefits."
//
// This is the first half of that step: the academic spine — ED5, Trade
// School, College and University, plus the Honors program — with the
// Apply / Pass-Fail / Waiver / Graduation machinery they share, and
// Major and Minor selection.
//
// What is deliberately not here yet, because each pulls in a mechanic of
// its own rather than more of this one:
//
//   - Service Academy, OTC and NOTC confer a military Commission and
//     oblige a term of service, which is a career interaction rather
//     than a schooling one.
//   - Masters, Professors, Medical School and Law School gate on a
//     degree this step now produces (BA, MA, Honors BA), so they chain
//     off it rather than sitting beside it.
//   - Flight School gates on an Honors BA and grants a Branch.
//   - Apprenticeship, Mentor and Training Course are the Tra path.
//     p.59 puts Sophonts with Tra there and notes Humans may use
//     Training Courses "using Edu/2 in lieu of Training"; GenerateUPP
//     only produces Humans, whose C5 is always Edu, so none of the three
//     has a caller yet.
//
// Also not here: the Major and Minor selected below do not yet resolve
// the "Major"/"Minor" cells of the thirteen career skill tables
// (career_generate.go's own resolveSkillCell). Those cells are read deep
// inside per-term skill resolution, which has no view of the character;
// giving it one is a mechanical change across every career and is kept
// separate from the rules work here.

// educationInstitution is one row of p.60's own "EDUCATION FOR
// CHARACTERS" table, recovered from the PDF's word bounding boxes
// against the header's column origins — name at x=55, Pre-Requisites at
// 134, To Apply at 183, Duration at 227, Pass/Fail at 265, Rolls at 309,
// Provides at 325 and Graduation at 437.
type educationInstitution struct {
	Name string
	// MinEdu is the "Edu 5+" style prerequisite, and MaxEdu the "Edu 4 -"
	// ceiling ED5 alone carries. Zero means unbounded in that direction.
	MinEdu, MaxEdu ehex.Value
	// ApplyCheck is the To Apply column; nil is that column's "auto".
	ApplyCheck []Position
	PassCheck  []Position
	// Rolls is the How Many Rolls? column — one Pass/Fail check per year
	// of the process.
	Rolls int
	// GraduationEdu is the Graduation column's own "Edu=8" target. p.59's
	// aside beside that column — "(If Edu already at this level, award
	// Edu+1)" — is what stops a graduate ever going backwards.
	GraduationEdu ehex.Value
	// Degree is what Graduation confers, if anything.
	Degree string
	// GrantsMajorMinor marks p.60's own merged Provides cell, "Major+1
	// per Pass and Minor+1 per 2 Passes", which the printed table braces
	// across College, University, Service Academy and Masters together.
	GrantsMajorMinor bool
	// MajorBonus is Trade School's own flat "Major+2" instead.
	MajorBonus int
}

// educationInstitutions are the rows this step can send a character to,
// in the order a character prefers them: the most demanding school he
// qualifies for. Which school to attend is an open choice Book 1 leaves
// to the player, resolved here toward the better outcome — the same
// treatment nextCC and BeginScout's own Begin-target choice already get,
// and for the same reason: a genuinely better option exists, so a
// uniform random pick would be a worse default than the obvious one.
var educationInstitutions = []educationInstitution{
	{
		Name: "University", MinEdu: 7,
		ApplyCheck: []Position{C4, C5}, PassCheck: []Position{C4, C5}, Rolls: 4,
		GraduationEdu: 9, Degree: educationDegreeBachelors, GrantsMajorMinor: true,
	},
	{
		Name: "College", MinEdu: 5,
		ApplyCheck: []Position{C4, C5}, PassCheck: []Position{C4, C5}, Rolls: 4,
		GraduationEdu: 8, Degree: educationDegreeBachelors, GrantsMajorMinor: true,
	},
	{
		Name: "Trade School", MinEdu: 5,
		ApplyCheck: []Position{C4}, PassCheck: []Position{C4, C5}, Rolls: 1,
		MajorBonus: 2,
	},
	{
		// "Edu 4 -" — a floor-raising program for characters below the
		// Trade School minimum. p.60: "A character with Edu less than 5
		// can attempt the ED5 program at the start the Education process.
		// Check Int: if successful, Edu is raised to 5."
		Name: "ED5", MaxEdu: 4,
		PassCheck: []Position{C4}, Rolls: 1,
		GraduationEdu: 5,
	},
}

// Degrees p.60's Graduation column confers.
const (
	educationDegreeBachelors = "BA"
	educationDegreeHonours   = "Honors BA"
)

// Education is what CharGen step C did to one character.
type Education struct {
	// School is the institution attended, empty if none was.
	School string
	// Major and Minor are p.59's own "The character attending an
	// Educational Institution must select a Major and a Minor from the
	// appropriate Skill and Knowledge list."
	Major, Minor string
	// Passes is how many of the institution's years were passed.
	Passes int
	// Graduated is p.59's "An individual who has passed all years
	// Graduates and receives the Graduation benefits" — including a
	// character whose only failure was Waived.
	Graduated bool
	// Honors records the optional extra pass/fail of p.59's Honors row.
	Honors bool
	// Degree is what Graduation conferred, if anything.
	Degree string
	// Waivers is how many Waiver rolls were attempted, successful or not
	// — p.59's own Mod is "minus number of previous waivers rolled
	// (successful or not)", so the count includes the failures.
	Waivers int
	// Skills are the levels the process awarded, already merged.
	Skills []SkillLevel
}

// Attended reports whether this character went to school at all.
func (e Education) Attended() bool {
	return e.School != ""
}

// resolveEducation runs step C for one character and returns the result
// alongside the possibly-improved UPP.
//
// Draws no dice at all for a character no institution will take, which
// keeps the reproducibility contract intact for them: p.60's own lowest
// bar is ED5's "Edu 4 -", and every character sits either under that or
// over Trade School's "Edu 5+", so in practice one institution always
// applies. The nil return is kept anyway rather than assumed away.
func resolveEducation(r *dice.Roller, upp UPP) (Education, UPP) {
	school, ok := chooseInstitution(upp)
	if !ok {
		return Education{}, upp
	}

	edu := Education{School: school.Name}

	// p.59: "To Apply (for Admission), a character must Check one of the
	// stated Characteristics. A failure disallows admission and consumes
	// one year." A Waiver can still get him in — p.60's own worked
	// example is exactly that: Eneri is rejected by the College of Regina
	// on Check Edu, applies for a Waiver, and is admitted.
	if len(school.ApplyCheck) > 0 && !checkAgainst(r, upp, school.ApplyCheck) && !tryWaiver(r, upp, &edu.Waivers) {
		return edu, upp
	}

	runEducationYears(r, upp, school, &edu)

	if !edu.Graduated {
		return edu, upp
	}

	upp = applyGraduation(upp, school, &edu)
	resolveHonors(r, upp, school, &edu)

	edu.Skills = aggregateSkills(edu.Skills)

	return edu, upp
}

// runEducationYears is p.59's own per-year Pass/Fail loop: "A character
// must Pass the required number of times (typically once for each year
// of attendance). Check the stated Characteristic the stated number of
// times. Each Success is one year; success in all rolls awards
// Graduation. Failure ends attendance (subject to Waiver)."
//
// A Waived failure reinstates the character but awards nothing — "no
// skill is received" — so it costs the year without advancing the Major.
func runEducationYears(r *dice.Roller, upp UPP, school educationInstitution, edu *Education) {
	for range school.Rolls {
		if checkAgainst(r, upp, school.PassCheck) {
			edu.Passes++
			awardEducationSkills(r, school, edu)

			continue
		}

		if !tryWaiver(r, upp, &edu.Waivers) {
			return
		}
	}

	edu.Graduated = true
}

// awardEducationSkills applies p.60's Provides column for one passed
// year: "Major+1 per Pass and Minor+1 per 2 Passes" for the degree
// programs, Trade School's flat "Major+2" instead.
//
// The Major is declared on the first Pass and the Minor on the second,
// which is what p.60's own worked example does — Eneri "declares his
// Major is Psychology" on his freshman pass, and "with his second Pass
// declares his Minor as Robotics". Both are drawn from the College
// column of p.60's skills matrix, per p.59's "select a Major and a Minor
// from the appropriate Skill and Knowledge list"; they cannot be the
// same, which p.59 states outright.
func awardEducationSkills(r *dice.Roller, school educationInstitution, edu *Education) {
	if edu.Major == "" {
		edu.Major = pickEducationSubject(r, "")
	}

	switch {
	case school.MajorBonus > 0:
		edu.Skills = append(edu.Skills, SkillLevel{Name: edu.Major, Level: school.MajorBonus, Kind: Skill})
	case school.GrantsMajorMinor:
		edu.Skills = append(edu.Skills, skillLevel1(edu.Major, Skill))

		if edu.Passes%2 == 0 {
			if edu.Minor == "" {
				edu.Minor = pickEducationSubject(r, edu.Major)
			}

			edu.Skills = append(edu.Skills, skillLevel1(edu.Minor, Skill))
		}
	}
}

// pickEducationSubject draws a Major or Minor from the College column of
// p.60's matrix, excluding one already-taken subject — p.59: a character
// "may select any Major and Minor (they cannot be the same)".
func pickEducationSubject(r *dice.Roller, exclude string) string {
	pool := skillsForSchool(schoolCollege, false)

	available := make([]string, 0, len(pool))

	for _, name := range pool {
		if name != exclude {
			available = append(available, name)
		}
	}

	if len(available) == 0 {
		return ""
	}

	return available[r.Uniform(len(available))-1]
}

// applyGraduation applies p.60's Graduation column. Its own aside — "(If
// Edu already at this level, award Edu+1)" — is why a graduate whose Edu
// already meets the target still gains a point rather than standing
// still.
func applyGraduation(upp UPP, school educationInstitution, edu *Education) UPP {
	edu.Degree = school.Degree

	if school.GraduationEdu == 0 {
		return upp
	}

	if upp.Characteristics[C5] >= school.GraduationEdu {
		upp.Characteristics[C5] = awardCharacteristic(upp.Characteristics[C5], 1)

		return upp
	}

	upp.Characteristics[C5] = school.GraduationEdu

	return upp
}

// resolveHonors is p.59's own Honors row: "A character can optionally
// make one additional pass/fail roll: success confers one level of his
// Major and confers Honors status. Failure has no effect."
//
// Offered only where there is a Major to advance, which is the degree
// programs. "Failure has no effect" is taken literally — no Waiver is
// attempted, because there is no adverse consequence to waive.
func resolveHonors(r *dice.Roller, upp UPP, school educationInstitution, edu *Education) {
	if !school.GrantsMajorMinor || edu.Major == "" {
		return
	}

	if !checkAgainst(r, upp, school.PassCheck) {
		return
	}

	edu.Honors = true
	edu.Degree = educationDegreeHonours
	edu.Skills = append(edu.Skills, skillLevel1(edu.Major, Skill))
}

// tryWaiver is p.59's own Educational Waiver:
//
//	"A student attending an Education Institution and who receives an
//	adverse die roll or decision (Prerequisite, Application Check,
//	Pass/Fail Check, Honors) may try for a Waiver. To receive a Waiver,
//	Check Soc or less (2D); Mod minus number of previous waivers rolled
//	(successful or not). Waivers are unique to Education and apply only
//	to Schools and Education (and to the Scholar career, but not other
//	careers)."
//
// Shared with the Scholar career, which p.59 names as the one career
// entitled to Waivers and which p.76 restates in its own terms: "An
// adverse die roll or decision (in Position, Promotion, Research,
// Publication, Tenure, or Continue) may be waived. Check Soc (2D); Mod
// minus previous waivers (successful or not)." The two statements agree
// exactly, so one implementation serves both.
//
// The Mod counts every waiver rolled rather than every waiver granted,
// which p.59 says in as many words and p.60's own worked example
// demonstrates: Eneri's fourth attempt carries Mod -4 having succeeded
// at only two of the previous three.
func tryWaiver(r *dice.Roller, upp UPP, waivers *int) bool {
	mod := -*waivers
	*waivers++

	return succeedsAgainst(r.TwoD6(), int(upp.Characteristics[C6]), mod)
}

// checkAgainst rolls 2D against the better of the given characteristics
// — p.60's own "Int or Edu" columns, resolved toward the higher of the
// two the same way every other open "X or Y" choice in this package is.
func checkAgainst(r *dice.Roller, upp UPP, positions []Position) bool {
	return succeedsAgainst(r.TwoD6(), int(upp.Characteristics[highestOf(upp, positions...)]), 0)
}

// chooseInstitution returns the most demanding school on p.60's table
// whose prerequisite this character meets.
func chooseInstitution(upp UPP) (educationInstitution, bool) {
	edu := upp.Characteristics[C5]

	for _, school := range educationInstitutions {
		if school.MinEdu > 0 && edu < school.MinEdu {
			continue
		}

		if school.MaxEdu > 0 && edu > school.MaxEdu {
			continue
		}

		return school, true
	}

	return educationInstitution{}, false
}

// generateStart is CharGen steps A, B and C in order: create the UPP,
// determine a homeworld and its skills, then run Education.
//
// Extracted so all twelve entry points — GenerateCareerChainCharacter
// and the eleven standalone Generate<X>Character functions — draw the
// same dice in the same order. That is not a tidiness point: the two
// paths are asserted to produce identical characters from identical
// seeds (TestCareerChainSingleEntryMatchesLegacyGenerator), so step C
// sitting one draw later on one path than the other would fail that test
// with a whole-struct diff.
//
// The Education's own skill grants are folded in beside the homeworld
// skills, which is where a character's pre-career skills already live.
func generateStart(r *dice.Roller) (UPP, string, []SkillLevel, Education) {
	upp := GenerateUPP(r)
	homeworld, skills := GenerateHomeworldSkills(r)

	edu, upp := resolveEducation(r, upp)

	return upp, homeworld, append(skills, edu.Skills...), edu
}

// Cell names in the thirteen career skill tables that stand for the
// character's own declared Major and Minor rather than for a skill.
//
// Book 1 prints them in 48 of the 546 cells — 8.8% of every career skill
// table — and footnotes what happens without one: "If the character does
// not have a Major/Minor this benefit is lost".
const (
	majorSkillCell = "Major"
	minorSkillCell = "Minor"
)

// resolveMajorMinorSkills replaces the Major and Minor markers a
// career's skill tables recorded with the subjects this character
// actually declared in CharGen step C, and drops them for a character
// who declared none — Book 1's own "this benefit is lost".
//
// Done at final assembly rather than at the point the cell is rolled,
// because the subject belongs to the character while the roll sees only
// a table. The markers stay on the Term as the career's own record of
// what it granted, which is accurate: the term awarded "a level of your
// Major", and which subject that was is a fact about the character.
func resolveMajorMinorSkills(skills []SkillLevel, edu Education) []SkillLevel {
	resolved := make([]SkillLevel, 0, len(skills))

	for _, s := range skills {
		switch s.Name {
		case majorSkillCell:
			if edu.Major != "" {
				s.Name = edu.Major

				resolved = append(resolved, s)
			}
		case minorSkillCell:
			if edu.Minor != "" {
				s.Name = edu.Minor

				resolved = append(resolved, s)
			}
		default:
			resolved = append(resolved, s)
		}
	}

	return resolved
}

// applyEducation attaches a finished step C result to a finished
// Character and resolves the Major/Minor markers its careers recorded.
//
// Re-aggregates afterwards because resolution can merge: a character
// whose Major is Psychology and whose career also granted Psychology
// separately holds one entry at the summed level, not two.
func applyEducation(c Character, edu Education) Character {
	c.Education = edu
	c.Skills = aggregateSkills(resolveMajorMinorSkills(c.Skills, effectiveMajorMinor(c, edu)))

	return c
}

// effectiveMajorMinor is the Major and Minor the character's skill cells
// resolve against: the ones a degree conferred, or failing that the ones
// a Scholar career declared on its own.
//
// The fallback is Book 1 p.76 taken at its word — "Every Scholar has a
// Major and a Minor. If no degree (and an associated Major and Minor)
// then select any Skill or Knowledge from the Skills List" — and it is
// not a corner case: 461 of 3,000 generated Scholars reach the career
// without a degree, between them drawing 1,830 Major/Minor cells that
// would otherwise be discarded against a rule that says they have one.
//
// Only Scholar declares its own. For every other career the cells stay
// governed by Book 1's "If the character does not have a Major/Minor
// this benefit is lost".
func effectiveMajorMinor(c Character, edu Education) Education {
	if edu.Major != "" && edu.Minor != "" {
		return edu
	}

	for _, career := range c.Careers {
		if career.Name != ScholarCareerName || len(career.Terms) == 0 {
			continue
		}

		if edu.Major == "" {
			edu.Major = career.Major
		}

		if edu.Minor == "" {
			edu.Minor = career.Minor
		}

		break
	}

	return edu
}

// allEducationSubjects is every distinct skill or knowledge p.60's
// matrix names, which is the closest thing this codebase has to Book 1's
// own "Skills List". Used where a character selects a subject freely
// rather than being granted one by a school.
func allEducationSubjects() []string {
	names := make([]string, 0, len(educationSkills))
	seen := make(map[string]bool, len(educationSkills))

	for _, skill := range educationSkills {
		if seen[skill.Name] {
			continue
		}

		seen[skill.Name] = true

		names = append(names, skill.Name)
	}

	return names
}

// pickSubjectExcluding draws from pool, avoiding one already-taken
// subject — p.59's own "they cannot be the same".
func pickSubjectExcluding(r *dice.Roller, pool []string, exclude string) string {
	available := make([]string, 0, len(pool))

	for _, name := range pool {
		if name != exclude {
			available = append(available, name)
		}
	}

	if len(available) == 0 {
		return ""
	}

	return available[r.Uniform(len(available))-1]
}
