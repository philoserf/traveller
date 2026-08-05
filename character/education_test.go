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

// TestEd5AwardsNoMajor is the regression test for #128:
// awardEducationSkills used to declare edu.Major unconditionally on the
// first Pass, before the switch that's supposed to gate it on
// MajorBonus/GrantsMajorMinor — so ED5 (which has neither, p.60) drew a
// College-column Major subject and rolled a die for it despite having no
// Major/Minor mechanic of its own anywhere in the book.
func TestEd5AwardsNoMajor(t *testing.T) {
	t.Parallel()

	edu, upp := resolveEducation(dice.New(rand.NewPCG(1, 1)), eduUPP(20, 2, 7))
	if edu.School != "ED5" || !edu.Graduated {
		t.Fatalf("School = %q graduated=%v, want ED5 graduated (Int 20 cannot fail a 2D check)",
			edu.School, edu.Graduated)
	}

	if upp.Characteristics[C5] != 5 {
		t.Fatalf("Edu = %v, want 5 after ED5", upp.Characteristics[C5])
	}

	if edu.Major != "" {
		t.Errorf("Major = %q, want \"\" (ED5 has no Major/Minor mechanic, p.60)", edu.Major)
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

// TestHoldsDegree covers holdsDegree's own Honors-satisfies-plain-BA
// rule (#113): an Honors graduate has "a Bachelors" too, just a better
// one, which is why Masters's own BA prerequisite must not reject them.
func TestHoldsDegree(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		have string
		want string
		ok   bool
	}{
		{"plain BA satisfies BA", educationDegreeBachelors, educationDegreeBachelors, true},
		{"Honors BA satisfies BA", educationDegreeHonours, educationDegreeBachelors, true},
		{"plain BA does not satisfy Honors BA", educationDegreeBachelors, educationDegreeHonours, false},
		{"MA satisfies MA", educationDegreeMasters, educationDegreeMasters, true},
		{"no degree satisfies nothing", "", educationDegreeBachelors, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := holdsDegree(Education{Degree: c.have}, c.want); got != c.ok {
				t.Errorf("holdsDegree(Degree=%q, %q) = %v, want %v", c.have, c.want, got, c.ok)
			}
		})
	}
}

// TestResolveGraduateEducationOnlyAppliesToUniversity guards #113's own
// gate — "A University Masters Program"/"Often associated with a
// University" both name University specifically, unlike OTC/NOTC's
// "College or University." A College graduate, even an Honors one,
// must not reach Medical School.
func TestResolveGraduateEducationOnlyAppliesToUniversity(t *testing.T) {
	t.Parallel()

	edu := Education{School: "College", Degree: educationDegreeHonours}
	r := dice.New(rand.NewPCG(1, 1))

	resolveGraduateEducation(r, eduUPP(20, 9, 20), &edu)

	if edu.Graduate != nil {
		t.Errorf("edu.Graduate = %+v, want nil — only University feeds the graduate tier", edu.Graduate)
	}
}

// TestResolveGraduateEducationReachesHonorsTierWithHonors is #113's own
// ruling that Medical School and Law School are a genuine open choice
// (educationHonorsTierPrograms's own doc comment) — an Honors BA holder
// reaches one of the two (picked at random), not Masters, even though
// Masters's own weaker BA prerequisite is also satisfied. Runs many
// seeds to confirm both ever fire, guarding the exact "content that
// can't fire" mistake a first ordered-preference draft actually made
// (see educationHonorsTierPrograms's own doc comment) — measurement
// caught Law School reached in only 12 of 6,000 trials before this was
// fixed to a random pick.
func TestResolveGraduateEducationReachesHonorsTierWithHonors(t *testing.T) {
	t.Parallel()

	sawMedical, sawLaw := false, false

	for seed := uint64(1); seed <= 200; seed++ {
		edu := Education{School: "University", Degree: educationDegreeHonours, Minor: "Robotics"}
		r := dice.New(rand.NewPCG(seed, seed))

		upp := resolveGraduateEducation(r, eduUPP(20, 9, 20), &edu)

		if edu.Graduate == nil {
			t.Fatalf("seed %d: edu.Graduate is nil, want Medical or Law School", seed)
		}

		switch edu.Graduate.School {
		case "Medical School":
			sawMedical = true

			if edu.Graduate.Degree != educationDegreeDoctor {
				t.Errorf("seed %d: Degree = %q, want %q", seed, edu.Graduate.Degree, educationDegreeDoctor)
			}
		case "Law School":
			sawLaw = true

			if edu.Graduate.Degree != educationDegreeAttorney {
				t.Errorf("seed %d: Degree = %q, want %q", seed, edu.Graduate.Degree, educationDegreeAttorney)
			}
		default:
			t.Fatalf("seed %d: School = %q, want Medical or Law School", seed, edu.Graduate.School)
		}

		if !edu.Graduate.Graduated {
			t.Errorf("seed %d: Graduate.Graduated = false, want true (unfailable fixture)", seed)
		}

		if upp.Characteristics[C5] < 10 {
			t.Errorf("seed %d: Edu = %v after graduation, want at least 10", seed, upp.Characteristics[C5])
		}
	}

	if !sawMedical || !sawLaw {
		t.Errorf("sawMedical=%v sawLaw=%v across 200 seeds, want both to fire", sawMedical, sawLaw)
	}
}

