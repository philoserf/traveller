package starship

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
// or a supplemental power source.
type Drive struct {
	Category     DriveCategory
	Type         DriveType
	Letter       string // size code, e.g. "A".."Z"
	Tons         float64
	EnergyPoints int
	Potential    int // performance rating: Maneuver-Potential-N = N Gs, Jump-Potential-N = Jump-N
	TechLevel    int
	Stage        StageEffect
	Efficiency   int     // percent
	Cost         float64 // MCr
}
