package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
)

// TestRogueSchemeTablesMatchBook1P84 is a full-pin regression test for
// rogueSchemeCareerNames/rogueSchemeValues' own transcription.
func TestRogueSchemeTablesMatchBook1P84(t *testing.T) {
	t.Parallel()

	wantNames := [13]string{
		"Craftsman", "Scholar", "Entertainer", "Citizen", "Scout", "Merchant",
		"Spacer", "Soldier", "Agent", "Rogue", "Noble", "Marine", "Functionary",
	}
	if rogueSchemeCareerNames != wantNames {
		t.Errorf("rogueSchemeCareerNames =\n%v\nwant\n%v", rogueSchemeCareerNames, wantNames)
	}

	wantValues := [13]string{
		"Cr200,000", "Cr100,000", "Cr300,000", "Cr50,000", "one Ship Share",
		"one Ship Share", "Cr100,000", "Cr50,000", "Cr100,000", "Cr100,000",
		"Cr500,000", "Cr50,000", "Cr100,000",
	}
	if rogueSchemeValues != wantValues {
		t.Errorf("rogueSchemeValues =\n%v\nwant\n%v", rogueSchemeValues, wantValues)
	}
}

// TestRogueSkillTableMatchesBook1P84 is a full-pin regression test for
// rogueSkillTable's own transcription, including the "JOT" ->
// "Jack of all Trades" normalization.
func TestRogueSkillTableMatchesBook1P84(t *testing.T) {
	t.Parallel()

	want := [7][6]string{
		{"Str", "Dex", "End", "Int", "Edu", "Soc"},
		{"One Science", "Major", "Minor", "One Art", "One Trade", "Gambler"},
		{"Driver", "Flyer", "Hostile Environment", "High-G", "Vacc Suit", "Navigation"},
		{"Starship Skill", "Pilot", "Engineer", "Zero-G", "Vacc Suit", "Astrogator"},
		{"Trader", "Broker", "Computer", "Jack of all Trades", "Teacher", "Fighter"},
		{"Advocate", "Counsellor", "Language", "Leader", "Streetwise", "Comms"},
		{"One Art", "One Science", "Athlete", "Soldier Skill", "Starship Skill", "One Trade"},
	}

	if rogueSkillTable != want {
		t.Errorf("rogueSkillTable = %v, want %v", rogueSkillTable, want)
	}
}

// TestRogueSucceedsNatural12AlwaysFails confirms the "But, 12 is always
// automatic failure" exception overrides succeedsAgainst regardless of
// how favorable target/mod are — seed 11's own first TwoD6() roll is a
// natural 12, confirmed by direct inspection (dice.New(rand.NewPCG(11,
// 11)) is deterministic).
func TestRogueSucceedsNatural12AlwaysFails(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(11, 11))

	if got := rogueSucceeds(r, 12, 12); got {
		t.Error("rogueSucceeds with a natural 12 = true, want false even against target+mod=24")
	}
}

// TestRogueSucceedsNormalRollComparesNormally confirms a non-12 roll
// falls through to the ordinary succeedsAgainst comparison — seed 1's
// own first TwoD6() roll is 7, confirmed by direct inspection.
func TestRogueSucceedsNormalRollComparesNormally(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		target, mod int
		want        bool
	}{
		{"7 <= 8 succeeds", 8, 0, true},
		{"7 <= 7 succeeds", 7, 0, true},
		{"7 <= 6 fails", 6, 0, false},
		{"7 <= 6+1 succeeds via mod", 6, 1, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			r := dice.New(rand.NewPCG(1, 1))
			if got := rogueSucceeds(r, c.target, c.mod); got != c.want {
				t.Errorf("rogueSucceeds(roll=7, target=%d, mod=%d) = %v, want %v", c.target, c.mod, got, c.want)
			}
		})
	}
}

