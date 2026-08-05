package character

import "github.com/philoserf/traveller/dice"

// scoutRiskRewardPositions is Scout's own Risk & Reward Controlling
// Characteristic set (Book 1 p.79: "Risk & Reward C1 C2 C3") — C4 Int is
// deliberately excluded even though the generic per-career summary table
// elsewhere (p.64's "5 Scout C1 C2 C3 C4" row) lists a fourth entry; Scout's
// own dedicated box (p.79) is the authoritative, career-specific source and
// reserves C4 Int exclusively for the Continue check (p.79's separate
// "Continue Int" line).
var scoutRiskRewardPositions = []Position{C1, C2, C3}

// nextCC picks the Controlling Characteristic for one term from positions
// and records it in usedThisCycle, implementing Book 1 p.65's generic
// rule: the player picks any not-yet-used Characteristic from the career's
// own set; it "cannot be used again until all of the others in the sequence
// have been used." A career's CC therefore rotates through its own set
// rather than staying fixed for the whole career — that fixed-for-career
// behavior is Rogue's own explicitly-marked "*Special Case" (p.83: a
// Rogue's Controlling Characteristic "is then used throughout his career,
// not just in the current Term"), a carve-out that would be redundant if
// every career already worked that way. Which still-available
// characteristic to pick is, like BeginScout's own Begin-target choice, an
// open choice Book 1 leaves to the player; resolved here via highestOf
// (same rationale as BeginScout's own doc comment: a genuinely better
// option exists, so a uniform random pick would be a worse default).
// usedThisCycle resets once every position in positions has been used,
// starting a fresh cycle. Shared by Scout's own C1 C2 C3 set (nextScoutCC,
// below) and Citizen's C1 C2 C3 C4 set (citizen_loop.go) — the two
// concrete instances that justify generalizing this out of nextScoutCC's
// own original, Scout-only body.
func nextCC(upp UPP, positions []Position, usedThisCycle map[Position]bool) Position {
	if len(usedThisCycle) == len(positions) {
		clear(usedThisCycle)
	}

	var available []Position

	for _, pos := range positions {
		if !usedThisCycle[pos] {
			available = append(available, pos)
		}
	}

	pos := highestOf(upp, available...)
	usedThisCycle[pos] = true

	return pos
}

func nextScoutCC(upp UPP, usedThisCycle map[Position]bool) Position {
	return nextCC(upp, scoutRiskRewardPositions, usedThisCycle)
}

// continueScoutOutcome is continueScout's own dice-free decision, split out
// the same way succeedsAgainst/riskOutcome are so the boundary is
// directly testable against a fixed roll instead of a real 2D6 draw. roll==2
// always succeeds (Book 1's "Mandatory Continue"); otherwise roll<=intChar.
func continueScoutOutcome(roll, intChar int) bool {
	return mandatoryContinueOutcome(roll, intChar)
}

// continueScout resolves Book 1's end-of-Term Continue check (p.65-66: "the
// Character must successfully roll (2D) to Continue (or less) in the
// career... If the Continue roll is 2 exactly, the character is required to
// Continue") against Scout's own p.79 Continue target, Int (C4) — unlike
// Agent's "Continue Str*" or Rogue's "Continue CC*" (each footnoted "*Mod
// +Terms"), Scout's own box has no asterisk and no +Terms Mod: the target is
// bare Int every term. Note: GenerateUPP only ever produces characteristics
// in [2,12] (a 2D6 roll) and Scout's own Risk & Reward set
// (scoutRiskRewardPositions) never includes C4, so Int can never fall below 2
// for any character this codebase produces — the roll==2 mandatory-continue
// branch is implemented because it's the literal rule, but is currently
// unreachable as an actual behavior change (2<=Int already covers it).
func continueScout(r *dice.Roller, upp UPP) bool {
	return continueScoutOutcome(r.TwoD6(), int(upp.Characteristics[C4]))
}

