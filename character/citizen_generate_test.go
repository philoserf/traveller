package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

func TestBeginCitizenAlwaysSucceeds(t *testing.T) {
	t.Parallel()

	if !BeginCitizen() {
		t.Error("BeginCitizen() = false, want true (Book 1 p.78: \"To Begin Auto\")")
	}
}

// TestCitizenTableCMatchesBook1P78 pins every one of the 42 cells in
// Book 1 p.78's Citizen Skills table, transcribed directly from the page
// image — full-pinned, not sampled, matching
// TestScoutSkillTableMapping's own "no partial pins" convention.
func TestCitizenTableCMatchesBook1P78(t *testing.T) {
	t.Parallel()

	want := [7][6]string{
		{"Str", "Dex", "End", "Int", "Edu", "Soc"},
		{"Major", "Major", "Minor", "Minor", "One Trade", "One Trade"},
		{"Seafarer", "Navigator", "Hostile Environment", "Flyer", "Driver", "Vacc Suit"},
		{"Admin", "Broker", "Computer", "Animals", "Bureaucrat", "Trader"},
		{"Advocate", "Broker", "Trader", "Liaison", "Counsellor", "Teacher"},
		{"One Art", "One Science", "One Trade", "Driver", "Bureaucrat", "Computer"},
		{"One Art", "One Science", "Jack of all Trades", "Athlete", "Medic", "One Trade"},
	}

	if citizenTableC != want {
		t.Errorf("citizenTableC =\n%v\nwant\n%v", citizenTableC, want)
	}
}

// TestResolveCitizenLifeAlwaysSucceedsAtMaxCC and
// TestResolveCitizenLifeNeverSucceedsAtZeroCC confirm resolveCitizenLife
// is correctly wired to rollAgainstTarget (target=cc, mod=0) at both
// extremes — deterministic, not statistical: 2D6's range is [2,12], so
// cc=12 makes the roll mathematically guaranteed to succeed and cc=0
// mathematically guaranteed to fail.
func TestResolveCitizenLifeAlwaysSucceedsAtMaxCC(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))

	for range 100 {
		if !resolveCitizenLife(r, 12) {
			t.Fatal("resolveCitizenLife(cc=12) = false, want true (2D6 <= 12 always)")
		}
	}
}

func TestResolveCitizenLifeNeverSucceedsAtZeroCC(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(2, 2))

	for range 100 {
		if resolveCitizenLife(r, 0) {
			t.Fatal("resolveCitizenLife(cc=0) = true, want false (2D6 minimum is 2)")
		}
	}
}

func TestResolveCitizenTermLengthIsFour(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{6, 6, 6, 6, 6, 6}}
	r := dice.New(rand.NewPCG(3, 3))

	if term := ResolveCitizenTerm(r, upp, C1); term.Length != 4 {
		t.Errorf("term.Length = %d, want 4", term.Length)
	}
}

// TestResolveCitizenTermNeverTouchesRiskResult is the regression test for
// this slice's core design decision: Citizen Life has no wound mechanic,
// so RiskResult must stay at its zero value (Unharmed) regardless of
// CitizenLifeSucceeded — matching ResolveScoutTerm's own Courier Duty
// precedent (RiskResult stays Unharmed when no Risk roll happens).
func TestResolveCitizenTermNeverTouchesRiskResult(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{6, 6, 6, 6, 6, 6}}
	r := dice.New(rand.NewPCG(4, 4))

	sawSuccess, sawFailure := false, false

	for range 500 {
		term := ResolveCitizenTerm(r, upp, C1)
		if term.RiskResult != Unharmed {
			t.Fatalf("term.RiskResult = %v, want Unharmed (Citizen Life has no wound mechanic)", term.RiskResult)
		}

		if term.CitizenLifeSucceeded {
			sawSuccess = true
		} else {
			sawFailure = true
		}
	}

	if !sawSuccess || !sawFailure {
		t.Fatalf(
			"500 trials at cc=6 didn't produce both outcomes (sawSuccess=%v, sawFailure=%v) — can't trust the assertion above",
			sawSuccess,
			sawFailure,
		)
	}
}

// TestResolveCitizenTermGrantsAtMostFourSkillsRegardlessOfLifeOutcome
// confirms Table C's own 4-skills-per-term grant (p.78: "Per Term 4 on
// Table C") is unconditional — unlike Scout's duty-dependent skill
// count, it doesn't depend on CitizenLifeSucceeded at all. "At most"
// since a roll landing on an unresolvable Table C entry (Major/Minor/One
// Science) still grants nothing that slot, per rollSkillFromTable's own
// doc comment.
func TestResolveCitizenTermGrantsAtMostFourSkillsRegardlessOfLifeOutcome(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{6, 6, 6, 6, 6, 6}}
	r := dice.New(rand.NewPCG(5, 5))

	for range 500 {
		term := ResolveCitizenTerm(r, upp, C1)
		if len(term.SkillsAwarded) > 4 {
			t.Fatalf("term granted %d skills, want at most 4", len(term.SkillsAwarded))
		}
	}
}

func TestResolveCitizenTermDeterminism(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{6, 6, 6, 6, 6, 6}}

	r1 := dice.New(rand.NewPCG(11, 11))
	r2 := dice.New(rand.NewPCG(11, 11))

	term1 := ResolveCitizenTerm(r1, upp, C2)
	term2 := ResolveCitizenTerm(r2, upp, C2)

	if term1.CitizenLifeSucceeded != term2.CitizenLifeSucceeded {
		t.Fatalf("identical seeds produced different CitizenLifeSucceeded: %v vs %v",
			term1.CitizenLifeSucceeded, term2.CitizenLifeSucceeded)
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
