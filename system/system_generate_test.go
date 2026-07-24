package system

import (
	"math/rand/v2"
	"reflect"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/world"
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

// TestHabitableZoneOrbitDLuminosity pins Book 3's own H1/J1/K1 "D"
// column, verified against the PDF page image: 1 for O, 0 for every
// other spectral type. A normal (non-Degenerate) SpectralType can
// independently roll "D" luminosity — stellarSizeTable's flux+5 row is
// "D" across every column — so this is reachable for O just as much as
// for B/A/F/G/K/M.
func TestHabitableZoneOrbitDLuminosity(t *testing.T) {
	t.Parallel()

	cases := []struct {
		spectral SpectralType
		want     int
	}{
		{SpectralO, 1},
		{SpectralB, 0},
		{SpectralK, 0},
		{SpectralM, 0},
	}

	for _, c := range cases {
		hz, ok := habitableZoneOrbit(c.spectral, "D")
		if !ok || hz != c.want {
			t.Errorf("habitableZoneOrbit(%v, \"D\") = (%d, %v), want (%d, true)", c.spectral, hz, ok, c.want)
		}
	}
}

// TestHabitableZoneOrbitTrueDegenerateHasNoRow confirms SpectralDegenerate
// itself — the star *type* (a white dwarf/brown dwarf from the type
// roll onward), distinct from a normal type that independently rolled
// "D" luminosity — has no row in habitableZoneTable at all: Book 3's
// H1/J1/K1 tables have no O-M spectral column for it.
func TestHabitableZoneOrbitTrueDegenerateHasNoRow(t *testing.T) {
	t.Parallel()

	if hz, ok := habitableZoneOrbit(SpectralDegenerate, "D"); ok || hz != 0 {
		t.Errorf("habitableZoneOrbit(SpectralDegenerate, \"D\") = (%d, %v), want (0, false)", hz, ok)
	}
}

// TestRollLuminosityClassEnforcesFootnote pins Book 3 p.28's own margin
// note ("Size IV not for K5-K9 and M0-M9. Size VI not for A0-A9 and
// F0-F4"): at sizeFlux -3, K would raw-table to "IV" and at sizeFlux 4,
// F would raw-table to "VI" — both forbidden for the given decimal, so
// both clamp to "V". The un-clamped half of each pair (K decimal <5, F
// decimal >4) still gets the table's own raw result.
func TestRollLuminosityClassEnforcesFootnote(t *testing.T) {
	t.Parallel()

	if got := rollLuminosityClass(-3, SpectralK, 9); got != "V" {
		t.Errorf("rollLuminosityClass(-3, K, decimal=9) = %q, want \"V\" (K5-K9/IV forbidden)", got)
	}

	if got := rollLuminosityClass(-3, SpectralK, 2); got != "IV" {
		t.Errorf("rollLuminosityClass(-3, K, decimal=2) = %q, want \"IV\" (K0-K4/IV allowed)", got)
	}

	if got := rollLuminosityClass(4, SpectralF, 0); got != "V" {
		t.Errorf("rollLuminosityClass(4, F, decimal=0) = %q, want \"V\" (F0-F4/VI forbidden)", got)
	}

	if got := rollLuminosityClass(4, SpectralF, 7); got != "VI" {
		t.Errorf("rollLuminosityClass(4, F, decimal=7) = %q, want \"VI\" (F5-F9/VI allowed)", got)
	}
}

// TestGasGiantSizeByRollMapping pins the GG table's direct row=roll
// indexing against the Regina worked example (2D=7 -> Siz S, 2D=2 ->
// Siz M, 2D=5 -> Siz Q) plus roll 12 -> Siz X, only reachable once the
// table's row 13 (Y, BD-only) is correctly excluded rather than
// absorbing 12's own slot.
func TestGasGiantSizeByRollMapping(t *testing.T) {
	t.Parallel()

	cases := map[int]byte{2: 'M', 5: 'Q', 7: 'S', 12: 'X'}

	for roll, want := range cases {
		if got := gasGiantSizeByRoll[roll]; got != want {
			t.Errorf("gasGiantSizeByRoll[%d] = %q, want %q", roll, got, want)
		}
	}
}

// TestRollGasGiantSGGLGGThreshold confirms the SGG/LGG split lands at
// 2D6<=6 (~15/36, 41.7%) rather than <=4 — verified against the page
// image's own bracket, cross-checked against the Regina worked example
// (2D=7/Siz S labeled "LGG"; 2D=2/Siz M labeled "Small Gas Giant SGG").
func TestRollGasGiantSGGLGGThreshold(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(11, 12))

	const trials = 20000

	sgg := 0

	for range trials {
		if rollGasGiant(r).Bracket == "SGG" {
			sgg++
		}
	}

	gotPct := 100 * float64(sgg) / trials
	if wantPct := 100.0 * 15 / 36; gotPct < wantPct-3 || gotPct > wantPct+3 {
		t.Errorf("rollGasGiant SGG rate = %.1f%% of %d trials, want ~%.1f%% (2D6<=6)", gotPct, trials, wantPct)
	}
}

