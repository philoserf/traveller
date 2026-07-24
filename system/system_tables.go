package system

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
// The table's own footnotes ("Size IV not for K5-K9 and M0-M9. Size VI
// not for A0-A9 and F0-F4") restrict some cells by the star's spectral
// decimal (0-9), which this table — keyed only on SpectralType, not the
// decimal — can't express directly. The book gives no reroll or
// substitution instruction for a forbidden result, so
// rollLuminosityClass enforces the footnote itself after this table's
// own lookup, clamping a forbidden result to "V" (the neighboring
// main-sequence class) — see violatesLuminosityFootnote.
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

// violatesLuminosityFootnote reports whether stellarSizeTable's own
// margin note ("Size IV not for K5-K9 and M0-M9. Size VI not for A0-A9
// and F0-F4") forbids pairing SpectralType t at the given SpectralDecimal
// with LuminosityClass class. Both named ranges are contiguous in
// spectralOrdinal's own rank*10+decimal encoding (K5-K9 is 55-59, M0-M9
// is 60-69, together 55-69; A0-A9 is 20-29, F0-F4 is 30-34, together
// 20-34) — reusing it here, the same ordering precludedOrbitCeiling
// already keys its own range checks on, instead of a second, independent
// hand-rolled encoding of spectral order. Checked for the full named
// ranges even though only K5-K9/IV and F0-F4/VI are actually reachable
// given stellarSizeTable's own contents today (M never rolls IV and A
// never rolls VI at any Flux row) — so a future correction to that table
// can't silently reintroduce an uncaught violation.
func violatesLuminosityFootnote(t SpectralType, decimal int, class string) bool {
	ordinal := spectralOrdinal(t, decimal)

	switch class {
	case "IV":
		return ordinal >= 55 // K5-K9, M0-M9
	case "VI":
		return ordinal >= 20 && ordinal <= 34 // A0-A9, F0-F4
	default:
		return false
	}
}

// rollLuminosityClass looks up LuminosityClass for an already-determined
// SpectralType and SpectralDecimal at the given flux, clamping flux to
// the table's -6..+8 range (a companion's size flux is Primary Flux +
// 1D+2, which can run higher than a single Flux roll's own -5..+5 native
// range). Degenerate stars skip the table entirely ("If Spectral=BD
// ignore remaining rolls") — their LuminosityClass is always "D". A
// table result the footnote forbids for this spectral/decimal
// combination (violatesLuminosityFootnote) clamps to "V" instead — see
// stellarSizeRow's own doc comment for why "V" and not a reroll.
func rollLuminosityClass(flux int, t SpectralType, decimal int) string {
	if t == SpectralDegenerate {
		return "D"
	}

	switch {
	case flux < -6:
		flux = -6
	case flux > 8:
		flux = 8
	}

	class := "V" // unreachable given the switch below covers every non-Degenerate SpectralType

	for _, row := range stellarSizeTable {
		if row.Flux != flux {
			continue
		}

		switch t { //nolint:exhaustive // SpectralDegenerate returns above; every other SpectralType value is a table column
		case SpectralO:
			class = row.O
		case SpectralB:
			class = row.B
		case SpectralA:
			class = row.A
		case SpectralF:
			class = row.F
		case SpectralG:
			class = row.G
		case SpectralK:
			class = row.K
		case SpectralM:
			class = row.M
		}

		break
	}

	if violatesLuminosityFootnote(t, decimal, class) {
		return "V"
	}

	return class
}

