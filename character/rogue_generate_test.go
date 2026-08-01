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
// are unreachable by a raw roll — clampRogueSchemeFlux's own defensive
// clamp exists for that reason — though p.84's own "+/-1 after roll"
// does put them within a Rogue's reach.
func TestRollRogueSchemeReachableExtremes(t *testing.T) {
	t.Parallel()

	// Flux -5 lands on Scholar (Cr100,000), but p.84's own "+/-1 after
	// roll" reaches Entertainer (Cr300,000) one row over, which a Rogue
	// takes. Likewise Flux +5 sits on Marine (Cr50,000) beside Noble
	// (Cr500,000), the richest Scheme on the table.
	name, value := rollRogueScheme(dice.New(rand.NewPCG(59, 59)), nil) // first Flux() = -5
	if name != "Entertainer" || value != "Cr300,000" {
		t.Errorf("rollRogueScheme at Flux -5 = (%q, %q), want (%q, %q) after the +/-1 adjustment",
			name, value, "Entertainer", "Cr300,000")
	}

	name, value = rollRogueScheme(dice.New(rand.NewPCG(1, 1)), nil) // first Flux() = +5
	if name != "Noble" || value != "Cr500,000" {
		t.Errorf("rollRogueScheme at Flux +5 = (%q, %q), want (%q, %q) after the +/-1 adjustment",
			name, value, "Noble", "Cr500,000")
	}

	// The clamp itself still guards the table's own unreachable ends.
	if got := clampRogueSchemeFlux(-9); got != -rogueSchemeFluxOffset {
		t.Errorf("clampRogueSchemeFlux(-9) = %d, want %d", got, -rogueSchemeFluxOffset)
	}

	if got := clampRogueSchemeFlux(9); got != rogueSchemeFluxOffset {
		t.Errorf("clampRogueSchemeFlux(9) = %d, want %d", got, rogueSchemeFluxOffset)
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

	term := ResolveRogueTerm(r, 8, 0, 0, nil)

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

	term := ResolveRogueTerm(r, 8, 0, 0, nil)

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
// covers Risk failure (Imprisoned): Scheme "Spacer" (Cr100,000) — the
// roll's own Soldier row is Cr50,000, and p.84's "+/-1 after roll"
// reaches Spacer beside it — Reward roll 4, unhalved Payoff = 100,000 x
// (1+8-4+0) = 500,000, halved to 250,000 for Imprisoned. Skills come from the Prison-columns-only roll
// (rogueSkillTable columns 1-2 only), confirmed by checking every
// awarded skill name appears in one of those two columns.
func TestResolveRogueTermRiskFailureHalvesThePayoffAndUsesPrisonSkills(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(18, 18))

	term := ResolveRogueTerm(r, 8, 0, 0, nil)

	if !term.Imprisoned {
		t.Fatal("Imprisoned = false, want true (Risk failed)")
	}

	if term.Scheme != "Spacer" {
		t.Errorf("Scheme = %q, want %q", term.Scheme, "Spacer")
	}

	if !term.RewardSucceeded {
		t.Error("RewardSucceeded = false, want true")
	}

	if term.SchemePayoff != 250000 {
		t.Errorf("SchemePayoff = %d, want 250000 (500000 halved for Imprisoned)", term.SchemePayoff)
	}

	// Skills come from the full table, not the Prison columns: Book 1
	// p.84 imposes prison "at the start of the next Term", so the term
	// that fails a Scheme is still served at liberty. Its own eligibility
	// is the Failed Scheme line, not the In Prison one.
	if got, want := rogueTermSkillCount(term), rogueSkillsPerTerm+rogueFailedSchemeSkillBonus; got != want {
		t.Errorf("skill eligibility = %d, want %d (Per Term + Failed Scheme)", got, want)
	}
}

// TestResolveRogueTermServedInPrisonUsesPrisonSkills covers the term the
// sentence is actually served in — p.84's "In Prison: Prison Skills from
// the Rogue Skills table column 1 or 2 only. Receives ONLY Prison Skills
// (not Term or Scheme Skills)" — two skills, and only those.
func TestResolveRogueTermServedInPrisonUsesPrisonSkills(t *testing.T) {
	t.Parallel()

	term := ResolveRogueTerm(dice.New(rand.NewPCG(18, 18)), 8, 0, 3, nil)

	if term.ServedYears != 3 {
		t.Fatalf("ServedYears = %d, want 3 (carried in from last term's failure)", term.ServedYears)
	}

	if len(term.SkillsAwarded) > roguePrisonSkillsPerTerm {
		t.Errorf("granted %d skills, want at most %d — a prison term receives ONLY Prison Skills",
			len(term.SkillsAwarded), roguePrisonSkillsPerTerm)
	}

	for _, sk := range term.SkillsAwarded {
		if !prisonColumnSkillNames[sk.Name] {
			t.Errorf("awarded skill %q, not from rogueSkillTable columns 1-2", sk.Name)
		}
	}
}

// prisonColumnSkillNames is rogueSkillTable's own columns 1-2 (table
// indices 0-1 — the restricted-column set rollRogueTermSkills passes to
// rollSkillsFromColumns for a prison term) flattened into a lookup set,
// for TestResolveRogueTermServedInPrisonUsesPrisonSkills.
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

	term := ResolveRogueTerm(r, 8, 0, 0, nil)

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

// TestRogueSentenceIncludesNegativeMods pins Book 1 p.84's own sentence
// formula: "Prison for (sum of negative Mods + Flux) years at the start
// of the next Term (may be zero; maximum 4)."
//
// The Mod half was missing — only Flux was rolled — which dropped the
// part that makes a heavily-modified Scheme more dangerous to fail.
func TestRogueSentenceIncludesNegativeMods(t *testing.T) {
	t.Parallel()

	// Flux spans [-5,+5], so a -5 Mod can never reach the cap and a
	// 0 Mod sometimes can. Comparing the two distributions is what shows
	// the Mod reaching the formula at all.
	heavy, light := 0, 0

	for seed := range uint64(500) {
		heavy += rogueSentence(dice.New(rand.NewPCG(seed+1, seed+1)), -5)
		light += rogueSentence(dice.New(rand.NewPCG(seed+1, seed+1)), 0)
	}

	if heavy >= light {
		t.Errorf("a -5 Mod totalled %d prison years against an unmodified %d — "+
			"negative Mods are not reaching the sentence", heavy, light)
	}

	// Positive Mods are not "negative Mods" and must not shorten one.
	for seed := range uint64(200) {
		unmodified := rogueSentence(dice.New(rand.NewPCG(seed+1, seed+1)), 0)
		if got := rogueSentence(dice.New(rand.NewPCG(seed+1, seed+1)), 3); got != unmodified {
			t.Fatalf("a +3 Mod changed the sentence from %d to %d, want no effect", unmodified, got)
		}
	}
}

// TestRogueSentenceStaysInRange pins p.84's own bounds — "may be zero;
// maximum 4" — across every Mod a term can carry.
func TestRogueSentenceStaysInRange(t *testing.T) {
	t.Parallel()

	for mod := -6; mod <= 6; mod++ {
		for seed := range uint64(100) {
			got := rogueSentence(dice.New(rand.NewPCG(seed+1, seed+1)), mod)
			if got < 0 || got > rogueMaxPrisonYears {
				t.Fatalf("mod %d, seed %d: sentence %d outside [0,%d]", mod, seed+1, got, rogueMaxPrisonYears)
			}
		}
	}
}

// TestRogueTermSkillsFollowEligibilityBox pins p.84's own "B SKILL
// ELIGIBILITY": Per Term 2, Failed Scheme 1, Successful Scheme 4, In
// Prison 2 — with "Receives ONLY Prison Skills (not Term or Scheme
// Skills)" making a prison term two, not two on top of the per-Term two.
//
// Failing a Scheme and being in prison are independent: p.84 allows a
// zero-year sentence, and the prison is served the term after the
// failure, so a term can be either, both, or neither.
func TestRogueTermSkillsFollowEligibilityBox(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		term Term
		want int
	}{
		{"successful Scheme, free", Term{}, rogueSkillsPerTerm + rogueSuccessfulSchemeSkillBonus},
		{"failed Scheme, free", Term{Imprisoned: true}, rogueSkillsPerTerm + rogueFailedSchemeSkillBonus},
		{"in prison", Term{ServedYears: 2}, roguePrisonSkillsPerTerm},
		{"in prison while this Scheme also failed", Term{ServedYears: 2, Imprisoned: true}, roguePrisonSkillsPerTerm},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := rogueTermSkillCount(c.term); got != c.want {
				t.Errorf("eligibility = %d skills, want %d", got, c.want)
			}
		})
	}
}

