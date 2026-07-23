// Package system models a Traveller5 star system: stars, orbits, gas
// giants, satellites, and their generation (GenerateSystem) — everything
// Book 3's system-generation tables place around an already-generated
// mainworld (package world). Sector-scale generation lives in package
// sector.
package system

import (
	"sort"

	"github.com/philoserf/traveller/world"
)

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

// String returns the role's display name (Primary/Close/Near/Far), or
// "Unknown" for any other value. Unlike TravelZone.String()'s "" default
// (where the zero value is itself the common, expected "not set yet"
// case), StellarRole's zero value is Primary — a meaningfully valid role
// — so reaching this default at all means a Star ended up with a role
// outside the four known constants, worth a visible marker rather than a
// silently blank label.
func (role StellarRole) String() string {
	switch role {
	case Primary:
		return "Primary"
	case Close:
		return "Close"
	case Near:
		return "Near"
	case Far:
		return "Far"
	default:
		return "Unknown"
	}
}

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
	Size byte // 'L'..'Y', per the GG table (Book 3 p.29)
	// Bracket is "SGG" (Small Gas Giant), "LGG" (Large Gas Giant), or
	// "IG" (Ice Giant — every second SGG rolled converts to one, same
	// Size, per the GG table's own note).
	Bracket string
	// Ring is Book 3 p.29's "S Number of Satellites": a satellite-count
	// roll of exactly 0 gives the parent a Ring (and rerolls the count
	// once more).
	Ring bool
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
	// Close is meaningful only when Satellite is true: "Close" (2D<=7,
	// tidally locked to its parent) vs "Far" (2D>=8) — Book 3 p.21/24.
	Close bool
	AU    float64
	// HostHZOrbit is the HabitableZoneOrbit of whichever star actually
	// placed this body (the Primary for the mainworld and Star itself;
	// whichever host placeGasGiants/placeBelts/placeOtherWorlds rotated
	// to for everything else). Meaningless (zero) for a Star entry
	// itself. Recorded at placement time rather than reconstructed later
	// — a body's own orbit Number alone doesn't reliably identify which
	// host star placed it, especially in a multi-star system.
	HostHZOrbit int
	// HostRole is the StellarRole of whichever star actually placed this
	// body — unambiguous, unlike HostHZOrbit (distinct stars can share
	// the same HabitableZoneOrbit). Meaningless (Primary, the zero
	// value) for a Star entry itself or a Satellite entry.
	HostRole StellarRole
	Star     *Star
	GasGiant *GasGiant
	World    *world.World
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

// GasGiantAt returns the Gas Giant occupying orbit number, or nil if none
// does. For finding the parent of a satellite sharing that Number — see
// Orbit's doc comment on why a satellite shares its parent's Number
// instead of a distinct one.
func (s StarSystem) GasGiantAt(number int) *GasGiant {
	for i := range s.Orbits {
		if s.Orbits[i].Number == number && s.Orbits[i].GasGiant != nil {
			return s.Orbits[i].GasGiant
		}
	}

	return nil
}

// Worlds returns every world in the system, collected from Orbits.
func (s StarSystem) Worlds() []*world.World {
	var worlds []*world.World

	for i := range s.Orbits {
		if s.Orbits[i].World != nil {
			worlds = append(worlds, s.Orbits[i].World)
		}
	}

	return worlds
}

// IsMainworld reports whether o is the system's own mainworld — either as
// a top-level body (in SystemBodies's bodiesByRole) or as one of a Gas
// Giant's satellites (in satellitesOf; a mainworld can itself be a
// satellite). Compares by World pointer identity, not by Number (which a
// satellite intentionally shares with its parent) or index (already lost
// once SystemBodies groups Orbits into its maps). Guards against a nil
// World on either side — s.Orbits[s.MainworldOrbit].World should never be
// nil per StarSystem's own doc comment, but if it ever were, every Gas
// Giant orbit (World also nil) would otherwise compare equal to it too.
func (s StarSystem) IsMainworld(o Orbit) bool {
	mw := s.Orbits[s.MainworldOrbit].World

	return mw != nil && o.World == mw
}

// SystemBodies groups every Orbit in s besides every Star (starOrbits
// collects those separately, sorted by StellarRole — Primary, then
// Close/Near/Far, the same close-to-far ordering the role constants
// themselves are declared in; Orbit.Number can't be the sort key here
// since it's a sentinel, not a real orbit slot, for the Primary) into:
// bodiesByRole, every top-level (non-Satellite) Gas Giant/World grouped
// by the StellarRole that hosts it (Orbit.HostRole) and sorted by orbit
// Number within each group; and satellitesOf, grouped by the Number they
// share with their parent and sorted Close before Far — the same
// close-to-far ordering applied one level up. The mainworld's own Orbit
// isn't special-cased: it flows through the same two buckets as
// everything else (bodiesByRole if freestanding, satellitesOf if it's
// itself a satellite of a Gas Giant) — callers wanting to point it out
// distinctly should use IsMainworld. The single source both render.System
// and the JSON API's toSystemResponse group by, so the two stay
// consistent.
func (s StarSystem) SystemBodies() ([]Orbit, map[StellarRole][]Orbit, map[int][]Orbit) {
	var starOrbits []Orbit

	bodiesByRole := map[StellarRole][]Orbit{}
	satellitesOf := map[int][]Orbit{}

	for _, o := range s.Orbits {
		switch {
		case o.Star != nil:
			starOrbits = append(starOrbits, o)
		case o.Satellite:
			satellitesOf[o.Number] = append(satellitesOf[o.Number], o)
		default:
			bodiesByRole[o.HostRole] = append(bodiesByRole[o.HostRole], o)
		}
	}

	sort.Slice(starOrbits, func(i, j int) bool { return starOrbits[i].Star.Role < starOrbits[j].Star.Role })

	for role := range bodiesByRole {
		sort.Slice(
			bodiesByRole[role],
			func(i, j int) bool { return bodiesByRole[role][i].Number < bodiesByRole[role][j].Number },
		)
	}

	for number := range satellitesOf {
		sort.SliceStable(
			satellitesOf[number],
			func(i, j int) bool { return satellitesOf[number][i].Close && !satellitesOf[number][j].Close },
		)
	}

	return starOrbits, bodiesByRole, satellitesOf
}
