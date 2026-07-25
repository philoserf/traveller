package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

func TestSucceedsAgainstBoundaries(t *testing.T) {
	t.Parallel()

	for roll := 2; roll <= 12; roll++ {
		if got, want := succeedsAgainst(roll, roll, 0), true; got != want {
			t.Errorf("succeedsAgainst(%d, target=%d, 0) = %v, want %v (exact match succeeds)", roll, roll, got, want)
		}

		if got, want := succeedsAgainst(roll, roll-1, 0), false; got != want {
			t.Errorf("succeedsAgainst(%d, target=%d, 0) = %v, want %v (one below fails)", roll, roll-1, got, want)
		}

		if got, want := succeedsAgainst(roll, roll-1, 1), true; got != want {
			t.Errorf(
				"succeedsAgainst(%d, target=%d, mod=1) = %v, want %v (mod closes the gap)",
				roll,
				roll-1,
				got,
				want,
			)
		}
	}
}

func TestHighestOf(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{5, 9, 3, 7, 2, 8}}

	if got, want := highestOf(upp, C1, C2, C3), C2; got != want {
		t.Errorf("highestOf(C1,C2,C3) = %v, want %v (C2=9 is highest)", got, want)
	}
}

// TestRiskOutcomeBoundaries pins Book 1 p.65's universal four Risk
// outcomes against fixed original/reduced pairs, no dice involved —
// originally Scout-specific (TestScoutRiskOutcomeBoundaries), renamed
// alongside riskOutcome itself once Marine became a second real caller.
func TestRiskOutcomeBoundaries(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name              string
		original, reduced int
		want              RiskResult
	}{
		{"no net loss (Flux offset the Mods)", 8, 8, Unharmed},
		{"1-point loss", 8, 7, Wounded},
		{"3-point loss", 8, 5, Wounded},
		{"exactly 4-point loss", 8, 4, Disabled},
		{"more than 4-point loss, not yet 0", 8, 1, Disabled},
		{"reduced to 0", 8, 0, Dead},
		{"reduced to 0 from a small original", 2, 0, Dead},
		{"already 0, stays 0 (regression: loss=0 must not misreport as Unharmed)", 0, 0, Dead},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := riskOutcome(c.original, c.reduced); got != c.want {
				t.Errorf("riskOutcome(%d, %d) = %v, want %v", c.original, c.reduced, got, c.want)
			}
		})
	}
}

// TestBeginScoutRetryImprovesOdds confirms the two-roll Begin+Retry
// structure is wired correctly: a character with a poor C1-C3 but a
// strong Education (C5) should succeed far more often than a single
// C1-C3-only roll would predict, since a failed first attempt gets a
// second chance against C5.
func TestBeginScoutRetryImprovesOdds(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{2, 2, 2, 2, 12, 2}}
	r := dice.New(rand.NewPCG(3, 3))

	const trials = 2000

	successes := 0

	for range trials {
		if _, ok := BeginScout(r, upp); ok {
			successes++
		}
	}

	// 2D<=2 alone succeeds ~2.8% of the time; with the C5=12 retry on
	// failure (2D<=12 always succeeds), the combined rate should be
	// close to 100%, not close to 2.8%.
	gotPct := 100 * float64(successes) / trials
	if gotPct < 95 {
		t.Errorf("BeginScout success rate = %.1f%% over %d trials, want >95%% (retry against C5=12 should dominate)",
			gotPct, trials)
	}
}

// TestScoutSkillTableMapping pins every one of the 42 cells in Book 1
// p.79's Scout Skills table, transcribed directly from the page image —
// full-pinned, not sampled: TestHomeworldSkillByTradeCodeMapping in
// homeworld_generate_test.go was originally a partial pin, and a real
// transcription error ("JOT" instead of "Jack of all Trades") slipped
// through undetected until the map was expanded to a full pin — see
// that test's own doc comment.
func TestScoutSkillTableMapping(t *testing.T) {
	t.Parallel()

	want := [7][6]string{
		{"Str", "Dex", "End", "Int", "Edu", "Soc"},
		{"Major", "Major", "Minor", "Minor", "One Trade", "One Trade"},
		{"Comms", "Language", "Computer", "Jack of all Trades", "Gunner", "Starship Skill"},
		{"Survey", "Survival", "Hostile Environment", "Animals", "Vacc Suit", "Navigation"},
		{"Diplomat", "Sensors", "Trader", "Teacher", "Fighter", "Streetwise"},
		{"Survey", "Flyer", "Language", "Starship Skill", "Engineer", "Comms"},
		{"One Art", "One Science", "Seafarer", "Athlete", "Medic", "One Trade"},
	}

	if scoutSkillTable != want {
		t.Errorf("scoutSkillTable =\n%v\nwant\n%v", scoutSkillTable, want)
	}
}

func TestResolveScoutTermDeterminism(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}

	r1 := dice.New(rand.NewPCG(11, 11))
	r2 := dice.New(rand.NewPCG(11, 11))

	term1, updatedUPP1, survived1 := ResolveScoutTerm(r1, upp, C1)
	term2, updatedUPP2, survived2 := ResolveScoutTerm(r2, upp, C1)

	if survived1 != survived2 {
		t.Fatalf("identical seeds produced different survival: %v vs %v", survived1, survived2)
	}

	if term1.RiskResult != term2.RiskResult || term1.RewardResult != term2.RewardResult {
		t.Fatalf("identical seeds produced different outcomes: %+v vs %+v", term1, term2)
	}

	if updatedUPP1 != updatedUPP2 {
		t.Fatalf("identical seeds produced different updated UPPs: %+v vs %+v", updatedUPP1, updatedUPP2)
	}

	if len(term1.SkillsAwarded) != len(term2.SkillsAwarded) {
		t.Fatalf("identical seeds produced different skill counts: %v vs %v", term1.SkillsAwarded, term2.SkillsAwarded)
	}

	for i := range term1.SkillsAwarded {
		if term1.SkillsAwarded[i] != term2.SkillsAwarded[i] {
			t.Fatalf("identical seeds produced different skills at %d: %+v vs %+v",
				i, term1.SkillsAwarded[i], term2.SkillsAwarded[i])
		}
	}
}

// TestResolveScoutTermSkillCountMatchesDuty confirms Courier Duty grants
// (at most) 4 skill rolls and Explorer Duty grants (at most) 8 — "at
// most" since a roll landing on an unresolvable table entry (Major/
// Minor/One Science) grants nothing, per rollScoutSkill's own doc
// comment.
func TestResolveScoutTermSkillCountMatchesDuty(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}
	r := dice.New(rand.NewPCG(17, 18))

	for range 500 {
		term, _, _ := ResolveScoutTerm(r, upp, C1)

		if len(term.SkillsAwarded) > 8 {
			t.Fatalf("term granted %d skills, want at most 8 (Explorer Duty's own ceiling)", len(term.SkillsAwarded))
		}

		if term.Length != 4 {
			t.Fatalf("term.Length = %d, want 4", term.Length)
		}
	}
}
