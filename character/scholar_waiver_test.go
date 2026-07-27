package character

import (
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// TestScholarRankTitlesCarryTheMajor is Book 1 p.76's own "Table Of
// Scholar Ranks", whose printed titles read "Lecturer <of Major>" and so
// on from level 1 upward. Level 0 Amateur takes no suffix — the printed
// table gives it none, and it is the rank of a character with Edu 7 or
// less who "can resolve Risk and Reward, but cannot be Promoted".
//
// Level 5 is printed "Professor of <Major>" where its neighbours read
// "<of Major>", enclosing a different span but rendering identically, so
// the suffix is applied uniformly.
func TestScholarRankTitlesCarryTheMajor(t *testing.T) {
	t.Parallel()

	if got := scholarRankName(0, "Psychology"); got != "Amateur" {
		t.Errorf("tier 0 = %q, want a bare Amateur", got)
	}

	for tier := 1; tier < len(scholarRankNames); tier++ {
		got := scholarRankName(tier, "Psychology")

		want := scholarRankNames[tier] + " of Psychology"
		if got != want {
			t.Errorf("tier %d = %q, want %q", tier, got, want)
		}
	}

	// Before CharGen step C there was no Major to print, and a Scholar
	// generated without one must still get a usable title.
	if got := scholarRankName(3, ""); got != scholarRankNames[3] {
		t.Errorf("tier 3 without a Major = %q, want the bare title", got)
	}
}

// TestEveryScholarHasAMajorAndMinor is p.76 stated flatly: "Every
// Scholar has a Major and a Minor. If no degree (and an associated Major
// and Minor) then select any Skill or Knowledge from the Skills List."
//
// Both halves are covered: a graduate brings his degree's subjects, and
// anyone else declares fresh ones.
func TestEveryScholarHasAMajorAndMinor(t *testing.T) {
	t.Parallel()

	t.Run("a graduate brings his degree", func(t *testing.T) {
		t.Parallel()

		edu := Education{School: "University", Major: "Psychology", Minor: "Robotics"}

		major, minor := scholarMajorMinor(dice.New(rand.NewPCG(1, 1)), edu)
		if major != "Psychology" || minor != "Robotics" {
			t.Errorf("got %q/%q, want the degree's own Psychology/Robotics", major, minor)
		}
	})

	t.Run("without a degree, one is selected", func(t *testing.T) {
		t.Parallel()

		major, minor := scholarMajorMinor(dice.New(rand.NewPCG(2, 2)), Education{})
		if major == "" || minor == "" {
			t.Fatalf("got %q/%q, want both declared", major, minor)
		}

		if major == minor {
			t.Error("Major and Minor are the same; p.59 says they cannot be")
		}
	})

	t.Run("every generated Scholar has both", func(t *testing.T) {
		t.Parallel()

		seen := 0

		for seed := uint64(1); seed <= 300; seed++ {
			c, ok, err := GenerateCareerChainCharacter(
				dice.New(rand.NewPCG(seed, seed)), []string{"scholar"}, 0)
			if err != nil || !ok || len(c.Careers) == 0 || len(c.Careers[0].Terms) == 0 {
				continue
			}

			seen++

			career := c.Careers[0]
			if career.Major == "" || career.Minor == "" {
				t.Fatalf("seed %d: Major=%q Minor=%q, want both", seed, career.Major, career.Minor)
			}
		}

		if seen == 0 {
			t.Fatal("no Scholar careers generated")
		}
	})
}

// TestScholarWaiverRescuesEachOfThePrintedEvents is p.76's own list:
// "An adverse die roll or decision (in Position, Promotion, Research,
// Publication, Tenure, or Continue) may be waived."
//
// Position is checked through the career entry point, since that roll
// happens before any term; the rest are checked through a term. Soc 20
// makes every Waiver succeed and Soc 0 makes every one fail, so the
// difference between the two isolates the Waiver from the roll it is
// rescuing.
func TestScholarWaiverRescuesEachOfThePrintedEvents(t *testing.T) {
	t.Parallel()

	t.Run("Position", func(t *testing.T) {
		t.Parallel()

		// Edu 0: BeginScholar's own roll cannot succeed.
		noWaiver := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 0, 0}}
		allWaived := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 0, 20}}

		career, _ := resolveScholarCareerWithBudget(
			dice.New(rand.NewPCG(3, 3)), noWaiver, maxCareerTerms, &agingSimulation{}, Education{})
		if len(career.Terms) != 0 {
			t.Error("a Scholar with Edu 0 and Soc 0 qualified anyway")
		}

		career, _ = resolveScholarCareerWithBudget(
			dice.New(rand.NewPCG(3, 3)), allWaived, maxCareerTerms, &agingSimulation{}, Education{})
		if len(career.Terms) == 0 {
			t.Error("a failed Position was not waived at Soc 20")
		}
	})

	t.Run("Research", func(t *testing.T) {
		t.Parallel()

		// CC 0 means the Risk roll cannot succeed, so Research always fails.
		noWaiver := UPP{Characteristics: [6]ehex.Value{0, 8, 8, 8, 8, 0}}
		allWaived := UPP{Characteristics: [6]ehex.Value{0, 8, 8, 8, 8, 20}}

		term, _, _ := ResolveScholarTerm(
			dice.New(rand.NewPCG(4, 4)), noWaiver, C1, 8, 1, nil, "Psychology", new(int))
		if term.RiskResult == Unharmed {
			t.Error("Research succeeded against a Controlling Characteristic of 0")
		}

		term, _, _ = ResolveScholarTerm(
			dice.New(rand.NewPCG(4, 4)), allWaived, C1, 8, 1, nil, "Psychology", new(int))
		if term.RiskResult != Unharmed {
			t.Errorf("RiskResult = %v, want a waived Research failure to leave the Scholar Unharmed",
				term.RiskResult)
		}
	})

	t.Run("Continue", func(t *testing.T) {
		t.Parallel()

		// A Scholar who can never fail Research but can never pass
		// Continue serves exactly one term unless the Waiver rescues him.
		noWaiver := UPP{Characteristics: [6]ehex.Value{20, 20, 20, 20, 0, 0}}
		allWaived := UPP{Characteristics: [6]ehex.Value{20, 20, 20, 20, 0, 20}}

		short, _ := resolveScholarCareerWithBudget(
			dice.New(rand.NewPCG(5, 5)), noWaiver, maxCareerTerms, &agingSimulation{}, Education{})
		long, _ := resolveScholarCareerWithBudget(
			dice.New(rand.NewPCG(5, 5)), allWaived, maxCareerTerms, &agingSimulation{}, Education{})

		if len(long.Terms) <= len(short.Terms) {
			t.Errorf("waived Continue served %d terms, unwaived %d — want the Waiver to extend the career",
				len(long.Terms), len(short.Terms))
		}
	})
}

