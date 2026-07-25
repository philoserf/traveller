package character

import (
	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// succeedsAgainst is rollAgainstTarget's own dice-free comparison,
// split out the same way riskOutcome is so the pass/fail boundary
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

// operationsEduDM is the Armed Forces' own universal "DM +2 if Edu 10+"
// (Book 1 p.82/p.86 alike) — originally marineOperationsEduDM, which
// already had zero Marine-specific content, generalized once Soldier
// became a second real caller.
func operationsEduDM(edu int) int {
	if edu >= 10 {
		return 2
	}

	return 0
}

// rollHighestOfFour rolls 1D+dm four times against a musterOutRow-style
// table (names/mods must be the same length) and returns the row with the
// highest Mod — Book 1's own "Roll 4 times per Term for Operations;
// select the highest Mod from the four," shared by every Armed Forces
// career's own Operations roll. Originally rollMarineOperations's own
// loop body, generalized once Soldier became a second real caller of the
// identical shape.
func rollHighestOfFour(r *dice.Roller, dm int, names []string, mods []int) (string, int) {
	// bestIdx starts from the first of the 4 rolls, not a hardcoded
	// index — seeding it with an unrolled row would bias the result
	// toward that row whenever none of the real rolls beat its Mod.
	bestIdx := musterOutRow(r.D6()+dm, len(names))

	for range 3 {
		row := musterOutRow(r.D6()+dm, len(names))
		if mods[row] > mods[bestIdx] {
			bestIdx = row
		}
	}

	return names[bestIdx], mods[bestIdx]
}

// branchAutomaticSkill is Book 1's own Armed Forces branch-tied automatic
// skill, verbatim identical text on both Marine's (p.86) and Soldier's
// (p.82) own pages ("if Medical Branch=Medic-1. If Technical Branch=any
// Trade"), granted once at career start — reuses theTradeChoices/
// rollChoice (homeworld_generate.go) for Technical's own open "any Trade"
// pick, the same resolution already used for "One Trade" table entries.
// Originally marineBranchAutomaticSkill, generalized once Soldier
// confirmed the text (and logic) is Armed-Forces-wide, not
// Marine-specific.
func branchAutomaticSkill(r *dice.Roller, branch string) (SkillLevel, bool) {
	switch branch {
	case "Medical":
		return skillLevel1("Medic", Skill), true
	case "Technical":
		return skillLevel1(rollChoice(r, theTradeChoices), Skill), true
	default:
		return SkillLevel{}, false
	}
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
// success bool. career_loop.go's ResolveScoutCareer does not thread
// this Position into term 1's Controlling Characteristic — its own
// nextScoutCC, called against an empty used-set, independently
// computes the identical highestOf(upp, C1, C2, C3) value, since
// nothing has reduced upp yet at career entry. The two calls are kept
// separate rather than coupled through a shared return value because
// nextScoutCC's job (rotate through C1/C2/C3 per Book 1 p.64) only
// coincides with BeginScout's job (pick the best of C1/C2/C3 to
// attempt Begin against) on the very first term.
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

// riskOutcome maps a Risk failure's original and reduced characteristic
// values to a RiskResult, split out from resolveRisk's own dice rolls so
// the mapping itself is unit-testable against fixed fixtures — matching
// this project's existing pattern of keeping a rule's own logic
// dice-free and testable (e.g. world.techLevelModifier). Checked in
// priority order since more than one case can technically apply to the
// same reduction: reduced to 0 is Dead first and foremost, even when
// original was already 0 (loss=0 there too, which would otherwise
// misreport as Unharmed); otherwise no net loss (Flux offset the Mods)
// is Unharmed even though the roll technically failed; a 4+ loss is
// Disabled; anything less is Wounded.
//
// Book 1 p.65's universal Risk-consequence rule, shared by every career
// that rolls Risk — originally written for Scout alone (as
// scoutRiskOutcome), generalized here now that Marine is a second real
// caller with verbatim-matching consequence text of its own (p.86:
// "Reduce CC by negative Mods and Flux... If reduced by 4 or more, then
// Disabled"), the same "generalize on 2nd instance" discipline
// musterOutRow already established.
func riskOutcome(original, reduced int) RiskResult {
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

// resolveRisk resolves one Risk roll against cc+mod — Book 1 p.65's
// universal Risk mechanic, originally written for Scout alone
// (resolveScoutRisk, p.79's own Risk & Reward box) and generalized once
// Marine became a second real caller with the identical mechanic (p.86).
// On success, Unharmed. On failure: reduce cc by mod (if negative) and a
// Flux roll, floored at 0 and never increased above the original cc
// ("CC may not be increased"), then riskOutcome maps the result. Returns
// the outcome and the (possibly reduced) characteristic value — applying
// that reduction back onto a persistent Character.UPP is a caller's
// concern, not this function's.
func resolveRisk(r *dice.Roller, cc ehex.Value, mod int) (RiskResult, ehex.Value) {
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

	return riskOutcome(int(cc), reduced), ehex.Value(reduced)
}

// resolveReward resolves one Reward roll against cc-mod (opposite sign
// from Risk, per Book 1 p.65's universal "Roll for Reward against CC+
// (opposite sign) Mods" — originally resolveScoutReward, generalized
// once Marine became a second real caller), reporting whether the
// attempt succeeds (a Scout Discovery, a Marine XS Exemplary Service —
// the specific consequence is each caller's own concern) and the raw,
// unmodified roll. Scout ignores the raw roll; Marine needs it to
// consult Book 1 p.70's own Medals table ("Rew= Successful unmodified
// Reward Roll") — widening the return here, rather than duplicating this
// function's single-line body in a Marine-only variant, is the same
// "shared primitive, second real caller" treatment resolveRisk/
// riskOutcome already got in Phase T1.
func resolveReward(r *dice.Roller, cc ehex.Value, mod int) (bool, int) {
	roll := r.TwoD6()

	return succeedsAgainst(roll, int(cc), -mod), roll
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

// rollSkillFromTable rolls one column (uniform pick among all 7 — the
// book gives no mechanic for which column a career draws from beyond
// "Column 1 is always available," so this generator resolves the choice
// the same way it resolves every other open player pick) and row (1D) of
// a 7-column/1D-row skill table — the shape both scoutSkillTable and
// citizen_generate.go's own citizenTableC share, generalized here after
// Citizen became the second concrete instance of the identical logic.
// Returns the granted skill and true, or false if the roll landed on an
// entry this generator can't resolve yet: "Major"/"Minor" (Book 1's own
// footnote — "If the character does not have a Major/Minor this benefit
// is lost" — and this codebase has no education/schooling generation
// code anywhere yet, so no character ever has a Major or Minor), "One
// Science" (no defined science list exists yet in this codebase, unlike
// "One Art"/"One Trade," which reuse homeworld_generate.go's own
// oneArtChoices/theTradeChoices), or "Capital" (Noble's own
// nobleSkillTable entry — Book 1 p.85's own footnote: "Capital= World
// Knowledge (of world of highest held noble Land Grant) (value=1D)" —
// this codebase has no Land Grant tracking anywhere yet). "C6+1"'s own
// footnote (lost if C6=Caste) never applies here — GenerateUPP
// (characteristic_generate.go) only generates Human characters, whose C6
// is always Social Standing, never Caste.
func rollSkillFromTable(r *dice.Roller, table [7][6]string) (SkillLevel, bool) {
	column := r.Uniform(7) - 1
	row := r.Uniform(6) - 1
	name := table[column][row]

	if column == 0 {
		return skillLevel1(name, Personal), true
	}

	switch name {
	case "Major", "Minor", "One Science", "Capital":
		return SkillLevel{}, false
	case "One Art":
		return skillLevel1(rollChoice(r, oneArtChoices), Skill), true
	case "One Trade":
		return skillLevel1(rollChoice(r, theTradeChoices), Skill), true
	default:
		return skillLevel1(name, Skill), true
	}
}

// rollSkillsFromTable rolls count skills via rollSkillFromTable. A roll
// that lands on an unresolvable entry (see rollSkillFromTable) grants
// nothing rather than being rerolled — the book's own wording is that the
// benefit is "lost," not replaced — so the returned slice can be shorter
// than count.
func rollSkillsFromTable(r *dice.Roller, table [7][6]string, count int) []SkillLevel {
	var skills []SkillLevel

	for range count {
		if skill, ok := rollSkillFromTable(r, table); ok {
			skills = append(skills, skill)
		}
	}

	return skills
}

func rollScoutSkills(r *dice.Roller, count int) []SkillLevel {
	return rollSkillsFromTable(r, scoutSkillTable, count)
}

// ResolveScoutTerm resolves one 4-year Scout term (Book 1 p.79) for the
// controlling characteristic ccPos. Begin happens once at career entry,
// not per term, so it's its own function (BeginScout) rather than
// folded in here. Courier Duty skips Risk & Reward entirely
// (Term.RiskResult stays at its zero value, Unharmed — no harm came to
// a Scout who never rolled for it) and grants 4 skills; Explorer Duty
// rolls Risk then, unless Dead, Reward too (Book 1's own worked
// example, Eneri Dinsha, rolls Reward even after failing Risk — a
// Wounded or Disabled Scout still gets a Reward chance and skills; only
// Dead stops everything) and grants 8. The Reward roll uses the
// Risk-reduced characteristic value, not the original — a Wounded or
// Disabled Scout's Reward odds reflect their now-lower characteristic,
// same as the book's own "Roll for Reward against CC+..." wording
// implies (CC having just been reduced by the preceding Risk roll).
// Term.RewardResult is "Discovery" or "None" (Term.RewardResult is a
// string field; there's no separate bool for this).
//
// Returns the resolved Term, an updated UPP, and whether the character
// survived to potentially serve another (RiskResult.Survived() — false
// only for Dead). Per Book 1 p.65, a Risk-caused reduction is
// "permanently reduced," not scoped to the current term — the returned
// UPP carries that reduction in Characteristics[ccPos] (unchanged for
// Courier Duty, which never rolls Risk) so a multi-term caller such as
// career_loop.go's ResolveScoutCareer can thread it into the next
// term's own Risk roll and Controlling Characteristic selection. upp is
// passed by value with Characteristics a fixed-size array, so mutating
// the local copy and returning it has no aliasing effect on the
// caller's own upp.
func ResolveScoutTerm(r *dice.Roller, upp UPP, ccPos Position) (Term, UPP, bool) {
	cc := upp.Characteristics[ccPos]

	term := Term{Length: 4, ControllingCharacteristic: ccPos, RewardResult: "None"}

	duty := rollScoutDuty(r)

	skillCount := 4
	if duty == ExplorerDuty {
		skillCount = 8

		risk, reducedCC := resolveRisk(r, cc, 0)
		term.RiskResult = risk
		upp.Characteristics[ccPos] = reducedCC

		if risk == Dead {
			return term, upp, false
		}

		if ok, _ := resolveReward(r, reducedCC, 0); ok {
			term.RewardResult = "Discovery"
		}
	}

	term.SkillsAwarded = rollScoutSkills(r, skillCount)

	return term, upp, term.RiskResult.Survived()
}