// TestRogueSucceedsRawReturnsTheRollAlongsideTheOutcome confirms the raw
// roll is returned unmodified in both the natural-12 and normal paths.
func TestRogueSucceedsRawReturnsTheRollAlongsideTheOutcome(t *testing.T) {
	t.Parallel()

	r12 := dice.New(rand.NewPCG(11, 11))

	ok, roll := rogueSucceedsRaw(r12, 12, 12)
	if ok {
		t.Error("rogueSucceedsRaw with a natural 12: ok = true, want false")
	}

	if roll != 12 {
		t.Errorf("rogueSucceedsRaw with a natural 12: roll = %d, want 12", roll)
	}

	r7 := dice.New(rand.NewPCG(1, 1))

	ok, roll = rogueSucceedsRaw(r7, 8, 0)
	if !ok {
		t.Error("rogueSucceedsRaw(roll=7, target=8): ok = false, want true")
	}

	if roll != 7 {
		t.Errorf("rogueSucceedsRaw(roll=7, target=8): roll = %d, want 7", roll)
	}
}

func TestBeginRogue(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1)) // first TwoD6() = 7

	if !BeginRogue(r, 8) {
		t.Error("BeginRogue(cc=8) with roll 7 = false, want true")
	}
}

// TestRollRogueSchemeReachableExtremes confirms the two Flux extremes a
// real Roller can actually produce (-5, +5 — dice.Roller.Flux's own
// documented range, one D6 minus another) index correctly into the
// 13-row table (index 1 = Scholar, index 11 = Marine). The table's own
// index 0 (Craftsman, "Flux -6") and index 12 (Functionary, "Flux +6")
// are unreachable via real dice — rollRogueScheme's own defensive clamp
// exists for exactly that reason, but real dice can never exercise it.
func TestRollRogueSchemeReachableExtremes(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(59, 59)) // first Flux() = -5

	name, value := rollRogueScheme(r)
	if name != "Scholar" || value != "Cr100,000" {
		t.Errorf("rollRogueScheme at Flux -5 = (%q, %q), want (%q, %q)", name, value, "Scholar", "Cr100,000")
	}

	r = dice.New(rand.NewPCG(1, 1)) // first Flux() = +5

	name, value = rollRogueScheme(r)
	if name != "Marine" || value != "Cr50,000" {
		t.Errorf("rollRogueScheme at Flux +5 = (%q, %q), want (%q, %q)", name, value, "Marine", "Cr50,000")
	}
}

// The four ResolveRogueTerm fixtures below were seed-hunted against
// cc=8, mod=0 (see this slice's own plan-file Testing section) and
// verified by direct inspection of the resulting Term before being
// pinned here — not assumed from the formula alone.

// TestResolveRogueTermRiskSuccessRewardSuccessGrantsScaledPayoff covers
// Risk success + Reward success: Scheme "Entertainer" (Cr300,000),
// Reward roll 7, Payoff = 300,000 x (1+8-7+0) = 600,000.
func TestResolveRogueTermRiskSuccessRewardSuccessGrantsScaledPayoff(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(2, 2))

	term := ResolveRogueTerm(r, 8, 0)

	if term.Imprisoned {
		t.Error("Imprisoned = true, want false (Risk succeeded)")
	}

	if term.Scheme != "Entertainer" {
		t.Errorf("Scheme = %q, want %q", term.Scheme, "Entertainer")
	}

	if !term.RewardSucceeded {
		t.Error("RewardSucceeded = false, want true")
	}

	if term.SchemePayoff != 600000 {
		t.Errorf("SchemePayoff = %d, want 600000", term.SchemePayoff)
	}

	if term.SchemeShipShare {
		t.Error("SchemeShipShare = true, want false (this Scheme is cash-valued)")
	}
}

// TestResolveRogueTermRiskSuccessRewardFailureGrantsNoPayoff covers
// Risk success + Reward failure.
func TestResolveRogueTermRiskSuccessRewardFailureGrantsNoPayoff(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(6, 6))

	term := ResolveRogueTerm(r, 8, 0)

	if term.Imprisoned {
		t.Error("Imprisoned = true, want false (Risk succeeded)")
	}

	if term.RewardSucceeded {
		t.Error("RewardSucceeded = true, want false")
	}

	if term.SchemePayoff != 0 {
		t.Errorf("SchemePayoff = %d, want 0 (Reward failed)", term.SchemePayoff)
	}

	if term.SchemeShipShare {
		t.Error("SchemeShipShare = true, want false (Reward failed)")
	}
}

