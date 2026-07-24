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
// bool: a Citizen attempt can't fail. Begin is Automatic (BeginCitizen
// always returns true) and Citizen Life has no wound or death mechanic
// at all, so every possible roll sequence produces a career with at
// least one term and no "this attempt didn't survive" outcome to
// report — a bool that can never observably be false would be dead
// surface area, not mirrored from Scout's own shape just for
// consistency's sake.
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
// count. Character.UPP is set directly from upp, not from
// ResolveCitizenCareer's return — it only returns Career, since Citizen
// Life never modifies a characteristic.
func buildCitizenCharacter(r *dice.Roller, upp UPP, homeworld string, homeworldSkills []SkillLevel) Character {
	career := ResolveCitizenCareer(r, upp)
	career.MusteringOut = ResolveCitizenMusterOut(r, career)

	skills := append(slices.Clone(homeworldSkills), allSkillsFromTerms(career.Terms)...)

	return Character{
		Species:        "Human",
		GeneticProfile: humanGeneticProfile,
		UPP:            upp,
		Homeworld:      homeworld,
		Birthworld:     homeworld,
		Careers:        []Career{career},
		Skills:         skills,
	}
}
