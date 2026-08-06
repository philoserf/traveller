package clitest

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "update clitest golden files instead of comparing against them")

// AssertGolden compares r.Stdout against testdata/<name>.golden,
// relative to the calling test's package directory. Run with -update
// to (re)write the fixture instead of comparing, e.g.
// `go test ./cmd/worldgen/... -run TestGolden -update`. Review the diff
// before committing an updated fixture — a golden change silently
// encodes a generation-output change, and nothing else distinguishes
// "expected" from "regression".
func (r Result) AssertGolden(t *testing.T, name string) {
	t.Helper()

	path := filepath.Join("testdata", name+".golden")

	if *update {
		if err := os.MkdirAll("testdata", 0o750); err != nil {
			t.Fatalf("clitest: creating testdata dir: %v", err)
		}

		if err := os.WriteFile(path, []byte(r.Stdout), 0o600); err != nil {
			t.Fatalf("clitest: writing golden file %s: %v", path, err)
		}

		return
	}

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("clitest: reading golden file %s (run with -update to create it): %v", path, err)
	}

	if r.Stdout != string(want) {
		t.Errorf(
			"clitest: stdout does not match %s (run with -update to review/accept)\n--- got ---\n%s\n--- want ---\n%s",
			path,
			r.Stdout,
			string(want),
		)
	}
}
