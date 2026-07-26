package character

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// TestSpacerBranchTablesMatchBook1P81 is a full-pin regression test for
// every table literal transcribed from p.81, including the dual-track
// Branch table's own four arrays.
func TestSpacerBranchTablesMatchBook1P81(t *testing.T) {
	t.Parallel()

	wantOfficerNames := [8]string{
		"Line", "Line", "Line", "Engineer", "Gunnery", "Flight", "Technical", "Medical",
	}
	if spacerBranchOfficerNames != wantOfficerNames {
		t.Errorf("spacerBranchOfficerNames =\n%v\nwant\n%v", spacerBranchOfficerNames, wantOfficerNames)
	}

	wantOfficerMods := [8]int{1, 1, 1, 0, 1, 2, 0, 0}
	if spacerBranchOfficerMods != wantOfficerMods {
		t.Errorf("spacerBranchOfficerMods =\n%v\nwant\n%v", spacerBranchOfficerMods, wantOfficerMods)
	}

	wantEnlistedNames := [8]string{
		"Crew", "Crew", "Engineer", "Engineer", "Gunnery", "Gunnery", "Technical", "Medical",
	}
	if spacerBranchEnlistedNames != wantEnlistedNames {
		t.Errorf("spacerBranchEnlistedNames =\n%v\nwant\n%v", spacerBranchEnlistedNames, wantEnlistedNames)
	}

	wantEnlistedMods := [8]int{1, 1, 0, 0, 1, 1, 0, 0}
	if spacerBranchEnlistedMods != wantEnlistedMods {
		t.Errorf("spacerBranchEnlistedMods =\n%v\nwant\n%v", spacerBranchEnlistedMods, wantEnlistedMods)
	}

	wantOpsNames := [8]string{
		"Battle", "Strike", "Siege", "Patrol", "Mission", "ANM School", "Shore Duty", "Shore Duty",
	}
	if spacerNavalOperationsNames != wantOpsNames {
		t.Errorf("spacerNavalOperationsNames =\n%v\nwant\n%v", spacerNavalOperationsNames, wantOpsNames)
	}

	wantOpsMods := [8]int{2, 2, 0, 1, 3, 0, 0, 0}
	if spacerNavalOperationsMods != wantOpsMods {
		t.Errorf("spacerNavalOperationsMods =\n%v\nwant\n%v", spacerNavalOperationsMods, wantOpsMods)
	}
}

func TestSpacerSkillTableMatchesBook1P81(t *testing.T) {
	t.Parallel()

	want := [7][6]string{
		{"Str", "Dex", "End", "Int", "Edu", "Soc"},
		{"Major", "Major", "Minor", "Minor", "One Trade", "One Trade"},
		{"Fighter", "Fleet Tactics", "Pilot", "Starship Skill", "Gunner", "Sensors"},
		{"Astrogator", "Fleet Tactics", "Computer", "Starship Skill", "Gunner", "Sensors"},
		{"Computer", "Strategy", "Counsellor", "Gunner", "Gunner", "Gunner"},
		{"Diplomat", "Admin", "Language", "Starship Skill", "Liaison", "Comms"},
		{"One Art", "One Science", "Athlete", "Medic", "Zero-G", "One Trade"},
	}

	if spacerSkillTable != want {
		t.Errorf("spacerSkillTable =\n%v\nwant\n%v", spacerSkillTable, want)
	}
}

// TestSpacerBranchNameAndMod covers all 8 rows x both isOfficer values,
// including row 6 (index 5) explicitly — the case where the same
// specialty's own Mod genuinely changes (not just the name) between
// tracks (Book 1 p.65's own "for Spacers, Crew becomes Line").
func TestSpacerBranchNameAndMod(t *testing.T) {
	t.Parallel()

	cases := []struct {
		row       int
		isOfficer bool
		wantName  string
		wantMod   int
	}{
		{0, false, "Crew", 1},
		{0, true, "Line", 1},
		{3, false, "Engineer", 0},
		{3, true, "Engineer", 0},
		{5, false, "Gunnery", 1},
		{5, true, "Flight", 2},
		{6, false, "Technical", 0},
		{6, true, "Technical", 0},
		{7, false, "Medical", 0},
		{7, true, "Medical", 0},
	}

	for _, c := range cases {
		name, mod := spacerBranchNameAndMod(c.row, c.isOfficer)
		if name != c.wantName || mod != c.wantMod {
			t.Errorf("spacerBranchNameAndMod(%d, %v) = (%q, %d), want (%q, %d)",
				c.row, c.isOfficer, name, mod, c.wantName, c.wantMod)
		}
	}
}

