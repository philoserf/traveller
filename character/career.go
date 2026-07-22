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

// Term is a single term (typically 4 years) served within a career.
type Term struct {
	Length                    int
	ControllingCharacteristic Position
	Branch                    string // Armed Forces only
	Assignment                string // Armed Forces only
	Rank                      string
	Commissioned              bool
	RiskResult                RiskResult
	RewardResult              string
	SkillsAwarded             []SkillLevel
	Survived                  bool
	Promoted                  bool
}

// MusteringOut is the benefits package awarded when a character leaves a career.
type MusteringOut struct {
	Automatics    []string
	Benefits      []string
	Entitlements  []string
	Pension       int
	RetirementPay int
}

// Career is a full career history within a single career track.
type Career struct {
	Name         string
	HasRank      bool
	Terms        []Term
	MusteringOut MusteringOut
}
