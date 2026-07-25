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
	upp := GenerateUPP(r)
	homeworld, homeworldSkills := GenerateHomeworldSkills(r)

	return buildRogueCharacter(r, upp, homeworld, homeworldSkills)
}

// buildRogueCharacter assembles a Character from an already-rolled upp,
// homeworld, and homeworldSkills, mirroring buildNobleCharacter's own
// split for testability. Does not delegate to buildRiskCareerCharacter —
// Rogue has neither Wounded/Disabled/Dead nor WoundBadges, a genuinely
// different shape from Marine/Soldier/Spacer, not just different data.
//
// Fame and Cash are computed here rather than solely through
// ApplyMusteringOut, since Rogue's own intrinsic Fame (Book 1 p.91:
// "Successful Schemes x2 / Failed Schemes x3" — see this slice's own
// plan-file Context for why this formula is implemented instead of the
// career box's own "+1 Infamy" line) and Cash (each term's own
// SchemePayoff) both come from the career's own Terms, not from
// Mustering Out rolls alone.
func buildRogueCharacter(r *dice.Roller, upp UPP, homeworld string, homeworldSkills []SkillLevel) (Character, bool) {
	career := ResolveRogueCareer(r, upp)
	career.MusteringOut = ResolveRogueMusterOut(r, career)

	boostedUPP, bonuses := ApplyMusteringOut(career.MusteringOut, upp)

	skills := append(slices.Clone(homeworldSkills), allSkillsFromTerms(career.Terms)...)

	ok := len(career.Terms) > 0

	finalUPP, age, lifeStage, notes := finalizeAging(r, boostedUPP, len(career.Terms), ok)
	birthdate := GenerateBirthdate(r, age)

	cash := bonuses.Cash
	fame := bonuses.Fame

	if ok {
		for _, t := range career.Terms {
			cash += t.SchemePayoff

			if t.Imprisoned {
				fame += 3 // Book 1 p.91's own "Failed Schemes x3"
			} else {
				fame += 2 // "Successful Schemes x2"
			}
		}
	}

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
		Fame:           fame,
		Cash:           cash,
		Careers:        []Career{career},
		Skills:         skills,
	}, ok
}
