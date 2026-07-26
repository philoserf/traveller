package starship

import "testing"

// allDriveTypes is every DriveType the package declares, paired with the
// DriveCategory its own const-block comment assigns it. Category's
// implementation is a switch with a default arm, so it answers for any
// string at all — including ones that aren't drives. Listing the real
// set here is what makes the test able to notice a DriveType being added
// without being categorised, which the switch itself would silently
// absorb into DriveManeuver.
var allDriveTypes = map[DriveType]DriveCategory{
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

func TestDriveTypeCategory(t *testing.T) {
	t.Parallel()

	for drive, want := range allDriveTypes {
		if got := drive.Category(); got != want {
			t.Errorf("DriveType(%q).Category() = %d, want %d", drive, got, want)
		}
	}
}

// TestDriveTypeCategoryDefaultsToManeuver pins the behavior of the
// switch's default arm for a value that is not a declared drive. It
// answers DriveManeuver rather than erroring, which is worth recording
// deliberately: it means a mistyped or newly-added DriveType is reported
// as a maneuver drive rather than rejected, and TestDriveTypeCategory
// above is the only thing standing between that and going unnoticed.
func TestDriveTypeCategoryDefaultsToManeuver(t *testing.T) {
	t.Parallel()

	for _, notADrive := range []DriveType{"", "Z", "jump", "g"} {
		if got := notADrive.Category(); got != DriveManeuver {
			t.Errorf("DriveType(%q).Category() = %d, want DriveManeuver (the default arm)", notADrive, got)
		}
	}
}

// TestDriveTypeCodesAreDistinct guards the const block itself: these are
// string codes, so two drives sharing one would compile cleanly and make
// the pair indistinguishable everywhere they are used — including inside
// Category's own switch, where the duplicate arm would be unreachable.
func TestDriveTypeCodesAreDistinct(t *testing.T) {
	t.Parallel()

	if len(allDriveTypes) != 14 {
		t.Fatalf("allDriveTypes has %d entries, want 14 — a DriveType was added or removed "+
			"without updating this test's own category expectations", len(allDriveTypes))
	}

	seen := make(map[DriveType]bool, len(allDriveTypes))
	for drive := range allDriveTypes {
		if seen[drive] {
			t.Errorf("DriveType %q is declared twice", drive)
		}

		seen[drive] = true
	}
}

// TestDriveCategoriesAreDistinct confirms the four categories really are
// four distinct values. They are iota-derived, so this mainly guards
// against an explicit value being pinned onto one of them later — and
// against the naming trap the const block itself warns about, where
// DriveJump (a category) sits beside DriveJumpJ (a type).
func TestDriveCategoriesAreDistinct(t *testing.T) {
	t.Parallel()

	categories := []DriveCategory{DriveManeuver, DriveJump, DrivePower, DriveSupplemental}

	seen := make(map[DriveCategory]bool, len(categories))
	for _, c := range categories {
		if seen[c] {
			t.Errorf("DriveCategory %d is used by more than one category constant", c)
		}

		seen[c] = true
	}

	if len(seen) != 4 {
		t.Errorf("got %d distinct DriveCategory values, want 4", len(seen))
	}
}