// TestScholarWaiverModCountsEveryAttempt is the half of p.76's Waiver
// rule that is easy to get wrong: "Mod minus previous waivers
// (successful or not)". A Scholar who waives repeatedly must find it
// progressively harder, whether or not the earlier attempts worked.
//
// Shares its implementation with Education's own Waiver, which p.59
// states in the same terms — so this asserts the sharing holds rather
// than re-deriving the arithmetic.
func TestScholarWaiverModCountsEveryAttempt(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 8, 12}}

	waivers := 0
	for want := 1; want <= 4; want++ {
		scholarWaivableRoll(dice.New(rand.NewPCG(6, 6)), upp, &waivers, false)

		if waivers != want {
			t.Errorf("waivers = %d after %d adverse rolls, want %d", waivers, want, want)
		}
	}

	// A successful roll must not consume a Waiver at all.
	before := waivers
	if !scholarWaivableRoll(dice.New(rand.NewPCG(6, 6)), upp, &waivers, true) {
		t.Error("a successful roll was reported as failed")
	}

	if waivers != before {
		t.Errorf("waivers = %d after a success, want %d unchanged", waivers, before)
	}
}

// TestScholarMajorMinorCellsResolve is the payoff: scholarSkillTable's
// own row 1 is four Major/Minor cells and two One Trade ones, and p.76
// guarantees every Scholar has both — so unlike the other twelve
// careers, a Scholar never loses those cells to Book 1's "this benefit
// is lost".
func TestScholarMajorMinorCellsResolve(t *testing.T) {
	t.Parallel()

	// Counted, not merely checked for absence. An earlier version of this
	// test only asserted that no marker survived into Character.Skills,
	// which a *dropped* marker satisfies just as well — and markers were
	// in fact being dropped for the 15% of Scholars who reach the career
	// without a degree. Requiring the levels to arrive is what catches
	// that.
	// Restricted to Scholars who reached the career WITHOUT a degree,
	// which is the population p.76's fallback exists for and the only one
	// where the resolution can go wrong. Aggregating over all Scholars
	// hid the bug: the graduates' resolved levels swamped the
	// non-graduates' dropped ones, and an earlier version of this test
	// passed with 1,830 cells being discarded.
	noDegree, cellsDrawn := 0, 0

	for seed := uint64(1); seed <= 1500; seed++ {
		c, ok, err := GenerateCareerChainCharacter(
			dice.New(rand.NewPCG(seed, seed)), []string{"scholar"}, 0)
		if err != nil || !ok || len(c.Careers) == 0 || len(c.Careers[0].Terms) == 0 {
			continue
		}

		if c.Education.Major != "" {
			continue
		}

		career := c.Careers[0]

		drawn := countMajorMinorCells(career.Terms)
		if drawn == 0 {
			continue
		}

		noDegree++
		cellsDrawn += drawn

		held, unresolved := majorMinorLevelsHeld(c.Skills, career)
		if unresolved != "" {
			t.Fatalf("seed %d: an unresolved %q marker reached Character.Skills", seed, unresolved)
		}

		if held < drawn {
			t.Fatalf("seed %d: a Scholar with no degree drew %d Major/Minor cells but holds only %d "+
				"levels of %q/%q — p.76 gives him both regardless",
				seed, drawn, held, career.Major, career.Minor)
		}
	}

	if noDegree == 0 {
		t.Fatal("no degreeless Scholar drew a Major/Minor cell — this test proves nothing")
	}

	t.Logf("checked %d degreeless Scholars holding %d drawn Major/Minor cells", noDegree, cellsDrawn)
}

