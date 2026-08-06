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

	r := clitest.Run(t, "-career", "scout", "-seed", "42")
	r.AssertExitCode(t, 0)
	r.AssertGolden(t, "scout-seed42")
}

func TestInvalidCareer(t *testing.T) {
	t.Parallel()

	r := clitest.Run(t, "-career", "bogus", "-seed", "42")
	r.AssertExitCode(t, 1)
	r.AssertStderrContains(t, `chargen: -career must be`)
}

func TestInvalidAge(t *testing.T) {
	t.Parallel()

	r := clitest.Run(t, "-career", "scout", "-age", "-1", "-seed", "42")
	r.AssertExitCode(t, 1)
	r.AssertStderrContains(t, "chargen: -age must not be negative")
}
