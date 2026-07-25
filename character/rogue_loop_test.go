package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// TestContinueRogueBoundaries mirrors TestContinueNobleOutcomeBoundaries'
// own shape, but continueRogue additionally shares rogueSucceeds' own
// natural-12-always-fails exception (Book 1's "Continue CC*" shares the
// same "*...But, 12 is always automatic failure" footnote), so a
// natural-12 case is added alongside the natural-2 override and the
// normal comparison. Seeds 21/11/1 were confirmed by direct inspection
// to produce first-TwoD6() rolls of 2/12/7 respectively.
func TestContinueRogueBoundaries(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		seed    uint64
		cc, mod int
		want    bool
	}{
		{"natural 2 always succeeds, even against target 0", 21, 0, 0, true},
		{"natural 12 always fails, even against a trivially easy target", 11, 12, 12, false},
		{"normal roll (7) exactly at target succeeds", 1, 7, 0, true},
		{"normal roll (7) one above target fails", 1, 6, 0, false},
		{"normal roll (7) pushed to success by mod", 1, 6, 1, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			r := dice.New(rand.NewPCG(c.seed, c.seed))
			if got := continueRogue(r, c.cc, c.mod); got != c.want {
				t.Errorf("continueRogue(seed=%d, cc=%d, mod=%d) = %v, want %v", c.seed, c.cc, c.mod, got, c.want)
			}
		})
	}
}

// TestResolveRogueCareerNeverQualifiedReturnsZeroTermsCareer mirrors
// TestResolveNobleCareerNeverQualifiedReturnsZeroTermsCareer's own
// shape: a zero UPP guarantees BeginRogue fails regardless of which CC
// rollRogueCC happens to pick.
func TestResolveRogueCareerNeverQualifiedReturnsZeroTermsCareer(t *testing.T) {
	t.Parallel()

	upp := UPP{}
	r := dice.New(rand.NewPCG(1, 1))

	career := ResolveRogueCareer(r, upp)

	if career.Name != RogueCareerName {
		t.Errorf("career.Name = %q, want %q", career.Name, RogueCareerName)
	}

	if career.HasRank {
		t.Error("career.HasRank = true, want false (Rogue is in Book 1's no-rank career list)")
	}

	if len(career.Terms) != 0 {
		t.Errorf("career.Terms = %v, want empty (BeginRogue fails against a zero UPP)", career.Terms)
	}
}

// TestResolveRogueCareerFixedCCNeverRotates confirms every term of a
// multi-term career records the same ControllingCharacteristic — the
// defining difference from every other career's own nextCC rotation
// (see this slice's own plan-file Context for why passing a
// single-element positions slice to resolveCareerLoop is sufficient).
// Seed 2 against an all-12 UPP was confirmed by direct inspection to
// reach the full 14-term cap.
func TestResolveRogueCareerFixedCCNeverRotates(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{12, 12, 12, 12, 12, 12}}
	r := dice.New(rand.NewPCG(2, 2))

	career := ResolveRogueCareer(r, upp)

	if len(career.Terms) == 0 {
		t.Fatal("career.Terms is empty, want at least one term")
	}

	want := career.Terms[0].ControllingCharacteristic

	for i, term := range career.Terms {
		if term.ControllingCharacteristic != want {
			t.Errorf("term %d: ControllingCharacteristic = %v, want %v (fixed for the whole career)",
				i+1, term.ControllingCharacteristic, want)
		}
	}
}

// TestResolveRogueCareerRespectsMaxTermsCap mirrors
// TestResolveNobleCareerRespectsMaxTermsCap's own reasoning: confirms
// the shared resolveCareerLoop cap actually applies to Rogue and that
// reaching it is observable, using the same seed already confirmed by
// direct inspection to reach 14 terms rather than a large statistical
// trial count (Rogue's own per-term retention against an all-12 UPP is
// very high — only a natural 12 fails Continue — so a full-length
// career is common, not a rare tail event the way Citizen's/Noble's own
// UPP-independent Continue targets are).
func TestResolveRogueCareerRespectsMaxTermsCap(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{12, 12, 12, 12, 12, 12}}
	r := dice.New(rand.NewPCG(2, 2))

	career := ResolveRogueCareer(r, upp)

	if len(career.Terms) != maxCareerTerms {
		t.Errorf("len(career.Terms) = %d, want %d (maxCareerTerms, seed 2 confirmed to reach the cap)",
			len(career.Terms), maxCareerTerms)
	}
}
