package character

import (
	"errors"
	"fmt"
	"slices"

	"github.com/philoserf/traveller/dice"
)

// careerSegment is the result of resolving one career attempt within a
// multi-career chain: everything buildRiskCareerCharacter and its
// per-career bespoke equivalents (buildMerchantCharacter,
// buildScholarCharacter, buildRogueCharacter,
// buildEntertainerCharacter, buildCitizenCharacter) already compute,
// stopped short of the final Character assembly — Age/LifeStage/
// Birthdate/Notes come from a single finalizeAging pass over the whole
// chain's total terms served, not per segment.
type careerSegment struct {
	Career      Career
	UPP         UPP // boosted by this segment's own Mustering Out
	Survived    bool
	Fame        int
	Cash        int
	WoundBadges int
	Skills      []SkillLevel
	Medals      []string
	Equipment   []string // Craftsman only: Masterpieces created this segment
}

func segmentEndsCareerResolution(seg careerSegment) bool {
	if len(seg.Career.Terms) == 0 {
		return false
	}

	last := seg.Career.Terms[len(seg.Career.Terms)-1]

	return !last.RiskResult.Survived() || last.RiskResult == Disabled
}

// segmentContext carries the things a handful of adapters need that
// most ignore: the running terms-served total and the immediately-
// preceding chain entry's own career name (Functionary's own Begin
// target and F6 Rank-title association), and the character's own
// aggregated skills accumulated before this segment is attempted
// (Craftsman's own Begin prerequisite and "New Trade, not already
// held" skill cell). Every adapter takes this parameter; only the one
// that actually needs a given field reads it — the same low-risk,
// mechanical threading pattern the -age target already used for
// maxTerms.
// Aging is the one field every segment reads rather than only the odd
// one out: it's a single simulation for the whole chain, so a checkpoint
// reached during a later career still counts the terms served in earlier
// ones (see agingSimulation's own doc comment). It is deliberately a
// pointer — each segment mutates the shared simulation as its terms
// advance, and the chain reads the accumulated result afterward.
type segmentContext struct {
	PrecedingCareer  string
	TermsServedSoFar int
	SkillsSoFar      []SkillLevel
	Aging            *agingSimulation
}

// aging returns the chain-wide simulation, or a fresh throwaway one when
// a caller resolving a single segment in isolation (every direct
// segment-resolver call in the tests) supplied none. Segment resolvers
// go through this rather than reading ctx.Aging directly so a zero-value
// segmentContext stays usable instead of panicking on a nil pointer.
func (c segmentContext) aging() *agingSimulation {
	if c.Aging == nil {
		return &agingSimulation{}
	}

	return c.Aging
}

type careerSegmentResolver func(r *dice.Roller, upp UPP, maxTerms int, ctx segmentContext) careerSegment

// careerChainRegistry is every career resolvable as a link in a
// multi-career chain.
var careerChainRegistry = map[string]careerSegmentResolver{
	"scout":       resolveScoutSegment,
	"marine":      resolveMarineSegment,
	"soldier":     resolveSoldierSegment,
	"spacer":      resolveSpacerSegment,
	"rogue":       resolveRogueSegment,
	"scholar":     resolveScholarSegment,
	"entertainer": resolveEntertainerSegment,
	"merchant":    resolveMerchantSegment,
	"agent":       resolveAgentSegment,
	"citizen":     resolveCitizenSegment,
	"functionary": resolveFunctionarySegment,
	"craftsman":   resolveCraftsmanSegment,
	"noble":       resolveNobleSegment,
}

// resolveRiskCareerSegment mirrors buildRiskCareerCharacter's own body
// (character_generate.go) at the segment level, stopping short of
// finalizeAging — the shared shape behind Scout/Marine/Soldier/Spacer/
// Agent's own adapters below, the same generalization
// buildRiskCareerCharacter itself already applies to their full
// Character-assembly functions.
func resolveRiskCareerSegment(
	r *dice.Roller,
	upp UPP,
	maxTerms int,
	ctx segmentContext,
	resolveCareer func(r *dice.Roller, upp UPP, maxTerms int, aging *agingSimulation) (Career, UPP),
	resolveMusterOut func(r *dice.Roller, career Career) MusteringOut,
	careerFame func(career Career) int,
) careerSegment {
	aging := ctx.aging()

	career, careerUPP := resolveCareer(r, upp, maxTerms, aging)

	if aging.alive() {
		career.MusteringOut = resolveMusterOut(r, career)
	}

	boostedUPP, bonuses := ApplyMusteringOut(career.MusteringOut, careerUPP)

	ok := len(career.Terms) > 0 && career.Terms[len(career.Terms)-1].RiskResult != Dead

	fame := bonuses.Fame
	if ok {
		fame += careerFame(career)
	}

	return careerSegment{
		Career: career, UPP: boostedUPP, Survived: ok,
		Fame: fame, Cash: bonuses.Cash, WoundBadges: scoutWoundBadges(career),
		Skills: allSkillsFromTerms(career.Terms), Medals: allMedalsFromTerms(career.Terms),
	}
}

