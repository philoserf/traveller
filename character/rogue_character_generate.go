package character

import (
	"slices"

	"github.com/philoserf/traveller/dice"
)

// GenerateRogueCharacter generates a full Human Rogue Character end to
// end: a UPP, a homeworld and its background skills, a full multi-term
// Rogue career, and Rogue's own Mustering Out benefits.
//
// Returns (Character, bool) mirroring GenerateNobleCharacter's own
// signature, not GenerateMarineCharacter's own risk-career shape: Rogue
// has a real "never qualified" outcome (BeginRogue can fail) but no
// death mechanic at all (Risk failure means Prison, not a fatal
// characteristic reduction), the same reasoning buildNobleCharacter's
// own doc comment already gives for its own identical shape.
func GenerateRogueCharacter(r *dice.Roller) (Character, bool) {
	upp, homeworld, homeworldSkills, education := generateStart(r)

	c, ok := buildRogueCharacter(r, upp, homeworld, homeworldSkills, &education)
	c = applyEducation(c, education)

	return c, ok
}

// buildRogueCharacter assembles a Character from an already-rolled upp,
// homeworld, and homeworldSkills, mirroring buildNobleCharacter's own
// split for testability. Does not delegate to buildRiskCareerCharacter —
// Rogue has neither Wounded/Disabled/Dead nor WoundBadges, a genuinely
// different shape from Marine/Soldier/Spacer, not just different data.
//
// Fame and Cash are computed via rogueTermsCash rather than solely
// through ApplyMusteringOut, since Rogue's own intrinsic Fame (Book 1
// p.91: "Successful Schemes x2 / Failed Schemes x3" — see this slice's
// own plan-file Context for why this formula is implemented instead of
// the career box's own "+1 Infamy" line) and Cash (each term's own
// SchemePayoff) both come from the career's own Terms, not from
// Mustering Out rolls alone.
func buildRogueCharacter(
	r *dice.Roller, upp UPP, homeworld string, homeworldSkills []SkillLevel, education *Education,
) (Character, bool) {
	var aging agingSimulation

	career, careerUPP := // A standalone Rogue has served nothing else, so p.84's
		// select-a-previous-career option never applies.
		resolveRogueCareerAndUPPWithBudget(r, upp, maxCareerTerms, &aging, nil, education)
	if aging.alive() {
		career.MusteringOut = ResolveRogueMusterOut(r, career)
	}

	boostedUPP, bonuses := ApplyMusteringOut(r, career, careerUPP)

	skills := append(slices.Clone(homeworldSkills), allSkillsFromTerms(career.Terms)...)
	skills = append(skills, bonuses.Skills...)

	survivedCareer := len(career.Terms) > 0

	age, lifeStage, notes, ok := finalizeAging(&aging, survivedCareer)
	birthdate := GenerateBirthdate(r, age)

	cash := bonuses.Cash
	// Book 1 p.91 Fame Stacks over the individual awards.
	fameAwards := bonuses.FameAwards

	if survivedCareer {
		termCash := rogueTermsCash(career.Terms)
		fameAwards = append(fameAwards, rogueTermFameAwards(career.Terms)...)
		cash += termCash
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
		Fame:           fame,
		Cash:           cash,
		Careers:        []Career{career},
		Skills:         aggregateSkills(skills),
		LandGrants:     bonuses.LandGrants,
	}, ok
}

// rogueTermsCash sums each term's own Cash payoff (SchemePayoff) —
// shared by buildRogueCharacter and resolveRogueSegment
// (career_chain.go), extracted once both were confirmed to duplicate
// the identical loop body. The p.91 "Successful Schemes x2 / Failed
// Schemes x3" Fame these terms also earn is rogueTermFameAwards' job
// (fame.go): the Fame Stacks cap needs the awards individually, so
// returning a pre-summed total here would discard what it reads.
func rogueTermsCash(terms []Term) int {
	cash := 0

	for _, t := range terms {
		cash += t.SchemePayoff
	}

	return cash
}
