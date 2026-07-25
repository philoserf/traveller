package character

import "github.com/philoserf/traveller/dice"

// RogueCareerName is Rogue's own Career.Name value — exported and
// shared, matching every other career's own CareerName rationale.
const RogueCareerName = "Rogue"

// rogueSchemeCareerNames/rogueSchemeValues are Book 1 p.84's own "ROGUE
// SCHEMES" table, Flux-keyed (index 0 = Flux -6, index 12 = Flux +6).
// Values are literal strings ("CrN,NNN" or "one Ship Share") — reuses
// musterOutCashAmount (character/muster_out_apply.go) to distinguish
// them, since it already returns (0, false) for anything not
// "Cr"-prefixed, exactly the right behavior for the two Ship-Share rows.
var rogueSchemeCareerNames = [13]string{
	"Craftsman", "Scholar", "Entertainer", "Citizen", "Scout", "Merchant",
	"Spacer", "Soldier", "Agent", "Rogue", "Noble", "Marine", "Functionary",
}

var rogueSchemeValues = [13]string{
	"Cr200,000", "Cr100,000", "Cr300,000", "Cr50,000", "one Ship Share",
	"one Ship Share", "Cr100,000", "Cr50,000", "Cr100,000", "Cr100,000",
	"Cr500,000", "Cr50,000", "Cr100,000",
}

// rogueSkillTable is Book 1 p.84's own "C ROGUE SKILLS" table. "JOT"
// (column 5, row 4 in the printed table, cross-checked against the
// page's own .txt OCR extraction to confirm it isn't a misread) is
// normalized here to "Jack of all Trades" — the same skill Scout's own
// table spells out in full; storing the book's own abbreviation
// literally would silently create a second, unrelated skill name
// instead of the one this codebase already grants elsewhere.
var rogueSkillTable = [7][6]string{
	{"Str", "Dex", "End", "Int", "Edu", "Soc"},
	{"One Science", "Major", "Minor", "One Art", "One Trade", "Gambler"},
	{"Driver", "Flyer", "Hostile Environment", "High-G", "Vacc Suit", "Navigation"},
	{"Starship Skill", "Pilot", "Engineer", "Zero-G", "Vacc Suit", "Astrogator"},
	{"Trader", "Broker", "Computer", "Jack of all Trades", "Teacher", "Fighter"},
	{"Advocate", "Counsellor", "Language", "Leader", "Streetwise", "Comms"},
	{"One Art", "One Science", "Athlete", "Soldier Skill", "Starship Skill", "One Trade"},
}

// rogueSkillsPerTerm/rogueSuccessfulSchemeSkillBonus/roguePrisonSkillsPerTerm
// are Book 1 p.84's own "B SKILL ELIGIBILITY" box. "Failed Scheme: 1" is
// deliberately not modeled as a separate bonus — see this slice's own
// plan-file Context for why a failed Risk roll uses the Prison-columns-
// only skill roll unconditionally instead.
const (
	rogueSkillsPerTerm              = 2
	rogueSuccessfulSchemeSkillBonus = 4
	roguePrisonSkillsPerTerm        = 2
)

// rollRogueCC resolves "A Rogue selects one Controlling Characteristic
// (C1 C2 C3 C4 C5 C6)" — a genuine open choice with no book-given
// tiebreaker, resolved via the same uniform-random-choice convention
// already established for Scout's own Duty pick and every Armed Forces
// career's own Branch pick.
func rollRogueCC(r *dice.Roller) Position {
	return Position(r.Uniform(6) - 1)
}

// rogueSucceeds is Book 1 p.84's own universal Rogue resolution
// primitive: "But, 12 is always automatic failure" overrides the
// standard succeedsAgainst comparison regardless of target/Mods — the
// mirror image of continueMarineOutcome's own "natural 2 always
// succeeds," but specific to Rogue's own Begin/Risk/Reward/Continue
// checks, not a universal rule for every career.
func rogueSucceeds(r *dice.Roller, target, mod int) bool {
	roll := r.TwoD6()
	if roll == 12 {
		return false
	}

	return succeedsAgainst(roll, target, mod)
}