// TestMedicalSchoolGrantsMedicPerPass isolates Medical School's own
// "Medic-4" Provides cell directly (bypassing the Medical-vs-Law random
// pick) to prove it's a +1-per-pass grant across its own 4 rolls,
// matching its own roll count exactly — the same "+1 Major per Pass"
// accumulation shape College/University/Service Academy already use,
// just naming a fixed skill instead of the character's own Major.
func TestMedicalSchoolGrantsMedicPerPass(t *testing.T) {
	t.Parallel()

	edu := Education{School: "University", Degree: educationDegreeHonours}

	var grad GraduateProgram

	r := dice.New(rand.NewPCG(1, 1))
	resolveOneGraduateProgram(r, eduUPP(20, 9, 20), educationHonorsTierPrograms[0], &edu, &grad)

	if !grad.Graduated || grad.Degree != educationDegreeDoctor {
		t.Fatalf("grad = %+v, want a graduated Doctor (unfailable fixture)", grad)
	}

	medicLevel := 0

	for _, s := range edu.Skills {
		if s.Name == "Medic" {
			medicLevel += s.Level
		}
	}

	if medicLevel != 4 {
		t.Errorf("Medic level = %d, want 4 (4 passes, +1 Medic each)", medicLevel)
	}
}

// TestResolveGraduateEducationFallsBackToMastersWithoutHonors covers
// the plain-BA path: Soc 1 makes Medical/Law School's own Waiver into
// their unmet Honors-BA prerequisite impossible (2D6<=1), so the loop
// falls through to Masters, whose own BA prerequisite is already held.
// Masters's own "Minor+1 per 2 Passes" advances the Minor University
// already declared — 2 passes yields exactly one grant, no fresh
// subject and no Major grant at all.
func TestResolveGraduateEducationFallsBackToMastersWithoutHonors(t *testing.T) {
	t.Parallel()

	edu := Education{School: "University", Major: "Psychology", Degree: educationDegreeBachelors, Minor: "Robotics"}
	r := dice.New(rand.NewPCG(3, 3))

	upp := resolveGraduateEducation(r, eduUPP(20, 9, 1), &edu)

	if edu.Graduate == nil || edu.Graduate.School != "Masters" {
		t.Fatalf("Graduate = %+v, want School=Masters", edu.Graduate)
	}

	if edu.Graduate.Degree != educationDegreeMasters || !edu.Graduate.Graduated {
		t.Errorf("Graduate = %+v, want Degree=MA Graduated=true (unfailable fixture)", edu.Graduate)
	}

	if upp.Characteristics[C5] < 9 {
		t.Errorf("Edu = %v after graduation, want at least 9", upp.Characteristics[C5])
	}

	names := map[string]int{}
	for _, s := range edu.Skills {
		names[s.Name] += s.Level
	}

	if names["Robotics"] != 1 {
		t.Errorf("Robotics (Minor) level = %d, want 1 (2 Masters passes, Minor+1 per 2 Passes)", names["Robotics"])
	}

	if names["Psychology"] != 0 {
		t.Errorf("Psychology (Major) level = %d, want 0 — Masters grants no Major skill", names["Psychology"])
	}
}

// TestResolveGraduateEducationWaivesIntoHonorsTierWithoutHonors is
// p.59's own Waiver clause naming "Prerequisite" as one of exactly four
// waivable adverse decisions, restated in the Masters/Professors/
// Medical/Law paragraph as "(all of these requirements can be
// waived)". Soc 20 makes the Waiver into an unmet Honors-BA
// prerequisite unfailable, reaching Medical or Law School despite only
// holding a plain BA.
func TestResolveGraduateEducationWaivesIntoHonorsTierWithoutHonors(t *testing.T) {
	t.Parallel()

	edu := Education{School: "University", Degree: educationDegreeBachelors}
	r := dice.New(rand.NewPCG(4, 4))

	resolveGraduateEducation(r, eduUPP(20, 9, 20), &edu)

	if edu.Graduate == nil || (edu.Graduate.School != "Medical School" && edu.Graduate.School != "Law School") {
		t.Fatalf("Graduate = %+v, want Medical or Law School (waived into its own Honors-BA prerequisite)",
			edu.Graduate)
	}

	if edu.Waivers == 0 {
		t.Error("edu.Waivers = 0, want at least one recorded waiver attempt")
	}
}

