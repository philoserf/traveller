package character

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// TestSoldierBranchAndOperationsTablesMatchBook1P82 is a full-pin
// regression test for every table literal transcribed from p.82.
func TestSoldierBranchAndOperationsTablesMatchBook1P82(t *testing.T) {
	t.Parallel()

	wantBranchNames := [8]string{
		"Infantry", "Infantry", "Artillery", "Cavalry",
		"Protected", "Protected", "Technical", "Medical",
	}
	if soldierBranchNames != wantBranchNames {
		t.Errorf("soldierBranchNames =\n%v\nwant\n%v", soldierBranchNames, wantBranchNames)
	}

	wantBranchMods := [8]int{1, 1, 1, 1, 2, 2, 0, 0}
	if soldierBranchMods != wantBranchMods {
		t.Errorf("soldierBranchMods =\n%v\nwant\n%v", soldierBranchMods, wantBranchMods)
	}

	wantOpsNames := [9]string{
		"Combat", "Combat", "Peace Keeper", "Mission", "ANM School",
		"Combat", "Peace Keeper", "Mission", "Base",
	}
	if soldierOperationsNames != wantOpsNames {
		t.Errorf("soldierOperationsNames =\n%v\nwant\n%v", soldierOperationsNames, wantOpsNames)
	}

	wantOpsMods := [9]int{2, 2, 1, 2, 0, 3, 1, 2, 0}
	if soldierOperationsMods != wantOpsMods {
		t.Errorf("soldierOperationsMods =\n%v\nwant\n%v", soldierOperationsMods, wantOpsMods)
	}

	wantBranchDM := map[string]int{
		"Protected": 0, "Infantry": 1, "Cavalry": 3,
		"Medical": 4, "Artillery": 5, "Technical": 6,
	}
	if len(soldierOperationsBranchDM) != len(wantBranchDM) {
		t.Fatalf("soldierOperationsBranchDM has %d entries, want %d", len(soldierOperationsBranchDM), len(wantBranchDM))
	}

	for name, want := range wantBranchDM {
		if got := soldierOperationsBranchDM[name]; got != want {
			t.Errorf("soldierOperationsBranchDM[%q] = %d, want %d", name, got, want)
		}
	}
}

func TestSoldierSkillTableMatchesBook1P82(t *testing.T) {
	t.Parallel()

	want := [7][6]string{
		{"Str", "Dex", "End", "Int", "Edu", "Soc"},
		{"Major", "Major", "Minor", "Minor", "One Trade", "One Trade"},
		{"Fighter", "Vacc Suit", "Fighter", "Stealth", "Leader", "Tactics"},
		{"Admin", "Fighter", "Hostile Environment", "Animals", "Liaison", "Navigation"},
		{"Fighter", "Vacc Suit", "Driver", "Stealth", "Heavy Weapons", "Sensors"},
		{"Soldier Skill", "Liaison", "Language", "Soldier Skill", "Computer", "Tactics"},
		{"One Art", "One Science", "Explosives", "Medic", "Seafarer", "One Trade"},
	}

	if soldierSkillTable != want {
		t.Errorf("soldierSkillTable =\n%v\nwant\n%v", soldierSkillTable, want)
	}
}

func TestBeginSoldierRate(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))

	const trials = 20000

	fired := 0

	for range trials {
		if BeginSoldier(r, 7) {
			fired++
		}
	}

	gotPct := 100 * float64(fired) / trials
	if wantPct := 100.0 * 21 / 36; gotPct < wantPct-1 || gotPct > wantPct+1 {
		t.Errorf("BeginSoldier(str=7) fired %.2f%% of %d trials, want ~%.2f%%", gotPct, trials, wantPct)
	}
}

// TestRollSoldierBranchAllRowsReachable is statistical, mirroring
// TestRollMarineBranchAllRowsReachable's own never-fired-in-N-trials
// style, generalized to all 8 branch names.
func TestRollSoldierBranchAllRowsReachable(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(3, 5))

	seen := make(map[string]bool, len(soldierBranchNames))

	for range 2000 {
		name, _ := rollSoldierBranch(r)
		seen[name] = true

		if len(seen) == len(uniqueStrings(soldierBranchNames[:])) {
			return
		}
	}

	t.Fatalf("not all branch names appeared across 2000 trials: saw %v", seen)
}

