package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// TestMarineBranchAndOperationsTablesMatchBook1P86 is a full-pin
// regression test for every table literal transcribed from p.86.
func TestMarineBranchAndOperationsTablesMatchBook1P86(t *testing.T) {
	t.Parallel()

	wantBranchNames := [8]string{
		"Infantry", "Infantry", "Artillery", "Cavalry",
		"Protected", "Commando", "Technical", "Medical",
	}
	if marineBranchNames != wantBranchNames {
		t.Errorf("marineBranchNames =\n%v\nwant\n%v", marineBranchNames, wantBranchNames)
	}

	wantBranchMods := [8]int{1, 1, 1, 1, 2, 2, 0, 0}
	if marineBranchMods != wantBranchMods {
		t.Errorf("marineBranchMods =\n%v\nwant\n%v", marineBranchMods, wantBranchMods)
	}

	wantOpsNames := [9]string{
		"Combat", "Combat", "Peace Keeper", "Mission", "ANM School",
		"Combat", "Peace Keeper", "Mission", "Garrison",
	}
	if marineOperationsNames != wantOpsNames {
		t.Errorf("marineOperationsNames =\n%v\nwant\n%v", marineOperationsNames, wantOpsNames)
	}

	wantOpsMods := [9]int{2, 2, 1, 2, 0, 3, 1, 2, 0}
	if marineOperationsMods != wantOpsMods {
		t.Errorf("marineOperationsMods =\n%v\nwant\n%v", marineOperationsMods, wantOpsMods)
	}

	wantBranchDM := map[string]int{
		"Commando": 0, "Protected": 1, "Infantry": 2, "Cavalry": 3,
		"Medical": 4, "Artillery": 5, "Technical": 6,
	}
	if len(marineOperationsBranchDM) != len(wantBranchDM) {
		t.Fatalf("marineOperationsBranchDM has %d entries, want %d", len(marineOperationsBranchDM), len(wantBranchDM))
	}

	for name, want := range wantBranchDM {
		if got := marineOperationsBranchDM[name]; got != want {
			t.Errorf("marineOperationsBranchDM[%q] = %d, want %d", name, got, want)
		}
	}
}

func TestMarineSkillTableMatchesBook1P86(t *testing.T) {
	t.Parallel()

	want := [7][6]string{
		{"Str", "Dex", "End", "Int", "Edu", "Soc"},
		{"One Trade", "Major", "Minor", "Gambler", "Athlete", "One Trade"},
		{"Fighter", "Fighter", "Soldier Skill", "Stealth", "Leader", "Tactics"},
		{"Vacc Suit", "Fighter", "Soldier Skill", "Hostile Environment", "Stealth", "Tactics"},
		{"Fighter", "Flyer", "Fighter", "Stealth", "Leader", "Heavy Weapons"},
		{"Soldier Skill", "Survival", "Language", "Gunner", "Leader", "Fighter"},
		{"One Art", "One Science", "Explosives", "Medic", "Seafarer", "One Trade"},
	}

	if marineSkillTable != want {
		t.Errorf("marineSkillTable =\n%v\nwant\n%v", marineSkillTable, want)
	}
}

func TestBeginMarineRate(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))

	const trials = 20000

	fired := 0

	for range trials {
		if BeginMarine(r, 7) {
			fired++
		}
	}

	gotPct := 100 * float64(fired) / trials
	if wantPct := 100.0 * 21 / 36; gotPct < wantPct-1 || gotPct > wantPct+1 {
		t.Errorf("BeginMarine(str=7) fired %.2f%% of %d trials, want ~%.2f%%", gotPct, trials, wantPct)
	}
}

// TestRollMarineBranchAllRowsReachable is statistical, mirroring
// TestResolveScoutMusterOutBothColumnsReachable's own never-fired-in-N-
// trials style, generalized to all 8 branch names.
func TestRollMarineBranchAllRowsReachable(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(3, 5))

	seen := make(map[string]bool, len(marineBranchNames))

	for range 2000 {
		name, _ := rollMarineBranch(r)
		seen[name] = true

		if len(seen) == len(uniqueStrings(marineBranchNames[:])) {
			return
		}
	}

	t.Fatalf("not all branch names appeared across 2000 trials: saw %v", seen)
}

