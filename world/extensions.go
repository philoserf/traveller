package world

import (
	"fmt"

	"github.com/philoserf/traveller/ehex"
)

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

// TravelZone values, from least to most restricted: Green is unrestricted,
// Amber warns of danger (see TradeCode Puzzle/Dangerous), Red forbids entry.
const (
	ZoneGreen TravelZone = 'G'
	ZoneAmber TravelZone = 'A'
	ZoneRed   TravelZone = 'R'
)

// String returns the zone's display name (Green/Amber/Red), or "" for the
// zero value or any other byte that isn't one of the three constants.
// world.Generate always sets a real zone, so the zero value only arises
// from a hand-built or otherwise-partial World — not a case Generate's
// own callers need to handle, but one this method still handles safely.
func (z TravelZone) String() string {
	switch z {
	case ZoneGreen:
		return "Green"
	case ZoneAmber:
		return "Amber"
	case ZoneRed:
		return "Red"
	default:
		return ""
	}
}

// PBG encodes a system's Population digit, Belts, and Gas Giants.
type PBG struct {
	PopulationDigit ehex.Value
	Belts           ehex.Value
	GasGiants       ehex.Value
}

// String renders the PBG as its 3-character PopulationDigit-Belts-GasGiants
// code, e.g. "703". If any field is out of range, it falls back to
// ehex.Value's descriptive form for that digit instead of silently
// emitting '?', matching UWP.String() and UPP.String().
func (p PBG) String() string {
	fields := [3]ehex.Value{p.PopulationDigit, p.Belts, p.GasGiants}

	for _, f := range fields {
		if !f.Valid() {
			return fmt.Sprintf("%s%s%s", p.PopulationDigit, p.Belts, p.GasGiants)
		}
	}

	s := [3]byte{p.PopulationDigit.Byte(), p.Belts.Byte(), p.GasGiants.Byte()}

	return string(s[:])
}
