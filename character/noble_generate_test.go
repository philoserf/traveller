package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

func TestBeginNoble(t *testing.T) {
	t.Parallel()

	cases := []struct {
		soc  ehex.Value
		want bool
	}{
		{10, false}, // 'A'
		{11, true},  // 'B'
		{12, true},
	}

	for _, c := range cases {
		if got := BeginNoble(c.soc); got != c.want {
			t.Errorf("BeginNoble(%v) = %v, want %v", c.soc, got, c.want)
		}
	}
}

func TestNobleBaseFame(t *testing.T) {
	t.Parallel()

	cases := []struct {
		soc  ehex.Value
		want int
	}{
		{10, 15},
		{11, 16}, // floor(16.5)
		{12, 18},
	}

	for _, c := range cases {
		if got := nobleBaseFame(c.soc); got != c.want {
			t.Errorf("nobleBaseFame(%v) = %d, want %d", c.soc, got, c.want)
		}
	}
}

func TestReturnIntrigueMod(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		terms []Term
		want  int
	}{
		{"no terms", nil, 0},
		{"one successful intrigue", []Term{{NobleAction: "Intrigue", NobleSucceeded: true}}, -1},
		{"one failed intrigue", []Term{{NobleAction: "Intrigue", NobleSucceeded: false}}, 1},
		{
			"mixed",
			[]Term{
				{NobleAction: "Intrigue", NobleSucceeded: true},
				{NobleAction: "Intrigue", NobleSucceeded: true},
				{NobleAction: "Intrigue", NobleSucceeded: false},
			},
			-1, // timesExiled(1) - successfulIntrigues(2)
		},
		{
			"Return terms never affect the count",
			[]Term{
				{NobleAction: "Return", NobleSucceeded: true},
				{NobleAction: "Return", NobleSucceeded: false},
			},
			0,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := returnIntrigueMod(c.terms); got != c.want {
				t.Errorf("returnIntrigueMod(%v) = %d, want %d", c.terms, got, c.want)
			}
		})
	}
}

func TestNobleIntrigueCounts(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                       string
		terms                      []Term
		wantSuccessful, wantExiled int
	}{
		{"no terms", nil, 0, 0},
		{"one successful intrigue", []Term{{NobleAction: "Intrigue", NobleSucceeded: true}}, 1, 0},
		{"one failed intrigue", []Term{{NobleAction: "Intrigue", NobleSucceeded: false}}, 0, 1},
		{
			"mixed",
			[]Term{
				{NobleAction: "Intrigue", NobleSucceeded: true},
				{NobleAction: "Intrigue", NobleSucceeded: true},
				{NobleAction: "Intrigue", NobleSucceeded: false},
			},
			2, 1,
		},
		{
			"Return terms never affect the count",
			[]Term{
				{NobleAction: "Return", NobleSucceeded: true},
				{NobleAction: "Return", NobleSucceeded: false},
			},
			0, 0,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			gotSuccessful, gotExiled := nobleIntrigueCounts(c.terms)
			if gotSuccessful != c.wantSuccessful || gotExiled != c.wantExiled {
				t.Errorf("nobleIntrigueCounts(%v) = (%d, %d), want (%d, %d)",
					c.terms, gotSuccessful, gotExiled, c.wantSuccessful, c.wantExiled)
			}
		})
	}
}

func TestNobleExileFame(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		terms []Term
		want  int
	}{
		{"no terms", nil, 0},
		{
			"mixed successes and exiles",
			[]Term{
				{NobleAction: "Intrigue", NobleSucceeded: true},
				{NobleAction: "Intrigue", NobleSucceeded: false},
				{NobleAction: "Return", NobleSucceeded: true},
				{NobleAction: "Intrigue", NobleSucceeded: false},
			},
			2,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := nobleExileFame(c.terms); got != c.want {
				t.Errorf("nobleExileFame(%v) = %d, want %d", c.terms, got, c.want)
			}
		})
	}
}

func TestNobleIsExiled(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		terms []Term
		want  bool
	}{
		{"no terms", nil, false},
		{"last term Return success", []Term{{NobleAction: "Return", NobleSucceeded: true}}, false},
		{"last term Return failure", []Term{{NobleAction: "Return", NobleSucceeded: false}}, true},
		{"last term Intrigue success", []Term{{NobleAction: "Intrigue", NobleSucceeded: true}}, false},
		{"last term Intrigue failure", []Term{{NobleAction: "Intrigue", NobleSucceeded: false}}, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := nobleIsExiled(c.terms); got != c.want {
				t.Errorf("nobleIsExiled(%v) = %v, want %v", c.terms, got, c.want)
			}
		})
	}
}

func TestNobleFluxAlreadyUsed(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		terms []Term
		want  bool
	}{
		{"no terms", nil, false},
		{"successful intrigue present", []Term{{NobleAction: "Intrigue", NobleSucceeded: true}}, true},
		{
			"only failed intrigues and returns",
			[]Term{
				{NobleAction: "Intrigue", NobleSucceeded: false},
				{NobleAction: "Return", NobleSucceeded: true},
			},
			false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := nobleFluxAlreadyUsed(c.terms); got != c.want {
				t.Errorf("nobleFluxAlreadyUsed(%v) = %v, want %v", c.terms, got, c.want)
			}
		})
	}
}

