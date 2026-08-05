package character

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

func TestContinueScoutOutcomeBoundaries(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		roll, intChar int
		want          bool
	}{
		{"natural 2 overrides a target below 2", 2, 1, true},
		{"natural 2 always succeeds even at a high target", 2, 12, true},
		{"roll one above a low target fails", 3, 2, false},
		{"exact match succeeds", 7, 7, true},
		{"roll one above target fails", 8, 7, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := continueScoutOutcome(c.roll, c.intChar); got != c.want {
				t.Errorf("continueScoutOutcome(%d, %d) = %v, want %v", c.roll, c.intChar, got, c.want)
			}
		})
	}
}

// TestNextScoutCCRotatesThroughAllThreeBeforeRepeating pins Book 1 p.64's
// generic rule against a fixed upp (no dice involved): each Controlling
// Characteristic "cannot be used again until all of the others in the
// sequence have been used".
func TestNextScoutCCRotatesThroughAllThreeBeforeRepeating(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{5, 9, 3, 0, 0, 0}}
	used := make(map[Position]bool)

	want := []Position{C2, C1, C3, C2}
	for i, w := range want {
		if got := nextScoutCC(upp, used); got != w {
			t.Errorf("call %d: nextScoutCC = %v, want %v", i+1, got, w)
		}
	}
}

// TestNextScoutCCFirstCallMatchesBeginScoutChoice confirms nextScoutCC's
// first-ever call (an empty used set) independently derives the same
// Position BeginScout would pick — both are highestOf(upp, C1, C2, C3)
// over the same untouched characteristic set — which is why
// ResolveScoutCareer doesn't need to thread BeginScout's own returned
// Position into term 1 (see BeginScout's own doc comment).
func TestNextScoutCCFirstCallMatchesBeginScoutChoice(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{4, 11, 7, 0, 0, 0}}

	want := highestOf(upp, C1, C2, C3)
	if got := nextScoutCC(upp, map[Position]bool{}); got != want {
		t.Errorf("nextScoutCC(empty used set) = %v, want %v (BeginScout's own choice)", got, want)
	}
}

// TestNextScoutCCNeverReturnsC4 guards against a future maintainer "fixing"
// this by consulting the generic p.64 table's stale "C1 C2 C3 C4" Scout row
// instead of Scout's own dedicated p.79 box ("Risk & Reward C1 C2 C3" — C4
// reserved for Continue).
func TestNextScoutCCNeverReturnsC4(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{1, 1, 1, 12, 0, 0}}
	used := make(map[Position]bool)

	for i := range 6 {
		if got := nextScoutCC(upp, used); got == C4 {
			t.Fatalf("call %d: nextScoutCC returned C4, want one of C1/C2/C3", i+1)
		}
	}
}

func TestResolveScoutCareerNeverQualifiedReturnsZeroTermsCareer(t *testing.T) {
	t.Parallel()

	upp := UPP{}
	r := dice.New(rand.NewPCG(1, 1))

	career, _ := ResolveScoutCareer(r, upp)

	if career.Name != "Scout" {
		t.Errorf("career.Name = %q, want %q", career.Name, "Scout")
	}

	if len(career.Terms) != 0 {
		t.Errorf("career.Terms = %v, want empty (Begin and Retry both impossible against target 0)", career.Terms)
	}
}

// TestResolveScoutCareerRespectsMaxTermsCap uses a provably immortal
// fixture (Str=Dex=End=Int=12: Risk can never fail since TwoD6's maximum is
// 12, and Continue can never fail for the same reason) to confirm the loop
// still stops at maxCareerTerms rather than running forever.
func TestResolveScoutCareerRespectsMaxTermsCap(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{12, 12, 12, 12, 12, 0}}

	for _, seed := range []uint64{1, 2, 3} {
		r := dice.New(rand.NewPCG(seed, seed))

		career, _ := ResolveScoutCareer(r, upp)
		if len(career.Terms) != maxCareerTerms {
			t.Errorf("seed %d: len(career.Terms) = %d, want %d", seed, len(career.Terms), maxCareerTerms)
		}
	}
}

