package character

import (
	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// succeedsAgainst is rollAgainstTarget's own dice-free comparison,
// split out the same way scoutRiskOutcome is so the pass/fail boundary
// is directly testable against a fixed roll value instead of a real
// 2D6 draw.
func succeedsAgainst(roll, target, mod int) bool {
	return roll <= target+mod
}

// rollAgainstTarget is Book 1 p.63's universal resolution primitive:
// "2D, result <= Target = success." mod raises or lowers the effective
// target (a positive mod makes success easier, matching the book's own
// "a Mod raises the Target" framing) — shared by every Begin/Retry/
// Risk/Reward/Continue check across all 13 careers, not just Scout.
func rollAgainstTarget(r *dice.Roller, target, mod int) bool {
	return succeedsAgainst(r.TwoD6(), target, mod)
}

// highestOf returns whichever of positions indexes the highest
// Characteristics value in upp, first one wins on a tie.
func highestOf(upp UPP, positions ...Position) Position {
	best := positions[0]

	for _, p := range positions[1:] {
		if upp.Characteristics[p] > upp.Characteristics[best] {
			best = p
		}
	}

	return best
}

// BeginScout resolves Book 1 p.79's "To Begin" check for the Scout
// career ("To Begin: C1 or C2 or C3"; "Retry R&R: C5" — confirmed
// against Book 1's own generic-engine text, "Some Careers allow Retry.
// If Begin fails, the character may immediately Retry," to mean a Begin
// retry, not a Risk/Reward retry, despite the "R&R" in the label): 2D
// against the highest of Str/Dex/End, with one retry against Education
// on failure. Picking the highest of the three for the initial attempt
// is this generator's own resolution of an open choice the book leaves
// to the player — unlike homeworld_generate.go's Art/Trade picks
// (rollChoice), there's a genuinely better option here (higher
// characteristic, better odds), not just flavor, so a uniform random
// pick would be a worse default, not just a different one.
//
// Returns the Position chosen for the initial attempt alongside the
// success bool — not just the bool — so a caller can thread the same
// Position into ResolveScoutTerm's ControllingCharacteristic instead of
// re-deriving it from upp a second time. Re-deriving would silently
// diverge from BeginScout's own choice if a future caller ever persists
// a Risk roll's characteristic reduction back onto a Character's UPP
// between terms — resolveScoutRisk's own returned reduced value isn't
// written back onto upp anywhere in this file today, but nothing stops
// a future caller from doing so.
func BeginScout(r *dice.Roller, upp UPP) (Position, bool) {
	ccPos := highestOf(upp, C1, C2, C3)
	if rollAgainstTarget(r, int(upp.Characteristics[ccPos]), 0) {
		return ccPos, true
	}

	return ccPos, rollAgainstTarget(r, int(upp.Characteristics[C5]), 0)
}

// ScoutDuty is which of Book 1 p.79's two Scout duty assignments a term
// serves under — Courier Duty skips the Risk & Reward roll entirely but
// grants fewer skills; Explorer Duty takes the roll for more skills.
type ScoutDuty int

// ScoutDuty values.
const (
	ExplorerDuty ScoutDuty = iota
	CourierDuty
)

// rollScoutDuty picks a Scout's duty for the term. The book offers this
// as a genuine safety-vs-reward strategic choice ("A Scout may avoid
// the Risk and Reward rolls by volunteering for Courier Duty") with no
// dice mechanic for which a character picks — resolved here via a
// uniform random pick, the same non-interactive-generator resolution
// homeworld_generate.go's rollChoice uses for its own open player
// choices (Ri/In).
func rollScoutDuty(r *dice.Roller) ScoutDuty {
	if r.Uniform(2) == 1 {
		return ExplorerDuty
	}

	return CourierDuty
}

// scoutRiskOutcome maps a Risk failure's original and reduced
// characteristic values to a RiskResult, split out from
// resolveScoutRisk's own dice rolls so the mapping itself is
// unit-testable against fixed fixtures — matching this project's
// existing pattern of keeping a rule's own logic dice-free and
// testable (e.g. world.techLevelModifier). Checked in priority order
// since more than one case can technically apply to the same reduction:
// reduced to 0 is Dead first and foremost, even when original was
// already 0 (loss=0 there too, which would otherwise misreport as
// Unharmed); otherwise no net loss (Flux offset the Mods) is Unharmed
// even though the roll technically failed; a 4+ loss is Disabled;
// anything less is Wounded.
func scoutRiskOutcome(original, reduced int) RiskResult {
	loss := original - reduced

	switch {
	case reduced == 0:
		return Dead
	case loss <= 0:
		return Unharmed
	case loss >= 4:
		return Disabled
	default:
		return Wounded
	}
}

// resolveScoutRisk resolves one Risk roll against cc+mod (Book 1 p.79's
// Scout Risk & Reward box, itself an instance of the universal Risk
// mechanic — Book 1 p.65). On success, Unharmed. On failure: reduce cc
// by mod (if negative) and a Flux roll, floored at 0 and never
// increased above the original cc ("CC may not be increased"), then
// scoutRiskOutcome maps the result. Returns the outcome and the
// (possibly reduced) characteristic value — applying that reduction
// back onto a persistent Character.UPP is a future caller's concern;
// this file only ever produces one detached Term (see ResolveScoutTerm),
// never a full multi-term Character.
func resolveScoutRisk(r *dice.Roller, cc ehex.Value, mod int) (RiskResult, ehex.Value) {
	if rollAgainstTarget(r, int(cc), mod) {
		return Unharmed, cc
	}

	reduced := int(cc)
	if mod < 0 {
		reduced += mod
	}

	reduced += r.Flux()

	switch {
	case reduced < 0:
		reduced = 0
	case reduced > int(cc):
		reduced = int(cc)
	}

	return scoutRiskOutcome(int(cc), reduced), ehex.Value(reduced)
}

// resolveScoutReward resolves one Reward roll against cc-mod (opposite
// sign from Risk, per Book 1 p.79's "Roll for Reward against CC+
// (opposite sign) Mods"), reporting whether the Scout makes a Discovery.
func resolveScoutReward(r *dice.Roller, cc ehex.Value, mod int) bool {
	return rollAgainstTarget(r, int(cc), -mod)
}

// scoutSkillTable is Book 1 p.79's Scout Skills table (7 columns, rows
// keyed by 1D 1-6), transcribed directly from the page image. Column 1
// (Personal) entries are characteristic boosts, not named skills at all
// — rollScoutSkill records these as a SkillLevel using the
// characteristic's own short name (Str/Dex/End/Int/Edu/Soc, matching
// characteristic.go's own abbreviations) and Kind: Personal, Level: 1
// meaning "+1," not a proficiency level — this project's data model has
// no separate "characteristic boost" record type, and Personal is
// otherwise unused, so this is the closest honest fit.
var scoutSkillTable = [7][6]string{
	{"Str", "Dex", "End", "Int", "Edu", "Soc"},
	{"Major", "Major", "Minor", "Minor", "One Trade", "One Trade"},
	{"Comms", "Language", "Computer", "Jack of all Trades", "Gunner", "Starship Skill"},
	{"Survey", "Survival", "Hostile Environment", "Animals", "Vacc Suit", "Navigation"},
	{"Diplomat", "Sensors", "Trader", "Teacher", "Fighter", "Streetwise"},
	{"Survey", "Flyer", "Language", "Starship Skill", "Engineer", "Comms"},
	{"One Art", "One Science", "Seafarer", "Athlete", "Medic", "One Trade"},
}

// rollScoutSkill rolls one column (uniform pick among all 7 — the book
// gives no mechanic for which column a Scout draws from beyond "Column
// 1 is always available," so this generator resolves the choice the
// same way it resolves every other open player pick) and row (1D),
// returning the granted skill and true, or false if the roll landed on
// an entry this generator can't resolve yet: "Major"/"Minor" (Book 1's
// own footnote — "If the character does not have a Major/Minor this
// benefit is lost" — and this codebase has no education/schooling
// generation code anywhere yet, so no character ever has a Major or
// Minor) or "One Science" (no defined science list exists yet in this
// codebase, unlike "One Art"/"One Trade," which reuse
// homeworld_generate.go's own oneArtChoices/theTradeChoices).
// "C6+1"'s own footnote (lost if C6=Caste) never applies here —
// GenerateUPP (characteristic_generate.go) only generates Human
// characters, whose C6 is always Social Standing, never Caste.
func rollScoutSkill(r *dice.Roller) (SkillLevel, bool) {
	column := r.Uniform(7) - 1
	row := r.Uniform(6) - 1
	name := scoutSkillTable[column][row]

	if column == 0 {
		return skillLevel1(name, Personal), true
	}

	switch name {
	case "Major", "Minor", "One Science":
		return SkillLevel{}, false
	case "One Art":
		return skillLevel1(rollChoice(r, oneArtChoices), Skill), true
	case "One Trade":
		return skillLevel1(rollChoice(r, theTradeChoices), Skill), true
	default:
		return skillLevel1(name, Skill), true
	}
}

// rollScoutSkills rolls count skills via rollScoutSkill. A roll that
// lands on an unresolvable entry (see rollScoutSkill) grants nothing
// rather than being rerolled — the book's own wording is that the
// benefit is "lost," not replaced — so the returned slice can be
// shorter than count.
func rollScoutSkills(r *dice.Roller, count int) []SkillLevel {
	var skills []SkillLevel

	for range count {
		if skill, ok := rollScoutSkill(r); ok {
			skills = append(skills, skill)
		}
	}

	return skills
}

// ResolveScoutTerm resolves one 4-year Scout term (Book 1 p.79) for the
// controlling characteristic ccPos — the same Position BeginScout chose
// at career entry, passed in by the caller rather than re-derived from
// upp here, so the two can't silently diverge once a later phase starts
// persisting Risk-roll characteristic reductions between terms. Begin
// happens once at career entry, not per term, so it's its own function
// (BeginScout) rather than folded in here. Courier Duty skips Risk &
// Reward entirely (Term.RiskResult stays at its zero value, Unharmed —
// no harm came to a Scout who never rolled for it) and grants 4 skills;
// Explorer Duty rolls Risk then, unless Dead, Reward too (Book 1's own
// worked example, Eneri Dinsha, rolls Reward even after failing Risk —
// a Wounded or Disabled Scout still gets a Reward chance and skills;
// only Dead stops everything) and grants 8. The Reward roll uses the
// Risk-reduced characteristic value, not the original — a Wounded or
// Disabled Scout's Reward odds reflect their now-lower characteristic,
// same as the book's own "Roll for Reward against CC+..." wording
// implies (CC having just been reduced by the preceding Risk roll).
// Term.RewardResult is "Discovery" or "None" (Term.RewardResult is a
// string field; there's no separate bool for this). Returns the
// resolved Term and whether the character survived to potentially serve
// another (RiskResult.Survived() — false only for Dead).
func ResolveScoutTerm(r *dice.Roller, upp UPP, ccPos Position) (Term, bool) {
	cc := upp.Characteristics[ccPos]

	term := Term{Length: 4, ControllingCharacteristic: ccPos, RewardResult: "None"}

	duty := rollScoutDuty(r)

	skillCount := 4
	if duty == ExplorerDuty {
		skillCount = 8

		risk, reducedCC := resolveScoutRisk(r, cc, 0)
		term.RiskResult = risk

		if risk == Dead {
			return term, false
		}

		if resolveScoutReward(r, reducedCC, 0) {
			term.RewardResult = "Discovery"
		}
	}

	term.SkillsAwarded = rollScoutSkills(r, skillCount)

	return term, term.RiskResult.Survived()
}
