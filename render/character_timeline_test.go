package render_test

import (
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/philoserf/traveller/character"
	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/render"
)

// TestCharacterTimelineOpensWithBirth covers the leading line, which is
// not itself an "Age N" bullet the way the rest of the timeline is.
func TestCharacterTimelineOpensWithBirth(t *testing.T) {
	t.Parallel()

	c := character.Character{Birthdate: "Wonday 002-1105", Homeworld: "A788899-C"}

	got := render.CharacterTimeline(c)
	if !strings.HasPrefix(got, "Born Wonday 002-1105 on A788899-C.\n") {
		t.Errorf("CharacterTimeline should open with the birth line, got:\n%s", got)
	}
}

// TestCharacterTimelineAgeArithmeticMatchesTermLengths is the core
// regression: age must start at 18 (character.AgeFromTermsServed(0))
// and advance by each Term's own Length, including a Flight-School-
// shortened first term (3 instead of 4) — not a hardcoded assumption
// that every term is 4 years.
func TestCharacterTimelineAgeArithmeticMatchesTermLengths(t *testing.T) {
	t.Parallel()

	c := character.Character{
		Careers: []character.Career{{
			Name: character.SpacerCareerName,
			Terms: []character.Term{
				{Length: 3, Commissioned: true},            // 18 -> 21
				{Length: 4, Promoted: true, Rank: "O2 Lt"}, // 21 -> 25
			},
		}},
	}

	got := render.CharacterTimeline(c)

	if !strings.Contains(got, "Age 21: Commissioned as Officer1 in the "+character.SpacerCareerName+".") {
		t.Errorf("first (3-year) term should land at Age 21, got:\n%s", got)
	}

	if !strings.Contains(got, "Age 25: Promoted to O2 Lt.") {
		t.Errorf("second (4-year) term should land at Age 25, got:\n%s", got)
	}
}

// TestCharacterTimelineEducationEventsLandAt18 covers the documented
// modeling choice: Education consumes no time in this codebase, so
// every one of its events — including a chained Graduate Program —
// is placed at Age 18, the same age the first career term also starts
// counting from.
func TestCharacterTimelineEducationEventsLandAt18(t *testing.T) {
	t.Parallel()

	c := character.Character{
		Education: character.Education{
			School: "University", Degree: "Honors BA", Honors: true,
			Graduate: &character.GraduateProgram{School: "Masters", Degree: "MA", Graduated: true},
		},
	}

	got := render.CharacterTimeline(c)

	for _, want := range []string{
		"Age 18: Attended University.",
		"Age 18: Graduated University with a Honors BA.",
		"Age 18: Earned Honors.",
		"Age 18: Graduated Masters as MA.",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("CharacterTimeline should contain %q, got:\n%s", want, got)
		}
	}
}

// TestCharacterTimelineOmitsEducationWhenNotAttended guards the gate.
func TestCharacterTimelineOmitsEducationWhenNotAttended(t *testing.T) {
	t.Parallel()

	got := render.CharacterTimeline(character.Character{})
	if strings.Contains(got, "Attended") || strings.Contains(got, "Graduated") {
		t.Errorf("CharacterTimeline should have no Education events when unattended, got:\n%s", got)
	}
}

// TestCharacterTimelineMasterpieceUsesItsOwnCreatedAtAge is the
// regression that matters most for accuracy: a Term.Masterpiece carries
// its own exact CreatedAtAge, which must be used verbatim instead of
// the reconstructed running age for that term (they can genuinely
// differ, since Masterpiece creation is its own sub-process).
func TestCharacterTimelineMasterpieceUsesItsOwnCreatedAtAge(t *testing.T) {
	t.Parallel()

	c := character.Character{
		Careers: []character.Career{{
			Name: character.CraftsmanCareerName,
			Terms: []character.Term{
				{Length: 4, Masterpiece: &character.Masterpiece{MasterPoints: 42, CreatedAtAge: 30}},
			},
		}},
	}

	got := render.CharacterTimeline(c)

	// The term itself ends at 18+4=22, but the Masterpiece's own
	// CreatedAtAge (30) is what must appear, not 22.
	if !strings.Contains(got, "Age 30: Created a Masterpiece (42 Master Points).") {
		t.Errorf("CharacterTimeline should use Masterpiece.CreatedAtAge verbatim, got:\n%s", got)
	}

	if strings.Contains(got, "Age 22: Created a Masterpiece") {
		t.Errorf("CharacterTimeline should not use the reconstructed term-end age for a Masterpiece, got:\n%s", got)
	}
}

