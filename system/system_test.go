package system_test

import (
	"testing"

	"github.com/philoserf/traveller/system"
	"github.com/philoserf/traveller/world"
)

func TestStellarRoleString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		role system.StellarRole
		want string
	}{
		{system.Primary, "Primary"},
		{system.Close, "Close"},
		{system.Near, "Near"},
		{system.Far, "Far"},
		{system.StellarRole(99), "Unknown"},
	}

	for _, c := range cases {
		if got := c.role.String(); got != c.want {
			t.Errorf("StellarRole(%d).String() = %q, want %q", c.role, got, c.want)
		}
	}
}

// TestSystemBodiesSortsStarsCloseToFar builds a StarSystem with its Star
// Orbits appended out of role order (Far, Close, Primary, Near) —
// GenerateSystem itself always appends Primary/Close/Near/Far in that
// order, so relying on append order alone would never exercise this sort
// — and confirms SystemBodies returns starOrbits sorted Primary, Close,
// Near, Far regardless: Orbit.Number can't be the sort key for this,
// since it's a sentinel (not a real orbit slot) for the Primary.
func TestSystemBodiesSortsStarsCloseToFar(t *testing.T) {
	t.Parallel()

	far := system.Star{Role: system.Far}
	closeStar := system.Star{Role: system.Close}
	primary := system.Star{Role: system.Primary}
	near := system.Star{Role: system.Near}

	sys := system.StarSystem{
		Orbits: []system.Orbit{
			{Number: 14, Star: &far},
			{Number: 3, Star: &closeStar},
			{Number: -1, Star: &primary},
			{Number: 8, Star: &near},
		},
		MainworldOrbit: -1,
	}

	starOrbits, _, _ := sys.SystemBodies()

	want := []system.StellarRole{system.Primary, system.Close, system.Near, system.Far}
	if len(starOrbits) != len(want) {
		t.Fatalf("SystemBodies() returned %d star orbits, want %d", len(starOrbits), len(want))
	}

	for i, o := range starOrbits {
		if o.Star.Role != want[i] {
			t.Errorf("starOrbits[%d].Star.Role = %v, want %v", i, o.Star.Role, want[i])
		}
	}
}

// TestSystemBodiesSortsSatellitesCloseToFar builds a Gas Giant at orbit 3
// with its two satellites appended out of order (Far, then Close) — the
// order generateSatellitesForBody's own per-satellite Close/Far roll can
// produce — and confirms SystemBodies returns satellitesOf[3] sorted
// Close before Far, the same close-to-far ordering applied to star groups.
func TestSystemBodiesSortsSatellitesCloseToFar(t *testing.T) {
	t.Parallel()

	primary := system.Star{Role: system.Primary}
	gg := system.GasGiant{Size: 'S', Bracket: "LGG"}
	farSat := world.World{}
	closeSat := world.World{}

	sys := system.StarSystem{
		Orbits: []system.Orbit{
			{Number: -1, Star: &primary},
			{Number: 3, HostRole: system.Primary, GasGiant: &gg},
			{Number: 3, Satellite: true, Close: false, World: &farSat},
			{Number: 3, Satellite: true, Close: true, World: &closeSat},
		},
		MainworldOrbit: -1,
	}

	_, _, satellitesOf := sys.SystemBodies()

	sats := satellitesOf[3]
	if len(sats) != 2 {
		t.Fatalf("satellitesOf[3] has %d entries, want 2", len(sats))
	}

	if !sats[0].Close || sats[1].Close {
		t.Errorf("satellitesOf[3] Close values = [%v, %v], want [true, false] (Close before Far)",
			sats[0].Close, sats[1].Close)
	}
}

