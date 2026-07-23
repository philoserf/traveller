// Command worldgen rolls a Traveller5 world (UWP + trade codes) and prints
// it. Extensions (Importance/Economic/Cultural), star systems, and
// markdown rendering are not implemented yet — see world/generate.go.
package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/world"
)

func main() {
	seed := flag.Int64("seed", 0, "PRNG seed (0 = derive from current time)")

	flag.Parse()

	s := dice.ResolveSeed(*seed)
	w := world.Generate(dice.RollerFromSeed(s))

	// This is where a future render.World(w) (markdown) replaces the
	// Printf block below.
	fmt.Printf("UWP: %s\n", w.UWP)
	fmt.Printf("Trade Codes: %s\n", strings.Join(world.TradeCodeStrings(w.TradeCodes), " "))
	fmt.Printf("(seed: %d)\n", s)
}
