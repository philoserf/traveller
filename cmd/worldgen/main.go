// Command worldgen rolls a Traveller5 world (UWP, trade codes, bases, PBG,
// and the Importance/Economic/Cultural extensions) and prints it. Star
// systems and markdown rendering are not implemented yet — see
// world/generate.go.
package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/philoserf/traveller/dice"
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

	// This is where a future render.World(w) (markdown) replaces the
	// Printf block below. Extensions are printed as plain labeled values
	// rather than the rulebook's "{+4}"/"(RLI+E)"/"[HASS]" bracket
	// notation — that's presentation-layer polish for the render package
	// (see issue #7), not warranted for a CLI debug dump.
	fmt.Printf("UWP: %s\n", w.UWP)
	fmt.Printf("Trade Codes: %s\n", strings.Join(world.TradeCodeStrings(w.TradeCodes), " "))
	fmt.Printf("Bases: %s\n", strings.Join(world.BaseStrings(w.Bases), " "))
	fmt.Printf("PBG: %s\n", w.PBG)
	fmt.Printf("Importance: %+d\n", int(w.Importance))
	fmt.Printf("Economic: Resources=%d Labor=%d Infrastructure=%d Efficiency=%+d\n",
		w.Economic.Resources, w.Economic.Labor, w.Economic.Infrastructure, w.Economic.Efficiency)
	fmt.Printf("Cultural: Heterogeneity=%d Acceptance=%d Strangeness=%d Symbols=%d\n",
		w.Cultural.Heterogeneity, w.Cultural.Acceptance, w.Cultural.Strangeness, w.Cultural.Symbols)
	fmt.Printf("(seed: %d)\n", s)
}
