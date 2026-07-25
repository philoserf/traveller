// Command chargen generates a Traveller5 character and renders it as
// Markdown. Five careers exist so far — Scout, Citizen, Noble, Marine,
// and Soldier, selected via -career — see character/character_generate.go,
// character/citizen_character_generate.go,
// character/noble_character_generate.go,
// character/marine_character_generate.go, and
// character/soldier_character_generate.go for what's generated and
// what's still deferred (Name, Equipment, Education, Command College,
// Spacer, and the other 7 T5 careers).
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
	careerName := flag.String("career", "scout", "career to generate: scout, citizen, noble, marine, or soldier")

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
	case "noble":
		c, ok = character.GenerateNobleCharacter(r)
	case "marine":
		c, ok = character.GenerateMarineCharacter(r)
	case "soldier":
		c, ok = character.GenerateSoldierCharacter(r)
	default:
		fmt.Fprintf(
			os.Stderr,
			"chargen: -career must be \"scout\", \"citizen\", \"noble\", \"marine\", or \"soldier\", got %q\n",
			*careerName,
		)
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
			"chargen: this attempt did not survive character generation, or never qualified "+
				"(Scout/Marine/Soldier: Book 1 p.69; Noble: Soc < B, p.85)",
		)
		os.Exit(1)
	}
}
