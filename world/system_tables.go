package world

import "github.com/philoserf/traveller/dice"

// starPresenceFlux is the threshold a Flux roll must meet for a Close,
// Near, or Far star (or a Companion to any star already present) to
// exist. Book 3 p.28, Table 1: "Flux for Close, Near, and Far stars in
// the system. Flux for Companions for each Star present." — read here as
// one independent Flux roll per candidate position, each checked against
// this same threshold, rather than a single roll deciding all three at
// once (matching the page 21 narrative's "up to eight stars: a Primary
// and a Companion, a Close star... and its Companion, ...", i.e. up to
// two stars — itself and one Companion — per one of four positions).
const starPresenceFlux = 3

// rollSpectralType rolls a star's SpectralType from flux, per Book 3 p.28
// Table 2's "Sp" column. The book prints "OB" ("Select further between O
// or B") as its own row at flux<=-6, with "A" starting at flux<=-5 — but
// the flux values this project ever passes in bottom out at -5 (a
// Primary's raw Flux(), range -5..+5; every other star's is that same
// range plus a non-negative 1D-1 offset), so a literal "<=-6" threshold
// is never reachable and OB-class stars could never be generated at all.
// Reads the row boundary as "-5 or lower" instead — folding the book's
// unreachable -6 row into A's own -5 row — so this branch (and
// SpectralO/SpectralB) are actually reachable, at the cost of narrowing
// A to just flux==-4 alone rather than its literal two-row width.
// Resolved with an extra D6 for O vs B: 1-3 is O, 4-6 is B. Flux>=5 rows
// are "D" or "BD" (white dwarf / brown dwarf) — both collapse to
// SpectralDegenerate, matching that constant's own doc comment
// ("includes brown dwarfs"), a design decision from earlier in this
// project, not new here.
func rollSpectralType(r *dice.Roller, flux int) SpectralType {
	switch {
	case flux <= -5:
		if r.D6() <= 3 {
			return SpectralO
		}

		return SpectralB
	case flux <= -4:
		return SpectralA
	case flux <= -2:
		return SpectralF
	case flux <= 0:
		return SpectralG
	case flux <= 2:
		return SpectralK
	case flux <= 4:
		return SpectralM
	default:
		return SpectralDegenerate
	}
}

// stellarSizeRow is one row of Book 3 p.28 Table 2's luminosity-class
// lookup: given a Flux roll and an already-determined SpectralType, the
// column for that type gives the star's LuminosityClass.
//
// Not enforced here, and documented as a known simplification: the
// table's own footnotes ("Size IV not for K5-K9 and M0-M9. Size VI not
// for A0-A9 and F0-F4") restrict some cells by the star's spectral
// decimal (0-9), which this lookup — keyed only on SpectralType, not the
// decimal — can't express without a second table dimension.
type stellarSizeRow struct {
	Flux                int
	O, B, A, F, G, K, M string
}

var stellarSizeTable = []stellarSizeRow{
	{-6, "Ia", "Ia", "Ia", "II", "II", "II", "II"},
	{-5, "Ia", "Ia", "Ia", "II", "II", "II", "II"},
	{-4, "Ib", "Ib", "Ib", "III", "III", "III", "II"},
	{-3, "II", "II", "II", "IV", "IV", "IV", "II"},
	{-2, "III", "III", "III", "V", "V", "V", "III"},
	{-1, "III", "III", "IV", "V", "V", "V", "V"},
	{0, "III", "III", "V", "V", "V", "V", "V"},
	{1, "V", "III", "V", "V", "V", "V", "V"},
	{2, "V", "V", "V", "V", "V", "V", "V"},
	{3, "V", "V", "V", "V", "V", "V", "V"},
	{4, "IV", "IV", "V", "VI", "VI", "VI", "VI"},
	{5, "D", "D", "D", "D", "D", "D", "D"},
	{6, "IV", "IV", "V", "VI", "VI", "VI", "VI"},
	{7, "IV", "IV", "V", "VI", "VI", "VI", "VI"},
	{8, "IV", "IV", "V", "VI", "VI", "VI", "VI"},
}