func resolveScoutSegment(r *dice.Roller, upp UPP, maxTerms int, ctx segmentContext) careerSegment {
	return resolveRiskCareerSegment(
		r,
		upp,
		maxTerms,
		ctx,
		resolveScoutCareerWithBudget,
		ResolveScoutMusterOut,
		scoutDiscoveryFame,
	)
}

func resolveMarineSegment(r *dice.Roller, upp UPP, maxTerms int, ctx segmentContext) careerSegment {
	return resolveRiskCareerSegment(
		r, upp, maxTerms, ctx, resolveMarineCareerWithBudget, ResolveMarineMusterOut, marineCareerFame)
}

func resolveSoldierSegment(r *dice.Roller, upp UPP, maxTerms int, ctx segmentContext) careerSegment {
	return resolveRiskCareerSegment(
		r, upp, maxTerms, ctx, resolveSoldierCareerWithBudget, ResolveSoldierMusterOut, soldierCareerFame)
}

func resolveSpacerSegment(r *dice.Roller, upp UPP, maxTerms int, ctx segmentContext) careerSegment {
	return resolveRiskCareerSegment(
		r, upp, maxTerms, ctx, resolveSpacerCareerWithBudget, ResolveSpacerMusterOut, spacerCareerFame)
}

func resolveAgentSegment(r *dice.Roller, upp UPP, maxTerms int, ctx segmentContext) careerSegment {
	return resolveRiskCareerSegment(
		r,
		upp,
		maxTerms,
		ctx,
		resolveAgentCareerWithBudget,
		ResolveAgentMusterOut,
		agentCareerFame,
	)
}

// resolveRogueSegment mirrors buildRogueCharacter's own body
// (rogue_character_generate.go), stopping short of finalizeAging. Rogue
// has no Risk-driven characteristic reduction; Personal awards can
// still improve UPP. There is no death mechanic, so Survived is always true.
// Fame/Cash share rogueTermsFameCash with buildRogueCharacter.
func resolveRogueSegment(r *dice.Roller, upp UPP, maxTerms int, ctx segmentContext) careerSegment {
	aging := ctx.aging()

	career, careerUPP := resolveRogueCareerAndUPPWithBudget(r, upp, maxTerms, aging)
	if aging.alive() {
		career.MusteringOut = ResolveRogueMusterOut(r, career)
	}

	boostedUPP, bonuses := ApplyMusteringOut(career.MusteringOut, careerUPP)

	ok := len(career.Terms) > 0

	cash := bonuses.Cash
	fame := bonuses.Fame

	if ok {
		termFame, termCash := rogueTermsFameCash(career.Terms)
		fame += termFame
		cash += termCash
	}

	return careerSegment{
		Career: career, UPP: boostedUPP, Survived: ok,
		Fame: fame, Cash: cash, Skills: allSkillsFromTerms(career.Terms),
	}
}

// resolveScholarSegment mirrors buildScholarCharacter's own body
// (scholar_character_generate.go), stopping short of finalizeAging.
func resolveScholarSegment(r *dice.Roller, upp UPP, maxTerms int, ctx segmentContext) careerSegment {
	aging := ctx.aging()

	career, careerUPP := resolveScholarCareerWithBudget(r, upp, maxTerms, aging)
	if aging.alive() {
		career.MusteringOut = ResolveScholarMusterOut(r, career, careerUPP)
	}

	boostedUPP, bonuses := ApplyMusteringOut(career.MusteringOut, careerUPP)

	ok := len(career.Terms) > 0 && career.Terms[len(career.Terms)-1].RiskResult != Dead

	fame := bonuses.Fame

	if ok {
		fame += scholarSegmentFame(careerUPP, career.Terms)
	}

	return careerSegment{
		Career: career, UPP: boostedUPP, Survived: ok,
		Fame: fame, Cash: bonuses.Cash, WoundBadges: scoutWoundBadges(career),
		Skills: allSkillsFromTerms(career.Terms), Medals: allMedalsFromTerms(career.Terms),
	}
}