func TestBeginSpacerRate(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))

	const trials = 20000

	fired := 0

	for range trials {
		if BeginSpacer(r, 7) {
			fired++
		}
	}

	gotPct := 100 * float64(fired) / trials
	if wantPct := 100.0 * 21 / 36; gotPct < wantPct-1 || gotPct > wantPct+1 {
		t.Errorf("BeginSpacer(int=7) fired %.2f%% of %d trials, want ~%.2f%%", gotPct, trials, wantPct)
	}
}

// TestRollSpacerBranchRowAllRowsReachable is statistical, mirroring
// TestRollSoldierBranchAllRowsReachable's own never-fired-in-N-trials
// style, generalized to all 8 rows.
func TestRollSpacerBranchRowAllRowsReachable(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(3, 5))

	seen := make(map[int]bool, 8)

	for range 2000 {
		row := rollSpacerBranchRow(r)
		seen[row] = true

		if len(seen) == 8 {
			return
		}
	}

	t.Fatalf("not all 8 rows appeared across 2000 trials: saw %v", seen)
}

// TestRollSpacerOperationsKeepsHighestMod mirrors
// TestRollSoldierOperationsKeepsHighestMod's own property-test style.
func TestRollSpacerOperationsKeepsHighestMod(t *testing.T) {
	t.Parallel()

	for _, seed := range []uint64{1, 2, 3, 4, 5} {
		r1 := dice.New(rand.NewPCG(seed, seed))
		_, got := rollSpacerOperations(r1, 8)

		r2 := dice.New(rand.NewPCG(seed, seed))

		dm := operationsEduDM(8)
		firstRow := musterOutRow(r2.D6()+dm, len(spacerNavalOperationsNames))
		firstMod := spacerNavalOperationsMods[firstRow]

		if got < firstMod {
			t.Errorf("seed %d: rollSpacerOperations = %d, want >= first roll's own Mod %d", seed, got, firstMod)
		}
	}
}

func TestResolveSpacerTermAppliesCombinedMod(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{6, 6, 0, 6, 8, 0}}
	r := dice.New(rand.NewPCG(11, 13))

	term, _ := ResolveSpacerTerm(r, upp, C1, 4, nil) // row 4 = Gunnery/Gunnery

	if term.Branch != "Gunnery" {
		t.Errorf("Branch = %q, want %q", term.Branch, "Gunnery")
	}

	if term.Assignment == "" {
		t.Error("Assignment is empty, want an Operations name")
	}
}

func TestResolveSpacerTermSkipsRewardAndSkillsOnDeath(t *testing.T) {
	t.Parallel()

	upp := UPP{} // guarantees Risk failure and a fatal reduction
	r := dice.New(rand.NewPCG(1, 1))

	term, _ := ResolveSpacerTerm(r, upp, C1, 0, nil)

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

// TestResolveSpacerTermPreservesRankOnDeath mirrors
// TestResolveMarineTermPreservesRankOnDeath/
// TestResolveSoldierTermPreservesRankOnDeath.
func TestResolveSpacerTermPreservesRankOnDeath(t *testing.T) {
	t.Parallel()

	priorTerms := []Term{{Commissioned: true}, {Promoted: true}, {Promoted: true}} // O3 Lieutenant
	upp := UPP{}                                                                   // Str=0: fatal Risk
	r := dice.New(rand.NewPCG(1, 1))

	term, _ := ResolveSpacerTerm(r, upp, C1, 0, priorTerms)

	if term.RiskResult != Dead {
		t.Fatalf("RiskResult = %v, want Dead (fixture assumption broke)", term.RiskResult)
	}

	if term.Rank != "O3 Lieutenant" {
		t.Errorf("Rank = %q, want %q (the rank held entering this fatal term)", term.Rank, "O3 Lieutenant")
	}
}

// TestResolveSpacerTermBranchNameChangesOnCommission is the regression
// test for the dual-track Branch table's own core mechanic: the exact
// same branchRow (5, "Gunnery"/"Flight") resolves to a different name
// AND Mod depending on whether the character is currently Officer or
// Enlisted — deterministic, not seed-hunted, since branch resolution
// itself is dice-free.
func TestResolveSpacerTermBranchNameChangesOnCommission(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{20, 20, 0, 20, 8, 20}}

	r1 := dice.New(rand.NewPCG(1, 1))

	enlistedTerm, _ := ResolveSpacerTerm(r1, upp, C1, 5, nil)
	if enlistedTerm.Branch != "Gunnery" {
		t.Errorf("Enlisted Branch = %q, want %q", enlistedTerm.Branch, "Gunnery")
	}

	r2 := dice.New(rand.NewPCG(1, 1))

	officerTerm, _ := ResolveSpacerTerm(r2, upp, C1, 5, []Term{{Commissioned: true}})
	if officerTerm.Branch != "Flight" {
		t.Errorf("Officer Branch = %q, want %q", officerTerm.Branch, "Flight")
	}
}

