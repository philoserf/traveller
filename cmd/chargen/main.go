// Command chargen generates a Traveller5 character and renders it as
// Markdown. Nine careers exist so far — Scout, Citizen, Noble, Marine,
// Soldier, Spacer, Rogue, Scholar, and Entertainer, selected via
// -career — see character/character_generate.go,
// character/citizen_character_generate.go,
// character/noble_character_generate.go,
// character/marine_character_generate.go,
// character/soldier_character_generate.go,
// character/spacer_character_generate.go,
// character/rogue_character_generate.go,
// character/scholar_character_generate.go, and
// character/entertainer_character_generate.go for what's generated and
// what's still deferred (Name, Equipment, Education, Command College, a
// multi-term Prison-sentence simulation for Rogue, Scholar's own
// Major/Minor selection and Waivers, Entertainer's own optional 2nd/3rd
// Flux rolls and Comeback, Craftsman/Functionary (both architecturally
// blocked, see the chargen plan history), and the other 3 T5 careers).
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
	careerName := flag.String(
		"career",
		"scout",
		"career to generate: scout, citizen, noble, marine, soldier, spacer, rogue, scholar, or entertainer",
	)

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
	case "spacer":
		c, ok = character.GenerateSpacerCharacter(r)
	case "rogue":
		c, ok = character.GenerateRogueCharacter(r)
	case "scholar":
		c, ok = character.GenerateScholarCharacter(r)
	case "entertainer":
		c, ok = character.GenerateEntertainerCharacter(r)
	default:
		fmt.Fprintf(
			os.Stderr,
			"chargen: -career must be \"scout\", \"citizen\", \"noble\", \"marine\", \"soldier\", \"spacer\", "+
				"\"rogue\", \"scholar\", or \"entertainer\", got %q\n",
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
				"(Scout/Marine/Soldier/Spacer: Book 1 p.69; Noble: Soc < B, p.85; Rogue: Begin failed, p.84; "+
				"Scholar: Begin failed or died, p.76; Entertainer: Begin failed, p.77)",
		)
		os.Exit(1)
	}
}
