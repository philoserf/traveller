// Command sysgen rolls a Traveller5 star system around a freshly
// generated mainworld — stars, orbit placement, habitable zone, and
// mainworld placement — and renders it as Markdown. Placing every other
// body in the system is not implemented yet — see world/system_generate.go.
package main

import (
	"flag"
	"fmt"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/render"
	"github.com/philoserf/traveller/world"
)

func main() {
	seed := flag.Int64("seed", 0, "PRNG seed (omit the flag entirely to derive one from current time)")

	flag.Parse()

	var seedPtr *int64

	flag.Visit(func(f *flag.Flag) {
		if f.Name == "seed" {
			seedPtr = seed
		}
	})

	s := dice.ResolveSeed(seedPtr)
	r := dice.RollerFromSeed(s)
	mw := world.Generate(r)
	sys := world.GenerateSystem(r, mw)

	fmt.Print(render.System(sys))
	fmt.Printf("\n_(seed: %d)_\n", s)
}
