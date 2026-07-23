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
// world.Generate currently populates (UWP, TradeCodes, Bases, PBG,
// TravelZone, and the Importance/Economic/Cultural extensions). The title falls
// back to the UWP code when Name is empty — Generate never sets Name.
// Travel Zone is shown only when TravelZone.String() is non-empty: real
// generated worlds always have one, but this keeps World's zero value
// (e.g. from a hand-built fixture, or a future partial construction)
// rendering cleanly instead of showing a blank label. Sector, Hex, Worlds,
// Notes, Nobility, and Allegiance are never rendered at all — see
// world/generate.go's own doc comment for what's not generated, and why.
func World(w world.World) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# %s\n\n", title(w))
	fmt.Fprintf(&b, "**UWP:** %s\n\n", w.UWP)
	fmt.Fprintf(&b, "**Trade Codes:** %s\n\n", joinOrNone(world.TradeCodeStrings(w.TradeCodes)))
	fmt.Fprintf(&b, "**Bases:** %s\n\n", joinOrNone(world.BaseStrings(w.Bases)))
	fmt.Fprintf(&b, "**PBG:** %s\n\n", w.PBG)

	if zone := w.TravelZone.String(); zone != "" {
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
