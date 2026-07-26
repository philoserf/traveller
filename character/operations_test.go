package character

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/philoserf/traveller/ehex"

	"github.com/philoserf/traveller/dice"
)

// TestEligibleSkillColumnsAlwaysIncludesPersonal is Book 1 p.65's
// "Column 1-Personal Skills may always be rolled" — true even for a term
// whose every Operations result maps to no column at all.
func TestEligibleSkillColumnsAlwaysIncludesPersonal(t *testing.T) {
	t.Parallel()

	cases := [][]string{
		nil,
		{},
		{"ANM School"},
		{"ANM School", "ANM School", "ANM School", "ANM School"},
		{"not an operation"},
	}

	for _, received := range cases {
		got := eligibleSkillColumns(received, marineOperationsColumns)
		if !slices.Equal(got, []int{personalSkillColumn}) {
			t.Errorf("eligibleSkillColumns(%v) = %v, want only the Personal column", received, got)
		}
	}
}

// TestEligibleSkillColumnsForMixedOperations is the issue's own
// mixed-Operations case: a term receiving four different assignments
// makes four different columns eligible, plus Personal.
func TestEligibleSkillColumnsForMixedOperations(t *testing.T) {
	t.Parallel()

	// Marine columns: 0 Personal, 1 Garrison, 2 Combat, 3 Peacekeeper,
	// 5 Mission.
	got := eligibleSkillColumns([]string{"Combat", "Mission", "Peace Keeper", "Garrison"}, marineOperationsColumns)
	if want := []int{0, 1, 2, 3, 5}; !slices.Equal(got, want) {
		t.Errorf("mixed Marine Operations = %v, want %v", got, want)
	}

	// Spacer's own Patrol and Strike share the printed column heading
	// "4 Patrol/Strike", so receiving both makes one column eligible,
	// not two.
	got = eligibleSkillColumns([]string{"Patrol", "Strike"}, spacerOperationsColumns)
	if want := []int{0, 3}; !slices.Equal(got, want) {
		t.Errorf("Patrol+Strike = %v, want %v (one shared column)", got, want)
	}
}

// TestEligibleSkillColumnsDedupesRepeatedOperations is the issue's own
// repeated-Operations case, and matches p.65's own second worked example
// — Eneri rolls "3, 3, 4, 4" for "Siege, Siege, Patrol, Patrol". Four
// results, two distinct columns, plus Personal.
func TestEligibleSkillColumnsDedupesRepeatedOperations(t *testing.T) {
	t.Parallel()

	got := eligibleSkillColumns([]string{"Siege", "Siege", "Patrol", "Patrol"}, spacerOperationsColumns)
	if want := []int{0, 3, 4}; !slices.Equal(got, want) {
		t.Errorf("Siege,Siege,Patrol,Patrol = %v, want %v", got, want)
	}

	// Order must not matter: the same set of assignments must offer the
	// same columns however they were rolled.
	shuffled := eligibleSkillColumns([]string{"Patrol", "Siege", "Patrol", "Siege"}, spacerOperationsColumns)
	if !slices.Equal(got, shuffled) {
		t.Errorf("column set depends on roll order: %v vs %v", got, shuffled)
	}
}

