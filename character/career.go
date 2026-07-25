package character

// RiskResult is the outcome of a term's Risk roll.
type RiskResult int

// RiskResult values, from best to worst outcome.
const (
	Unharmed RiskResult = iota
	Wounded
	Disabled
	Dead
)

// Survived reports whether this RiskResult left the character alive.
func (r RiskResult) Survived() bool {
	return r != Dead
}

// Term is a single term (typically 4 years) served within a career.
type Term struct {
	Length                    int
	ControllingCharacteristic Position
	Branch                    string   // Armed Forces only
	Assignment                string   // Armed Forces only
	Medals                    []string // Armed Forces only: medal codes earned this term (XS/MCUF/MCG/SEH)
	Rank                      string   // Armed Forces only: rank held after this term (e.g. "M3 Sergeant")
	Commissioned              bool     // Armed Forces only: Commission succeeded this term
	RiskResult                RiskResult
	RewardResult              string
	SkillsAwarded             []SkillLevel
	Promoted                  bool   // Armed Forces only: Officer/Enlisted Promotion succeeded this term
	CitizenLifeSucceeded      bool   // Citizen only
	NobleAction               string // "Return" or "Intrigue" — Noble only
	NobleSucceeded            bool   // Noble only
	Elevated                  bool   // Noble only: Elevation succeeded this term
	Scheme                    string // Rogue only: this term's own Scheme flavor name
	Imprisoned                bool   // Rogue only: Risk failed this term
	PrisonYears               int    // Rogue only: sentence length if Imprisoned (0-4)
	RewardSucceeded           bool   // Rogue only: this term's own Reward roll outcome
	SchemePayoff              int    // Rogue only: Cr payoff this term (0 if Reward failed; halved if Imprisoned)
	SchemeShipShare           bool   // Rogue only: Reward succeeded on a Ship-Share-valued Scheme
	PublicationSucceeded      bool   // Scholar only: this term's own Publication roll outcome
	AwardWinning              bool   // Scholar only: Publication beat CC by 4+, counts as two Publications
	TenureGranted             bool   // Scholar only: Tenure roll succeeded this term
	FameAfterTerm             int    // Entertainer only: Fame value after this term's own Flux roll
	TalentAfterTerm           int    // Entertainer only: Talent value after this term (Fame increases grant +1)
	FameIncreased             bool   // Entertainer only: this term's own Flux roll was positive
	UndercoverCareer          string // Agent only: which career's own skill table this term's own skill was drawn from
	Perfect                   bool   // Craftsman only: this term's own Masterpiece reached 55+ Master Points
}

// MusteringOut is the benefits package awarded when a character leaves a career.
type MusteringOut struct {
	Automatics    []string
	Benefits      []string
	Money         []string
	Entitlements  []string
	Pension       int
	RetirementPay int
}

// Career is a full career history within a single career track.
type Career struct {
	Name         string
	HasRank      bool
	JobSkill     string // Citizen only
	HobbySkill   string // Citizen only
	Specialty    string // Entertainer only: Artist/Actor/Author/Dancer/Musician/Chef
	Terms        []Term
	MusteringOut MusteringOut
}
