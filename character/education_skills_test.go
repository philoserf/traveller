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
	pool := skillsForSchool(schoolNavalAcademy, false)
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

// TestKnowledgeGroupsMatchPage61 pins the Parent column against Book 1
// p.61's own "THE KNOWLEDGES-ONLY SKILLS" lists, which enumerate every
// parent skill and the Knowledges under it.
//
// This is a genuinely independent check on the p.60 transcription. The
// Parent values were derived from p.60's gutter labels and the row runs
// beneath them; p.61 states the same groupings in prose on a different
// page. Agreement between the two means the specialty blocks were
// segmented correctly, which nothing else verifies.
func TestKnowledgeGroupsMatchPage61(t *testing.T) {
	t.Parallel()

	// p.61: "Some skills (Animals, Driver, Engineer, Fighter, Flyer,
	// Gunner, Heavy Weapons, Pilot, Seafarer) include within them several
	// Knowledges." Sciences are the stand-alone Knowledges of the same
	// page's "other Knowledges are stand-alone sciences".
	want := map[string]int{
		"Animals":  3,  // Rider, Teamster, Trainer
		"Driver":   7,  // ACV, Automotive, Grav, Legged, Mole, Tracked, Wheeled
		"Engineer": 4,  // Jump Drives, Life Support, Maneuver Drive, Power Systems
		"Fighter":  7,  // Battle Dress, Beams, Blades, Exotics, Slug Throwers, Sprays, Unarmed
		"Flyer":    6,  // Aeronautics, Flapper, Grav, LTA, Rotor, Winged
		"Gunner":   5,  // Bay Weapons, Ortillery, Screens, Spines, Turrets
		"Hvy Wpns": 4,  // Artillery, Launcher, Ordnance, WMD
		"Pilot":    3,  // Small Craft, Spacecraft ACS, Spacecraft BCS
		"Seafarer": 5,  // Aquanautics, Grav, Boat, Ship, Sub
		"Sciences": 13, // Archeology through Sophontology
	}

	got := make(map[string]int, len(want))

	for _, skill := range educationSkills {
		if skill.Parent != "" {
			got[skill.Parent]++
		}
	}

	for parent, n := range want {
		if got[parent] != n {
			t.Errorf("%s has %d Knowledges, want %d (p.61's own list)", parent, got[parent], n)
		}
	}

	for parent := range got {
		if _, ok := want[parent]; !ok {
			t.Errorf("unexpected parent %q — p.61 names nine parents plus the Sciences", parent)
		}
	}
}

// TestANMSchoolPoolIsKnowledgesOnly is p.59's "Knowledge-2 from
// School=ANM", against the Training Course row one line away that reads
// "Skill-2 or Knowledge-2" — the book distinguishes the two, so the pool
// has to as well.
//
// The Army column is the corroboration worth pinning: every entry it
// flags is already a Knowledge, which is unlikely to be a coincidence
// and is what makes the restriction a reading of the page rather than an
// imposition on it.
func TestANMSchoolPoolIsKnowledgesOnly(t *testing.T) {
	t.Parallel()

	byName := make(map[string]educationSkill, len(educationSkills))
	for _, skill := range educationSkills {
		byName[skill.Name] = skill
	}

	for _, school := range []schoolSet{schoolMilitaryAcademy, schoolNavalAcademy, schoolMarine} {
		for _, name := range skillsForSchool(school, true) {
			if !byName[name].isKnowledge() {
				t.Errorf("%q is in an ANM School pool but is not a Knowledge", name)
			}
		}
	}

	// The Army column needs no restricting at all.
	all := skillsForSchool(schoolMilitaryAcademy, false)
	knowledges := skillsForSchool(schoolMilitaryAcademy, true)

	if len(all) != len(knowledges) {
		t.Errorf("Army column: %d flagged, %d of them Knowledges — want every one a Knowledge",
			len(all), len(knowledges))
	}

	// The other two do get restricted, or this rule would be inert.
	for _, school := range []schoolSet{schoolNavalAcademy, schoolMarine} {
		if len(skillsForSchool(school, true)) >= len(skillsForSchool(school, false)) {
			t.Error("the Knowledge restriction dropped nothing from a column that carries plain skills")
		}
	}

	// The Knowledge-Only parents themselves must never be drawable: p.61
	// says "the Skills themselves are not obtainable in Education or
	// Training", which is what excludes Gunner, Fighter and Hvy Wpns
	// despite p.60 flagging all three.
	for _, parent := range []string{"Gunner", "Fighter", "Hvy Wpns", "Engineer", "Pilot"} {
		for _, school := range []schoolSet{schoolMilitaryAcademy, schoolNavalAcademy, schoolMarine} {
			for _, name := range skillsForSchool(school, true) {
				if name == parent {
					t.Errorf("%q is a Knowledge-Only skill and must not be obtainable in Education", parent)
				}
			}
		}
	}
}