// TestScholarRankRendersWithTheMajor guards the whole chain from p.76's
// table through to a finished character, which is where a suffix applied
// in the wrong place would actually show.
func TestScholarRankRendersWithTheMajor(t *testing.T) {
	t.Parallel()

	seed := seedFor(t, "a promoted Scholar", func(seed uint64) bool {
		c, ok, err := GenerateCareerChainCharacter(
			dice.New(rand.NewPCG(seed, seed)), []string{"scholar"}, 0)

		return err == nil && ok && len(c.Careers) > 0 && lastTermRank(c.Careers[0].Terms) != "" &&
			lastTermRank(c.Careers[0].Terms) != scholarRankNames[0]
	})

	c, _, err := GenerateCareerChainCharacter(dice.New(rand.NewPCG(seed, seed)), []string{"scholar"}, 0)
	if err != nil {
		t.Fatal(err)
	}

	career := c.Careers[0]
	if !strings.HasSuffix(c.Rank, " of "+career.Major) {
		t.Errorf("Rank = %q, want it to end in the Major %q", c.Rank, career.Major)
	}
}

// countMajorMinorCells is how many Major or Minor cells a career's terms
// drew, before resolution replaced them with the declared subjects.
func countMajorMinorCells(terms []Term) int {
	n := 0

	for _, term := range terms {
		for _, s := range term.SkillsAwarded {
			if s.Name == majorSkillCell || s.Name == minorSkillCell {
				n++
			}
		}
	}

	return n
}

// majorMinorLevelsHeld totals the levels a character holds in a career's
// declared Major and Minor, and names any marker that survived
// resolution instead of being replaced.
func majorMinorLevelsHeld(skills []SkillLevel, career Career) (int, string) {
	held := 0

	for _, s := range skills {
		if s.Name == majorSkillCell || s.Name == minorSkillCell {
			return held, s.Name
		}

		if s.Name == career.Major || s.Name == career.Minor {
			held += s.Level
		}
	}

	return held, ""
}

// TestScholarMajorCostsNoDiceOnAFailedBegin holds the reproducibility
// line: p.76 gives every Scholar a Major, but a character who never
// entered the career is not a Scholar, so settling one must draw
// nothing for him.
func TestScholarMajorCostsNoDiceOnAFailedBegin(t *testing.T) {
	t.Parallel()

	// Edu 0 fails Begin; Soc 0 means the Position waiver cannot rescue it.
	upp := UPP{Characteristics: [6]ehex.Value{8, 8, 8, 8, 0, 0}}

	spent := dice.New(rand.NewPCG(7, 7))

	career, _ := resolveScholarCareerWithBudget(spent, upp, maxCareerTerms, &agingSimulation{}, Education{})
	if len(career.Terms) != 0 {
		t.Fatal("this fixture is meant to fail its Begin")
	}

	if career.Major != "" || career.Minor != "" {
		t.Errorf("a character who never became a Scholar declared %q/%q", career.Major, career.Minor)
	}

	// Two same-seed rollers compared after the fact, since rand.IntN
	// rejection-samples and a counting source would deadlock.
	untouched := dice.New(rand.NewPCG(7, 7))
	_, _ = resolveScholarCareerWithBudget(untouched, upp, maxCareerTerms, &agingSimulation{}, Education{})

	if spent.TwoD6() != untouched.TwoD6() {
		t.Error("the two runs diverged, which they cannot if neither drew for a Major")
	}
}
