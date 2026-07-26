package character

import (
	"slices"

	"github.com/philoserf/traveller/dice"
)

// GenerateEntertainerCharacter generates a full Human Entertainer
// Character end to end: a UPP, a homeworld and its background skills, a
// full multi-term Entertainer career, and Entertainer's own Mustering
// Out benefits.
func GenerateEntertainerCharacter(r *dice.Roller) (Character, bool) {
	upp := GenerateUPP(r)
	homeworld, homeworldSkills := GenerateHomeworldSkills(r)

	return buildEntertainerCharacter(r, upp, homeworld, homeworldSkills)
}

// buildEntertainerCharacter assembles a Character from an already-rolled
// upp, homeworld, and homeworldSkills, mirroring buildScholarCharacter's
// own split for testability. Does not delegate to
// buildRiskCareerCharacter — ResolveEntertainerMusterOut needs fame,
// which buildRiskCareerCharacter's shared resolveMusterOut signature has
// no room for (see this slice's own plan-file Context).
func buildEntertainerCharacter(
	r *dice.Roller,
	upp UPP,
	homeworld string,
	homeworldSkills []SkillLevel,
) (Character, bool) {
	career, fame, careerUPP := resolveEntertainerCareerAndUPPWithBudget(r, upp, maxCareerTerms)
	career.MusteringOut = ResolveEntertainerMusterOut(r, career, fame)

	boostedUPP, bonuses := ApplyMusteringOut(career.MusteringOut, careerUPP)

	skills := append(slices.Clone(homeworldSkills), allSkillsFromTerms(career.Terms)...)

	// ok is purely "did the character ever qualify" (mirroring Rogue's
	// own no-death-mechanic convention), NOT gated on the last term's
	// RiskResult: Entertainer's own Dead means Talent completely spent,
	// not physical death (see ResolveEntertainerTerm's own doc comment),
	// so it must not skip aging the way a real Scout/Marine death does.
	// A code-review pass caught an earlier version of this checking
	// RiskResult != Dead here, which silently skipped ResolveAging for a
	// perfectly alive Entertainer whose Talent had merely run out.
	ok := len(career.Terms) > 0

	finalUPP, age, lifeStage, notes := finalizeAging(r, boostedUPP, len(career.Terms), ok)
	birthdate := GenerateBirthdate(r, age)

	// fame is already the character's own current Fame (the initial 2D
	// roll, or the last term's own FameAfterTerm) — set unconditionally,
	// per this slice's own plan-file Context ("Before Begin" ordering:
	// the initial roll happens before qualification is even determined).
	// bonuses.Fame is added unconditionally too, matching Rogue's own
	// buildRogueCharacter: it's already naturally 0 whenever there were
	// no Mustering Out rolls (the only case ok is false), so no
	// additional ok gate is needed — a code-review pass caught an
	// earlier version adding a redundant one here.
	totalFame := fame + bonuses.Fame

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
		// WoundBadges deliberately left at its zero value: Entertainer's
		// own Wounded/Disabled RiskResults represent a Talent setback,
		// not a physical wound (scoutWoundBadges' own reasoning is
		// specifically about physical characteristic reduction) — a
		// code-review pass caught an earlier version calling
		// scoutWoundBadges(career) here, which misrepresented a bad
		// show as a combat injury on the rendered sheet.
		Fame:    totalFame,
		Cash:    bonuses.Cash,
		Careers: []Career{career},
		Skills:  aggregateSkills(skills),
	}, ok
}
