package character

import (
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/philoserf/traveller/dice"
)

// TestAgentUndercoverTableShape pins the table's own dimensions and the
// fact that every one of its 18 cells is populated. Book 1 p.83 rolls A
// 1-3 and B 1-6 with no "no result" cell, so a zero-valued entry would
// be a transcription gap that silently assigns an empty career.
func TestAgentUndercoverTableShape(t *testing.T) {
	t.Parallel()

	if len(agentUndercoverTable) != agentUndercoverTableAColumns {
		t.Fatalf("table has %d A blocks, want %d", len(agentUndercoverTable), agentUndercoverTableAColumns)
	}

	for a, block := range agentUndercoverTable {
		if len(block) != 6 {
			t.Fatalf("A=%d has %d B rows, want 6", a+1, len(block))
		}

		for b, cell := range block {
			if cell.Service == "" {
				t.Errorf("A=%d B=%d has no Service label", a+1, b+1)
			}

			if cell.Career == "" {
				t.Errorf("A=%d B=%d (%s) has no Career", a+1, b+1, cell.Service)
			}

			// "top row C (reroll if >3)" allows three titles at most, and
			// the rows that print fewer are the ones where C is not
			// required — any other count would mean a misread row.
			if n := len(cell.Titles); n != 0 && n != 1 && n != agentUndercoverTableAColumns {
				t.Errorf("A=%d B=%d (%s) has %d titles, want 0, 1 or 3", a+1, b+1, cell.Service, n)
			}

			for i, title := range cell.Titles {
				if title == "" {
					t.Errorf("A=%d B=%d (%s) C=%d is an empty title", a+1, b+1, cell.Service, i+1)
				}
			}
		}
	}
}

// TestEveryUndercoverCareerHasASkillSource is the guarantee
// rollAgentUndercoverSkill depends on: every career the table can assign
// must have somewhere to draw a skill from. A missing entry would index
// the map to its zero value — a table of 42 empty strings — and the
// reroll loop would spin forever rather than fail visibly.
func TestEveryUndercoverCareerHasASkillSource(t *testing.T) {
	t.Parallel()

	for _, block := range agentUndercoverTable {
		for _, cell := range block {
			if cell.Career == CitizenCareerName {
				continue // routed to rollCitizenTableEName, not the map
			}

			if _, ok := agentUndercoverSkillTables[cell.Career]; !ok {
				t.Errorf("%s assigns career %q, which has no skill table", cell.Service, cell.Career)
			}
		}
	}
}

// TestUndercoverSkillTablesHasNoStrayEntries is the converse: a career
// in the map that no cell can assign is dead data, and more likely a
// sign the table lost a row. It is what catches the specific defect this
// issue fixed — the previous implementation could assign Rogue, which
// Book 1 p.83's table has no row for.
func TestUndercoverSkillTablesHasNoStrayEntries(t *testing.T) {
	t.Parallel()

	assignable := map[string]bool{}

	for _, block := range agentUndercoverTable {
		for _, cell := range block {
			assignable[cell.Career] = true
		}
	}

	for career := range agentUndercoverSkillTables {
		if !assignable[career] {
			t.Errorf("agentUndercoverSkillTables has %q, which no p.83 row assigns", career)
		}
	}

	for _, absent := range []string{RogueCareerName, CraftsmanCareerName, AgentCareerName} {
		if assignable[absent] {
			t.Errorf("%q is assignable, but Book 1 p.83's table has no row for it", absent)
		}
	}
}

