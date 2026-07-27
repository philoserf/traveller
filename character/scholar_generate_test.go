package character

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

func TestScholarRankNamesMatchBook1P76(t *testing.T) {
	t.Parallel()

	want := [7]string{
		"Amateur", "Lecturer", "Instructor", "Assistant Professor",
		"Associate Professor", "Professor", "Distinguished Professor",
	}

	if scholarRankNames != want {
		t.Errorf("scholarRankNames = %v, want %v", scholarRankNames, want)
	}
}

func TestScholarSkillTableMatchesBook1P76(t *testing.T) {
	t.Parallel()

	want := [7][6]string{
		{"Str", "C2", "C3", "Int", "C5", "C6"},
		{"Major", "Major", "Minor", "Minor", "One Trade", "One Trade"},
		{"Seafarer", "Navigation", "Hostile Environment", "Flyer", "Driver", "Vacc Suit"},
		{"Survey", "Astrogation", "Hostile Environment", "Survival", "Animals", "Bureaucrat"},
		{"Fighter", "Fighter", "Stealth", "Flyer", "Gunner", "Sensors"},
		{"Admin", "Language", "One Science", "Comms", "Starship Skill", "Bureaucrat"},
		{"Seafarer", "One Art", "One Science", "Athlete", "Medic", "One Trade"},
	}

	if scholarSkillTable != want {
		t.Errorf("scholarSkillTable = %v, want %v", scholarSkillTable, want)
	}
}

// TestBeginScholar covers both of Book 1 p.76's own Begin branches:
// Edu>=8 is automatic (no roll at all, dice-free by construction), Edu<8
// rolls against Edu with no Mod — seeds 3/1 were confirmed by direct
// inspection to produce a success/failure respectively at Edu=6.
func TestBeginScholar(t *testing.T) {
	t.Parallel()

	t.Run("Edu 8+ is automatic", func(t *testing.T) {
		t.Parallel()

		r := dice.New(rand.NewPCG(999, 999))

		ok, tier := BeginScholar(r, 8)
		if !ok {
			t.Error("BeginScholar(edu=8) ok = false, want true (automatic)")
		}

		if tier != 1 {
			t.Errorf("BeginScholar(edu=8) tier = %d, want 1 (Lecturer)", tier)
		}
	})

	t.Run("Edu<8 roll success enters at tier 0", func(t *testing.T) {
		t.Parallel()

		r := dice.New(rand.NewPCG(3, 3))

		ok, tier := BeginScholar(r, 6)
		if !ok {
			t.Fatal("BeginScholar(edu=6) ok = false, want true (seed 3 rolls a success)")
		}

		if tier != 0 {
			t.Errorf("BeginScholar(edu=6) tier = %d, want 0 (Amateur)", tier)
		}
	})

	t.Run("Edu<8 roll failure never qualifies", func(t *testing.T) {
		t.Parallel()

		r := dice.New(rand.NewPCG(1, 1))

		ok, _ := BeginScholar(r, 6)
		if ok {
			t.Error("BeginScholar(edu=6) ok = true, want false (seed 1 rolls a failure)")
		}
	})
}

func TestScholarPublicationsTotal(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		terms []Term
		want  int
	}{
		{"no terms", nil, 0},
		{"one ordinary success", []Term{{PublicationSucceeded: true}}, 1},
		{"one Award-Winning counts as two", []Term{{PublicationSucceeded: true, AwardWinning: true}}, 2},
		{"a failed Publication contributes nothing", []Term{{PublicationSucceeded: false}}, 0},
		{
			"mixed",
			[]Term{
				{PublicationSucceeded: true},
				{PublicationSucceeded: true, AwardWinning: true},
				{PublicationSucceeded: false},
			},
			3,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := scholarPublicationsTotal(c.terms); got != c.want {
				t.Errorf("scholarPublicationsTotal(%v) = %d, want %d", c.terms, got, c.want)
			}
		})
	}
}

func TestHasTenure(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		terms []Term
		want  bool
	}{
		{"no terms", nil, false},
		{"tenure granted", []Term{{TenureGranted: true}}, true},
		{"tenure never granted", []Term{{TenureGranted: false}, {PublicationSucceeded: true}}, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := hasTenure(c.terms); got != c.want {
				t.Errorf("hasTenure(%v) = %v, want %v", c.terms, got, c.want)
			}
		})
	}
}

