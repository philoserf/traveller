package character

import (
	"slices"

	"github.com/philoserf/traveller/dice"
)

// GenerateCitizenCharacter generates a full Human Citizen Character end
// to end: a UPP, a homeworld and its background skills, a full
// multi-term Citizen career, and Citizen's own Mustering Out benefits.
//
// Unlike GenerateScoutCharacter, this returns only Character, no ok
// bool: a Citizen attempt can't fail Career Resolution itself. Begin is
// Automatic (BeginCitizen always returns true) and Citizen Life has no
// wound or death mechanic at all, so every possible roll sequence
// produces a career with at least one term and no "this attempt didn't
// survive Career Resolution" outcome to report — a bool that can never
// observably be false would be dead surface area, not mirrored from
// Scout's own shape just for consistency's sake.
//
// This is narrower than "can't die at all": finalizeAging (below) still
// runs the full Aging simulation (character/aging.go's ResolveAging)
// against the resulting Character, which genuinely can end in death
// decades later (Book 1 p.89) — the same "Aging death isn't a voided
// attempt" distinction GenerateScoutCharacter's own doc comment makes.
// That death, if it happens, is recorded in Notes; it doesn't reintroduce
// an ok bool here, for the same reason.
func GenerateCitizenCharacter(r *dice.Roller) Character {
	upp := GenerateUPP(r)
	homeworld, homeworldSkills := GenerateHomeworldSkills(r)

	return buildCitizenCharacter(r, upp, homeworld, homeworldSkills)
}

// buildCitizenCharacter assembles a Character from an already-rolled
// upp, homeworld, and homeworldSkills, mirroring buildScoutCharacter's
// own split for testability (fixed upp fixtures instead of seed-hunting
// rare GenerateUPP outcomes). WoundBadges is left at its zero value —
// correct for Citizen, not a gap: Citizen Life has no wound mechanic to
// count. Character.UPP comes from finalizeAging after career Personal
// awards, Mustering Out characteristic boosts, and Aging reductions.
func buildCitizenCharacter(r *dice.Roller, upp UPP, homeworld string, homeworldSkills []SkillLevel) Character {
	career, careerUPP := resolveCitizenCareerAndUPPWithBudget(r, upp, maxCareerTerms)
	career.MusteringOut = ResolveCitizenMusterOut(r, career)

	boostedUPP, bonuses := ApplyMusteringOut(career.MusteringOut, careerUPP)

	skills := append(slices.Clone(homeworldSkills), allSkillsFromTerms(career.Terms)...)

	// survivedCareer is always true: Citizen Life has no death mechanic,
	// so there's always "the rest of their life" for finalizeAging to
	// simulate.
	finalUPP, age, lifeStage, notes := finalizeAging(r, boostedUPP, len(career.Terms), true)
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
		Skills:         aggregateSkills(skills),
	}
}
