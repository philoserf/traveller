package character

import (
	"slices"

	"github.com/philoserf/traveller/dice"
)

// GenerateNobleCharacter generates a full Human Noble Character end to
// end: a UPP, a homeworld and its background skills, a full multi-term
// Noble career, and Noble's own Mustering Out benefits.
//
// Returns a genuinely new combination, not a copy of either existing
// career's own shape: like GenerateScoutCharacter, this returns
// (Character, bool) — unlike GenerateCitizenCharacter's ok-less
// signature — because BeginNoble can fail (Soc < B, Book 1 p.85's own
// hard prerequisite), a real "never qualified" outcome that needs a
// signal, the same reason Scout needs one. But ok here only ever checks
// len(career.Terms) > 0, with no Risk-death branch at all: Return &
// Intrigue has no death mechanic (matching Citizen's own precedent on
// that front), so there's nothing analogous to Scout's own
// RiskResult != Dead check to make.
func GenerateNobleCharacter(r *dice.Roller) (Character, bool) {
	upp := GenerateUPP(r)
	homeworld, homeworldSkills := GenerateHomeworldSkills(r)

	return buildNobleCharacter(r, upp, homeworld, homeworldSkills)
}

// buildNobleCharacter assembles a Character from an already-rolled upp,
// homeworld, and homeworldSkills, mirroring buildScoutCharacter's/
// buildCitizenCharacter's own split for testability. WoundBadges is left
// at its zero value — correct for Noble, not a gap: Return & Intrigue
// has no wound mechanic to count.
func buildNobleCharacter(r *dice.Roller, upp UPP, homeworld string, homeworldSkills []SkillLevel) (Character, bool) {
	career := ResolveNobleCareer(r, upp)
	career.MusteringOut = ResolveNobleMusterOut(r, career)

	boostedUPP, bonuses := ApplyMusteringOut(career.MusteringOut, upp)

	skills := append(slices.Clone(homeworldSkills), allSkillsFromTerms(career.Terms)...)

	ok := len(career.Terms) > 0

	finalUPP, age, lifeStage, notes := finalizeAging(r, boostedUPP, len(career.Terms), ok)
	birthdate := GenerateBirthdate(r, age)

	return Character{
		Species:        "Human",
		GeneticProfile: humanGeneticProfile,
		UPP:            finalUPP,
		Homeworld:      homeworld,
		Birthworld:     homeworld,
		Birthdate:      birthdate,
		Age:            age,
		LifeStage:      lifeStage,
		Notes:          notes,
		Fame:           bonuses.Fame,
		Cash:           bonuses.Cash,
		Careers:        []Career{career},
		Skills:         skills,
	}, ok
}
