package character

import (
	"math/rand/v2"
	"slices"
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

	term, _, _ := ResolveCitizenTerm(r, upp, C1, 0, "", "")
	if term.Length != 4 {
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
		term, _, _ := ResolveCitizenTerm(r, upp, C1, 0, "", "")
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

// TestResolveCitizenTermGrantsUpToFiveSkillsOnSuccess confirms Table C's
// own 4-skills-per-term grant (p.78: "Per Term 4 on Table C") is
// unconditional — unlike Scout's duty-dependent skill count, it doesn't
// depend on CitizenLifeSucceeded at all — and that a successful term may
// grant a 5th skill (the Job/Hobby bonus, p.78's own separate
// entitlement line) on top of it. "At most" in both cases since a roll
// landing on an unresolvable Table C or Table E entry still grants
// nothing that slot.
func TestResolveCitizenTermGrantsUpToFiveSkillsOnSuccess(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{6, 6, 6, 6, 6, 6}}
	r := dice.New(rand.NewPCG(5, 5))

	for range 500 {
		term, _, _ := ResolveCitizenTerm(r, upp, C1, 0, "", "")

		want := 4
		if term.CitizenLifeSucceeded {
			want = 5
		}

		if len(term.SkillsAwarded) > want {
			t.Fatalf(
				"term (succeeded=%v) granted %d skills, want at most %d",
				term.CitizenLifeSucceeded,
				len(term.SkillsAwarded),
				want,
			)
		}
	}
}

func TestResolveCitizenTermDeterminism(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{6, 6, 6, 6, 6, 6}}

	r1 := dice.New(rand.NewPCG(11, 11))
	r2 := dice.New(rand.NewPCG(11, 11))

	term1, job1, hobby1 := ResolveCitizenTerm(r1, upp, C2, 0, "", "")
	term2, job2, hobby2 := ResolveCitizenTerm(r2, upp, C2, 0, "", "")

	if job1 != job2 || hobby1 != hobby2 {
		t.Fatalf("identical seeds produced different job/hobby state: (%q,%q) vs (%q,%q)", job1, hobby1, job2, hobby2)
	}

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

// TestCitizenTableEMatchesBook1P78 pins every one of the 108 cells in
// Book 1 p.78's "E CITIZEN SKILLS AND KNOWLEDGES" table, transcribed
// directly from the page image — full-pinned, not sampled, matching
// TestScoutSkillTableMapping's own "no partial pins" convention.
func TestCitizenTableEMatchesBook1P78(t *testing.T) {
	t.Parallel()

	want := [3][6][6]string{
		{
			{"ACV", "Comms", "High-G", "Steward", "Ordnance", "Naval Arch"},
			{"Jack of all Trades", "Rider", "Sensors", "Forward Observer", "Survival", "Streetwise"},
			{"LTA", "Spines", "Flapper", "Seafarer", "No Skill", "Astrogator"},
			{"WMD", "Leader", "Tracked", "Engineer", "Computer", "Navigation"},
			{"Chef", "Survey", "Animals", "Fluidics", "Bay Weapons", "Explosives"},
			{"Mole", "Dancer", "Tactics", "Launcher", "Magnetics", "Jump Drive"},
		},
		{
			{"Grav", "Artist", "Turrets", "Teamster", "Photonics", "Counsellor"},
			{"Boat", "Legged", "Teacher", "Designer", "Vacc Suit", "Submersible"},
			{"Ship", "Sapper", "Unarmed", "Engineer", "Artillery", "Aeronautics"},
			{"Wing", "Driver", "Exotics", "Language", "Craftsman", "Aquanautics"},
			{"Recon", "Gunner", "Stealth", "Musician", "Gravitics", "Battle Dress"},
			{"Actor", "Blades", "Trainer", "Strategy", "Forensics", "Electronics"},
		},
		{
			{"Flyer", "Zero-G", "Animals", "Maneuver", "Biologics", "Hostile Environment"},
			{"Pilot", "Author", "Liaison", "Polymers", "Ortillery", "Power Systems"},
			{"Rotor", "Broker", "Athlete", "Advocate", "Automotive", "Life Support"},
			{"Admin", "Trader", "Fighter", "Computer", "Bureaucrat", "Slug Thrower"},
			{"Beams", "Sprays", "Wheeled", "Diplomat", "Heavy Weapons", "Fleet Tactics"},
			{"Medic", "Gambler", "Screens", "Mechanic", "Programmer", "Spacecraft"},
		},
	}

	if citizenTableE != want {
		t.Errorf("citizenTableE =\n%v\nwant\n%v", citizenTableE, want)
	}
}

// TestCitizenLifeGrantIsJob pins the odd/even alternation against Book 1
// p.78's own "Citizen Life Roll" table exactly: First=Job, Second=Hobby,
// Third=Job, Fourth=Hobby, Fifth=Job, Sixth=Hobby, Seventh=Job,
// Eighth=Hobby.
func TestCitizenLifeGrantIsJob(t *testing.T) {
	t.Parallel()

	want := []bool{true, false, true, false, true, false, true, false}

	for priorSuccesses, w := range want {
		if got := citizenLifeGrantIsJob(priorSuccesses); got != w {
			t.Errorf("citizenLifeGrantIsJob(%d) = %v, want %v", priorSuccesses, got, w)
		}
	}
}

// TestCitizenLifeSkillGrantReusesPersistedJobName uses priorSuccesses=2
// (the 3rd success — a Job turn per citizenLifeGrantIsJob) with jobSkill
// already set, the realistic case where a name would already be
// persisted: priorSuccesses=0 with a persisted name is an impossible
// combination (nothing could have set it yet).
func TestCitizenLifeSkillGrantReusesPersistedJobName(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))

	grant, ok, job, hobby := citizenLifeSkillGrant(r, 2, "Pilot", "")

	want := SkillLevel{Name: "Pilot", Level: 1, Kind: Skill}
	if !ok || grant != want {
		t.Errorf("citizenLifeSkillGrant(priorSuccesses=2, job=Pilot) = %+v, %v, want %+v, true", grant, ok, want)
	}

	if job != "Pilot" || hobby != "" {
		t.Errorf("returned job/hobby = %q/%q, want unchanged %q/%q", job, hobby, "Pilot", "")
	}
}