// habitableZoneTable is the HZ orbit number by (SpectralType,
// LuminosityClass), Book 3 p.20/29/30 (H1/J1/K1) — byte-identical across
// all three pages, used here as the single source. Includes the table's
// own "D" (Degenerate luminosity) column: 1 for O, 0 for every other
// spectral type — reachable for any SpectralType whose independent size
// roll landed on "D" (stellarSizeTable's flux+5 row is "D" across every
// spectral column), not just the SpectralDegenerate star *type*. A
// combination absent from a row (the book's "-" or a blank cell, e.g.
// O/VI or M/IV) has no entry; habitableZoneOrbit reports ok=false for it.
var habitableZoneTable = map[SpectralType]map[string]int{
	SpectralO: {"Ia": 15, "Ib": 15, "II": 14, "III": 13, "IV": 12, "V": 11, "D": 1},
	SpectralB: {"Ia": 13, "Ib": 13, "II": 12, "III": 11, "IV": 10, "V": 9, "D": 0},
	SpectralA: {"Ia": 12, "Ib": 11, "II": 9, "III": 7, "IV": 7, "V": 7, "D": 0},
	SpectralF: {"Ia": 11, "Ib": 10, "II": 9, "III": 6, "IV": 6, "V": 5, "VI": 3, "D": 0},
	SpectralG: {"Ia": 12, "Ib": 10, "II": 9, "III": 7, "IV": 5, "V": 3, "VI": 2, "D": 0},
	SpectralK: {"Ia": 12, "Ib": 10, "II": 9, "III": 8, "IV": 5, "V": 2, "VI": 1, "D": 0},
	SpectralM: {"Ia": 12, "Ib": 11, "II": 10, "III": 9, "V": 0, "VI": 0, "D": 0},
}

