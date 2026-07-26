package starship

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"testing"
)

// declaredDriveTypes reads drive.go and returns the value of every
// constant declared with type DriveType, keyed by its Go identifier.
//
// Reading the source is the only way to enumerate declared constants in
// Go: they leave no runtime trace, so a test that lists them by hand can
// only ever check the ones its author remembered. An earlier version of
// this file did exactly that, and adding a DriveType to the const block
// alone left every test in it green — the omission it claimed to catch
// was the one thing it could not see.
func declaredDriveTypes(t *testing.T) map[string]DriveType {
	t.Helper()

	file, err := parser.ParseFile(token.NewFileSet(), "drive.go", nil, 0)
	if err != nil {
		t.Fatalf("parsing drive.go: %v", err)
	}

	declared := map[string]DriveType{}

	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}

		for _, spec := range gen.Specs {
			collectDriveTypeSpec(t, spec, declared)
		}
	}

	if len(declared) == 0 {
		t.Fatal("found no DriveType constants in drive.go — the parser found nothing to check")
	}

	return declared
}

// collectDriveTypeSpec adds spec's constants into declared when spec is
// a DriveType declaration, and does nothing otherwise — the const blocks
// in drive.go also declare DriveCategory and StageEffect.
func collectDriveTypeSpec(t *testing.T, spec ast.Spec, declared map[string]DriveType) {
	t.Helper()

	value, ok := spec.(*ast.ValueSpec)
	if !ok {
		return
	}

	if ident, ok := value.Type.(*ast.Ident); !ok || ident.Name != "DriveType" {
		return
	}

	for i, name := range value.Names {
		lit, ok := value.Values[i].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			t.Fatalf("%s: DriveType constants must be string literals, got %v", name.Name, value.Values[i])
		}

		code, err := strconv.Unquote(lit.Value)
		if err != nil {
			t.Fatalf("%s: unquoting %s: %v", name.Name, lit.Value, err)
		}

		declared[name.Name] = DriveType(code)
	}
}

// TestEveryDeclaredDriveTypeIsCategorized is the guarantee the rest of
// this file depends on: every DriveType the const block declares appears
// in driveCategories. Without it, Category's fallback silently reports a
// newly declared drive as a maneuver drive, and no test notices.
func TestEveryDeclaredDriveTypeIsCategorized(t *testing.T) {
	t.Parallel()

	for name, code := range declaredDriveTypes(t) {
		if _, ok := driveCategories[code]; !ok {
			t.Errorf("%s (%q) is declared but missing from driveCategories — Category would silently "+
				"report it as a maneuver drive", name, code)
		}
	}
}

// TestDriveCategoriesHasNoStrayEntries is the converse: nothing in the
// map that isn't a declared constant. A stray entry would be dead data,
// and more likely a typo'd code that leaves the real drive uncategorized
// while this file's other tests still pass.
func TestDriveCategoriesHasNoStrayEntries(t *testing.T) {
	t.Parallel()

	declared := map[DriveType]bool{}
	for _, code := range declaredDriveTypes(t) {
		declared[code] = true
	}

	for code := range driveCategories {
		if !declared[code] {
			t.Errorf("driveCategories has entry %q, which no DriveType constant declares", code)
		}
	}
}

// TestDriveTypeCategory pins the actual grouping, so the registry can't
// be "complete" but wrong. Expectations are written out by hand on
// purpose — deriving them from driveCategories would make the test
// tautological, asserting only that the map equals itself.
func TestDriveTypeCategory(t *testing.T) {
	t.Parallel()

	want := map[DriveType]DriveCategory{
		DriveGravitic: DriveManeuver,
		DriveRocket:   DriveManeuver,
		DriveHEPlaR:   DriveManeuver,

		DriveJumpJ: DriveJump,
		DriveHop:   DriveJump,
		DriveSkip:  DriveJump,
		DriveNAFAL: DriveJump,

		DrivePowerPlant: DrivePower,
		DriveFission:    DrivePower,
		DriveAntiMatter: DrivePower,
		DriveCollector:  DrivePower,

		DriveBattery:    DriveSupplemental,
		DriveFuelCell:   DriveSupplemental,
		DriveFusionPlus: DriveSupplemental,
	}

	for drive, category := range want {
		if got := drive.Category(); got != category {
			t.Errorf("DriveType(%q).Category() = %d, want %d", drive, got, category)
		}
	}

	// Every declared drive must have an expectation here too, or a new
	// one could be categorized in the map and never actually checked.
	for name, code := range declaredDriveTypes(t) {
		if _, ok := want[code]; !ok {
			t.Errorf("%s (%q) is declared but has no expected category in this test", name, code)
		}
	}
}

// TestDriveTypeCategoryDefaultsToManeuver pins the fallback for a value
// that is not a declared drive at all. Recorded deliberately: it means a
// mistyped code is reported as a maneuver drive rather than rejected,
// which is tolerable only because TestEveryDeclaredDriveTypeIsCategorized
// rules out the case that actually matters.
func TestDriveTypeCategoryDefaultsToManeuver(t *testing.T) {
	t.Parallel()

	for _, notADrive := range []DriveType{"", "Z", "jump", "g"} {
		if got := notADrive.Category(); got != DriveManeuver {
			t.Errorf("DriveType(%q).Category() = %d, want DriveManeuver (the fallback)", notADrive, got)
		}
	}
}

// TestDriveTypeCodesAreDistinct is a backstop, not the primary guard.
// Now that driveCategories must contain every declared drive, two
// sharing a code collide as duplicate keys in that map literal and the
// package no longer compiles ("duplicate key \"G\" in map literal") —
// a stronger check than any test, and one this test cannot reach past.
// It is kept for the case that stops being true: a code duplicated
// between a declared constant and something outside the registry.
func TestDriveTypeCodesAreDistinct(t *testing.T) {
	t.Parallel()

	seen := map[DriveType]string{}

	for name, code := range declaredDriveTypes(t) {
		if other, dup := seen[code]; dup {
			t.Errorf("%s and %s both use the code %q", other, name, code)
		}

		seen[code] = name
	}
}

// TestDriveCategoriesAreDistinct confirms the four categories really are
// four distinct values — they are iota-derived, so this mainly guards
// against an explicit value being pinned onto one later, and against the
// naming trap the const block itself warns about, where DriveJump (a
// category) sits beside DriveJumpJ (a type).
func TestDriveCategoriesAreDistinct(t *testing.T) {
	t.Parallel()

	categories := []DriveCategory{DriveManeuver, DriveJump, DrivePower, DriveSupplemental}

	seen := make(map[DriveCategory]bool, len(categories))
	for _, c := range categories {
		seen[c] = true
	}

	if len(seen) != len(categories) {
		t.Errorf("got %d distinct DriveCategory values, want %d", len(seen), len(categories))
	}
}
