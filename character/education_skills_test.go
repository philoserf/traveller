package character

import "testing"

// TestSkillMatrixMatchesThePrintedPage pins the p.60 transcription the
// way TestNobleRanksMatchThePrintedTable pins p.88 — because the same
// hazard applies and worse. The page is four side-by-side skill groups
// whose flag columns interleave in the text extract, and the values were
// recovered from the PDF's word bounding boxes rather than read off the
// page, so the failure mode to guard against is a whole column being
// read one position out.
//
// Per-column totals catch exactly that: shifting any column by one
// changes at least two of these counts, and no plausible single
// mis-paired row changes any of them by more than one.
func TestSkillMatrixMatchesThePrintedPage(t *testing.T) {
	t.Parallel()

	if len(educationSkills) != 121 {
		t.Fatalf("educationSkills has %d rows, want 121", len(educationSkills))
	}

	counts := map[string]struct {
		school schoolSet
		want   int
	}{
		"C College":          {schoolCollege, 40},
		"L Law":              {schoolLaw, 2},
		"M Medical":          {schoolMedical, 2},
		"A Military/Army":    {schoolMilitaryAcademy, 33},
		"N Naval":            {schoolNavalAcademy, 34},
		"M Marine":           {schoolMarine, 36},
		"S Apprentice/Trade": {schoolTrade, 67},
	}

	for name, c := range counts {
		got := 0

		for _, skill := range educationSkills {
			if skill.Schools&c.school != 0 {
				got++
			}
		}

		if got != c.want {
			t.Errorf("column %s carries %d skills, want %d", name, got, c.want)
		}
	}
}

// TestSkillMatrixSpotRows checks the rows most likely to be mis-paired:
// the multi-flag ones, where a name and its flags sit on different
// baselines and a proximity match could attach a flag to its neighbour.
// A row carrying every column at once is the strongest single check
// available, since it can only be right if all five columns line up.
func TestSkillMatrixSpotRows(t *testing.T) {
	t.Parallel()

	want := map[string]schoolSet{
		// The one row flagged in every column on the page.
		"Robotics": schoolCollege | schoolMilitaryAcademy | schoolNavalAcademy | schoolMarine | schoolTrade,
		// A left-group row using all four of its columns.
		"Language": schoolCollege | schoolNavalAcademy | schoolMarine | schoolTrade,
		// The row that proves the two M codes are distinct: Medic is
		// taught at Medical School (left column) and Marine School
		// (third), which is only expressible if the columns are read
		// separately rather than folded into one "M".
		"Medic": schoolMedical | schoolNavalAcademy | schoolMarine | schoolTrade,
		// The only other Medical M, and one of only two L rows.
		"Forensics": schoolMedical,
		"Advocate":  schoolLaw,
		// A right-group row: College plus Army plus Navy plus Trade,
		// with the Marine column deliberately empty.
		"Aeronautics": schoolCollege | schoolMilitaryAcademy | schoolNavalAcademy | schoolTrade,
		// Sciences are College-only, which is what makes the C-flagged
		// block an enumerable science list.
		"Psychohistory": schoolCollege,
		// A skill no institution teaches at all — the page leaves these
		// blank on purpose, and a transcription that gave every row at
		// least one flag would be wrong.
		"Gambler": 0,
	}

	got := make(map[string]schoolSet, len(educationSkills))
	for _, skill := range educationSkills {
		if _, dup := got[skill.Name]; dup && skill.Name != "Grav" {
			t.Errorf("%q appears twice in the matrix", skill.Name)
		}

		got[skill.Name] = skill.Schools
	}

	for name, wantSet := range want {
		if got[name] != wantSet {
			t.Errorf("%s = %05b, want %05b", name, got[name], wantSet)
		}
	}
}

// TestGravIsTheOnlyRepeatedSkillName guards skillsForSchool's
// deduplication. Three separate p.60 rows are printed "Grav" — under
// Driver, Flyer and Seafarer — and if another duplicate ever appeared
// without the same treatment it would quietly bias every draw.
func TestGravIsTheOnlyRepeatedSkillName(t *testing.T) {
	t.Parallel()

	seen := make(map[string]int, len(educationSkills))
	for _, skill := range educationSkills {
		seen[skill.Name]++
	}

	for name, n := range seen {
		if n > 1 && name != "Grav" {
			t.Errorf("%q is printed %d times; only Grav should repeat", name, n)
		}
	}

	if seen["Grav"] != 3 {
		t.Errorf("Grav appears %d times, want 3 (Driver, Flyer, Seafarer)", seen["Grav"])
	}

	// The three Grav rows must agree, or collapsing them would change
	// which schools teach it.
	pool := skillsForSchool(schoolNavalAcademy)
	gravs := 0

	for _, name := range pool {
		if name == "Grav" {
			gravs++
		}
	}

	if gravs != 1 {
		t.Errorf("Naval pool contains Grav %d times, want exactly 1", gravs)
	}
}
