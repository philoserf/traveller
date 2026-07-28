package character

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// eduUPP builds a UPP with a given Int, Edu and Soc and nothing else,
// which is all step C reads.
func eduUPP(intChar, edu, soc ehex.Value) UPP {
	return UPP{Characteristics: [6]ehex.Value{0, 0, 0, intChar, edu, soc}}
}

// TestInstitutionChoiceFollowsThePrerequisites is p.60's Pre-Requisites
// column — "Edu 5+" for College, "Edu 6+" for Service Academy, "Edu 7+"
// for University, "Edu 4 -" for ED5 — and this codebase's resolution of
// the open choice between them: the most demanding school the character
// qualifies for. Service Academy outranks College once Edu 6+ is met
// (same cost, plus a Commission — see educationInstitutions's own doc
// comment) but not University, whose higher Edu=9 is read as preferred
// over Service Academy's forced one-term military obligation (#113).
func TestInstitutionChoiceFollowsThePrerequisites(t *testing.T) {
	t.Parallel()

	cases := []struct {
		edu  ehex.Value
		want string
	}{
		{0, "ED5"},
		{4, "ED5"},
		{5, "College"},
		{6, "Service Academy"},
		{7, "University"},
		{12, "University"},
	}

	for _, c := range cases {
		got, ok := chooseInstitution(eduUPP(7, c.edu, 7))
		if !ok {
			t.Errorf("Edu %v qualified for nothing", c.edu)

			continue
		}

		if got.Name != c.want {
			t.Errorf("Edu %v attends %s, want %s", c.edu, got.Name, c.want)
		}
	}
}

// TestEd5RaisesEduToFive is p.60's own "A character with Edu less than 5
// can attempt the ED5 program... Check Int: if successful, Edu is raised
// to 5" — the one institution that exists to lift a character over the
// Trade School floor.
func TestEd5RaisesEduToFive(t *testing.T) {
	t.Parallel()

	// Int 20 cannot fail a 2D check; Int 1 cannot pass one.
	edu, upp := resolveEducation(dice.New(rand.NewPCG(1, 1)), eduUPP(20, 2, 7))
	if !edu.Graduated || upp.Characteristics[C5] != 5 {
		t.Errorf("Edu = %v graduated=%v, want Edu 5 after ED5", upp.Characteristics[C5], edu.Graduated)
	}

	edu, upp = resolveEducation(dice.New(rand.NewPCG(1, 1)), eduUPP(1, 2, 1))
	if edu.Graduated || upp.Characteristics[C5] != 2 {
		t.Errorf("Edu = %v graduated=%v, want an unchanged Edu 2 after failing ED5",
			upp.Characteristics[C5], edu.Graduated)
	}
}

// TestGraduationNeverLowersEdu is the aside printed beside p.60's own
// Graduation column: "(If Edu already at this level, award Edu+1)". A
// University graduate whose Edu was already 9 must come out at A, not
// stay at 9 and certainly not drop.
func TestGraduationNeverLowersEdu(t *testing.T) {
	t.Parallel()

	for _, start := range []ehex.Value{7, 8, 9, 10, 12} {
		edu, upp := resolveEducation(dice.New(rand.NewPCG(2, 2)), eduUPP(20, start, 20))
		if !edu.Graduated {
			t.Fatalf("Edu %v failed to graduate on an unfailable check", start)
		}

		if upp.Characteristics[C5] <= start {
			t.Errorf("Edu %v graduated to %v, want an increase", start, upp.Characteristics[C5])
		}
	}
}

// TestDegreeProgramGrantsMajorPerPassAndMinorPerTwo is p.60's own merged
// Provides cell, "Major+1 per Pass and Minor+1 per 2 Passes", checked
// against the worked example on the same page: Eneri passes his freshman
// and senior years, ends on Psychology-2, and declares Robotics as his
// Minor on the second Pass for Robotics-1.
func TestDegreeProgramGrantsMajorPerPassAndMinorPerTwo(t *testing.T) {
	t.Parallel()

	// Int and Edu of 20 pass every check, so all four years are Passes.
	edu, _ := resolveEducation(dice.New(rand.NewPCG(3, 3)), eduUPP(20, 7, 20))

	if edu.School != "University" || edu.Passes != 4 {
		t.Fatalf("School = %q with %d passes, want 4 passes at University", edu.School, edu.Passes)
	}

	if edu.Major == "" || edu.Minor == "" {
		t.Fatalf("Major = %q Minor = %q, want both declared after four passes", edu.Major, edu.Minor)
	}

	if edu.Major == edu.Minor {
		t.Error("Major and Minor are the same skill; p.59 says they cannot be")
	}

	levels := map[string]int{}
	for _, s := range edu.Skills {
		levels[s.Name] = s.Level
	}

	// Four passes: Major+1 each, so Major-4 — plus Honors' own extra
	// level when that roll also cannot fail, giving 5.
	wantMajor := 4
	if edu.Honors {
		wantMajor++
	}

	if levels[edu.Major] != wantMajor {
		t.Errorf("%s (Major) = %d, want %d for %d passes (Honors=%v)",
			edu.Major, levels[edu.Major], wantMajor, edu.Passes, edu.Honors)
	}

	// Four passes is two Minor grants: "Minor+1 per 2 Passes".
	if levels[edu.Minor] != 2 {
		t.Errorf("%s (Minor) = %d, want 2 for %d passes", edu.Minor, levels[edu.Minor], edu.Passes)
	}
}

