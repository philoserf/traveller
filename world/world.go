// Package world models a single Traveller5 world: the UWP, trade codes,
// and Importance/Economic/Cultural extensions, plus the UWP-field-rolling
// primitives (RollStarport, RollSize, ...) both mainworld generation here
// and secondary-world generation in package system are built from.
// Star-system structure (stars, orbits, gas giants, satellites) lives in
// package system; sector-scale generation lives in package sector.
package world

import "strings"

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

// JoinOrNone space-joins items, or returns "None" if items is empty —
// the shared display convention package render's prose output
// (TradeCodeStrings and BaseStrings alike) uses so an empty field reads
// as an explicit "None" rather than silently going blank. The Second
// Survey Format table renderers (render.SectorCompact, cmd/client's
// printSectorHex) use BasesOrDash for Bases instead — see its own doc
// comment for why a table needs a different placeholder than prose does.
func JoinOrNone(items []string) string {
	if len(items) == 0 {
		return "None"
	}

	return strings.Join(items, " ")
}

// OrDash returns s, or "-" if s is empty — the same placeholder Book 3
// p.21's own Second Survey Format sector-table example uses for an
// absent field ("BcCeF NS - 703 8": a bare "-" where Regina's Travel
// Zone column would otherwise go).
func OrDash(s string) string {
	if s == "" {
		return "-"
	}

	return s
}

// BasesOrDash is OrDash for a Base slice: space-joins BaseStrings(bases),
// or "-" if there are none — the Second Survey Format table's own
// convention (see OrDash), as opposed to JoinOrNone's "None" for prose.
func BasesOrDash(bases []Base) string {
	return OrDash(strings.Join(BaseStrings(bases), " "))
}

// TravelZoneOrDash is OrDash for an already-stringified TravelZone
// (TravelZone.String(), or the same string decoded off the wire), with
// one addition: a Green zone also collapses to "-". Book 3 p.21's Second
// Survey Format table shows exactly this for Regina — a Green-zone
// world — despite Green.String() itself reading "Green" everywhere else
// (e.g. package render's prose output, where spelling it out is correct
// and this helper doesn't apply).
func TravelZoneOrDash(s string) string {
	if s == ZoneGreen.String() {
		return "-"
	}

	return OrDash(s)
}
