package character

// medalCodes is Book 1 p.70's own "IMPERIAL MEDALS" table, keyed by the
// raw Reward roll (2-13; index 0 = roll 2). Roll 13 ("SEHD", SEH With
// Diamonds) is only reachable via the table's own "If Officer, increase
// +1" bonus (each Armed Forces career's own ResolveXTerm applies it to
// the raw roll before this lookup) — a natural 12 plus that bonus.
//
// Shared by every Armed Forces career, not Marine-specific — the table's
// own closing line reads "Medals = Soldier, Spacer, Marine Promotion
// Mods," confirmed directly against the page image. Originally
// marineMedalCodes, generalized once Soldier became a second real
// caller with byte-identical values.
var medalCodes = [12]string{
	"XS", "XS", "XS", "XS", "XS", "XS", "XS", // rolls 2-8
	"MCUF", "MCUF", // rolls 9-10
	"MCG",  // roll 11
	"SEH",  // roll 12
	"SEHD", // roll 13 (Officer bonus only)
}

// medalNames are the Medals table's own "Medal Description" column, used
// as each Armed Forces career's own RewardResult text — render's existing
// termOutcomeLine already prints any non-"None" RewardResult generically,
// so no render change is needed for any career using this table.
var medalNames = map[string]string{
	"XS":   "XS Exemplary Service",
	"MCUF": "MCUF Meritorious Conduct Under Fire",
	"MCG":  "MCG Medal for Conspicuous Gallantry",
	"SEH":  "SEH Starburst for Extreme Heroism",
	"SEHD": "SEH With Diamonds",
}

// medalFame is Book 1 p.91's own per-medal Fame contribution — distinct
// from medalRewardMod (this table's own Mod column, p.70), which feeds
// Promotion rolls.
var medalFame = map[string]int{"XS": 0, "MCUF": 1, "MCG": 2, "SEH": 3, "SEHD": 4}

// medalRewardMod is Book 1 p.70's own Medals table Mod column — named
// for what it's keyed by (a medal earned via the Reward roll,
// medalFromReward) and distinctly from the local "medalMod" variable
// each career's own ResolveXTerm computes (the cumulative sum this map
// feeds into), not from Book 1's own Tier concept (career rank
// progression, scholarRankTier/commandCollegeOfficerTier) — this table
// has nothing to do with that.
var medalRewardMod = map[string]int{"XS": 1, "MCUF": 2, "MCG": 3, "SEH": 4, "SEHD": 5}

// medalFromReward converts a raw Reward roll (2-13, already including the
// Officer +1 bonus if applicable) into its Medals table code.
func medalFromReward(roll int) string {
	return medalCodes[roll-2]
}

// medalModSum sums medals' own Mod values (medalRewardMod).
func medalModSum(medals []string) int {
	total := 0
	for _, m := range medals {
		total += medalRewardMod[m]
	}

	return total
}

// medalModTotal sums every medal Mod earned across terms — Book 1 p.66's
// own fully-worked Eneri Dinsha example computes Promotion's own target
// as "Soc plus Medal Mods," cumulative across the whole career to that
// point, not just the current term's own medals.
func medalModTotal(terms []Term) int {
	total := 0
	for _, t := range terms {
		total += medalModSum(t.Medals)
	}

	return total
}
