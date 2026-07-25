package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

func TestContinueNobleOutcomeBoundaries(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		roll int
		want bool
	}{
		{"natural 2 always succeeds, even against a target it would otherwise beat easily", 2, true},
		{"exactly at the target succeeds", 7, true},
		{"one above the target fails", 8, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := continueNobleOutcome(c.roll); got != c.want {
				t.Errorf("continueNobleOutcome(%d) = %v, want %v", c.roll, got, c.want)
			}
		})
	}
}

func TestResolveNobleCareerNeverQualifiedReturnsZeroTermsCareer(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{0, 0, 0, 0, 0, 10}} // Soc 10 < B
	r := dice.New(rand.NewPCG(1, 1))

	career := ResolveNobleCareer(r, upp)

	if career.Name != NobleCareerName {
		t.Errorf("career.Name = %q, want %q", career.Name, NobleCareerName)
	}

	if !career.HasRank {
		t.Error("career.HasRank = false, want true (Noble is not in Book 1's no-rank career list)")
	}

	if len(career.Terms) != 0 {
		t.Errorf("career.Terms = %v, want empty (BeginNoble fails below Soc B)", career.Terms)
	}
}

// TestResolveNobleCareerRespectsMaxTermsCap is statistical, mirroring
// TestResolveCitizenCareerRespectsMaxTermsCap's own reasoning: Continue's
// fixed target (7) is UPP-independent, so there's no deterministic way
// to force every trial to reach the cap. Per-term retention is P(2D<=7)
// = 21/36 (the natural-2 override changes nothing here, since roll==2 is
// already <=7), so P(all 14 terms) = (21/36)^14 ≈ 0.05% — a large trial
// count is needed to reliably observe at least one full-length career.
func TestResolveNobleCareerRespectsMaxTermsCap(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{0, 12, 12, 12, 12, 12}}
	r := dice.New(rand.NewPCG(13, 17))

	const trials = 20000

	atCap := 0

	for range trials {
		career := ResolveNobleCareer(r, upp)
		if len(career.Terms) > maxCareerTerms {
			t.Fatalf("len(career.Terms) = %d, want <= %d", len(career.Terms), maxCareerTerms)
		}

		if len(career.Terms) == maxCareerTerms {
			atCap++
		}
	}

	if atCap == 0 {
		t.Errorf("0 of %d trials reached maxCareerTerms, want at least 1 (cap never observed as reachable)", trials)
	}
}

// TestResolveNobleCareerCCRotation reuses an immortal-at-Intrigue fixture
// (C2=C3=C4=C5=12: Intrigue's own target only ever grows more favorable
// as successful Intrigues accumulate — see returnIntrigueMod's own doc
// comment — so every term succeeds and Exile never triggers) to confirm
// nobleReturnIntriguePositions actually rotates via nextCC, checked
// against however many terms the (probabilistic) Continue check actually
// produced rather than requiring the full cap.
func TestResolveNobleCareerCCRotation(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{0, 12, 12, 12, 12, 12}}
	r := dice.New(rand.NewPCG(7, 7))

	career := ResolveNobleCareer(r, upp)

	if len(career.Terms) == 0 {
		t.Fatal("career.Terms is empty, want at least one term")
	}

	wantCycle := []Position{C2, C3, C4, C5}

	for i, term := range career.Terms {
		want := wantCycle[i%len(wantCycle)]
		if term.ControllingCharacteristic != want {
			t.Errorf("term %d: ControllingCharacteristic = %v, want %v", i+1, term.ControllingCharacteristic, want)
		}
	}
}

// TestResolveNobleCareerExileTransitions is an end-to-end confirmation
// that ResolveNobleCareer actually threads career.Terms into each new
// ResolveNobleTerm call: for every term in several real runs, the
// recorded NobleAction must match what nobleIsExiled derives from the
// terms preceding it — not just that the pure helpers are individually
// correct in isolation (already covered by TestNobleIsExiled and
// friends), but that the loop actually wires them together.
func TestResolveNobleCareerExileTransitions(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{0, 6, 6, 6, 6, 12}}

	for _, seed := range []uint64{1, 2, 3, 4, 5} {
		r := dice.New(rand.NewPCG(seed, seed))

		career := ResolveNobleCareer(r, upp)

		for i, term := range career.Terms {
			wantAction := "Intrigue"
			if nobleIsExiled(career.Terms[:i]) {
				wantAction = "Return"
			}

			if term.NobleAction != wantAction {
				t.Errorf("seed %d, term %d: NobleAction = %q, want %q (derived from prior terms)",
					seed, i+1, term.NobleAction, wantAction)
			}
		}
	}
}
