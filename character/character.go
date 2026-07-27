// Package character models Traveller5 player and non-player characters:
// characteristics, careers, skills, and combat-relevant stats.
package character

// Character is a full player or non-player character.
type Character struct {
	Name           string
	Species        string
	GeneticProfile string // 6-char code identifying which trait each UPP position represents, e.g. "SDEIES"
	UPP            UPP
	Homeworld      string
	Birthworld     string
	Birthdate      string
	Age            int
	LifeStage      int
	Careers        []Career
	Skills         []SkillLevel
	Rank           string
	Medals         []string
	Commendations  []string
	WoundBadges    int
	Fame           int
	Cash           int
	Equipment      []string
	Notes          string
	// Masterpieces are the structured records of everything a Craftsman
	// created — QREBS allocation, Master Points and creation age.
	// Equipment above still carries the same items as display strings.
	Masterpieces []Masterpiece
	// Education is what CharGen step C (Book 1 p.72) did to this
	// character before any career began — which school was attended, the
	// Major and Minor declared there, and whether a degree was earned.
	// Zero-valued for a character who attended nothing.
	Education Education
	// LandGrants are p.88's own awards of territory, retained at
	// Mustering Out per p.68. Cumulative: a Noble holds one per Soc
	// increase, a Scout one per Discovery.
	LandGrants []LandGrant
}

// NobleTitle is the noble title this character holds: the rank the
// Noble career actually walked to if it walked one, otherwise Book 1
// p.88's Soc-to-title mapping — "Gentleman" at Soc A through "Archduke"
// at G, and empty below Soc A, which is no noble rank at all.
//
// Rank leads and Soc follows, which is why the career's tracked ladder
// position wins wherever one exists:
//
//	p.65: "Elevation is Roll High (roll Soc or greater to be Elevated
//	to the next higher Noble rank) and its associated increase in
//	Social Standing (if any)."
//
// The "(if any)" is load-bearing. Three pairs of p.88 rows share a Soc —
// c Baronet and C Baron at 12, e Viscount and E Count at 14, f and F
// Duke at 15 — so a Soc value cannot name a rank on its own, and a
// Baronet elevated to Baron has genuinely advanced without any
// characteristic moving. Deriving the title from Soc alone reported a
// rank the character was never elevated to for 25.7% of generated Nobles
// (63 of 245 over 3,000 seeds), in both directions: ahead of the ladder
// where Mustering Out raised Soc after the career ended, behind it where
// an elevation raised no Soc.
//
// The Soc fallback is not a leftover. It is what makes p.68's Knighthood
// confer a title out of any career's Mustering Out — "A Knighthood
// raises any value of Soc to B" — for a character who never entered the
// Noble career and so has no ladder position to read.
//
// Note that a Mustering Out Soc increase still awards its Land Grant per
// p.85 even when it moves no title here; a fief follows Soc, a title
// follows the ladder. See ApplyMusteringOut.
func (c Character) NobleTitle() string {
	if title, ok := c.nobleLadderRank(); ok {
		return title
	}

	return NobleTitleForSoc(c.UPP.Characteristics[C6])
}

// MasterpieceValue is what every Masterpiece this character holds would
// fetch today, with Book 1 p.75's own Vintage appreciation applied
// against the character's current age. Sale-time Flux is not included —
// see Masterpiece.VintageValueAtSale.
func (c Character) MasterpieceValue() int {
	total := 0

	for _, m := range c.Masterpieces {
		total += m.VintageValue(c.Age)
	}

	return total
}

// LandGrantIncome is the total annual profit, in Credits, of every Land
// Grant this character holds — Book 1 p.88's own "LAND GRANT VALUE".
// A separate accessor rather than a stored field because it is derived:
// storing it would let it drift from the grants it summarizes.
func (c Character) LandGrantIncome() int {
	return totalLandGrantIncome(c.LandGrants)
}

// nobleLadderRank is the rank the Noble career walked to, if this
// character served one that reached a rank at all.
func (c Character) nobleLadderRank() (string, bool) {
	for _, career := range c.Careers {
		if career.Name != NobleCareerName {
			continue
		}

		if rank := lastTermRank(career.Terms); rank != "" {
			return rank, true
		}
	}

	return "", false
}