// TestNobleElevationOutcome is the regression test for the ">="
// direction, not succeedsAgainst's own "<=" convention.
func TestNobleElevationOutcome(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name            string
		roll, soc, flux int
		want            bool
	}{
		{"one below soc fails", 6, 7, 0, false},
		{"exactly equal succeeds", 7, 7, 0, true},
		{"one above soc succeeds", 8, 7, 0, true},
		{"flux pushes a failure into a success", 6, 7, 1, true},
		{"negative flux pushes a success into a failure", 7, 7, -1, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := nobleElevationOutcome(c.roll, c.soc, c.flux); got != c.want {
				t.Errorf("nobleElevationOutcome(%d, %d, %d) = %v, want %v", c.roll, c.soc, c.flux, got, c.want)
			}
		})
	}
}

// TestRollNobleElevationRate confirms useFlux actually widens the roll:
// at soc=9, P(2D>=9) = P(9)+P(10)+P(11)+P(12) = (4+3+2+1)/36 = 10/36 ≈
// 27.8% without Flux. With Flux (range -5..+5, mean 0, symmetric), the
// fired rate should differ measurably from the no-Flux baseline over
// enough trials, confirming rollNobleElevation actually wires useFlux
// into the roll rather than ignoring it.
func TestRollNobleElevationRate(t *testing.T) {
	t.Parallel()

	const trials = 20000

	withoutFlux := 0
	r1 := dice.New(rand.NewPCG(1, 1))

	for range trials {
		if rollNobleElevation(r1, 9, false) {
			withoutFlux++
		}
	}

	gotPct := 100 * float64(withoutFlux) / trials
	if wantPct := 100.0 * 10 / 36; gotPct < wantPct-1 || gotPct > wantPct+1 {
		t.Errorf(
			"rollNobleElevation(soc=9, useFlux=false) fired %.2f%% of %d trials, want ~%.2f%%",
			gotPct,
			trials,
			wantPct,
		)
	}

	withFlux := 0
	r2 := dice.New(rand.NewPCG(2, 2))

	for range trials {
		if rollNobleElevation(r2, 9, true) {
			withFlux++
		}
	}

	gotFluxPct := 100 * float64(withFlux) / trials
	if gotFluxPct <= gotPct {
		t.Errorf(
			"rollNobleElevation with Flux fired %.2f%%, want more than the no-Flux rate %.2f%% (Flux widens the roll)",
			gotFluxPct,
			gotPct,
		)
	}
}

func TestResolveNobleTermExiledRollsReturn(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 12}}
	priorTerms := []Term{{NobleAction: "Intrigue", NobleSucceeded: false}} // exiled
	r := dice.New(rand.NewPCG(1, 1))

	term := ResolveNobleTerm(r, upp, C2, priorTerms)

	if term.NobleAction != "Return" {
		t.Errorf("NobleAction = %q, want %q", term.NobleAction, "Return")
	}

	if term.Elevated {
		t.Error("Elevated = true, want false (Return terms never trigger Elevation)")
	}
}

func TestResolveNobleTermNotExiledRollsIntrigue(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 12}}
	r := dice.New(rand.NewPCG(1, 1))

	term := ResolveNobleTerm(r, upp, C2, nil) // no prior terms: not exiled

	if term.NobleAction != "Intrigue" {
		t.Errorf("NobleAction = %q, want %q", term.NobleAction, "Intrigue")
	}
}

// TestResolveNobleTermElevationOnlyOnSuccessfulIntrigue confirms a
// Return term, or a failed Intrigue term, never sets Elevated — using
// UPP{0} to guarantee every Return&Intrigue roll fails (target 0).
func TestResolveNobleTermElevationOnlyOnSuccessfulIntrigue(t *testing.T) {
	t.Parallel()

	upp := UPP{}
	r := dice.New(rand.NewPCG(1, 1))

	for range 100 {
		term := ResolveNobleTerm(r, upp, C2, nil)
		if term.NobleSucceeded {
			t.Fatalf("term unexpectedly succeeded against a zero UPP: %+v", term)
		}

		if term.Elevated {
			t.Errorf("Elevated = true on a failed %s term, want false: %+v", term.NobleAction, term)
		}
	}
}

// TestNobleSkillTableMatchesBook1P85 is a full-pin regression test for
// nobleSkillTable's own transcription.
func TestNobleSkillTableMatchesBook1P85(t *testing.T) {
	t.Parallel()

	want := [7][6]string{
		{"Str", "Dex", "End", "Int", "Edu", "Soc"},
		{"Major", "Major", "Minor", "Minor", "One Trade", "One Trade"},
		{"Driver", "Flyer", "Pilot", "Starship Skill", "High-G", "Zero-G"},
		{"Advocate", "Counsellor", "Bureaucrat", "Liaison", "Leader", "Leader"},
		{"Liaison", "Strategy", "Tactics", "Diplomat", "Advocate", "Leader"},
		{"Capital", "Admin", "Language", "Starship Skill", "Bureaucrat", "Comms"},
		{"One Art", "One Science", "Athlete", "Medic", "Seafarer", "One Trade"},
	}

	if nobleSkillTable != want {
		t.Errorf("nobleSkillTable = %v, want %v", nobleSkillTable, want)
	}
}
