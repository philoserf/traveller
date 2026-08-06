package main

import (
	"testing"

	"github.com/philoserf/traveller/cmd/internal/clitest"
)

func TestMain(m *testing.M) {
	clitest.Main(m, main)
}

// TestNotYetImplemented is shipgen's only case until #6 (starship
// generation) lands — there is no success path to test yet.
func TestNotYetImplemented(t *testing.T) {
	t.Parallel()

	r := clitest.Run(t)
	r.AssertExitCode(t, 1)
	r.AssertStderrContains(t, "shipgen: not yet implemented")
}
