package character

import (
	"slices"

	"github.com/philoserf/traveller/dice"
)

// GenerateScholarCharacter generates a full Human Scholar Character end
// to end: a UPP, a homeworld and its background skills, a full
// multi-term Scholar career, and Scholar's own Mustering Out benefits.
func GenerateScholarCharacter(r *dice.Roller) (Character, bool) {
	upp := GenerateUPP(r)
	homeworld, homeworldSkills := GenerateHomeworldSkills(r)

	return buildScholarCharacter(r, upp, homeworld, homeworldSkills)
}

// buildScholarCharacter assembles a Character from an already-rolled
// upp, homeworld, and homeworldSkills, mirroring buildRogueCharacter's
// own split for testability. Does not delegate to
// buildRiskCareerCharacter — ResolveScholarMusterOut needs upp (to know
// Scholar's Edu-dependent starting rank tier), which
// buildRiskCareerCharacter's shared resolveMusterOut signature has no
// room for (see this slice's own plan-file Context).
func buildScholarCharacter(r *dice.Roller, upp UPP, homeworld string, homeworldSkills []SkillLevel) (Character, bool) {
	var aging agingSimulation

	career, careerUPP := resolveScholarCareerWithBudget(r, upp, maxCareerTerms, &aging)
	if aging.alive() {
		career.MusteringOut = ResolveScholarMusterOut(r, career, careerUPP)
	}

	// careerUPP, not the original upp — carries forward any Risk-reduced
	// characteristic from a survived Wounded/Disabled term (mirroring
	// buildRiskCareerCharacter's own use of resolveCareer's returned
	// updatedUPP). A code-review pass caught an earlier version of this
	// silently reverting that reduction by using upp here instead.
	boostedUPP, bonuses := ApplyMusteringOut(career, careerUPP)

	skills := append(slices.Clone(homeworldSkills), allSkillsFromTerms(career.Terms)...)
	skills = append(skills, bonuses.Skills...)

	survivedCareer := len(career.Terms) > 0 && career.Terms[len(career.Terms)-1].RiskResult != Dead

	age, lifeStage, notes, ok := finalizeAging(&aging, survivedCareer)
	birthdate := GenerateBirthdate(r, age)

	// Book 1 p.91 Fame Stacks over the individual awards.
	fameAwards := bonuses.FameAwards

	if survivedCareer {
		fameAwards = append(fameAwards, scholarSegmentFameAwards(careerUPP, career.Terms)...)
	}

	fame := resolveFameStacks(fameAwards)

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
		Rank:           lastTermRank(career.Terms),
		WoundBadges:    scoutWoundBadges(career),
		Fame:           fame,
		Cash:           bonuses.Cash,
		Careers:        []Career{career},
		Skills:         aggregateSkills(skills),
	}, ok
}