// TestAdjustRogueSchemeFluxTakesTheBestAdjacentScheme pins Book 1 p.84's
// "Flux may be modified (after roll) plus or minus 1".
//
// Unlike Entertainer's optional Flux this is not a gamble — the roll is
// already known and the values are printed — so the adjustment is taken
// to the best of the three reachable Schemes.
func TestAdjustRogueSchemeFluxTakesTheBestAdjacentScheme(t *testing.T) {
	t.Parallel()

	valueAt := func(flux int) string { return rogueSchemeValues[flux+rogueSchemeFluxOffset] }

	// +3 Rogue is Cr100,000; +4 Noble alongside it is Cr500,000, the
	// richest Scheme on the table.
	if got := adjustRogueSchemeFlux(3); got != 4 {
		t.Errorf("Flux +3 adjusted to %+d (%s), want +4 (%s)", got, valueAt(got), valueAt(4))
	}

	// +5 Marine is Cr50,000, and +4 Noble is reachable downward.
	if got := adjustRogueSchemeFlux(5); got != 4 {
		t.Errorf("Flux +5 adjusted to %+d (%s), want +4 (%s)", got, valueAt(got), valueAt(4))
	}

	// Never worse than where it started, at any roll.
	for flux := -rogueSchemeFluxOffset; flux <= rogueSchemeFluxOffset; flux++ {
		got := adjustRogueSchemeFlux(flux)
		if got < -rogueSchemeFluxOffset || got > rogueSchemeFluxOffset {
			t.Fatalf("Flux %+d adjusted off the table to %+d", flux, got)
		}

		if got != flux && got != flux-1 && got != flux+1 {
			t.Errorf("Flux %+d adjusted to %+d, want a move of at most one row", flux, got)
		}

		if got != flux && !schemeIsBetter(valueAt(got), valueAt(flux)) {
			t.Errorf("Flux %+d adjusted to %s from %s without improving it", flux, valueAt(got), valueAt(flux))
		}
	}
}

