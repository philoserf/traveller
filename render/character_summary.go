package render

import (
	"fmt"
	"strings"

	"github.com/philoserf/traveller/character"
)

// CharacterSummary renders c as Book 1's own narrative "finished
// character" block — the format p.60-66's own worked examples use for
// Eneri Dinsha and Barr Vech:
//
//	Eneri Dinsha 9AB58A. Genetic 4553XX
//	from: Regina (1910 Spinward Marches)
//	Trader-1, Chef-1.
//	Psychology-2, Robotics-1, Astrogation-1, Pilot-4.
//	Imperial Navy Ensign O1.
//	Age 23. Born 069-1082 (or deferred)
//	Current Date: 001-1105 (if CharGen ends now).
//
// Three deliberate departures from that shape, all because the data to
// do better doesn't exist yet, not because the book's own layout was
// rejected:
//
//   - No sector/hex place name ("Regina (1910 Spinward Marches)") —
//     issue #118, no world-name generation exists yet. Prints the
//     Homeworld UWP instead.
//   - No homeworld-skills/career-skills split ("Trader-1, Chef-1." then
//     a separate line for the rest) — Character.Skills is a single
//     post-aggregateSkills bag with no surviving origin tag once
//     applyEducation/career resolution merge everything together, so
//     the book's own two-tier presentation isn't reconstructable from
//     what's stored. One combined, comma-joined line instead.
//   - No "Imperial Navy" branch prefix on Rank — that phrasing is this
//     specific worked example's own Third Imperium flavor, not a rule;
//     every other renderer in this codebase stays setting-agnostic (no
//     hardcoded "Imperial" anywhere), so Rank prints as-is ("O1
//     Ensign").
//
// Current Date is dropped entirely — an in-game campaign artifact, not
// a fact about the character.
func CharacterSummary(c character.Character) string {
	lines := []string{summaryNameLine(c), "from: " + c.Homeworld}

	if len(c.Skills) > 0 {
		skills := make([]string, len(c.Skills))
		for i, s := range c.Skills {
			skills[i] = skillNotation(s)
		}

		lines = append(lines, strings.Join(skills, ", ")+".")
	}

	if c.Rank != "" {
		lines = append(lines, c.Rank+".")
	}

	lines = append(lines, fmt.Sprintf("Age %d. Born %s.", c.Age, c.Birthdate))

	return strings.Join(lines, "\n") + "\n"
}

// summaryNameLine is "{Name} {UPP}. Genetic {Profile}" when a Name is
// set, or "{UPP}." alone otherwise — nothing in the character package
// generates Name yet (characterTitle's own doc comment), so the
// UPP-only form is the common case. Genetic Profile appears only away
// from its Human default, matching writeMetadata's own convention.
func summaryNameLine(c character.Character) string {
	line := c.UPP.String() + "."
	if c.Name != "" {
		line = c.Name + " " + c.UPP.String() + "."
	}

	if c.GeneticProfile != "" && c.GeneticProfile != character.HumanGeneticProfile {
		line += " Genetic " + c.GeneticProfile
	}

	return line
}
