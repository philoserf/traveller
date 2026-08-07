package character

import (
	"slices"

	"github.com/philoserf/traveller/dice"
)

// GenerateCitizenCharacter generates a full Human Citizen Character end
// to end: a UPP, a homeworld and its background skills, a full
// multi-term Citizen career, and Citizen's own Mustering Out benefits.
//
// Returns an ok bool like every other generator, but for a narrower
// reason than theirs: Citizen genuinely can't fail Career Resolution
// itself (Begin is Automatic — BeginCitizen always returns true — and
// Citizen Life has no wound or death mechanic at all, so every roll
// sequence produces a career with at least one term). ok is false here
// only for an Aging death (Book 1 p.89), which finalizeAging
// (character/aging.go) simulates over the career's own span and which
// a Citizen is no more immune to than anyone else. This function used
// to return Character alone on the reasoning that a never-false bool
// would be dead surface area; that held only while Aging death was
// treated as a Notes-only outcome — see finalizeAging's own doc comment
// for why it no longer is.
func GenerateCitizenCharacter(r *dice.Roller) (Character, bool) {
	upp, homeworld, homeworldSkills, education := generateStart(r)

	c, ok := buildCitizenCharacter(r, upp, homeworld, homeworldSkills, &education)
	c = applyEducation(c, education)

	return c, ok
}

// buildCitizenCharacter assembles a Character from an already-rolled
// upp, homeworld, and homeworldSkills, mirroring buildScoutCharacter's
// own split for testability (fixed upp fixtures instead of seed-hunting
// rare GenerateUPP outcomes). WoundBadges is left at its zero value —
// correct for Citizen, not a gap: Citizen Life has no wound mechanic to
// count. Character.UPP comes from finalizeAging after career Personal
// awards, Mustering Out characteristic boosts, and Aging reductions.
func buildCitizenCharacter(
	r *dice.Roller, upp UPP, homeworld string, homeworldSkills []SkillLevel, education *Education,
) (Character, bool) {
	var aging agingSimulation

	career, careerUPP := resolveCitizenCareerAndUPPWithBudget(r, upp, maxCareerTerms, &aging, education)
	if aging.alive() {
		career.MusteringOut = ResolveCitizenMusterOut(r, career)
	}

	boostedUPP, bonuses := ApplyMusteringOut(r, career, careerUPP)

	skills := append(slices.Clone(homeworldSkills), allSkillsFromTerms(career.Terms)...)
	skills = append(skills, bonuses.Skills...)

	// survivedCareer is always true: Citizen Life has no death mechanic,
	// so there's always "the rest of their life" for finalizeAging to
	// simulate. The returned ok can still be false — Aging death.
	age, lifeStage, notes, ok := finalizeAging(&aging, true)
	birthdate := GenerateBirthdate(r, age)

	return Character{
		Species:        "Human",
		GeneticProfile: humanGeneticProfile,
		UPP:            boostedUPP,
		Homeworld:      homeworld,
		Birthworld:     homeworld,
		Birthdate:      birthdate,
		Age:            age,
		LifeStage:      lifeStage,
		Notes:          notes,
		// p.91: "Citizen no intrinsic Fame" — Mustering Out awards only.
		Fame:       resolveFameStacks(bonuses.FameAwards),
		Cash:       bonuses.Cash,
		Careers:    []Career{career},
		Skills:     aggregateSkills(skills),
		LandGrants: bonuses.LandGrants,
	}, ok
}