// rollLuminosityClass looks up LuminosityClass for an already-determined
// SpectralType at the given flux, clamping flux to the table's -6..+8
// range (a companion's size flux is Primary Flux + 1D+2, which can run
// higher than a single Flux roll's own -5..+5 native range). Degenerate
// stars skip the table entirely ("If Spectral=BD ignore remaining
// rolls") — their LuminosityClass is always "D".
func rollLuminosityClass(flux int, t SpectralType) string {
	if t == SpectralDegenerate {
		return "D"
	}

	switch {
	case flux < -6:
		flux = -6
	case flux > 8:
		flux = 8
	}

	for _, row := range stellarSizeTable {
		if row.Flux != flux {
			continue
		}

		switch t { //nolint:exhaustive // SpectralDegenerate returns above; every other SpectralType value is a table column
		case SpectralO:
			return row.O
		case SpectralB:
			return row.B
		case SpectralA:
			return row.A
		case SpectralF:
			return row.F
		case SpectralG:
			return row.G
		case SpectralK:
			return row.K
		case SpectralM:
			return row.M
		}
	}

	return "V" // unreachable given the switch above covers every non-Degenerate SpectralType
}

// habitableZoneTable is the HZ orbit number by (SpectralType,
// LuminosityClass), Book 3 p.20/29/30 (H1) — identical across all three
// pages, used here as the single source. A combination absent from a
// row (the book's "-" or a blank cell, e.g. O/VI or M/IV) has no entry;
// habitableZoneOrbit reports ok=false for it.
var habitableZoneTable = map[SpectralType]map[string]int{
	SpectralO: {"Ia": 15, "Ib": 15, "II": 14, "III": 13, "IV": 12, "V": 11},
	SpectralB: {"Ia": 13, "Ib": 13, "II": 12, "III": 11, "IV": 10, "V": 9},
	SpectralA: {"Ia": 12, "Ib": 11, "II": 9, "III": 7, "IV": 7, "V": 7},
	SpectralF: {"Ia": 11, "Ib": 10, "II": 9, "III": 6, "IV": 6, "V": 5, "VI": 3},
	SpectralG: {"Ia": 12, "Ib": 10, "II": 9, "III": 7, "IV": 5, "V": 3, "VI": 2},
	SpectralK: {"Ia": 12, "Ib": 10, "II": 9, "III": 8, "IV": 5, "V": 2, "VI": 1},
	SpectralM: {"Ia": 12, "Ib": 11, "II": 10, "III": 9, "V": 0, "VI": 0},
}

// habitableZoneOrbit returns the HZ orbit number for a star's
// SpectralType and LuminosityClass. LuminosityClass "D" (a table row —
// stellarSizeTable's flux+5 row is "D" across every spectral column, so
// this is reachable for any SpectralType, not just SpectralDegenerate)
// isn't in habitableZoneTable — the book's own "D" column is nearly all
// zero (only O/D is 1) regardless of spectral row, and this
// implementation doesn't track which O-M row a size roll landing on "D"
// would otherwise have paired with (rollLuminosityClass discards it), so
// this collapses every "D"-luminosity star to the table's common case, 0
// — a documented simplification, not a table lookup.
func habitableZoneOrbit(t SpectralType, luminosityClass string) (int, bool) {
	if luminosityClass == "D" {
		return 0, true
	}

	row, ok := habitableZoneTable[t]
	if !ok {
		return 0, false
	}

	orbit, ok := row[luminosityClass]

	return orbit, ok
}

// mainworldHZVar returns Table 2B's (Book 3 p.24) HZ offset for placing
// the mainworld, from a Flux roll already adjusted for the caller's DM
// (+2 if primary Spectral M, -2 if O or B, per the table's own notes —
// applied by the caller before this lookup).
func mainworldHZVar(flux int) int {
	switch {
	case flux <= -6:
		return -2
	case flux <= -3:
		return -1
	case flux <= 2:
		return 0
	case flux <= 5:
		return 1
	default:
		return 2
	}
}

// mainworldPlacementKind is whether the mainworld is a freestanding
// Planet or a Satellite of another body.
type mainworldPlacementKind int

// mainworldPlacementKind values.
const (
	mainworldPlanet mainworldPlacementKind = iota
	mainworldCloseSatellite
	mainworldFarSatellite
)

