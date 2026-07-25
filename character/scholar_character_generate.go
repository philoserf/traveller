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
	career, careerUPP := ResolveScholarCareer(r, upp)
	career.MusteringOut = ResolveScholarMusterOut(r, career, careerUPP)

	// careerUPP, not the original upp — carries forward any Risk-reduced
	// characteristic from a survived Wounded/Disabled term (mirroring
	// buildRiskCareerCharacter's own use of resolveCareer's returned
	// updatedUPP). A code-review pass caught an earlier version of this
	// silently reverting that reduction by using upp here instead.
	boostedUPP, bonuses := ApplyMusteringOut(career.MusteringOut, careerUPP)

	skills := append(slices.Clone(homeworldSkills), allSkillsFromTerms(career.Terms)...)

	ok := len(career.Terms) > 0 && career.Terms[len(career.Terms)-1].RiskResult != Dead

	finalUPP, age, lifeStage, notes := finalizeAging(r, boostedUPP, len(career.Terms), ok)
	birthdate := GenerateBirthdate(r, age)

	fame := bonuses.Fame

	if ok {
		// Edu (C5) is never a Risk & Reward position (scholarRiskRewardPositions
		// is C1-C4), so it's never reduced — careerUPP and upp agree here,
		// but careerUPP is used anyway for consistency with the rest of
		// this function, not because the two would ever actually differ.
		startTier := scholarStartTier(int(careerUPP.Characteristics[C5]))
		fame += scholarCareerFame(career.Terms, startTier)
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
		Rank:           lastTermRank(career.Terms),
		WoundBadges:    scoutWoundBadges(career),
		Fame:           fame,
		Cash:           bonuses.Cash,
		Careers:        []Career{career},
		Skills:         skills,
	}, ok
}

// scholarCareerFame is Book 1 p.91's own Fame table: two separate,
// unconditional, additive rows for Scholar — "=Rank" and "=Publications"
// — confirmed from the page image directly (the .txt OCR extraction
// visually scrambles this dense multi-column table). No Wound-Badge or
// Medal-tier contribution — those rows are nested under the page's own
// "Armed Forces" bracket (Army/Marine/Navy only), not universal.
func scholarCareerFame(terms []Term, startTier int) int {
	return scholarRankTier(terms, startTier) + scholarPublicationsTotal(terms)
}
