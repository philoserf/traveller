package character

import (
	"fmt"

	"github.com/philoserf/traveller/dice"
)

// AgentCareerName is Agent's own Career.Name value — exported and
// shared, matching every other career's own CareerName rationale.
const AgentCareerName = "Agent"

var agentRiskRewardPositions = []Position{C1, C2, C3, C4}

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
	assignment, title := rollAgentUndercoverAssignment(r)
	term.UndercoverCareer = assignment.Career
	term.UndercoverAssignment = title

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
	// the surrounding count-based grants above).
	term.SkillsAwarded = append(term.SkillsAwarded, rollAgentUndercoverSkill(r, assignment.Career))

	return term, upp
}
