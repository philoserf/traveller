package character

import (
	"math/rand/v2"
	"reflect"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

func TestContinueCitizenOutcomeBoundaries(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		roll int
		want bool
	}{
		{"natural 2 is mandatory continue", 2, true},
		{"just under the fixed target", 9, true},
		{"exact match succeeds", 10, true},
		{"one over the fixed target fails", 11, false},
		{"well over the fixed target fails", 12, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := continueCitizenOutcome(c.roll); got != c.want {
				t.Errorf("continueCitizenOutcome(%d) = %v, want %v", c.roll, got, c.want)
			}
		})
	}
}

// TestNextCCHandlesFourPositions confirms the generalized nextCC
// correctly rotates a 4-element set (Citizen's own citizenLifePositions)
// — not just Scout's already-tested 3-element case.
func TestNextCCHandlesFourPositions(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{5, 9, 3, 7, 0, 0}}
	used := make(map[Position]bool)

	want := []Position{C2, C4, C1, C3, C2}
	for i, w := range want {
		if got := nextCC(upp, citizenLifePositions, used); got != w {
			t.Errorf("call %d: nextCC = %v, want %v", i+1, got, w)
		}
	}
}

func TestResolveCitizenCareerAlwaysProducesAtLeastOneTerm(t *testing.T) {
	t.Parallel()

	upps := []UPP{
		{},
		{Characteristics: [6]ehex.Value{2, 2, 2, 2, 2, 2}},
		{Characteristics: [6]ehex.Value{12, 12, 12, 12, 12, 12}},
	}

	for i, upp := range upps {
		r := dice.New(rand.NewPCG(uint64(i)+1, uint64(i)+1))

		career := ResolveCitizenCareer(r, upp)
		if career.Name != "Citizen" {
			t.Errorf("career.Name = %q, want %q", career.Name, "Citizen")
		}

		if len(career.Terms) < 1 {
			t.Errorf("upp %d: len(career.Terms) = %d, want >= 1 (Begin is Auto)", i, len(career.Terms))
		}
	}
}

// TestResolveCitizenCareerRespectsMaxTermsCap is statistical: Continue's
// fixed target (10) is UPP-independent, so unlike Scout's all-12s
// "immortal" fixture, there's no deterministic way to force every trial
// to reach the cap — instead this confirms the cap is never exceeded,
// and that it's actually reached often enough (Continue fails only on an
// 11 or 12, so per-term success is 33/36; (33/36)^14 ≈ 29.6% of trials,
// ignoring the mandatory-continue-on-2 edge) to trust the upper-bound
// assertion isn't vacuously true.
func TestResolveCitizenCareerRespectsMaxTermsCap(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}
	r := dice.New(rand.NewPCG(13, 17))

	const trials = 2000

	atCap := 0

	for range trials {
		career := ResolveCitizenCareer(r, upp)
		if len(career.Terms) > maxCareerTerms {
			t.Fatalf("len(career.Terms) = %d, want <= %d", len(career.Terms), maxCareerTerms)
		}

		if len(career.Terms) == maxCareerTerms {
			atCap++
		}
	}

	if atCap < 100 {
		t.Fatalf("only %d of %d trials reached the %d-term cap — want at least 100 to trust the upper bound above",
			atCap, trials, maxCareerTerms)
	}
}

// TestResolveCitizenCareerWithBudgetTruncatesALongerNaturalRun confirms
// citizen_loop.go's own hand-rolled loop honors the -age-derived budget
// (character/career_chain.go) exactly like resolveCareerLoop does —
// seed 5 confirmed by direct inspection to run 8 terms uncapped; a
// budget of 3 must produce exactly the first 3 of those same terms, not
// a different run (the same dice draws, just stopped earlier).
func TestResolveCitizenCareerWithBudgetTruncatesALongerNaturalRun(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{7, 7, 7, 7, 7, 0}}

	full := resolveCitizenCareerWithBudget(dice.New(rand.NewPCG(5, 5)), upp, maxCareerTerms)
	if len(full.Terms) != 8 {
		t.Fatalf("full.Terms = %d, want 8 (fixture assumption broken)", len(full.Terms))
	}

	const budget = 3

	truncated := resolveCitizenCareerWithBudget(dice.New(rand.NewPCG(5, 5)), upp, budget)

	if len(truncated.Terms) != budget {
		t.Fatalf("len(truncated.Terms) = %d, want %d", len(truncated.Terms), budget)
	}

	if !reflect.DeepEqual(truncated.Terms, full.Terms[:budget]) {
		t.Fatalf("truncated.Terms = %+v, want the first %d terms of the uncapped run %+v",
			truncated.Terms, budget, full.Terms[:budget])
	}
}