// TestResolveRogueTermRiskFailureHalvesThePayoffAndUsesPrisonSkills
// covers Risk failure (Imprisoned): Scheme "Soldier" (Cr50,000), Reward
// roll 4, unhalved Payoff = 50,000 x (1+8-4+0) = 250,000, halved to
// 125,000 for Imprisoned. Skills come from the Prison-columns-only roll
// (rogueSkillTable columns 1-2 only), confirmed by checking every
// awarded skill name appears in one of those two columns.
func TestResolveRogueTermRiskFailureHalvesThePayoffAndUsesPrisonSkills(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(18, 18))

	term := ResolveRogueTerm(r, 8, 0)

	if !term.Imprisoned {
		t.Fatal("Imprisoned = false, want true (Risk failed)")
	}

	if term.Scheme != "Soldier" {
		t.Errorf("Scheme = %q, want %q", term.Scheme, "Soldier")
	}

	if !term.RewardSucceeded {
		t.Error("RewardSucceeded = false, want true")
	}

	if term.SchemePayoff != 125000 {
		t.Errorf("SchemePayoff = %d, want 125000 (250000 halved for Imprisoned)", term.SchemePayoff)
	}

	if len(term.SkillsAwarded) != roguePrisonSkillsPerTerm {
		t.Fatalf(
			"len(SkillsAwarded) = %d, want %d (roguePrisonSkillsPerTerm)",
			len(term.SkillsAwarded),
			roguePrisonSkillsPerTerm,
		)
	}

	for _, sk := range term.SkillsAwarded {
		if !prisonColumnSkillNames[sk.Name] {
			t.Errorf("awarded skill %q, not from rogueSkillTable columns 1-2", sk.Name)
		}
	}
}

// prisonColumnSkillNames is rogueSkillTable's own columns 1-2 (table
// indices 0-1 — rollRogueSkillFromTable's own r.Uniform(2)-1 picks
// index 0 or 1) flattened into a lookup set, for
// TestResolveRogueTermRiskFailureHalvesThePayoffAndUsesPrisonSkills.
var prisonColumnSkillNames = func() map[string]bool {
	names := make(map[string]bool)

	for _, col := range rogueSkillTable[0:2] {
		for _, name := range col {
			names[name] = true
		}
	}

	// resolveSkillCell substitutes these table entries with a rolled
	// choice or drops them entirely — the substituted names are valid
	// too.
	names["One Art"] = true
	names["One Trade"] = true

	for _, name := range oneArtChoices {
		names[name] = true
	}

	for _, name := range theTradeChoices {
		names[name] = true
	}

	return names
}()

// TestResolveRogueTermShipShareSchemeGrantsNoScaledPayoff covers a
// Ship-Share-valued Scheme's own success path: Scheme "Merchant" (one
// Ship Share), Reward succeeds, and SchemeShipShare is set instead of a
// scaled Payoff.
func TestResolveRogueTermShipShareSchemeGrantsNoScaledPayoff(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(3, 3))

	term := ResolveRogueTerm(r, 8, 0)

	if term.Imprisoned {
		t.Error("Imprisoned = true, want false (Risk succeeded)")
	}

	if term.Scheme != "Merchant" {
		t.Errorf("Scheme = %q, want %q", term.Scheme, "Merchant")
	}

	if !term.RewardSucceeded {
		t.Error("RewardSucceeded = false, want true")
	}

	if !term.SchemeShipShare {
		t.Error("SchemeShipShare = false, want true (Merchant's own Scheme Value is \"one Ship Share\")")
	}

	if term.SchemePayoff != 0 {
		t.Errorf("SchemePayoff = %d, want 0 (Ship Share is a flat grant, not a scaled Payoff)", term.SchemePayoff)
	}
}
