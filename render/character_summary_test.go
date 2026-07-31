package render_test

import (
	"strings"
	"testing"

	"github.com/philoserf/traveller/character"
	"github.com/philoserf/traveller/render"
)

// TestCharacterSummaryMatchesTheWorkedExampleShape checks each line of
// the book's own Eneri Dinsha format against a hand-built fixture,
// documented departures aside (see CharacterSummary's own doc comment):
// no place name, one combined Skills line, no "Imperial Navy" prefix,
// no Current Date.
func TestCharacterSummaryMatchesTheWorkedExampleShape(t *testing.T) {
	t.Parallel()

	c := character.Character{
		Name:      "Eneri Dinsha",
		Homeworld: "A788899-C",
		Skills: []character.SkillLevel{
			{Name: "Psychology", Level: 2, Kind: character.Skill},
			{Name: "Robotics", Level: 1, Kind: character.Skill},
		},
		Rank:      "O1 Ensign",
		Age:       23,
		Birthdate: "Wonday 069-1082",
	}

	got := render.CharacterSummary(c)

	for _, want := range []string{
		"Eneri Dinsha " + c.UPP.String() + ".",
		"from: A788899-C",
		"Psychology-2, Robotics-1.",
		"O1 Ensign.",
		"Age 23. Born Wonday 069-1082.",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("CharacterSummary should contain %q, got:\n%s", want, got)
		}
	}

	if strings.Contains(got, "Current Date") {
		t.Errorf("CharacterSummary should never print a Current Date line, got:\n%s", got)
	}
}

// TestCharacterSummaryFallsBackToUPPWithoutAName mirrors
// characterTitle's own fallback: nothing in the character package
// generates Name yet, so the UPP-only first line is the common case.
func TestCharacterSummaryFallsBackToUPPWithoutAName(t *testing.T) {
	t.Parallel()

	c := character.Character{Homeworld: "A788899-C", Age: 18, Birthdate: "Wonday 002-1105"}

	got := render.CharacterSummary(c)

	firstLine := strings.SplitN(got, "\n", 2)[0]
	if !strings.HasPrefix(firstLine, c.UPP.String()+".") {
		t.Errorf("first line should be the bare UPP followed by a period, got %q", firstLine)
	}

	if strings.Contains(firstLine, "  ") {
		t.Errorf("first line should not carry a stray leading name separator, got %q", firstLine)
	}
}

// TestCharacterSummaryOmitsSkillsAndRankLinesWhenEmpty guards against a
// bare "." line for a character with no Skills, and no Rank line for a
// career with no rank concept (Book 1's own p.65 no-rank career list).
func TestCharacterSummaryOmitsSkillsAndRankLinesWhenEmpty(t *testing.T) {
	t.Parallel()

	c := character.Character{Homeworld: "A788899-C", Age: 18, Birthdate: "Wonday 002-1105"}

	got := render.CharacterSummary(c)

	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 3 {
		t.Errorf("expected exactly 3 lines (name, homeworld, age) with no Skills/Rank, got %d:\n%s", len(lines), got)
	}
}

// TestCharacterSummaryShowsGeneticProfileOnlyWhenNotDefault mirrors
// writeMetadata's own convention for the same field.
func TestCharacterSummaryShowsGeneticProfileOnlyWhenNotDefault(t *testing.T) {
	t.Parallel()

	human := character.Character{Homeworld: "A788899-C", Age: 18, Birthdate: "Wonday 002-1105"}
	if got := render.CharacterSummary(human); strings.Contains(got, "Genetic") {
		t.Errorf("CharacterSummary should omit Genetic Profile at the Human default, got:\n%s", got)
	}

	nonHuman := human
	nonHuman.GeneticProfile = "5435XX"

	if got := render.CharacterSummary(nonHuman); !strings.Contains(got, "Genetic 5435XX") {
		t.Errorf("CharacterSummary should show a non-default Genetic Profile, got:\n%s", got)
	}
}
