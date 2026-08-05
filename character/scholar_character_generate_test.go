package character

import (
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// TestBuildScholarCharacterNeverQualified mirrors
// TestBuildRogueCharacterNeverQualified's own shape: Edu 0 guarantees
// BeginScholar's own roll fails.
func TestBuildScholarCharacterNeverQualified(t *testing.T) {
	t.Parallel()

	// Soc 0 as well as Edu 0: p.76 lets a Scholar waive an adverse
	// "Position" result on a Check Soc, so a failed Begin at Soc 8
	// would simply be waived and the character would qualify after
	// all. A 2D check cannot come in at or below 0.
	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 0, 0}}
	r := dice.New(rand.NewPCG(1, 1))

	c, ok := buildScholarCharacter(r, upp, "hw", nil, Education{})

	if ok {
		t.Error("ok = true, want false (BeginScholar's own roll fails against Edu 0)")
	}

	if len(c.Careers) != 1 || len(c.Careers[0].Terms) != 0 {
		t.Errorf("Careers = %+v, want one Career with zero Terms", c.Careers)
	}

	if c.Age != 19 {
		t.Errorf("Age = %d, want 19 (18 plus the one year the failed Begin roll cost, Book 1 p.65)", c.Age)
	}

	if c.Fame != 0 {
		t.Errorf("Fame = %d, want 0 (never qualified)", c.Fame)
	}
}

// TestBuildScholarCharacterQualified pins seed 9 against Str/Dex/End/
// Int/Soc 8, Edu 9: one term, Research succeeds (Unharmed), Publication
// succeeds (not Award-Winning), Promoted true (tier 1 Lecturer -> 2
// Instructor). Fame = 4 (scholarCareerFame: tier 2 + pubs 1 = 3, plus a
// Mustering Out "Fame +1" Benefit) — confirmed by direct inspection
// before being pinned, not assumed from the formula alone.
func TestBuildScholarCharacterQualified(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 9, 8}}
	homeworldSkills := []SkillLevel{{Name: "Vacc Suit", Level: 1, Kind: Skill}}
	r := dice.New(rand.NewPCG(9, 9))

	c, ok := buildScholarCharacter(r, upp, "hw", homeworldSkills, Education{})

	if !ok {
		t.Fatal("ok = false, want true")
	}

	if len(c.Careers) != 1 || c.Careers[0].Name != ScholarCareerName {
		t.Fatalf("Careers = %+v, want one Career named %q", c.Careers, ScholarCareerName)
	}

	if !c.Careers[0].HasRank {
		t.Error("Careers[0].HasRank = false, want true")
	}

	// At least one term, not exactly one. This fixture pinned 1 back when
	// its Continue roll happened to fail immediately; p.76's Waivers now
	// rescue such a failure on a Check Soc, and #110 lets the career run
	// to its own natural end besides. The term inspected below is the
	// first either way.
	if len(c.Careers[0].Terms) == 0 {
		t.Fatal("Careers[0].Terms is empty, want at least one term")
	}

	term := c.Careers[0].Terms[0]
	if term.RiskResult != Unharmed {
		t.Errorf("Terms[0].RiskResult = %v, want Unharmed", term.RiskResult)
	}

	if !term.PublicationSucceeded || term.AwardWinning {
		t.Errorf("Terms[0].PublicationSucceeded = %v, AwardWinning = %v, want true, false",
			term.PublicationSucceeded, term.AwardWinning)
	}

	if !term.Promoted {
		t.Error("Terms[0].Promoted = false, want true")
	}

	// Derived, not pinned. p.76's rank titles carry the Scholar's Major
	// from Lecturer upward — "Assistant Professor of Psychology" — so the
	// title is the ladder entry the tier reached plus that suffix, and
	// which tier a seed reaches moves with the dice stream.
	career := c.Careers[0]
	if want := lastTermRank(career.Terms); c.Rank != want {
		t.Errorf("Rank = %q, want the career's own %q", c.Rank, want)
	}

	assertScholarRankNamesTheMajor(t, c.Rank, career)

	if want := resolveFameStacks(scholarSegmentFameAwards(c.UPP, career.Terms)); c.Fame < want {
		t.Errorf("Fame = %d, want at least %d (Rank plus Publications, before Mustering Out)", c.Fame, want)
	}

	if len(c.Skills) < len(homeworldSkills) {
		t.Errorf("len(Skills) = %d, want at least %d (homeworld skills alone)", len(c.Skills), len(homeworldSkills))
	}

	assertBirthdateFormat(t, c.Birthdate, c.Age)
}

