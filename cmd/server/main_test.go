package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/philoserf/traveller/cmd/internal/clitest"
)

func TestMain(m *testing.M) {
	clitest.Main(m, main)
}

// testAddr is a fixed, atypical port rather than ":0": cmd/server never
// reports back which port an ephemeral bind resolved to (Addr isn't
// rewritten after Listen), and this test runs in a separate process
// from the server it starts, so there's no in-process way to recover
// it. Collision risk is accepted the same way a fixed 8080 default is.
const testAddr = "127.0.0.1:18080"

func TestHealthz(t *testing.T) {
	t.Parallel()

	clitest.RunBackground(t, "-addr", testAddr)

	deadline := time.Now().Add(5 * time.Second)

	var lastErr error

	for time.Now().Before(deadline) {
		if err := pollHealthz(); err != nil {
			lastErr = err

			time.Sleep(50 * time.Millisecond)

			continue
		}

		return
	}

	t.Fatalf("server: never became healthy at %s: %v", testAddr, lastErr)
}

func pollHealthz() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+testAddr+"/healthz", nil)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("healthz returned %s", resp.Status)
	}

	return nil
}

func TestInvalidAddr(t *testing.T) {
	t.Parallel()

	r := clitest.Run(t, "-addr", "not-a-valid-address")
	r.AssertExitCode(t, 1)
	r.AssertStderrContains(t, "listen")
}
