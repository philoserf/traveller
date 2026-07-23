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

// TestSystemListsCloseStarAndAttributesHost builds a two-star system (a
// Primary and a Close star) with a Gas Giant hosted by the Close star —
// confirming the merged system list surfaces the Close star at its own
// orbit Number, and tags the Gas Giant's host now that it's ambiguous
// which star placed it just from the orbit number alone.
func TestSystemListsCloseStarAndAttributesHost(t *testing.T) {
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

	if strings.Contains(out, "## Other Bodies") {
		t.Errorf("output still contains the old \"## Other Bodies\" heading:\n%s", out)
	}

	if !strings.Contains(out, "Orbit 5: Close: M6 V (HZ orbit 0)") {
		t.Errorf("output missing the Close star's own list entry:\n%s", out)
	}

	if !strings.Contains(out, "Orbit 2: Gas Giant, Size S (LGG) (hosted by Close)") {
		t.Errorf("output missing the Gas Giant's host attribution:\n%s", out)
	}
}

// TestSystemOmitsHostAttributionForSingleStar confirms the "(hosted
// by ...)" suffix — only meaningful once a body's Number alone can't say
// which star placed it — doesn't appear for the common single-star case.
func TestSystemOmitsHostAttributionForSingleStar(t *testing.T) {
	t.Parallel()

	primary := world.Star{
		SpectralType: world.SpectralG, SpectralDecimal: 2, LuminosityClass: "V",
		Role: world.Primary, HabitableZoneOrbit: 3,
	}

	sys := world.StarSystem{
		Orbits: []world.Orbit{
			{Number: -1, Star: &primary},
			{Number: 3, HostRole: world.Primary, World: fixtureMainworld()},
			{Number: 5, HostRole: world.Primary, GasGiant: &world.GasGiant{Size: 'S', Bracket: "LGG"}},
		},
		MainworldOrbit: 1,
	}

	out := render.System(sys)

	if strings.Contains(out, "hosted by") {
		t.Errorf("single-star system output shouldn't show host attribution:\n%s", out)
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
