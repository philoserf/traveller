package render

import (
	"fmt"
	"strings"

	"github.com/philoserf/traveller/character"
)

// timelineEvent is one bio-list bullet: the in-game age it happened at,
// and its own description.
type timelineEvent struct {
	Age  int
	Text string
}

// CharacterTimeline renders c as a chronological list of life events —
// something neither Book 1's own Character Card nor its narrative
// worked examples provide, built instead from what this codebase
// already tracks per Career/Term.
//
// No stored field on Character, Career, or Term records an age or
// sequence position directly (character/aging.go's own agingSimulation
// computes one internally per term, then discards it once finalizeAging
// collapses everything to a single final Age). This reconstructs the
// same arithmetic from outside the package: age starts at
// character.AgeFromTermsServed(0) (18) and advances by each Term's own
// Length (normally 4, sometimes 3 for a Flight-School-shortened first
// term) as Careers are walked in order — the same "read the chain in
// slice order" rule career_chain.go's own accumulator already follows.
//
// Two honest limitations, not silently smoothed over:
//
//   - A failed Begin/Retry attempt costs a year (Book 1 p.65) that this
//     reconstruction has no way to see — agingSimulation.failedYears is
//     never stored on Character/Career/Term, so every age after a
//     failed attempt undercounts by however many attempts failed. A
//     trailing footnote says so on every call, not just when it might
//     matter, since there's no way to tell from outside whether it does.
//   - Education (CharGen step C) consumes no modeled years at all in
//     this codebase — resolveEducation runs before any agingSimulation
//     exists, so every character's career clock starts at exactly 18
//     regardless of how many years of schooling Education.Passes/
//     Graduate.Passes represent. Its own events are placed at Age 18
//     because that is what this codebase's model actually says
//     happened, not an approximation of it.
func CharacterTimeline(c character.Character) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Born %s on %s.\n\n", c.Birthdate, c.Homeworld)

	events := educationTimelineEvents(c.Education)
	age := character.AgeFromTermsServed(0)

careers:
	for _, career := range c.Careers {
		for _, term := range career.Terms {
			age += term.Length

			termEvents, dead := termTimelineEvents(career, term, age)
			events = append(events, termEvents...)

			if dead {
				break careers
			}
		}

		if musteringOutHasContent(career.MusteringOut) {
			events = append(events, timelineEvent{
				Age: age, Text: fmt.Sprintf("Mustered out of the %s.", career.Name),
			})
		}
	}

	for _, e := range events {
		fmt.Fprintf(&b, "- Age %d: %s\n", e.Age, e.Text)
	}

	fmt.Fprint(&b, "\n_Ages are reconstructed from term lengths and may undercount by up to a "+
		"year per failed Begin/Retry attempt, which this codebase does not record on the character._\n")

	return b.String()
}

// educationTimelineEvents is CharGen step C's own events, all placed at
// Age 18 — see CharacterTimeline's own doc comment for why that is
// exact, not an approximation, in this codebase's model. Nothing is
// returned for a character who attended no institution at all.
//
// CommissionCareers is reported as "Earned a Commission" (singular,
// generic) rather than naming which careers it covers — which one it
// actually applied to is a fact about a later Career/Term (Commissioned
// true), not about Education itself; see termTimelineEvents.
func educationTimelineEvents(edu character.Education) []timelineEvent {
	if !edu.Attended() {
		return nil
	}

	const startAge = 18

	var events []timelineEvent

	add := func(format string, args ...any) {
		events = append(events, timelineEvent{Age: startAge, Text: fmt.Sprintf(format, args...)})
	}

	add("Attended %s.", edu.School)

	if edu.Degree != "" {
		add("Graduated %s with a %s.", edu.School, edu.Degree)
	}

	if edu.Honors {
		add("Earned Honors.")
	}

	for grad := edu.Graduate; grad != nil; grad = grad.Next {
		if grad.Graduated {
			add("Graduated %s as %s.", grad.School, grad.Degree)
		} else {
			add("Attended %s.", grad.School)
		}
	}

	if len(edu.CommissionCareers) > 0 {
		add("Earned a Commission.")
	}

	return events
}

