package render

import (
	"fmt"
	"strings"

	"github.com/philoserf/traveller/sector"
	"github.com/philoserf/traveller/world"
)

// Sector renders sec as a Markdown document: a title, then every Hex in
// order — "**Hex CCRR:** empty" for an empty hex, or a "**Hex CCRR**"
// locator line followed by that hex's full System(...) output. Flat
// concatenation, not heading-depth composition: this is exactly what
// running cmd/sysgen repeatedly already produces (a sequence of
// independent "# ... System" blocks), so System itself needs no changes
// to compose here.
func Sector(sec sector.Sector) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# %s Sector\n\n", sec.Name)

	for _, hex := range sec.Hexes {
		if hex.System == nil {
			fmt.Fprintf(&b, "**Hex %s:** empty\n\n", hex.Location)

			continue
		}

		fmt.Fprintf(&b, "**Hex %s**\n\n", hex.Location)
		b.WriteString(System(*hex.System))
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n") + "\n"
}

// SectorCompact renders sec as a Markdown table, one row per populated
// Hex — the mainworld's UWP, Trade Codes, Bases, PBG, and Travel Zone
// only, none of the star/orbit/satellite detail Sector's full System(...)
// output includes. Empty hexes are omitted entirely: a "map overview" is
// meant to be scanned for what's actually there, and most sectors are
// mostly empty space. Bases and Travel Zone use "-" for an absent/Green
// field (world.BasesOrDash/TravelZoneOrDash) rather than the "None"/"Green"
// package render's prose output uses — this table follows Book 3 p.21's
// own Second Survey Format convention instead, which is a table too.
func SectorCompact(sec sector.Sector) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# %s Sector (compact)\n\n", sec.Name)
	fmt.Fprint(&b, "| Hex  | UWP       | Trade Codes | Bases | PBG | Zone  |\n")
	fmt.Fprint(&b, "| ---- | --------- | ----------- | ----- | --- | ----- |\n")

	for _, hex := range sec.Hexes {
		if hex.System == nil {
			continue
		}

		mw := hex.System.Orbits[hex.System.MainworldOrbit].World

		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s |\n",
			hex.Location, mw.UWP,
			world.JoinOrNone(world.TradeCodeStrings(mw.TradeCodes)),
			world.BasesOrDash(mw.Bases),
			mw.PBG, world.TravelZoneOrDash(mw.TravelZone.String()))
	}

	return b.String()
}
