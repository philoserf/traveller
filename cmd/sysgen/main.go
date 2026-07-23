// Command sysgen rolls a full Traveller5 star system around a freshly
// generated mainworld — stars, every other gas giant/belt/secondary
// world, satellites, and rings — and renders it as Markdown. See
// system/system_generate.go for what's placed and why.
package main

import (
	"fmt"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/render"
	"github.com/philoserf/traveller/system"
	"github.com/philoserf/traveller/world"
)

func main() {
	s := dice.SeedFlag()
	r := dice.RollerFromSeed(s)
	mw := world.Generate(r)
	sys := system.GenerateSystem(r, mw)

	fmt.Print(render.System(sys))
	fmt.Printf("\n_(seed: %d)_\n", s)
}