// termTimelineEvents is one term's own notable events, tagged at age
// (the running age after this term completes) — an ordinary Unharmed
// term with no Commission/Promotion/Reward/special outcome contributes
// nothing, the same "don't bury the real events" reasoning
// termOutcomeLine (character.go) already documents for the sheet.
// term.Medals is deliberately not listed separately: a Reward-earned
// medal already appears, by its own full name, in the RewardResult
// event below, and an Unharmed-only "XS Exemplary Service" grant isn't
// exceptional enough on its own to earn a line — Unharmed is the
// expected outcome, not a notable one.
//
// Returns dead=true when this term's own RiskResult is a real death
// (not Entertainer's own reused "Talent Exhausted" or Functionary's own
// reused "Office Politics" meaning — both override the generic Dead/
// Disabled labels the same way termOutcomeLine's own family already
// does), so the caller can stop walking further terms.
func termTimelineEvents(career character.Career, term character.Term, age int) ([]timelineEvent, bool) {
	var events []timelineEvent

	add := func(format string, args ...any) {
		events = append(events, timelineEvent{Age: age, Text: fmt.Sprintf(format, args...)})
	}

	if term.Commissioned {
		add("Commissioned as Officer1 in the %s.", career.Name)
	}

	if term.Promoted {
		add("Promoted to %s.", term.Rank)
	}

	dead := addRiskResultEvent(add, career.Name, term.RiskResult)

	if term.RewardResult != "" && term.RewardResult != "None" {
		add("Reward: %s.", term.RewardResult)
	}

	addCareerHighlightEvents(add, career.Name, term)

	// Masterpiece.CreatedAtAge is exact where the running total isn't —
	// use it verbatim rather than the reconstructed age.
	if term.Masterpiece != nil {
		events = append(events, timelineEvent{
			Age:  term.Masterpiece.CreatedAtAge,
			Text: fmt.Sprintf("Created a Masterpiece (%d Master Points).", term.Masterpiece.MasterPoints),
		})
	}

	return events, dead
}

// addRiskResultEvent reports term.RiskResult's own event, if any, and
// returns true only for a real death — Entertainer's own reused Dead
// ("Talent Exhausted") and Functionary's own reused Disabled ("Office
// Politics") are overridden the same way termOutcomeLine's family
// already does, and neither one ends the timeline the way a real death
// does.
func addRiskResultEvent(add func(string, ...any), careerName string, result character.RiskResult) bool {
	switch result {
	case character.Unharmed:
		// The expected outcome; earns no line.
	case character.Wounded:
		add("Wounded in the %s.", careerName)
	case character.Disabled:
		if careerName == character.FunctionaryCareerName {
			add("Office Politics ended %s service.", careerName)
		} else {
			add("Disabled, ending %s service.", careerName)
		}
	case character.Dead:
		if careerName == character.EntertainerCareerName {
			add("Talent exhausted, ending %s career.", careerName)
		} else {
			add("Died in the %s.", careerName)

			return true
		}
	}

	return false
}

// addCareerHighlightEvents reports the career-specific highlights
// render/character.go's own termOutcomeLine family already names:
// Noble Elevation, Scholar Tenure/award-winning works, Rogue
// imprisonment.
func addCareerHighlightEvents(add func(string, ...any), careerName string, term character.Term) {
	switch careerName {
	case character.NobleCareerName:
		if term.Elevated {
			add("Elevated to %s.", term.Rank)
		}
	case character.ScholarCareerName:
		if term.TenureGranted {
			add("Granted Tenure.")
		}

		if term.AwardWinning {
			add("Published an award-winning work.")
		}
	case character.RogueCareerName:
		if term.Imprisoned {
			add("Imprisoned for %d years.", term.PrisonYears)
		}
	}
}

// musteringOutHasContent mirrors writeMusteringOut's own "nothing was
// granted" gate (character.go) — a never-qualified or died-before-
// Mustering-Out career leaves every field empty, and that earns no
// timeline entry.
func musteringOutHasContent(m character.MusteringOut) bool {
	return len(m.Automatics) > 0 || len(m.Benefits) > 0 || len(m.Money) > 0 || len(m.Entitlements) > 0 ||
		m.Pension != 0 || m.RetirementPay != 0
}
