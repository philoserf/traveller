// Command chargen generates a Traveller5 Scout character and renders it
// as Markdown. Only the Scout career exists so far — see
// character/character_generate.go for what's generated and what's still
// deferred (Name, Age, Cash amounts, Equipment, and the other 12 T5
// careers).
package main

import (
	"fmt"
	"os"

	"github.com/philoserf/traveller/character"
	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/render"
)

func main() {
	s := dice.SeedFlag()
	c, ok := character.GenerateScoutCharacter(dice.RollerFromSeed(s))

	fmt.Print(render.Character(c))
	fmt.Printf("\n_(seed: %d)_\n", s)

	if !ok {
		// A died-or-never-qualified attempt is a valid, expected RNG
		// outcome, not a tool malfunction — the sheet above still prints in
		// full. Exiting 1 here (rather than always 0) preserves the exit
		// code as a real signal for any caller that wants to detect this
		// outcome programmatically, not just by scraping stdout.
		fmt.Fprintln(
			os.Stderr,
			"chargen: this attempt did not survive character generation, or never qualified for Scout (Book 1 p.69)",
		)
		os.Exit(1)
	}
}
