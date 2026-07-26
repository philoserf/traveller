package character

import (
	"fmt"

	"github.com/philoserf/traveller/dice"
)

// AgentCareerName is Agent's own Career.Name value — exported and
// shared, matching every other career's own CareerName rationale.
const AgentCareerName = "Agent"

var agentRiskRewardPositions = []Position{C1, C2, C3, C4}

// agentUndercoverSkillTables maps each of this codebase's own nine
// already-implemented career names to its own skill table — the
// deliberately simplified stand-in for Book 1 p.83's own 18-row x
// 3-column "AGENT UNDERCOVER ASSIGNMENT" table (a real Undercover
// Assignment names a specific rank title within one of eleven other
// careers per cell, resolved via a nested three-die A/B/C reroll
// mechanic — this codebase instead uniformly picks one of its own
// already-implemented career skill tables, deferring the rank-title
// flavor text, the three-die mechanic, and Citizen's/Scout's own
// special-cased rows, since the full table doesn't actually run a real
// career underneath — it's narrative dressing for which skill table to
// draw from).
var agentUndercoverSkillTables = map[string][7][6]string{
	"Scout":       scoutSkillTable,
	"Marine":      marineSkillTable,
	"Soldier":     soldierSkillTable,
	"Spacer":      spacerSkillTable,
	"Rogue":       rogueSkillTable,
	"Scholar":     scholarSkillTable,
	"Entertainer": entertainerSkillTable,
	"Merchant":    merchantSkillTable,
	"Noble":       nobleSkillTable,
}

var agentUndercoverCareerNames = []string{
	"Scout", "Marine", "Soldier", "Spacer", "Rogue",
	"Scholar", "Entertainer", "Merchant", "Noble",
}

// agentSkillTable is Book 1 p.83's own "AGENT SKILLS" table. Its own
// "Any Knowledge" cell (column 6, row 1) is handled by
// resolveSkillCell's own unresolvable-cell case (career_generate.go) —
// no enumerable Knowledge list exists anywhere in this codebase.
var agentSkillTable = [7][6]string{
	{"Str", "C2", "C3", "Int", "C5", "C6"},
	{"Major", "Major", "Minor", "Minor", "One Trade", "One Trade"},
	{"Zero-G", "Vacc Suit", "Pilot", "Starship Skill", "Gunner", "Sensors"},
	{"Survey", "Survival", "Hostile Environment", "Animals", "Bureaucrat", "Navigation"},
	{"Fighter", "Soldier Skill", "Flyer", "Stealth", "Gunner", "Streetwise"},
	{"Any Knowledge", "Admin", "Language", "Starship Skill", "Comms", "One Trade"},
	{"One Art", "One Science", "Athlete", "Medic", "Seafarer", "One Trade"},
}

const (
	agentSkillsPerTerm               = 2
	agentSuccessfulMissionSkillBonus = 4
)

// BeginAgent is Book 1 p.83's own "To Begin C3" — a single ordinary
// roll against End, no cascading tiers like Merchant.
func BeginAgent(r *dice.Roller, c3 int) bool {
	return rollAgainstTarget(r, c3, 0)
}

// rollAgentUndercoverCareer resolves "Roll for Undercover Assignment" —
// simplified to a uniform pick among this codebase's own nine
// already-implemented career skill tables (see this slice's own
// plan-file Context). "Select (not Roll) one skill" (the box's own
// wording for the skill draw itself) is resolved via
// rollSkillFromTable's own dice-driven uniform pick across the whole
// table — the same "open player choice resolved via uniform random
// pick through the dice roller" convention already established for
// Scout's Duty and Entertainer's Specialty, not a literal contradiction
// of "not Roll" (the book means "not looked up on a weighted table,"
// not "not resolved via any RNG at all" — this codebase has no
// interactive player to make the pick).
func rollAgentUndercoverCareer(r *dice.Roller) string {
	return agentUndercoverCareerNames[r.Uniform(len(agentUndercoverCareerNames))-1]
}

// agentCommendationCount sums how many prior terms earned a
// Commendation — the shared Fame/Mustering-Out-DM count. Mirrors
// merchantRewardCount's own generic RewardResult check.
func agentCommendationCount(terms []Term) int {
	n := 0

	for _, t := range terms {
		if t.RewardResult != "" && t.RewardResult != "None" {
			n++
		}
	}

	return n
}

// ResolveAgentTerm resolves one 4-year Agent term (Book 1 p.83). The
// Undercover Assignment is rolled first, since the Commendation's own
// label needs term.UndercoverCareer already set.
func ResolveAgentTerm(r *dice.Roller, upp UPP, ccPos Position) (Term, UPP) {
	term := Term{Length: 4, ControllingCharacteristic: ccPos}
	term.UndercoverCareer = rollAgentUndercoverCareer(r)

	cc := upp.Characteristics[ccPos]
	riskResult, reducedCC := resolveRisk(r, cc, 0)
	term.RiskResult = riskResult
	upp.Characteristics[ccPos] = reducedCC

	if riskResult == Dead {
		return term, upp
	}

	// Commendation is Book 1 p.91's own "N = C-R". Both the Reward target
	// and C use the term-start, pre-Risk CC; the reduced value persists
	// only for later terms.
	if rewardOK, rewardRoll := resolveReward(r, cc, 0); rewardOK {
		n := int(cc) - rewardRoll
		term.RewardResult = fmt.Sprintf("%s Commendation-%d", term.UndercoverCareer, n)
	} else {
		term.RewardResult = "None"
	}

	skillCount := agentSkillsPerTerm
	if term.RewardResult != "" && term.RewardResult != "None" {
		skillCount += agentSuccessfulMissionSkillBonus
	}

	term.SkillsAwarded = rollSkillsFromTable(r, agentSkillTable, skillCount)

	// The Undercover skill — exactly one, unconditionally, per Book 1's
	// own "Select ... one skill" (singular, no "lost" language, unlike
	// the surrounding count-based grants above) — rerolled until a
	// resolvable cell is drawn, rather than reusing rollSkillsFromTable's
	// own "unresolvable = lost" convention, since this specific grant is
	// guaranteed by the book's own text.
	var undercoverSkill SkillLevel

	for {
		skill, ok := rollSkillFromTable(r, agentUndercoverSkillTables[term.UndercoverCareer])
		if ok {
			undercoverSkill = skill

			break
		}
	}

	term.SkillsAwarded = append(term.SkillsAwarded, undercoverSkill)

	return term, upp
}
