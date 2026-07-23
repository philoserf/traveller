// Package world models Traveller5 worlds and star systems: the UWP,
// trade codes, extensions, and system/star/orbit structure.
package world

// World is a single mainworld's full profile.
type World struct {
	Name       string
	Sector     string
	Hex        string
	UWP        UWP
	TradeCodes []TradeCode
	Importance Importance
	Economic   Economic
	Cultural   Cultural
	Nobility   []string
	Allegiance string
	Bases      []Base
	TravelZone TravelZone
	PBG        PBG
	Worlds     int
	Notes      string
}

// stringsOf converts a slice of any string-backed enum type (TradeCode,
// Base, ...) to plain strings, e.g. for joining into display text.
func stringsOf[T ~string](vals []T) []string {
	s := make([]string, len(vals))
	for i, v := range vals {
		s[i] = string(v)
	}

	return s
}