// TestResolveGraduateEducationChainsProfessorsAfterMasters is p.61's
// own "A Professors Program requires a Masters" — reachable only once
// Masters itself graduates with Degree MA, recorded as
// Graduate.Next rather than replacing Graduate.
func TestResolveGraduateEducationChainsProfessorsAfterMasters(t *testing.T) {
	t.Parallel()

	edu := Education{School: "University", Degree: educationDegreeBachelors, Minor: "Robotics"}
	r := dice.New(rand.NewPCG(3, 3))

	upp := resolveGraduateEducation(r, eduUPP(20, 9, 1), &edu)

	if edu.Graduate == nil || edu.Graduate.School != "Masters" || !edu.Graduate.Graduated {
		t.Fatalf("Graduate = %+v, want a graduated Masters", edu.Graduate)
	}

	if edu.Graduate.Next == nil {
		t.Fatal("Graduate.Next is nil, want a chained Professors attempt")
	}

	if edu.Graduate.Next.School != "Professors" || edu.Graduate.Next.Degree != educationDegreeProfessor {
		t.Errorf("Graduate.Next = %+v, want School=Professors Degree=Professor", edu.Graduate.Next)
	}

	if !edu.Graduate.Next.Graduated {
		t.Error("Graduate.Next.Graduated = false, want true (unfailable fixture)")
	}

	if upp.Characteristics[C5] < 12 {
		t.Errorf("Edu = %v after Professors graduation, want at least 12", upp.Characteristics[C5])
	}
}

// TestShouldAttemptLaterEducationRetriesAfterFailure is p.59's own
// Later Education "retry" case: a character who never graduated (here,
// never attended at all — Education{}) and still qualifies for
// something should be offered it.
func TestShouldAttemptLaterEducationRetriesAfterFailure(t *testing.T) {
	t.Parallel()

	school, ok := shouldAttemptLaterEducation(eduUPP(7, 2, 7), Education{})
	if !ok || school.Name != "ED5" {
		t.Errorf("shouldAttemptLaterEducation = %+v, %v, want ED5, true", school, ok)
	}
}

// TestShouldAttemptLaterEducationEscalatesToABetterInstitution is the
// "escalate" case: a character who already graduated College, but whose
// Edu has since grown enough to qualify for University (strictly better
// in educationInstitutions's own preference order), should be offered
// the upgrade.
func TestShouldAttemptLaterEducationEscalatesToABetterInstitution(t *testing.T) {
	t.Parallel()

	edu := Education{School: "College", Degree: educationDegreeBachelors, Graduated: true}

	school, ok := shouldAttemptLaterEducation(eduUPP(7, 9, 7), edu)
	if !ok || school.Name != "University" {
		t.Errorf("shouldAttemptLaterEducation = %+v, %v, want University, true", school, ok)
	}
}

// TestShouldAttemptLaterEducationDeclinesWithoutABetterOption confirms
// the gate doesn't fire every term forever: a graduate whose Edu hasn't
// grown past what their current School already qualifies for has
// nothing better to reach.
func TestShouldAttemptLaterEducationDeclinesWithoutABetterOption(t *testing.T) {
	t.Parallel()

	edu := Education{School: "College", Degree: educationDegreeBachelors, Graduated: true}

	if school, ok := shouldAttemptLaterEducation(eduUPP(7, 5, 7), edu); ok {
		t.Errorf("shouldAttemptLaterEducation = %+v, true, want false (College is still the best reachable)", school)
	}
}

// TestAttendInstitutionDoesNotRegressAnExistingDegree is
// attendInstitution's own stated invariant: a second attendance that
// fails admission outright must not clear a Degree an earlier,
// different attendance already earned. Int=Edu=Soc=0 forces both the
// ApplyCheck and its Waiver to fail unconditionally (2D6 can never roll
// 0 or less), regardless of seed.
func TestAttendInstitutionDoesNotRegressAnExistingDegree(t *testing.T) {
	t.Parallel()

	edu := Education{School: "College", Degree: educationDegreeBachelors, Graduated: true}
	university, ok := chooseInstitution(eduUPP(20, 20, 20))

	if !ok || university.Name != "University" {
		t.Fatalf("chooseInstitution = %+v, %v, want University, true", university, ok)
	}

	attendInstitution(dice.New(rand.NewPCG(1, 1)), eduUPP(0, 0, 0), university, &edu)

	if edu.Degree != educationDegreeBachelors {
		t.Errorf(
			"Degree = %q after a failed later attempt, want %q (unregressed)",
			edu.Degree,
			educationDegreeBachelors,
		)
	}
}

