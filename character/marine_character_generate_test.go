package character

import (
	"math/rand/v2"
	"slices"
	"strings"
	"testing"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

func TestBuildMarineCharacterNeverQualified(t *testing.T) {
	t.Parallel()

	upp := UPP{}
	homeworldSkills := []SkillLevel{{Name: "Vacc Suit", Level: 1, Kind: Skill}}
	r := dice.New(rand.NewPCG(1, 1))

	c, ok := buildMarineCharacter(r, upp, "A000000-0", homeworldSkills, false)

	if ok {
		t.Error("ok = true, want false (never qualified)")
	}

	if len(c.Careers) != 1 || len(c.Careers[0].Terms) != 0 {
		t.Errorf("Careers = %+v, want one Career with zero Terms", c.Careers)
	}

	if c.WoundBadges != 0 {
		t.Errorf("WoundBadges = %d, want 0", c.WoundBadges)
	}

	if c.Age != 19 {
		t.Errorf("Age = %d, want 19 (18 plus the one year the failed Begin roll cost, Book 1 p.65)", c.Age)
	}

	if !slices.Equal(c.Skills, homeworldSkills) {
		t.Errorf("Skills = %v, want exactly homeworldSkills %v (no career skills)", c.Skills, homeworldSkills)
	}
}

// TestBuildMarineCharacterDies confirms ok exactly matches the last
// term's own RiskResult != Dead — mirroring TestBuildScoutCharacterDies.
// C1=12 guarantees Begin always succeeds (Marine has no Retry, unlike
// Scout, so Begin needs to be reliable on its own); C4=1 makes death
// likely whenever the C1/C4 rotation lands on C4 (Risk target =
// 1-combinedMod, effectively always <=0).
func TestBuildMarineCharacterDies(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{12, 0, 0, 1, 8, 0}}
	r := dice.New(rand.NewPCG(41, 43))

	sawDeath := false

	for range 200 {
		c, ok := buildMarineCharacter(r, upp, "hw", nil, false)

		terms := c.Careers[0].Terms
		if len(terms) == 0 {
			t.Fatal("career has zero terms, want Begin to always succeed (C1=12)")
		}

		// Either kill path clears ok: a Dead last term (Book 1 p.69) or
		// an Aging death (p.89), which now runs between terms.
		wantOK := terms[len(terms)-1].RiskResult != Dead &&
			!strings.Contains(c.Notes, "died of natural causes")
		if ok != wantOK {
			t.Fatalf("ok = %v, want %v (last term RiskResult = %v, Notes = %q)",
				ok, wantOK, terms[len(terms)-1].RiskResult, c.Notes)
		}

		if terms[len(terms)-1].RiskResult == Dead {
			sawDeath = true
		}
	}

	if !sawDeath {
		t.Fatal("no trial produced a death across 200 trials — fixture can't verify the ok=false path")
	}
}

// TestBuildMarineCharacterQualified confirms a real career actually
// populates Age/LifeStage/Birthdate/Skills/WoundBadges/Rank, mirroring
// the equivalent Scout/Citizen/Noble build tests. C1=C4=20 is high enough
// that even the worst-case combined Mod (Branch's own max 2 +
// Operations' own max 3 = 5) still leaves a target above 2D6's own
// maximum, guaranteeing survival and Continue (see
// TestResolveMarineCareerRespectsMaxTermsCap's own identical reasoning).
// C3=End=0 keeps Officer Commission's own target unreachable (2D6's own
// minimum is 2), so this character stays Enlisted the whole career;
// C1=20 guarantees Enlisted Promotion always succeeds, deterministically
// capping at M6 Sergeant Major well within maxCareerTerms.
func TestBuildMarineCharacterQualified(t *testing.T) {
	t.Parallel()

	upp := UPP{Characteristics: [6]ehex.Value{20, 0, 0, 20, 8, 0}}
	homeworldSkills := []SkillLevel{{Name: "Vacc Suit", Level: 1, Kind: Skill}}
	r := dice.New(rand.NewPCG(1, 1))

	c, ok := buildMarineCharacter(r, upp, "hw", homeworldSkills, false)

	if !ok {
		t.Fatal("ok = false, want true (fixture guarantees survival)")
	}

	if len(c.Careers) != 1 || c.Careers[0].Name != MarineCareerName {
		t.Fatalf("Careers = %+v, want one Career named %q", c.Careers, MarineCareerName)
	}

	if !c.Careers[0].HasRank {
		t.Error("Careers[0].HasRank = false, want true")
	}

	if len(c.Careers[0].Terms) == 0 {
		t.Fatal("Careers[0].Terms is empty, want at least one term")
	}

	if c.Age < 18 {
		t.Errorf("Age = %d, want >= 18", c.Age)
	}

	if want := LifeStageForAge(c.Age); c.LifeStage != want {
		t.Errorf("LifeStage = %d, want %d (LifeStageForAge(%d))", c.LifeStage, want, c.Age)
	}

	assertBirthdateFormat(t, c.Birthdate, c.Age)

	if len(c.Skills) < len(homeworldSkills) {
		t.Errorf("len(Skills) = %d, want at least %d (homeworld skills alone)", len(c.Skills), len(homeworldSkills))
	}

	// Fame, WoundBadges and Medals are all checked against what this
	// character's own terms actually record rather than against pinned
	// totals. They used to be pinned, on the premise that a high-UPP
	// fixture made every Risk roll succeed for all 14 terms — a premise
	// Aging retired: it now runs between terms and erodes C1-C3, so even
	// a maximal starting characteristic can fall far enough for a late
	// Risk roll to fail. Deriving keeps what these assertions were
	// actually for (marineCareerFameAwards wired end to end, and the
	// code-review-caught bug where Term.Medals fed Fame but never reached
	// Character.Medals) without re-pinning a number every Aging change
	// would invalidate again.
	if want := sumInts(marineCareerFameAwards(c.Careers[0])); c.Fame < want {
		t.Errorf("Fame = %d, want at least %d (career Fame, before Mustering Out's own additions)", c.Fame, want)
	}

	if want := scoutWoundBadges(c.Careers[0]); c.WoundBadges != want {
		t.Errorf("WoundBadges = %d, want %d (one per Wounded/Disabled term)", c.WoundBadges, want)
	}

	if want := allMedalsFromTerms(c.Careers[0].Terms); len(c.Medals) != len(want) {
		t.Errorf("len(Medals) = %d, want %d (every Term.Medals entry must reach Character.Medals)",
			len(c.Medals), len(want))
	}

	if len(c.Medals) == 0 {
		t.Error("len(Medals) = 0, want some (fixture can't verify Medals propagation at all)")
	}

	// Regression test for Phase V's own core mechanic: promotion always
	// succeeds against C1=20, so the rank reaches the top of its ladder
	// well before the career's own 14-term run ends.
	//
	// Which ladder is no longer fixed. This previously asserted the
	// Enlisted cap outright, on the premise that Commission could never
	// succeed against C3=End=0 — a premise that stopped holding once Term
	// skills were restricted to the Operations columns actually received
	// (Book 1 p.65). Fewer eligible columns means the Personal column,
	// which grants characteristics, is drawn far more often, so C3 no
	// longer stays at 0. The cap itself is what this test is for, so it
	// is asserted against whichever ladder the character ended on.
	topEnlisted := marineRankName(false, len(marineEnlistedRankNames))
	topOfficer := marineRankName(true, len(marineOfficerRankNames))

	if c.Rank != topEnlisted && c.Rank != topOfficer {
		t.Errorf("Rank = %q, want the top of a ladder (%q or %q)", c.Rank, topEnlisted, topOfficer)
	}
}
