// Command client is a CLI for talking to the traveller API server.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/philoserf/traveller/api"
	"github.com/philoserf/traveller/world"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: client <healthz|world|system> [flags]")
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

	for _, s := range sys.Stars {
		fmt.Println(starLine(s))
	}

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

	multiStar := len(sys.Stars) > 1
	for _, o := range sys.OtherBodies {
		fmt.Println(otherBodyLine(o, multiStar))
	}

	fmt.Printf("(seed: %d)\n", sys.Seed)

	return nil
}

// otherBodyLine formats one non-mainworld, non-star body: a Gas Giant, or
// a placed World with its own Trade Codes — matching render/system.go's
// otherBodyLine, including its "(hosted by <Role>)" suffix once more than
// one star is present (o.HostRole alone doesn't say that on its own).
func otherBodyLine(o api.OtherBodyResponse, multiStar bool) string {
	line := fmt.Sprintf("Orbit %d: %s — %s", o.Orbit, o.UWP, strings.Join(world.TradeCodeStrings(o.TradeCodes), " "))
	if o.GasGiant != nil {
		line = fmt.Sprintf("Orbit %d: Gas Giant, Size %s (%s)", o.Orbit, o.GasGiant.Size, o.GasGiant.Bracket)
	}

	if multiStar {
		line += fmt.Sprintf(" (hosted by %s)", o.HostRole)
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

// starLine formats one star for display, matching render/system.go's
// starLine: Degenerate stars (white dwarfs/brown dwarfs) omit
// SpectralDecimal, since api.StarResponse's SpectralDecimal is
// meaningless for them (mirroring world.Star's own doc comment).
func starLine(s api.StarResponse) string {
	spec := fmt.Sprintf("%s%d %s", s.SpectralType, s.SpectralDecimal, s.LuminosityClass)
	if s.SpectralType == string(world.SpectralDegenerate) {
		spec = fmt.Sprintf("%s %s", s.SpectralType, s.LuminosityClass)
	}

	orbit := "center"
	if s.Orbit != nil {
		orbit = fmt.Sprintf("orbit %d", *s.Orbit)
	}

	companion := ""
	if s.HasCompanion {
		companion = ", with a Companion"
	}

	return fmt.Sprintf("%s: %s (%s, HZ orbit %d)%s", s.Role, spec, orbit, s.HabitableZoneOrbit, companion)
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