// TestBuildScholarCharacterMusterOutUsesEntryTimeEdu is the regression
// test for #136: buildScholarCharacter used to pass careerUPP (the
// post-career UPP) to both ResolveScholarMusterOut and
// scholarSegmentFameAwards, when both need entry-time Edu to know
// Scholar's starting rank tier (scholarStartTier, gated at Edu 8) —
// ResolveScholarMusterOut's own doc comment says so explicitly.
// resolveScholarSegment (career_chain.go) had the identical bug at both
// call sites; fixed there too so a chain segment and a standalone career
// still produce the identical character from the identical seed.
//
// Seed 3 against Str/Dex/End/Int 8, Edu 7, Soc 8 (BeginScholar rolls
// against Edu 7, tier starts at 0/Amateur): a 4-term career with a
// Personal award raising Edu to 8 mid-career, but never a Promoted term
// — Rank stays Amateur (tier 0) throughout. Confirmed by direct
// comparison against the pre-fix implementation: it wrongly read the
// post-career Edu 8 as the starting tier (scholarStartTier(8) = 1),
// crediting a Rank-tier-1 Fame award and a +1 Mustering Out DM the
// character never earned — Fame 6 pre-fix vs. the correct 4 post-fix.
func TestBuildScholarCharacterMusterOutUsesEntryTimeEdu(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 7, 8}}
	r := dice.New(rand.NewPCG(3, 3))

	c, ok := buildScholarCharacter(r, upp, "hw", nil, Education{})

	if !ok {
		t.Fatal("ok = false, want true")
	}

	if c.Rank != "Amateur" {
		t.Fatalf(
			"Rank = %q, want %q (no Promoted term this seed, so tier must stay at its entry-time 0)",
			c.Rank,
			"Amateur",
		)
	}

	if c.Fame != 4 {
		t.Errorf("Fame = %d, want 4 (6 was the pre-fix value: a spurious Rank-tier-1 Fame award "+
			"plus a Mustering Out benefit reached only via the inflated DM the stale post-career Edu produced)",
			c.Fame)
	}
}

// TestBuildScholarCharacterDiedMidCareer pins seed 1 against a very low
// Str/Dex/End/Int (2) UPP: the character dies on term 2's own Risk roll
// (RiskResult == Dead), so ok is false despite two real Terms having
// been resolved — mirrors buildRiskCareerCharacter's own Dead-check
// convention, confirmed here since buildScholarCharacter reimplements
// it directly (bespoke, not delegating to buildRiskCareerCharacter).
func TestBuildScholarCharacterDiedMidCareer(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{2, 2, 2, 2, 10, 8}}
	r := dice.New(rand.NewPCG(1, 1))

	c, ok := buildScholarCharacter(r, upp, "hw", nil, Education{})

	if ok {
		t.Error("ok = true, want false (the character dies mid-career)")
	}

	if len(c.Careers[0].Terms) == 0 {
		t.Fatal("Careers[0].Terms is empty, want at least one term before death")
	}

	last := c.Careers[0].Terms[len(c.Careers[0].Terms)-1]
	if last.RiskResult != Dead {
		t.Errorf("last term's RiskResult = %v, want Dead", last.RiskResult)
	}

	if c.Fame != 0 {
		t.Errorf("Fame = %d, want 0 (a died character doesn't retain intrinsic career Fame)", c.Fame)
	}

	if c.Cash != 0 {
		t.Errorf("Cash = %d, want 0 (Dead zeroes Mustering Out, per musterOutRollCount)", c.Cash)
	}
}

// TestGenerateScholarCharacterProducesAHumanCharacter is a smoke test
// confirming the full public entry point wires GenerateUPP/
// GenerateHomeworldSkills into buildScholarCharacter, mirroring every
// other career's own GenerateXCharacter smoke test.
func TestGenerateScholarCharacterProducesAHumanCharacter(t *testing.T) {
	t.Parallel()

	r := dice.New(rand.NewPCG(1, 1))

	c, _ := GenerateScholarCharacter(r)

	if c.Species != "Human" {
		t.Errorf("Species = %q, want %q", c.Species, "Human")
	}

	if len(c.Careers) != 1 || c.Careers[0].Name != ScholarCareerName {
		t.Errorf("Careers = %+v, want one Career named %q", c.Careers, ScholarCareerName)
	}
}

// assertScholarRankNamesTheMajor checks p.76's own printed titles: every
// rank from Lecturer upward carries the Scholar's Major, and Amateur —
// the rank of a character with Edu 7 or less — carries none.
func assertScholarRankNamesTheMajor(t *testing.T, rank string, career Career) {
	t.Helper()

	if career.Major == "" {
		t.Error("Careers[0].Major is empty; p.76 gives every Scholar a Major")

		return
	}

	if rank == scholarRankNames[0] {
		return
	}

	if !strings.HasSuffix(rank, " of "+career.Major) {
		t.Errorf("Rank = %q, want it to name the Major %q (or be the suffix-less Amateur)", rank, career.Major)
	}
}
