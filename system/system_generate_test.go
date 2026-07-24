package system

import (
	"math/rand/v2"
	"reflect"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/world"
)

func TestGenerateSystemDeterminism(t *testing.T) {
	t.Parallel()

	r1 := dice.New(rand.NewPCG(7, 7))
	r2 := dice.New(rand.NewPCG(7, 7))

	mw1 := world.Generate(r1)
	mw2 := world.Generate(r2)

	sys1 := GenerateSystem(r1, mw1)
	sys2 := GenerateSystem(r2, mw2)

	if !reflect.DeepEqual(sys1, sys2) {
		t.Errorf("identical seeds produced different systems:\n%+v\nvs\n%+v", sys1, sys2)
	}
}

func TestGenerateSystemMainworldOrbitIndexIsCorrect(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(20, 20))
	mw := world.Generate(r)
	sys := GenerateSystem(r, mw)

	got := sys.Orbits[sys.MainworldOrbit].World
	if got == nil || got.UWP != mw.UWP {
		t.Errorf("Orbits[MainworldOrbit].World = %+v, want a world with UWP %s", got, mw.UWP)
	}
}

// TestGenerateSystemInvariants runs many seeds and checks every invariant
// that must hold regardless of what the dice produced: exactly one
// Primary-role star at primaryOrbitNumber, a valid mainworld orbit index,
// a Gas Giant count that never exceeds PBG.GasGiants, no duplicate orbit
// numbers among non-Satellite entries (the collision guard placeInOrbit
// exists for), and a mainworld orbit number that's always non-negative
// and never collides with primaryOrbitNumber's sentinel — a regression
// guard for a real bug: an M-type primary's hzOrbit=0 plus a negative
// HZVar used to produce orbit -1, exactly primaryOrbitNumber.
func TestGenerateSystemInvariants(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(42, 42))

	for range 2000 {
		mw := world.Generate(r)
		sys := GenerateSystem(r, mw)

		if sys.MainworldOrbit < 0 || sys.MainworldOrbit >= len(sys.Orbits) {
			t.Fatalf("MainworldOrbit = %d out of range for %d orbits", sys.MainworldOrbit, len(sys.Orbits))
		}

		if n := sys.Orbits[sys.MainworldOrbit].Number; n < 0 {
			t.Fatalf("mainworld orbit number = %d, want >= 0 (and not primaryOrbitNumber's sentinel)", n)
		}

		primaryCount, gasGiantCount := scanSystemOrbits(t, sys)

		if primaryCount != 1 {
			t.Fatalf("found %d Primary-orbit stars, want exactly 1", primaryCount)
		}

		// The mainworld BigWorld fallback ("If Satellite and No Giants,
		// place a BigWorld in MW Orbit") guarantees a satellite mainworld
		// only ever exists when its own (possibly regenerated)
		// PBG.GasGiants is >= 1 — so the system's own final mainworld's
		// PBG.GasGiants (not the pre-GenerateSystem mw passed in above,
		// which the fallback can replace entirely) is always a hard
		// ceiling on the total Gas Giant count, satellite host included.
		maxGG := int(sys.Orbits[sys.MainworldOrbit].World.PBG.GasGiants)

		if gasGiantCount > maxGG {
			t.Fatalf("found %d Gas Giants, want at most %d", gasGiantCount, maxGG)
		}

		assertNoOrphanedSatellites(t, sys)
		assertNoPrecludedOrbitViolations(t, sys)
	}
}

// assertNoPrecludedOrbitViolations fails t if any non-Star, non-Satellite
// entry sits at or below its host star's own precluded-orbit ceiling
// (precludedOrbitHost) — the floor placeInOrbit's minOrbit exists to
// enforce. Keyed by Orbit.HostRole, not HostHZOrbit: a system has at most
// one Star per StellarRole (sys.Stars() never returns two Primaries, two
// Closes, ...), so this lookup has no collision case to guard against,
// unlike the HZ-orbit value, which distinct stars can share.
func assertNoPrecludedOrbitViolations(t *testing.T, sys StarSystem) {
	t.Helper()

	minOrbitByRole := map[StellarRole]int{}

	for _, star := range sys.Stars() {
		minOrbitByRole[star.Role] = precludedOrbitHost(*star)
	}

	for _, o := range sys.Orbits {
		if o.Satellite || o.Star != nil || (o.World == nil && o.GasGiant == nil) {
			continue
		}

		if minOrbit, ok := minOrbitByRole[o.HostRole]; ok && o.Number < minOrbit {
			t.Fatalf(
				"orbit %d (host=%v) is below its host's precluded-orbit floor %d",
				o.Number,
				o.HostRole,
				minOrbit,
			)
		}
	}
}

// TestGenerateSystemAvoidsPrecludedOrbitsForGiantPrimary pins a real
// generated case with a giant Primary (seed 21: A1 II — Book 3 p.20's
// Stellar Surface table precludes orbit 0 for this combination), rather
// than only the synthetic table tests in system_tables_precluded_test.go.
func TestGenerateSystemAvoidsPrecludedOrbitsForGiantPrimary(t *testing.T) {
	t.Parallel()

	r := dice.RollerFromSeed(21)
	mw := world.Generate(r)
	sys := GenerateSystem(r, mw)

	primary := sys.Stars()[0]
	if primary.SpectralType != SpectralA || primary.SpectralDecimal != 1 || primary.LuminosityClass != "II" {
		t.Fatalf("seed 21's Primary changed (%s%d %s) — this test needs re-pinning to a fresh giant-primary seed",
			string(primary.SpectralType), primary.SpectralDecimal, primary.LuminosityClass)
	}

	for _, o := range sys.Orbits {
		if o.Satellite || o.Star != nil || o.HostRole != Primary {
			continue
		}

		if o.Number < 1 {
			t.Fatalf("orbit %d hosted by the Primary (A1 II, precluded ceiling 0), want >= 1", o.Number)
		}
	}
}