// maxCareerTerms is where Book 1 p.89's own description of a life runs
// out: the traditional human lifespan (74 years) minus the traditional
// adventuring start (Young Adult, 18), in 4-year terms — (74-18)/4 = 14.
//
// This was originally written as a placeholder — "a defensive
// engineering cap, not a Book 1 rule... since Aging (which would
// otherwise end this naturally) isn't modeled yet." Aging is modeled
// now, and applied chronologically between terms (aging.go's own
// agingSimulation), so that reasoning has been tested rather than
// assumed — and it does not hold. Three independent things in p.89
// agree on 74 rather than pointing past it:
//
//   - the "Stages of Life" table's last row is Retirement, whose upper
//     bound is 71 (LifeStageForAge's own doc comment); nothing beyond it
//     is defined, so there is no Aging Check the book describes for a
//     75-year-old;
//   - 74 is named as the traditional lifespan; and
//   - this constant, derived independently from that lifespan, lands on
//     exactly the same boundary.
//
// Removing the cap was tried and measured before being rejected. Careers
// do terminate on their own without it — nothing reached even a 200-term
// guard across 21,000 characters — but only because LifeStageForAge
// clamps every age past 71 to Life Stage 9 and keeps rolling forever,
// which is this codebase extrapolating into a region the table does not
// cover. The resulting distribution (p99 of 32 terms, a maximum of 57 —
// age 246) is not "somewhat longer than traditional"; it is a lifespan
// the rules never bound at all. Honoring the end of the table is the
// more faithful reading than inventing checks past it.
//
// So a career reaching 14 terms has reached the end of Book 1's own
// account of a human life, not an arbitrary engineering limit. It is
// still a rare outcome per career, but a real one, not a bug.
const maxCareerTerms = 14

// resolveCareerLoop is the shared multi-term Career-resolution loop
// shape both ResolveScoutCareer and ResolveMarineCareer follow: rotate
// the Controlling Characteristic via nextCC, resolve one term via the
// caller-supplied resolveTerm, thread the returned UPP forward into the
// next iteration, and stop when the character didn't survive
// (RiskResult.Survived(), false only on Dead), was left Disabled (p.65:
// mandatory regardless of the Continue roll), or failed the
// caller-supplied continueCareer check — capped at maxTerms iterations.
// Extracted once Marine became a second real caller of this identical
// shape (the same "generalize on 2nd instance" discipline already
// applied to musterOutRow and resolveRisk/resolveReward/riskOutcome) —
// a defect in this control flow (e.g. the missing-Disabled-check class
// of defect an early Scout draft had) now only needs fixing once, not
// once per career.
//
// maxTerms is a per-call budget, not always maxCareerTerms: every
// existing single-career caller passes maxCareerTerms itself (see each
// ResolveXCareer's own thin wrapper around its resolveXCareerWithBudget)
// for unchanged behavior, while character/career_chain.go's multi-career
// orchestrator can pass a tighter budget derived from an -age target —
// once maxTerms iterations have run, the loop simply never attempts
// another one (no Continue roll, no new term), rather than cutting an
// already-resolved term short.
// aging is threaded in rather than created here so one simulation spans
// a whole life: a multi-career chain passes the same one through every
// segment, and its own checkpoint after each term is what stops a career
// when Aging kills the character mid-service.
// afterContinue is the hook Book 1 p.61's Command College needs: a
// check that runs after a successful Continue and can still end the
// career ("A character who fails Command College may not Continue in the
// service"). Nil for every career that has no such school — only the
// three Armed Forces careers do.
//
// Placed after the Continue roll rather than before it because p.61
// gates attendance on continuing — "provided he successfully Continues"
// — so a character who was leaving anyway never attends and never rolls
// for it.
type afterContinue func(r *dice.Roller, upp UPP, terms []Term) bool

// beforeTerm is p.59's own Later Education hook: "At the beginning of
// any term, the character may apply for any Educational Institution or
// Training, and if accepted substitutes that process for the entire
// term." Checked once per iteration, before nextCC/resolveTerm — if it
// reports true, its returned Term and UPP substitute for what
// resolveTerm would have produced, and nextCC's own rotation is not
// advanced, since no Controlling Characteristic was used. Nil for every
// career that doesn't yet offer it.
type beforeTerm func(r *dice.Roller, upp UPP) (Term, UPP, bool)

