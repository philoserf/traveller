// Command worldgen rolls a Traveller5 world (UWP + trade codes) and prints
// it. Extensions (Importance/Economic/Cultural), star systems, and
// markdown rendering are not implemented yet — see world/generate.go.
package main

import (
	"flag"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/world"
)

func main() {
	seed := flag.Int64("seed", 0, "PRNG seed (0 = derive from current time)")

	flag.Parse()

	s := *seed
	if s == 0 {
		s = time.Now().UnixNano()
	}

	roller := dice.New(rand.NewPCG(uint64(s), uint64(s)))
	w := world.Generate(roller)

	codes := make([]string, len(w.TradeCodes))
	for i, c := range w.TradeCodes {
		codes[i] = string(c)
	}

	// This is where a future render.World(w) (markdown) replaces the
	// Printf block below.
	fmt.Printf("UWP: %s\n", w.UWP)
	fmt.Printf("Trade Codes: %s\n", strings.Join(codes, " "))
	fmt.Printf("(seed: %d)\n", s)
}
