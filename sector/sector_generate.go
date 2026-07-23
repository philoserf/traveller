package sector

import (
	"fmt"
	"hash/fnv"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/system"
	"github.com/philoserf/traveller/world"
)

// hexLocation formats col/row (1-based: col 1-32, row 1-40) as Book 3's
// zero-padded "CCRR" — e.g. (1,1) -> "0101", (32,40) -> "3240".
func hexLocation(col, row int) string {
	return fmt.Sprintf("%02d%02d", col, row)
}

// deriveSeed derives a deterministic seed from sectorSeed, location, and
// purpose (a plain FNV-1a hash of the three combined, not a security
// boundary). Two different purpose tags for the same (sectorSeed,
// location) produce independent, uncorrelated seeds — used so a hex's
// System Presence roll and its own system generation each get their own
// Roller, with no shared stream between them to keep in sync.
func deriveSeed(sectorSeed int64, location, purpose string) int64 {
	h := fnv.New64a()
	fmt.Fprintf(h, "%d:%s:%s", sectorSeed, location, purpose)

	//nolint:gosec // any uint64 bit pattern is a valid PRNG seed; this is a hash digest, not a security boundary
	return int64(h.Sum64())
}

// HexSeed derives the seed that reproduces a specific hex's own system —
// dice.RollerFromSeed(HexSeed(sectorSeed, location)) fed through Generate
// then GenerateSystem, with no other rolls first, exactly reproduces
// what GenerateSector rolled for that hex: its own Roller is never
// shared with the hex's System Presence roll (a separate, independently
// seeded Roller — see GenerateSector), so there's no prior consumption
// to replicate.
func HexSeed(sectorSeed int64, location string) int64 {
	return deriveSeed(sectorSeed, location, "system")
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

// GenerateSector rolls a full 1280-hex Sector named name, seeded by seed:
// for each hex, in column-major order (matching Sector.Hexes' own
// documented order), its own System Presence Roller (seeded via
// deriveSeed(seed, location, "presence") — independent of, and never
// shared with, the hex's own system-generation Roller) decides
// (rollSystemPresent) whether a system exists there; if so, a second,
// separately-seeded Roller (HexSeed(seed, location)) feeds Generate and
// GenerateSystem to produce its mainworld and system exactly as sysgen
// already does standalone, with Sector/Hex stamped onto both the
// mainworld World and its StarSystem — the previously-unused fields both
// types already carry. Splitting presence and generation into two
// independent Rollers (rather than one shared stream, per hex or across
// the whole grid) is what makes HexSeed alone — with no other rolls
// first — enough to reproduce a populated hex's own system standalone.
//
// The Classic System Contents Table's own separate "Gas Giant Presence"
// (2D<=8) and "Asteroid" (2D=2) sub-rolls are deliberately not
// reproduced here: they're a coarse map-symbol approximation, and
// Generate/GenerateSystem already produce the real, detailed answer to
// both questions (an actual PBG.GasGiants count, and an actual Size-0
// Asteroid-Belt mainworld when the dice land there) without a separate
// roll.
func GenerateSector(seed int64, name string, density Density) Sector {
	hexes := make([]Hex, 0, sectorWidth*sectorHeight)

	for col := 1; col <= sectorWidth; col++ {
		for row := 1; row <= sectorHeight; row++ {
			location := hexLocation(col, row)
			presenceRoller := dice.RollerFromSeed(deriveSeed(seed, location, "presence"))

			if !rollSystemPresent(presenceRoller, density) {
				hexes = append(hexes, Hex{Location: location})

				continue
			}

			r := dice.RollerFromSeed(HexSeed(seed, location))

			mw := world.Generate(r)
			mw.Sector = name
			mw.Hex = location

			sys := system.GenerateSystem(r, mw)
			sys.Sector = name
			sys.Hex = location

			hexes = append(hexes, Hex{Location: location, System: &sys})
		}
	}

	return Sector{Name: name, Hexes: hexes}
}
