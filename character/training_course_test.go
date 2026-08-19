package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
)

// TestHumanTraRoundsFractionsUp pins the rounding direction against p.62's
// own worked example (Kyle Martin, Apprenticeship, the same Edu/2
// substitution): Edu=7 gives Tra-in-lieu 4, not the 3 plain integer
// division would produce.
func TestHumanTraRoundsFractionsUp(t *testing.T) {
	t.Parallel()

	if got := humanTra(7); got != 4 {
		t.Errorf("humanTra(7) = %d, want 4 (Kyle Martin: Edu/2 = 4, round fractions up)", got)
	}
}

// TestTrainingCourseEligibleRequiresEduNine confirms Training Course's
// "Tra 5+" Pre-Requisite, applied through the Human Edu/2 substitution,
// lands at Edu 9 (ceil(9/2)=5) and not Edu 8 (ceil(8/2)=4 still fails).
func TestTrainingCourseEligibleRequiresEduNine(t *testing.T) {
	t.Parallel()

	if trainingCourseEligible(eduUPP(0, 8, 0), Education{}) {
		t.Error("trainingCourseEligible(Edu 8) = true, want false (Tra-in-lieu 4 < 5)")
	}

	if !trainingCourseEligible(eduUPP(0, 9, 0), Education{}) {
		t.Error("trainingCourseEligible(Edu 9) = false, want true (Tra-in-lieu 5 >= 5)")
	}
}

// TestTrainingCourseEligibleDeclinesWhenBarred confirms p.63's own
// permanent consequence — "Failure prohibits additional Training
// Courses" — overrides an otherwise-qualifying Edu.
func TestTrainingCourseEligibleDeclinesWhenBarred(t *testing.T) {
	t.Parallel()

	edu := Education{TrainingCourses: []TrainingCourse{{Passed: false, Subject: "Rotor"}}}

	if trainingCourseEligible(eduUPP(0, 20, 0), edu) {
		t.Error("trainingCourseEligible = true after a failed course, want false (permanently barred)")
	}
}

// TestTrainingCourseEligibleAfterAPass confirms a passed course does not
// bar a further attempt — only a failure does (trainingCourseBarred's
// own contract).
func TestTrainingCourseEligibleAfterAPass(t *testing.T) {
	t.Parallel()

	edu := Education{TrainingCourses: []TrainingCourse{{Passed: true, Subject: "Wheeled"}}}

	if !trainingCourseEligible(eduUPP(0, 20, 0), edu) {
		t.Error("trainingCourseEligible = false after a passed course, want true (not barred)")
	}
}

// TestAttemptTrainingCourseRejectedEnrollmentRecordsNoAttempt confirms a
// failed "Check Int to enroll" (p.62-63) costs nothing beyond the roll —
// no TrainingCourse entry is appended, and the character isn't barred.
// Int=0 forces the enroll check to fail unconditionally (2D6 can never
// roll 0 or less), regardless of seed.
func TestAttemptTrainingCourseRejectedEnrollmentRecordsNoAttempt(t *testing.T) {
	t.Parallel()

	var edu Education

	skills, admitted := attemptTrainingCourse(dice.New(rand.NewPCG(1, 1)), eduUPP(0, 250, 0), &edu)

	if admitted {
		t.Error("admitted = true, want false (Int=0 fails the enroll check unconditionally)")
	}

	if skills != nil {
		t.Errorf("skills = %v, want nil", skills)
	}

	if len(edu.TrainingCourses) != 0 {
		t.Errorf("TrainingCourses = %+v, want none recorded for a rejected enrollment", edu.TrainingCourses)
	}
}

