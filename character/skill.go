package character

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
