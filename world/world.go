// Package world models a single Traveller5 world: the UWP, trade codes,
// and Importance/Economic/Cultural extensions, plus the UWP-field-rolling
// primitives (RollStarport, RollSize, ...) both mainworld generation here
// and secondary-world generation in package system are built from.
// Star-system structure (stars, orbits, gas giants, satellites) lives in
// package system; sector-scale generation lives in package sector.
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
	// Ring is Book 3 p.29's "S Number of Satellites": a satellite-count
	// roll of exactly 0 gives this world a Ring (and rerolls the count
	// once more). Only meaningful for a world with satellites generated
	// at all — see system/satellite_generate.go.
	Ring bool
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