// resolveEntertainerSegment mirrors buildEntertainerCharacter's own body
// (entertainer_character_generate.go), stopping short of finalizeAging.
// Entertainer's Risk/Reward never touches UPP.Characteristics (Fame/
// Talent are separate scalars), though Personal awards improve UPP; its
// "Dead" means Talent exhausted, not physical death, so Survived is
// always true and WoundBadges is always 0 (a Talent setback, not a
// physical wound).
func resolveEntertainerSegment(r *dice.Roller, upp UPP, maxTerms int, ctx segmentContext) careerSegment {
	aging := ctx.aging()

	career, fame, careerUPP := resolveEntertainerCareerAndUPPWithBudget(r, upp, maxTerms, aging)
	if aging.alive() {
		career.MusteringOut = ResolveEntertainerMusterOut(r, career, fame)
	}

	boostedUPP, bonuses := ApplyMusteringOut(career.MusteringOut, careerUPP)

	ok := len(career.Terms) > 0

	return careerSegment{
		Career: career, UPP: boostedUPP, Survived: ok,
		Fame: fame + bonuses.Fame, Cash: bonuses.Cash,
		Skills: allSkillsFromTerms(career.Terms),
	}
}

// resolveMerchantSegment mirrors buildMerchantCharacter's own body
// (merchant_character_generate.go), stopping short of finalizeAging.
// buildMerchantCharacter's own comment notes BeginMerchant never fails,
// so len(career.Terms) > 0 used to be unconditional — that stops being
// true once maxTerms can be 0 (an exhausted -age budget), so this
// checks explicitly rather than indexing career.Terms unconditionally.
func resolveMerchantSegment(r *dice.Roller, upp UPP, maxTerms int, ctx segmentContext) careerSegment {
	aging := ctx.aging()

	career, careerUPP, isOfficer, tier := resolveMerchantCareerWithBudget(r, upp, maxTerms, aging)
	if aging.alive() {
		career.MusteringOut = ResolveMerchantMusterOut(r, career, isOfficer, tier)
	}

	boostedUPP, bonuses := ApplyMusteringOut(career.MusteringOut, careerUPP)

	ok := len(career.Terms) > 0 && career.Terms[len(career.Terms)-1].RiskResult != Dead

	fame := bonuses.Fame
	if ok {
		fame += merchantCareerFame(tier)
	}

	return careerSegment{
		Career: career, UPP: boostedUPP, Survived: ok,
		Fame: fame, Cash: bonuses.Cash, WoundBadges: scoutWoundBadges(career),
		Skills: allSkillsFromTerms(career.Terms),
	}
}

// resolveCitizenSegment mirrors buildCitizenCharacter's own body
// (citizen_character_generate.go), stopping short of finalizeAging.
// Citizen Life can't fail Career Resolution at all, so Survived is
// always true.
func resolveCitizenSegment(r *dice.Roller, upp UPP, maxTerms int, ctx segmentContext) careerSegment {
	aging := ctx.aging()

	career, careerUPP := resolveCitizenCareerAndUPPWithBudget(r, upp, maxTerms, aging)
	if aging.alive() {
		career.MusteringOut = ResolveCitizenMusterOut(r, career)
	}

	boostedUPP, bonuses := ApplyMusteringOut(career.MusteringOut, careerUPP)

	return careerSegment{
		Career: career, UPP: boostedUPP, Survived: true,
		Fame: bonuses.Fame, Cash: bonuses.Cash,
		Skills: allSkillsFromTerms(career.Terms),
	}
}

// resolveFunctionarySegment is the only adapter that actually consumes
// ctx (character/functionary_generate.go's own BeginFunctionary needs
// ctx.TermsServedSoFar, and functionaryRankName needs
// ctx.PrecedingCareer for the F6 title). Office Politics has no death
// mechanic at all — Survived is always true, the same as Citizen/Rogue/
// Entertainer — and WoundBadges is never set (Disabled here means
// "career ends," not a physical wound; see ResolveFunctionaryTerm's own
// doc comment for why scoutWoundBadges must not be called on this
// segment).
func resolveFunctionarySegment(r *dice.Roller, upp UPP, maxTerms int, ctx segmentContext) careerSegment {
	aging := ctx.aging()

	career, careerUPP, finalTier := resolveFunctionaryCareerWithBudget(r, upp, maxTerms, ctx, aging)
	if aging.alive() {
		career.MusteringOut = ResolveFunctionaryMusterOut(r, career, finalTier)
	}

	boostedUPP, bonuses := ApplyMusteringOut(career.MusteringOut, careerUPP)

	return careerSegment{
		Career: career, UPP: boostedUPP, Survived: true,
		Fame: bonuses.Fame, Cash: bonuses.Cash,
		Skills: allSkillsFromTerms(career.Terms),
	}
}

