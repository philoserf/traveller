package render

import (
	"fmt"
	"strings"

	"github.com/philoserf/traveller/sector"
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
