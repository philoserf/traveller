package main

import (
	"testing"

	"github.com/philoserf/traveller/cmd/internal/clitest"
)

func TestMain(m *testing.M) {
	clitest.Main(m, main)
}

func TestGolden(t *testing.T) {
	t.Parallel()

	r := clitest.Run(t, "-seed", "42")
	r.AssertExitCode(t, 0)
	r.AssertGolden(t, "seed42")
}

func TestInvalidSeed(t *testing.T) {
	t.Parallel()

	r := clitest.Run(t, "-seed", "notanumber")
	r.AssertExitCode(t, 2)
	r.AssertStderrContains(t, "invalid value")
}
