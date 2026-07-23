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
	"github.com/philoserf/traveller/sector"
)

func main() {
	name := flag.String("name", "Unnamed", "sector name")
	densityName := flag.String(
		"density", sector.DensityStandard.String(),
		"System Presence density: Extra Galactic, Rift, Sparse, Scattered, Standard, Dense, Cluster, Core",
	)
	subsector := flag.String("subsector", "", "single letter A-P — limit output to that 80-hex block only")

	// dice.SeedFlag itself calls flag.Parse, so every other flag must be
	// registered above this line.
	s := dice.SeedFlag()

	density, ok := sector.ParseDensity(*densityName)
	if !ok {
		fmt.Fprintf(os.Stderr, "secgen: unknown density %q\n", *densityName)
		os.Exit(1)
	}

	if *subsector != "" && (len(*subsector) != 1 || !sector.ValidSubsectorLetter((*subsector)[0])) {
		fmt.Fprintln(os.Stderr, "secgen: -subsector must be a single letter A-P")
		os.Exit(1)
	}

	sec := sector.GenerateSector(s, *name, density)

	if *subsector != "" {
		sec.Hexes = sec.Subsector((*subsector)[0])
	}

	fmt.Print(render.Sector(sec))
	fmt.Printf("\n_(seed: %d)_\n", s)
}
