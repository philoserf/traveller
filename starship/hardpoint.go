package starship

import "github.com/philoserf/traveller/ehex"

// Mount is the physical installation a weapon or other item occupies.
type Mount string

// Mount values, ordered from smallest to largest tonnage footprint
// (turrets take ~1 ton, barbettes ~3-5 tons, bays ~50 tons).
const (
	MountTurretSingle Mount = "TurretSingle"
	MountTurretDual   Mount = "TurretDual"
	MountTurretTriple Mount = "TurretTriple"
	MountTurretQuad   Mount = "TurretQuad"
	MountBarbette     Mount = "Barbette"
	MountDualBarbette Mount = "DualBarbette"
	MountBay          Mount = "Bay"
	MountLargeBay     Mount = "LargeBay"
	MountMain         Mount = "Main"
)

// Hardpoint is one of a hull's weapon/installation slots (HullTons/100 of
// them are available).
type Hardpoint struct {
	Number       int
	Mount        Mount
	WeaponCode   string
	TechLevel    ehex.Value
	Cost         float64 // MCr
	Firmpoints   []string
	SurfaceMount bool // mounted on hull surface, doesn't consume a hardpoint
}

// SensorType is the functional category of a sensor installation.
type SensorType string

// SensorType values, by target: CommVisual/Space detect other ships and
// objects, World scans planetary/system bodies, Specialized covers
// purpose-built roles, Deception is a countermeasure, not detection.
const (
	SensorCommVisual  SensorType = "CommVisual"
	SensorSpace       SensorType = "Space"
	SensorWorld       SensorType = "World"
	SensorSpecialized SensorType = "Specialized"
	SensorDeception   SensorType = "Deception"
)

// Sensor is one shipboard sensor installation.
type Sensor struct {
	Type   SensorType
	Range  int // R= rating
	Signal int // S= rating
	Mount  string
}
