package starship

import "github.com/philoserf/traveller/ehex"

// DriveCategory is the functional role a drive fills aboard ship.
type DriveCategory int

// DriveCategory values. Note these are distinct from, and easily confused
// with, the similarly-named DriveType constants below (e.g. DriveJump here
// is a category; DriveJumpJ there is a specific jump-drive technology).
const (
	DriveManeuver DriveCategory = iota
	DriveJump
	DrivePower
	DriveSupplemental
)

// DriveType is the specific technology used by a drive.
type DriveType string

// DriveType values, grouped by DriveCategory (Maneuver, Interstellar, Power, Supplemental).
const (
	DriveGravitic DriveType = "G" // Maneuver
	DriveRocket   DriveType = "R" // Maneuver
	DriveHEPlaR   DriveType = "H" // Maneuver

	DriveJumpJ DriveType = "J" // Interstellar
	DriveHop   DriveType = "Hop"
	DriveSkip  DriveType = "S"
	DriveNAFAL DriveType = "N"

	DrivePowerPlant DriveType = "P" // Power
	DriveFission    DriveType = "U"
	DriveAntiMatter DriveType = "A"
	DriveCollector  DriveType = "C"

	DriveBattery    DriveType = "B" // Supplemental
	DriveFuelCell   DriveType = "FC"
	DriveFusionPlus DriveType = "F+"
)

// driveCategories is the canonical mapping from every declared
// DriveType to its DriveCategory — the same grouping the const block
// above documents in comments, in a form code can actually read.
//
// It exists as production data rather than as a switch so that "which
// drives are there, and what is each one?" has a single answer that both
// Category and the package's tests read. A switch cannot be enumerated,
// which meant a DriveType added to the const block and nowhere else was
// silently categorised as a maneuver drive by the default arm, with
// nothing able to notice.
var driveCategories = map[DriveType]DriveCategory{
	DriveGravitic: DriveManeuver,
	DriveRocket:   DriveManeuver,
	DriveHEPlaR:   DriveManeuver,

	DriveJumpJ: DriveJump,
	DriveHop:   DriveJump,
	DriveSkip:  DriveJump,
	DriveNAFAL: DriveJump,

	DrivePowerPlant: DrivePower,
	DriveFission:    DrivePower,
	DriveAntiMatter: DrivePower,
	DriveCollector:  DrivePower,

	DriveBattery:    DriveSupplemental,
	DriveFuelCell:   DriveSupplemental,
	DriveFusionPlus: DriveSupplemental,
}

// Category returns the DriveCategory this DriveType belongs to, per
// driveCategories above. An unrecognized DriveType reports DriveManeuver
// — retained from the original switch's own default arm, so behavior is
// unchanged for any value that was already valid. TestEveryDeclaredDriveTypeIsCategorized
// (drive_test.go) is what keeps that fallback from quietly absorbing a
// newly declared drive: it reads the const block from source and fails
// on any DriveType missing from the map above.
func (t DriveType) Category() DriveCategory {
	if category, ok := driveCategories[t]; ok {
		return category
	}

	return DriveManeuver
}

// StageEffect modifies a drive's TL offset, cost multiplier, efficiency,
// fuel use, and tonnage relative to the standard baseline.
type StageEffect int

// StageEffect values, ordered from least refined (Experimental) to most
// refined (Ultimate).
const (
	StageExperimental StageEffect = iota
	StagePrototype
	StageEarly
	StageStandard
	StageBasic
	StageAlternate
	StageImproved
	StageGeneric
	StageModified
	StageAdvanced
	StageUltimate
)

// Drive is one installed drive: maneuver, jump/hop/skip/NAFAL, power plant,
// or a supplemental power source. Its DriveCategory is derivable from Type
// via Type.Category() and so isn't stored separately.
type Drive struct {
	Type         DriveType
	Letter       string // size code, e.g. "A".."Z", "N2".."Z2" for oversize
	Tons         float64
	EnergyPoints int
	Potential    int // performance rating: Maneuver-Potential-N = N Gs, Jump-Potential-N = Jump-N
	TechLevel    ehex.Value
	Stage        StageEffect
	Efficiency   int     // percent
	Cost         float64 // MCr
}