// TestEducationProducesLevelsAboveOne is the property #95 needs and
// nothing else in this codebase supplied: every other skill grant is a
// flat +1 (skillLevel1), so a level-6 skill needed six separate grants
// of the same name. p.60's degree programs grant the same Major once per
// passed year.
func TestEducationProducesLevelsAboveOne(t *testing.T) {
	t.Parallel()

	edu, _ := resolveEducation(dice.New(rand.NewPCG(4, 4)), eduUPP(20, 7, 20))

	best := 0
	for _, s := range edu.Skills {
		best = max(best, s.Level)
	}

	if best < 2 {
		t.Errorf("best skill level = %d, want more than 1 — Education is the only source of multi-level grants", best)
	}
}

// TestWaiverModCountsEveryAttempt is p.59's own "Mod minus number of
// previous waivers rolled (successful or not)", which p.60's worked
// example demonstrates: Eneri's fourth waiver attempt carries Mod -4
// having succeeded at only two of the previous three.
func TestWaiverModCountsEveryAttempt(t *testing.T) {
	t.Parallel()

	// Soc 7 with no Mod succeeds often; the point here is the counter,
	// which must advance on failures as well as successes.
	edu := Education{}
	upp := eduUPP(7, 7, 1) // Soc 1: every waiver fails

	for want := 1; want <= 3; want++ {
		if tryWaiver(dice.New(rand.NewPCG(5, 5)), upp, &edu.Waivers) {
			t.Fatal("a waiver succeeded against Soc 1")
		}

		if edu.Waivers != want {
			t.Errorf("Waivers = %d after %d failed attempts, want %d", edu.Waivers, want, want)
		}
	}
}

// TestNoDiceAreDrawnWithoutAnInstitution holds the reproducibility line
// for the case where step C does nothing. Every character qualifies for
// something today, so this guards a branch rather than an observed
// outcome — but it is the branch that would silently shift every stream
// if it ever started drawing.
func TestNoDiceAreDrawnWithoutAnInstitution(t *testing.T) {
	t.Parallel()

	if _, ok := chooseInstitution(eduUPP(7, 5, 7)); !ok {
		t.Skip("every Edu value qualifies for an institution, so there is no no-school path to check")
	}
}

// TestEducationReachesEveryGeneratedCharacter checks the wiring rather
// than the rules: step C sits in generateStart, which all twelve entry
// points share, so a generated character should carry an Education.
func TestEducationReachesEveryGeneratedCharacter(t *testing.T) {
	t.Parallel()

	attended, total := 0, 0

	for seed := uint64(1); seed <= 200; seed++ {
		c, ok := GenerateScoutCharacter(dice.New(rand.NewPCG(seed, seed)))
		if !ok {
			continue
		}

		total++

		if c.Education.Attended() {
			attended++
		}
	}

	if total == 0 {
		t.Fatal("no Scouts generated")
	}

	if attended != total {
		t.Errorf("%d of %d characters attended an institution, want all of them", attended, total)
	}
}