// rollMainworldPlacementKind resolves Table 2C (Book 3 p.24) from a Flux
// roll. The table's extreme rows (flux<=-6, flux>=6) print "(none)" for
// Satellite? — treated here as Planet, the same as the table's other
// non-satellite rows, since "(none)" isn't itself one of the three named
// outcomes (Planet/Close Satellite/Far Satellite) and this project found
// no clarifying text for what else it could mean.
//
// The table's "GG?" column (is a satellite's parent a Gas Giant or a
// plain world) reads "GG" on every row where Satellite? is anything
// other than Planet — so under the simplification above, a mainworld
// that IS a satellite always has a Gas Giant parent; there's no case in
// this table where it wouldn't. This implementation places a freshly
// rolled Gas Giant unconditionally for either satellite kind, matching
// the table exactly rather than modeling a branch the table never takes.
func rollMainworldPlacementKind(flux int) mainworldPlacementKind {
	switch flux {
	case -5, -4:
		return mainworldFarSatellite
	case -3:
		return mainworldCloseSatellite
	default:
		return mainworldPlanet
	}
}

// gasGiantSizeByRoll maps a 2D6 roll to a Gas Giant's Size letter, Book 3
// p.29's GG table. That table's own rows are printed 1-13 against a 2D6
// roll (which only produces 11 distinct values, 2-12) — an indexing
// mismatch this project couldn't resolve from the source text. Modeled
// here as a direct 2D6 lookup over the table's first 11 rows (L through
// W, skipping I/O as ehex digits do), dropping the two largest sizes (X,
// Y — over 10 Jupiter masses) as unreachable; a documented simplification
// given the ambiguity, not a claim that giants of that size can't exist.
var gasGiantSizeByRoll = map[int]byte{
	2: 'L', 3: 'M', 4: 'N',
	5: 'P', 6: 'Q', 7: 'R', 8: 'S', 9: 'T', 10: 'U', 11: 'V', 12: 'W',
}

// rollGasGiant rolls a Gas Giant's Size and Bracket. SGG (Small Gas
// Giant) is 2D6 2-4; LGG (Large Gas Giant) is 2D6 5-12 — matching the
// GG table's own SGG/LGG split, shifted for the same reindexing as
// gasGiantSizeByRoll.
func rollGasGiant(r *dice.Roller) GasGiant {
	roll := r.TwoD6()

	bracket := "LGG"
	if roll <= 4 {
		bracket = "SGG"
	}

	return GasGiant{Size: gasGiantSizeByRoll[roll], Bracket: bracket}
}

// beltOffsetByRoll maps a 2D6 roll to an orbit offset from the system's
// HZ orbit, for placing a mainworld that is itself an Asteroid Belt (Book
// 3 p.29 P2 table, Belt column — "GG and Belt placement is relative to
// HZ"). Same 2D6-vs-13-rows reindexing as gasGiantSizeByRoll; the "HZ"
// cell (2D6=4 here) is offset 0, since it sits exactly where 0 falls in
// the column's numeric sequence.
var beltOffsetByRoll = map[int]int{
	2: -2, 3: -1, 4: 0,
	5: 1, 6: 2, 7: 3, 8: 4, 9: 5, 10: 6, 11: 7, 12: 8,
}

// rollBeltOffset rolls the HZ-relative orbit offset for a mainworld that
// is an Asteroid Belt.
func rollBeltOffset(r *dice.Roller) int {
	return beltOffsetByRoll[r.TwoD6()]
}

// orbitAUTable is orbit number -> distance in AU, Book 3 p.20's A1
// Orbital Distances table, indexed by orbit number (0-20).
var orbitAUTable = []float64{
	0.2, 0.4, 0.7, 1.0, 1.6, 2.8, 5.2, 10, 20, 40, 77,
	154, 308, 615, 1230, 2500, 4900, 9800, 19500, 39500, 78700,
}

// orbitAU returns the AU distance for a numbered orbit, or 0 if number is
// outside the table's 0-20 range (every orbit this generator places falls
// within that range, so this is a defensive fallback, not a real case).
func orbitAU(number int) float64 {
	if number < 0 || number >= len(orbitAUTable) {
		return 0
	}

	return orbitAUTable[number]
}