// TestP2BasicPlacementChartMapping pins every reachable cell (2D6 rolls
// 2-12) of Book 3 p.29's "P2 Basic Placement Chart", transcribed
// directly from the page image, row by row:
//
//	2D  LGG SGG  IG Belt World1 World2
//	1   -4  -3  HZ   -2    11     18
//	2   -3  -2  +1   -1    10     17
//	3   -2  -1  +2   HZ     8     16
//	4   -1  HZ  +3   +1     6     15
//	5   HZ  +1  +4   +2     4     14
//	6   +1  +2  +5   +3     2     13
//	7   +2  +3  +6   +4     1     11
//	8   +3  +4  +7   +5     3     10
//	9   +4  +5  +8   +6     5      9
//	10  +5  +6  +9   +7     7      8
//	11  +6  +7 +10   +8     9      7
//	12  +7  +8 +11   +9     9      7
//
// Row N's own value belongs at roll N (direct indexing, matching
// gasGiantSizeByRoll — see its own doc comment), not row N-1 as an
// earlier reading of this whole table assumed. World1's column is
// notably non-linear (a V bottoming out at orbit 1 for roll 7, not the
// smoother, one-roll-earlier valley the old reading produced) — the
// case most likely to hide a transcription slip, so pinned in full here
// rather than only spot-checked.
func TestP2BasicPlacementChartMapping(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		got  map[int]int
		want map[int]int
	}{
		{"lggOffsetByRoll", lggOffsetByRoll, map[int]int{
			2: -3, 3: -2, 4: -1, 5: 0, 6: 1, 7: 2, 8: 3, 9: 4, 10: 5, 11: 6, 12: 7,
		}},
		{"sggOffsetByRoll", sggOffsetByRoll, map[int]int{
			2: -2, 3: -1, 4: 0, 5: 1, 6: 2, 7: 3, 8: 4, 9: 5, 10: 6, 11: 7, 12: 8,
		}},
		{"igOffsetByRoll", igOffsetByRoll, map[int]int{
			2: 1, 3: 2, 4: 3, 5: 4, 6: 5, 7: 6, 8: 7, 9: 8, 10: 9, 11: 10, 12: 11,
		}},
		{"beltOffsetByRoll", beltOffsetByRoll, map[int]int{
			2: -1, 3: 0, 4: 1, 5: 2, 6: 3, 7: 4, 8: 5, 9: 6, 10: 7, 11: 8, 12: 9,
		}},
		{"world1OrbitByRoll", world1OrbitByRoll, map[int]int{
			2: 10, 3: 8, 4: 6, 5: 4, 6: 2, 7: 1, 8: 3, 9: 5, 10: 7, 11: 9, 12: 9,
		}},
		{"world2OrbitByRoll", world2OrbitByRoll, map[int]int{
			2: 17, 3: 16, 4: 15, 5: 14, 6: 13, 7: 11, 8: 10, 9: 9, 10: 8, 11: 7, 12: 7,
		}},
	}

	for _, c := range cases {
		for roll, want := range c.want {
			if got := c.got[roll]; got != want {
				t.Errorf("%s[%d] = %d, want %d", c.name, roll, got, want)
			}
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
