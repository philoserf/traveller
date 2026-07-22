// Package character models Traveller5 player and non-player characters:
// characteristics, careers, skills, and combat-relevant stats.
package character

// Character is a full player or non-player character.
type Character struct {
	Name           string
	Species        string
	GeneticProfile string // 6-char code identifying which trait each UPP position represents, e.g. "SDEIES"
	UPP            UPP
	Homeworld      string
	Birthworld     string
	Birthdate      string
	Age            int
	LifeStage      int
	Careers        []Career
	Skills         []SkillLevel
	Rank           string
	Medals         []string
	Commendations  []string
	WoundBadges    int
	Fame           int
	Cash           int
	Equipment      []string
	Notes          string
}
