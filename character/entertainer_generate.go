package character

import (
	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// EntertainerCareerName is Entertainer's own Career.Name value —
// exported and shared, matching every other career's own CareerName
// rationale.
const EntertainerCareerName = "Entertainer"

// entertainerSpecialties is Book 1 p.77's own "Select A Specialty" list —
// a genuine open choice with no book-given tiebreaker, resolved via the
// same uniform-random-pick convention as Scout's Duty/Rogue's CC.
var entertainerSpecialties = [6]string{"Artist", "Actor", "Author", "Dancer", "Musician", "Chef"}

// entertainerBeginPositions maps each Specialty to its own two Begin
// characteristics (Book 1 p.77's own "Begin <Specialty> X or Y" lines) —
// "or" is read as "the higher of," reusing highestOf directly.
var entertainerBeginPositions = map[string][2]Position{
	"Artist":   {C3, C4}, // C3 or Int
	"Actor":    {C2, C3},
	"Author":   {C4, C5}, // Int or C5
	"Dancer":   {C2, C3},
	"Musician": {C2, C3},
	"Chef":     {C2, C4}, // C2 or Int
}

// entertainerSkillTable is Book 1 p.77's own "ENTERTAINER SKILLS" table.
// Its own Personal row prints "Str/C2/C3/Int/C5/C6" literally, the same
// raw position-code notation already confirmed for Scholar's own table
// (career_generate.go's resolveSkillCell stores it verbatim).
var entertainerSkillTable = [7][6]string{
	{"Str", "C2", "C3", "Int", "C5", "C6"},
	{"Major", "Major", "Minor", "Minor", "One Trade", "One Trade"},
	{"Zero-G", "Vacc Suit", "Pilot", "Astrogator", "Sensors", "Starship Skill"},
	{"Survey", "Survival", "Hostile Environment", "Animals", "Bureaucrat", "Navigation"},
	{"Broker", "Trader", "Advocate", "Liaison", "Diplomat", "Bureaucrat"},
	{"Broker", "One Art", "Language", "Admin", "One Art", "Bureaucrat"},
	{"One Art", "One Trade", "Athlete", "Medic", "One Trade", "One Trade"},
}

const (
	entertainerSkillsPerTerm          = 4
	entertainerFameIncreaseSkillBonus = 2
)

// rollEntertainerSpecialty resolves "Select A Specialty".
func rollEntertainerSpecialty(r *dice.Roller) string {
	return entertainerSpecialties[r.Uniform(6)-1]
}

// rollEntertainerFameTalent is Book 1 p.77's own "roll initial Fame and
// Talent (with one 2D roll; they are equal)" — a single roll producing
// both starting values, "Before Begin" (i.e. independent of whether
// BeginEntertainer's own roll later succeeds — see this slice's own
// plan-file Context for why Character.Fame is still set even on a
// never-qualified attempt).
// entertainerOptionalFluxRolls is the count of "F*" entries on Book 1
// p.77's own Fame progression — "Fame +F +F* +F*", footnoted "F= Flux.
// F*= Optional Flux" — so two per term beyond the mandatory one.
const entertainerOptionalFluxRolls = 2

// twoD6Expectation is the mean of 2D6 — the Fame a Comeback resets to,
// in expectation.
const twoD6Expectation = 7

// entertainerTakesComeback decides whether a fading Entertainer stages
// Book 1 p.77's own Comeback: "Reset Fame to 2D; Talent is unchanged.
// Comeback is possible any number of times."
//
// The book gives no criterion, but unlike the optional Flux there is an
// unambiguously better answer here, so this picks it rather than
// flipping a coin — the same reasoning BeginScout's own doc comment
// gives for choosing the highest characteristic. A reset trades current
// Fame for 2D, worth taking exactly when Fame has fallen below what 2D
// pays on average. Talent, which drives Risk and Reward, is untouched
// either way, so there is nothing to lose by resetting a low Fame.
//
// It also matters more than it looks: continueEntertainer rolls against
// Fame, so a collapsed Fame ends the career. Comeback is how the rules
// let a washed-up performer keep working.
func entertainerTakesComeback(fame int) bool {
	return fame < twoD6Expectation
}

// rollEntertainerFameChange resolves one term's Fame movement — the
// mandatory Flux plus however many of the two optional ones this
// character takes.
//
// Whether to take an optional Flux is a player decision the book states
// without giving any criterion for, so each is resolved as an
// independent coin flip. That follows this package's established
// treatment of an open choice with no book-given mechanic (rollScoutDuty
// and the Art/Trade picks) rather than BeginScout's, which picks the
// best option — here there isn't an unambiguously best one. Flux
// averages zero, so an extra roll doesn't raise expected Fame; it widens
// the spread, which helps a character whose mandatory Flux went badly
// and hurts one whose went well. Choosing on that basis would be this
// codebase inventing a strategy the book leaves to the table.
func rollEntertainerFameChange(r *dice.Roller) int {
	delta := r.Flux()

	for range entertainerOptionalFluxRolls {
		if r.Uniform(2) == 1 {
			delta += r.Flux()
		}
	}

	return delta
}

func rollEntertainerFameTalent(r *dice.Roller) int {
	return r.TwoD6()
}

// BeginEntertainer resolves "Begin <Specialty> X or Y" — target is the
// higher of the Specialty's own two characteristics (highestOf).
func BeginEntertainer(r *dice.Roller, upp UPP, specialty string) bool {
	pair := entertainerBeginPositions[specialty]
	best := highestOf(upp, pair[0], pair[1])

	return rollAgainstTarget(r, int(upp.Characteristics[best]), 0)
}

// ResolveEntertainerTerm resolves one 4-year Entertainer term (Book 1
// p.77). fame/talent are the values entering this term. Returns the
// term and the fame/talent values exiting it.
//
// Fame/Talent evolve first (Book 1's own "At the start of each Term...");
// Risk & Reward then rolls against the (possibly just-increased) Talent.
// Only the mandatory first Flux roll is implemented — see this slice's
// own plan-file Context for why the optional second/third rolls are
// deferred.
//
// Risk & Reward target Talent, not a UPP characteristic — resolveRisk/
// resolveReward never touch UPP.Characteristics themselves (confirmed
// against ResolveScholarTerm's own body last slice), so they work
// unmodified against a plain local int cast through ehex.Value. A Talent
// reduced to 0 reports RiskResult == Dead via the same universal
// mechanic every other career uses — here that means the entertainer's
// Talent is completely spent, ending the career (resolveCareerLoop's own
// stop condition already treats any Dead as "stop"), not physical death.
func ResolveEntertainerTerm(r *dice.Roller, fame, talent int) (Term, int, int) {
	term := Term{Length: 4}

	fameDelta := rollEntertainerFameChange(r)
	fame += fameDelta
	term.FameIncreased = fameDelta > 0

	if term.FameIncreased {
		talent++
	}

	term.FameAfterTerm = fame

	// Talent can only grow by +1 per Fame increase across at most 14
	// terms, so it never realistically approaches ehex.Max — clamped
	// anyway (matching ApplyMusteringOut's own precedent,
	// muster_out_apply.go) so the int->uint8 cast is provably safe, not
	// just practically safe.
	clampedTalent := min(talent, int(ehex.Max))

	//nolint:gosec // bounded by the min(...) clamp above, gosec can't see that
	riskResult, reducedTalent := resolveRisk(r, ehex.Value(clampedTalent), 0)
	term.RiskResult = riskResult
	talent = int(reducedTalent)
	term.TalentAfterTerm = talent

	// Reward is rolled whenever Risk isn't Dead (Scout's/Marine's own
	// default), not Scholar's own "only on Unharmed" restriction —
	// Entertainer's page has no dedicated Risk & Reward sidebar
	// describing a Scholar-style override, so the universal default
	// applies. Reward success has no career-specific consequence
	// described on this page, so the outcome is recorded via the
	// existing generic RewardResult string, not a new typed field.
	if riskResult != Dead {
		// Reward uses the term-start Talent; the Risk reduction applies
		// to later terms, just as it does for a UPP-based CC.
		//nolint:gosec // bounded by the min(...) clamp above, gosec can't see that
		if ok, _ := resolveReward(r, ehex.Value(clampedTalent), 0); ok {
			term.RewardResult = "Success"
		} else {
			term.RewardResult = "None"
		}
	}

	skillCount := entertainerSkillsPerTerm
	if term.FameIncreased {
		skillCount += entertainerFameIncreaseSkillBonus
	}

	term.SkillsAwarded = rollSkillsFromTable(r, entertainerSkillTable, skillCount)

	return term, fame, talent
}