// TestRollSoldierOperationsKeepsHighestMod mirrors
// TestRollMarineOperationsKeepsHighestMod's own property-test style.
func TestRollSoldierOperationsKeepsHighestMod(t *testing.T) {
	t.Parallel()

	for _, seed := range []uint64{1, 2, 3, 4, 5} {
		r1 := dice.New(rand.NewPCG(seed, seed))
		_, got := rollSoldierOperations(r1, "Cavalry", 8)

		r2 := dice.New(rand.NewPCG(seed, seed))

		dm := soldierOperationsBranchDM["Cavalry"] + operationsEduDM(8)
		firstRow := musterOutRow(r2.D6()+dm, len(soldierOperationsNames))
		firstMod := soldierOperationsMods[firstRow]

		if got < firstMod {
			t.Errorf("seed %d: rollSoldierOperations = %d, want >= first roll's own Mod %d", seed, got, firstMod)
		}
	}
}

func TestResolveSoldierTermAppliesCombinedMod(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{6, 0, 6, 6, 8, 0}}
	r := dice.New(rand.NewPCG(11, 13))

	term, _ := ResolveSoldierTerm(r, upp, C1, "Cavalry", soldierBranchMods[3], nil) // index 3 = Cavalry

	if term.Branch != "Cavalry" {
		t.Errorf("Branch = %q, want %q", term.Branch, "Cavalry")
	}

	if term.Assignment == "" {
		t.Error("Assignment is empty, want an Operations name")
	}
}

func TestResolveSoldierTermSkipsRewardAndSkillsOnDeath(t *testing.T) {
	t.Parallel()

	upp := UPP{} // guarantees Risk failure and a fatal reduction
	r := dice.New(rand.NewPCG(1, 1))

	term, _ := ResolveSoldierTerm(r, upp, C1, "Infantry", 1, nil)

	if term.RiskResult != Dead {
		t.Fatalf("RiskResult = %v, want Dead (fixture assumption broke)", term.RiskResult)
	}

	if term.RewardResult != "None" {
		t.Errorf(
			"RewardResult = %q, want %q (Dead skips the Reward roll, but the initial value stands)",
			term.RewardResult,
			"None",
		)
	}

	if term.SkillsAwarded != nil {
		t.Errorf("SkillsAwarded = %v, want nil (Dead skips Skills)", term.SkillsAwarded)
	}
}

// TestResolveSoldierTermPreservesRankOnDeath mirrors
// TestResolveMarineTermPreservesRankOnDeath.
func TestResolveSoldierTermPreservesRankOnDeath(t *testing.T) {
	t.Parallel()

	priorTerms := []Term{{Commissioned: true}, {Promoted: true}, {Promoted: true}} // O3 Captain
	upp := UPP{}                                                                   // Str=0: fatal Risk
	r := dice.New(rand.NewPCG(1, 1))

	term, _ := ResolveSoldierTerm(r, upp, C1, "Infantry", 1, priorTerms)

	if term.RiskResult != Dead {
		t.Fatalf("RiskResult = %v, want Dead (fixture assumption broke)", term.RiskResult)
	}

	if term.Rank != "O3 Captain" {
		t.Errorf("Rank = %q, want %q (the rank held entering this fatal term)", term.Rank, "O3 Captain")
	}
}

// soldierMedalFixtureUPP is shared by the Medals-granting tests below —
// Str/End 10 make Risk failures/successes both reachable across seeds.
var soldierMedalFixtureUPP = UPP{Characteristics: [6]ehex.Value{10, 0, 10, 10, 8, 0}}

// TestResolveSoldierTermGrantsFlatXSOnRiskSuccess mirrors
// TestResolveMarineTermGrantsFlatXSOnRiskSuccess — seed 39 (found by
// direct search) produces RiskResult == Unharmed with a failed Reward
// roll.
func TestResolveSoldierTermGrantsFlatXSOnRiskSuccess(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(39, 39))

	term, _ := ResolveSoldierTerm(r, soldierMedalFixtureUPP, C1, "Cavalry", 0, nil)

	if term.RiskResult != Unharmed {
		t.Fatalf("RiskResult = %v, want Unharmed (fixture assumption broke)", term.RiskResult)
	}

	if want := []string{"XS"}; !slices.Equal(term.Medals, want) {
		t.Errorf("Medals = %v, want %v (flat Risk-success grant, Reward failed)", term.Medals, want)
	}

	if term.RewardResult != "None" {
		t.Errorf("RewardResult = %q, want %q (Reward failed)", term.RewardResult, "None")
	}
}

// TestResolveSoldierTermGrantsRewardMedal mirrors
// TestResolveMarineTermGrantsRewardMedal — seed 3 (found by direct
// search) resolves to MCUF.
func TestResolveSoldierTermGrantsRewardMedal(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(3, 3))

	term, _ := ResolveSoldierTerm(r, soldierMedalFixtureUPP, C1, "Cavalry", 0, nil)

	if want := []string{"XS", "MCUF"}; !slices.Equal(term.Medals, want) {
		t.Errorf("Medals = %v, want %v (fixture assumption broke)", term.Medals, want)
	}

	if want := medalNames["MCUF"]; term.RewardResult != want {
		t.Errorf("RewardResult = %q, want %q (Medals table lookup)", term.RewardResult, want)
	}
}

