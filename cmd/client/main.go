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
		fmt.Fprintln(os.Stderr, "usage: client <healthz|world> [flags]")
		os.Exit(1)
	}

	var err error

	switch os.Args[1] {
	case "healthz":
		err = runHealthz(os.Args[2:])
	case "world":
		err = runWorld(os.Args[2:])
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

	fmt.Printf("UWP: %s\n", w.UWP)
	fmt.Printf("Trade Codes: %s\n", strings.Join(world.TradeCodeStrings(w.TradeCodes), " "))
	fmt.Printf("Bases: %s\n", strings.Join(world.BaseStrings(w.Bases), " "))
	fmt.Printf("PBG: %s\n", w.PBG)
	fmt.Printf("Importance: %+d\n", w.Importance)
	fmt.Printf("Economic: Resources=%d Labor=%d Infrastructure=%d Efficiency=%+d\n",
		w.Economic.Resources, w.Economic.Labor, w.Economic.Infrastructure, w.Economic.Efficiency)
	fmt.Printf("Cultural: Heterogeneity=%d Acceptance=%d Strangeness=%d Symbols=%d\n",
		w.Cultural.Heterogeneity, w.Cultural.Acceptance, w.Cultural.Strangeness, w.Cultural.Symbols)
	fmt.Printf("(seed: %d)\n", w.Seed)

	return nil
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
