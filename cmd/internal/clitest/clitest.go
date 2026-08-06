// Package clitest is a subprocess-based harness for testing cmd/*
// commands end to end. Main re-execs the test binary itself as the
// command under test (the same trick os/exec's own tests use for a
// "helper process": TestMain checks an env var and, if set, calls the
// package's real main directly instead of running go test's usual
// m.Run()) — so Run exercises the actual flag-parsing/stdout/stderr/
// exit-code boundary, not an in-process call to an internal function.
package clitest

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// runTimeout bounds how long a single Run subprocess may run before
// it's killed — generous for a CLI command, but short enough that a
// hung child (e.g. a command that unexpectedly blocks on stdin) fails
// the test instead of hanging `go test` indefinitely.
const runTimeout = 30 * time.Second

const (
	childEnvVar = "TRAVELLER_CLITEST_CHILD"
	childEnvVal = "1"
)

// Main is the TestMain a cmd/* package's own main_test.go should
// delegate to: `func TestMain(m *testing.M) { clitest.Main(m, main) }`.
// In the child process started by Run, childEnvVar is set, so realMain
// (the package's own main) runs directly and the process exits without
// ever entering go test's usual m.Run() — realMain may itself call
// os.Exit on an error path, which is fine: that only ends this child
// process, not the parent test process that's waiting on it.
//
// Skipping m.Run() in the child path matters beyond speed: testing's
// own flag.Parse() call (registering -test.v, -test.run, etc.) lives
// inside m.Run(), not before TestMain runs. Since the child never calls
// m.Run(), that parse never happens, and realMain's own flag.Parse()
// (e.g. dice.SeedFlag(), or cmd/server's -addr) is the first and only
// one to consume args — so Run/RunBackground's args reach the command
// under test exactly as if a user had typed them.
func Main(m *testing.M, realMain func()) {
	if os.Getenv(childEnvVar) == childEnvVal {
		realMain()
		os.Exit(0)
	}

	os.Exit(m.Run())
}

// Result is the outcome of one Run: captured stdout/stderr and the
// process's exit code.
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Run re-execs the test binary with args, as if args were passed to the
// cmd/* package's own main. Panics/build failures aside, this never
// fails the test itself on a nonzero exit — that's a normal, assertable
// outcome (AssertExitCode) for a command under test, not a harness
// error.
func Run(t *testing.T, args ...string) Result {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), runTimeout)
	defer cancel()

	//nolint:gosec // os.Args[0] is this test binary's own path, not external input.
	cmd := exec.CommandContext(ctx, os.Args[0], args...)

	cmd.Env = append(os.Environ(), childEnvVar+"="+childEnvVal)

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	exitCode := 0

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError

		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("clitest: running subprocess: %v", err)
		}
	}

	return Result{Stdout: stdout.String(), Stderr: stderr.String(), ExitCode: exitCode}
}

// Background is a subprocess started by RunBackground and left running
// — for a command like cmd/server that doesn't exit on its own.
type Background struct {
	cmd *exec.Cmd
}

// RunBackground starts the test binary re-exec'd as the command under
// test (same mechanism as Run) but does not wait for it to exit; the
// caller is responsible for polling whatever readiness signal the
// command under test exposes. The process is killed via t.Cleanup.
func RunBackground(t *testing.T, args ...string) *Background {
	t.Helper()

	//nolint:gosec // os.Args[0] is this test binary's own path, not external input.
	cmd := exec.CommandContext(context.Background(), os.Args[0], args...)

	cmd.Env = append(os.Environ(), childEnvVar+"="+childEnvVal)

	if err := cmd.Start(); err != nil {
		t.Fatalf("clitest: starting background subprocess: %v", err)
	}

	bg := &Background{cmd: cmd}
	t.Cleanup(bg.Kill)

	return bg
}

// Kill terminates the background process and waits for it to exit.
// Safe to call more than once (via t.Cleanup and explicitly).
func (b *Background) Kill() {
	if b.cmd.Process == nil {
		return
	}

	_ = b.cmd.Process.Kill()
	_ = b.cmd.Wait()
}

// AssertExitCode fails the test if r.ExitCode != want.
func (r Result) AssertExitCode(t *testing.T, want int) {
	t.Helper()

	if r.ExitCode != want {
		t.Errorf("clitest: exit code = %d, want %d\nstdout:\n%s\nstderr:\n%s", r.ExitCode, want, r.Stdout, r.Stderr)
	}
}

// AssertStderrContains fails the test if r.Stderr does not contain substr.
func (r Result) AssertStderrContains(t *testing.T, substr string) {
	t.Helper()

	if !strings.Contains(r.Stderr, substr) {
		t.Errorf("clitest: stderr does not contain %q\nstderr:\n%s", substr, r.Stderr)
	}
}