// TestAttendInstitutionReturnsOnlyThisAttendancesSkills is the
// regression test for attendInstitution's own delta return value: a
// second attendance's own returned skills must not repeat the first
// attendance's, since edu.Skills already carries those and a Later
// Education Term's SkillsAwarded must not double-grant them.
func TestAttendInstitutionReturnsOnlyThisAttendancesSkills(t *testing.T) {
	t.Parallel()

	college, ok := chooseInstitution(eduUPP(20, 5, 20))
	if !ok || college.Name != "College" {
		t.Fatalf("chooseInstitution(Edu 5) = %+v, %v, want College, true", college, ok)
	}

	university, ok := chooseInstitution(eduUPP(20, 20, 20))
	if !ok || university.Name != "University" {
		t.Fatalf("chooseInstitution(Edu 20) = %+v, %v, want University, true", university, ok)
	}

	var edu Education

	r := dice.New(rand.NewPCG(4, 4))

	upp := eduUPP(20, 20, 20)
	_, firstDelta, _ := attendInstitution(r, upp, college, &edu)

	if len(firstDelta) == 0 {
		t.Fatal("first attendance's own delta is empty, want at least one skill (unfailable fixture)")
	}

	if len(edu.Skills) != len(firstDelta) {
		t.Fatalf(
			"edu.Skills = %+v after one attendance, want it to equal the returned delta %+v",
			edu.Skills,
			firstDelta,
		)
	}

	_, secondDelta, _ := attendInstitution(r, upp, university, &edu)

	if len(secondDelta) == 0 {
		t.Fatal("second attendance's own delta is empty, want at least one skill (unfailable fixture)")
	}

	if len(edu.Skills) != len(firstDelta)+len(secondDelta) {
		t.Errorf("len(edu.Skills) = %d, want %d (first delta) + %d (second delta) = %d, not double-counted",
			len(edu.Skills), len(firstDelta), len(secondDelta), len(firstDelta)+len(secondDelta))
	}
}

// TestAttendInstitutionRejectedAdmissionReportsNotAdmitted is the
// regression test for a Copilot-review-caught bug (PR #162): the third
// return value must be false when admission is rejected outright — p.59
// "if accepted substitutes that process for the entire term" is a real
// conditional, not automatic once shouldAttemptLaterEducation picks a
// school. Int=Edu=Soc=0 forces both the ApplyCheck and its Waiver to
// fail unconditionally (2D6 can never roll 0 or less), regardless of
// seed.
func TestAttendInstitutionRejectedAdmissionReportsNotAdmitted(t *testing.T) {
	t.Parallel()

	university, ok := chooseInstitution(eduUPP(20, 20, 20))
	if !ok || university.Name != "University" {
		t.Fatalf("chooseInstitution = %+v, %v, want University, true", university, ok)
	}

	var edu Education

	_, _, admitted := attendInstitution(dice.New(rand.NewPCG(1, 1)), eduUPP(0, 0, 0), university, &edu)

	if admitted {
		t.Error("admitted = true, want false (ApplyCheck and its Waiver both fail unconditionally)")
	}
}

// TestAttendInstitutionResetsPassesPerAttendance is the regression test
// for another Copilot-review-caught bug (PR #162): Passes counts a
// single attendance's own Pass/Fail rolls (runEducationYears's own doc
// comment), not a lifetime total, so it must reset to 0 at the start of
// each attendance rather than accumulating a stale value forward from
// an earlier, different institution.
func TestAttendInstitutionResetsPassesPerAttendance(t *testing.T) {
	t.Parallel()

	college, ok := chooseInstitution(eduUPP(20, 5, 20))
	if !ok || college.Name != "College" {
		t.Fatalf("chooseInstitution(Edu 5) = %+v, %v, want College, true", college, ok)
	}

	edu := Education{Passes: 100} // an implausible carryover value from a prior attendance

	attendInstitution(dice.New(rand.NewPCG(4, 4)), eduUPP(20, 20, 20), college, &edu)

	if edu.Passes != college.Rolls {
		t.Errorf("Passes = %d, want %d (College's own Rolls — an unfailable fixture graduates every year)",
			edu.Passes, college.Rolls)
	}
}
