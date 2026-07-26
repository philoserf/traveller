package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

func TestEntertainerSpecialtiesMatchBook1P77(t *testing.T) {
	t.Parallel()

	want := [6]string{"Artist", "Actor", "Author", "Dancer", "Musician", "Chef"}
	if entertainerSpecialties != want {
		t.Errorf("entertainerSpecialties = %v, want %v", entertainerSpecialties, want)
	}
}

func TestEntertainerBeginPositionsMatchBook1P77(t *testing.T) {
	t.Parallel()

	want := map[string][2]Position{
		"Artist":   {C3, C4},
		"Actor":    {C2, C3},
		"Author":   {C4, C5},
		"Dancer":   {C2, C3},
		"Musician": {C2, C3},
		"Chef":     {C2, C4},
	}

	if len(entertainerBeginPositions) != len(want) {
		t.Fatalf("entertainerBeginPositions has %d entries, want %d", len(entertainerBeginPositions), len(want))
	}

	for specialty, wantPair := range want {
		if got := entertainerBeginPositions[specialty]; got != wantPair {
			t.Errorf("entertainerBeginPositions[%q] = %v, want %v", specialty, got, wantPair)
		}
	}
}

func TestEntertainerSkillTableMatchesBook1P77(t *testing.T) {
	t.Parallel()

	want := [7][6]string{
		{"Str", "C2", "C3", "Int", "C5", "C6"},
		{"Major", "Major", "Minor", "Minor", "One Trade", "One Trade"},
		{"Zero-G", "Vacc Suit", "Pilot", "Astrogator", "Sensors", "Starship Skill"},
		{"Survey", "Survival", "Hostile Environment", "Animals", "Bureaucrat", "Navigation"},
		{"Broker", "Trader", "Advocate", "Liaison", "Diplomat", "Bureaucrat"},
		{"Broker", "One Art", "Language", "Admin", "One Art", "Bureaucrat"},
		{"One Art", "One Trade", "Athlete", "Medic", "One Trade", "One Trade"},
	}

	if entertainerSkillTable != want {
		t.Errorf("entertainerSkillTable = %v, want %v", entertainerSkillTable, want)
	}
}

// TestRollEntertainerFameTalent confirms the one-2D-roll produces the
// same starting value used for both Fame and Talent — dice-free, since
// rollEntertainerFameTalent's own return is a single int by construction
// (both being set from it is the caller's own doing), so this just pins
// the roll itself, not any equality invariant.
func TestRollEntertainerFameTalent(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1)) // first TwoD6() = 7, confirmed this session

	if got := rollEntertainerFameTalent(r); got != 7 {
		t.Errorf("rollEntertainerFameTalent = %d, want 7", got)
	}
}

// TestBeginEntertainer covers both outcomes at Str/Dex/End/Int/Edu/Soc 8
// — Actor's own target is highestOf(C2, C3), both 8 here, so target 8.
// Seeds 1/4 were confirmed by direct inspection to roll a success/
// failure respectively.
func TestBeginEntertainer(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		r := dice.New(rand.NewPCG(1, 1))
		if !BeginEntertainer(r, upp, "Actor") {
			t.Error("BeginEntertainer = false, want true (seed 1 rolls a success)")
		}
	})

	t.Run("failure", func(t *testing.T) {
		t.Parallel()

		r := dice.New(rand.NewPCG(4, 4))
		if BeginEntertainer(r, upp, "Actor") {
			t.Error("BeginEntertainer = true, want false (seed 4 rolls a failure)")
		}
	})
}

// The ResolveEntertainerTerm fixtures below were seed-hunted and
// verified by direct inspection of the resulting Term before being
// pinned here — not assumed from the mechanic alone.

func TestResolveEntertainerTermFameIncreases(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(132, 132))

	term, fame, talent := ResolveEntertainerTerm(r, 8, 8)

	if !term.FameIncreased {
		t.Fatal("FameIncreased = false, want true")
	}

	if term.FameAfterTerm != 10 {
		t.Errorf("FameAfterTerm = %d, want 10", term.FameAfterTerm)
	}

	if fame != 10 {
		t.Errorf("returned fame = %d, want 10", fame)
	}

	// Talent started 8, +1 for the Fame increase (9), then this fixture's
	// own Risk roll (Disabled) reduces it further to 5 — TalentAfterTerm
	// reflects the post-Risk value, not the pre-Risk +1 alone.
	if term.TalentAfterTerm != talent {
		t.Errorf("term.TalentAfterTerm = %d, returned talent = %d, want equal", term.TalentAfterTerm, talent)
	}

	wantSkillCount := entertainerSkillsPerTerm + entertainerFameIncreaseSkillBonus
	if len(term.SkillsAwarded) > wantSkillCount {
		t.Errorf("len(SkillsAwarded) = %d, want at most %d (base %d + Fame-increase bonus %d)",
			len(term.SkillsAwarded), wantSkillCount, entertainerSkillsPerTerm, entertainerFameIncreaseSkillBonus)
	}
}

