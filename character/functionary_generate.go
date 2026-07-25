package character

import (
	"fmt"

	"github.com/philoserf/traveller/dice"
)

// FunctionaryCareerName is Functionary's own Career.Name value —
// exported and shared, matching every other career's own CareerName
// rationale.
const FunctionaryCareerName = "Functionary"

// functionaryRiskRewardPositions is Office Politics' own Controlling
// Characteristic set (Book 1 p.87: "Office Politics C2 C3 C4 C5").
var functionaryRiskRewardPositions = []Position{C2, C3, C4, C5}

// functionarySkillTable is Book 1 p.87's own "FUNCTIONARY SKILLS" table.
// Its own "Any Skill" cell (column 4, row 4) is resolved by
// resolveSkillCell's own shared switch (career_generate.go), which
// draws from citizenTableE per the book's own footnote ("from Citizen
// Life Skills and Knowledges"). "C6+1"'s own "lost if C6=Caste" footnote
// never applies — GenerateUPP only produces Human characters (C6 always
// Social Standing), the same already-documented reasoning as every
// other career's own C6 column.
var functionarySkillTable = [7][6]string{
	{"Str", "C2", "C3", "Int", "C5", "C6"},
	{"Major", "Major", "Minor", "Minor", "One Trade", "One Trade"},
	{"High-G", "Vacc Suit", "Driver", "Flyer", "Navigation", "Seafarer"},
	{"One Trade", "One Art", "One Science", "Any Skill", "Bureaucrat", "Leader"},
	{"Advocate", "Broker", "Trader", "Teacher", "One Trade", "Driver"},
	{"Advocate", "Comms", "Language", "Admin", "Bureaucrat", "Comms"},
	{"One Art", "One Science", "Athlete", "Designer", "Seafarer", "One Trade"},
}

const functionarySkillsPerTerm = 4

// functionaryRankNames is Book 1 p.87's own "Functionary Ranks" table,
// F0-F8. F6's own name varies by preceding career (see
// functionaryF6TitleByPrecedingCareer); F7's own "Nth UnderSecretary"
// needs its own N=1D roll (functionaryOrdinals), resolved once, the
// first (and only) time tier reaches 7 — promotion is strictly +1 per
// success, so a tier can never reach 7 more than once in a career.
var functionaryRankNames = [9]string{
	"Clerk", "Supervisor", "Senior Supervisor", "Manager", "Senior Manager",
	"Assistant Director", "Director", "Nth UnderSecretary", "Secretary",
}

// functionaryOrdinals renders the N in "Nth UnderSecretary" for a 1D
// roll (1-6) — Book 1's own worked example is "3rd UnderSecretary".
var functionaryOrdinals = [6]string{"1st", "2nd", "3rd", "4th", "5th", "6th"}

// functionaryTierAutoSkills is the Rank table's own "Auto Skill"
// column — granted once, the first time a tier is reached (F0 at
// career start, F2/F4 on the promotion that reaches them), the same
// "one-level increase, or level-1 if not already held" shape documented
// for Merchant's own merchantRankAutoSkill. F1/F3/F5-F8 grant nothing
// (absent from the map).
var functionaryTierAutoSkills = map[int]string{0: "Bureaucrat", 2: "Admin", 4: "Bureaucrat"}

// functionaryF6TitleByPrecedingCareer is Book 1 p.87's own four named
// F6 titles, keyed by which prior career the Functionary is associated
// with — the only four Book 1 actually names. Every other preceding
// career (Scout, Marine, Soldier, Spacer, Agent, Citizen, or none) falls
// back to the generic "Director" already in functionaryRankNames, a
// documented simplification: the book gives no title for those. Keyed
// by careerChainRegistry's own lowercase names (segmentContext.PrecedingCareer
// is always one of those), not the exported, capitalized CareerName
// constants.
var functionaryF6TitleByPrecedingCareer = map[string]string{
	"scholar":     "College President",
	"entertainer": "Association Director",
	"merchant":    "Starport Warden",
	"rogue":       "Bank President",
}