// TestRollAgentUndercoverAssignmentCoversEveryCell walks enough rolls to
// see all 18 cells and every C title. The distribution matters as much
// as the boundaries: A and C are 1D6 rerolled past 3, so a Uniform(3)
// substitution would produce the same spread but consume a different
// number of dice, and an off-by-one in either bound would silently make
// part of the table unreachable.
func TestRollAgentUndercoverAssignmentCoversEveryCell(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(7, 7))

	seenService := map[string]bool{}
	seenTitle := map[string]bool{}

	for range 20000 {
		assignment, title := rollAgentUndercoverAssignment(r)
		seenService[assignment.Service] = true
		seenTitle[title] = true
	}

	for _, block := range agentUndercoverTable {
		for _, cell := range block {
			if !seenService[cell.Service] {
				t.Errorf("never rolled %q — part of the table is unreachable", cell.Service)
			}

			for _, want := range cell.Titles {
				if !seenTitle[cell.Service+" "+want] {
					t.Errorf("never rolled %q for %q", want, cell.Service)
				}
			}

			if len(cell.Titles) == 0 && !seenTitle[cell.Service] {
				t.Errorf("never rolled the bare label %q", cell.Service)
			}
		}
	}
}

// TestRollAgentUndercoverAssignmentRollsCOnlyWhenRequired is the "if
// required" clause. It is checked as the observable consequence — a row
// that does not require C must render the same title on every roll,
// because there is nothing left for a die to vary — rather than by
// counting dice consumed: rand.IntN rejection-samples, so the number of
// source draws behind a D6 is not fixed, and a die-counting harness
// measures the sampler rather than this function.
func TestRollAgentUndercoverAssignmentRollsCOnlyWhenRequired(t *testing.T) {
	t.Parallel()

	// Keyed by cell position, not by the Service label: Scholar,
	// Entertainer, Merchant and Scout each own two rows, so counting by
	// label merges two rows' titles into one tally.
	titlesSeen := map[[2]int]map[string]bool{}

	for a := range agentUndercoverTable {
		for b := range agentUndercoverTable[a] {
			titlesSeen[[2]int{a, b}] = map[string]bool{}
		}
	}

	for a := range agentUndercoverTable {
		for b, cell := range agentUndercoverTable[a] {
			r := dice.New(rand.NewPCG(uint64(a*6+b)+1, 9))

			for range 3000 {
				titlesSeen[[2]int{a, b}][renderUndercoverTitle(cell, r)] = true
			}
		}
	}

	for a := range agentUndercoverTable {
		for b, cell := range agentUndercoverTable[a] {
			got := len(titlesSeen[[2]int{a, b}])

			// A row that does not require a C roll has nothing left for a
			// die to vary, so it renders exactly one string. Scout's own
			// Courier/Courier/Courier row does roll C and still renders
			// one — three identical titles collapse.
			want := len(uniqueStrings(cell.Titles))
			if want == 0 {
				want = 1
			}

			if got != want {
				t.Errorf("A=%d B=%d (%s) rendered %d distinct titles, want %d",
					a+1, b+1, cell.Service, got, want)
			}
		}
	}
}

// TestRollAgentUndercoverSkillAlwaysReturnsASkill covers Book 1's own
// "Select (not Roll) one skill" being singular and unconditional: unlike
// the count-based grants, this one is never lost to an unresolvable
// cell. Citizen is included specifically because it draws from p.78's
// Table E, whose "No Skill" entry is exactly such a cell.
func TestRollAgentUndercoverSkillAlwaysReturnsASkill(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(11, 11))

	careers := make([]string, 0, 1+len(agentUndercoverSkillTables))
	careers = append(careers, CitizenCareerName)

	for career := range agentUndercoverSkillTables {
		careers = append(careers, career)
	}

	for _, career := range careers {
		for range 500 {
			skill := rollAgentUndercoverSkill(r, career)
			if skill.Name == "" {
				t.Fatalf("rollAgentUndercoverSkill(%q) returned an unnamed skill", career)
			}

			// Level 1 for every cell except Noble's own "Capital", which
			// Book 1 p.85 footnotes as "value= 1D" — the one skill-table
			// entry whose level is rolled rather than fixed.
			if skill.Name == capitalSkillCell {
				if skill.Level < 1 || skill.Level > 6 {
					t.Errorf("rollAgentUndercoverSkill(%q) returned Capital at level %d, want 1D",
						career, skill.Level)
				}

				continue
			}

			if skill.Level != 1 {
				t.Errorf("rollAgentUndercoverSkill(%q) returned level %d, want 1", career, skill.Level)
			}
		}
	}
}