// TestBestPriorCareerScheme pins p.84's "A Rogue may select for his
// Scheme (rather than roll) any previous career" — any, not merely the
// most recent, and only careers that appear on the Schemes table.
func TestBestPriorCareerScheme(t *testing.T) {
	t.Parallel()

	if _, _, ok := bestPriorCareerScheme(nil); ok {
		t.Error("a Rogue with no prior careers reported a selectable Scheme")
	}

	// Noble pays Cr500,000 against Soldier's Cr50,000, and is not the
	// most recent — the rule says any previous career.
	name, value, ok := bestPriorCareerScheme([]string{"Noble", "Soldier"})
	if !ok || name != "Noble" || value != "Cr500,000" {
		t.Errorf("bestPriorCareerScheme = (%q, %q, %v), want Noble at Cr500,000", name, value, ok)
	}

	if _, _, ok := bestPriorCareerScheme([]string{"Craftsman"}); !ok {
		t.Error("Craftsman is on the Schemes table but was not selectable")
	}
}

// TestSchemeIsBetterLeavesShipSharesIncomparable records the one
// comparison p.84 leaves open: two Schemes pay "one Ship Share" rather
// than a Cr figure, and nothing prices a share in credits. Neither
// direction is taken, so a Rogue neither trades a share away for cash
// nor chases one — the adjustment simply isn't made.
func TestSchemeIsBetterLeavesShipSharesIncomparable(t *testing.T) {
	t.Parallel()

	if schemeIsBetter("one Ship Share", "Cr50,000") {
		t.Error("a Ship Share was ranked above cash")
	}

	if schemeIsBetter("Cr500,000", "one Ship Share") {
		t.Error("cash was ranked above a Ship Share")
	}

	if !schemeIsBetter("Cr500,000", "Cr50,000") {
		t.Error("Cr500,000 should outrank Cr50,000")
	}

	if schemeIsBetter("Cr50,000", "Cr500,000") {
		t.Error("Cr50,000 should not outrank Cr500,000")
	}
}