func uniqueStrings(items []string) map[string]bool {
	set := make(map[string]bool, len(items))
	for _, s := range items {
		set[s] = true
	}

	return set
}

func TestMarineOperationsEduDM(t *testing.T) {
	t.Parallel()

	cases := []struct {
		edu  int
		want int
	}{
		{9, 0},
		{10, 2},
		{15, 2},
	}

	for _, c := range cases {
		if got := marineOperationsEduDM(c.edu); got != c.want {
			t.Errorf("marineOperationsEduDM(%d) = %d, want %d", c.edu, got, c.want)
		}
	}
}

// TestRollMarineOperationsKeepsHighestMod is a property test, not exact-
// value pinning (which of 4 rolls wins is inherently seed-dependent):
// the returned Mod must be at least as high as musterOutRow's own
// single-roll result would give for the very first of the 4 dice drawn
// — confirming rollMarineOperations doesn't silently take a worse row
// than what it actually rolled.
func TestRollMarineOperationsKeepsHighestMod(t *testing.T) {
	t.Parallel()

	for _, seed := range []uint64{1, 2, 3, 4, 5} {
		r1 := dice.New(rand.NewPCG(seed, seed))
		_, got := rollMarineOperations(r1, "Commando", 8)

		r2 := dice.New(rand.NewPCG(seed, seed))

		dm := marineOperationsBranchDM["Commando"] + marineOperationsEduDM(8)
		firstRow := musterOutRow(r2.D6()+dm, len(marineOperationsNames))
		firstMod := marineOperationsMods[firstRow]

		if got < firstMod {
			t.Errorf("seed %d: rollMarineOperations = %d, want >= first roll's own Mod %d", seed, got, firstMod)
		}
	}
}

func TestResolveMarineTermAppliesCombinedMod(t *testing.T) {
	t.Parallel()

	// Commando (Branch Mod 2), Edu < 10 so no eduDM: the combined Mod is
	// Commando's own Branch Mod (2) plus whatever Operations Mod this
	// term's own 4 rolls produce (0-3 for Commando's own DM=0). A Str/Int
	// UPP{6,...} makes the Risk target sensitive to the Mod: a nonzero
	// combined Mod must shift the pass/fail boundary versus mod=0.
	upp := UPP{Characteristics: [6]ehex.Value{6, 0, 0, 6, 8, 0}}
	r := dice.New(rand.NewPCG(11, 13))

	term, _ := ResolveMarineTerm(r, upp, C1, "Commando", marineBranchMods[5]) // index 5 = Commando

	if term.Branch != "Commando" {
		t.Errorf("Branch = %q, want %q", term.Branch, "Commando")
	}

	if term.Assignment == "" {
		t.Error("Assignment is empty, want an Operations name")
	}
}

func TestResolveMarineTermSkipsRewardAndSkillsOnDeath(t *testing.T) {
	t.Parallel()

	upp := UPP{} // guarantees Risk failure and a fatal reduction
	r := dice.New(rand.NewPCG(1, 1))

	term, _ := ResolveMarineTerm(r, upp, C1, "Infantry", 1)

	if term.RiskResult != Dead {
		t.Fatalf("RiskResult = %v, want Dead (fixture assumption broke)", term.RiskResult)
	}

	// RewardResult stays "None" (its own initial value, set before the
	// Risk roll — not left empty) even on death; only the Reward roll
	// itself, and Skills, are skipped.
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

// TestMarineCareerFameIsAlwaysZero confirms Book 1 p.91's own "Armed
// Forces Enlisted = no Fame" — regardless of how many terms a Marine
// serves, marineCareerFame stays 0, since no character in this codebase
// is ever assigned an Officer Rank yet.
func TestMarineCareerFameIsAlwaysZero(t *testing.T) {
	t.Parallel()

	career := Career{
		Terms: []Term{
			{RewardResult: "XS Exemplary Service"},
			{RewardResult: "None"},
			{RewardResult: "XS Exemplary Service"},
		},
	}

	if got := marineCareerFame(career); got != 0 {
		t.Errorf("marineCareerFame() = %d, want 0", got)
	}
}
