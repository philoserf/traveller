package character

import (
	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// Training Course (#163) — p.59's Training Institutions counterpart to
// the Educational Institutions pipeline in education.go. p.59 prints the
// two as separate lists ("EDUCATIONAL INSTITUTIONS" / "TRAINING
// INSTITUTIONS"), and this one is not a good fit for another
// educationInstitution row: it has no Graduation or Degree, it is a
// repeatable enroll-then-roll loop rather than one multi-year multi-roll
// attendance, and its Mod accumulates across separate courses instead of
// within one. See PLAN.md and this file's own doc comments for the
// design reasoning.
//
// p.59: "it is possible for Humans to pursue Training Courses using
// Edu/2 in lieu of Training." p.62-63:
//
//	"Training Course. Sophonts with C5=Tra often use Training Courses
//	instead of Colleges and Universities. A focused training course
//	provides intensive hands-on experience in one specific skill or
//	knowledge... Each Training Course requires Check Int to enroll,
//	followed by a Pass/Fail Check Tra. A continual series of Training
//	courses imposes its own burden: the Pass/Fail Check Tra is subject
//	to Mod minus the number of Training courses taken. Failure prohibits
//	additional Training Courses. Successfully passing a Training Course
//	awards Skill-2 or Knowledge-2."
//
// Wired into laterEducationHook (career_loop.go) as a fallback tried only
// once the academic Later Education path has nothing to offer — the same
// "resolve toward the better outcome" convention chooseInstitution
// already establishes elsewhere in this package: College/University's
// own payoff (a degree, Major growth, a shot at Masters/Medical/Law) is
// always better than one Skill-2 when both are available.

// TrainingCourse is one attempt at Training Course (#163) — Education's
// own TrainingCourses field holds one of these per attempt, oldest first.
type TrainingCourse struct {
	// Passed is whether this attempt's own Pass/Fail Check succeeded — a
	// false here permanently bars any further Training Course attempt,
	// per trainingCourseBarred.
	Passed bool
	// Subject is the Skill or Knowledge enrolled in — chosen before the
	// Pass/Fail roll, so it is recorded even on a failed attempt (p.62's
	// own Barr Vech example still names "Legged" as what he failed).
	Subject string
	// SchoolNameRoll, SchoolName and SchoolRank mirror GraduateProgram's
	// own fields — set only when Passed, since only a passed attempt
	// reaches p.72's "For Each School Attended" step.
	SchoolNameRoll int
	SchoolName     string
	SchoolRank     int
}

// humanTra is a Human's own stand-in for Tra wherever the Training
// Institutions table calls for it — p.59's "Edu/2 in lieu of Training."
// p.52 gives the general rule for using an analog characteristic at
// half-value: "he can use the analog... at half-value (round fractions
// up)." Confirmed against Kyle Martin's own worked example (p.62,
// Apprenticeship, the same Edu/2 substitution): Edu=7 gives "Edu/2 = 4",
// not the 3 that plain integer division alone would produce.
func humanTra(edu ehex.Value) int {
	return (int(edu) + 1) / 2
}

// trainingCourseBarred is p.63's own permanent consequence: "Failure
// prohibits additional Training Courses." True only once the character's
// most recent course failed — a passed course, or no attempt at all,
// leaves the door open for another.
func trainingCourseBarred(edu Education) bool {
	n := len(edu.TrainingCourses)

	return n > 0 && !edu.TrainingCourses[n-1].Passed
}

// trainingCourseEligible applies the Human Edu/2-for-Tra substitution to
// Training Course's own "Tra 5+" Pre-Requisite (p.59's table): ceil(Edu/2)
// >= 5, i.e. Edu >= 9. Pure — draws no dice — so a character this
// declines costs nothing to the dice stream, the same contract
// shouldAttemptLaterEducation already holds for the academic path.
func trainingCourseEligible(upp UPP, edu Education) bool {
	return !trainingCourseBarred(edu) && humanTra(upp.Characteristics[C5]) >= 5
}

// attemptTrainingCourse runs one Training Course attempt, appending its
// own record to edu.TrainingCourses rather than overwriting Education's
// primary School/SchoolName fields — the same "For Each School Attended"
// (p.72) plurality GraduateProgram already models for the University
// follow-on tier.
//
// Waiver is deliberately not offered here, unlike every check in
// education.go: p.59's Educational Waiver clause is introduced under a
// heading naming "Schools and Education" specifically, Training Course's
// own paragraph (p.62-63) describes a complete adverse-consequence rule
// of its own ("Failure prohibits additional Training Courses") with no
// mention of Waiver, and the Barr Vech worked example (p.62, three
// courses run to a failure) shows no waiver attempt on any of them.
//
// The subject is drawn fresh from p.60's own "S" column
// (skillsForSchool(schoolTrade, ...), p.61's legend: "School including
// Apprentice, Training Course, Trade School") whether or not the
// enrollment goes on to pass — p.62's own Barr Vech example enrolls in
// "Legged" and only then rolls Pass/Fail, so the subject is chosen at
// enrollment, not awarded after the fact. No exclusion against a subject
// an earlier course already granted: the book states none, and a repeat
// draw simply adds a further level via aggregateSkills, same as any other
// repeat grant in this package.
//
// Only a passed course rolls a School Name (rollInstitution) — the same
// "only a graduate gets a name" convention attendInstitution's own
// Provides-column institutions already follow (its own doc comment: "a
// character who was refused admission never gets a school name"; an
// admitted-but-failed attempt gets none either, since rollInstitution
// there only runs on the Graduated path).
//
// Returns the skill this attempt granted (nil on a rejected enrollment or
// a failed course) and whether the character was admitted — the same
// "if accepted substitutes that process for the entire term" contract
// (p.59) attendInstitution's own admitted return already honors, and
// which laterEducationHook uses to decide whether this Later Education
// attempt consumes the term.
func attemptTrainingCourse(r *dice.Roller, upp UPP, edu *Education) ([]SkillLevel, bool) {
	if !checkAgainst(r, upp, []Position{C4}) {
		return nil, false
	}

	subject := pickSubjectExcluding(r, skillsForSchool(schoolTrade, false), "")
	mod := -len(edu.TrainingCourses)
	passed := rollAgainstTarget(r, humanTra(upp.Characteristics[C5]), mod)

	course := TrainingCourse{Passed: passed, Subject: subject}

	var granted []SkillLevel

	if passed {
		granted = []SkillLevel{{Name: subject, Level: 2, Kind: Skill}}
		edu.Skills = aggregateSkills(append(edu.Skills, granted...))
		course.SchoolNameRoll, course.SchoolName, course.SchoolRank = rollInstitution(r, "Training Course", subject)
	}

	edu.TrainingCourses = append(edu.TrainingCourses, course)

	return granted, true
}
