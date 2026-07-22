package world

import "github.com/philoserf/traveller/ehex"

// Importance (Ix) ranks a world's significance, roughly -3 to +5.
type Importance int

// Economic (Ex) extension: Resources, Labor, Infrastructure, Efficiency.
type Economic struct {
	Resources      int
	Labor          int
	Infrastructure int
	Efficiency     int
}

// Cultural (Cx) extension: Heterogeneity, Acceptance, Strangeness, Symbols.
type Cultural struct {
	Heterogeneity int
	Acceptance    int
	Strangeness   int
	Symbols       int
}

// TravelZone is a world's danger/restriction rating.
type TravelZone byte

const (
	ZoneGreen TravelZone = 'G'
	ZoneAmber TravelZone = 'A'
	ZoneRed   TravelZone = 'R'
)

// PBG encodes a system's Population digit, Belts, and Gas Giants.
type PBG struct {
	PopulationDigit ehex.Value
	Belts           int
	GasGiants       int
}
