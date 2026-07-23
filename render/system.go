package render

import (
	"fmt"
	"strings"

	"github.com/philoserf/traveller/system"
	"github.com/philoserf/traveller/world"
)

// System renders s as a Markdown system sheet: the mainworld's own orbit
// placement and UWP/TradeCodes/PBG/extensions, then every star grouped
// with the bodies it hosts (Gas Giants, Belts, secondary worlds — see
// system.GenerateSystem for what's placed and why) and their satellites,
// sorted by orbit number within each star's group. Doesn't call World for
// the mainworld section — World's own "#"/"##" headers don't compose
// cleanly nested inside a larger document, so this renders the same
// fields directly at its own heading level instead.
func System(s system.StarSystem) string {
	mwOrbit := s.Orbits[s.MainworldOrbit]
	mw := mwOrbit.World

	var b strings.Builder

	fmt.Fprintf(&b, "# %s System\n\n", title(*mw))

	starOrbits, bodiesByRole, satellitesOf := s.SystemBodies()

	writeMainworld(&b, mwOrbit)

	fmt.Fprint(&b, "\n## System\n\n")

	for _, o := range starOrbits {
		fmt.Fprintf(&b, "### %s\n\n", starHeading(o))

		bodies := bodiesByRole[o.Star.Role]
		if len(bodies) == 0 {
			fmt.Fprint(&b, "None.\n\n")

			continue
		}

		for _, body := range bodies {
			fmt.Fprintf(&b, "- %s\n", otherBodyLine(body, s.IsMainworld(body)))

			for _, sat := range satellitesOf[body.Number] {
				fmt.Fprintf(&b, "  - %s\n", satelliteLine(sat, s.IsMainworld(sat)))
			}
		}

		fmt.Fprint(&b, "\n")
	}

	return strings.TrimRight(b.String(), "\n") + "\n"
}

// writeMainworld renders the "## Mainworld" section: orbit placement
// (Close/Far-aware when it's a satellite of a Gas Giant), UWP, Trade
// Codes, Bases, PBG, Travel Zone, and Extensions. The mainworld's own
// satellites (if any) aren't repeated here — they're already visible,
// marked "(Mainworld)", nested under its entry in "## System" below,
// alongside every other body's satellites.
func writeMainworld(b *strings.Builder, mwOrbit system.Orbit) {
	mw := mwOrbit.World

	fmt.Fprint(b, "## Mainworld\n\n")

	if mwOrbit.Satellite {
		// AU is left unset for a satellite orbit (see Orbit's doc comment)
		// — it orbits the body sharing this Number, not the star directly,
		// at a sub-orbit distance the book doesn't tabulate. Showing "(0.0
		// AU)" here would read as a real distance, not an unset field.
		fmt.Fprintf(b, "**Orbit:** %d\n\n", mwOrbit.Number)
		fmt.Fprintf(b, "**%s satellite of:** a Gas Giant sharing this orbit\n\n", closeFarLabel(mwOrbit.Close))
	} else {
		fmt.Fprintf(b, "**Orbit:** %d (%.1f AU)\n\n", mwOrbit.Number, mwOrbit.AU)
	}

	fmt.Fprintf(b, "**UWP:** %s\n\n", mw.UWP)
	fmt.Fprintf(b, "**Trade Codes:** %s\n\n", joinOrNone(world.TradeCodeStrings(mw.TradeCodes)))
	fmt.Fprintf(b, "**Bases:** %s\n\n", joinOrNone(world.BaseStrings(mw.Bases)))
	fmt.Fprintf(b, "**PBG:** %s\n\n", mw.PBG)

	if zone := mw.TravelZone.String(); zone != "" {
		fmt.Fprintf(b, "**Travel Zone:** %s\n\n", zone)
	}

	if mw.Ring {
		fmt.Fprint(b, "**Ring:** yes\n\n")
	}

	fmt.Fprint(b, "### Extensions\n\n")
	fmt.Fprintf(b, "- **Importance:** %+d\n", int(mw.Importance))
	fmt.Fprintf(b, "- **Economic:** Resources %d, Labor %d, Infrastructure %d, Efficiency %+d\n",
		mw.Economic.Resources, mw.Economic.Labor, mw.Economic.Infrastructure, mw.Economic.Efficiency)
	fmt.Fprintf(b, "- **Cultural:** Heterogeneity %d, Acceptance %d, Strangeness %d, Symbols %d\n",
		mw.Cultural.Heterogeneity, mw.Cultural.Acceptance, mw.Cultural.Strangeness, mw.Cultural.Symbols)
}

