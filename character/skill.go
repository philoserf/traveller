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
