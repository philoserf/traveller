package character

import "github.com/philoserf/traveller/ehex"

// DamageType determines which characteristics a hit reduces. T5 has no
// separate hit-point track: damage is applied directly against physical
// characteristics.
type DamageType int

const (
	Hit         DamageType = iota // C1 C2 C3
	Cut                           // C1 C2 C3, per round
	Suffocation                   // C3 C4 C5
	Heat                          // C1 C2 C3 C4 C5
	Freeze                        // C1 C2 C3 C4 C5
)

// Protection lists specialized damage-type ratings an armor can provide,
// each deflecting damage up to its rating before penetration.
type Protection struct {
	Corrosion  int // Ca
	Incendiary int // In
	Flame      int // Fl
	Radiation  int // Ra
	Sonic      int // So
	Psi        int // Ps
	Sensors    int // Se
}

// Armor absorbs Hits up to its Rating before penetration to the wearer.
type Armor struct {
	Name       string
	Rating     ehex.Value // Ar
	Protection Protection
}
