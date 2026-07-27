package character

import (
	"math/rand/v2"
	"slices"
	"strings"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

func TestMusterOutCashAmount(t *testing.T) {
	t.Parallel()

	cases := []struct {
		entry      string
		wantAmount int
		wantOK     bool
	}{
		{"Cr30,000", 30000, true},
		{"Cr80,000", 80000, true},
		{"Low Passage", 0, false},
		{"StarPass", 0, false},
	}

	for _, c := range cases {
		amount, ok := musterOutCashAmount(c.entry)
		if amount != c.wantAmount || ok != c.wantOK {
			t.Errorf("musterOutCashAmount(%q) = (%d, %v), want (%d, %v)", c.entry, amount, ok, c.wantAmount, c.wantOK)
		}
	}
}

func TestMusterOutFameBonus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		entry      string
		wantAmount int
		wantOK     bool
	}{
		{"Fame +2", 2, true},
		{"Str +1", 0, false},
		{"Knighthood", 0, false},
	}

	for _, c := range cases {
		amount, ok := musterOutFameBonus(c.entry)
		if amount != c.wantAmount || ok != c.wantOK {
			t.Errorf("musterOutFameBonus(%q) = (%d, %v), want (%d, %v)", c.entry, amount, ok, c.wantAmount, c.wantOK)
		}
	}
}

func TestMusterOutCharacteristicBoost(t *testing.T) {
	t.Parallel()

	cases := []struct {
		entry      string
		wantPos    Position
		wantAmount int
		wantOK     bool
	}{
		{"Str +1", C1, 1, true},
		{"C5 +1", C5, 1, true},
		{"C2 +1", C2, 1, true},
		{"Fame +2", 0, 0, false},
		{"Ship Share", 0, 0, false},
	}

	for _, c := range cases {
		pos, amount, ok := musterOutCharacteristicBoost(c.entry)
		if pos != c.wantPos || amount != c.wantAmount || ok != c.wantOK {
			t.Errorf("musterOutCharacteristicBoost(%q) = (%v, %d, %v), want (%v, %d, %v)",
				c.entry, pos, amount, ok, c.wantPos, c.wantAmount, c.wantOK)
		}
	}
}

// TestMusterOutTableEntriesParseAsExpected is a full-pin regression test:
// every literal string in every real Mustering Out table (Scout and
// Citizen, Money and Benefits) must classify exactly as expected. Catches
// a table literal being edited (a new abbreviation, a typo) without the
// parser being updated to match.
func TestMusterOutTableEntriesParseAsExpected(t *testing.T) {
	t.Parallel()

	for _, entry := range scoutMusterOutMoney {
		assertMoneyEntryClassification(t, entry)
	}

	for _, entry := range citizenMusterOutMoney {
		assertMoneyEntryClassification(t, entry)
	}

	for _, entry := range scoutMusterOutBenefits {
		assertBenefitEntryClassification(t, entry)
	}

	for _, entry := range citizenMusterOutBenefits {
		assertBenefitEntryClassification(t, entry)
	}
}

func assertMoneyEntryClassification(t *testing.T, entry string) {
	t.Helper()

	_, isCash := musterOutCashAmount(entry)
	wantCash := strings.HasPrefix(entry, "Cr")

	if isCash != wantCash {
		t.Errorf("musterOutCashAmount(%q) ok = %v, want %v", entry, isCash, wantCash)
	}
}

// narrativeBenefitEntries are every Benefits-column literal, across both
// tables, that carries no mechanical effect — everything else in either
// table is either "Fame +N" or a characteristic-boost token.
var narrativeBenefitEntries = map[string]bool{
	"Ship Share": true, "Forbidden Knowledge": true, "Wafer Jack": true,
	"Life Insurance": true, "TAS Fellow Membership": true, "Knighthood": true,
}

func assertBenefitEntryClassification(t *testing.T, entry string) {
	t.Helper()

	_, isFame := musterOutFameBonus(entry)
	_, _, isBoost := musterOutCharacteristicBoost(entry)

	if isFame && isBoost {
		t.Errorf("%q classified as both a Fame bonus and a characteristic boost", entry)
	}

	wantParsed := !narrativeBenefitEntries[entry]
	if got := isFame || isBoost; got != wantParsed {
		t.Errorf("%q: parsed = %v, want %v", entry, got, wantParsed)
	}
}

func TestApplyMusteringOut(t *testing.T) {
	t.Parallel()

	career := Career{Name: "Scout", MusteringOut: MusteringOut{
		Money:    []string{"Cr30,000", "Cr20,000", "Low Passage"},
		Benefits: []string{"Fame +2", "Str +1", "Ship Share"},
	}}
	upp := UPP{Characteristics: [6]ehex.Value{5, 5, 5, 5, 5, 5}}

	gotUPP, gotBonuses := ApplyMusteringOut(dice.New(rand.NewPCG(1, 1)), career, upp)

	if gotBonuses.Cash != 50000 {
		t.Errorf("cash = %d, want 50000", gotBonuses.Cash)
	}

	if gotBonuses.Fame != 2 {
		t.Errorf("fame = %d, want 2", gotBonuses.Fame)
	}

	want := upp
	want.Characteristics[C1] = 6 // Str +1

	if gotUPP != want {
		t.Errorf("UPP = %+v, want %+v (only Str boosted, everything else unchanged)", gotUPP, want)
	}
}