// TestCitizenLifeSkillGrantReusesPersistedHobbyName uses priorSuccesses=3
// (the 4th success — a Hobby turn) with hobbySkill already set, for the
// same reason TestCitizenLifeSkillGrantReusesPersistedJobName uses
// priorSuccesses=2 rather than 1.
func TestCitizenLifeSkillGrantReusesPersistedHobbyName(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(2, 2))

	grant, ok, job, hobby := citizenLifeSkillGrant(r, 3, "", "Broker")

	want := SkillLevel{Name: "Broker", Level: 1, Kind: Skill}
	if !ok || grant != want {
		t.Errorf("citizenLifeSkillGrant(priorSuccesses=3, hobby=Broker) = %+v, %v, want %+v, true", grant, ok, want)
	}

	if job != "" || hobby != "Broker" {
		t.Errorf("returned job/hobby = %q/%q, want unchanged %q/%q", job, hobby, "", "Broker")
	}
}

// TestCitizenGrantLevel pins Book 1 p.78's own "Citizen Life Roll" table
// exactly: First=Job-4 (priorSuccesses=0), Second=Hobby-2
// (priorSuccesses=1), every later success (Third through Eighth and
// beyond) = Level 1.
func TestCitizenGrantLevel(t *testing.T) {
	t.Parallel()

	want := []int{4, 2, 1, 1, 1, 1, 1, 1}

	for priorSuccesses, w := range want {
		if got := citizenGrantLevel(priorSuccesses); got != w {
			t.Errorf("citizenGrantLevel(%d) = %d, want %d", priorSuccesses, got, w)
		}
	}
}

