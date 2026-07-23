package system

import "testing"

func TestSpectralOrdinal(t *testing.T) {
	t.Parallel()

	cases := []struct {
		t       SpectralType
		decimal int
		want    int
	}{
		{SpectralO, 0, 0},
		{SpectralA, 0, 20},
		{SpectralA, 5, 25},
		{SpectralM, 9, 69},
	}

	for _, c := range cases {
		if got := spectralOrdinal(c.t, c.decimal); got != c.want {
			t.Errorf("spectralOrdinal(%v, %d) = %d, want %d", c.t, c.decimal, got, c.want)
		}
	}
}

func TestPrecludedOrbitCeilingLiteralTableEntries(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name            string
		t               SpectralType
		decimal         int
		luminosityClass string
		wantOrbit       int
	}{
		// Book 3 p.20's own printed cells.
		{"Ib row1 A0", SpectralA, 0, "Ib", 1},
		{"II row0 A0", SpectralA, 0, "II", 0},
		{"II row0 F5 (upper end)", SpectralF, 5, "II", 0},
		{"III row0 K0 (upper end)", SpectralK, 0, "III", 0},
		{"Ia row9 M9 (max ordinal)", SpectralM, 9, "Ia", 9},
		// Regression guard: Ib row4 and II row1 are both printed "G5" (not
		// "G4"), ordinal 45 — a code-review catch against a real
		// transcription slip in an earlier version of this table (both
		// cells had been encoded as ordinal 44, one off).
		{"Ib row4 G5 (upper end, not G4)", SpectralG, 5, "Ib", 4},
		{"II row1 G5 (upper end, not G4)", SpectralG, 5, "II", 1},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			orbit, ok := precludedOrbitCeiling(c.t, c.decimal, c.luminosityClass)
			if !ok {
				t.Fatalf(
					"precludedOrbitCeiling(%v, %d, %q) reported ok=false, want true",
					c.t,
					c.decimal,
					c.luminosityClass,
				)
			}

			if orbit != c.wantOrbit {
				t.Errorf(
					"precludedOrbitCeiling(%v, %d, %q) = %d, want %d",
					c.t,
					c.decimal,
					c.luminosityClass,
					orbit,
					c.wantOrbit,
				)
			}
		})
	}
}

// TestPrecludedOrbitCeilingGapRoundsUp pins the documented "round toward
// more precluded" behavior for an ordinal in a gap the table's own rows
// don't cover directly: Ib has A5-G0 at row2 and G5 at row4, with no
// row3 — a G2 star (between G0 and G5) should round up to row4, not
// row2.
func TestPrecludedOrbitCeilingGapRoundsUp(t *testing.T) {
	t.Parallel()

	orbit, ok := precludedOrbitCeiling(SpectralG, 2, "Ib")
	if !ok {
		t.Fatal("precludedOrbitCeiling(G2, Ib) reported ok=false, want true")
	}

	if orbit != 4 {
		t.Errorf("precludedOrbitCeiling(G2, Ib) = %d, want 4 (rounds up to the G5 row, not down to A5-G0)", orbit)
	}
}

// TestPrecludedOrbitCeilingHotterThanTableClamps pins the documented
// clamp-to-first-row behavior for a star hotter than any of a column's
// listed entries (the table has no O/B rows at all, since it starts at
// A0 for every column).
func TestPrecludedOrbitCeilingHotterThanTableClamps(t *testing.T) {
	t.Parallel()

	orbit, ok := precludedOrbitCeiling(SpectralO, 0, "Ia")
	if !ok {
		t.Fatal("precludedOrbitCeiling(O0, Ia) reported ok=false, want true")
	}

	if orbit != 4 {
		t.Errorf("precludedOrbitCeiling(O0, Ia) = %d, want 4 (clamps to Ia's first listed row)", orbit)
	}
}

func TestPrecludedOrbitCeilingNoPreclusionForSmallStars(t *testing.T) {
	t.Parallel()

	for _, luminosityClass := range []string{"IV", "V", "VI", "D", ""} {
		if _, ok := precludedOrbitCeiling(SpectralG, 0, luminosityClass); ok {
			t.Errorf(
				"precludedOrbitCeiling(G0, %q) reported ok=true, want false (no Stellar Surface column for this class)",
				luminosityClass,
			)
		}
	}
}