// habitableZoneOrbit returns the HZ orbit number for a star's
// SpectralType and LuminosityClass, straight from habitableZoneTable.
// The SpectralDegenerate star *type* (as opposed to a normal type whose
// independent size roll landed on "D" luminosity) has no row at all —
// ok=false, leaving the caller's HabitableZoneOrbit at its zero value —
// since Book 3's H1/J1/K1 tables have no O-M spectral column to key a
// row on for a star that's Degenerate from the type roll onward.
func habitableZoneOrbit(t SpectralType, luminosityClass string) (int, bool) {
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

// rollMainworldPlacementKind resolves Table "2C SATELLITE?" (Book 3
// p.24) from a Flux roll. The table's extreme rows (flux=-6, flux=+6)
// leave the Satellite? cell genuinely blank in the book — not "(none)"
// — reachable only via a Referee-imposed DM on Flux (p.24's own 2A
// note: "Flux= can equal -6 or +6 with Referee imposed DM"). This
// project's own Flux() is DM-free (dice.go: D6()-D6(), range -5..+5),
// so those rows never actually fire here; the default->Planet branch
// below is a harmless unreachable catch-all for them, not a lossy
// coercion of a real outcome. Separately, "no mainworld" was never a
// coherent reading of a blank cell here regardless: the mainworld
// always already exists (world.Generate runs before GenerateSystem ever
// calls this) — Table 2C only decides where/how it sits, never whether
// it exists at all.
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

// gasGiantSizeByRoll maps a 2D6 roll directly to a Gas Giant's Size
// letter, Book 3 p.29's GG table — verified against the page image: the
// table's row number IS the roll value (row 7 = roll 7 = Size S,
// matching the Regina worked example's "GG Table = 2D = 7 = Siz S"
// exactly), not an off-by-one index as an earlier reading of this table
// assumed. Row 1 (Size L) is unreachable via a plain 2D6 roll (minimum
// 2); row 13 (Size Y) is reserved for the page's own separate "All BD
// Brown Dwarfs are Siz=Y" rule, not a plain roll outcome either.
var gasGiantSizeByRoll = map[int]byte{
	2: 'M', 3: 'N', 4: 'P', 5: 'Q', 6: 'R',
	7: 'S', 8: 'T', 9: 'U', 10: 'V', 11: 'W', 12: 'X',
}

// rollGasGiant rolls a Gas Giant's Size and Bracket. The GG table's own
// left-margin bracket divides SGG (Small Gas Giant) from LGG (Large Gas
// Giant) between rows 6 and 7 — SGG for 2D6 2-6, LGG for 7-12 — verified
// against the page image, and cross-checked against the Regina worked
// example (2D=7/Size S explicitly labeled "LGG"; 2D=2/Size M labeled
// "Small Gas Giant SGG"), which only agrees with the bracket read this
// way.
func rollGasGiant(r *dice.Roller) GasGiant {
	roll := r.TwoD6()

	bracket := "LGG"
	if roll <= 6 {
		bracket = "SGG"
	}

	return GasGiant{Size: gasGiantSizeByRoll[roll], Bracket: bracket}
}

// beltOffsetByRoll maps a 2D6 roll directly to an orbit offset from the
// system's HZ orbit, for placing a mainworld that is itself an Asteroid
// Belt (Book 3 p.29 P2 "Basic Placement Chart", Belt column — "GG and
// Belt placement is relative to HZ"). Verified against the page image:
// the table's row number IS the roll value, the same direct indexing
// gasGiantSizeByRoll uses — an earlier reading of every P2 column (and
// of the GG table before it) wrongly shifted each one by one row/roll.
// Row 1 (offset -2) is unreachable via a plain 2D6 roll (minimum 2),
// same as the GG table's row 1.
var beltOffsetByRoll = map[int]int{
	2: -1, 3: 0, 4: 1,
	5: 2, 6: 3, 7: 4, 8: 5, 9: 6, 10: 7, 11: 8, 12: 9,
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
// direct row=roll indexing as beltOffsetByRoll/gasGiantSizeByRoll,
// verified against the page image.
var lggOffsetByRoll = map[int]int{
	2: -3, 3: -2, 4: -1, 5: 0, 6: 1, 7: 2, 8: 3, 9: 4, 10: 5, 11: 6, 12: 7,
}

var sggOffsetByRoll = map[int]int{
	2: -2, 3: -1, 4: 0, 5: 1, 6: 2, 7: 3, 8: 4, 9: 5, 10: 6, 11: 7, 12: 8,
}

var igOffsetByRoll = map[int]int{
	2: 1, 3: 2, 4: 3, 5: 4, 6: 5, 7: 6, 8: 7, 9: 8, 10: 9, 11: 10, 12: 11,
}

func rollLGGOffset(r *dice.Roller) int { return lggOffsetByRoll[r.TwoD6()] }
func rollSGGOffset(r *dice.Roller) int { return sggOffsetByRoll[r.TwoD6()] }
func rollIGOffset(r *dice.Roller) int  { return igOffsetByRoll[r.TwoD6()] }

// world1OrbitByRoll, world2OrbitByRoll: P2's World1/World2 columns —
// literal orbit numbers (not HZ-relative, unlike every other P2 column:
// "World placement is based on Orbit"), used to place "Other Worlds."
// Same direct row=roll indexing as the other P2 columns, verified
// against the page image — including World1's own real shape, a V
// bottoming out at orbit 1 for roll 7 (11,10,8,6,4,2,1,3,5,7,9,9), not
// the smoother, one-roll-earlier valley an earlier reading produced.
var world1OrbitByRoll = map[int]int{
	2: 10, 3: 8, 4: 6, 5: 4, 6: 2, 7: 1, 8: 3, 9: 5, 10: 7, 11: 9, 12: 9,
}

var world2OrbitByRoll = map[int]int{
	2: 17, 3: 16, 4: 15, 5: 14, 6: 13, 7: 11, 8: 10, 9: 9, 10: 8, 11: 7, 12: 7,
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

// spectralRank orders SpectralType hottest-to-coolest: O=0 .. M=6,
// matching the type's own doc comment ordering. SpectralDegenerate never
// reaches the Stellar Surface table below (see precludedOrbitCeiling,
// which rejects every luminosity class Degenerate stars actually have
// before ever calling this) — the fallback here is defensive, not a real
// case.
func spectralRank(t SpectralType) int {
	switch t {
	case SpectralO:
		return 0
	case SpectralB:
		return 1
	case SpectralA:
		return 2
	case SpectralF:
		return 3
	case SpectralG:
		return 4
	case SpectralK:
		return 5
	case SpectralM:
		return 6
	default:
		return 7
	}
}

// spectralOrdinal combines rank and decimal into one comparable value
// (rank*10 + decimal, range 0-69) for range-matching against the Stellar
// Surface table's printed spectral ranges (e.g. "A5-G0").
func spectralOrdinal(t SpectralType, decimal int) int {
	return spectralRank(t)*10 + decimal
}

// stellarSurfaceEntry is one row of Book 3 p.20's "Stellar Surface"
// table for a single luminosity-class column: UpperOrdinal is the
// spectralOrdinal of the coolest (latest-decimal) spectral type whose
// photosphere still reaches Orbit. Rows are sorted by Orbit ascending —
// precludedOrbitCeiling finds the first (innermost) row whose
// UpperOrdinal is >= a star's own ordinal.
type stellarSurfaceEntry struct {
	UpperOrdinal int
	Orbit        int
}

// stellarSurfaceIa/Ib/II/III: Book 3 p.20's four Stellar Surface columns.
// The printed ranges (e.g. Ib row2 = "A5-G0") become one entry per row,
// UpperOrdinal set from the range's cooler (higher-ordinal) end. Rows the
// book leaves blank for a column (e.g. Ia has no rows 0-3) simply have no
// entry — "first row whose UpperOrdinal is >= ordinal" naturally rounds
// an ordinal that falls in a blank gap up to the next tabulated
// (cooler, farther-out) row, and clamps an ordinal hotter than a
// column's very first row to that row — see the sysgen precluded-orbit
// plan for why "round toward more precluded, not less" is the right
// direction to be wrong in, matching this project's other tables.
var stellarSurfaceIa = []stellarSurfaceEntry{
	{35, 4}, {40, 5}, {50, 6}, {55, 7}, {60, 8}, {69, 9},
}

var stellarSurfaceIb = []stellarSurfaceEntry{
	{20, 1}, {40, 2}, {45, 4}, {50, 5}, {55, 6}, {60, 7}, {69, 8},
}

var stellarSurfaceII = []stellarSurfaceEntry{
	{35, 0}, {45, 1}, {50, 2}, {55, 4}, {60, 5}, {65, 6}, {69, 7},
}

var stellarSurfaceIII = []stellarSurfaceEntry{
	{50, 0}, {55, 1}, {60, 2}, {65, 5}, {69, 6},
}

// precludedOrbitCeiling returns the highest orbit number a star of
// SpectralType t, SpectralDecimal decimal, and LuminosityClass
// luminosityClass physically engulfs with its own photosphere ("Book 3
// p.21: Some stars are so large that they engulf some of the orbits in
// the system"), and whether any preclusion applies at all. ok is false
// for every luminosity class besides Ia/Ib/II/III — Book 3's Stellar
// Surface table has no column for IV/V/VI/D, since main-sequence and
// smaller stars are never physically large enough to reach even orbit 0.
func precludedOrbitCeiling(t SpectralType, decimal int, luminosityClass string) (int, bool) {
	var table []stellarSurfaceEntry

	switch luminosityClass {
	case "Ia":
		table = stellarSurfaceIa
	case "Ib":
		table = stellarSurfaceIb
	case "II":
		table = stellarSurfaceII
	case "III":
		table = stellarSurfaceIII
	default:
		return 0, false
	}

	ordinal := spectralOrdinal(t, decimal)

	for _, entry := range table {
		if ordinal <= entry.UpperOrdinal {
			return entry.Orbit, true
		}
	}

	// Unreachable given every populated column's last row already covers
	// ordinal 69, the maximum possible (M9) — defensive, not a real case.
	return 0, false
}