// TestGenerateSystemPreservesMainworldSatelliteCloseFar pins two real
// generated cases (seeds found by search, matching this file's other
// pinned-seed tests) — one where the mainworld's Book 3 Table 2C roll
// (rollMainworldPlacementKind) lands mainworldCloseSatellite, one
// mainworldFarSatellite — and confirms placeMainworld's own Orbit.Close
// reflects that roll rather than defaulting to false for both.
func TestGenerateSystemPreservesMainworldSatelliteCloseFar(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		seed  int64
		close bool
	}{
		{"close satellite mainworld", 5, true},
		{"far satellite mainworld", 15, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			r := dice.RollerFromSeed(c.seed)
			mw := world.Generate(r)
			sys := GenerateSystem(r, mw)

			mwOrbit := sys.Orbits[sys.MainworldOrbit]
			if !mwOrbit.Satellite {
				t.Fatalf("seed %d's mainworld is no longer a satellite — this test needs re-pinning to a fresh seed",
					c.seed)
			}

			if mwOrbit.Close != c.close {
				t.Errorf("seed %d: mainworld satellite Orbit.Close = %v, want %v", c.seed, mwOrbit.Close, c.close)
			}
		})
	}
}

// TestGenerateSystemMainworldFallsBackToBigWorld pins seed 34 (found by
// search: its Table 2C roll says the mainworld is a satellite, but its
// own rolled PBG.GasGiants is 0) — confirming placeMainworld's "If
// Satellite and No Giants, place a BigWorld in MW Orbit" fallback (Book 3
// p.24) actually fires: the mainworld ends up an ordinary (non-satellite)
// planet with a regenerated Size. Checked against rollBigWorldSize's own
// mechanical floor (2D+7, TwoD6 min 2, so >= 9) rather than the book's
// separate "Siz=B+ is BW" labeling threshold (>= 11) — a legitimate
// 2D+7 roll of 9 or 10 is still exactly what this fallback should
// produce, even though the book wouldn't print "BW" next to it.
func TestGenerateSystemMainworldFallsBackToBigWorld(t *testing.T) {
	t.Parallel()

	r := dice.RollerFromSeed(34)
	mw := world.Generate(r)

	if mw.PBG.GasGiants != 0 {
		t.Fatalf("seed 34's mainworld PBG.GasGiants = %d, want 0 — this test needs re-pinning to a fresh seed",
			mw.PBG.GasGiants)
	}

	sys := GenerateSystem(r, mw)
	mwOrbit := sys.Orbits[sys.MainworldOrbit]

	if mwOrbit.Satellite {
		t.Fatalf("seed 34: mainworld is still a Satellite, want the BigWorld fallback to have converted it to a planet")
	}

	if mwOrbit.World.UWP == mw.UWP {
		t.Fatalf("seed 34: mainworld UWP unchanged (%s), want a regenerated BigWorld UWP", mw.UWP)
	}

	if mwOrbit.World.UWP.Size < 9 {
		t.Errorf("seed 34: mainworld Size = %s, want >= 9 (2D+7 floor: TwoD6 min 2, +7)",
			mwOrbit.World.UWP.Size)
	}
}

// assertNoOrphanedSatellites fails t if any Satellite:true entry's
// Number doesn't match an existing top-level (non-Satellite) entry —
// every satellite must share its parent's Number, never a Number nothing
// else occupies.
func assertNoOrphanedSatellites(t *testing.T, sys StarSystem) {
	t.Helper()

	topLevelNumbers := map[int]bool{}

	for _, o := range sys.Orbits {
		if !o.Satellite && (o.World != nil || o.GasGiant != nil) {
			topLevelNumbers[o.Number] = true
		}
	}

	for _, o := range sys.Orbits {
		if o.Satellite && !topLevelNumbers[o.Number] {
			t.Fatalf("orphaned satellite at orbit %d: no top-level body shares that Number", o.Number)
		}
	}
}

// scanSystemOrbits counts Primary-role stars and Gas Giants in sys, and
// fails t if any non-Satellite entry shares an orbit Number with another
// — the collision guard placeInOrbit exists for.
func scanSystemOrbits(t *testing.T, sys StarSystem) (int, int) {
	t.Helper()

	primaryCount, gasGiantCount := 0, 0
	seenNumbers := map[int]bool{}

	for _, o := range sys.Orbits {
		if o.Star != nil && o.Number == primaryOrbitNumber {
			primaryCount++
		}

		if o.GasGiant != nil {
			gasGiantCount++
		}

		if o.Satellite {
			continue
		}

		if seenNumbers[o.Number] {
			t.Fatalf("orbit %d occupied by more than one non-Satellite entry", o.Number)
		}

		seenNumbers[o.Number] = true
	}

	return primaryCount, gasGiantCount
}
