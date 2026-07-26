// Command chargen generates a Traveller5 character and renders it as
// Markdown. All thirteen T5 careers are implemented — Scout, Citizen,
// Noble, Marine, Soldier, Spacer, Rogue, Scholar, Entertainer, Merchant,
// Agent, Functionary, and Craftsman, selected via -career — see
// character/character_generate.go, character/citizen_character_generate.go,
// character/noble_character_generate.go,
// character/marine_character_generate.go,
// character/soldier_character_generate.go,
// character/spacer_character_generate.go,
// character/rogue_character_generate.go,
// character/scholar_character_generate.go,
// character/entertainer_character_generate.go,
// character/merchant_character_generate.go,
// character/agent_character_generate.go,
// character/functionary_generate.go/functionary_loop.go, and
// character/craftsman_generate.go/craftsman_loop.go for what's generated
// and what's still deferred (Name, Education, Command College, a
// multi-term Prison-sentence simulation for Rogue, Scholar's own
// Major/Minor selection and Waivers, Entertainer's own optional 2nd/3rd
// Flux rolls and Comeback, Merchant's own Ship Owner Fame bonus (this
// codebase tracks Ship Shares, not outright ownership), Agent's own full
// Undercover Assignment table (rank titles, the three-die A/B/C
// mechanic, and Citizen's/Scout's own special-cased rows — simplified to
// a uniform pick among this codebase's own already-implemented career
// skill tables), Functionary's own F6 Rank title for preceding careers
// Book 1 doesn't name (falls back to the generic "Director"), and
// Craftsman's own QREBS Masterpiece-dimension allocation and Vintage
// Masterpiece value appreciation (no structured item record or
// time-since-creation concept exists in this codebase to attach either
// to).
//
// -career also accepts a comma-separated ordered list ("scout,spacer")
// to chain multiple careers over one lifetime (character/career_chain.go:
// GenerateCareerChainCharacter), falling back to Citizen if every listed
// career fails to Begin. Each non-final entry means "transfer after one
// term"; the final career follows its normal Continue rolls. Functionary
// and Noble must be final entries, and Functionary and Craftsman may
// never be first.
//
// -age N stops the chain from attempting any further term or career
// once the character's age would reach or exceed N — never a
// retroactive cut of an already-resolved term, only "don't attempt what
// comes next." A nonzero -age always routes through the chain function,
// even for a single -career name, since the legacy single-career
// functions have no way to honor an age target.
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
		"career (or comma-separated ordered list, e.g. \"scholar,functionary\") to generate: scout, citizen, "+
			"noble, marine, soldier, spacer, rogue, scholar, entertainer, merchant, agent, functionary, or "+
			"craftsman — each non-final career lasts one term; functionary and noble must be final; "+
			"functionary and craftsman may not be first",
	)
	ageTarget := flag.Int(
		"age",
		0,
		"stop attempting further career progression once age would reach or exceed this (0 = no target)",
	)

	// dice.SeedFlag itself calls flag.Parse, so every other flag must be
	// registered above this line.
	s := dice.SeedFlag()

	r := dice.RollerFromSeed(s)

	if *ageTarget < 0 {
		fmt.Fprintf(os.Stderr, "chargen: -age must not be negative, got %d\n", *ageTarget)
		os.Exit(1)
	}

	names := splitCareerNames(*careerName)

	var c character.Character

	var ok bool

	if len(names) == 1 && *ageTarget == 0 {
		ok = generateSingleCareer(r, names[0], *careerName, &c)
	} else {
		var err error

		c, ok, err = character.GenerateCareerChainCharacter(r, names, *ageTarget)
		if err != nil {
			fmt.Fprintf(os.Stderr, "chargen: %v\n", err)
			os.Exit(1)
		}
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
				"(Scout/Marine/Soldier/Spacer/Agent: Book 1 p.69/p.83; Noble: Soc < B, p.85; Rogue: Begin failed, "+
				"p.84; Scholar: Begin failed or died, p.76; Entertainer: Begin failed, p.77)",
		)
		os.Exit(1)
	}
}

// splitCareerNames parses -career's comma-separated form, trimming and
// lowercasing each entry — shared by both the single-name and chain
// paths so "-career Scout" and "-career scout" behave identically.
func splitCareerNames(careerName string) []string {
	parts := strings.Split(careerName, ",")
	names := make([]string, len(parts))

	for i, p := range parts {
		names[i] = strings.ToLower(strings.TrimSpace(p))
	}

	return names
}

// generateSingleCareer keeps today's exact single-career dispatch
// unchanged — every one of these 11 values, including "noble" and
// "citizen", byte-for-byte matches pre-chaining behavior. Multi-entry
// lists go through character.GenerateCareerChainCharacter instead.
// rawCareerName is the untouched -career flag value (not the lowercased/
// trimmed name) so an invalid single value's error message echoes
// exactly what the user typed, matching pre-chaining behavior.
func generateSingleCareer(r *dice.Roller, name, rawCareerName string, c *character.Character) bool {
	ok := true

	switch name {
	case "scout":
		*c, ok = character.GenerateScoutCharacter(r)
	case "citizen":
		*c, ok = character.GenerateCitizenCharacter(r)
	case "noble":
		*c, ok = character.GenerateNobleCharacter(r)
	case "marine":
		*c, ok = character.GenerateMarineCharacter(r)
	case "soldier":
		*c, ok = character.GenerateSoldierCharacter(r)
	case "spacer":
		*c, ok = character.GenerateSpacerCharacter(r)
	case "rogue":
		*c, ok = character.GenerateRogueCharacter(r)
	case "scholar":
		*c, ok = character.GenerateScholarCharacter(r)
	case "entertainer":
		*c, ok = character.GenerateEntertainerCharacter(r)
	case "merchant":
		*c, ok = character.GenerateMerchantCharacter(r)
	case "agent":
		*c, ok = character.GenerateAgentCharacter(r)
	default:
		fmt.Fprintf(
			os.Stderr,
			"chargen: -career must be \"scout\", \"citizen\", \"noble\", \"marine\", \"soldier\", \"spacer\", "+
				"\"rogue\", \"scholar\", \"entertainer\", \"merchant\", \"agent\", or a comma-separated list "+
				"of those, got %q\n",
			rawCareerName,
		)
		os.Exit(1)
	}

	return ok
}