// rogueSucceedsRaw is rogueSucceeds's own twin, additionally returning
// the raw roll — Reward's own Payoff formula needs R (the raw Reward die
// roll), the same "widen for a second real need" precedent already
// applied to resolveReward in Phase U.
func rogueSucceedsRaw(r *dice.Roller, target, mod int) (bool, int) {
	roll := r.TwoD6()
	if roll == 12 {
		return false, roll
	}

	return succeedsAgainst(roll, target, mod), roll
}

// BeginRogue reports Book 1 p.84's own "To Begin CC" — roll 2D <= the
// player's own selected Controlling Characteristic, no Mod. No Retry,
// matching every other career researched this session with this box
// shape.
func BeginRogue(r *dice.Roller, cc int) bool {
	return rogueSucceeds(r, cc, 0)
}

// rollRogueScheme resolves the Flux-keyed "ROGUE SCHEMES" table. Flux
// ranges roughly -5..+5 (dice.Roller's own Flux, a D6-D6 difference) but
// the table only defines -6..+6 rows; clamped defensively even though
// -6/+6 themselves are already at the edge of Flux's own natural range.
func rollRogueScheme(r *dice.Roller) (string, string) {
	flux := r.Flux()

	switch {
	case flux < -6:
		flux = -6
	case flux > 6:
		flux = 6
	}

	i := flux + 6

	return rogueSchemeCareerNames[i], rogueSchemeValues[i]
}

// ResolveRogueTerm resolves one 4-year Rogue term (Book 1 p.84). cc is
// the character's own fixed Controlling Characteristic value (selected
// once for the whole career, see ResolveRogueCareer); mod is the
// combined Risk-direction Mod ("+Terms" only in this codebase — see this
// slice's own plan-file Context for why Bravery/Caution's own "sum of
// negative Mods" contribution is always 0 here, matching every other
// career).
//
// Reward is always rolled, regardless of Risk's own outcome — Rogue has
// no Dead state to skip it, unlike Marine/Soldier/Spacer. Risk failure
// means Prison (PrisonYears, 0-4, from Flux alone per the Context above),
// Fame consequences handled by the caller (buildRogueCharacter), and a
// halved Payoff if Reward also succeeded. Skill eligibility drops to the
// Prison-columns-only roll unconditionally on Risk failure, per this
// slice's own documented simplification of "In Prison"/"Failed Scheme".
func ResolveRogueTerm(r *dice.Roller, cc, mod int) Term {
	schemeName, schemeValue := rollRogueScheme(r)

	term := Term{
		Length: 4,
		Scheme: schemeName,
	}

	if !rogueSucceeds(r, cc, mod) {
		term.Imprisoned = true
		term.PrisonYears = min(4, max(0, r.Flux()))
	}

	rewardSucceeded, rewardRoll := rogueSucceedsRaw(r, cc, -mod)
	term.RewardSucceeded = rewardSucceeded

	if rewardSucceeded {
		if cashValue, ok := musterOutCashAmount(schemeValue); ok {
			payoff := cashValue * (1 + cc - rewardRoll + (-mod))
			if term.Imprisoned {
				payoff /= 2
			}

			term.SchemePayoff = payoff
		} else {
			term.SchemeShipShare = true
		}
	}

	if term.Imprisoned {
		term.SkillsAwarded = rollRogueSkillsFromTable(r, roguePrisonSkillsPerTerm)
	} else {
		term.SkillsAwarded = rollSkillsFromTable(r, rogueSkillTable, rogueSkillsPerTerm+rogueSuccessfulSchemeSkillBonus)
	}

	return term
}

// rollRogueSkillFromTable/rollRogueSkillsFromTable are Book 1 p.84's own
// "In Prison: Prison Skills from the Rogue Skills table column 1 or 2
// only" — reuses the shared resolveSkillCell (career_generate.go).
func rollRogueSkillFromTable(r *dice.Roller) (SkillLevel, bool) {
	return resolveSkillCell(r, rogueSkillTable, r.Uniform(2)-1, r.Uniform(6)-1)
}

func rollRogueSkillsFromTable(r *dice.Roller, count int) []SkillLevel {
	var skills []SkillLevel

	for range count {
		if skill, ok := rollRogueSkillFromTable(r); ok {
			skills = append(skills, skill)
		}
	}

	return skills
}
