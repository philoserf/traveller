package sector

import "testing"

func TestDensityString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		density Density
		want    string
	}{
		{DensityExtraGalactic, "Extra Galactic"},
		{DensityRift, "Rift"},
		{DensitySparse, "Sparse"},
		{DensityScattered, "Scattered"},
		{DensityStandard, "Standard"},
		{DensityDense, "Dense"},
		{DensityCluster, "Cluster"},
		{DensityCore, "Core"},
		{Density(99), "Unknown"},
	}

	for _, c := range cases {
		if got := c.density.String(); got != c.want {
			t.Errorf("Density(%d).String() = %q, want %q", c.density, got, c.want)
		}
	}
}

// TestParseDensityRoundTrips confirms every real Density's own String()
// parses back to itself via ParseDensity — the two share one source
// table (densityNames) precisely so they can't drift apart.
func TestParseDensityRoundTrips(t *testing.T) {
	t.Parallel()

	densities := []Density{
		DensityExtraGalactic, DensityRift, DensitySparse, DensityScattered,
		DensityStandard, DensityDense, DensityCluster, DensityCore,
	}

	for _, d := range densities {
		got, ok := ParseDensity(d.String())
		if !ok || got != d {
			t.Errorf("ParseDensity(%q) = (%v, %v), want (%v, true)", d.String(), got, ok, d)
		}
	}

	if _, ok := ParseDensity("Nonexistent"); ok {
		t.Error(`ParseDensity("Nonexistent") reported ok=true, want false`)
	}
}

// buildTestSector builds a Sector with the full 1280 Hexes, each stamped
// with a unique, predictable Location via hexLocation — enough to test
// Subsector's slicing without running actual dice-based generation.
func buildTestSector() Sector {
	hexes := make([]Hex, 0, sectorWidth*sectorHeight)

	for col := 1; col <= sectorWidth; col++ {
		for row := 1; row <= sectorHeight; row++ {
			hexes = append(hexes, Hex{Location: hexLocation(col, row)})
		}
	}

	return Sector{Name: "Test", Hexes: hexes}
}

func TestSectorSubsectorBounds(t *testing.T) {
	t.Parallel()

	sec := buildTestSector()

	cases := []struct {
		letter              byte
		wantFirst, wantLast string
	}{
		{'A', "0101", "0810"},
		{'D', "2501", "3210"},
		{'E', "0111", "0820"},
		{'P', "2531", "3240"},
	}

	for _, c := range cases {
		hexes := sec.Subsector(c.letter)
		if len(hexes) != subsectorWidth*subsectorHeight {
			t.Fatalf("Subsector(%c) returned %d hexes, want %d", c.letter, len(hexes), subsectorWidth*subsectorHeight)
		}

		if hexes[0].Location != c.wantFirst {
			t.Errorf("Subsector(%c)[0].Location = %q, want %q", c.letter, hexes[0].Location, c.wantFirst)
		}

		if last := hexes[len(hexes)-1].Location; last != c.wantLast {
			t.Errorf("Subsector(%c)[last].Location = %q, want %q", c.letter, last, c.wantLast)
		}
	}
}

func TestSectorSubsectorInvalidLetter(t *testing.T) {
	t.Parallel()

	sec := buildTestSector()

	for _, letter := range []byte{'Q', 'Z', '0', 'a'} {
		if hexes := sec.Subsector(letter); hexes != nil {
			t.Errorf("Subsector(%c) = %v, want nil", letter, hexes)
		}
	}
}

func TestValidSubsectorLetter(t *testing.T) {
	t.Parallel()

	for letter := byte('A'); letter <= 'P'; letter++ {
		if !ValidSubsectorLetter(letter) {
			t.Errorf("ValidSubsectorLetter(%c) = false, want true", letter)
		}
	}

	for _, letter := range []byte{'Q', 'Z', '0', 'a'} {
		if ValidSubsectorLetter(letter) {
			t.Errorf("ValidSubsectorLetter(%c) = true, want false", letter)
		}
	}
}
