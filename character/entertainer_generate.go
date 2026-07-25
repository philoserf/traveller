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

	fameDelta := r.Flux()
	fame += fameDelta
	term.FameIncreased = fameDelta > 0

	if term.FameIncreased {
		talent++
	}

	term.FameAfterTerm = fame
	term.TalentAfterTerm = talent

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
		clampedTalent = min(talent, int(ehex.Max))

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
