package world

// Base is a facility present at a world.
type Base string

// Base values. NavalBase, ScoutBase, NavalDepot, and WayStation use their
// official T5 single-letter abbreviations; Military/Scientific/Diplomatic/
// Cultural bases have no defined single-letter code — the rulebook calls
// them out as referee-assigned exceptions instead of giving one.
const (
	NavalBase      Base = "N"
	ScoutBase      Base = "S"
	NavalDepot     Base = "D"
	WayStation     Base = "W"
	MilitaryBase   Base = "Military"
	ScientificBase Base = "Scientific"
	DiplomaticBase Base = "Diplomatic"
	CulturalBase   Base = "Cultural"
)

// BaseStrings converts bases to their plain string form, e.g. for joining
// into display text.
func BaseStrings(bases []Base) []string {
	return stringsOf(bases)
}