func TestResolveEntertainerTermFameDoesNotIncrease(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(2, 2))

	term, fame, talent := ResolveEntertainerTerm(r, 8, 8)

	if term.FameIncreased {
		t.Fatal("FameIncreased = true, want false")
	}

	if term.RiskResult != Unharmed {
		t.Fatalf("RiskResult = %v, want Unharmed", term.RiskResult)
	}

	// Talent is unaffected: no Fame-increase bonus, no Risk reduction.
	if talent != 8 {
		t.Errorf("talent = %d, want 8 (unchanged)", talent)
	}

	if term.RewardResult != "Success" {
		t.Errorf("RewardResult = %q, want %q", term.RewardResult, "Success")
	}

	if fame != 5 {
		t.Errorf("fame = %d, want 5", fame)
	}
}

// TestResolveEntertainerTermWoundedStillRollsReward confirms Reward is
// rolled whenever Risk isn't Dead — a real divergence from Scholar's own
// "only on Unharmed" restriction, per this slice's own plan-file
// Context (no override box exists for Entertainer, so the universal
// Scout/Marine default applies).
func TestResolveEntertainerTermWoundedStillRollsReward(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(5, 5))

	term, _, _ := ResolveEntertainerTerm(r, 8, 8)

	if term.RiskResult != Wounded {
		t.Fatalf("RiskResult = %v, want Wounded", term.RiskResult)
	}

	if term.RewardResult == "" {
		t.Error("RewardResult is empty, want a Reward roll to have occurred despite Wounded")
	}
}

// TestResolveEntertainerTermTalentExhaustedIsDead confirms a Talent
// reduced to 0 reports RiskResult == Dead (the same universal p.65
// mechanic every other career uses) and that Reward is correctly
// skipped, matching the Scout/Marine "unless Dead" convention.
func TestResolveEntertainerTermTalentExhaustedIsDead(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))

	term, _, talent := ResolveEntertainerTerm(r, 2, 2)

	if term.RiskResult != Dead {
		t.Fatalf("RiskResult = %v, want Dead", term.RiskResult)
	}

	if talent != 0 {
		t.Errorf("talent = %d, want 0", talent)
	}

	if term.RewardResult != "" {
		t.Errorf("RewardResult = %q, want empty (Reward skipped on Dead)", term.RewardResult)
	}
}

// TestRollEntertainerFameChangeUsesOptionalFluxes covers Book 1 p.77's
// "Fame +F +F* +F*" — one mandatory Flux and two optional ones. Only the
// mandatory roll was made, so Fame moved across a third of its range.
//
// Checked by range and spread rather than by pinned values, since which
// optional rolls are taken is itself random: one Flux spans [-5,+5],
// three span [-15,+15], so observing any result outside the single-Flux
// range proves the optional ones are reaching the total.
func TestRollEntertainerFameChangeUsesOptionalFluxes(t *testing.T) {
	t.Parallel()

	const (
		singleFluxMax = 5
		allFluxMax    = 15
	)

	sawBeyondOneFlux := false

	for seed := range uint64(400) {
		got := rollEntertainerFameChange(dice.New(rand.NewPCG(seed+1, seed+1)))

		if got < -allFluxMax || got > allFluxMax {
			t.Fatalf("seed %d: Fame change %d is outside three Fluxes' own [-15,+15]", seed+1, got)
		}

		if got < -singleFluxMax || got > singleFluxMax {
			sawBeyondOneFlux = true
		}
	}

	if !sawBeyondOneFlux {
		t.Error("no result exceeded a single Flux's own range across 400 rolls — " +
			"the two optional Fluxes are not reaching the total")
	}
}

// TestEntertainerTakesComeback pins when a fading Entertainer stages
// p.77's Comeback ("Reset Fame to 2D; Talent is unchanged"). The reset
// is worth taking exactly when Fame has fallen below what 2D pays on
// average, and never when it hasn't.
func TestEntertainerTakesComeback(t *testing.T) {
	t.Parallel()

	for fame := range 7 {
		if !entertainerTakesComeback(fame) {
			t.Errorf("Fame %d: no Comeback, want one (below 2D's own average of %d)", fame, twoD6Expectation)
		}
	}

	for _, fame := range []int{twoD6Expectation, 8, 12, 20} {
		if entertainerTakesComeback(fame) {
			t.Errorf("Fame %d: Comeback taken, want none — a reset to 2D would not improve it", fame)
		}
	}
}