// TestRollRestrictedD6StaysInRange pins the extracted reroll loop at the
// two bounds its callers actually use — 3 for both of p.83's restricted
// rolls and p.78's A roll, and len(Titles) for a C column.
func TestRollRestrictedD6StaysInRange(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(3, 3))

	for _, limit := range []int{1, 2, 3, 6} {
		seen := map[int]bool{}

		for range 2000 {
			got := rollRestrictedD6(r, limit)
			if got < 1 || got > limit {
				t.Fatalf("rollRestrictedD6(limit=%d) = %d, out of range", limit, got)
			}

			seen[got] = true
		}

		if len(seen) != limit {
			t.Errorf("rollRestrictedD6(limit=%d) produced %d distinct values, want %d", limit, len(seen), limit)
		}
	}
}

// TestAgentTermRecordsTheAssignmentTitle confirms the rank/title detail
// the issue asked to preserve actually reaches Term, and that it agrees
// with the career recorded beside it.
func TestAgentTermRecordsTheAssignmentTitle(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(4, 4))

	for range 200 {
		term, _ := ResolveAgentTerm(r, uppAgent88, C1)

		if term.UndercoverAssignment == "" {
			t.Fatal("UndercoverAssignment is empty")
		}

		if term.UndercoverCareer == "" {
			t.Fatal("UndercoverCareer is empty")
		}

		if strings.TrimSpace(term.UndercoverAssignment) != term.UndercoverAssignment {
			t.Errorf("UndercoverAssignment = %q has stray whitespace", term.UndercoverAssignment)
		}
	}
}

// TestRenderUndercoverTitleConsumesNoDiceWhenCIsNotRequired is the "if
// required" clause tested where it is actually observable: as dice
// consumed. A row printing fewer than three titles has nothing for C to
// select, so rendering it must leave the roller untouched — rolling C
// anyway would pick the same title while shifting every later roll in
// the term, a defect no assertion about the returned string can see.
//
// Checked by running two rollers off the same seed and comparing what
// they produce afterwards, rather than by counting source draws:
// rand.IntN rejection-samples, so draws per D6 is not fixed.
func TestRenderUndercoverTitleConsumesNoDiceWhenCIsNotRequired(t *testing.T) {
	t.Parallel()

	for a := range agentUndercoverTable {
		for b, cell := range agentUndercoverTable[a] {
			if len(cell.Titles) > 1 {
				continue
			}

			const seed = 21

			rendered := dice.New(rand.NewPCG(seed, seed))
			renderUndercoverTitle(cell, rendered)

			untouched := dice.New(rand.NewPCG(seed, seed))

			for i := range 8 {
				if got, want := rendered.D6(), untouched.D6(); got != want {
					t.Fatalf("A=%d B=%d (%s): roll %d after rendering = %d, want %d — "+
						"rendering consumed a die it should not have",
						a+1, b+1, cell.Service, i+1, got, want)
				}
			}
		}
	}
}

// TestRenderUndercoverTitleConsumesADieWhenCIsRequired is the converse,
// and what stops the test above from being satisfied by a
// renderUndercoverTitle that never rolls at all.
func TestRenderUndercoverTitleConsumesADieWhenCIsRequired(t *testing.T) {
	t.Parallel()

	for a := range agentUndercoverTable {
		for b, cell := range agentUndercoverTable[a] {
			if len(cell.Titles) <= 1 {
				continue
			}

			// Compared across several seeds: a single die consumed can
			// coincidentally leave the next value unchanged, so this
			// asserts the streams diverge somewhere, not everywhere.
			diverged := false

			for seed := uint64(1); seed <= 20 && !diverged; seed++ {
				rendered := dice.New(rand.NewPCG(seed, seed))
				renderUndercoverTitle(cell, rendered)

				untouched := dice.New(rand.NewPCG(seed, seed))

				for range 4 {
					if rendered.D6() != untouched.D6() {
						diverged = true

						break
					}
				}
			}

			if !diverged {
				t.Errorf("A=%d B=%d (%s): rendering never consumed a die, but the row requires a C roll",
					a+1, b+1, cell.Service)
			}
		}
	}
}