// TestResolveSoldierTermRewardUsesTermStartCC mirrors the Marine #50
// regression: Risk reduces the CC, but Reward still succeeds against
// the term-start value.
func TestResolveSoldierTermRewardUsesTermStartCC(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(8, 8))

	term, _ := ResolveSoldierTerm(r, soldierMedalFixtureUPP, C1, "Cavalry", 0, nil)

	if term.RiskResult == Unharmed || term.RiskResult == Dead {
		t.Fatalf("RiskResult = %v, want Wounded or Disabled (fixture assumption broke)", term.RiskResult)
	}

	if want := []string{"XS"}; !slices.Equal(term.Medals, want) {
		t.Errorf("Medals = %v, want %v (Reward must use the term-start CC)", term.Medals, want)
	}
}

// TestResolveSoldierTermOfficerRewardBonusReachesSEHD mirrors
// TestResolveMarineTermOfficerRewardBonusReachesSEHD — seed 30 (found by
// direct search, priorTerms showing a Commissioned character) rolls a
// natural 12 that resolves to SEHD only once the Officer bonus applies.
func TestResolveSoldierTermOfficerRewardBonusReachesSEHD(t *testing.T) {
	t.Parallel()

	priorTerms := []Term{{Commissioned: true}}
	r := dice.New(rand.NewPCG(30, 30))

	term, _ := ResolveSoldierTerm(r, soldierMedalFixtureUPP, C1, "Cavalry", 0, priorTerms)

	if !slices.Contains(term.Medals, "SEHD") {
		t.Errorf("Medals = %v, want to contain %q (Officer +1 Reward bonus reaching roll 13)", term.Medals, "SEHD")
	}

	if want := medalNames["SEHD"]; term.RewardResult != want {
		t.Errorf("RewardResult = %q, want %q", term.RewardResult, want)
	}
}

// soldierPromotionFixtureUPP is shared by the Commission/Promotion tests
// below — moderate values across Str (Risk & Reward's own ccPos here),
// End (Commission and Enlisted Promotion's own shared target for
// Soldier), and Soc (Officer Promotion's own target) keep both success
// and failure reachable across seeds.
var soldierPromotionFixtureUPP = UPP{Characteristics: [6]ehex.Value{8, 0, 8, 8, 8, 8}}

// TestResolveSoldierTermSetsRankEveryTerm mirrors
// TestResolveMarineTermSetsRankEveryTerm — seed 19 (found by direct
// search) produces neither Commission nor Promotion for a fresh
// character, who should show Book 1 p.65's own starting rank, "S1
// Private".
func TestResolveSoldierTermSetsRankEveryTerm(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(19, 19))

	term, _ := ResolveSoldierTerm(r, soldierPromotionFixtureUPP, C1, "Cavalry", 0, nil)

	if term.RiskResult == Dead {
		t.Fatalf("RiskResult = Dead (fixture assumption broke)")
	}

	if term.Commissioned || term.Promoted {
		t.Fatalf(
			"Commissioned=%v Promoted=%v, want both false (fixture assumption broke)",
			term.Commissioned,
			term.Promoted,
		)
	}

	if term.Rank != "S1 Private" {
		t.Errorf("Rank = %q, want %q", term.Rank, "S1 Private")
	}

	if len(term.SkillsAwarded) != soldierSkillsPerTerm {
		t.Errorf(
			"len(SkillsAwarded) = %d, want %d (no Commission/Promotion bonus)",
			len(term.SkillsAwarded),
			soldierSkillsPerTerm,
		)
	}
}

// TestResolveSoldierTermGrantsCommission mirrors
// TestResolveMarineTermGrantsCommission — seed 2 (found by direct
// search) succeeds, moving the character to the Officer track at O1.
func TestResolveSoldierTermGrantsCommission(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(2, 2))

	term, _ := ResolveSoldierTerm(r, soldierPromotionFixtureUPP, C1, "Cavalry", 0, nil)

	if !term.Commissioned {
		t.Fatalf("Commissioned = false, want true (fixture assumption broke)")
	}

	if term.Promoted {
		t.Errorf("Promoted = true, want false (Commission and Enlisted Promotion don't both fire the same term)")
	}

	if term.Rank != "O1 2nd Lieutenant" {
		t.Errorf("Rank = %q, want %q", term.Rank, "O1 2nd Lieutenant")
	}

	if len(term.SkillsAwarded) != soldierSkillsPerTerm+1 {
		t.Errorf("len(SkillsAwarded) = %d, want %d (per-term 4 + Commission's own +1)",
			len(term.SkillsAwarded), soldierSkillsPerTerm+1)
	}
}

