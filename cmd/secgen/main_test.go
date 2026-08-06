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

	r := clitest.Run(t, "-seed", "42", "-subsector", "A", "-format", "short")
	r.AssertExitCode(t, 0)
	r.AssertGolden(t, "seed42-subsector-a-short")
}

func TestInvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"density", []string{"-density", "bogus"}, "secgen: unknown density"},
		{"subsector", []string{"-subsector", "Q"}, "secgen: -subsector must be a single letter A-P"},
		{"format", []string{"-format", "bogus"}, "secgen: -format must be"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := clitest.Run(t, append([]string{"-seed", "42"}, tt.args...)...)
			r.AssertExitCode(t, 1)
			r.AssertStderrContains(t, tt.wantErr)
		})
	}
}
