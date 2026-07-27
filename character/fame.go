package character

// fameStackCap is Book 1 p.91's own threshold in the Fame Stacks rule.
// On the same page's Fame descriptor scale it is Sector — the reach at
// which accumulation stops meaning anything, with Subsector below it and
// Domain, Empire and eventually "All Reality" above.
const fameStackCap = 20

// resolveFameStacks applies Book 1 p.91's Fame Stacks rule to the
// individual Fame awards a character received:
//
//	"A character's Fame is the sum of all Fame points received to 20;
//	beyond 20, only the highest Fame applies."
//
// Fame is a scale of reach, not a quantity — p.91's descriptors run
// Neighborhood, Town, City, ... Subsector, Sector at 20, then Domain and
// beyond. That is what makes the two clauses coherent: local reputations
// genuinely add up, so a character known in several regions is known
// more widely, but past Sector they stop adding, because being famous in
// two sectors is not the same as being famous across a Domain. Past that
// point the only thing that matters is the single furthest-reaching
// thing the character did.
//
// So the sum is taken, capped at 20, and a single award that already
// exceeds 20 on its own overrides that cap. A Scout whose Discoveries
// alone made him known across a Domain does not have that reduced to
// Sector because the rest of his career was unremarkable.
//
// Written to be monotonic: adding another award can never lower a
// character's Fame. A plain reading of "only the highest Fame applies"
// on its own would do exactly that — a character with awards of 12 would
// drop to Fame 12 on receiving a second award of 8 — which is why the
// two clauses have to be read together rather than in sequence.
func resolveFameStacks(awards []int) int {
	sum, highest := 0, 0

	for _, award := range awards {
		sum += award
		highest = max(highest, award)
	}

	return max(min(sum, fameStackCap), highest)
}

// sumInts totals Fame awards for the places that still report a single
// per-career figure alongside the awards themselves.
func sumInts(values []int) int {
	total := 0
	for _, v := range values {
		total += v
	}

	return total
}

// The Fame sources below report each award separately rather than a
// total, because Book 1 p.91's Fame Stacks rule counts "Fame points
// received" individually: it caps their sum at 20 but lets a single
// award larger than that stand alone. A Scout's twelve Discoveries are
// twelve awards of 4, not one award of 48 — the difference decides
// whether the cap binds at all.

// scoutDiscoveryFameAwards is p.91's "Scout: Discoveries x4", one award
// per Discovery.
//
// Book 1 states this twice and the two disagree. p.79's Scout Reward
// Success box says a Discovery "receives a Land Grant, and Fame +1";
// p.91's Fame table row reads "Scout | Discoveries | x4". Ruled for
// p.91, against this codebase's usual "a career's own box beats the
// generic summary" precedent (the reading behind scoutRiskRewardPositions
// following p.79's C1 C2 C3 over p.64's four-entry row), because p.91 is
// not a summary. It is the Fame chapter's own table under a "Mult"
// column that carries every source's award formula, including the
// non-multiplier ones — "Scholar =Rank", "Scholar =Publications",
// "Merchant Ship Owner = 1D", "Craftsman Masterpieces x3", "Craftsman
// Perfect Masterpieces x5". A column that has to say "=Rank" to mean
// rank is defining awards, not restating them.
//
// Leaving x4 in place also keeps #82's Fame Stacks decomposition intact:
// the per-award granularity that makes p.91's cap of 20 bind at all was
// derived against awards of 4.
func scoutDiscoveryFameAwards(career Career) []int {
	var awards []int

	for _, t := range career.Terms {
		if t.RewardResult == "Discovery" {
			awards = append(awards, scoutDiscoveryFameMultiplier)
		}
	}

	return awards
}

// scoutDiscoveryFameMultiplier is the "x4" in that row.
const scoutDiscoveryFameMultiplier = 4

// rankBasedCareerFameAwards is p.91's Armed Forces Fame, one award per
// medal and per Wound Badge, plus one for Officer Rank.
func rankBasedCareerFameAwards(career Career, enlistedRankCount, officerRankCount int) []int {
	var awards []int

	for _, t := range career.Terms {
		for _, medal := range t.Medals {
			if fame := medalFame[medal]; fame > 0 {
				awards = append(awards, fame)
			}
		}
	}

	for range scoutWoundBadges(career) {
		awards = append(awards, 1) // "Wound Badge WB x1"
	}

	if isOfficer, tier := rankState(career.Terms, enlistedRankCount, officerRankCount); isOfficer && tier > 0 {
		awards = append(awards, tier)
	}

	return awards
}

// rogueTermFameAwards is p.91's "Rogue: Successful Schemes x2, Failed
// Schemes x3", one award per term.
func rogueTermFameAwards(terms []Term) []int {
	awards := make([]int, 0, len(terms))

	for _, t := range terms {
		if t.Imprisoned {
			awards = append(awards, 3)
		} else {
			awards = append(awards, 2)
		}
	}

	return awards
}

// scholarSegmentFameAwards is p.91's "Scholar =Rank, =Publications" as
// separate awards: one for Rank Tier, then one per publication (2 for an
// award-winning one, per scholarPublicationsTotal's own scale).
func scholarSegmentFameAwards(upp UPP, terms []Term) []int {
	var awards []int

	if tier := scholarRankTier(terms, scholarStartTier(int(upp.Characteristics[C5]))); tier > 0 {
		awards = append(awards, tier)
	}

	for _, t := range terms {
		switch {
		case t.AwardWinning:
			awards = append(awards, 2)
		case t.PublicationSucceeded:
			awards = append(awards, 1)
		}
	}

	return awards
}