// TestResolveSoldierTermGrantsEnlistedPromotion mirrors
// TestResolveMarineTermGrantsEnlistedPromotion — seed 4 (found by direct
// search) succeeds via Enlisted Promotion (targeting C3, End, for
// Soldier — not C1 like Marine).
func TestResolveSoldierTermGrantsEnlistedPromotion(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(4, 4))

	term, _ := ResolveSoldierTerm(r, soldierPromotionFixtureUPP, C1, "Cavalry", 0, nil)

	if term.Commissioned {
		t.Fatalf("Commissioned = true, want false (fixture assumption broke)")
	}

	if !term.Promoted {
		t.Fatalf("Promoted = false, want true (fixture assumption broke)")
	}

	if term.Rank != "S2 Corporal" {
		t.Errorf("Rank = %q, want %q", term.Rank, "S2 Corporal")
	}

	if len(term.SkillsAwarded) != soldierSkillsPerTerm+1 {
		t.Errorf("len(SkillsAwarded) = %d, want %d (per-term 4 + Promotion's own +1)",
			len(term.SkillsAwarded), soldierSkillsPerTerm+1)
	}
}

// TestResolveSoldierTermGrantsOfficerPromotion mirrors
// TestResolveMarineTermGrantsOfficerPromotion — seed 3 (found by direct
// search) succeeds via Officer Promotion (targeting Soc for Soldier —
// not Int like Marine).
func TestResolveSoldierTermGrantsOfficerPromotion(t *testing.T) {
	t.Parallel()

	priorTerms := []Term{{Commissioned: true}}
	r := dice.New(rand.NewPCG(3, 3))

	term, _ := ResolveSoldierTerm(r, soldierPromotionFixtureUPP, C1, "Cavalry", 0, priorTerms)

	if term.Commissioned {
		t.Errorf("Commissioned = true, want false (already an Officer)")
	}

	if !term.Promoted {
		t.Fatalf("Promoted = false, want true (fixture assumption broke)")
	}

	if term.Rank != "O2 1st Lieutenant" {
		t.Errorf("Rank = %q, want %q", term.Rank, "O2 1st Lieutenant")
	}

	if len(term.SkillsAwarded) != soldierSkillsPerTerm+1 {
		t.Errorf("len(SkillsAwarded) = %d, want %d (per-term 4 + Promotion's own +1)",
			len(term.SkillsAwarded), soldierSkillsPerTerm+1)
	}
}

// TestResolveSoldierTermNeverPromotesPastTheRankCap mirrors
// TestResolveMarineTermNeverPromotesPastTheRankCap — deterministic, not
// seed-hunted. Unlike Marine (whose Commission and Enlisted Promotion
// target different characteristics, C3 and C1), Soldier's own Commission
// and Enlisted Promotion both target C3 (End) — so End=0 guarantees
// Commission itself always fails (target unreachable via 2D6), letting
// the switch reach the Enlisted Promotion case at all; priorTerms
// already show the character at the S6 cap, so that case's own
// `tier < len(soldierEnlistedRankNames)` guard is false and short-
// circuits before any roll — Promoted must stay false regardless of
// what any hypothetical roll would have been.
func TestResolveSoldierTermNeverPromotesPastTheRankCap(t *testing.T) {
	t.Parallel()

	priorTerms := []Term{
		{Promoted: true},
		{Promoted: true},
		{Promoted: true},
		{Promoted: true},
		{Promoted: true}, // tier now 6, S6 Sergeant Major
	}
	upp := UPP{Characteristics: [6]ehex.Value{20, 0, 0, 8, 8, 0}}
	r := dice.New(rand.NewPCG(1, 1))

	term, _ := ResolveSoldierTerm(r, upp, C1, "Cavalry", 0, priorTerms)

	if term.Promoted {
		t.Errorf("Promoted = true, want false (already at the S6 cap)")
	}

	if term.Rank != "S6 Sergeant Major" {
		t.Errorf("Rank = %q, want %q (capped, unchanged)", term.Rank, "S6 Sergeant Major")
	}

	if len(term.SkillsAwarded) > soldierSkillsPerTerm {
		t.Errorf("len(SkillsAwarded) = %d, want at most %d (no unearned Promotion bonus)",
			len(term.SkillsAwarded), soldierSkillsPerTerm)
	}
}