// BeginFunctionary implements Book 1 p.87's own "To Begin Total Terms
// x3" — the roll target is three times the character's total terms
// served in all prior careers (termsServedSoFar), not a characteristic.
// With 0 prior terms the target is 0, which a 2D6 roll (minimum 2) can
// never meet — "Functionary is never a first career" (p.87) falls out
// of this formula for free; validateCareerChain (career_chain.go) still
// enforces it explicitly too, per this codebase's own "implement the
// literal rule, not just its currently-observable effect" discipline.
func BeginFunctionary(r *dice.Roller, termsServedSoFar int) bool {
	return rollAgainstTarget(r, 3*termsServedSoFar, 0)
}

// functionaryRankName formats "F<tier> <name>", substituting the
// preceding-career-specific title at F6 and the rolled ordinal at F7.
func functionaryRankName(tier int, precedingCareer string, nthUnderSecretary int) string {
	name := functionaryRankNames[tier]

	switch tier {
	case 6:
		if title, ok := functionaryF6TitleByPrecedingCareer[precedingCareer]; ok {
			name = title
		}
	case 7:
		name = functionaryOrdinals[nthUnderSecretary-1] + " UnderSecretary"
	}

	return fmt.Sprintf("F%d %s", tier, name)
}

// ResolveFunctionaryTerm resolves one 4-year Functionary term (Book 1
// p.87) via Office Politics — a named Risk/Reward variant, not the
// universal p.65 mechanic: no Mods, no characteristic reduction of any
// kind, and Reward is only rolled when Risk succeeds (unlike Scout/
// Marine/Agent, where Reward is rolled whenever Risk isn't Dead).
//
// Risk failure is recorded as RiskResult Disabled — a deliberate reuse
// of Risk vocabulary for "career ends, cannot Continue" (Book 1's own
// literal words), not a physical injury: resolveCareerLoop's own stop
// check already treats Disabled as an unconditional, mandatory stop
// regardless of the continueCareer callback's answer, which is exactly
// Office Politics' own "Risk Failure: ... may not Continue" — reused
// for what it already means, the same class of reuse Entertainer's own
// Talent-exhaustion "Dead" already established this session. Callers
// must not call scoutWoundBadges on a Functionary career (WoundBadges
// is always 0 — a career-ending setback, not a wound) and must not
// render the literal word "Disabled" for it (render/character.go's own
// functionaryTermLabel overrides this, mirroring
// entertainerTermLabel's own "Talent Exhausted" override).
//
// priorTier/nthUnderSecretary are threaded through the closure the same
// way Scholar's own tier is (resolveCareerLoop builds its own local
// Terms slice; career.Terms isn't available until the whole loop
// returns). Skills are never lost on a Risk-failure term (Book 1 says
// the career ends, not that the term's own benefit is lost) — the F0
// starting Auto Skill is applied once by resolveFunctionaryCareerWithBudget
// after the loop completes (Marine's own Branch-skill precedent), since
// F0 is never reached via promotion the way F2/F4 are.
func ResolveFunctionaryTerm(
	r *dice.Roller,
	upp UPP,
	ccPos Position,
	priorTier int,
	nthUnderSecretary int,
	precedingCareer string,
) (Term, int, int) {
	term := Term{Length: 4, ControllingCharacteristic: ccPos}
	tier := priorTier

	cc := int(upp.Characteristics[ccPos])

	if !rollAgainstTarget(r, cc, 0) {
		term.RiskResult = Disabled
	} else {
		term.RiskResult = Unharmed

		if rollAgainstTarget(r, cc, 0) {
			term.Promoted = true
			tier = min(tier+1, len(functionaryRankNames)-1)

			if tier == 7 {
				nthUnderSecretary = r.D6()
			}

			term.RewardResult = "Promoted to " + functionaryRankName(tier, precedingCareer, nthUnderSecretary)
		} else {
			term.RewardResult = "None"
		}
	}

	term.Rank = functionaryRankName(tier, precedingCareer, nthUnderSecretary)

	skillCount := functionarySkillsPerTerm
	if term.Promoted {
		skillCount++

		if skillName, ok := functionaryTierAutoSkills[tier]; ok {
			term.SkillsAwarded = append(term.SkillsAwarded, skillLevel1(skillName, Skill))
		}
	}

	term.SkillsAwarded = append(term.SkillsAwarded, rollSkillsFromTable(r, functionarySkillTable, skillCount)...)

	return term, tier, nthUnderSecretary
}