// lggOffsetByRoll, sggOffsetByRoll, igOffsetByRoll: Book 3 p.29's P2
// Basic Placement Chart LGG/SGG/IG columns — an orbit offset from the
// system's HZ orbit ("GG and Belt placement is relative to HZ"). Same
// 2D6-vs-13-rows reindexing as beltOffsetByRoll/gasGiantSizeByRoll: read
// directly against the table's first 11 rows.
var lggOffsetByRoll = map[int]int{
	2: -4, 3: -3, 4: -2, 5: -1, 6: 0, 7: 1, 8: 2, 9: 3, 10: 4, 11: 5, 12: 6,
}

var sggOffsetByRoll = map[int]int{
	2: -3, 3: -2, 4: -1, 5: 0, 6: 1, 7: 2, 8: 3, 9: 4, 10: 5, 11: 6, 12: 7,
}

var igOffsetByRoll = map[int]int{
	2: 0, 3: 1, 4: 2, 5: 3, 6: 4, 7: 5, 8: 6, 9: 7, 10: 8, 11: 9, 12: 10,
}

func rollLGGOffset(r *dice.Roller) int { return lggOffsetByRoll[r.TwoD6()] }
func rollSGGOffset(r *dice.Roller) int { return sggOffsetByRoll[r.TwoD6()] }
func rollIGOffset(r *dice.Roller) int  { return igOffsetByRoll[r.TwoD6()] }

// world1OrbitByRoll, world2OrbitByRoll: P2's World1/World2 columns —
// literal orbit numbers (not HZ-relative, unlike every other P2 column:
// "World placement is based on Orbit"), used to place "Other Worlds."
// Same reindexing as the other P2 columns.
var world1OrbitByRoll = map[int]int{
	2: 11, 3: 10, 4: 8, 5: 6, 6: 4, 7: 2, 8: 0, 9: 1, 10: 3, 11: 5, 12: 7,
}

var world2OrbitByRoll = map[int]int{
	2: 18, 3: 17, 4: 16, 5: 15, 6: 14, 7: 13, 8: 12, 9: 11, 10: 10, 11: 9, 12: 8,
}

func rollWorld1Orbit(r *dice.Roller) int { return world1OrbitByRoll[r.TwoD6()] }
func rollWorld2Orbit(r *dice.Roller) int { return world2OrbitByRoll[r.TwoD6()] }

// secondaryWorldCategory is which of Book 3 p.29's eight non-mainworld
// world types a placed "Other World" turns out to be.
type secondaryWorldCategory int

// secondaryWorldCategory values.
const (
	categoryInferno secondaryWorldCategory = iota
	categoryInnerWorld
	categoryBigWorld
	categoryStormWorld
	categoryRadWorld
	categoryHospitable
	categoryWorldlet
	categoryIceworld
)

// innerHZCategoryByRoll, outerCategoryByRoll: Book 3 p.29's Inner/HZ and
// Outer world-category tables (1D each). "Inner" is inside HZ-1, "HZ" is
// HZ-1..HZ+1, "Outer" is beyond HZ+1 — rollSecondaryWorldCategory reads
// this from the caller's orbit-relative-to-HZ delta.
var innerHZCategoryByRoll = map[int]secondaryWorldCategory{
	1: categoryInferno, 2: categoryInnerWorld, 3: categoryBigWorld,
	4: categoryStormWorld, 5: categoryRadWorld, 6: categoryHospitable,
}

var outerCategoryByRoll = map[int]secondaryWorldCategory{
	1: categoryWorldlet, 2: categoryIceworld, 3: categoryBigWorld,
	4: categoryIceworld, 5: categoryRadWorld, 6: categoryIceworld,
}

// rollSecondaryWorldCategory rolls a placed "Other World"'s category.
// delta<=0 covers both "Inner" (delta<=-1) and "HZ" (delta==0) — the book
// gives them a single shared 1D column, not two.
func rollSecondaryWorldCategory(r *dice.Roller, delta int) secondaryWorldCategory {
	if delta <= 0 {
		return innerHZCategoryByRoll[r.D6()]
	}

	return outerCategoryByRoll[r.D6()]
}
