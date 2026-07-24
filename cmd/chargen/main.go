// Command chargen generates a Traveller5 character and renders it as
// Markdown. Two careers exist so far — Scout and Citizen, selected via
// -career — see character/character_generate.go and
// character/citizen_character_generate.go for what's generated and
// what's still deferred (Name, Age, Cash amounts, Equipment, Aging,
// Education, and the other 11 T5 careers).
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/philoserf/traveller/character"
	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/render"
)

func main() {
	careerName := flag.String("career", "scout", "career to generate: scout or citizen")

	// dice.SeedFlag itself calls flag.Parse, so every other flag must be
	// registered above this line.
	s := dice.SeedFlag()

	r := dice.RollerFromSeed(s)

	var c character.Character

	ok := true

	switch strings.ToLower(*careerName) {
	case "scout":
		c, ok = character.GenerateScoutCharacter(r)
	case "citizen":
		c = character.GenerateCitizenCharacter(r)
	default:
		fmt.Fprintf(os.Stderr, "chargen: -career must be \"scout\" or \"citizen\", got %q\n", *careerName)
		os.Exit(1)
	}

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
