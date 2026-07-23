package world_test

import (
	"testing"

	"github.com/philoserf/traveller/world"
)

func TestStellarRoleString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		role world.StellarRole
		want string
	}{
		{world.Primary, "Primary"},
		{world.Close, "Close"},
		{world.Near, "Near"},
		{world.Far, "Far"},
		{world.StellarRole(99), "Unknown"},
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

	far := world.Star{Role: world.Far}
	closeStar := world.Star{Role: world.Close}
	primary := world.Star{Role: world.Primary}
	near := world.Star{Role: world.Near}

	sys := world.StarSystem{
		Orbits: []world.Orbit{
			{Number: 14, Star: &far},
			{Number: 3, Star: &closeStar},
			{Number: -1, Star: &primary},
			{Number: 8, Star: &near},
		},
		MainworldOrbit: -1,
	}

	starOrbits, _, _ := sys.SystemBodies()

	want := []world.StellarRole{world.Primary, world.Close, world.Near, world.Far}
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

	primary := world.Star{Role: world.Primary}
	gg := world.GasGiant{Size: 'S', Bracket: "LGG"}
	farSat := world.World{}
	closeSat := world.World{}

	sys := world.StarSystem{
		Orbits: []world.Orbit{
			{Number: -1, Star: &primary},
			{Number: 3, HostRole: world.Primary, GasGiant: &gg},
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