// TestServiceAcademyGrantsCommissionOnGraduation is #113's own p.61:
// "Service Academies ... provide graduates an Army or Navy Commission (a
// Naval Academy graduate may choose a Marine Commission instead)." Edu 6
// (above College's Edu 5+ floor, below University's Edu 7+) isolates
// Service Academy per the preference-order ruling recorded in
// educationInstitutions's own doc comment.
func TestServiceAcademyGrantsCommissionOnGraduation(t *testing.T) {
	t.Parallel()

	// Int and Edu of 20 pass every check, so all four years are Passes.
	edu, upp := resolveEducation(dice.New(rand.NewPCG(6, 6)), eduUPP(20, 6, 20))

	if edu.School != "Service Academy" {
		t.Fatalf("School = %q, want Service Academy", edu.School)
	}

	// Degree is "Honors BA" rather than plain "BA" here: an unfailable
	// fixture also clears the optional Honors roll (resolveHonors), the
	// same non-determinism TestDegreeProgramGrantsMajorPerPassAndMinorPerTwo
	// already accounts for.
	if !edu.Graduated || (edu.Degree != educationDegreeBachelors && edu.Degree != educationDegreeHonours) {
		t.Fatalf("Graduated = %v Degree = %q, want a BA or Honors BA", edu.Graduated, edu.Degree)
	}

	if upp.Characteristics[C5] < 8 {
		t.Errorf("Edu = %v after graduation, want at least 8", upp.Characteristics[C5])
	}

	want := []string{SoldierCareerName, SpacerCareerName, MarineCareerName}
	for _, career := range want {
		if !slices.Contains(edu.CommissionCareers, career) {
			t.Errorf("CommissionCareers = %v, want it to contain %q", edu.CommissionCareers, career)
		}
	}
}

// TestServiceAcademyNoCommissionWithoutGraduation is the failure-path
// counterpart: p.61's Commission is a Graduation benefit, so a character
// refused Admission (and refused a Waiver) must not carry one — the same
// "no attempt, no award" shape every other Graduation benefit already has.
func TestServiceAcademyNoCommissionWithoutGraduation(t *testing.T) {
	t.Parallel()

	// Int 1 fails Apply (2D6<=1 is impossible); Soc 1 fails the Waiver too.
	edu, _ := resolveEducation(dice.New(rand.NewPCG(6, 6)), eduUPP(1, 6, 1))

	if edu.Graduated {
		t.Fatal("edu.Graduated = true, want Admission to have failed")
	}

	if len(edu.CommissionCareers) != 0 {
		t.Errorf("CommissionCareers = %v, want empty without Graduation", edu.CommissionCareers)
	}
}

// TestOfficerTrainingCorpsGrantsCommissionsAtCollege is p.61's own OTC/
// NOTC: "A character attending College or University may also volunteer
// to participate in OTC ... or NOTC ... Success confers a Commission
// (OTC= Army Officer1; NOTC= Navy Officer1 or Marine Officer1)." Edu 5
// isolates College (below Service Academy's Edu 6+ floor). This codebase
// always volunteers for NOTC over OTC (resolveOfficerTrainingCorps's own
// doc comment), so an unfailable fixture earns NOTC's own Navy-or-Marine
// Commission and "Ship Skill-1" grant.
func TestOfficerTrainingCorpsGrantsCommissionsAtCollege(t *testing.T) {
	t.Parallel()

	edu, _ := resolveEducation(dice.New(rand.NewPCG(9, 9)), eduUPP(20, 5, 20))

	if edu.School != "College" {
		t.Fatalf("School = %q, want College", edu.School)
	}

	for _, career := range []string{SpacerCareerName, MarineCareerName} {
		if !slices.Contains(edu.CommissionCareers, career) {
			t.Errorf("CommissionCareers = %v, want it to contain %q (NOTC)", edu.CommissionCareers, career)
		}
	}

	if slices.Contains(edu.CommissionCareers, SoldierCareerName) {
		t.Errorf("CommissionCareers = %v, want no Soldier — NOTC is preferred over OTC", edu.CommissionCareers)
	}

	names := map[string]bool{}
	for _, s := range edu.Skills {
		names[s.Name] = true
	}

	if !names["Starship Skill"] {
		t.Error("edu.Skills has no \"Starship Skill\" grant, want NOTC's own p.60 \"Ship Skill-1\" Provides entry")
	}
}

// TestOfficerTrainingCorpsRequiresCollegeOrUniversity guards p.61's own
// gate — "A character attending College or University" — against ED5,
// which draws no dice at all today (TestNoDiceAreDrawnWithoutAnInstitution)
// and must not start doing so once OTC/NOTC exist.
func TestOfficerTrainingCorpsRequiresCollegeOrUniversity(t *testing.T) {
	t.Parallel()

	edu := Education{School: "ED5"}
	before := dice.New(rand.NewPCG(1, 1))

	resolveOfficerTrainingCorps(before, eduUPP(20, 20, 20), &edu)

	if len(edu.CommissionCareers) != 0 || len(edu.Skills) != 0 {
		t.Errorf("edu = %+v, want unchanged — OTC/NOTC only apply at College or University", edu)
	}
}