func TestScholarRankTier(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		terms     []Term
		startTier int
		want      int
	}{
		{"no terms, Amateur start", nil, 0, 0},
		{"no terms, Lecturer start", nil, 1, 1},
		{"one promotion from Lecturer", []Term{{Promoted: true}}, 1, 2},
		{"non-promoted terms don't advance", []Term{{}, {}}, 1, 1},
		{
			"capped at the top of the table",
			[]Term{
				{Promoted: true},
				{Promoted: true},
				{Promoted: true},
				{Promoted: true},
				{Promoted: true},
				{Promoted: true},
				{Promoted: true},
			},
			0,
			6,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := scholarRankTier(c.terms, c.startTier); got != c.want {
				t.Errorf("scholarRankTier(%v, %d) = %d, want %d", c.terms, c.startTier, got, c.want)
			}
		})
	}
}

// The ResolveScholarTerm fixtures below were seed-hunted against an
// all-8 UPP (upp84) and verified by direct inspection of the resulting
// Term before being pinned here — not assumed from the mechanic alone.
var upp84 = UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 8}}

// upp84NoWaiver is upp84 with Soc 0, which is how these fixtures keep
// testing the roll they are about.
//
// Book 1 p.76 lets a Scholar waive an adverse result "in Position,
// Promotion, Research, Publication, Tenure, or Continue" on a Check Soc,
// so at Soc 8 a fixture built to fail one of those rolls is liable to be
// rescued and stop exercising it. A 2D check cannot come in at or below
// 0, so Soc 0 disables Waivers by construction and leaves the underlying
// mechanic exposed. The Waiver behaviour itself is tested separately, in
// scholar_waiver_test.go.
var upp84NoWaiver = UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 0}}

func TestResolveScholarTermResearchUnharmedPublicationSuccess(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(2, 2))

	term, _, tier := ResolveScholarTerm(r, upp84, C1, 8, 5, nil, "", new(int))

	if term.RiskResult != Unharmed {
		t.Fatalf("RiskResult = %v, want Unharmed", term.RiskResult)
	}

	if !term.PublicationSucceeded {
		t.Error("PublicationSucceeded = false, want true")
	}

	if term.AwardWinning {
		t.Error("AwardWinning = true, want false")
	}

	if tier != 5 {
		t.Errorf("tier = %d, want 5 (unchanged, no Promotion this term)", tier)
	}
}

func TestResolveScholarTermPublicationAwardWinning(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(5, 5))

	term, _, _ := ResolveScholarTerm(r, upp84, C1, 8, 5, nil, "", new(int))

	if term.RiskResult != Unharmed {
		t.Fatalf("RiskResult = %v, want Unharmed", term.RiskResult)
	}

	if !term.PublicationSucceeded || !term.AwardWinning {
		t.Errorf(
			"PublicationSucceeded = %v, AwardWinning = %v, want true, true",
			term.PublicationSucceeded,
			term.AwardWinning,
		)
	}
}

func TestResolveScholarTermResearchUnharmedPublicationFailure(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))

	term, _, _ := ResolveScholarTerm(r, upp84NoWaiver, C1, 8, 5, nil, "", new(int))

	if term.RiskResult != Unharmed {
		t.Fatalf("RiskResult = %v, want Unharmed", term.RiskResult)
	}

	if term.PublicationSucceeded {
		t.Error("PublicationSucceeded = true, want false")
	}
}

// TestResolveScholarTermWoundedSkipsPublication confirms Publication is
// only rolled when Research succeeded (RiskResult == Unharmed) — a real
// divergence from Scout's/Marine's own "Reward rolled unless Dead"
// pattern, per this slice's own plan-file Context.
func TestResolveScholarTermWoundedSkipsPublication(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(23, 23))

	term, _, _ := ResolveScholarTerm(r, upp84NoWaiver, C1, 8, 5, nil, "", new(int))

	if term.RiskResult != Wounded {
		t.Fatalf("RiskResult = %v, want Wounded", term.RiskResult)
	}

	if term.PublicationSucceeded || term.AwardWinning {
		t.Errorf("PublicationSucceeded = %v, AwardWinning = %v, want false, false (no Reward on a non-Unharmed term)",
			term.PublicationSucceeded, term.AwardWinning)
	}
}

