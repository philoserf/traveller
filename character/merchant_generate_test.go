package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

func TestMerchantRankNamesMatchBook1P80(t *testing.T) {
	t.Parallel()

	wantEnlisted := [4]string{"Temp", "Spacehand", "Steward Apprentice", "Drive Helper"}
	if merchantEnlistedRankNames != wantEnlisted {
		t.Errorf("merchantEnlistedRankNames = %v, want %v", merchantEnlistedRankNames, wantEnlisted)
	}

	wantOfficer := [6]string{
		"Fourth Officer", "Third Officer", "Second Officer",
		"First Officer", "Captain", "Senior Captain",
	}
	if merchantOfficerRankNames != wantOfficer {
		t.Errorf("merchantOfficerRankNames = %v, want %v", merchantOfficerRankNames, wantOfficer)
	}
}

func TestMerchantSkillTableMatchesBook1P80(t *testing.T) {
	t.Parallel()

	want := [7][6]string{
		{"Str", "C2", "C3", "Int", "C5", "C6"},
		{"Major", "Major", "Minor", "Minor", "One Trade", "One Trade"},
		{"Astrogator", "Pilot", "Medic", "Sensors", "Steward", "Gunner"},
		{"Broker", "Trader", "Diplomat", "Admin", "Steward", "Trader"},
		{"Broker", "Trader", "Diplomat", "Advocate", "Steward", "Comms"},
		{"Broker", "Admin", "Language", "Starship Skill", "Jack of all Trades", "Vacc Suit"},
		{"One Art", "One Science", "Computer", "Comms", "Medic", "One Trade"},
	}

	if merchantSkillTable != want {
		t.Errorf("merchantSkillTable = %v, want %v", merchantSkillTable, want)
	}
}

// TestBeginMerchant covers all three of Book 1 p.80's own cascading
// Begin outcomes. Seeds were confirmed by direct inspection: seed 1
// against an all-8 UPP rolls a success against Int (Officer); seed 2
// against an Int-2 UPP fails Int but succeeds against Dex (Spacehand);
// seed 1 against a Dex-1/Int-1 UPP fails both (Temp, automatic).
func TestBeginMerchant(t *testing.T) {
	t.Parallel()

	t.Run("Officer", func(t *testing.T) {
		t.Parallel()

		upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}
		r := dice.New(rand.NewPCG(1, 1))

		isOfficer, tier := BeginMerchant(r, upp)
		if !isOfficer || tier != 1 {
			t.Errorf("BeginMerchant = (%v, %d), want (true, 1)", isOfficer, tier)
		}
	})

	t.Run("Spacehand", func(t *testing.T) {
		t.Parallel()

		upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 2, 8, 8}}
		r := dice.New(rand.NewPCG(2, 2))

		isOfficer, tier := BeginMerchant(r, upp)
		if isOfficer || tier != 1 {
			t.Errorf("BeginMerchant = (%v, %d), want (false, 1)", isOfficer, tier)
		}
	})

	t.Run("Temp", func(t *testing.T) {
		t.Parallel()

		upp := UPP{Characteristics: [6]ehex.Value{8, 1, 8, 1, 8, 8}}
		r := dice.New(rand.NewPCG(1, 1))

		isOfficer, tier := BeginMerchant(r, upp)
		if isOfficer || tier != 0 {
			t.Errorf("BeginMerchant = (%v, %d), want (false, 0)", isOfficer, tier)
		}
	})
}

func TestMerchantPromotionMod(t *testing.T) {
	t.Parallel()

	cases := []struct {
		intChar int
		want    int
	}{
		{7, 0},
		{8, 3},
		{12, 3},
	}

	for _, c := range cases {
		if got := merchantPromotionMod(c.intChar); got != c.want {
			t.Errorf("merchantPromotionMod(%d) = %d, want %d", c.intChar, got, c.want)
		}
	}
}

func TestMerchantRankName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		isOfficer bool
		tier      int
		want      string
	}{
		{"RX Temp", false, 0, "RX Temp"},
		{"R0 Spacehand", false, 1, "R0 Spacehand"},
		{"R2 Drive Helper", false, 3, "R2 Drive Helper"},
		{"M1 Fourth Officer", true, 1, "M1 Fourth Officer"},
		{"M6 Senior Captain", true, 6, "M6 Senior Captain"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := merchantRankName(c.isOfficer, c.tier); got != c.want {
				t.Errorf("merchantRankName(%v, %d) = %q, want %q", c.isOfficer, c.tier, got, c.want)
			}
		})
	}
}

func TestMerchantRewardCount(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		terms []Term
		want  int
	}{
		{"no terms", nil, 0},
		{"one success", []Term{{RewardResult: "1 Ship Share"}}, 1},
		{"one failure", []Term{{RewardResult: "None"}}, 0},
		{"Dead term with no RewardResult at all", []Term{{RiskResult: Dead}}, 0},
		{
			"mixed",
			[]Term{
				{RewardResult: "1 Ship Share"},
				{RewardResult: "None"},
				{RewardResult: "2 Ship Shares"},
			},
			2,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := merchantRewardCount(c.terms); got != c.want {
				t.Errorf("merchantRewardCount(%v) = %d, want %d", c.terms, got, c.want)
			}
		})
	}
}