// TestAttemptTrainingCourseSuccessGrantsSkillAndSchoolName is the
// unfailable-fixture pass case: Int and Edu both far above any 2D6 roll
// guarantee enrollment and a Pass on the very first attempt (Mod 0).
func TestAttemptTrainingCourseSuccessGrantsSkillAndSchoolName(t *testing.T) {
	t.Parallel()

	var edu Education

	skills, admitted := attemptTrainingCourse(dice.New(rand.NewPCG(1, 1)), eduUPP(250, 250, 0), &edu)

	if !admitted {
		t.Fatal("admitted = false, want true (unfailable fixture)")
	}

	if len(skills) != 1 || skills[0].Level != 2 || skills[0].Kind != Skill {
		t.Errorf("skills = %+v, want one Skill at Level 2 (p.60's own \"Skill-2\")", skills)
	}

	if len(edu.TrainingCourses) != 1 {
		t.Fatalf("TrainingCourses = %+v, want exactly one entry", edu.TrainingCourses)
	}

	course := edu.TrainingCourses[0]
	if !course.Passed || course.Subject != skills[0].Name {
		t.Errorf("TrainingCourses[0] = %+v, want Passed=true and Subject=%q", course, skills[0].Name)
	}

	if course.SchoolNameRoll == 0 || course.SchoolRank == 0 {
		t.Errorf(
			"SchoolNameRoll/SchoolRank = %d/%d, want both nonzero (p.72's own roll happens on any Graduate)",
			course.SchoolNameRoll, course.SchoolRank,
		)
	}
}

// TestAttemptTrainingCourseFailureBarsFurtherAttempts is the
// unfailable-fixture fail case: Edu=0 makes the Tra-in-lieu Pass/Fail
// target 0, below any possible 2D6 roll, so the course always fails once
// enrolled. Confirms the failed attempt is still recorded (naming what
// was attempted, same "names what was applied to" precedent
// attendInstitution's own doc comment establishes) and that it
// permanently bars a further course.
func TestAttemptTrainingCourseFailureBarsFurtherAttempts(t *testing.T) {
	t.Parallel()

	var edu Education

	skills, admitted := attemptTrainingCourse(dice.New(rand.NewPCG(1, 1)), eduUPP(250, 0, 0), &edu)

	if !admitted {
		t.Fatal("admitted = false, want true (enrollment succeeds; only the Pass/Fail Check fails)")
	}

	if skills != nil {
		t.Errorf("skills = %v, want nil (no grant on failure)", skills)
	}

	if len(edu.TrainingCourses) != 1 || edu.TrainingCourses[0].Passed {
		t.Fatalf("TrainingCourses = %+v, want one failed entry", edu.TrainingCourses)
	}

	if edu.TrainingCourses[0].Subject == "" {
		t.Error("Subject = \"\", want the enrolled-in subject recorded even on failure (p.62's own Barr Vech example)")
	}

	if edu.TrainingCourses[0].SchoolNameRoll != 0 {
		t.Errorf(
			"SchoolNameRoll = %d, want 0 (only a passed attempt reaches the chart roll)",
			edu.TrainingCourses[0].SchoolNameRoll,
		)
	}

	if !trainingCourseBarred(edu) {
		t.Error("trainingCourseBarred = false after a failed course, want true")
	}
}

// TestAttemptTrainingCourseModAccumulatesAcrossPriorAttempts confirms
// p.63's own cumulative penalty — "the Pass/Fail Check Tra is subject to
// Mod minus the number of Training courses taken" — is actually applied
// as -len(edu.TrainingCourses), not just checked once and ignored
// thereafter. Both ends are deterministic, not seed-dependent: Edu=24
// gives an unmodified Tra-in-lieu target of 12, which a 2D6 roll can
// never exceed, so with zero prior courses this attempt cannot fail; 11
// prior passed courses (Mod -11) drag the effective target to 1, which a
// 2D6 roll (minimum 2) can never meet, so with them present the identical
// check cannot succeed. Both boundaries are load-bearing: a version of
// this test using a merely-likely-to-fail target instead didn't actually
// distinguish "Mod applied" from "Mod ignored" on this seed.
func TestAttemptTrainingCourseModAccumulatesAcrossPriorAttempts(t *testing.T) {
	t.Parallel()

	prior := make([]TrainingCourse, 11)
	for i := range prior {
		prior[i] = TrainingCourse{Passed: true, Subject: "Wheeled"}
	}

	edu := Education{TrainingCourses: prior}

	skills, admitted := attemptTrainingCourse(dice.New(rand.NewPCG(1, 1)), eduUPP(250, 24, 0), &edu)

	if !admitted {
		t.Fatal("admitted = false, want true (Int=250 enrolls unconditionally)")
	}

	if skills != nil {
		t.Errorf("skills = %v, want nil (Mod -11 against target 12 cannot succeed)", skills)
	}

	if len(edu.TrainingCourses) != 12 || edu.TrainingCourses[11].Passed {
		t.Fatalf("TrainingCourses = %+v, want 12 entries with the last one failed", edu.TrainingCourses)
	}
}
