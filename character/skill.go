package character

// applyPersonalAwards applies the Personal column's characteristic
// improvements to the live UPP. The award remains on its Term as career
// history, but is excluded from Character.Skills.
func applyPersonalAwards(upp UPP, awards []SkillLevel) UPP {
	for _, award := range awards {
		if award.Kind != Personal {
			continue
		}

		position, ok := musterOutCharacteristicNames[award.Name]
		if !ok {
			continue
		}

		upp.Characteristics[position] = awardCharacteristic(upp.Characteristics[position], award.Level)
	}

	return upp
}

// SkillKind categorizes an acquired skill entry.
type SkillKind int

// SkillKind values. Skill, Personal, and Intuition come from T5's closed
// master list; Knowledge and Talent are open-ended/advisory categories.
const (
	Skill SkillKind = iota
	Knowledge
	Talent
	Personal
	Intuition
)

// SkillLevel is a single acquired skill, knowledge, or talent and its level.
// Level 0 ("default skill") is implicit and commonly omitted in notation
// (e.g. "Pilot-4").
type SkillLevel struct {
	Name  string
	Level int
	Kind  SkillKind
}

// skillLevel1 builds a freshly-granted, level-1 SkillLevel — the shape
// every generation-time skill grant across this package takes (a
// homeworld skill, a career skill, a characteristic boost), so callers
// don't each repeat the same three-field literal.
func skillLevel1(name string, kind SkillKind) SkillLevel {
	return SkillLevel{Name: name, Level: 1, Kind: kind}
}

// aggregateSkills merges every grant sharing a (Name, Kind) into one
// summed entry, in first-grant order. Every skill grant in this
// codebase is already a flat "+1 unit" (skillLevel1's own Level: 1),
// so summing is exactly Book 1's own "further grants add levels" rule
// (p.65-66) — not a new one, just finally applied at final assembly
// time rather than left as separate level-1 records forever. A skill
// with no repeat grants passes through unchanged (same Name/Level/Kind,
// same position). Confirmed directly against the two existing
// "automatic skill" functions whose own doc comments already describe
// this exact rule (merchantRankAutoSkill, branchAutomaticSkill): both
// just call skillLevel1 unconditionally today, with no "already held?"
// check — because without this function, it never mattered. Callers
// don't need to change: summing a flat +1 onto an existing entry here
// already produces the "one-level increase if held" behavior those
// comments describe.
func aggregateSkills(skills []SkillLevel) []SkillLevel {
	type key struct {
		name string
		kind SkillKind
	}

	var order []key

	totals := make(map[key]SkillLevel, len(skills))

	for _, s := range skills {
		k := key{s.Name, s.Kind}

		if existing, ok := totals[k]; ok {
			existing.Level += s.Level
			totals[k] = existing
		} else {
			order = append(order, k)
			totals[k] = s
		}
	}

	merged := make([]SkillLevel, len(order))
	for i, k := range order {
		merged[i] = totals[k]
	}

	return merged
}