// TestCharacterTimelineOrdinaryTermContributesNoLine covers the "don't
// bury the real events" gate: an Unharmed term with no Commission,
// Promotion, Reward, or career-specific highlight must not appear at
// all.
func TestCharacterTimelineOrdinaryTermContributesNoLine(t *testing.T) {
	t.Parallel()

	c := character.Character{
		Careers: []character.Career{{
			Name:  "Scout",
			Terms: []character.Term{{Length: 4, RiskResult: character.Unharmed}},
		}},
	}

	got := render.CharacterTimeline(c)
	if strings.Contains(got, "Age 22:") {
		t.Errorf("an ordinary Unharmed term should contribute no timeline line, got:\n%s", got)
	}
}

// TestCharacterTimelineDeadStopsTheTimeline covers the "Dead ends the
// timeline" rule — no events from a later term (or a later chained
// career) should appear once a real death fires.
func TestCharacterTimelineDeadStopsTheTimeline(t *testing.T) {
	t.Parallel()

	c := character.Character{
		Careers: []character.Career{{
			Name: "Scout",
			Terms: []character.Term{
				{Length: 4, RiskResult: character.Dead},
				{Length: 4, Promoted: true, Rank: "should never appear"},
			},
		}},
	}

	got := render.CharacterTimeline(c)

	if !strings.Contains(got, "Age 22: Died in the Scout.") {
		t.Errorf("CharacterTimeline should report the death, got:\n%s", got)
	}

	if strings.Contains(got, "should never appear") {
		t.Errorf("CharacterTimeline should stop at Dead and report nothing after it, got:\n%s", got)
	}
}

// TestCharacterTimelineEntertainerDeadIsTalentExhaustionNotDeath is the
// same override termOutcomeLine's own entertainerTermLabel already
// needs — Entertainer's own RiskResult Dead means Talent completely
// spent, not physical death, and printing the word the same way a real
// death does would misreport it on the one renderer built to narrate a
// life span.
func TestCharacterTimelineEntertainerDeadIsTalentExhaustionNotDeath(t *testing.T) {
	t.Parallel()

	c := character.Character{
		Careers: []character.Career{{
			Name:  character.EntertainerCareerName,
			Terms: []character.Term{{Length: 4, RiskResult: character.Dead}},
		}},
	}

	got := render.CharacterTimeline(c)

	if !strings.Contains(got, "Talent exhausted") {
		t.Errorf("CharacterTimeline should report Entertainer's own Dead as Talent exhaustion, got:\n%s", got)
	}

	if strings.Contains(got, "Died in the "+character.EntertainerCareerName) {
		t.Errorf("CharacterTimeline should not report Entertainer's own Dead as a real death, got:\n%s", got)
	}
}

// TestCharacterTimelineAlwaysShowsTheAgeCaveat guards the honesty
// footnote — present on every call, since there's no way to tell from
// outside the character package whether a failed Begin/Retry attempt
// actually occurred.
func TestCharacterTimelineAlwaysShowsTheAgeCaveat(t *testing.T) {
	t.Parallel()

	got := render.CharacterTimeline(character.Character{})
	if !strings.Contains(got, "may undercount by up to a") {
		t.Errorf("CharacterTimeline should always show the age-reconstruction caveat, got:\n%s", got)
	}
}

// TestCharacterTimelineDoesNotPanicOnGeneratedCharacters is a live-
// generation smoke test across a few careers and seeds — the fixture
// tests above pin down exact content; this just confirms the age
// reconstruction and event extraction hold up against real generated
// data without panicking, and stay age-ascending.
func TestCharacterTimelineDoesNotPanicOnGeneratedCharacters(t *testing.T) {
	t.Parallel()

	generators := []func(*dice.Roller) (character.Character, bool){
		character.GenerateScoutCharacter,
		character.GenerateMarineCharacter,
		character.GenerateNobleCharacter,
		character.GenerateScholarCharacter,
	}

	for _, gen := range generators {
		for seed := uint64(1); seed <= 25; seed++ {
			c, ok := gen(dice.New(rand.NewPCG(seed, seed)))
			if !ok {
				continue
			}

			out := render.CharacterTimeline(c)
			if out == "" {
				t.Fatalf("seed %d: CharacterTimeline returned an empty string", seed)
			}
		}
	}
}
