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
// Primary hosting the mainworld — with a satellite of its own — and a
// Close star hosting a Gas Giant) — confirming the Gas Giant appears
// nested under the Close star's own "### " group heading (at its own
// orbit Number), not the Primary's; that the mainworld itself appears as
// a normal entry under the Primary's group, marked "(Mainworld)"; that
// its own satellite nests under that entry without the marker (it's a
// regular moon, not the mainworld itself); and that the old standalone
// "### Satellites" subsection is gone.
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
	moon := &world.World{UWP: fixtureUWP()}

	sys := world.StarSystem{
		Orbits: []world.Orbit{
			{Number: -1, Star: &primary},
			{Number: 5, Star: &closeStar},
			{Number: 3, HostRole: world.Primary, World: fixtureMainworld()},
			{Number: 3, Satellite: true, Close: true, World: moon},
			{Number: 2, HostRole: world.Close, GasGiant: &world.GasGiant{Size: 'S', Bracket: "LGG"}},
		},
		MainworldOrbit: 2,
	}

	out := render.System(sys)

	if !strings.Contains(out, "## System") {
		t.Errorf("output missing \"## System\" heading:\n%s", out)
	}

	if strings.Contains(out, "### Satellites") {
		t.Errorf("output still contains the old standalone \"### Satellites\" subsection:\n%s", out)
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

	mwLineIdx := strings.Index(out, "Orbit 3: A788899-C — None (Mainworld)")
	if mwLineIdx == -1 {
		t.Fatalf("output missing the mainworld's own line, marked (Mainworld), under Primary's group:\n%s", out)
	}

	moonIdx := strings.Index(out, "Close satellite: A788899-C — None\n")
	if moonIdx == -1 {
		t.Fatalf("output missing the mainworld's own (unmarked) satellite:\n%s", out)
	}

	if moonIdx < mwLineIdx {
		t.Errorf("mainworld's satellite line appears before the mainworld's own line, want nested under it:\n%s", out)
	}
}

// TestSystemMainworldSatelliteReportsCloseOrFar confirms the mainworld
// section shows the correct Close/Far phrasing when the mainworld is
// itself a satellite of a Gas Giant, rather than the old unconditional
// "Satellite of" text that couldn't distinguish the two — and that the
// same mainworld also appears nested under its host Gas Giant's own entry
// in "## System", marked "(Mainworld)", now that SystemBodies no longer
// excludes it.
func TestSystemMainworldSatelliteReportsCloseOrFar(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		close     bool
		want      string
		wantInSys string
	}{
		{"close", true, "**Close satellite of:**", "Close satellite: A788899-C — None (Mainworld)"},
		{"far", false, "**Far satellite of:**", "Far satellite: A788899-C — None (Mainworld)"},
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

			if !strings.Contains(out, c.wantInSys) {
				t.Errorf("output missing %q under the Gas Giant's entry in \"## System\":\n%s", c.wantInSys, out)
			}
		})
	}
}
