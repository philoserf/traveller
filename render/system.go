package render

import (
	"fmt"
	"sort"
	"strings"

	"github.com/philoserf/traveller/world"
)

// System renders s as a Markdown system sheet: its stars (spectral type,
// size, orbit, and HZ orbit), the mainworld's own orbit placement and
// UWP/TradeCodes/PBG/extensions, and every other body placed in the
// system (Gas Giants, Belts, secondary worlds — see world.GenerateSystem
// for what's placed and why), sorted by orbit number. Doesn't call World
// for the mainworld section — World's own "#"/"##" headers don't compose
// cleanly nested inside a larger document, so this renders the same
// fields directly at its own heading level instead.
func System(s world.StarSystem) string {
	mwOrbit := s.Orbits[s.MainworldOrbit]
	mw := mwOrbit.World

	var b strings.Builder

	fmt.Fprintf(&b, "# %s System\n\n", title(*mw))

	fmt.Fprint(&b, "## Stars\n\n")

	for _, star := range s.Stars() {
		fmt.Fprintf(&b, "- %s\n", starLine(*star))
	}

	fmt.Fprint(&b, "\n## Mainworld\n\n")

	if mwOrbit.Satellite {
		// AU is left unset for a satellite orbit (see Orbit's doc comment)
		// — it orbits the body sharing this Number, not the star directly,
		// at a sub-orbit distance the book doesn't tabulate. Showing "(0.0
		// AU)" here would read as a real distance, not an unset field.
		fmt.Fprintf(&b, "**Orbit:** %d\n\n", mwOrbit.Number)
		fmt.Fprint(&b, "**Satellite of:** a Gas Giant sharing this orbit\n\n")
	} else {
		fmt.Fprintf(&b, "**Orbit:** %d (%.1f AU)\n\n", mwOrbit.Number, mwOrbit.AU)
	}

	fmt.Fprintf(&b, "**UWP:** %s\n\n", mw.UWP)
	fmt.Fprintf(&b, "**Trade Codes:** %s\n\n", joinOrNone(world.TradeCodeStrings(mw.TradeCodes)))
	fmt.Fprintf(&b, "**Bases:** %s\n\n", joinOrNone(world.BaseStrings(mw.Bases)))
	fmt.Fprintf(&b, "**PBG:** %s\n\n", mw.PBG)

	if zone := mw.TravelZone.String(); zone != "" {
		fmt.Fprintf(&b, "**Travel Zone:** %s\n\n", zone)
	}

	fmt.Fprint(&b, "### Extensions\n\n")
	fmt.Fprintf(&b, "- **Importance:** %+d\n", int(mw.Importance))
	fmt.Fprintf(&b, "- **Economic:** Resources %d, Labor %d, Infrastructure %d, Efficiency %+d\n",
		mw.Economic.Resources, mw.Economic.Labor, mw.Economic.Infrastructure, mw.Economic.Efficiency)
	fmt.Fprintf(&b, "- **Cultural:** Heterogeneity %d, Acceptance %d, Strangeness %d, Symbols %d\n",
		mw.Cultural.Heterogeneity, mw.Cultural.Acceptance, mw.Cultural.Strangeness, mw.Cultural.Symbols)

	fmt.Fprint(&b, "\n## Other Bodies\n\n")

	others := otherBodies(s)
	if len(others) == 0 {
		fmt.Fprint(&b, "None.\n")
	}

	for _, o := range others {
		fmt.Fprintf(&b, "- %s\n", otherBodyLine(o))
	}

	return b.String()
}

// otherBodies returns every Orbit in s besides the mainworld's own and
// the stars', sorted by orbit number for a readable listing.
func otherBodies(s world.StarSystem) []world.Orbit {
	var others []world.Orbit

	for i, o := range s.Orbits {
		if i == s.MainworldOrbit || o.Star != nil {
			continue
		}

		others = append(others, o)
	}

	sort.Slice(others, func(i, j int) bool { return others[i].Number < others[j].Number })

	return others
}

// otherBodyLine renders one non-mainworld, non-star body: a Gas Giant
// (Size letter and Bracket), or a placed World with its Trade Codes.
func otherBodyLine(o world.Orbit) string {
	if o.GasGiant != nil {
		return fmt.Sprintf("Orbit %d: Gas Giant, Size %c (%s)", o.Number, o.GasGiant.Size, o.GasGiant.Bracket)
	}

	return fmt.Sprintf(
		"Orbit %d: %s — %s",
		o.Number,
		o.World.UWP,
		joinOrNone(world.TradeCodeStrings(o.World.TradeCodes)),
	)
}

// starLine renders one star's spectral classification, role, orbit, HZ
// orbit, and whether it has a Companion.
func starLine(star world.Star) string {
	spec := fmt.Sprintf("%s%d %s", string(star.SpectralType), star.SpectralDecimal, star.LuminosityClass)
	if star.SpectralType == world.SpectralDegenerate {
		spec = string(star.SpectralType) + " " + star.LuminosityClass // Degenerate stars have no decimal
	}

	line := fmt.Sprintf("%s: %s (HZ orbit %d)", star.Role, spec, star.HabitableZoneOrbit)
	if star.Companion != nil {
		line += ", with a Companion"
	}

	return line
}
