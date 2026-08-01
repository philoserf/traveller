package character

import (
	"slices"
	"strconv"
	"strings"

	"github.com/philoserf/traveller/dice"
)

// merchantCareerFame is Book 1 p.91's own "Merchant =Rank" Fame row —
// the character's current numeric tier regardless of track (0 while RX
// Temp, 1-3 while Enlisted R0-R2, 1-6 while Officer M1-M6), so it takes
// tier alone, not isOfficer — a code-review pass caught an earlier
// version threading a discarded isOfficer parameter through here, which
// hid the "Fame depends only on tier" rule behind a no-op argument. The
// page's own second row, "Ship Owner = 1D", is merchantShipOwnerFame
// below.
func merchantCareerFame(tier int) int {
	return tier
}

// shipOwnershipMinimumShares is the fewest shares that buy any ship at
// all on Book 1 p.90's own Ship Shares table — a 100-ton Scout, at 2.
// The rest cost more (Free Trader and Yacht 4, Lab Ship and Close
// Escort 8), so 2 is the point at which a character stops holding a
// fraction of nothing and owns something.
//
// p.91's Fame row is "Ship Owner = 1D" with no threshold of its own,
// which is why this was previously read as unspecified. The share costs
// are a page earlier, in the Ship Shares section rather than the Fame
// table.
const shipOwnershipMinimumShares = 2

// merchantShipShares totals the ship shares a Merchant career produced,
// from both sources p.90 describes as acquiring them.
//
// Term rewards escalate — Book 1 p.80's Escalating Ship Shares, where
// the Nth successful Reward grants N shares — and are recorded in the
// term's own Reward result ("2 Ship Shares"). Mustering Out grants one
// per "Ship Share" benefit rolled.
func merchantShipShares(career Career) int {
	shares := 0

	for _, term := range career.Terms {
		if n, ok := shipSharesAwarded(term.RewardResult); ok {
			shares += n
		}
	}

	for _, benefit := range career.MusteringOut.Benefits {
		if benefit == shipShareBenefit {
			shares++
		}
	}

	return shares
}

// shipShareBenefit is the Mustering Out table entry granting one share.
const shipShareBenefit = "Ship Share"

// shipSharesAwarded parses a term's own "N Ship Share"/"N Ship Shares"
// Reward result back into its count.
func shipSharesAwarded(rewardResult string) (int, bool) {
	count, rest, found := strings.Cut(rewardResult, " Ship Share")
	if !found || (rest != "" && rest != "s") {
		return 0, false
	}

	n, err := strconv.Atoi(count)
	if err != nil {
		return 0, false
	}

	return n, true
}

// merchantShipOwnerFame rolls Book 1 p.91's own "Merchant: Ship Owner =
// 1D" — Fame for a Merchant whose accumulated shares actually buy a
// ship, which p.90 makes possible from two shares up.
//
// Rolled rather than fixed: the row gives 1D, not a constant. Returns 0
// for a Merchant still short of a ship, who owns a fraction and is no
// more famous for it.
func merchantShipOwnerFame(r *dice.Roller, career Career) int {
	if merchantShipShares(career) < shipOwnershipMinimumShares {
		return 0
	}

	return r.D6()
}

// GenerateMerchantCharacter generates a full Human Merchant Character
// end to end: a UPP, a homeworld and its background skills, a full
// multi-term Merchant career, and Merchant's own Mustering Out
// benefits.
func GenerateMerchantCharacter(r *dice.Roller) (Character, bool) {
	upp, homeworld, homeworldSkills, education := generateStart(r)

	c, ok := buildMerchantCharacter(r, upp, homeworld, homeworldSkills)
	c = applyEducation(c, education)

	return c, ok
}

// buildMerchantCharacter assembles a Character from an already-rolled
// upp, homeworld, and homeworldSkills, mirroring buildScholarCharacter's
// own split for testability. Does not delegate to
// buildRiskCareerCharacter — ResolveMerchantMusterOut and
// merchantCareerFame both need the final (isOfficer, tier), which
// buildRiskCareerCharacter's shared signature has no room for (the same
// Scholar-shaped mismatch from last slice).
func buildMerchantCharacter(r *dice.Roller, upp UPP, homeworld string, homeworldSkills []SkillLevel) (Character, bool) {
	var aging agingSimulation

	career, careerUPP, isOfficer, tier := resolveMerchantCareerWithBudget(r, upp, maxCareerTerms, &aging)
	if aging.alive() {
		career.MusteringOut = ResolveMerchantMusterOut(r, career, isOfficer, tier)
	}

	// careerUPP, not the original upp — carries forward any Risk-reduced
	// characteristic from a survived Wounded/Disabled term. A code-review
	// pass caught an earlier version using upp here, silently reverting
	// that reduction (the same class of bug caught for Scholar last
	// slice).
	boostedUPP, bonuses := ApplyMusteringOut(r, career, careerUPP)

	skills := append(slices.Clone(homeworldSkills), allSkillsFromTerms(career.Terms)...)
	skills = append(skills, bonuses.Skills...)

	// len(career.Terms) > 0 is unconditionally true (BeginMerchant never
	// fails) — ok collapses to "didn't die on the last term."
	survivedCareer := career.Terms[len(career.Terms)-1].RiskResult != Dead

	// Fame is rolled before the birthdate, matching resolveMerchantSegment
	// (career_chain.go) so both generation paths draw the Ship Owner die
	// at the same point in the sequence and produce the same character
	// for a given seed.
	// Book 1 p.91 Fame Stacks: the individual awards, not a running
	// total — a single award past 20 stands where an accumulated one
	// would be capped there.
	fameAwards := bonuses.FameAwards
	if survivedCareer {
		fameAwards = append(fameAwards, merchantCareerFame(tier))
		if owner := merchantShipOwnerFame(r, career); owner > 0 {
			fameAwards = append(fameAwards, owner)
		}
	}

	fame := resolveFameStacks(fameAwards)

	age, lifeStage, notes, ok := finalizeAging(&aging, survivedCareer)
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
		Rank:           lastTermRank(career.Terms),
		WoundBadges:    careerWoundBadges(career),
		Fame:           fame,
		Cash:           bonuses.Cash,
		Careers:        []Career{career},
		Skills:         aggregateSkills(skills),
		LandGrants:     bonuses.LandGrants,
	}, ok
}
