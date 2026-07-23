// Command client is a CLI for talking to the traveller API server.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/philoserf/traveller/api"
	"github.com/philoserf/traveller/system"
	"github.com/philoserf/traveller/world"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: client <healthz|world|system|sector> [flags]")
		os.Exit(1)
	}

	var err error

	switch os.Args[1] {
	case "healthz":
		err = runHealthz(os.Args[2:])
	case "world":
		err = runWorld(os.Args[2:])
	case "system":
		err = runSystem(os.Args[2:])
	case "sector":
		err = runSector(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "client: unknown command %q\n", os.Args[1])
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runHealthz(args []string) error {
	fs := flag.NewFlagSet("healthz", flag.ExitOnError)
	addr := fs.String("server", "http://localhost:8080", "traveller API server address")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("client: parsing flags: %w", err)
	}

	status, _, body, err := get(*addr + "/healthz")
	if err != nil {
		return err
	}

	fmt.Printf("client: %s -> %s (%s)\n", *addr, status, body)

	return nil
}

func runWorld(args []string) error {
	fs := flag.NewFlagSet("world", flag.ExitOnError)
	addr := fs.String("server", "http://localhost:8080", "traveller API server address")
	seed := fs.Int64("seed", 0, "seed to request (0 = server picks)")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("client: parsing flags: %w", err)
	}

	url := *addr + "/worlds/random"
	if *seed != 0 {
		url += "?seed=" + strconv.FormatInt(*seed, 10)
	}

	status, statusCode, body, err := get(url)
	if err != nil {
		return err
	}

	if statusCode != http.StatusOK {
		return fmt.Errorf("client: server returned %s: %s", status, body)
	}

	var w api.WorldResponse
	if err := json.Unmarshal(body, &w); err != nil {
		return fmt.Errorf("client: decoding response: %w", err)
	}

	printWorldFields(w.UWP, w.TradeCodes, w.Bases, w.PBG, w.TravelZone, w.Importance, w.Economic, w.Cultural)
	fmt.Printf("(seed: %d)\n", w.Seed)

	return nil
}

func runSystem(args []string) error {
	fs := flag.NewFlagSet("system", flag.ExitOnError)
	addr := fs.String("server", "http://localhost:8080", "traveller API server address")
	seed := fs.Int64("seed", 0, "seed to request (0 = server picks)")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("client: parsing flags: %w", err)
	}

	url := *addr + "/systems/random"
	if *seed != 0 {
		url += "?seed=" + strconv.FormatInt(*seed, 10)
	}

	status, statusCode, body, err := get(url)
	if err != nil {
		return err
	}

	if statusCode != http.StatusOK {
		return fmt.Errorf("client: server returned %s: %s", status, body)
	}

	var sys api.SystemResponse
	if err := json.Unmarshal(body, &sys); err != nil {
		return fmt.Errorf("client: decoding response: %w", err)
	}

	printSystem(sys)

	return nil
}

// printSystem prints one system's Mainworld orbit/UWP/extensions fields
// followed by every star group's bodies and satellites — shared by
// runSystem (one system) and runSector (one per populated hex).
func printSystem(sys api.SystemResponse) {
	mw := sys.Mainworld
	if mw.Satellite {
		orbitKind := "Far"
		if mw.Close {
			orbitKind = "Close"
		}

		fmt.Printf("Mainworld orbit: %d (%s satellite of a Gas Giant)\n", mw.Orbit, orbitKind)
	} else {
		fmt.Printf("Mainworld orbit: %d (%.1f AU)\n", mw.Orbit, mw.AU)
	}

	printWorldFields(mw.UWP, mw.TradeCodes, mw.Bases, mw.PBG, mw.TravelZone, mw.Importance, mw.Economic, mw.Cultural)

	for _, group := range sys.StarGroups {
		fmt.Println(starHeadingLine(group.Star))

		if len(group.Bodies) == 0 {
			fmt.Println("  None.")

			continue
		}

		for _, b := range group.Bodies {
			fmt.Println("  " + bodyLine(b))

			for _, sat := range b.Satellites {
				fmt.Println("    " + satelliteLine(sat))
			}
		}
	}

	fmt.Printf("(seed: %d)\n", sys.Seed)
}

func runSector(args []string) error {
	fs := flag.NewFlagSet("sector", flag.ExitOnError)
	addr := fs.String("server", "http://localhost:8080", "traveller API server address")
	seed := fs.Int64("seed", 0, "seed to request (0 = server picks)")
	name := fs.String("name", "", "sector name")
	density := fs.String("density", "", "System Presence density (default: Standard)")
	subsector := fs.String("subsector", "", "single letter A-P — limit output to that 80-hex block only")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("client: parsing flags: %w", err)
	}

	query := neturl.Values{}
	if *seed != 0 {
		query.Set("seed", strconv.FormatInt(*seed, 10))
	}

	if *name != "" {
		query.Set("name", *name)
	}

	if *density != "" {
		query.Set("density", *density)
	}

	if *subsector != "" {
		query.Set("subsector", *subsector)
	}

	url := *addr + "/sectors/random"
	if encoded := query.Encode(); encoded != "" {
		url += "?" + encoded
	}

	status, statusCode, body, err := get(url)
	if err != nil {
		return err
	}

	if statusCode != http.StatusOK {
		return fmt.Errorf("client: server returned %s: %s", status, body)
	}

	var sec api.SectorResponse
	if err := json.Unmarshal(body, &sec); err != nil {
		return fmt.Errorf("client: decoding response: %w", err)
	}

	fmt.Printf("%s Sector\n", sec.Name)

	for _, hex := range sec.Hexes {
		if hex.System == nil {
			fmt.Printf("Hex %s: empty\n", hex.Location)

			continue
		}

		fmt.Printf("Hex %s\n", hex.Location)
		printSystem(*hex.System)
	}

	fmt.Printf("(seed: %d)\n", sec.Seed)

	return nil
}

