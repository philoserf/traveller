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
// the shared display convention every renderer of TradeCodeStrings/
// BaseStrings output uses (package render's Markdown, cmd/client's
// terminal output) so an empty field reads as an explicit "None" rather
// than silently going blank.
func JoinOrNone(items []string) string {
	if len(items) == 0 {
		return "None"
	}

	return strings.Join(items, " ")
}

// OrDash returns s, or "—" if s is empty — the shared placeholder every
// renderer of an optional already-stringified field (e.g. TravelZone's
// String(), which is "" for the common no-zone case) uses instead of
// silently leaving the field blank.
func OrDash(s string) string {
	if s == "" {
		return "—"
	}

	return s
}
