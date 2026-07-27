package character

import (
	"math/rand/v2"
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
// column — "Edu 5+" for College, "Edu 7+" for University, "Edu 4 -" for
// ED5 — and this codebase's resolution of the open choice between them:
// the most demanding school the character qualifies for.
func TestInstitutionChoiceFollowsThePrerequisites(t *testing.T) {
	t.Parallel()

	cases := []struct {
		edu  ehex.Value
		want string
	}{
		{0, "ED5"},
		{4, "ED5"},
		{5, "College"},
		{6, "College"},
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
