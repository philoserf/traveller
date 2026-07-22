package world

// SpectralType is a star's spectral classification.
type SpectralType byte

const (
	SpectralO          SpectralType = 'O'
	SpectralB          SpectralType = 'B'
	SpectralA          SpectralType = 'A'
	SpectralF          SpectralType = 'F'
	SpectralG          SpectralType = 'G'
	SpectralK          SpectralType = 'K'
	SpectralM          SpectralType = 'M'
	SpectralDegenerate SpectralType = 'D' // includes brown dwarfs
)

// StellarRole is a star's position within a multiple-star system.
type StellarRole int

const (
	Primary StellarRole = iota
	Close
	Near
	Far
)

// Star is a single star in a system, e.g. "F7 V".
type Star struct {
	SpectralType       SpectralType
	SpectralDecimal    int    // 0-9, ignored when SpectralType is SpectralDegenerate
	LuminosityClass    string // Ia, Ib, II, III, IV, V, VI, D
	Role               StellarRole
	Companion          *Star
	HabitableZoneOrbit int
}

// Orbit is a single numbered orbit slot within a system.
type Orbit struct {
	Number int
	AU     float64
	Star   *Star
	World  *World
}

// StarSystem is a full system: its stars, orbits, and worlds.
type StarSystem struct {
	Sector         string
	Hex            string
	Stars          []Star
	Orbits         []Orbit
	Worlds         []World
	MainworldIndex int
}
