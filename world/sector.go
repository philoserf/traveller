package world

// Density is one of Book 3 p.13's eight named System Presence densities,
// each a dice-count + "N or less" target for whether a given hex holds a
// star system.
type Density int

// Density values. DensityStandard is the default — identical to the
// "Classic System Contents Table"'s own System Presence figure (1D<=3,
// 50%).
const (
	DensityExtraGalactic Density = iota
	DensityRift
	DensitySparse
	DensityScattered
	DensityStandard
	DensityDense
	DensityCluster
	DensityCore
)

// String returns the density's book name (e.g. "Standard"), or "Unknown"
// for any other value.
func (d Density) String() string {
	switch d {
	case DensityExtraGalactic:
		return "Extra Galactic"
	case DensityRift:
		return "Rift"
	case DensitySparse:
		return "Sparse"
	case DensityScattered:
		return "Scattered"
	case DensityStandard:
		return "Standard"
	case DensityDense:
		return "Dense"
	case DensityCluster:
		return "Cluster"
	case DensityCore:
		return "Core"
	default:
		return "Unknown"
	}
}

// densityNames maps every Density's own String() back to its value — the
// single source both ParseDensity and String draw from, so the two can't
// drift out of sync with each other.
var densityNames = map[string]Density{
	DensityExtraGalactic.String(): DensityExtraGalactic,
	DensityRift.String():          DensityRift,
	DensitySparse.String():        DensitySparse,
	DensityScattered.String():     DensityScattered,
	DensityStandard.String():      DensityStandard,
	DensityDense.String():         DensityDense,
	DensityCluster.String():       DensityCluster,
	DensityCore.String():          DensityCore,
}

// ParseDensity parses name (one of Density's own String() values, e.g.
// "Standard") back into a Density. ok is false for any other string.
func ParseDensity(name string) (Density, bool) {
	d, ok := densityNames[name]

	return d, ok
}

// densityRoll is one Density's dice count and "N or less" target (Book 3
// p.13's System Presence table). Core's own printed figures are
// internally inconsistent — "11 or less (on 2D) 91%" — verified directly
// against the page image, not just text extraction: 2D6<=11 is
// mathematically 35/36 (97.2%), excluding only a roll of 12, not 91%.
// The dice mechanic ("11 or less on 2D") is the unambiguous, literal
// rule and is what's implemented here; the book's own quoted percentage
// is simply wrong and isn't reproduced anywhere in this code.
type densityRoll struct {
	dice   int
	target int
}

var densityTable = map[Density]densityRoll{
	DensityExtraGalactic: {dice: 3, target: 3},
	DensityRift:          {dice: 2, target: 2},
	DensitySparse:        {dice: 1, target: 1},
	DensityScattered:     {dice: 1, target: 2},
	DensityStandard:      {dice: 1, target: 3},
	DensityDense:         {dice: 1, target: 4},
	DensityCluster:       {dice: 1, target: 5},
	DensityCore:          {dice: 2, target: 11},
}

// Hex is one located slot in a Sector's grid. System is nil for an empty
// (deep space) hex.
type Hex struct {
	Location string // "CCRR", e.g. "0101".."3240"
	System   *StarSystem
}

// Sector is a full 32x40 (1280-hex) grid, per Book 3 p.12-13. Hexes is
// always exactly 1280 entries, ordered column-major (column 01's 40
// rows, then column 02's, ...) matching the book's own CCRR numbering —
// the hex in the upper-left corner is "0101", lower-right is "3240".
type Sector struct {
	Name  string
	Hexes []Hex
}

// subsectorWidth/subsectorHeight: a Subsector is 8 columns x 10 rows of
// hexes (Book 3 p.12: "each containing 80 locations: 8 columns of 10
// rows"). sectorWidth/sectorHeight are the full Sector's own 32x40 grid,
// i.e. 4x4 Subsectors.
const (
	subsectorWidth  = 8
	subsectorHeight = 10
	sectorWidth     = 4 * subsectorWidth
	sectorHeight    = 4 * subsectorHeight
)

// ValidSubsectorLetter reports whether letter is a valid Subsector
// identifier (A-P, Book 3 p.15) — the single source Subsector itself,
// api/sectors.go, and cmd/secgen all check against, so the valid range
// can't drift between them.
func ValidSubsectorLetter(letter byte) bool {
	return letter >= 'A' && letter <= 'P'
}

// Subsector returns the 80 Hexes (8 columns x 10 rows) belonging to
// letter — A-P, laid out in Book 3 p.15's own 4x4 lettered grid (A B C D
// / E F G H / I J K L / M N O P, reading left-to-right then top-to-bottom)
// — as a slice of Sector's own Hexes. Not a separate generation path:
// "Subsector Map creation uses the same procedures as creating a Sector
// Map," just at a smaller, sliced scale. Returns nil if letter fails
// ValidSubsectorLetter.
func (s Sector) Subsector(letter byte) []Hex {
	if !ValidSubsectorLetter(letter) {
		return nil
	}

	index := int(letter - 'A')
	subCol := index % 4
	subRow := index / 4

	var hexes []Hex

	for col := subCol * subsectorWidth; col < (subCol+1)*subsectorWidth; col++ {
		start := col*sectorHeight + subRow*subsectorHeight
		hexes = append(hexes, s.Hexes[start:start+subsectorHeight]...)
	}

	return hexes
}
