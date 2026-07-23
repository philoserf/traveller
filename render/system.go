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

	tops, satellitesOf := systemBodies(s)
	multiStar := len(s.Stars()) > 1

	writeMainworld(&b, mwOrbit, satellitesOf)

	fmt.Fprint(&b, "\n## System\n\n")

	if len(tops) == 0 {
		fmt.Fprint(&b, "None.\n")
	}

	for _, o := range tops {
		if o.Star != nil {
			fmt.Fprintf(&b, "- Orbit %d: %s\n", o.Number, starLine(*o.Star))

			continue
		}

		fmt.Fprintf(&b, "- %s\n", otherBodyLine(o, multiStar))

		for _, sat := range satellitesOf[o.Number] {
			fmt.Fprintf(&b, "  - %s\n", satelliteLine(sat))
		}
	}

	return b.String()
}

// writeMainworld renders the "## Mainworld" section: orbit placement
// (Close/Far-aware when it's a satellite of a Gas Giant), UWP, Trade
// Codes, Bases, PBG, Travel Zone, Extensions, and — for a non-satellite
// mainworld — its own Satellites subsection.
func writeMainworld(b *strings.Builder, mwOrbit world.Orbit, satellitesOf map[int][]world.Orbit) {
	mw := mwOrbit.World

	fmt.Fprint(b, "\n## Mainworld\n\n")

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

	// A satellite mainworld shares its Number with its host Gas Giant —
	// any satellitesOf that Number are the Gas Giant's own (already
	// shown under its own "System" listing below), not the mainworld's.
	// A mainworld never gets satellites of its own when it's itself a
	// satellite (satellites don't get satellites — see
	// world.GenerateSystem's topLevel snapshot).
	if !mwOrbit.Satellite {
		if mwSatellites := satellitesOf[mwOrbit.Number]; len(mwSatellites) > 0 {
			fmt.Fprint(b, "\n### Satellites\n\n")

			for _, sat := range mwSatellites {
				fmt.Fprintf(b, "- %s\n", satelliteLine(sat))
			}
		}
	}
}

// systemBodies splits every Orbit in s besides the mainworld's own and the
// Primary (whose Number is the primaryOrbitNumber sentinel, not a real
// numbered slot) into top-level entries — every Close/Near/Far Star, Gas
// Giant, and placed World — sorted by orbit number, and their satellites,
// grouped by the Number they share with their parent.
func systemBodies(s world.StarSystem) ([]world.Orbit, map[int][]world.Orbit) {
	var tops []world.Orbit

	satellitesOf := map[int][]world.Orbit{}

	for i, o := range s.Orbits {
		if i == s.MainworldOrbit || o.Star != nil && o.Star.Role == world.Primary {
			continue
		}

		if o.Satellite {
			satellitesOf[o.Number] = append(satellitesOf[o.Number], o)

			continue
		}

		tops = append(tops, o)
	}

	sort.Slice(tops, func(i, j int) bool { return tops[i].Number < tops[j].Number })

	return tops, satellitesOf
}

// otherBodyLine renders one non-mainworld, non-star, non-Satellite body:
// a Gas Giant (Size letter and Bracket), or a placed World with its
// Trade Codes — either way with a Ring suffix when it has one. multiStar
// adds a "(hosted by <Role>)" suffix — the shared orbit-numbering means a
// body's Number alone doesn't say which star placed it once more than one
// star is present; omitted in the (common) single-star case where it
// would only ever read "(hosted by Primary)".
func otherBodyLine(o world.Orbit, multiStar bool) string {
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

	if multiStar {
		line += fmt.Sprintf(" (hosted by %s)", o.HostRole)
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

// satelliteLine renders one satellite: Close or Far, its UWP, and its
// Trade Codes.
func satelliteLine(o world.Orbit) string {
	return fmt.Sprintf(
		"%s satellite: %s — %s",
		closeFarLabel(o.Close),
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
