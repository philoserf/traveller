// Command secgen rolls a full Traveller5 sector — a 32x40 hex grid, each
// hex either empty or holding a complete generated star system — and
// renders it as Markdown.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/render"
	"github.com/philoserf/traveller/world"
)

func main() {
	name := flag.String("name", "Unnamed", "sector name")
	densityName := flag.String(
		"density", world.DensityStandard.String(),
		"System Presence density: Extra Galactic, Rift, Sparse, Scattered, Standard, Dense, Cluster, Core",
	)
	subsector := flag.String("subsector", "", "single letter A-P — limit output to that 80-hex block only")

	// dice.SeedFlag itself calls flag.Parse, so every other flag must be
	// registered above this line.
	s := dice.SeedFlag()

	density, ok := world.ParseDensity(*densityName)
	if !ok {
		fmt.Fprintf(os.Stderr, "secgen: unknown density %q\n", *densityName)
		os.Exit(1)
	}

	if *subsector != "" && (len(*subsector) != 1 || (*subsector)[0] < 'A' || (*subsector)[0] > 'P') {
		fmt.Fprintln(os.Stderr, "secgen: -subsector must be a single letter A-P")
		os.Exit(1)
	}

	r := dice.RollerFromSeed(s)
	sec := world.GenerateSector(r, *name, density)

	if *subsector != "" {
		sec.Hexes = sec.Subsector((*subsector)[0])
	}

	fmt.Print(render.Sector(sec))
	fmt.Printf("\n_(seed: %d)_\n", s)
}