// TestResolveScoutCareerWithBudgetTruncatesAnImmortalCareer confirms
// resolveCareerLoop's own maxTerms parameter (added for the -age target,
// character/career_chain.go) actually caps the loop rather than just
// documenting an intent — reusing the immortal fixture above (Risk can
// never fail, so an uncapped run always hits the full maxCareerTerms),
// a tighter budget must bind, and bind exactly.
func TestResolveScoutCareerWithBudgetTruncatesAnImmortalCareer(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{12, 12, 12, 12, 12, 0}}

	const budget = 5

	career, _ := resolveScoutCareerWithBudget(dice.New(rand.NewPCG(1, 1)), upp, budget, &agingSimulation{})
	if len(career.Terms) != budget {
		t.Errorf("len(career.Terms) = %d, want %d", len(career.Terms), budget)
	}
}

func TestResolveCareerLoopBudgetEndsBeforeContinueRoll(t *testing.T) {
	t.Parallel()

	continueCalled := false
	terms, _ := resolveCareerLoop(
		dice.New(rand.NewPCG(1, 1)),
		UPP{Characteristics: [6]ehex.Value{8}},
		[]Position{C1},
		func(_ *dice.Roller, upp UPP, _ Position) (Term, UPP) {
			return Term{Length: 4, RiskResult: Unharmed}, upp
		},
		func(_ *dice.Roller, _ UPP) bool {
			continueCalled = true

			return false
		},
		1,
		&agingSimulation{},
		nil,
		nil,
	)

	if len(terms) != 1 {
		t.Fatalf("len(terms) = %d, want 1", len(terms))
	}

	if continueCalled {
		t.Fatal("Continue was rolled after the caller had already chosen to transfer")
	}
}

// TestResolveCareerLoopBeforeTermSubstitutesForResolveTerm is the
// regression test for the beforeTerm hook (#113 item 5, stage 2): when
// it fires, its own Term/UPP must be what lands in the returned Terms
// slice, resolveTerm must not run at all, and nextCC's own rotation
// must not advance — no Controlling Characteristic was used.
func TestResolveCareerLoopBeforeTermSubstitutesForResolveTerm(t *testing.T) {
	t.Parallel()

	resolveTermCalled := false
	terms, _ := resolveCareerLoop(
		dice.New(rand.NewPCG(1, 1)),
		UPP{Characteristics: [6]ehex.Value{8}},
		[]Position{C1, C2},
		func(_ *dice.Roller, upp UPP, _ Position) (Term, UPP) {
			resolveTermCalled = true

			return Term{Length: 4, RiskResult: Unharmed}, upp
		},
		func(_ *dice.Roller, _ UPP) bool { return false },
		1,
		&agingSimulation{},
		nil,
		func(_ *dice.Roller, upp UPP) (Term, UPP, bool) {
			return Term{Length: 4, LaterEducation: true, LaterEducationSchool: "Test University"}, upp, true
		},
	)

	if resolveTermCalled {
		t.Error("resolveTerm was called even though beforeTerm elected Later Education")
	}

	if len(terms) != 1 || !terms[0].LaterEducation || terms[0].LaterEducationSchool != "Test University" {
		t.Fatalf("terms = %+v, want one LaterEducation term naming Test University", terms)
	}

	if terms[0].ControllingCharacteristic != 0 {
		t.Errorf(
			"ControllingCharacteristic = %v, want the zero value — no CC was used",
			terms[0].ControllingCharacteristic,
		)
	}
}