// resolveCraftsmanSegment is the second adapter (after Functionary's
// own) that consumes ctx — BeginCraftsman and the "New Trade" skill
// cell both need ctx.SkillsSoFar. Craftsman "does not roll Risk and
// Reward" at all (Book 1 p.75) — no death mechanic, Survived is always
// true, WoundBadges is never set, matching Citizen/Rogue/Entertainer's
// own precedent. Masterpieces are objects the character created, not
// liquid Cash — recorded in Equipment (the first real use of that
// field), not folded into bonuses.Cash.
func resolveCraftsmanSegment(r *dice.Roller, upp UPP, maxTerms int, ctx segmentContext) careerSegment {
	aging := ctx.aging()

	career, careerUPP := resolveCraftsmanCareerWithBudget(r, upp, maxTerms, ctx, aging)
	if aging.alive() {
		career.MusteringOut = ResolveCraftsmanMusterOut(r, career)
	}

	boostedUPP, bonuses := ApplyMusteringOut(career.MusteringOut, careerUPP)

	var equipment []string

	for _, t := range career.Terms {
		if t.RewardResult != "" && t.RewardResult != "None" {
			equipment = append(equipment, t.RewardResult)
		}
	}

	return careerSegment{
		Career: career, UPP: boostedUPP, Survived: true,
		Fame: bonuses.Fame + craftsmanCareerFame(career.Terms), Cash: bonuses.Cash,
		Skills: allSkillsFromTerms(career.Terms), Equipment: equipment,
	}
}

func resolveNobleSegment(r *dice.Roller, upp UPP, maxTerms int, ctx segmentContext) careerSegment {
	aging := ctx.aging()

	career, careerUPP := resolveNobleCareerAndUPPWithBudget(r, upp, maxTerms, aging)
	if aging.alive() {
		career.MusteringOut = ResolveNobleMusterOut(r, career)
	}

	boostedUPP, bonuses := ApplyMusteringOut(career.MusteringOut, careerUPP)
	ok := len(career.Terms) > 0

	fame := bonuses.Fame
	if ok {
		fame += nobleBaseFame(upp.Characteristics[C6]) + nobleExileFame(career.Terms)
	}

	return careerSegment{
		Career: career, UPP: boostedUPP, Survived: ok,
		Fame: fame, Cash: bonuses.Cash, Skills: allSkillsFromTerms(career.Terms),
	}
}

// validateCareerChain checks careerNames (already lowercased/trimmed by
// the caller) against the rules this slice's own plan-file Context
// derives from Book 1 pp.63-66: every name must be a known, chainable
// career; "citizen" may only be the first entry (p.64: "may not
// transfer to Citizen"); "functionary" and "craftsman" may never be the
// first entry (p.63: both "unavailable as initial careers" — stated
// explicitly rather than relying solely on each one's own emergent
// Begin-failure at 0 prior terms/skills); Functionary and Noble must be
// terminal because a character may not transfer out of either; and no
// two adjacent entries may repeat the
// same career ("selecting a different career" — re-entering the very
// career just left is meaningless when Continue already offers that for
// free).
func validateCareerChain(careerNames []string) error {
	if len(careerNames) == 0 {
		return errors.New("career chain must not be empty")
	}

	for i, name := range careerNames {
		if _, ok := careerChainRegistry[name]; !ok {
			return fmt.Errorf("unknown career %q; valid careers are %v", name, sortedCareerChainNames())
		}

		if name == "citizen" && i != 0 {
			return fmt.Errorf(
				"%q may only be the first career in a chain (a character may not transfer to Citizen)",
				name,
			)
		}

		if (name == "functionary" || name == "craftsman") && i == 0 {
			return fmt.Errorf("%q may not be the first career in a chain (never a first career, Book 1 p.63)", name)
		}

		if (name == "functionary" || name == "noble") && i != len(careerNames)-1 {
			return fmt.Errorf("%q must be the final career in a chain (a character may not transfer from it)", name)
		}

		if i > 0 && careerNames[i-1] == name {
			return fmt.Errorf("career %q may not immediately follow itself at position %d", name, i)
		}
	}

	return nil
}

