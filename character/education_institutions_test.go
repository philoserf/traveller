package character

import (
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/philoserf/traveller/dice"
)

// TestInstitutionChartMatchesThePrintedPage pins the p.92-93
// transcription. The page prints its twelve sub-tables two abreast, so a
// visual read splices the left table's entries into the right one's;
// these were recovered from the PDF's word bounding boxes instead.
//
// The rank dice are the values most easily got wrong, since they vary
// per institution and one of them is not a die at all.
func TestInstitutionChartMatchesThePrintedPage(t *testing.T) {
	t.Parallel()

	wantRankDice := map[string]int{
		"ED5":              0, // "Rank= Inconsequential"
		"Trade School":     1,
		"College":          2,
		"University":       3, // the only 3D on the chart
		"Medical School":   2,
		"Law School":       2,
		"Naval Academy":    2,
		"Military Academy": 2,
		"Flight School":    2,
		"Command College":  2,
		"Training Course":  1,
		"ANM School":       1,
	}

	if len(educationInstitutionCharts) != len(wantRankDice) {
		t.Fatalf("chart has %d institutions, want the printed %d",
			len(educationInstitutionCharts), len(wantRankDice))
	}

	for name, want := range wantRankDice {
		chart, ok := educationInstitutionCharts[name]
		if !ok {
			t.Errorf("%s is missing from the chart", name)

			continue
		}

		if chart.RankDice != want {
			t.Errorf("%s rank = %dD, want %dD", name, chart.RankDice, want)
		}

		for i, entry := range chart.Names {
			if entry == "" {
				t.Errorf("%s entry %d is empty", name, i+1)
			}
		}
	}
}

// TestInstitutionNamesResolveOnlyWhenComplete is the rule that keeps
// half-substituted names off a character sheet: <Skill> and <Armed
// Force> have sources in this codebase, and the five place-name
// placeholders do not.
func TestInstitutionNamesResolveOnlyWhenComplete(t *testing.T) {
	t.Parallel()

	cases := []struct {
		template   string
		skill      string
		armedForce string
		want       string
	}{
		{"Imperial War College", "", "", "Imperial War College"},
		{"Imperial Command College (<Armed Force>)", "", "Navy", "Imperial Command College (Navy)"},
		{"Institute of <Skill>", "Robotics", "", "Institute of Robotics"},
		{"Imperial <Armed Force> <Skill> Course", "Recon", "Marine", "Imperial Marine Recon Course"},
		// No source for any of these, so nothing is emitted.
		{"College of <World>", "Robotics", "Navy", ""},
		{"<City> College", "Robotics", "Navy", ""},
		{"The <Color> Institute", "", "", ""},
		{"<Company> School of <Skill>", "Robotics", "", ""},
	}

	for _, c := range cases {
		got, ok := resolveInstitutionName(c.template, c.skill, c.armedForce)
		if c.want == "" {
			if ok || got != "" {
				t.Errorf("%q resolved to %q, want it withheld as incomplete", c.template, got)
			}

			continue
		}

		if !ok || got != c.want {
			t.Errorf("%q resolved to %q, want %q", c.template, got, c.want)
		}
	}
}

// TestNoResolvedNameKeepsAPlaceholder is the property the case list
// above is a sample of: whatever the chart holds, nothing with an
// unfilled placeholder ever escapes.
func TestNoResolvedNameKeepsAPlaceholder(t *testing.T) {
	t.Parallel()

	for school, chart := range educationInstitutionCharts {
		for i, entry := range chart.Names {
			got, ok := resolveInstitutionName(entry, "Robotics", "Navy")
			if ok && strings.ContainsRune(got, '<') {
				t.Errorf("%s entry %d resolved to %q, which still holds a placeholder", school, i+1, got)
			}

			if !ok && got != "" {
				t.Errorf("%s entry %d withheld but returned %q", school, i+1, got)
			}
		}
	}
}

// TestSchoolRankIsRolledForEveryAttendanceButED5 covers p.72's own "For
// Each School Attended: Note School Name and Rank", and ED5's own
// exception — its chart entry prints "Rank= Inconsequential" rather than
// a die, so it must draw none.
func TestSchoolRankIsRolledForEveryAttendanceButED5(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))

	roll, _, rank := rollInstitution(r, "ED5", "")
	if roll < 1 || roll > 6 {
		t.Errorf("ED5 name roll = %d, want 1D", roll)
	}

	if rank != 0 {
		t.Errorf("ED5 rank = %d, want none — the chart prints Inconsequential", rank)
	}

	// University rolls 3D, so it cannot come in below 3.
	_, _, rank = rollInstitution(r, "University", "")
	if rank < 3 || rank > 18 {
		t.Errorf("University rank = %d, want 3D", rank)
	}

	// An institution absent from the chart draws nothing at all.
	spent := dice.New(rand.NewPCG(2, 2))
	if roll, name, rank := rollInstitution(spent, "Nowhere Polytechnic", ""); roll != 0 || name != "" || rank != 0 {
		t.Errorf("an unknown institution produced %d/%q/%d, want nothing", roll, name, rank)
	}

	untouched := dice.New(rand.NewPCG(2, 2))
	if spent.TwoD6() != untouched.TwoD6() {
		t.Error("an unknown institution consumed dice")
	}
}

// TestGeneratedCharactersRecordSchoolRank checks the wiring: step C
// attends an institution for nearly every character, so School Rank
// should reach the finished Education.
func TestGeneratedCharactersRecordSchoolRank(t *testing.T) {
	t.Parallel()

	graduates, ranked, named := 0, 0, 0

	for seed := uint64(1); seed <= 500; seed++ {
		c, ok := GenerateScoutCharacter(dice.New(rand.NewPCG(seed, seed)))
		if !ok || !c.Education.Graduated {
			continue
		}

		graduates++

		if c.Education.SchoolNameRoll < 1 || c.Education.SchoolNameRoll > 6 {
			t.Fatalf("seed %d: SchoolNameRoll = %d, want 1D", seed, c.Education.SchoolNameRoll)
		}

		if c.Education.SchoolRank > 0 {
			ranked++
		}

		if c.Education.SchoolName != "" {
			named++
		}
	}

	if graduates == 0 {
		t.Fatal("no graduates generated")
	}

	if ranked == 0 {
		t.Error("no graduate carries a School Rank")
	}

	// Names are expected to be mostly absent: College and University
	// entries all name a place, and this codebase has no place names.
	// Recorded as a measurement rather than asserted, since the ratio
	// will move when those acquire sources.
	t.Logf("%d graduates: %d carry a School Rank, %d a School Name", graduates, ranked, named)
}
