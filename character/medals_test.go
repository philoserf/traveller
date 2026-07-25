package character

import "testing"

// TestMedalTableMatchesBook1P70 is a full-pin regression test for the
// universal Medals table literals transcribed from p.70 — shared by
// every Armed Forces career, confirmed by the table's own closing line
// "Medals = Soldier, Spacer, Marine Promotion Mods".
func TestMedalTableMatchesBook1P70(t *testing.T) {
	t.Parallel()

	wantCodes := [12]string{
		"XS", "XS", "XS", "XS", "XS", "XS", "XS",
		"MCUF", "MCUF",
		"MCG",
		"SEH",
		"SEHD",
	}
	if medalCodes != wantCodes {
		t.Errorf("medalCodes =\n%v\nwant\n%v", medalCodes, wantCodes)
	}

	wantNames := map[string]string{
		"XS":   "XS Exemplary Service",
		"MCUF": "MCUF Meritorious Conduct Under Fire",
		"MCG":  "MCG Medal for Conspicuous Gallantry",
		"SEH":  "SEH Starburst for Extreme Heroism",
		"SEHD": "SEH With Diamonds",
	}
	if len(medalNames) != len(wantNames) {
		t.Fatalf("medalNames has %d entries, want %d", len(medalNames), len(wantNames))
	}

	for code, want := range wantNames {
		if got := medalNames[code]; got != want {
			t.Errorf("medalNames[%q] = %q, want %q", code, got, want)
		}
	}

	wantFame := map[string]int{"XS": 0, "MCUF": 1, "MCG": 2, "SEH": 3, "SEHD": 4}
	if len(medalFame) != len(wantFame) {
		t.Fatalf("medalFame has %d entries, want %d", len(medalFame), len(wantFame))
	}

	for code, want := range wantFame {
		if got := medalFame[code]; got != want {
			t.Errorf("medalFame[%q] = %d, want %d", code, got, want)
		}
	}

	wantMod := map[string]int{"XS": 1, "MCUF": 2, "MCG": 3, "SEH": 4, "SEHD": 5}
	if len(medalTierMod) != len(wantMod) {
		t.Fatalf("medalTierMod has %d entries, want %d", len(medalTierMod), len(wantMod))
	}

	for code, want := range wantMod {
		if got := medalTierMod[code]; got != want {
			t.Errorf("medalTierMod[%q] = %d, want %d", code, got, want)
		}
	}
}

// TestMedalFromReward is a boundary pin over the raw-roll-to-code mapping.
func TestMedalFromReward(t *testing.T) {
	t.Parallel()

	cases := []struct {
		roll int
		want string
	}{
		{2, "XS"}, {8, "XS"}, {9, "MCUF"}, {10, "MCUF"}, {11, "MCG"}, {12, "SEH"}, {13, "SEHD"},
	}

	for _, c := range cases {
		if got := medalFromReward(c.roll); got != c.want {
			t.Errorf("medalFromReward(%d) = %q, want %q", c.roll, got, c.want)
		}
	}
}

func TestMedalModSum(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		medals []string
		want   int
	}{
		{"no medals", nil, 0},
		{"one XS", []string{"XS"}, 1},
		{"mixed", []string{"XS", "MCUF", "MCG", "SEH"}, 1 + 2 + 3 + 4},
	}

	for _, c := range cases {
		if got := medalModSum(c.medals); got != c.want {
			t.Errorf("%s: medalModSum(%v) = %d, want %d", c.name, c.medals, got, c.want)
		}
	}
}

func TestMedalModTotal(t *testing.T) {
	t.Parallel()

	terms := []Term{
		{Medals: []string{"XS"}},
		{Medals: []string{"MCUF", "XS"}},
		{},
	}

	if got, want := medalModTotal(terms), 1+(2+1); got != want {
		t.Errorf("medalModTotal(%+v) = %d, want %d", terms, got, want)
	}
}