// TestResolveCareerLoopBeforeTermDecliningFallsThroughToResolveTerm
// confirms the hook is opt-in per term, not all-or-nothing: when it
// reports false, the ordinary resolveTerm/nextCC path still runs.
func TestResolveCareerLoopBeforeTermDecliningFallsThroughToResolveTerm(t *testing.T) {
	t.Parallel()

	resolveTermCalled := false
	terms, _ := resolveCareerLoop(
		dice.New(rand.NewPCG(1, 1)),
		UPP{Characteristics: [6]ehex.Value{8}},
		[]Position{C1},
		func(_ *dice.Roller, upp UPP, _ Position) (Term, UPP) {
			resolveTermCalled = true

			return Term{Length: 4, RiskResult: Unharmed}, upp
		},
		func(_ *dice.Roller, _ UPP) bool { return false },
		1,
		&agingSimulation{},
		nil,
		func(_ *dice.Roller, upp UPP) (Term, UPP, bool) {
			return Term{}, upp, false
		},
	)

	if !resolveTermCalled {
		t.Error("resolveTerm was never called even though beforeTerm declined")
	}

	if len(terms) != 1 || terms[0].LaterEducation {
		t.Fatalf("terms = %+v, want one ordinary (non-LaterEducation) term", terms)
	}
}

// TestResolveScoutCareerCCRotationAcrossTerms reuses the immortal fixture:
// since Risk can never fail, C1/C2/C3 all stay tied at 12 for the whole
// career, so highestOf's first-wins-on-tie makes the rotation fully
// predictable across all maxCareerTerms terms — an end-to-end confirmation
// that ResolveScoutCareer actually rotates the Controlling Characteristic
// rather than reusing a single fixed one for every term.
func TestResolveScoutCareerCCRotationAcrossTerms(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{12, 12, 12, 12, 12, 0}}
	r := dice.New(rand.NewPCG(7, 7))

	career, _ := ResolveScoutCareer(r, upp)

	assertCCRotationCycles(t, career.Terms, scoutRiskRewardPositions)
}

func assertCCRotationCycles(t *testing.T, terms []Term, positions []Position) {
	t.Helper()

	for start := 0; start < len(terms); start += len(positions) {
		seen := make(map[Position]bool, len(positions))
		end := min(start+len(positions), len(terms))

		for _, term := range terms[start:end] {
			position := term.ControllingCharacteristic
			if !slices.Contains(positions, position) {
				t.Errorf("ControllingCharacteristic %v is not in %v", position, positions)
			}

			if seen[position] {
				t.Errorf("ControllingCharacteristic %v repeated within rotation cycle", position)
			}

			seen[position] = true
		}
	}
}

// TestResolveScoutCareerPersistsCharacteristicReduction is the regression
// test for Book 1 p.65's "permanently reduced" rule. It uses a fixture
// where C1=C2=C3=6 start tied (so, absent any reduction, the rotation's
// tie-break would always pick C1 first at the start of every three-term
// cycle) and Int=Edu=12 (guaranteeing Begin and Continue always succeed, so
// every trial runs a full, comparable set of terms). Risk against target 6
// can never reduce a characteristic to 0 (Flux's floor is -5, so the worst
// case is 6-5=1) — death is impossible here, isolating the characteristic
// persistence effect from the death-stopping behavior
// TestResolveScoutCareerStopsOnDeath already covers.
//
// Filtered to trials where term 1 actually wounded C1 (RiskResult ==
// Wounded or Disabled — only possible if Risk failed AND Flux left a net
// reduction, as opposed to Risk failing but Flux fully compensating, which
// reports Unharmed with no real change) and terms 2 and 3 left C2/C3
// unchanged (RiskResult == Unharmed): under exactly these conditions, C1's
// value at the start of the second three-term cycle (term 4) is
// mathematically guaranteed to be below C2 and C3's still-6 values, so term
// 4's Controlling Characteristic must be C2 (the first of the two
// remaining tied-at-6 positions, by nextScoutCC's own list order). Without
// this slice's fix — if ResolveScoutCareer didn't thread ResolveScoutTerm's
// returned UPP into the next term — term 4 would incorrectly still be C1,
// since the stale, unreduced upp would still show C1=C2=C3=6 tied.
func TestResolveScoutCareerPersistsCharacteristicReduction(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{6, 6, 6, 12, 12, 0}}
	r := dice.New(rand.NewPCG(23, 29))

	const trials = 3000

	matched := 0

	for range trials {
		career, _ := ResolveScoutCareer(r, upp)
		// A Disabled term (loss>=4, i.e. Flux<=-4) mandatorily ends the
		// career at that term (p.65), same as Dead — a legitimate way for
		// a trial to fall short of 4 terms despite Int=12 guaranteeing
		// Continue, so this skips rather than treating it as a failure.
		if len(career.Terms) < 4 {
			continue
		}

		term1, term2, term3 := career.Terms[0], career.Terms[1], career.Terms[2]

		c1Wounded := term1.RiskResult == Wounded || term1.RiskResult == Disabled
		if !c1Wounded || term2.RiskResult != Unharmed || term3.RiskResult != Unharmed {
			continue
		}

		matched++
	}

	if matched < 20 {
		t.Fatalf(
			"only %d of %d trials matched the filter (C1 wounded, C2/C3 unchanged) — want at least 20",
			matched,
			trials,
		)
	}
}

