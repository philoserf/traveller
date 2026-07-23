package world

// SpectralType is a star's spectral classification.
type SpectralType byte

// SpectralType values, ordered hottest to coolest per the standard OBAFGKM
// sequence; SpectralDegenerate is the odd one out, covering white dwarfs
// and brown dwarfs rather than a point on that temperature scale.
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

// StellarRole values, ordered by increasing distance from the system's
// center: Close stars orbit nearest the primary, Near and Far further out.
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

// GasGiant is a gas giant occupying an orbit — its own kind of body, not a
// UWP World (Book 3's GG table gives it a Size and Bracket only, no
// Atmosphere/Hydrographics/Population/...).
type GasGiant struct {
	Size    byte   // 'L'..'Y', per the GG table (Book 3 p.29)
	Bracket string // "SGG" (Small Gas Giant) or "LGG" (Large Gas Giant)
}

// Orbit is a single numbered orbit slot within a system. Number may repeat:
// a satellite shares its parent body's Number and sets Satellite to true,
// distinguishing "orbits the star at slot N" from "orbits whatever
// occupies slot N" — e.g. a Gas Giant's moon, or (per Book 3's "G Placing
// Worlds" narrative) a mainworld that is itself a satellite of a Gas
// Giant. AU is left zero for a Satellite entry: the orbit-to-AU table
// (Book 3 p.20) only covers primary numbered orbits, not sub-orbit
// distances.
type Orbit struct {
	Number    int
	Satellite bool
	AU        float64
	Star      *Star
	GasGiant  *GasGiant
	World     *World
}

// StarSystem is a full system. Orbits is the single source of truth for
// its stars and worlds; use the Stars and Worlds methods to collect them.
type StarSystem struct {
	Sector         string
	Hex            string
	Orbits         []Orbit
	MainworldOrbit int // index into Orbits; Orbits[MainworldOrbit].World is the mainworld
}

// Stars returns every star in the system, collected from Orbits.
func (s StarSystem) Stars() []*Star {
	var stars []*Star

	for i := range s.Orbits {
		if s.Orbits[i].Star != nil {
			stars = append(stars, s.Orbits[i].Star)
		}
	}

	return stars
}

// Worlds returns every world in the system, collected from Orbits.
func (s StarSystem) Worlds() []*World {
	var worlds []*World

	for i := range s.Orbits {
		if s.Orbits[i].World != nil {
			worlds = append(worlds, s.Orbits[i].World)
		}
	}

	return worlds
}