func sortedCareerChainNames() []string {
	names := make([]string, 0, len(careerChainRegistry))
	for name := range careerChainRegistry {
		names = append(names, name)
	}

	slices.Sort(names)

	return names
}

// chainRank returns the Rank held after the whole chain, scanning
// careers in reverse for the first non-empty lastTermRank — Book 1
// p.66's own Reserves rule: "A character who leaves a military, naval,
// or marine career ... maintains his or her last held rank as a Reserve
// Rank," so a later rankless career (e.g. Scout after Marine) must not
// blank out an earlier Armed-Forces rank.
func chainRank(careers []Career) string {
	for _, career := range slices.Backward(careers) {
		if rank := lastTermRank(career.Terms); rank != "" {
			return rank
		}
	}

	return ""
}

// careerChainAccumulator carries the running totals GenerateCareerChainCharacter
// builds up across segments — split out so the main function's own loop
// only has to decide continue/break, not also thread six separate
// running variables through it.
type careerChainAccumulator struct {
	careers                              []Career
	skills                               []SkillLevel
	medals                               []string
	equipment                            []string
	fame, cash, woundBadges, termsServed int
}

// addSegment folds seg into the accumulator. A zero-term seg (Begin
// failed) still contributes its own Career entry — matching every
// existing single-career "never qualified" precedent — but nothing
// else.
func (acc *careerChainAccumulator) addSegment(seg careerSegment) {
	acc.careers = append(acc.careers, seg.Career)

	if len(seg.Career.Terms) == 0 {
		return
	}

	acc.termsServed += len(seg.Career.Terms)
	acc.skills = append(acc.skills, seg.Skills...)
	acc.medals = append(acc.medals, seg.Medals...)
	acc.equipment = append(acc.equipment, seg.Equipment...)
	acc.fame += seg.Fame
	acc.cash += seg.Cash
	acc.woundBadges += seg.WoundBadges
}

// segmentBudget returns the term budget the next segment attempt should
// run with, and false if an -age target (ageTarget != 0) has already
// been reached — meaning the chain must stop attempting anything
// further, including the Citizen fallback, without even calling that
// segment's own resolver (the literal "don't attempt," never a
// retroactive cut of an already-resolved term). ageTarget == 0 means no
// cap at all — every segment gets the same maxCareerTerms budget every
// single-career caller has always used.
func segmentBudget(ageTarget, currentAge int) (int, bool) {
	if ageTarget == 0 {
		return maxCareerTerms, true
	}

	// Measured from the character's actual age rather than from terms
	// alone: a failed Begin costs a year (Book 1 p.65) without serving a
	// term, so a terms-only count would quietly hand back years the
	// character had already spent and overshoot the target.
	remaining := (ageTarget - currentAge) / 4
	if remaining <= 0 {
		return 0, false
	}

	return min(maxCareerTerms, remaining), true
}

func generateCareerChainStart(r *dice.Roller) (UPP, string, []SkillLevel) {
	upp := GenerateUPP(r)
	homeworld, skills := GenerateHomeworldSkills(r)

	return upp, homeworld, skills
}

// chainSegmentContext builds the per-segment context from the chain's
// running state — the shared Aging simulation, terms served so far, the
// skills accumulated to date, and the immediately-preceding career name.
func chainSegmentContext(
	aging *agingSimulation,
	acc *careerChainAccumulator,
	precedingCareer string,
) segmentContext {
	return segmentContext{
		Aging:            aging,
		PrecedingCareer:  precedingCareer,
		TermsServedSoFar: acc.termsServed,
		SkillsSoFar:      aggregateSkills(acc.skills),
	}
}

// applyCitizenFallback runs Book 1 p.64's guaranteed Citizen fallback
// ("Begin Citizen Life is Automatic") for a chain in which no listed
// career ever began, folding the result into acc and returning the
// resulting UPP. Subject to the same -age budget as any other segment,
// so a target already reached skips it entirely rather than handing out
// a free career.
func applyCitizenFallback(
	r *dice.Roller,
	upp UPP,
	acc *careerChainAccumulator,
	aging *agingSimulation,
	ageTarget int,
) UPP {
	maxTerms, attemptAllowed := segmentBudget(ageTarget, aging.age())
	if !attemptAllowed {
		return upp
	}

	seg := careerChainRegistry["citizen"](r, upp, maxTerms, segmentContext{Aging: aging})
	acc.addSegment(seg)

	return seg.UPP
}