// TestSystemBodiesIncludesFreestandingMainworld confirms a non-satellite
// mainworld's Orbit is no longer excluded from bodiesByRole — it flows
// through the same bucketing as every other top-level body, so callers
// (render.System, the API's toSystemResponse) can fold it into their
// normal body listing rather than needing a separate code path for it.
func TestSystemBodiesIncludesFreestandingMainworld(t *testing.T) {
	t.Parallel()

	primary := system.Star{Role: system.Primary}
	mw := world.World{}

	sys := system.StarSystem{
		Orbits: []system.Orbit{
			{Number: -1, Star: &primary},
			{Number: 4, HostRole: system.Primary, World: &mw},
		},
		MainworldOrbit: 1,
	}

	_, bodiesByRole, _ := sys.SystemBodies()

	bodies := bodiesByRole[system.Primary]
	if len(bodies) != 1 || bodies[0].World != &mw {
		t.Errorf("bodiesByRole[Primary] = %v, want exactly the mainworld's own Orbit", bodies)
	}
}

// TestSystemBodiesIncludesSatelliteMainworld confirms a mainworld that is
// itself a satellite of a Gas Giant is no longer excluded from
// satellitesOf either — it appears alongside the Gas Giant's other real
// satellites, the same way TestSystemBodiesIncludesFreestandingMainworld
// confirms for the freestanding case.
func TestSystemBodiesIncludesSatelliteMainworld(t *testing.T) {
	t.Parallel()

	primary := system.Star{Role: system.Primary}
	gg := system.GasGiant{Size: 'S', Bracket: "LGG"}
	mw := world.World{}
	otherMoon := world.World{}

	sys := system.StarSystem{
		Orbits: []system.Orbit{
			{Number: -1, Star: &primary},
			{Number: 3, HostRole: system.Primary, GasGiant: &gg},
			{Number: 3, Satellite: true, Close: true, World: &mw},
			{Number: 3, Satellite: true, Close: false, World: &otherMoon},
		},
		MainworldOrbit: 2,
	}

	_, _, satellitesOf := sys.SystemBodies()

	sats := satellitesOf[3]
	if len(sats) != 2 {
		t.Fatalf("satellitesOf[3] has %d entries, want 2 (the mainworld and the Gas Giant's other moon)", len(sats))
	}

	found := false

	for _, sat := range sats {
		if sat.World == &mw {
			found = true
		}
	}

	if !found {
		t.Errorf("satellitesOf[3] doesn't contain the mainworld's own Orbit: %v", sats)
	}
}

// TestIsMainworld confirms World-pointer identity correctly distinguishes
// the mainworld from an unrelated body sharing the same Number band, and
// that a nil World (either side) never reports a false match — guarding
// against every Gas Giant orbit (World always nil) comparing equal to an
// invalid mainworld Orbit whose own World was never set.
func TestIsMainworld(t *testing.T) {
	t.Parallel()

	primary := system.Star{Role: system.Primary}
	mw := world.World{}
	otherWorld := world.World{}
	gg := system.GasGiant{Size: 'S', Bracket: "LGG"}

	t.Run("matches the mainworld's own Orbit", func(t *testing.T) {
		t.Parallel()

		sys := system.StarSystem{
			Orbits:         []system.Orbit{{Number: -1, Star: &primary}, {Number: 4, World: &mw}},
			MainworldOrbit: 1,
		}

		if !sys.IsMainworld(sys.Orbits[1]) {
			t.Error("IsMainworld(mainworld's own Orbit) = false, want true")
		}

		if sys.IsMainworld(system.Orbit{Number: 4, World: &otherWorld}) {
			t.Error("IsMainworld(a different World at the same Number) = true, want false")
		}
	})

	t.Run("nil World never false-matches", func(t *testing.T) {
		t.Parallel()

		// An invalid StarSystem (MainworldOrbit points at a Star, not a
		// World) — every Gas Giant orbit has a nil World too, and must
		// not be reported as the mainworld just because both are nil.
		sys := system.StarSystem{
			Orbits:         []system.Orbit{{Number: -1, Star: &primary}, {Number: 4, GasGiant: &gg}},
			MainworldOrbit: 0,
		}

		if sys.IsMainworld(sys.Orbits[1]) {
			t.Error("IsMainworld(Gas Giant orbit) = true when mainworld's own World is nil, want false")
		}
	})
}
