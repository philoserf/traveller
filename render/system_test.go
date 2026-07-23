package render_test

import (
	"strings"
	"testing"

	"github.com/philoserf/traveller/render"
	"github.com/philoserf/traveller/world"
)

func fixtureUWP() world.UWP {
	return world.UWP{
		Starport: world.StarportA, Size: 7, Atmosphere: 8, Hydrographics: 8,
		Population: 8, Government: 9, Law: 9, TechLevel: 12,
	}
}

func fixtureMainworld() *world.World {
	return &world.World{
		UWP:        fixtureUWP(),
		TravelZone: world.ZoneGreen,
		Economic:   world.Economic{Resources: 10, Labor: 5, Infrastructure: 5, Efficiency: 1},
		Cultural:   world.Cultural{Heterogeneity: 5, Acceptance: 5, Strangeness: 5, Symbols: 5},
	}
}

// TestSystemGroupsBodiesUnderHostingStar builds a two-star system (a
// Primary and a Close star) with a Gas Giant hosted by the Close star —
// confirming the Gas Giant appears nested under the Close star's own
// "### " group heading (at its own orbit Number), not the Primary's, and
// that the Primary's own (empty) group is shown too.
func TestSystemGroupsBodiesUnderHostingStar(t *testing.T) {
	t.Parallel()

	primary := world.Star{
		SpectralType: world.SpectralG, SpectralDecimal: 2, LuminosityClass: "V",
		Role: world.Primary, HabitableZoneOrbit: 3,
	}
	closeStar := world.Star{
		SpectralType: world.SpectralM, SpectralDecimal: 6, LuminosityClass: "V",
		Role: world.Close, HabitableZoneOrbit: 0,
	}

	sys := world.StarSystem{
		Orbits: []world.Orbit{
			{Number: -1, Star: &primary},
			{Number: 5, Star: &closeStar},
			{Number: 3, HostRole: world.Primary, World: fixtureMainworld()},
			{Number: 2, HostRole: world.Close, GasGiant: &world.GasGiant{Size: 'S', Bracket: "LGG"}},
		},
		MainworldOrbit: 2,
	}

	out := render.System(sys)

	if !strings.Contains(out, "## System") {
		t.Errorf("output missing \"## System\" heading:\n%s", out)
	}

	closeIdx := strings.Index(out, "### Close: M6 V (Orbit 5, HZ orbit 0)")
	if closeIdx == -1 {
		t.Fatalf("output missing the Close star's own group heading:\n%s", out)
	}

	gasGiantIdx := strings.Index(out, "Orbit 2: Gas Giant, Size S (LGG)")
	if gasGiantIdx == -1 {
		t.Fatalf("output missing the Gas Giant's own line:\n%s", out)
	}

	if gasGiantIdx < closeIdx {
		t.Errorf("Gas Giant line appears before the Close star's group heading, want nested under it:\n%s", out)
	}

	if !strings.Contains(out, "### Primary: G2 V (HZ orbit 3)\n\nNone.") {
		t.Errorf("output missing the Primary's own (empty) group:\n%s", out)
	}
}

// TestSystemMainworldSatelliteReportsCloseOrFar confirms the mainworld
// section shows the correct Close/Far phrasing when the mainworld is
// itself a satellite of a Gas Giant, rather than the old unconditional
// "Satellite of" text that couldn't distinguish the two.
func TestSystemMainworldSatelliteReportsCloseOrFar(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		close bool
		want  string
	}{
		{"close", true, "**Close satellite of:**"},
		{"far", false, "**Far satellite of:**"},
	}

	primary := world.Star{
		SpectralType: world.SpectralG, SpectralDecimal: 2, LuminosityClass: "V",
		Role: world.Primary, HabitableZoneOrbit: 3,
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			sys := world.StarSystem{
				Orbits: []world.Orbit{
					{Number: -1, Star: &primary},
					{Number: 3, HostRole: world.Primary, GasGiant: &world.GasGiant{Size: 'S', Bracket: "LGG"}},
					{Number: 3, Satellite: true, Close: c.close, World: fixtureMainworld()},
				},
				MainworldOrbit: 2,
			}

			out := render.System(sys)

			if !strings.Contains(out, c.want) {
				t.Errorf("output missing %q:\n%s", c.want, out)
			}
		})
	}
}