// otherBodyLine renders one non-star, non-Satellite body: a Gas Giant
// (Size letter and Bracket), or a placed World with its Trade Codes —
// either way with a Ring suffix when it has one, and a "(Mainworld)"
// suffix when isMainworld (a Gas Giant is never the mainworld itself).
func otherBodyLine(o system.Orbit, isMainworld bool) string {
	var line string
	if o.GasGiant != nil {
		line = fmt.Sprintf("Orbit %d: Gas Giant, Size %c (%s)", o.Number, o.GasGiant.Size, o.GasGiant.Bracket)
		if o.GasGiant.Ring {
			line += ", with a Ring"
		}
	} else {
		line = fmt.Sprintf(
			"Orbit %d: %s — %s",
			o.Number,
			o.World.UWP,
			joinOrNone(world.TradeCodeStrings(o.World.TradeCodes)),
		)
		if o.World.Ring {
			line += ", with a Ring"
		}
	}

	if isMainworld {
		line += " (Mainworld)"
	}

	return line
}

// closeFarLabel is the shared "Close"/"Far" wording for an Orbit.Close
// value, per Book 3 p.21/24 (2D<=7 tidally locked "Close" vs 2D>=8 "Far").
func closeFarLabel(isClose bool) string {
	if isClose {
		return "Close"
	}

	return "Far"
}

// satelliteLine renders one satellite: Close or Far, its UWP, its Trade
// Codes, and a "(Mainworld)" suffix when isMainworld — a mainworld that
// is itself a satellite of a Gas Giant.
func satelliteLine(o system.Orbit, isMainworld bool) string {
	line := fmt.Sprintf(
		"%s satellite: %s — %s",
		closeFarLabel(o.Close),
		o.World.UWP,
		joinOrNone(world.TradeCodeStrings(o.World.TradeCodes)),
	)

	if isMainworld {
		line += " (Mainworld)"
	}

	return line
}

// starSpec renders a star's spectral classification, e.g. "G7 IV" — or,
// for a Degenerate star (white dwarf/brown dwarf), "D D": SpectralDecimal
// is meaningless for them (system.Star's own doc comment), so the type
// letter and LuminosityClass ("D") are shown instead of a decimal.
func starSpec(star system.Star) string {
	if star.SpectralType == system.SpectralDegenerate {
		return string(star.SpectralType) + " " + star.LuminosityClass
	}

	return fmt.Sprintf("%s%d %s", string(star.SpectralType), star.SpectralDecimal, star.LuminosityClass)
}

// starHeading renders one star's own group heading: role, spectral
// classification, its own orbit Number (omitted for the Primary, whose
// Number is the primaryOrbitNumber sentinel, not a real orbit slot — the
// same o.Number >= 0 check api.toStarResponse uses for this), HZ orbit,
// and whether it has a Companion.
func starHeading(o system.Orbit) string {
	star := *o.Star

	var orbitPart string
	if o.Number >= 0 {
		orbitPart = fmt.Sprintf("Orbit %d, ", o.Number)
	}

	line := fmt.Sprintf("%s: %s (%sHZ orbit %d)", star.Role, starSpec(star), orbitPart, star.HabitableZoneOrbit)
	if star.Companion != nil {
		line += ", with a Companion"
	}

	return line
}