// TestResolveScoutCareerStopsOnDeath uses a fixture where Str/Dex/End are
// all 1 (Risk against a target of 1 always fails, and reduced = 1+Flux
// clamped to [0,1] reaches 0 — Dead — whenever Flux is negative, about
// 15/36 of the time) while Edu=12 (guarantees Begin via Retry) and Int=12
// (isolates death as the stopping cause, since Continue never fails on its
// own).
func TestResolveScoutCareerStopsOnDeath(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{1, 1, 1, 12, 12, 0}}
	r := dice.New(rand.NewPCG(41, 43))

	const trials = 200

	died := 0

	for range trials {
		career, _ := ResolveScoutCareer(r, upp)
		if len(career.Terms) == 0 {
			t.Fatal("career has zero terms, want Begin to always succeed via Retry (Edu=12)")
		}

		if last := career.Terms[len(career.Terms)-1]; last.RiskResult == Dead {
			died++
		}
	}

	if gotPct := 100 * float64(died) / trials; gotPct < 80 {
		t.Errorf("%.1f%% of careers ended in death, want >80%% (Str/Dex/End=1 makes death likely most terms)", gotPct)
	}
}

func TestResolveScoutCareerDeterminism(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}

	r1 := dice.New(rand.NewPCG(99, 99))
	r2 := dice.New(rand.NewPCG(99, 99))

	career1, updatedUPP1 := ResolveScoutCareer(r1, upp)
	career2, updatedUPP2 := ResolveScoutCareer(r2, upp)

	if len(career1.Terms) != len(career2.Terms) {
		t.Fatalf("identical seeds produced different term counts: %d vs %d", len(career1.Terms), len(career2.Terms))
	}

	if updatedUPP1 != updatedUPP2 {
		t.Fatalf("identical seeds produced different final UPPs: %+v vs %+v", updatedUPP1, updatedUPP2)
	}

	for i := range career1.Terms {
		t1, t2 := career1.Terms[i], career2.Terms[i]

		if t1.RiskResult != t2.RiskResult || t1.RewardResult != t2.RewardResult ||
			t1.ControllingCharacteristic != t2.ControllingCharacteristic {
			t.Fatalf("term %d: identical seeds produced different outcomes: %+v vs %+v", i, t1, t2)
		}

		if !slices.Equal(t1.SkillsAwarded, t2.SkillsAwarded) {
			t.Fatalf(
				"term %d: identical seeds produced different skills: %v vs %v",
				i,
				t1.SkillsAwarded,
				t2.SkillsAwarded,
			)
		}
	}
}