// The ResolveMerchantTerm fixtures below were seed-hunted against an
// all-8 UPP and verified by direct inspection of the resulting Term
// before being pinned here — not assumed from the mechanic alone.
var upp88 = UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}

func TestResolveMerchantTermOfficerPromotion(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(7, 7))

	term, _, isOfficer, tier := ResolveMerchantTerm(r, upp88, C1, true, 1, nil)

	if !term.Promoted {
		t.Fatal("Promoted = false, want true")
	}

	if !isOfficer || tier != 2 {
		t.Errorf("(isOfficer, tier) = (%v, %d), want (true, 2)", isOfficer, tier)
	}

	if term.Rank != "M2 Third Officer" {
		t.Errorf("Rank = %q, want %q", term.Rank, "M2 Third Officer")
	}
}

func TestResolveMerchantTermCommissionFromEnlisted(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))

	term, _, isOfficer, tier := ResolveMerchantTerm(r, upp88, C1, false, 1, nil)

	if !term.Commissioned {
		t.Fatal("Commissioned = false, want true")
	}

	if !isOfficer || tier != 1 {
		t.Errorf("(isOfficer, tier) = (%v, %d), want (true, 1)", isOfficer, tier)
	}

	if term.Rank != "M1 Fourth Officer" {
		t.Errorf("Rank = %q, want %q", term.Rank, "M1 Fourth Officer")
	}
}

func TestResolveMerchantTermRatingPromotion(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(4, 4))

	term, _, isOfficer, tier := ResolveMerchantTerm(r, upp88, C1, false, 1, nil)

	if term.Commissioned {
		t.Fatal("Commissioned = true, want false (this fixture's own Commission roll fails)")
	}

	if !term.Promoted {
		t.Fatal("Promoted = false, want true")
	}

	if isOfficer || tier != 2 {
		t.Errorf("(isOfficer, tier) = (%v, %d), want (false, 2)", isOfficer, tier)
	}

	if term.Rank != "R1 Steward Apprentice" {
		t.Errorf("Rank = %q, want %q", term.Rank, "R1 Steward Apprentice")
	}
}

// TestResolveMerchantTermEscalatingShipSharesSecondReward confirms the
// "receipt number" actually escalates across terms, not just always
// granting 1 Ship Share.
func TestResolveMerchantTermEscalatingShipSharesSecondReward(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(2, 2))
	priorTerms := []Term{{RewardResult: "1 Ship Share"}}

	term, _, isOfficer, tier := ResolveMerchantTerm(r, upp88, C1, false, 1, priorTerms)

	if term.RewardResult != "2 Ship Shares" {
		t.Errorf("RewardResult = %q, want %q", term.RewardResult, "2 Ship Shares")
	}

	if !isOfficer || tier != 1 {
		t.Errorf(
			"(isOfficer, tier) = (%v, %d), want (true, 1) (this fixture's own Commission also succeeds)",
			isOfficer,
			tier,
		)
	}
}

// TestResolveMerchantTermRiskFailurePersistsReduction confirms Merchant
// reuses the exact universal Risk mechanic (a real physical
// characteristic reduction, unlike Entertainer's own Talent) — the
// returned UPP reflects the reduced C1.
func TestResolveMerchantTermRiskFailurePersistsReduction(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(23, 23))

	term, updatedUPP, _, _ := ResolveMerchantTerm(r, upp88, C1, false, 1, nil)

	if term.RiskResult != Wounded {
		t.Fatalf("RiskResult = %v, want Wounded", term.RiskResult)
	}

	if updatedUPP.Characteristics[C1] >= upp88.Characteristics[C1] {
		t.Errorf("updatedUPP.Characteristics[C1] = %v, want less than %v (a real reduction)",
			updatedUPP.Characteristics[C1], upp88.Characteristics[C1])
	}
}

// TestResolveMerchantTermRewardUsesTermStartCC is the regression for
// #50, mirroring TestResolveMarineTermRewardUsesTermStartCC. Unlike
// Marine/Soldier/Spacer, Merchant has no flat Risk-success grant to
// isolate Reward with, so this asserts on RewardResult directly. Seed
// 47 (found by direct search) reduces C1 from 8 to 6 during Risk, then
// rolls a 7 for Reward — a success against the term-start CC (7<=8)
// that would fail against the reduced value (7<=6 is false).
func TestResolveMerchantTermRewardUsesTermStartCC(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(47, 47))

	//nolint:dogsled // only RewardResult matters here; updatedUPP/isOfficer/tier are exercised by other tests
	term, _, _, _ := ResolveMerchantTerm(r, upp88, C1, false, 1, nil)

	if term.RiskResult != Wounded {
		t.Fatalf("RiskResult = %v, want Wounded (fixture assumption broke)", term.RiskResult)
	}

	if term.RewardResult != "1 Ship Share" {
		t.Errorf("RewardResult = %q, want %q (Reward must use the term-start CC)", term.RewardResult, "1 Ship Share")
	}
}
