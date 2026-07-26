package character

import (
	"slices"
	"testing"

	"github.com/philoserf/traveller/ehex"
)

func TestApplyPersonalAwardsUpdatesCharacteristics(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{5, 5, 5, 5, 5, 5}}
	awards := []SkillLevel{
		{Name: "Str", Level: 1, Kind: Personal},
		{Name: "C2", Level: 2, Kind: Personal},
		{Name: "End", Level: 1, Kind: Personal},
		{Name: "Int", Level: 1, Kind: Personal},
		{Name: "Edu", Level: 1, Kind: Personal},
		{Name: "Soc", Level: 1, Kind: Personal},
		{Name: "Pilot", Level: 9, Kind: Skill},
	}

	got := applyPersonalAwards(upp, awards)
	want := UPP{Characteristics: [6]ehex.Value{6, 7, 6, 6, 6, 6}}

	if got != want {
		t.Errorf("applyPersonalAwards(...) = %v, want %v", got, want)
	}
}

func TestAllSkillsFromTermsExcludesPersonalAwards(t *testing.T) {
	t.Parallel()

	terms := []Term{{SkillsAwarded: []SkillLevel{
		{Name: "Str", Level: 1, Kind: Personal},
		{Name: "Pilot", Level: 1, Kind: Skill},
	}}}

	want := []SkillLevel{{Name: "Pilot", Level: 1, Kind: Skill}}
	if got := allSkillsFromTerms(terms); !slices.Equal(got, want) {
		t.Errorf("allSkillsFromTerms(...) = %+v, want %+v", got, want)
	}
}

func TestAggregateSkillsNoDuplicatesPassesThroughUnchanged(t *testing.T) {
	t.Parallel()

	skills := []SkillLevel{
		{Name: "Pilot", Level: 1, Kind: Skill},
		{Name: "Str", Level: 1, Kind: Personal},
		{Name: "Vacc Suit", Level: 1, Kind: Skill},
	}

	got := aggregateSkills(skills)
	if !slices.Equal(got, skills) {
		t.Errorf("aggregateSkills(%+v) = %+v, want unchanged", skills, got)
	}
}

func TestAggregateSkillsSumsRepeatedSameNameSameKindGrants(t *testing.T) {
	t.Parallel()

	skills := []SkillLevel{
		{Name: "Bureaucrat", Level: 1, Kind: Skill},
		{Name: "Admin", Level: 1, Kind: Skill},
		{Name: "Bureaucrat", Level: 1, Kind: Skill},
		{Name: "Bureaucrat", Level: 1, Kind: Skill},
	}

	want := []SkillLevel{
		{Name: "Bureaucrat", Level: 3, Kind: Skill},
		{Name: "Admin", Level: 1, Kind: Skill},
	}

	if got := aggregateSkills(skills); !slices.Equal(got, want) {
		t.Errorf("aggregateSkills(%+v) = %+v, want %+v", skills, got, want)
	}
}

// TestAggregateSkillsKeepsDistinctKindsSeparate confirms a same-named
// entry under a different Kind (e.g. a "Str" Personal boost vs a
// hypothetical "Str" Skill) never merges — Kind is part of the grouping
// key, not just Name.
func TestAggregateSkillsKeepsDistinctKindsSeparate(t *testing.T) {
	t.Parallel()

	skills := []SkillLevel{
		{Name: "Str", Level: 1, Kind: Personal},
		{Name: "Str", Level: 1, Kind: Skill},
	}

	if got := aggregateSkills(skills); !slices.Equal(got, skills) {
		t.Errorf("aggregateSkills(%+v) = %+v, want unchanged (different Kind)", skills, got)
	}
}

// TestAggregateSkillsPreservesFirstGrantOrder confirms a merged entry
// appears at its first grant's own position, not its last (or moved to
// the front/back).
func TestAggregateSkillsPreservesFirstGrantOrder(t *testing.T) {
	t.Parallel()

	skills := []SkillLevel{
		{Name: "Admin", Level: 1, Kind: Skill},
		{Name: "Bureaucrat", Level: 1, Kind: Skill},
		{Name: "Trader", Level: 1, Kind: Skill},
		{Name: "Bureaucrat", Level: 1, Kind: Skill},
	}

	want := []SkillLevel{
		{Name: "Admin", Level: 1, Kind: Skill},
		{Name: "Bureaucrat", Level: 2, Kind: Skill},
		{Name: "Trader", Level: 1, Kind: Skill},
	}

	if got := aggregateSkills(skills); !slices.Equal(got, want) {
		t.Errorf("aggregateSkills(%+v) = %+v, want %+v", skills, got, want)
	}
}

func TestAggregateSkillsEmptyInput(t *testing.T) {
	t.Parallel()

	if got := aggregateSkills(nil); len(got) != 0 {
		t.Errorf("aggregateSkills(nil) = %+v, want empty", got)
	}
}