// GenerateCareerChainCharacter generates a full Human Character across
// an ordered sequence of careerNames (already lowercased/trimmed),
// implementing Book 1's own "Changing Careers" (pp.65-66). An ordered
// list means a voluntary transfer after one completed term in each
// non-final career; that choice is made instead of rolling Continue.
// The final career follows its normal Continue rolls, so a failed
// Continue always ends Career Resolution. A career whose Begin fails
// contributes a zero-term Career entry (matching every existing single-career
// "never qualified" precedent) and the chain moves on to the next name
// without consuming anything. If every listed career fails to Begin,
// Citizen is the guaranteed fallback (p.64: "Begin Citizen Life is
// Automatic") — no such thing as a character who never held any career
// at all.
//
// ageTarget (0 = unbounded) stops the chain from attempting any further
// term or career transition once the character's age would reach or
// exceed it — a target under 18 (or one reached before any listed
// career even Begins) degrades cleanly to an empty Careers slice, a
// valid "never had a job" outcome, not an error. A term is an atomic
// 4-year block, so a target that isn't 18+4N is reached from *under*,
// never over (e.g. -age 41 yields a final age of 38).
//
// Returns ok=false only when a segment ends in Dead (Book 1 p.69:
// "Dying During Character Generation" voids the whole attempt) — same
// as every existing single-career entry point, everything accumulated
// through the fatal segment is still returned, not discarded, matching
// cmd/chargen's own "the sheet above still prints in full" precedent.
func GenerateCareerChainCharacter(r *dice.Roller, careerNames []string, ageTarget int) (Character, bool, error) {
	if err := validateCareerChain(careerNames); err != nil {
		return Character{}, false, err
	}

	upp, homeworld, homeworldSkills := generateCareerChainStart(r)

	acc := careerChainAccumulator{skills: slices.Clone(homeworldSkills)}

	// One simulation for the whole chain — see agingSimulation's own doc
	// comment for why it cannot be per-segment.
	var aging agingSimulation

	everSucceeded, survived := false, true
	precedingCareer := ""

	for i, name := range careerNames {
		maxTerms, attemptAllowed := segmentBudget(ageTarget, aging.age())
		if !attemptAllowed {
			break
		}

		if i < len(careerNames)-1 {
			maxTerms = min(maxTerms, 1)
		}

		seg := careerChainRegistry[name](r, upp, maxTerms,
			chainSegmentContext(&aging, &acc, precedingCareer))
		upp = seg.UPP
		acc.addSegment(seg)

		// Checked before the zero-term continue below, not after it: a
		// segment that served no terms can still have killed the
		// character, since a failed Begin costs a year (Book 1 p.65) and
		// that year can cross an Aging checkpoint. seg.Survived answers
		// only "did this career kill them?", so Aging — which is tracked
		// separately and spans the whole chain — has to be asked on its
		// own or the next named career is attempted by someone dead.
		if !aging.alive() {
			break
		}

		if len(seg.Career.Terms) == 0 {
			continue
		}

		everSucceeded = true
		precedingCareer = name

		if !seg.Survived {
			survived = false

			break
		}

		if segmentEndsCareerResolution(seg) {
			break
		}
	}

	// Guaranteed fallback only for the living: a character Aging killed
	// during a failed Begin never reaches Citizen Life either.
	if !everSucceeded && aging.alive() {
		upp = applyCitizenFallback(r, upp, &acc, &aging, ageTarget)
	}

	age, lifeStage, notes, survived := finalizeAging(&aging, survived)
	birthdate := GenerateBirthdate(r, age)

	return Character{
		Species:        "Human",
		GeneticProfile: humanGeneticProfile,
		UPP:            upp,
		Homeworld:      homeworld,
		Birthworld:     homeworld,
		Birthdate:      birthdate,
		Age:            age,
		LifeStage:      lifeStage,
		Notes:          notes,
		Rank:           chainRank(acc.careers),
		Fame:           acc.fame,
		Cash:           acc.cash,
		WoundBadges:    acc.woundBadges,
		Careers:        acc.careers,
		Skills:         aggregateSkills(acc.skills),
		Medals:         acc.medals,
		Equipment:      acc.equipment,
	}, survived, nil
}
