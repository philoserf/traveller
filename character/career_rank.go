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
//
//nolint:unparam // see the enlistedTiers paragraph above.
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