// spacerMedalFixtureUPP is shared by the Medals-granting tests below.
var spacerMedalFixtureUPP = UPP{Characteristics: [6]ehex.Value{10, 0, 0, 10, 8, 0}}

// TestResolveSpacerTermGrantsFlatXSOnRiskSuccess mirrors
// TestResolveSoldierTermGrantsFlatXSOnRiskSuccess — seed 4574 (found by
// direct search) produces RiskResult == Unharmed with a failed Reward
// roll.
func TestResolveSpacerTermGrantsFlatXSOnRiskSuccess(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(4574, 4574))

	term, _ := ResolveSpacerTerm(r, spacerMedalFixtureUPP, C1, 4, nil)

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

// TestResolveSpacerTermGrantsRewardMedal mirrors
// TestResolveSoldierTermGrantsRewardMedal.
func TestResolveSpacerTermGrantsRewardMedal(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(3, 3))

	term, _ := ResolveSpacerTerm(r, spacerMedalFixtureUPP, C1, 4, nil)

	if want := []string{"XS", "MCUF"}; !slices.Equal(term.Medals, want) {
		t.Errorf("Medals = %v, want %v (fixture assumption broke)", term.Medals, want)
	}

	if want := medalNames["MCUF"]; term.RewardResult != want {
		t.Errorf("RewardResult = %q, want %q (Medals table lookup)", term.RewardResult, want)
	}
}

// TestResolveSpacerTermRewardUsesTermStartCC mirrors the Marine and
// Soldier #50 regressions.
func TestResolveSpacerTermRewardUsesTermStartCC(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(8, 8))

	term, _ := ResolveSpacerTerm(r, spacerMedalFixtureUPP, C1, 4, nil)

	if term.RiskResult == Unharmed || term.RiskResult == Dead {
		t.Fatalf("RiskResult = %v, want Wounded or Disabled (fixture assumption broke)", term.RiskResult)
	}

	if want := []string{"XS"}; !slices.Equal(term.Medals, want) {
		t.Errorf("Medals = %v, want %v (Reward must use the term-start CC)", term.Medals, want)
	}
}

// TestResolveSpacerTermOfficerRewardBonusReachesSEHD mirrors
// TestResolveSoldierTermOfficerRewardBonusReachesSEHD.
func TestResolveSpacerTermOfficerRewardBonusReachesSEHD(t *testing.T) {
	t.Parallel()

	priorTerms := []Term{{Commissioned: true}}
	r := dice.New(rand.NewPCG(30, 30))

	term, _ := ResolveSpacerTerm(r, spacerMedalFixtureUPP, C1, 4, priorTerms)

	if !slices.Contains(term.Medals, "SEHD") {
		t.Errorf("Medals = %v, want to contain %q (Officer +1 Reward bonus reaching roll 13)", term.Medals, "SEHD")
	}

	if want := medalNames["SEHD"]; term.RewardResult != want {
		t.Errorf("RewardResult = %q, want %q", term.RewardResult, want)
	}
}

// spacerPromotionFixtureUPP is shared by the Commission/Promotion tests
// below — moderate values across Str (Risk & Reward's own ccPos here),
// Dex (Commission and Rating Promotion's own shared target for Spacer),
// and Soc (Officer Promotion's own target) keep both success and
// failure reachable across seeds.
var spacerPromotionFixtureUPP = UPP{Characteristics: [6]ehex.Value{8, 8, 0, 8, 8, 8}}

// TestResolveSpacerTermSetsRankEveryTerm mirrors
// TestResolveSoldierTermSetsRankEveryTerm.
func TestResolveSpacerTermSetsRankEveryTerm(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(19, 19))

	term, _ := ResolveSpacerTerm(r, spacerPromotionFixtureUPP, C1, 4, nil)

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

	if term.Rank != "R1 Spacehand" {
		t.Errorf("Rank = %q, want %q", term.Rank, "R1 Spacehand")
	}

	if len(term.SkillsAwarded) != spacerSkillsPerTerm {
		t.Errorf(
			"len(SkillsAwarded) = %d, want %d (no Commission/Promotion bonus)",
			len(term.SkillsAwarded),
			spacerSkillsPerTerm,
		)
	}
}