// TestResolveCitizenCareerCCRotation uses a fixture with all four
// characteristics tied, so highestOf's first-wins-on-tie makes the
// rotation fully predictable: term 1 must always be C1 through term 4
// C4 (Begin is Auto and Continue succeeds independently of UPP, so every
// trial reaches at least 4 terms with overwhelming probability — no
// filtering needed, but the loop still checks length defensively).
func TestResolveCitizenCareerCCRotation(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}
	r := dice.New(rand.NewPCG(19, 23))

	want := []Position{C1, C2, C3, C4}

	for range 500 {
		career := ResolveCitizenCareer(r, upp)
		if len(career.Terms) < len(want) {
			continue
		}

		for i, w := range want {
			if got := career.Terms[i].ControllingCharacteristic; got != w {
				t.Fatalf("term %d: ControllingCharacteristic = %v, want %v", i+1, got, w)
			}
		}

		return
	}

	t.Fatal("no trial in 500 reached 4 terms — can't verify CC rotation")
}

func TestResolveCitizenCareerDeterminism(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{6, 7, 8, 9, 0, 0}}

	r1 := dice.New(rand.NewPCG(29, 31))
	r2 := dice.New(rand.NewPCG(29, 31))

	career1 := ResolveCitizenCareer(r1, upp)
	career2 := ResolveCitizenCareer(r2, upp)

	if len(career1.Terms) != len(career2.Terms) {
		t.Fatalf("identical seeds produced different term counts: %d vs %d", len(career1.Terms), len(career2.Terms))
	}

	if career1.JobSkill != career2.JobSkill || career1.HobbySkill != career2.HobbySkill {
		t.Fatalf("identical seeds produced different Job/Hobby: (%q,%q) vs (%q,%q)",
			career1.JobSkill, career1.HobbySkill, career2.JobSkill, career2.HobbySkill)
	}

	for i := range career1.Terms {
		t1, t2 := career1.Terms[i], career2.Terms[i]

		if t1.CitizenLifeSucceeded != t2.CitizenLifeSucceeded ||
			t1.ControllingCharacteristic != t2.ControllingCharacteristic {
			t.Fatalf("term %d: identical seeds produced different outcomes: %+v vs %+v", i, t1, t2)
		}

		if len(t1.SkillsAwarded) != len(t2.SkillsAwarded) {
			t.Fatalf(
				"term %d: identical seeds produced different skill counts: %v vs %v",
				i,
				t1.SkillsAwarded,
				t2.SkillsAwarded,
			)
		}

		for j := range t1.SkillsAwarded {
			if t1.SkillsAwarded[j] != t2.SkillsAwarded[j] {
				t.Fatalf("term %d: identical seeds produced different skills at %d: %+v vs %+v",
					i, j, t1.SkillsAwarded[j], t2.SkillsAwarded[j])
			}
		}
	}
}

// TestResolveCitizenCareerPersistsJobAndHobbyAcrossTerms uses the all-12s
// fixture (Citizen Life always succeeds, since 2D6<=12 always — the same
// guarantee TestResolveCitizenCareerRespectsMaxTermsCap's Scout analog
// relies on) to confirm ResolveCitizenCareer's loop actually carries
// career.JobSkill/HobbySkill forward: term 1 is always a Job success
// (citizenLifeGrantIsJob(0)==true) that determines JobSkill unless it
// happens to roll "No Skill" (~1/108 chance), so JobSkill should end up
// set in the overwhelming majority of trials — if the loop forgot to
// persist citizenLifeSkillGrant's returned name back onto career.JobSkill,
// this would stay empty in every trial instead.
func TestResolveCitizenCareerPersistsJobAndHobbyAcrossTerms(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{12, 12, 12, 12, 12, 12}}
	r := dice.New(rand.NewPCG(37, 41))

	const trials = 300

	jobSet, hobbySet := 0, 0

	for range trials {
		career := ResolveCitizenCareer(r, upp)
		if career.JobSkill != "" {
			jobSet++
		}

		if career.HobbySkill != "" {
			hobbySet++
		}
	}

	if jobSet < trials*95/100 {
		t.Errorf(
			"JobSkill set in %d/%d trials, want at least 95%% (term 1 is always a guaranteed Job success)",
			jobSet,
			trials,
		)
	}

	if hobbySet < trials*80/100 {
		t.Errorf("HobbySkill set in %d/%d trials, want at least 80%%", hobbySet, trials)
	}
}