func resolveCareerLoop(
	r *dice.Roller,
	upp UPP,
	positions []Position,
	resolveTerm func(r *dice.Roller, upp UPP, ccPos Position) (Term, UPP),
	continueCareer func(r *dice.Roller, upp UPP) bool,
	maxTerms int,
	aging *agingSimulation,
	afterContinueCheck afterContinue,
	beforeTermCheck beforeTerm,
) ([]Term, UPP) {
	var terms []Term

	usedThisCycle := make(map[Position]bool, len(positions))

	for range maxTerms {
		// Already dead — from an Aging checkpoint in an earlier career of
		// the same chain, since one simulation spans them all. Checked
		// before resolving anything, not after: the post-term check below
		// only stops the career that did the killing, and would still let
		// the next one in the chain serve a full term first.
		if !aging.alive() {
			break
		}

		var (
			term       Term
			updatedUPP UPP
			elected    bool
		)

		if beforeTermCheck != nil {
			term, updatedUPP, elected = beforeTermCheck(r, upp)
		}

		if !elected {
			ccPos := nextCC(upp, positions, usedThisCycle)
			term, updatedUPP = resolveTerm(r, upp, ccPos)
		}

		upp = applyPersonalAwards(updatedUPP, term.SkillsAwarded)

		terms = append(terms, term)

		// A term that killed the character outright (Book 1 p.69) ends
		// things before any Aging check — the dead don't grow older.
		// Disabled does still age: that career is over, the person isn't.
		if !term.RiskResult.Survived() {
			aging.recordFatalTerm()

			break
		}

		upp = aging.advanceTerm(r, upp)
		if !aging.alive() {
			break
		}

		if term.RiskResult == Disabled || len(terms) == maxTerms {
			break
		}

		if !continueCareer(r, upp) {
			break
		}

		if afterContinueCheck != nil && !afterContinueCheck(r, upp, terms) {
			break
		}
	}

	return terms, upp
}

// ResolveScoutCareer resolves a full multi-term Scout career (Book 1 p.79)
// via resolveCareerLoop. BeginScout's one-time "To Begin" check runs
// first; if it fails, this returns Career{Name: "Scout"} with a nil Terms
// slice and upp unchanged — a real, legitimate "never qualified" outcome
// per Book 1's own "If both Begin and Retry fail, this career may not be
// used," not an error.
//
// Returns the final UPP alongside the Career — not just the Career — so a
// caller building a full Character can persist any permanent reduction
// (RiskResult Wounded or Disabled) onto Character.UPP; returning only the
// Career would silently lose it one call frame up, reintroducing the exact
// p.65 persistence gap this function exists to fix.
//
// BeginScout's own returned Position is not threaded into the first term's
// ccPos — nextCC's first-ever call (against an empty usedThisCycle)
// independently computes the identical value, both being
// highestOf(upp, C1, C2, C3) over the same untouched characteristic set — so
// term 1's CC naturally coincides with BeginScout's own pick.
//
// Deferred, same treatment as Aging/Muster-Out: Book 1 p.79's own "reduce
// San= -1 for each TWO Terms served" Sanity rule — meaningless to apply
// since GenerateUPP never generates Sanity in the first place (p.57:
// "created only as needed").
func ResolveScoutCareer(r *dice.Roller, upp UPP) (Career, UPP) {
	return resolveScoutCareerWithBudget(r, upp, maxCareerTerms, &agingSimulation{})
}

// resolveScoutCareerWithBudget is ResolveScoutCareer's own body, with
// the resolveCareerLoop term cap threaded as a parameter instead of the
// implicit maxCareerTerms constant — see resolveCareerLoop's own doc
// comment for why. character/career_chain.go's multi-career orchestrator
// calls this directly with an -age-derived budget; ResolveScoutCareer
// itself is unchanged for every other caller.
func resolveScoutCareerWithBudget(r *dice.Roller, upp UPP, maxTerms int, aging *agingSimulation) (Career, UPP) {
	career := Career{Name: "Scout"}

	// Each attempt's own year is charged before the next is rolled, so
	// the Retry sees whatever the intervening Aging checkpoint did — and
	// isn't rolled at all if that checkpoint killed the character.
	if _, began := BeginScout(r, upp); !began {
		upp = aging.chargeFailedAttempt(r, upp)
		if !aging.alive() {
			return career, upp
		}

		if !RetryScout(r, upp) {
			return career, aging.chargeFailedAttempt(r, upp)
		}
	}

	terms, finalUPP := resolveCareerLoop(r, upp, scoutRiskRewardPositions,
		func(r *dice.Roller, upp UPP, ccPos Position) (Term, UPP) {
			term, updatedUPP, _ := ResolveScoutTerm(r, upp, ccPos)

			return term, updatedUPP
		},
		continueScout,
		maxTerms,
		aging,
		nil,
		nil,
	)
	career.Terms = terms

	return career, finalUPP
}
