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
	// LandGrants are p.88's own awards of territory, retained at
	// Mustering Out per p.68. Cumulative: a Noble holds one per Soc
	// increase, a Scout one per Discovery.
	LandGrants []LandGrant
}

// NobleTitle is Book 1 p.88's own Soc-to-title mapping — "Gentleman" at
// Soc A through "Archduke" at G — and empty below Soc A, which is no
// noble rank at all.
//
// Derived from Soc rather than stored, so it cannot disagree with the
// characteristic it comes from, and so it applies to every career
// without each builder having to remember: p.68's Knighthood raises Soc
// to B from any career's Mustering Out, and that is a Knight whether or
// not the character ever entered the Noble career.
func (c Character) NobleTitle() string {
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
