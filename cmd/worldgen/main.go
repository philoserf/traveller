// Command worldgen rolls a Traveller5 world (UWP, trade codes, bases, PBG,
// and the Importance/Economic/Cultural extensions) and renders it as
// Markdown. Star systems are not implemented yet — see world/generate.go.
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
	w := world.Generate(dice.RollerFromSeed(s))

	fmt.Print(render.World(w))
	fmt.Printf("\n_(seed: %d)_\n", s)
}
