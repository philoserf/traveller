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
	// Operations are every assignment received this term — Book 1's own
	// four 1D rolls per Term. Assignment above records only the
	// highest-Mod one; all four are kept because p.65 restricts Term
	// skills to "a column of the Skills table corresponding to an
	// Operations result received in the Term". Spacer, Soldier and
	// Marine only.
	Operations           []string
	Commissioned         bool // Armed Forces only: Commission succeeded this term
	RiskResult           RiskResult
	RewardResult         string
	SkillsAwarded        []SkillLevel
	Promoted             bool // Armed Forces only: Officer/Enlisted Promotion succeeded this term
	CitizenLifeSucceeded bool // Citizen only
	// CitizenLifeRoll/CitizenLifeTarget are the raw 2D6 and the
	// characteristic it was rolled against (Book 1 p.78: "roll 2D for the
	// Controlling Characteristic or less"). Retained so a rendered term
	// can show why it succeeded or failed rather than only that it did —
	// the outcome bool alone can't explain itself. Citizen only.
	CitizenLifeRoll   int
	CitizenLifeTarget int
	// CitizenLifeGrant names the extra Job/Hobby skill a successful
	// Citizen Life term grants, so a reader can tell it apart from the
	// four unconditional Table C skills sitting beside it in
	// SkillsAwarded (where it is also recorded, since it is a real
	// skill). Empty when the term failed or the grant resolved to
	// nothing. Citizen only.
	CitizenLifeGrant string
	// CitizenLifeGrantIsJob distinguishes which of Book 1 p.78's two
	// alternating tracks CitizenLifeGrant came from. The alternation is
	// driven by the count of prior successful terms, so it can't be
	// recovered from a single Term after the fact. Citizen only.
	CitizenLifeGrantIsJob bool
	NobleAction           string // "Return" or "Intrigue" — Noble only
	NobleSucceeded        bool   // Noble only
	Elevated              bool   // Noble only: Elevation succeeded this term
	Scheme                string // Rogue only: this term's own Scheme flavor name
	Imprisoned            bool   // Rogue only: Risk failed this term
	PrisonYears           int    // Rogue only: sentence length if Imprisoned (0-4)
	// ServedYears is the sentence carried into this term from the
	// previous one's failed Scheme — Book 1 p.84 imposes prison "at the
	// start of the next Term", so it is served a term later than the
	// PrisonYears that earned it. Distinct from Imprisoned, which marks
	// the term whose own Scheme failed. Rogue only.
	ServedYears          int
	RewardSucceeded      bool   // Rogue only: this term's own Reward roll outcome
	SchemePayoff         int    // Rogue only: Cr payoff this term (0 if Reward failed; halved if Imprisoned)
	SchemeShipShare      bool   // Rogue only: Reward succeeded on a Ship-Share-valued Scheme
	PublicationSucceeded bool   // Scholar only: this term's own Publication roll outcome
	AwardWinning         bool   // Scholar only: Publication beat CC by 4+, counts as two Publications
	TenureGranted        bool   // Scholar only: Tenure roll succeeded this term
	FameAfterTerm        int    // Entertainer only: Fame value after this term's own Flux roll
	TalentAfterTerm      int    // Entertainer only: Talent value after this term (Fame increases grant +1)
	FameIncreased        bool   // Entertainer only: this term's own Flux roll was positive
	UndercoverCareer     string // Agent only: which career's own skill table this term's own skill was drawn from
	// UndercoverAssignment is Book 1 p.83's own rolled Undercover
	// Assignment as printed — the table's service label plus the rank
	// title its C column selected ("Navy Officer Commander"), or the
	// label alone for the rows that print no title. Agent only.
	UndercoverAssignment string
	// Masterpiece is the structured record of what this term created —
	// its QREBS allocation, Master Points and creation age. RewardResult
	// still carries the display string beside it. Craftsman only, and nil
	// for a term whose creation attempt failed. See masterpiece.go.
	Masterpiece *Masterpiece
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
	Name       string
	HasRank    bool
	JobSkill   string // Citizen only
	HobbySkill string // Citizen only
	Specialty  string // Entertainer only: Artist/Actor/Author/Dancer/Musician/Chef
	// Major and Minor are Scholar only. Book 1 p.76: "A Scholar has an
	// area of interest and expertise called his Major, accompanied by a
	// companion area of knowledge called his Minor. Every Scholar has a
	// Major and a Minor." A Scholar who arrives with a degree brings its
	// subjects; one who does not declares them on entering the career, so
	// they are recorded here rather than on Education.
	Major, Minor string
	Terms        []Term
	MusteringOut MusteringOut
}
