package character

import (
	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// ScholarCareerName is Scholar's own Career.Name value — exported and
// shared, matching every other career's own CareerName rationale.
const ScholarCareerName = "Scholar"

var scholarRiskRewardPositions = []Position{C1, C2, C3, C4}

// scholarRankNames is Book 1 p.76's own "Table Of Scholar Ranks" (0-6) —
// printed titles include a `<of Major>` suffix this codebase drops, since
// Major/Minor selection depends on the Education mechanic, deferred all
// session (see this slice's own plan-file Context).
var scholarRankNames = [7]string{
	"Amateur", "Lecturer", "Instructor", "Assistant Professor",
	"Associate Professor", "Professor", "Distinguished Professor",
}

// scholarSkillTable is Book 1 p.76's own "SCHOLAR SKILLS" table. Its own
// Personal row prints "Str/C2/C3/Int/C5/C6" literally in the book —
// unlike Scout's/Rogue's own tables, which spell out "Str Dex End Int
// Edu Soc" in full, Scholar's own Personal row genuinely uses the raw
// position-code notation instead. Stored verbatim, not normalized to the
// spelled-out form — resolveSkillCell's column-0 branch stores whatever
// string is here directly as the Personal skill Name, so this also
// matches the book's own printed "C2 +1" (not "Dex +1") notation
// exactly.
var scholarSkillTable = [7][6]string{
	{"Str", "C2", "C3", "Int", "C5", "C6"},
	{"Major", "Major", "Minor", "Minor", "One Trade", "One Trade"},
	{"Seafarer", "Navigation", "Hostile Environment", "Flyer", "Driver", "Vacc Suit"},
	{"Survey", "Astrogation", "Hostile Environment", "Survival", "Animals", "Bureaucrat"},
	{"Fighter", "Fighter", "Stealth", "Flyer", "Gunner", "Sensors"},
	{"Admin", "Language", "One Science", "Comms", "Starship Skill", "Bureaucrat"},
	{"Seafarer", "One Art", "One Science", "Athlete", "Medic", "One Trade"},
}

const (
	scholarSkillsPerTerm             = 4
	scholarResearchSuccessSkillBonus = 2
)

// scholarStartTier is Book 1 p.76's own Edu>=8 threshold that decides
// both whether Begin is automatic (BeginScholar) and the DM/Fame
// starting point every other Scholar rank calculation needs
// (ResolveScholarMusterOut, buildScholarCharacter) — a single source for
// a rule three call sites otherwise had to restate independently.
func scholarStartTier(edu int) int {
	if edu >= 8 {
		return 1
	}

	return 0
}

// BeginScholar resolves Book 1 p.76's own "To Begin (Edu 8+) Automatic /
// To Begin Edu or Tra" — Human-only: "Tra" (Training) is a non-Human C5
// substitute for Edu, and this codebase only ever generates Human
// characters, so that branch is unreachable and not implemented; "Edu or
// Tra" reduces to "Edu" alone here. Returns (qualified, startTier):
// Edu>=8 is automatic entry at tier 1 (Lecturer), no roll; Edu<8 rolls
// against Edu (no Mod) for entry at tier 0 (Amateur) on success.
func BeginScholar(r *dice.Roller, edu int) (bool, int) {
	if edu >= 8 {
		return true, scholarStartTier(edu)
	}

	return rollAgainstTarget(r, edu, 0), scholarStartTier(edu)
}

// scholarRankName is p.76's own "Table Of Scholar Ranks", whose printed
// titles carry the character's Major from Lecturer upward:
//
//	0 Amateur
//	1 Lecturer <of Major>
//	2 Instructor <of Major>
//	3 Assistant Professor <of Major>
//	4 Associate Professor <of Major>
//	5 Professor of <Major>
//	6 Distinguished Professor <of Major>
//
// Level 5 encloses a different span than the rest — "Professor of
// <Major>" where its neighbours read "<of Major>" — but both render the
// same way, so the suffix is applied uniformly. Amateur takes none: the
// printed table gives it no suffix, and it is the rank of a character
// with Edu 7 or less, who p.76 says "can resolve Risk and Reward, but
// cannot be Promoted".
//
// Falls back to the bare title when no Major is known, which is how this
// read before CharGen step C existed to supply one.
func scholarRankName(tier int, major string) string {
	if tier == 0 || major == "" {
		return scholarRankNames[tier]
	}

	return scholarRankNames[tier] + " of " + major
}

// rollScholarTenure is Book 1 p.76's own "Tenure Pub x3" — no separate
// Mod, unlike Promotion/Continue.
func rollScholarTenure(r *dice.Roller, pubs int) bool {
	return rollAgainstTarget(r, pubs*3, 0)
}

// scholarPublicationsTotal sums 1 per successful Publication, 2 for an
// Award-Winning one — the shared Promotion/Continue/Tenure Mod (Book 1
// p.76's own "*Mod +Pubs").
func scholarPublicationsTotal(terms []Term) int {
	total := 0

	for _, t := range terms {
		total += scholarPublicationDelta(t)
	}

	return total
}

// scholarPublicationDelta is scholarPublicationsTotal's own per-term
// scale, factored out so a not-yet-appended term's contribution can be
// added to an already-summed priorTerms total (see ResolveScholarTerm's
// own pubs calculation) without appending onto — and so risking an
// aliased write into — the caller-owned priorTerms slice.
func scholarPublicationDelta(t Term) int {
	switch {
	case t.AwardWinning:
		return 2
	case t.PublicationSucceeded:
		return 1
	default:
		return 0
	}
}

// hasTenure reports whether Tenure was ever granted — derived from
// Terms, not separately tracked, matching this codebase's existing
// "derive don't track" discipline (nobleIsExiled, citizenLifeSuccessCount).
func hasTenure(terms []Term) bool {
	for _, t := range terms {
		if t.TenureGranted {
			return true
		}
	}

	return false
}

// scholarRankTier derives the current rank tier from prior terms and the
// character's own known startTier (0 or 1 — Edu-dependent, see
// BeginScholar) — Scholar's own rank progression is a flat 0-6 track, no
// Officer/Enlisted split, so this is simpler than rankState
// (character/career_rank.go) but follows the same "derive, explicit
// tier bounds" shape.
func scholarRankTier(terms []Term, startTier int) int {
	tier := startTier

	for _, t := range terms {
		if t.Promoted {
			tier = min(tier+1, len(scholarRankNames)-1)
		}
	}

	return tier
}

// ResolveScholarTerm resolves one 4-year Scholar term (Book 1 p.76). tier
// is the rank tier entering this term; edu is the character's own Edu
// value (Promotion and Tenure are both Edu-gated). Returns the term, the
// UPP with the Risk-reduced CC persisted (mirroring ResolveScoutTerm's
// own upp.Characteristics[ccPos] = reducedCC, confirmed directly from
// its body — the next term's own Risk roll must use the reduced value),
// and the tier exiting this term.
//
// Reward (Publication) is rolled only when Research succeeded
// (riskResult == Unharmed) — a real divergence from Scout's/Marine's own
// "Reward rolled unless Dead" pattern: the box's own Reward row is
// nested under "If Research Complete, roll for Publication," with no
// equivalent step on a Wounded/Disabled/Dead term.
func ResolveScholarTerm(
	r *dice.Roller, upp UPP, ccPos Position, edu, tier int, priorTerms []Term, major string, waivers *int,
) (Term, UPP, int) {
	term := Term{Length: 4, ControllingCharacteristic: ccPos}

	cc := upp.Characteristics[ccPos]
	upp.Characteristics[ccPos] = resolveScholarResearch(r, upp, cc, &term, waivers)

	pubs := scholarPublicationsTotal(priorTerms) + scholarPublicationDelta(term)

	if tier == 3 && edu >= 10 && !hasTenure(priorTerms) {
		term.TenureGranted = scholarWaivableRoll(r, upp, waivers, rollScholarTenure(r, pubs))
	}

	// Tenure only ever needs to be won once — "Promotion beyond Scholar3
	// not possible without Tenure" gates every transition past tier 3,
	// not just the 3->4 one, so hasTenure(priorTerms) (or a same-term
	// grant) permanently unblocks every later Promotion too, not only
	// this one. A code-review pass caught an earlier version of this
	// gate checking tier == 3 specifically, which silently capped every
	// tenured Scholar at Associate Professor (tier 4) forever.
	if edu >= 8 && (tier < 3 || term.TenureGranted || hasTenure(priorTerms)) {
		if scholarWaivableRoll(r, upp, waivers, rollAgainstTarget(r, int(upp.Characteristics[C4]), pubs)) {
			term.Promoted = true
			tier = min(tier+1, len(scholarRankNames)-1)
		}
	}

	term.Rank = scholarRankName(tier, major)

	skillCount := scholarSkillsPerTerm
	if term.Promoted {
		skillCount++
	}

	if term.RiskResult == Unharmed {
		skillCount += scholarResearchSuccessSkillBonus
	}

	term.SkillsAwarded = rollSkillsFromTable(r, scholarSkillTable, skillCount)

	return term, upp, tier
}

// Book 1 p.76's own account of what a Scholar studies:
//
//	"The Scholar's Major. A Scholar has an area of interest and expertise
//	called his Major, accompanied by a companion area of knowledge called
//	his Minor. Every Scholar has a Major and a Minor. If no degree (and
//	an associated Major and Minor) then select any Skill or Knowledge
//	from the Skills List."
//
// So a Scholar is the one career that always has a Major, whether or not
// CharGen step C gave him a degree. A graduate brings the subjects he
// declared at College or University; anyone else declares them on
// entering the career.
//
// This is why the Major/Minor cells of scholarSkillTable always resolve
// for a Scholar, where in the other twelve careers they are lost to
// Book 1's own "If the character does not have a Major/Minor this
// benefit is lost".
func scholarMajorMinor(r *dice.Roller, edu Education) (string, string) {
	if edu.Major != "" && edu.Minor != "" {
		return edu.Major, edu.Minor
	}

	// "select any Skill or Knowledge from the Skills List" — resolved
	// against p.60's own matrix, which is the enumerated Skill and
	// Knowledge list this codebase has. A free selection rather than an
	// Education grant, so it is not restricted to any school's column.
	pool := allEducationSubjects()

	major := edu.Major
	if major == "" {
		major = pool[r.Uniform(len(pool))-1]
	}

	minor := edu.Minor
	if minor == "" {
		minor = pickSubjectExcluding(r, pool, major)
	}

	return major, minor
}

// scholarWaivableRoll runs one of p.76's own waivable Scholar checks:
//
//	"Waivers. An adverse die roll or decision (in Position, Promotion,
//	Research, Publication, Tenure, or Continue) may be waived. Check Soc
//	(2D); Mod minus previous waivers (successful or not)."
//
// Six named events and no others — a Scholar cannot waive, say, an
// injury from a failed Research roll, only the failure itself. The
// Waiver machinery is Education's (education.go's own tryWaiver), which
// p.59 says applies "to Schools and Education (and to the Scholar
// career, but not other careers)"; the two statements of the rule agree
// exactly, down to the Mod counting attempts rather than successes.
func scholarWaivableRoll(r *dice.Roller, upp UPP, waivers *int, succeeded bool) bool {
	if succeeded {
		return true
	}

	return tryWaiver(r, upp, waivers)
}

// resolveScholarResearch runs p.76's own Research and Publication rolls
// — Scholar's names for Risk and Reward — and returns the Controlling
// Characteristic they leave behind.
//
// Both are on p.76's waivable list. A waived Research failure spares the
// characteristic reduction the failure would otherwise inflict: the
// adverse roll is set aside, not merely its consequences. A waived
// Publication rejection is an ordinary success rather than a
// distinguished one, since Award-Winning needs the roll itself to have
// come in 4 under the characteristic.
func resolveScholarResearch(r *dice.Roller, upp UPP, cc ehex.Value, term *Term, waivers *int) ehex.Value {
	riskResult, reducedCC := resolveRisk(r, cc, 0)
	if riskResult != Unharmed && tryWaiver(r, upp, waivers) {
		riskResult, reducedCC = Unharmed, cc
	}

	term.RiskResult = riskResult

	if riskResult != Unharmed {
		return reducedCC
	}

	succeeded, roll := resolveReward(r, cc, 0)
	term.PublicationSucceeded = scholarWaivableRoll(r, upp, waivers, succeeded)

	if succeeded && roll <= int(cc)-4 {
		term.AwardWinning = true
	}

	return reducedCC
}
