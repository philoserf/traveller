package world

import (
	"fmt"

	"github.com/philoserf/traveller/ehex"
)

// Starport is a world's starport quality code.
type Starport byte

// Starport values, ranked from best (A: Excellent) to worst (E: Frontier);
// StarportNone means no starport exists, not merely a poor one.
const (
	StarportA    Starport = 'A'
	StarportB    Starport = 'B'
	StarportC    Starport = 'C'
	StarportD    Starport = 'D'
	StarportE    Starport = 'E'
	StarportNone Starport = 'X'
)

func (s Starport) String() string { return string(s) }

// UWP is a world's Universal World Profile: the eight-field StSAHPGL-T code.
type UWP struct {
	Starport      Starport
	Size          ehex.Value
	Atmosphere    ehex.Value
	Hydrographics ehex.Value
	Population    ehex.Value
	Government    ehex.Value
	Law           ehex.Value
	TechLevel     ehex.Value
}

func (u UWP) String() string {
	return fmt.Sprintf("%s%s%s%s%s%s%s-%s",
		u.Starport, u.Size, u.Atmosphere, u.Hydrographics,
		u.Population, u.Government, u.Law, u.TechLevel)
}