// TestCitizenLifeSkillGrantLevelIsOneWhenDeterminedLate is the
// regression test for the bug the code review caught: if an earlier
// same-type roll landed on "No Skill" (leaving the name undetermined),
// a later determining roll must still grant only the level its own
// ordinal calls for (Level 1 here, priorSuccesses=2 is the 3rd success),
// not the elevated Level 4 a naive "level 4 whenever unset" rule would
// wrongly grant.
func TestCitizenLifeSkillGrantLevelIsOneWhenDeterminedLate(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(7, 8))

	for range 2000 {
		grant, ok, _, _ := citizenLifeSkillGrant(r, 2, "", "")
		if !ok {
			continue // "No Skill" roll; nothing to check this trial
		}

		if grant.Level != 1 {
			t.Fatalf("citizenLifeSkillGrant(priorSuccesses=2, job=\"\") granted Level %d, want 1", grant.Level)
		}
	}
}

// flatCitizenTableE returns every non-"No Skill" name in citizenTableE,
// for membership checks below.
func flatCitizenTableE() []string {
	var names []string

	for _, a := range citizenTableE {
		for _, b := range a {
			for _, name := range b {
				if name != "No Skill" {
					names = append(names, name)
				}
			}
		}
	}

	return names
}

// TestCitizenLifeSkillGrantRollsNewJobNameOnFirstSuccess confirms the
// elevated Job-4 starting level (Book 1 p.78's own "First=Job-4") and
// that the rolled name is correctly persisted as the new JobSkill.
func TestCitizenLifeSkillGrantRollsNewJobNameOnFirstSuccess(t *testing.T) {
	t.Parallel()

	names := flatCitizenTableE()
	r := dice.New(rand.NewPCG(3, 3))

	for range 2000 {
		grant, ok, job, hobby := citizenLifeSkillGrant(r, 0, "", "")
		if !ok {
			continue // "No Skill" roll; nothing to check this trial
		}

		if grant.Level != 4 || grant.Kind != Skill {
			t.Fatalf("grant = %+v, want Level 4, Kind Skill", grant)
		}

		if !slices.Contains(names, grant.Name) {
			t.Fatalf("grant.Name = %q, not in citizenTableE", grant.Name)
		}

		if job != grant.Name || hobby != "" {
			t.Fatalf("returned job/hobby = %q/%q, want %q/%q", job, hobby, grant.Name, "")
		}
	}
}

// TestCitizenLifeSkillGrantRollsNewHobbyNameOnFirstSuccess confirms the
// elevated Hobby-2 starting level (Book 1 p.78's own "Second=Hobby-2").
func TestCitizenLifeSkillGrantRollsNewHobbyNameOnFirstSuccess(t *testing.T) {
	t.Parallel()

	names := flatCitizenTableE()
	r := dice.New(rand.NewPCG(4, 4))

	for range 2000 {
		grant, ok, job, hobby := citizenLifeSkillGrant(r, 1, "", "")
		if !ok {
			continue
		}

		if grant.Level != 2 || grant.Kind != Skill {
			t.Fatalf("grant = %+v, want Level 2, Kind Skill", grant)
		}

		if !slices.Contains(names, grant.Name) {
			t.Fatalf("grant.Name = %q, not in citizenTableE", grant.Name)
		}

		if hobby != grant.Name || job != "" {
			t.Fatalf("returned job/hobby = %q/%q, want %q/%q", job, hobby, "", grant.Name)
		}
	}
}

// TestRollCitizenTableENameHitsNoSkillEventually confirms the "No Skill"
// unresolvable entry (A=1, B=3, C=5, a 1-in-108 cell) is actually
// reachable, at roughly the expected rate — mirroring
// rollDeepSpaceBonus's own never-fired-in-N-trials failure-mode test.
func TestRollCitizenTableENameHitsNoSkillEventually(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(5, 6))

	const trials = 50000

	misses := 0

	for range trials {
		if _, ok := rollCitizenTableEName(r); !ok {
			misses++
		}
	}

	gotPct := 100 * float64(misses) / trials
	if wantPct := 100.0 / 108; gotPct < wantPct-0.5 || gotPct > wantPct+0.5 {
		t.Errorf(
			"rollCitizenTableEName returned \"No Skill\" %.2f%% of %d trials, want ~%.2f%%",
			gotPct,
			trials,
			wantPct,
		)
	}
}
