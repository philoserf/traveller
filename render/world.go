// Package render turns domain values into human-readable text. Kept
// separate from the domain packages (world, character, starship) so they
// stay pure data with no presentation concerns — see cmd/worldgen's
// former Printf block, which this package replaces.
package render

import (
	"fmt"
	"strings"

	"github.com/philoserf/traveller/world"
)

// World renders w as a Markdown world sheet, covering everything
// world.Generate currently populates (UWP, TradeCodes, Bases, PBG, and the
// Importance/Economic/Cultural extensions). Fields Generate doesn't
// produce yet (Name, Sector, Hex, TravelZone, Worlds, Notes, Nobility,
// Allegiance) are omitted rather than shown as misleading blanks or
// zeros — see world/generate.go's own doc comment for what's not
// generated, and why.
func World(w world.World) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# %s\n\n", title(w))
	fmt.Fprintf(&b, "**UWP:** %s\n\n", w.UWP)
	fmt.Fprintf(&b, "**Trade Codes:** %s\n\n", joinOrNone(world.TradeCodeStrings(w.TradeCodes)))
	fmt.Fprintf(&b, "**Bases:** %s\n\n", joinOrNone(world.BaseStrings(w.Bases)))
	fmt.Fprintf(&b, "**PBG:** %s\n\n", w.PBG)

	if zone, ok := travelZoneName(w.TravelZone); ok {
		fmt.Fprintf(&b, "**Travel Zone:** %s\n\n", zone)
	}

	fmt.Fprint(&b, "## Extensions\n\n")
	fmt.Fprintf(&b, "- **Importance:** %+d\n", int(w.Importance))
	fmt.Fprintf(&b, "- **Economic:** Resources %d, Labor %d, Infrastructure %d, Efficiency %+d\n",
		w.Economic.Resources, w.Economic.Labor, w.Economic.Infrastructure, w.Economic.Efficiency)
	fmt.Fprintf(&b, "- **Cultural:** Heterogeneity %d, Acceptance %d, Strangeness %d, Symbols %d\n",
		w.Cultural.Heterogeneity, w.Cultural.Acceptance, w.Cultural.Strangeness, w.Cultural.Symbols)

	return b.String()
}

// title falls back to the UWP code when Name is empty — Generate never
// sets Name, so this is the common case, not an edge case.
func title(w world.World) string {
	if w.Name != "" {
		return w.Name
	}

	return w.UWP.String()
}

func joinOrNone(items []string) string {
	if len(items) == 0 {
		return "None"
	}

	return strings.Join(items, " ")
}

// travelZoneName returns the zone's display name and true, or false if z
// is the zero value (Generate never sets TravelZone today, so this is the
// common case a caller must handle, not a malformed-data edge case).
func travelZoneName(z world.TravelZone) (string, bool) {
	switch z {
	case world.ZoneGreen:
		return "Green", true
	case world.ZoneAmber:
		return "Amber", true
	case world.ZoneRed:
		return "Red", true
	default:
		return "", false
	}
}
