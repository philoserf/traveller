package world

import (
	"math/rand/v2"
	"reflect"
	"testing"

	"github.com/philoserf/traveller/dice"
)

// TestHabitableZoneOrbitMatchesReginaPrimary is a consistency check, not
// an exact reproduction: Book 3's worked Regina example (p.22) states her
// Primary is F7V and that she orbits a Gas Giant placed at orbit 4, but
// never states the *raw* HZ orbit for F7V directly — only the final,
// already-HZVar-adjusted placement. Unlike the UWP-based fixtures
// elsewhere in this package, there's no way to recover the book's own
// intermediate die rolls from a published narrative, so this only checks
// that this table's F/V value (5) is reachable as Regina's placement (4)
// via some valid HZVar (5 + (-1) = 4, i.e. mainworldHZVar(-3) — a
// plausible roll, not a proven one).
func TestHabitableZoneOrbitMatchesReginaPrimary(t *testing.T) {
	t.Parallel()

	hz, ok := habitableZoneOrbit(SpectralF, "V")
	if !ok {
		t.Fatalf("habitableZoneOrbit(F, V) reported no HZ, want a value")
	}

	if want := 4; hz+mainworldHZVar(-3) != want {
		t.Errorf("habitableZoneOrbit(F, V)=%d + mainworldHZVar(-3)=%d = %d, want %d (Regina's published orbit)",
			hz, mainworldHZVar(-3), hz+mainworldHZVar(-3), want)
	}
}

func TestHabitableZoneOrbitUnknownCombination(t *testing.T) {
	t.Parallel()

	if _, ok := habitableZoneOrbit(SpectralO, "VI"); ok {
		t.Error("habitableZoneOrbit(O, VI) reported a value, want ok=false (no HZ for this combination)")
	}
}

func TestHabitableZoneOrbitDegenerateCollapses(t *testing.T) {
	t.Parallel()

	// Both a "true" Degenerate spectral roll and a non-Degenerate type
	// that independently rolled a "D" luminosity (stellarSizeTable's
	// flux+5 row is "D" across every column) must collapse the same way.
	cases := []SpectralType{SpectralDegenerate, SpectralK, SpectralM}

	for _, t2 := range cases {
		hz, ok := habitableZoneOrbit(t2, "D")
		if !ok || hz != 0 {
			t.Errorf("habitableZoneOrbit(%v, \"D\") = (%d, %v), want (0, true)", t2, hz, ok)
		}
	}
}

func TestMainworldHZVarBoundaries(t *testing.T) {
	t.Parallel()

	cases := []struct {
		flux int
		want int
	}{
		{-7, -2},
		{-6, -2},
		{-5, -1},
		{-3, -1},
		{-2, 0},
		{2, 0},
		{3, 1},
		{5, 1},
		{6, 2},
		{7, 2},
	}

	for _, c := range cases {
		if got := mainworldHZVar(c.flux); got != c.want {
			t.Errorf("mainworldHZVar(%d) = %d, want %d", c.flux, got, c.want)
		}
	}
}

func TestRollMainworldPlacementKindBoundaries(t *testing.T) {
	t.Parallel()

	cases := []struct {
		flux int
		want mainworldPlacementKind
	}{
		{-6, mainworldPlanet},
		{-5, mainworldFarSatellite},
		{-4, mainworldFarSatellite},
		{-3, mainworldCloseSatellite},
		{-2, mainworldPlanet},
		{0, mainworldPlanet},
		{5, mainworldPlanet},
		{6, mainworldPlanet},
	}

	for _, c := range cases {
		if got := rollMainworldPlacementKind(c.flux); got != c.want {
			t.Errorf("rollMainworldPlacementKind(%d) = %v, want %v", c.flux, got, c.want)
		}
	}
}

func TestGenerateSystemDeterminism(t *testing.T) {
	t.Parallel()

	r1 := dice.New(rand.NewPCG(7, 7))
	r2 := dice.New(rand.NewPCG(7, 7))

	mw1 := Generate(r1)
	mw2 := Generate(r2)

	sys1 := GenerateSystem(r1, mw1)
	sys2 := GenerateSystem(r2, mw2)

	if !reflect.DeepEqual(sys1, sys2) {
		t.Errorf("identical seeds produced different systems:\n%+v\nvs\n%+v", sys1, sys2)
	}
}

func TestGenerateSystemPlacesMainworldLast(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(20, 20))
	mw := Generate(r)
	sys := GenerateSystem(r, mw)

	if sys.MainworldOrbit != len(sys.Orbits)-1 {
		t.Fatalf("MainworldOrbit = %d, want %d (last entry)", sys.MainworldOrbit, len(sys.Orbits)-1)
	}

	got := sys.Orbits[sys.MainworldOrbit].World
	if got == nil || got.UWP != mw.UWP {
		t.Errorf("Orbits[MainworldOrbit].World = %+v, want a world with UWP %s", got, mw.UWP)
	}
}

// TestGenerateSystemPrimaryAlwaysPresent runs many seeds and checks every
// invariant that must hold regardless of what the dice produced: exactly
// one Primary-role star at primaryOrbitNumber, a valid mainworld orbit
// index, no more than one Gas Giant (Phase 1 only ever places one, to
// host a satellite mainworld), and a mainworld orbit number that's always
// non-negative and never collides with primaryOrbitNumber's sentinel — a
// regression guard for a real bug: an M-type primary's hzOrbit=0 plus a
// negative HZVar used to produce orbit -1, exactly primaryOrbitNumber.
func TestGenerateSystemInvariants(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(42, 42))

	for range 2000 {
		mw := Generate(r)
		sys := GenerateSystem(r, mw)

		if sys.MainworldOrbit < 0 || sys.MainworldOrbit >= len(sys.Orbits) {
			t.Fatalf("MainworldOrbit = %d out of range for %d orbits", sys.MainworldOrbit, len(sys.Orbits))
		}

		if n := sys.Orbits[sys.MainworldOrbit].Number; n < 0 {
			t.Fatalf("mainworld orbit number = %d, want >= 0 (and not primaryOrbitNumber's sentinel)", n)
		}

		primaryCount, gasGiantCount := 0, 0

		for _, o := range sys.Orbits {
			if o.Star != nil && o.Number == primaryOrbitNumber {
				primaryCount++
			}

			if o.GasGiant != nil {
				gasGiantCount++
			}
		}

		if primaryCount != 1 {
			t.Fatalf("found %d Primary-orbit stars, want exactly 1", primaryCount)
		}

		if gasGiantCount > 1 {
			t.Fatalf("found %d Gas Giants, want at most 1", gasGiantCount)
		}
	}
}

// TestRollSpectralTypeReachesOAndB is a regression guard for a real bug:
// the OB branch's original threshold (flux<=-6) was mathematically
// unreachable given every flux value this project ever computes (see
// rollSpectralType's doc comment), so O/B-class stars could never be
// generated despite SpectralO/SpectralB being fully modeled types.
func TestRollSpectralTypeReachesOAndB(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(9, 9))

	var sawO, sawB bool

	for range 500 {
		switch rollSpectralType(r, -5) { //nolint:exhaustive // only checking that O and B are both reachable
		case SpectralO:
			sawO = true
		case SpectralB:
			sawB = true
		}
	}

	if !sawO || !sawB {
		t.Errorf("rollSpectralType(-5) over 500 rolls: sawO=%v sawB=%v, want both reachable", sawO, sawB)
	}
}