// TestApplyMusteringOutRespectsTheHumanCharacteristicCap is #57's
// regression. Book 1 p.70 caps a Human characteristic at 15 and loses
// any award that would take it past — this used to clamp at ehex.Max
// (33), which is how many values a single extended-hex digit can encode,
// not a limit on people. Repeated awards carried real generated
// characters to 17.
func TestApplyMusteringOutRespectsTheHumanCharacteristicCap(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		start   ehex.Value
		want    ehex.Value
		explain string
	}{
		{"below the cap, award applies", 9, 10, ""},
		{"one below the cap, award reaches it", 14, HumanCharacteristicMax, ""},
		{"at the cap, award is lost", HumanCharacteristicMax, HumanCharacteristicMax, ""},
		{
			"above the cap, award is lost without dragging the value down",
			20, 20,
			"p.70 loses the award; it does not clamp an already-higher value to 15",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got, _ := ApplyMusteringOut(
				dice.New(rand.NewPCG(1, 1)),
				Career{Name: "Scout", MusteringOut: MusteringOut{Benefits: []string{"Str +1"}}},
				UPP{Characteristics: [6]ehex.Value{c.start, 0, 0, 0, 0, 0}},
			)

			if got.Characteristics[C1] != c.want {
				t.Errorf("Str %v + 1 = %v, want %v. %s", c.start, got.Characteristics[C1], c.want, c.explain)
			}
		})
	}
}

// TestApplyKnighthood pins Book 1 p.68's Knighthood benefit:
//
//	"A Knighthood raises any value of Soc to B; if the character is
//	already Soc 11+, he receives Soc +1 instead."
//	"In the Spacer, Soldier, and Marine careers, Knighthood is only
//	available to Officers. A non-officer receives Soc +1 (even if it
//	advances Soc to 11 or beyond)."
//
// It was recorded as a narrative string and never applied.
func TestApplyKnighthood(t *testing.T) {
	t.Parallel()

	knighted := func(name, rank string, soc ehex.Value) ehex.Value {
		career := Career{
			Name:         name,
			Terms:        []Term{{Rank: rank}},
			MusteringOut: MusteringOut{Benefits: []string{"Knighthood"}},
		}

		got, _ := ApplyMusteringOut(
			dice.New(rand.NewPCG(1, 1)), career, UPP{Characteristics: [6]ehex.Value{0, 0, 0, 0, 0, soc}})

		return got.Characteristics[C6]
	}

	t.Run("raises any lower Soc to B", func(t *testing.T) {
		t.Parallel()

		for _, soc := range []ehex.Value{0, 4, 7, 10} {
			if got := knighted("Scout", "", soc); got != knighthoodSocFloor {
				t.Errorf("Soc %v knighted = %v, want %v (B)", soc, got, ehex.Value(knighthoodSocFloor))
			}
		}
	})

	t.Run("already B or above takes Soc +1 instead", func(t *testing.T) {
		t.Parallel()

		for _, soc := range []ehex.Value{11, 12, 14} {
			if got, want := knighted("Scout", "", soc), soc+1; got != want {
				t.Errorf("Soc %v knighted = %v, want %v", soc, got, want)
			}
		}
	})

	t.Run("Spacer, Soldier and Marine restrict it to Officers", func(t *testing.T) {
		t.Parallel()

		for _, career := range []string{SpacerCareerName, SoldierCareerName, MarineCareerName} {
			// Enlisted ranks: R/S/M prefixes. Soc +1, not a raise to B.
			if got := knighted(career, "M3 Sergeant", 5); got != 6 {
				t.Errorf("%s enlisted, Soc 5 knighted = %v, want 6 (Soc +1, not B)", career, got)
			}

			// And +1 even where that carries them past B.
			if got := knighted(career, "R4 Petty Officer First", 10); got != 11 {
				t.Errorf("%s enlisted, Soc 10 knighted = %v, want 11", career, got)
			}

			if got := knighted(career, "O4 Force Commander", 5); got != knighthoodSocFloor {
				t.Errorf("%s officer, Soc 5 knighted = %v, want %v (B)",
					career, got, ehex.Value(knighthoodSocFloor))
			}
		}
	})

	t.Run("other careers are unrestricted regardless of rank", func(t *testing.T) {
		t.Parallel()

		// Merchant formats Officer ranks with an M prefix and Enlisted
		// with R — the opposite of Marine. It isn't one of the three
		// restricted careers, so neither prefix may gate the benefit.
		if got := knighted(MerchantCareerName, "R0 Spacehand", 5); got != knighthoodSocFloor {
			t.Errorf("Merchant enlisted, Soc 5 knighted = %v, want %v (B)", got, ehex.Value(knighthoodSocFloor))
		}
	})

	t.Run("respects the Human characteristic cap", func(t *testing.T) {
		t.Parallel()

		if got := knighted("Scout", "", HumanCharacteristicMax); got != HumanCharacteristicMax {
			t.Errorf("Soc at the cap knighted = %v, want it unchanged", got)
		}
	})
}

// TestApplyForbiddenKnowledge pins Book 1 p.69: "Forbidden Knowledge
// refers to a skill or knowledge that is not, should not, or cannot, be
// mentioned in polite society… Each receipt provides skill-1."
//
// The book names no table, only the worked example this grants.
func TestApplyForbiddenKnowledge(t *testing.T) {
	t.Parallel()

	career := Career{Name: "Scout", MusteringOut: MusteringOut{
		Benefits: []string{"Forbidden Knowledge", "Forbidden Knowledge"},
	}}

	_, bonuses := ApplyMusteringOut(dice.New(rand.NewPCG(1, 1)), career, UPP{})

	want := []SkillLevel{
		{Name: forbiddenKnowledgeSkill, Level: 1, Kind: Skill},
		{Name: forbiddenKnowledgeSkill, Level: 1, Kind: Skill},
	}
	if !slices.Equal(bonuses.Skills, want) {
		t.Errorf("Skills = %+v, want %+v (each receipt provides skill-1)", bonuses.Skills, want)
	}
}