// bodyLine formats one non-star, non-satellite body: a Gas Giant, or a
// placed World with its own Trade Codes — matching render/system.go's
// otherBodyLine, including its "(Mainworld)" suffix.
func bodyLine(b api.BodyResponse) string {
	var line string
	if b.GasGiant != nil {
		line = fmt.Sprintf("Orbit %d: Gas Giant, Size %s (%s)", b.Orbit, b.GasGiant.Size, b.GasGiant.Bracket)
	} else {
		line = fmt.Sprintf("Orbit %d: %s — %s", b.Orbit, b.UWP, strings.Join(world.TradeCodeStrings(b.TradeCodes), " "))
	}

	if b.Ring {
		line += ", with a Ring"
	}

	if b.IsMainworld {
		line += " (Mainworld)"
	}

	return line
}

// satelliteLine formats one satellite — matching render/system.go's
// satelliteLine, including its "(Mainworld)" suffix.
func satelliteLine(s api.SatelliteResponse) string {
	orbit := "Far"
	if s.Close {
		orbit = "Close"
	}

	line := fmt.Sprintf("%s satellite: %s — %s", orbit, s.UWP, strings.Join(world.TradeCodeStrings(s.TradeCodes), " "))

	if s.IsMainworld {
		line += " (Mainworld)"
	}

	return line
}

// printWorldFields prints the fields api.WorldResponse and
// api.MainworldResponse share, so runWorld and runSystem don't each keep
// their own copy of this Printf block.
func printWorldFields(
	uwp string, tradeCodes []world.TradeCode, bases []world.Base, pbg, travelZone string,
	importance int, econ api.EconomicResponse, cult api.CulturalResponse,
) {
	fmt.Printf("UWP: %s\n", uwp)
	fmt.Printf("Trade Codes: %s\n", strings.Join(world.TradeCodeStrings(tradeCodes), " "))
	fmt.Printf("Bases: %s\n", strings.Join(world.BaseStrings(bases), " "))
	fmt.Printf("PBG: %s\n", pbg)
	fmt.Printf("Travel Zone: %s\n", travelZone)
	fmt.Printf("Importance: %+d\n", importance)
	fmt.Printf("Economic: Resources=%d Labor=%d Infrastructure=%d Efficiency=%+d\n",
		econ.Resources, econ.Labor, econ.Infrastructure, econ.Efficiency)
	fmt.Printf("Cultural: Heterogeneity=%d Acceptance=%d Strangeness=%d Symbols=%d\n",
		cult.Heterogeneity, cult.Acceptance, cult.Strangeness, cult.Symbols)
}

// starHeadingLine formats one star's own group heading — matching
// render/system.go's starHeading: Degenerate stars (white dwarfs/brown
// dwarfs) omit SpectralDecimal, since api.StarResponse's SpectralDecimal
// is meaningless for them (mirroring system.Star's own doc comment), and
// the orbit part is omitted entirely for the Primary (s.Orbit is nil —
// see StarResponse's own doc comment on the sentinel this maps from).
func starHeadingLine(s api.StarResponse) string {
	spec := fmt.Sprintf("%s%d %s", s.SpectralType, s.SpectralDecimal, s.LuminosityClass)
	if s.SpectralType == string(system.SpectralDegenerate) {
		spec = fmt.Sprintf("%s %s", s.SpectralType, s.LuminosityClass)
	}

	var orbitPart string
	if s.Orbit != nil {
		orbitPart = fmt.Sprintf("Orbit %d, ", *s.Orbit)
	}

	companion := ""
	if s.HasCompanion {
		companion = ", with a Companion"
	}

	return fmt.Sprintf("%s: %s (%sHZ orbit %d)%s", s.Role, spec, orbitPart, s.HabitableZoneOrbit, companion)
}

// get performs a GET request against url. It fully drains and closes the
// response body itself, returning only the pieces callers need — never a
// live *http.Response — so there's no way for a caller to forget to close
// it. status is the server's raw status line (e.g. "200 OK"), kept
// alongside statusCode rather than reconstructed via http.StatusText:
// StatusText returns "" for any status code outside net/http's built-in
// table, which would silently drop the reason phrase for a non-standard
// code from a proxy or future server change.
func get(url string) (string, int, []byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", 0, nil, fmt.Errorf("client: building request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", 0, nil, fmt.Errorf("client: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, nil, fmt.Errorf("client: reading response: %w", err)
	}

	return resp.Status, resp.StatusCode, body, nil
}
