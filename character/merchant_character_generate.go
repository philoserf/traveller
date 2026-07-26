package character

import (
	"slices"

	"github.com/philoserf/traveller/dice"
)

// merchantCareerFame is Book 1 p.91's own "Merchant =Rank" Fame row —
// the character's current numeric tier regardless of track (0 while RX
// Temp, 1-3 while Enlisted R0-R2, 1-6 while Officer M1-M6), so it takes
// tier alone, not isOfficer — a code-review pass caught an earlier
// version threading a discarded isOfficer parameter through here, which
// hid the "Fame depends only on tier" rule behind a no-op argument. The
// page's own second row, "Ship Owner = 1D", is deliberately not
// modeled: this codebase tracks Mustering Out Ship Shares, not outright
// ship ownership, and the page gives no share-count threshold for
// "Owner" status.
func merchantCareerFame(tier int) int {
	return tier
}

// GenerateMerchantCharacter generates a full Human Merchant Character
// end to end: a UPP, a homeworld and its background skills, a full
// multi-term Merchant career, and Merchant's own Mustering Out
// benefits.
func GenerateMerchantCharacter(r *dice.Roller) (Character, bool) {
	upp := GenerateUPP(r)
	homeworld, homeworldSkills := GenerateHomeworldSkills(r)

	return buildMerchantCharacter(r, upp, homeworld, homeworldSkills)
}

// buildMerchantCharacter assembles a Character from an already-rolled
// upp, homeworld, and homeworldSkills, mirroring buildScholarCharacter's
// own split for testability. Does not delegate to
// buildRiskCareerCharacter — ResolveMerchantMusterOut and
// merchantCareerFame both need the final (isOfficer, tier), which
// buildRiskCareerCharacter's shared signature has no room for (the same
// Scholar-shaped mismatch from last slice).
func buildMerchantCharacter(r *dice.Roller, upp UPP, homeworld string, homeworldSkills []SkillLevel) (Character, bool) {
	career, careerUPP, isOfficer, tier := ResolveMerchantCareer(r, upp)
	career.MusteringOut = ResolveMerchantMusterOut(r, career, isOfficer, tier)

	// careerUPP, not the original upp — carries forward any Risk-reduced
	// characteristic from a survived Wounded/Disabled term. A code-review
	// pass caught an earlier version using upp here, silently reverting
	// that reduction (the same class of bug caught for Scholar last
	// slice).
	boostedUPP, bonuses := ApplyMusteringOut(career.MusteringOut, careerUPP)

	skills := append(slices.Clone(homeworldSkills), allSkillsFromTerms(career.Terms)...)

	// len(career.Terms) > 0 is unconditionally true (BeginMerchant never
	// fails) — ok collapses to "didn't die on the last term."
	ok := career.Terms[len(career.Terms)-1].RiskResult != Dead

	finalUPP, age, lifeStage, notes, ok := finalizeAging(r, boostedUPP, len(career.Terms), ok)
	birthdate := GenerateBirthdate(r, age)

	fame := bonuses.Fame
	if ok {
		fame += merchantCareerFame(tier)
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
		Skills:         aggregateSkills(skills),
	}, ok
}
