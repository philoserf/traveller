package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

func TestMerchantCareerFame(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		tier int
		want int
	}{
		{"RX Temp", 0, 0},
		{"R2 Drive Helper", 3, 3},
		{"M1 Fourth Officer", 1, 1},
		{"M6 Senior Captain", 6, 6},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := merchantCareerFame(c.tier); got != c.want {
				t.Errorf("merchantCareerFame(%d) = %d, want %d", c.tier, got, c.want)
			}
		})
	}
}

// TestBuildMerchantCharacterQualified pins seed 1 against an all-8 UPP:
// one term, Risk fails (Disabled — confirmed by direct inspection),
// Rank "M1 Fourth Officer" (Begin as Officer succeeded), Fame 1 (tier 1)
// — confirmed by direct inspection before being pinned, not assumed
// from the formula alone.
func TestBuildMerchantCharacterQualified(t *testing.T) {
	t.Parallel()

	homeworldSkills := []SkillLevel{{Name: "Vacc Suit", Level: 1, Kind: Skill}}
	r := dice.New(rand.NewPCG(1, 1))

	c, ok := buildMerchantCharacter(r, upp88, "hw", homeworldSkills)

	if !ok {
		t.Fatal("ok = false, want true")
	}

	if len(c.Careers) != 1 || c.Careers[0].Name != MerchantCareerName {
		t.Fatalf("Careers = %+v, want one Career named %q", c.Careers, MerchantCareerName)
	}

	if !c.Careers[0].HasRank {
		t.Error("Careers[0].HasRank = false, want true")
	}

	if len(c.Careers[0].Terms) == 0 {
		t.Fatal("Careers[0].Terms is empty, want at least one term (BeginMerchant never fails)")
	}

	if c.Rank != "M1 Fourth Officer" {
		t.Errorf("Rank = %q, want %q", c.Rank, "M1 Fourth Officer")
	}

	if c.Fame != 1 {
		t.Errorf("Fame = %d, want 1", c.Fame)
	}

	if len(c.Skills) < len(homeworldSkills) {
		t.Errorf("len(Skills) = %d, want at least %d (homeworld skills alone)", len(c.Skills), len(homeworldSkills))
	}

	assertBirthdateFormat(t, c.Birthdate, c.Age)
}

// TestBuildMerchantCharacterNeverProducesAnEmptyCareer confirms the
// central finding of this slice across a spread of seeds/UPPs — unlike
// every other career this session, buildMerchantCharacter's own
// Careers[0].Terms is never empty (BeginMerchant cannot fail).
func TestBuildMerchantCharacterNeverProducesAnEmptyCareer(t *testing.T) {
	t.Parallel()

	upps := []UPP{
		{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}},
		{Characteristics: [6]ehex.Value{1, 1, 1, 1, 1, 1}},
		{Characteristics: [6]ehex.Value{0, 0, 0, 0, 0, 0}},
	}

	for _, upp := range upps {
		for seed := uint64(1); seed <= 10; seed++ {
			r := dice.New(rand.NewPCG(seed, seed))

			c, _ := buildMerchantCharacter(r, upp, "hw", nil)
			if len(c.Careers[0].Terms) == 0 {
				t.Fatalf("seed %d, upp %v: Careers[0].Terms is empty, want at least one term", seed, upp)
			}
		}
	}
}

// TestGenerateMerchantCharacterProducesAHumanCharacter is a smoke test
// confirming the full public entry point wires GenerateUPP/
// GenerateHomeworldSkills into buildMerchantCharacter, mirroring every
// other career's own GenerateXCharacter smoke test.
func TestGenerateMerchantCharacterProducesAHumanCharacter(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))

	c, _ := GenerateMerchantCharacter(r)

	if c.Species != "Human" {
		t.Errorf("Species = %q, want %q", c.Species, "Human")
	}

	if len(c.Careers) != 1 || c.Careers[0].Name != MerchantCareerName {
		t.Errorf("Careers = %+v, want one Career named %q", c.Careers, MerchantCareerName)
	}
}
