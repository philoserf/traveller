package world

import (
	"fmt"

	"github.com/philoserf/traveller/dice"
)

// hexLocation formats col/row (1-based: col 1-32, row 1-40) as Book 3's
// zero-padded "CCRR" — e.g. (1,1) -> "0101", (32,40) -> "3240".
func hexLocation(col, row int) string {
	return fmt.Sprintf("%02d%02d", col, row)
}

// rollNDice sums n D6 rolls. dice.Roller only has dedicated D6/TwoD6
// helpers for the common 1D/2D cases; density's own 3D (Extra Galactic)
// falls through to summing D6 three times.
func rollNDice(r *dice.Roller, n int) int {
	switch n {
	case 1:
		return r.D6()
	case 2:
		return r.TwoD6()
	default:
		sum := 0

		for range n {
			sum += r.D6()
		}

		return sum
	}
}

// rollSystemPresent rolls density's own dice count and compares to its
// "N or less" target (Book 3 p.13's System Presence table).
func rollSystemPresent(r *dice.Roller, density Density) bool {
	roll := densityTable[density]

	return rollNDice(r, roll.dice) <= roll.target
}

// GenerateSector rolls a full 1280-hex Sector named name: for each hex,
// in column-major order (matching Sector.Hexes' own documented order),
// rollSystemPresent decides whether a system exists there; if so,
// Generate and GenerateSystem produce its mainworld and system exactly
// as sysgen already does standalone, with Sector/Hex stamped onto both
// the mainworld World and its StarSystem — the previously-unused fields
// both types already carry.
//
// The Classic System Contents Table's own separate "Gas Giant Presence"
// (2D<=8) and "Asteroid" (2D=2) sub-rolls are deliberately not
// reproduced here: they're a coarse map-symbol approximation, and
// Generate/GenerateSystem already produce the real, detailed answer to
// both questions (an actual PBG.GasGiants count, and an actual Size-0
// Asteroid-Belt mainworld when the dice land there) without a separate
// roll.
func GenerateSector(r *dice.Roller, name string, density Density) Sector {
	hexes := make([]Hex, 0, sectorWidth*sectorHeight)

	for col := 1; col <= sectorWidth; col++ {
		for row := 1; row <= sectorHeight; row++ {
			location := hexLocation(col, row)

			if !rollSystemPresent(r, density) {
				hexes = append(hexes, Hex{Location: location})

				continue
			}

			mw := Generate(r)
			mw.Sector = name
			mw.Hex = location

			sys := GenerateSystem(r, mw)
			sys.Sector = name
			sys.Hex = location

			hexes = append(hexes, Hex{Location: location, System: &sys})
		}
	}

	return Sector{Name: name, Hexes: hexes}
}
