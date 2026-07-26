package character

// rankState derives the current Armed Forces rank track and tier from
// prior terms — derived, not separately tracked, matching this codebase's
// existing "derive from Terms" discipline (nobleIntrigueCounts,
// citizenLifeSuccessCount). Every Armed Forces character begins Enlisted
// at tier 1 (Book 1 p.65's own "begin with enlisted rank... Marine1"/
// "Soldier1"/"Spacer1") — the zero-terms case. A Commissioned term resets
// to Officer tier 1; a Promoted term advances the current track's own
// tier by 1, capped at enlistedTiers/officerTiers (each career's own rank
// table length — 6/7 for both Marine and Soldier, confirmed as a real
// shared convention now that two careers agree, not assumed).
//
// Originally marineRankState, generalized to take explicit tier counts
// once Soldier became a second real caller with the identical shape.
//
// enlistedTiers happens to be 6 for both current callers (Marine,
// Soldier), but that's the two careers' own rank tables agreeing, not a
// guarantee — passing it explicitly (derived from each career's own
// len(xEnlistedRankNames), not hardcoded here) keeps this function
// correct if a future Armed Forces career's own table differs, without
// needing to revisit this signature.
func rankState(terms []Term, enlistedTiers, officerTiers int) (bool, int) {
	isOfficer := false
	tier := 1

	for _, t := range terms {
		switch {
		case t.Commissioned:
			isOfficer = true
			tier = 1
		case t.Promoted:
			if isOfficer {
				tier = min(tier+1, officerTiers)
			} else {
				tier = min(tier+1, enlistedTiers)
			}
		}
	}

	return isOfficer, tier
}

// rankBasedCareerFame is Marine's/Soldier's/Spacer's own shared Fame
// body (Book 1 p.91's "Army/Marine/Navy: Officer Rank*" bracket) —
// confirmed byte-identical across all three careers except which
// rank-name tables get passed in, extracted per this codebase's own
// "generalize on 2nd instance" discipline once a third verbatim match
// (Spacer) appeared. See marineCareerFameAwards's own doc comment
// (marine_character_generate.go) for the full formula rationale: Medal
// Fame + Wound Badge Fame (x1 each) + Officer Rank Fame (=Rank, the
// numeric tier).
func rankBasedCareerFame(career Career, enlistedRankCount, officerRankCount int) int {
	fame := 0

	for _, t := range career.Terms {
		for _, medal := range t.Medals {
			fame += medalFame[medal]
		}
	}

	fame += scoutWoundBadges(career)

	if isOfficer, tier := rankState(career.Terms, enlistedRankCount, officerRankCount); isOfficer {
		fame += tier
	}

	return fame
}

// rankAutoSkillFromTables is Spacer's/Merchant's own shared "Automatic
// Skills by Rank" lookup body — confirmed byte-identical between
// spacerRankAutomaticSkill and merchantRankAutoSkill, extracted per this
// codebase's own "generalize on 2nd instance" discipline. enlisted/
// officer are each career's own tier -> skill-name map (an absent tier
// grants nothing). Marine's/Soldier's own switch-based lookups aren't
// folded in here: different code shape and different data, not
// copy-paste — Spacer only moved to a map because its own extra tier
// pushed a flat switch over golangci-lint's cyclomatic complexity limit.
func rankAutoSkillFromTables(enlisted, officer map[int]string, isOfficer bool, tier int) (SkillLevel, bool) {
	table := enlisted
	if isOfficer {
		table = officer
	}

	name, ok := table[tier]
	if !ok {
		return SkillLevel{}, false
	}

	return skillLevel1(name, Skill), true
}
