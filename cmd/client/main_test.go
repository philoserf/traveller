package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/philoserf/traveller/api"
	"github.com/philoserf/traveller/cmd/internal/clitest"
)

func TestMain(m *testing.M) {
	clitest.Main(m, main)
}

// queryRecorder wraps a handler and records the RawQuery of the most
// recent request it served, so tests can assert on what cmd/client
// actually sent over the wire — not just what the server's JSON
// response happens to report back.
type queryRecorder struct {
	http.Handler

	mu        sync.Mutex
	lastQuery string
}

func (q *queryRecorder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	q.mu.Lock()
	q.lastQuery = r.URL.RawQuery
	q.mu.Unlock()

	q.Handler.ServeHTTP(w, r)
}

func (q *queryRecorder) LastQuery() string {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.lastQuery
}

func newTestServer(t *testing.T) (*httptest.Server, *queryRecorder) {
	t.Helper()

	rec := &queryRecorder{Handler: api.NewMux()}
	srv := httptest.NewServer(rec)
	t.Cleanup(srv.Close)

	return srv, rec
}

// TestSeedZeroExplicit is the regression test for the bug this issue
// fixes: an explicit "-seed 0" must reach the server as ?seed=0, and
// the server must honor it (dice.ResolveSeed(&0) == 0), not treat it as
// "no seed given".
func TestSeedZeroExplicit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{"world", []string{"world"}},
		{"system", []string{"system"}},
		{"sector", []string{"sector"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv, rec := newTestServer(t)

			r := clitest.Run(t, append(tt.args, "-server", srv.URL, "-seed", "0")...)
			r.AssertExitCode(t, 0)

			if !strings.Contains(rec.LastQuery(), "seed=0") {
				t.Errorf("client: request query = %q, want it to contain \"seed=0\"", rec.LastQuery())
			}

			if !strings.Contains(r.Stdout, "(seed: 0)") {
				t.Errorf("client: stdout = %q, want it to report \"(seed: 0)\"", r.Stdout)
			}
		})
	}
}

// TestSeedOmitted confirms the fix didn't overcorrect: omitting -seed
// entirely must still ask the server to pick one, not send seed=0.
func TestSeedOmitted(t *testing.T) {
	t.Parallel()

	srv, rec := newTestServer(t)

	r := clitest.Run(t, "world", "-server", srv.URL)
	r.AssertExitCode(t, 0)

	if strings.Contains(rec.LastQuery(), "seed=") {
		t.Errorf("client: request query = %q, want no \"seed=\" param", rec.LastQuery())
	}
}

func TestHealthz(t *testing.T) {
	t.Parallel()

	srv, _ := newTestServer(t)

	r := clitest.Run(t, "healthz", "-server", srv.URL)
	r.AssertExitCode(t, 0)

	if !strings.Contains(r.Stdout, "200 OK") {
		t.Errorf("client: stdout = %q, want it to contain \"200 OK\"", r.Stdout)
	}
}

func TestInvalidFormat(t *testing.T) {
	t.Parallel()

	srv, _ := newTestServer(t)

	r := clitest.Run(t, "sector", "-server", srv.URL, "-format", "bogus")
	r.AssertExitCode(t, 1)
	r.AssertStderrContains(t, `client: -format must be "full" or "short"`)
}

// TestServerError covers the client's own error path when the server
// responds with a non-200 status — a fake handler standing in for a
// misbehaving or overloaded real server, distinct from api.NewMux's own
// validation errors.
func TestServerError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	r := clitest.Run(t, "world", "-server", srv.URL)
	r.AssertExitCode(t, 1)
	r.AssertStderrContains(t, "client: server returned")
}

// TestInvalidSeed covers the flag package's own ExitOnError handling
// for a malformed -seed value, matching the other cmd/* packages'
// coverage of the same boundary.
func TestInvalidSeed(t *testing.T) {
	t.Parallel()

	srv, _ := newTestServer(t)

	r := clitest.Run(t, "world", "-server", srv.URL, "-seed", "notanumber")
	r.AssertExitCode(t, 2)
	r.AssertStderrContains(t, "invalid value")
}

func TestUnknownCommand(t *testing.T) {
	t.Parallel()

	r := clitest.Run(t, "bogus")
	r.AssertExitCode(t, 1)
	r.AssertStderrContains(t, fmt.Sprintf("client: unknown command %q", "bogus"))
}