// TestEligibleSkillColumnsMatchesTheWorkedExample walks Book 1 p.65's
// own first example end to end. Eneri rolls Naval Operations (4, 1, 5)
// for "assignments in Patrol, Battle, and Mission", and the book says he
// "can consult the columns for 1 Personal, 3 Battle, 4 Patrol/Strike,
// 5 Siege, 6 Mission".
//
// Four of those five are what this function returns. The fifth — 5 Siege
// — is not an Operations result he received, and is reachable only
// because the rule exempts "commission, promotion, or other skill
// eligibilities": Eneri had four Term skills plus one promotion skill.
// Asserting Siege's *absence* here is what pins that reading, and what
// would catch an implementation that simply allowed every column.
func TestEligibleSkillColumnsMatchesTheWorkedExample(t *testing.T) {
	t.Parallel()

	// The rolls themselves, against the real table: 4 -> Patrol,
	// 1 -> Battle, 5 -> Mission.
	for roll, want := range map[int]string{4: "Patrol", 1: "Battle", 5: "Mission"} {
		if got := spacerNavalOperationsNames[roll-1]; got != want {
			t.Fatalf("Naval Operations roll %d = %q, want %q", roll, got, want)
		}
	}

	// "The highest associated Mod is 3."
	best := 0
	for _, name := range []string{"Patrol", "Battle", "Mission"} {
		best = max(best, operationMod(name, spacerNavalOperationsNames[:], spacerNavalOperationsMods[:]))
	}

	if best != 3 {
		t.Errorf("highest Mod of Patrol/Battle/Mission = %d, want 3", best)
	}

	got := eligibleSkillColumns([]string{"Patrol", "Battle", "Mission"}, spacerOperationsColumns)
	if want := []int{0, 2, 3, 5}; !slices.Equal(got, want) {
		t.Errorf("eligible columns = %v, want %v (1 Personal, 3 Battle, 4 Patrol/Strike, 6 Mission)", got, want)
	}

	const siegeColumn = 4
	if slices.Contains(got, siegeColumn) {
		t.Error("Siege is eligible for Term skills, but Eneri never received it — " +
			"the book reaches that column through the promotion skill the rule exempts")
	}
}

// TestNoOperationGrantsOccupationOrTechnical records a property of all
// three tables that would otherwise look like a transcription gap: no
// Operations result in any Armed Forces career maps to the Occupation/
// Siege or Technical column. Those are reachable only through the
// exempted eligibilities, which is exactly what p.65's example shows.
func TestNoOperationGrantsOccupationOrTechnical(t *testing.T) {
	t.Parallel()

	const technicalColumn = 6

	for name, columns := range map[string]map[string]int{
		"Marine":  marineOperationsColumns,
		"Soldier": soldierOperationsColumns,
	} {
		for op, column := range columns {
			if column == technicalColumn {
				t.Errorf("%s Operation %q maps to Technical, which no Operations table grants", name, op)
			}
		}
	}

	// Spacer is the exception worth stating: its own Siege Operation does
	// map to the Siege column. Marine's and Soldier's column 4 is
	// Occupation, which neither table's Operations can produce.
	if spacerOperationsColumns["Siege"] != 4 {
		t.Error("Spacer's Siege Operation should map to its own Siege column")
	}

	const occupationColumn = 4

	for name, columns := range map[string]map[string]int{
		"Marine":  marineOperationsColumns,
		"Soldier": soldierOperationsColumns,
	} {
		for op, column := range columns {
			if column == occupationColumn {
				t.Errorf("%s Operation %q maps to Occupation, which its table cannot grant", name, op)
			}
		}
	}
}

// TestOperationsColumnsNameRealOperations guards the maps against a
// typo'd key, which would silently make a real Operations result grant
// no column at all — indistinguishable at runtime from ANM School.
func TestOperationsColumnsNameRealOperations(t *testing.T) {
	t.Parallel()

	cases := []struct {
		career  string
		columns map[string]int
		names   []string
	}{
		{"Marine", marineOperationsColumns, marineOperationsNames[:]},
		{"Soldier", soldierOperationsColumns, soldierOperationsNames[:]},
		{"Spacer", spacerOperationsColumns, spacerNavalOperationsNames[:]},
	}

	for _, c := range cases {
		for op := range c.columns {
			if !slices.Contains(c.names, op) {
				t.Errorf("%s: %q is in the column map but is not an Operations result", c.career, op)
			}
		}

		// Every Operations result except ANM School must map somewhere,
		// or a real assignment would grant no column.
		for _, name := range c.names {
			if name == "ANM School" {
				continue
			}

			if _, ok := c.columns[name]; !ok {
				t.Errorf("%s: Operations result %q maps to no skill column", c.career, name)
			}
		}
	}
}

