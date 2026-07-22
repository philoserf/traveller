package character

import "github.com/philoserf/traveller/ehex"

// Position identifies one of the six characteristic slots every sophont
// species has. Which trait each position represents varies by species
// (e.g. human C2 is Dexterity; other species use Agility or Grace) — see
// Character.GeneticProfile.
type Position int

// Position values, in UPP string order.
const (
	C1 Position = iota // Strength or analog
	C2                 // Dexterity, Agility, or Grace
	C3                 // Endurance, Stamina, or Vigor
	C4                 // Intelligence (universal)
	C5                 // Education, Training, or Instinct
	C6                 // Social Standing, Charisma, or Caste
)

// UPP is a Universal Personality Profile: the six characteristics plus the
// two obscure characteristics (Sanity, Psionics) every character has.
type UPP struct {
	Characteristics [6]ehex.Value // indexed by Position
	Sanity          ehex.Value
	Psionics        ehex.Value
}

func (u UPP) String() string {
	var s [6]byte
	for i, c := range u.Characteristics {
		s[i] = c.Byte()
	}

	return string(s[:])
}