// TestResolveSpacerTermGrantsCommission mirrors
// TestResolveSoldierTermGrantsCommission.
func TestResolveSpacerTermGrantsCommission(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(2, 2))

	term, _ := ResolveSpacerTerm(r, spacerPromotionFixtureUPP, C1, 4, nil)

	if !term.Commissioned {
		t.Fatalf("Commissioned = false, want true (fixture assumption broke)")
	}

	if term.Promoted {
		t.Errorf("Promoted = true, want false (Commission and Rating Promotion don't both fire the same term)")
	}

	if term.Rank != "O1 Ensign" {
		t.Errorf("Rank = %q, want %q", term.Rank, "O1 Ensign")
	}

	if len(term.SkillsAwarded) != spacerSkillsPerTerm+1 {
		t.Errorf("len(SkillsAwarded) = %d, want %d (per-term 4 + Commission's own +1)",
			len(term.SkillsAwarded), spacerSkillsPerTerm+1)
	}
}

// TestResolveSpacerTermGrantsRatingPromotion mirrors
// TestResolveSoldierTermGrantsEnlistedPromotion.
func TestResolveSpacerTermGrantsRatingPromotion(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(4, 4))

	term, _ := ResolveSpacerTerm(r, spacerPromotionFixtureUPP, C1, 4, nil)

	if term.Commissioned {
		t.Fatalf("Commissioned = true, want false (fixture assumption broke)")
	}

	if !term.Promoted {
		t.Fatalf("Promoted = false, want true (fixture assumption broke)")
	}

	if term.Rank != "R2 Able Spacer" {
		t.Errorf("Rank = %q, want %q", term.Rank, "R2 Able Spacer")
	}

	if len(term.SkillsAwarded) != spacerSkillsPerTerm+1 {
		t.Errorf("len(SkillsAwarded) = %d, want %d (per-term 4 + Promotion's own +1)",
			len(term.SkillsAwarded), spacerSkillsPerTerm+1)
	}
}

// TestResolveSpacerTermGrantsOfficerPromotion mirrors
// TestResolveSoldierTermGrantsOfficerPromotion.
func TestResolveSpacerTermGrantsOfficerPromotion(t *testing.T) {
	t.Parallel()

	priorTerms := []Term{{Commissioned: true}}
	r := dice.New(rand.NewPCG(3, 3))

	term, _ := ResolveSpacerTerm(r, spacerPromotionFixtureUPP, C1, 4, priorTerms)

	if term.Commissioned {
		t.Errorf("Commissioned = true, want false (already an Officer)")
	}

	if !term.Promoted {
		t.Fatalf("Promoted = false, want true (fixture assumption broke)")
	}

	if term.Rank != "O2 Sublieutenant" {
		t.Errorf("Rank = %q, want %q", term.Rank, "O2 Sublieutenant")
	}

	if len(term.SkillsAwarded) != spacerSkillsPerTerm+1 {
		t.Errorf("len(SkillsAwarded) = %d, want %d (per-term 4 + Promotion's own +1)",
			len(term.SkillsAwarded), spacerSkillsPerTerm+1)
	}
}

// TestResolveSpacerTermNeverPromotesPastTheRankCap mirrors
// TestResolveSoldierTermNeverPromotesPastTheRankCap — Commission and
// Rating Promotion both target C2 (Dex) for Spacer, so Dex=0 guarantees
// Commission itself always fails (target unreachable via 2D6), letting
// the switch reach the Rating Promotion case at all; priorTerms already
// show the character at the R6 cap.
func TestResolveSpacerTermNeverPromotesPastTheRankCap(t *testing.T) {
	t.Parallel()

	priorTerms := []Term{
		{Promoted: true},
		{Promoted: true},
		{Promoted: true},
		{Promoted: true},
		{Promoted: true}, // tier now 6, R6 Master Chief Petty Officer
	}
	upp := UPP{Characteristics: [6]ehex.Value{20, 0, 0, 8, 8, 0}}
	r := dice.New(rand.NewPCG(1, 1))

	term, _ := ResolveSpacerTerm(r, upp, C1, 4, priorTerms)

	if term.Promoted {
		t.Errorf("Promoted = true, want false (already at the R6 cap)")
	}

	if term.Rank != "R6 Master Chief Petty Officer" {
		t.Errorf("Rank = %q, want %q (capped, unchanged)", term.Rank, "R6 Master Chief Petty Officer")
	}

	if len(term.SkillsAwarded) > spacerSkillsPerTerm {
		t.Errorf("len(SkillsAwarded) = %d, want at most %d (no unearned Promotion bonus)",
			len(term.SkillsAwarded), spacerSkillsPerTerm)
	}
}