// TestRollSkillsFromColumnsStaysInsideItsColumns is the restriction
// itself: every skill drawn must come from a cell in an eligible column.
// Checked by name against the table, since the returned SkillLevel does
// not record where it came from.
func TestRollSkillsFromColumnsStaysInsideItsColumns(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(13, 13))
	columns := []int{personalSkillColumn, 2}

	allowed := map[string]bool{}

	for _, column := range columns {
		for _, name := range marineSkillTable[column] {
			allowed[name] = true
		}
	}

	// Personal-column cells become "<name> +1" style characteristic
	// grants, so match on the raw table text for those too.
	for range 2000 {
		for _, skill := range rollSkillsFromColumns(r, marineSkillTable, columns, 4) {
			if !allowed[skill.Name] && skill.Kind != Personal {
				t.Fatalf("drew %q, which is in no eligible column %v", skill.Name, columns)
			}
		}
	}
}

// TestRollSkillsFromColumnsDrawsTwoDicePerSkill pins the draw shape: one
// die to pick the column, one for the row. Verified against a column set
// of size 1, where the column pick is forced and the whole draw is
// therefore predictable.
//
// Total dice per skill is not fixed and deliberately not asserted — some
// cells ("One Trade", "Any Skill") consume further dice to resolve
// themselves, in this function and the unrestricted one alike.
func TestRollSkillsFromColumnsDrawsTwoDicePerSkill(t *testing.T) {
	t.Parallel()

	// Column 2 of the Marine table is Combat, whose six cells are all
	// plain skill names that resolve without further dice.
	const combatColumn = 2

	for _, name := range marineSkillTable[combatColumn] {
		switch name {
		case "Major", "Minor", "One Science", "One Trade", "One Art", "Any Skill", "Any Knowledge", "Capital":
			t.Fatalf("column %d cell %q resolves with extra dice, so this test's premise is gone",
				combatColumn, name)
		}
	}

	const seed = 17

	got := dice.New(rand.NewPCG(seed, seed))
	skills := rollSkillsFromColumns(got, marineSkillTable, []int{combatColumn}, 3)

	if len(skills) != 3 {
		t.Fatalf("got %d skills, want 3 (every cell in this column resolves)", len(skills))
	}

	// Replay: with one eligible column the column die is forced, so each
	// skill is the row die alone against that column.
	want := dice.New(rand.NewPCG(seed, seed))

	for i, skill := range skills {
		want.Uniform(1) // the forced column pick still costs its die

		row := want.Uniform(6) - 1
		if expected := marineSkillTable[combatColumn][row]; skill.Name != expected {
			t.Errorf("skill %d = %q, want %q (row %d of the Combat column)", i, skill.Name, expected, row)
		}
	}
}

// TestArmedForcesTermsRecordEveryOperation confirms the four results
// reach Term.Operations for all three careers, and that Assignment is
// one of them rather than a fifth value.
func TestArmedForcesTermsRecordEveryOperation(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}

	r := dice.New(rand.NewPCG(23, 23))

	for range 100 {
		marine, _ := ResolveMarineTerm(r, upp, C1, "Commando", 0, nil)
		soldier, _ := ResolveSoldierTerm(r, upp, C1, "Infantry", 0, nil)
		spacer, _ := ResolveSpacerTerm(r, upp, C1, 0, nil)

		for _, c := range []struct {
			name string
			term Term
		}{{"Marine", marine}, {"Soldier", soldier}, {"Spacer", spacer}} {
			if len(c.term.Operations) != operationsRollsPerTerm {
				t.Fatalf("%s: Operations = %v, want %d entries",
					c.name, c.term.Operations, operationsRollsPerTerm)
			}

			if !slices.Contains(c.term.Operations, c.term.Assignment) {
				t.Errorf("%s: Assignment %q is not among Operations %v",
					c.name, c.term.Assignment, c.term.Operations)
			}
		}
	}
}
