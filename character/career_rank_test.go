package character

import "testing"

func TestRankState(t *testing.T) {
	t.Parallel()

	const enlistedTiers, officerTiers = 6, 7

	cases := []struct {
		name        string
		terms       []Term
		wantOfficer bool
		wantTier    int
	}{
		{"no terms", nil, false, 1},
		{"one Commissioned term", []Term{{Commissioned: true}}, true, 1},
		{
			"Commissioned then two Promoted",
			[]Term{{Commissioned: true}, {Promoted: true}, {Promoted: true}},
			true, 3,
		},
		{
			"six Enlisted Promoted terms cap at enlistedTiers",
			[]Term{
				{Promoted: true},
				{Promoted: true},
				{Promoted: true},
				{Promoted: true},
				{Promoted: true},
				{Promoted: true},
			},
			false,
			6,
		},
		{
			"a seventh Enlisted Promoted term stays capped at enlistedTiers",
			[]Term{
				{Promoted: true},
				{Promoted: true},
				{Promoted: true},
				{Promoted: true},
				{Promoted: true},
				{Promoted: true},
				{Promoted: true},
			},
			false, 6,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			gotOfficer, gotTier := rankState(c.terms, enlistedTiers, officerTiers)
			if gotOfficer != c.wantOfficer || gotTier != c.wantTier {
				t.Errorf("rankState(%+v, %d, %d) = (%v, %d), want (%v, %d)",
					c.terms, enlistedTiers, officerTiers, gotOfficer, gotTier, c.wantOfficer, c.wantTier)
			}
		})
	}
}

// TestRankStateCapsAtOfficerTiers confirms the Officer track's own cap
// (7, distinct from Enlisted's 6) is respected independently.
func TestRankStateCapsAtOfficerTiers(t *testing.T) {
	t.Parallel()

	terms := []Term{
		{Commissioned: true},
		{Promoted: true},
		{Promoted: true},
		{Promoted: true},
		{Promoted: true},
		{Promoted: true},
		{Promoted: true},
		{Promoted: true}, // 7th promotion attempt past O7
	}

	isOfficer, tier := rankState(terms, 6, 7)
	if !isOfficer || tier != 7 {
		t.Errorf("rankState(...) = (%v, %d), want (true, 7)", isOfficer, tier)
	}
}
