// Command worldgen rolls a Traveller5 world (UWP, trade codes, bases, PBG,
// and the Importance/Economic/Cultural extensions) and renders it as
// Markdown. Star systems are not implemented yet — see world/generate.go.
package main

import (
	"fmt"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/render"
	"github.com/philoserf/traveller/world"
)

func main() {
	s := dice.SeedFlag()
	w := world.Generate(dice.RollerFromSeed(s))

	fmt.Print(render.World(w))
	fmt.Printf("\n_(seed: %d)_\n", s)
}