func TestResolveScholarTermDisabledSkipsPublication(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(31, 31))

	term, _, _ := ResolveScholarTerm(r, upp84NoWaiver, C1, 8, 5, nil, "", new(int))

	if term.RiskResult != Disabled {
		t.Fatalf("RiskResult = %v, want Disabled", term.RiskResult)
	}

	if term.PublicationSucceeded {
		t.Error("PublicationSucceeded = true, want false")
	}
}

func TestResolveScholarTermPromotionBelowTier3(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))

	term, _, tier := ResolveScholarTerm(r, upp84, C1, 10, 1, nil, "", new(int))

	if !term.Promoted {
		t.Fatal("Promoted = false, want true")
	}

	if tier != 2 {
		t.Errorf("tier = %d, want 2 (Instructor)", tier)
	}

	if term.Rank != "Instructor" {
		t.Errorf("Rank = %q, want %q", term.Rank, "Instructor")
	}
}

// TestResolveScholarTermTenureGrantedAtTier3 confirms a Tenure roll is
// attempted (and can succeed) at tier 3 with Edu>=10 — the roll's own
// success/failure is independent of whether a same-term Promotion roll
// also then succeeds (Book 1's own "Promotion beyond Scholar3 not
// possible without Tenure" only removes the block; it doesn't itself
// grant the Promotion).
func TestResolveScholarTermTenureGrantedAtTier3(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(5, 5))

	term, _, tier := ResolveScholarTerm(r, upp84NoWaiver, C1, 10, 3, nil, "", new(int))

	if !term.TenureGranted {
		t.Fatal("TenureGranted = false, want true")
	}

	if tier != 3 {
		t.Errorf("tier = %d, want 3 (this fixture's own Promotion roll fails even though Tenure unblocked it)", tier)
	}
}

// TestResolveScholarTermPromotionBlockedWithoutTenureAtTier3 confirms
// Promotion is never even attempted at tier 3 without Tenure — tier
// stays 3 regardless of what the Promotion roll would have done.
func TestResolveScholarTermPromotionBlockedWithoutTenureAtTier3(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))

	term, _, tier := ResolveScholarTerm(r, upp84NoWaiver, C1, 10, 3, nil, "", new(int))

	if term.TenureGranted {
		t.Fatal("TenureGranted = true, want false (this fixture's own Tenure roll fails)")
	}

	if term.Promoted {
		t.Error("Promoted = true, want false (Promotion must not be attempted at tier 3 without Tenure)")
	}

	if tier != 3 {
		t.Errorf("tier = %d, want 3 (unchanged)", tier)
	}
}

// TestResolveScholarTermPromotionUnblockedByPriorTenure confirms
// hasTenure(priorTerms) — Tenure granted in an earlier term — is
// sufficient to unblock Promotion past tier 3, not just a same-term grant.
func TestResolveScholarTermPromotionUnblockedByPriorTenure(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))
	priorTerms := []Term{{TenureGranted: true}}

	term, _, tier := ResolveScholarTerm(r, upp84, C1, 10, 3, priorTerms, "", new(int))

	if !term.Promoted {
		t.Fatal("Promoted = false, want true")
	}

	if tier != 4 {
		t.Errorf("tier = %d, want 4 (Associate Professor)", tier)
	}
}

// TestResolveScholarTermPromotionContinuesPastTier4WithTenure is the
// regression test for a code-review-caught bug: an earlier version of
// the Promotion gate checked `tier == 3` specifically instead of
// `tier >= 3`, so a Tenured Scholar already promoted to tier 4
// (Associate Professor) could never be promoted again — Book 1's own
// "Promotion beyond Scholar3 not possible without Tenure" gates every
// transition past tier 3, not just the first one past it.
func TestResolveScholarTermPromotionContinuesPastTier4WithTenure(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))
	priorTerms := []Term{{TenureGranted: true}}

	term, _, tier := ResolveScholarTerm(r, upp84, C1, 12, 4, priorTerms, "", new(int))

	if !term.Promoted {
		t.Fatal("Promoted = false, want true (Tenure won in an earlier term must still unblock tier 4->5)")
	}

	if tier != 5 {
		t.Errorf("tier = %d, want 5 (Professor)", tier)
	}
}
