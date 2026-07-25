package character

import (
	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// NobleCareerName is Noble's own Career.Name value — exported and shared
// (ResolveNobleCareer's own Career{Name: NobleCareerName} literal below,
// a future render dispatch) as a single source of truth, rather than two
// independent "Noble" string literals that could silently drift apart on
// a future rename.
const NobleCareerName = "Noble"

// nobleReturnIntriguePositions is Return & Intrigue's own Controlling
// Characteristic set (Book 1 p.85: "Return & Intrigue C2 C3 C4 C5").
var nobleReturnIntriguePositions = []Position{C2, C3, C4, C5}

// nobleBaseFame is Book 1 p.85's own "Fame. Nobles have a Base Fame equal
// to 1.5 times Soc." Integer math (3*soc/2, floor-rounded) rather than
// float64 — the book doesn't specify rounding for a fractional result
// (odd Soc values), and floor is the simplest, most defensible default
// absent a stated rule. Distinct from ApplyMusteringOut's own Fame
// accumulation (Mustering Out's "Fame +N" rolls) — this is what a Noble
// starts with, not an end-of-career total.
func nobleBaseFame(soc ehex.Value) int {
	return 3 * int(soc) / 2
}

// nobleSkillTable is Book 1 p.85's "C NOBLE SKILLS" table (7 columns, 1D
// per row). Column 1 (Personal) is the universal six-characteristic
// boost list every career's own skill table shares (row N always names
// characteristic C(N+1) — matching scoutSkillTable's own "Str Dex End
// Int Edu Soc" literally, not the book's own mixed "Str +1"/"C2 +1"
// box shorthand for this column, which skillLevel1 already normalizes
// away). "Capital" (footnoted in the book as "*** Capital= World
// Knowledge (of world of highest held noble Land Grant) (value=1D)") has
// no structured Land Grant tracking in this codebase yet, so it's
// handled by rollSkillFromTable's own existing unresolvable-token
// bucket, the same treatment "Major"/"Minor"/"One Science" already get
// for the identical reason (a real mechanic this codebase doesn't yet
// implement). "C6 +1"'s own footnote ("if C6=Caste, lost") is provably
// unreachable — Caste is Soc's non-Human analog (p.58's own UPP table),
// and this codebase only generates Human characters — noted, not
// implemented, mirroring this codebase's existing precedent for
// provably-unreachable branches elsewhere (e.g. continueScoutOutcome's
// own roll==2 case).
var nobleSkillTable = [7][6]string{
	{"Str", "Dex", "End", "Int", "Edu", "Soc"},
	{"Major", "Major", "Minor", "Minor", "One Trade", "One Trade"},
	{"Driver", "Flyer", "Pilot", "Starship Skill", "High-G", "Zero-G"},
	{"Advocate", "Counsellor", "Bureaucrat", "Liaison", "Leader", "Leader"},
	{"Liaison", "Strategy", "Tactics", "Diplomat", "Advocate", "Leader"},
	{"Capital", "Admin", "Language", "Starship Skill", "Bureaucrat", "Comms"},
	{"One Art", "One Science", "Athlete", "Medic", "Seafarer", "One Trade"},
}

// nobleSkillsPerTerm is Book 1 p.85's own "Skill Eligibility: Per Term 4".
const nobleSkillsPerTerm = 4

// BeginNoble reports Book 1 p.85's own hard prerequisite: "To Begin
// Automatic* ... *if Soc B+" — Soc >= 11 (ehex B), not a roll. No
// fallback path is described for a lower Soc; this career simply isn't
// attemptable.
func BeginNoble(soc ehex.Value) bool {
	return soc >= 11
}

// nobleIntrigueCounts derives successfulIntrigues and timesExiled from
// terms — shared by returnIntrigueMod (their combined Mod, p.85) and
// nobleExileFame (p.91's own "Per Exile +1" Fame source), so the
// counting loop itself only exists once.
func nobleIntrigueCounts(terms []Term) (int, int) {
	successfulIntrigues, timesExiled := 0, 0

	for _, t := range terms {
		if t.NobleAction != "Intrigue" {
			continue
		}

		if t.NobleSucceeded {
			successfulIntrigues++
		} else {
			timesExiled++
		}
	}

	return successfulIntrigues, timesExiled
}

// returnIntrigueMod is the shared Return & Intrigue Mod, Book 1 p.85's
// own "Mod= -Successful Intrigues. Mod=+Times Exiled." — derived from
// priorTerms rather than tracked as separate mutable counters, so it
// can't drift out of sync with the Terms it's counting (the same
// discipline citizenLifeSuccessCount already established).
//
// Read as ONE combined value (timesExiled - successfulIntrigues) applied
// to BOTH Return and Intrigue via the book's own universal Risk(+Mod)/
// Reward(-Mod, opposite sign) convention — not two independent per-roll
// mods (Intrigue's own mod = -successfulIntrigues only; Return's own
// mod = +timesExiled only, no cross-term). Re-verified directly against
// two independent primary-source anchors, since the page text alone
// ("Mod=" stated twice) is genuinely terse enough to support either
// reading: (1) p.85 itself labels Return "Roll R&R CC +Mods" and
// Intrigue "Roll R&R CC +(opposite sign) Mods" — singular "Mods," the
// same phrasing the book uses everywhere else for a combined total, not
// "Intrigue's Mod" / "Return's Mod" as two separate named values; (2)
// p.65's own Armed Forces body text describes the identical shape for a
// career that indisputably sums multiple named Mod sources (Bravery/
// Caution, Branch, Operations) into one value before applying it:
// "Roll for Risk (2D <= Characteristic + Mods)... rolls again for Reward
// (2D <= Characteristic - Mods)" — one combined Mods, reused with a sign
// flip, exactly this function's own shape. Both anchors point the same
// direction; the "two independent mods" reading would be the outlier
// against this book's own established Career Resolution convention.
func returnIntrigueMod(priorTerms []Term) int {
	successfulIntrigues, timesExiled := nobleIntrigueCounts(priorTerms)

	return timesExiled - successfulIntrigues
}

// nobleExileFame is Book 1 p.91's own "Imperial Noble: Per Exile +1" — a
// second Base Fame source alongside nobleBaseFame's own "Soc x1.5,"
// missed in this codebase's own first Noble Fame pass since p.91 hadn't
// been read yet at that point.
func nobleExileFame(terms []Term) int {
	_, timesExiled := nobleIntrigueCounts(terms)

	return timesExiled
}

// nobleIsExiled derives current Exile status from the last Term's own
// outcome (Book 1 p.85's own table: a failed Intrigue causes Exile; a
// successful Return ends it; a failed Return leaves the character still
// exiled — all three collapse to the same "!NobleSucceeded" check). No
// terms yet means not exiled.
func nobleIsExiled(priorTerms []Term) bool {
	return len(priorTerms) > 0 && !priorTerms[len(priorTerms)-1].NobleSucceeded
}

// nobleFluxAlreadyUsed reports whether the one-time p.85 Elevation Flux
// bonus has already been spent — true iff any prior term already had a
// successful Intrigue (the only trigger for an Elevation roll at all).
func nobleFluxAlreadyUsed(priorTerms []Term) bool {
	for _, t := range priorTerms {
		if t.NobleAction == "Intrigue" && t.NobleSucceeded {
			return true
		}
	}

	return false
}

// nobleElevationOutcome is Book 1 p.85's own Elevation check: "Roll Soc
// or greater" — the opposite comparison direction from succeedsAgainst's
// own roll<=target convention, the same kind of deliberate, documented
// divergence agingCheckOutcome already established a precedent for.
func nobleElevationOutcome(roll, soc, flux int) bool {
	return roll+flux >= soc
}

// rollNobleElevation rolls Elevation, invoking the one-time p.85 Flux
// bonus automatically when useFlux is true — ResolveNobleTerm decides
// eligibility (nobleFluxAlreadyUsed) and always spends it the first time
// it's available: Book 1 leaves the timing open, and Flux itself already
// supplies the randomness, so there's no clear asymmetry to make a real
// strategic choice over (the same reasoning already applied to Scout's
// own Duty pick and Mustering Out's own column pick).
func rollNobleElevation(r *dice.Roller, soc int, useFlux bool) bool {
	flux := 0
	if useFlux {
		flux = r.Flux()
	}

	return nobleElevationOutcome(r.TwoD6(), soc, flux)
}

// ResolveNobleTerm resolves one 4-year Noble term (Book 1 p.85) for the
// controlling characteristic ccPos, given priorTerms for deriving Exile
// status, the running Mod, and Flux availability. Grants
// nobleSkillsPerTerm skills from nobleSkillTable regardless of
// Return/Intrigue outcome — Book 1 doesn't gate Noble's own skill grants
// on success the way Scout's Duty does.
func ResolveNobleTerm(r *dice.Roller, upp UPP, ccPos Position, priorTerms []Term) Term {
	mod := returnIntrigueMod(priorTerms)
	cc := int(upp.Characteristics[ccPos])

	term := Term{Length: 4, ControllingCharacteristic: ccPos}

	if nobleIsExiled(priorTerms) {
		term.NobleAction = "Return"
		term.NobleSucceeded = rollAgainstTarget(r, cc, mod)
	} else {
		term.NobleAction = "Intrigue"
		term.NobleSucceeded = rollAgainstTarget(r, cc, -mod)

		if term.NobleSucceeded {
			term.Elevated = rollNobleElevation(r, int(upp.Characteristics[C6]), !nobleFluxAlreadyUsed(priorTerms))
		}
	}

	term.SkillsAwarded = rollSkillsFromTable(r, nobleSkillTable, nobleSkillsPerTerm)

	return term
}
