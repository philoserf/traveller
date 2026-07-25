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
}

// segmentContext carries the two things Functionary's own Begin target
// and Rank-title association need that every other adapter ignores: the
// running terms-served total from before this segment is attempted, and
// the immediately-preceding chain entry's own career name (empty for
// the very first attempted entry). Every existing adapter takes this
// parameter and ignores it — the same low-risk, mechanical threading
// pattern the -age target already used for maxTerms.
type segmentContext struct {
	PrecedingCareer  string
	TermsServedSoFar int
}

type careerSegmentResolver func(r *dice.Roller, upp UPP, maxTerms int, ctx segmentContext) careerSegment

// careerChainRegistry is every career resolvable as a link in a
// multi-career chain. Noble is deliberately absent (its Begin is
// Soc-gated with no Continue/retry-a-different-career shape); Craftsman
// is absent because it isn't implemented at all yet (a separate,
// architecturally-blocked career).
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
	resolveCareer func(r *dice.Roller, upp UPP, maxTerms int) (Career, UPP),
	resolveMusterOut func(r *dice.Roller, career Career) MusteringOut,
	careerFame func(career Career) int,
) careerSegment {
	career, careerUPP := resolveCareer(r, upp, maxTerms)
	career.MusteringOut = resolveMusterOut(r, career)

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

func resolveScoutSegment(r *dice.Roller, upp UPP, maxTerms int, _ segmentContext) careerSegment {
	return resolveRiskCareerSegment(
		r,
		upp,
		maxTerms,
		resolveScoutCareerWithBudget,
		ResolveScoutMusterOut,
		scoutDiscoveryFame,
	)
}

func resolveMarineSegment(r *dice.Roller, upp UPP, maxTerms int, _ segmentContext) careerSegment {
	return resolveRiskCareerSegment(
		r, upp, maxTerms, resolveMarineCareerWithBudget, ResolveMarineMusterOut, marineCareerFame)
}

func resolveSoldierSegment(r *dice.Roller, upp UPP, maxTerms int, _ segmentContext) careerSegment {
	return resolveRiskCareerSegment(
		r, upp, maxTerms, resolveSoldierCareerWithBudget, ResolveSoldierMusterOut, soldierCareerFame)
}

func resolveSpacerSegment(r *dice.Roller, upp UPP, maxTerms int, _ segmentContext) careerSegment {
	return resolveRiskCareerSegment(
		r, upp, maxTerms, resolveSpacerCareerWithBudget, ResolveSpacerMusterOut, spacerCareerFame)
}

func resolveAgentSegment(r *dice.Roller, upp UPP, maxTerms int, _ segmentContext) careerSegment {
	return resolveRiskCareerSegment(
		r,
		upp,
		maxTerms,
		resolveAgentCareerWithBudget,
		ResolveAgentMusterOut,
		agentCareerFame,
	)
}

// resolveRogueSegment mirrors buildRogueCharacter's own body
// (rogue_character_generate.go), stopping short of finalizeAging. Rogue
// never modifies a characteristic at all (rogue_loop.go's own doc
// comment: "Rogue never modifies its own CC"), so upp passes through
// unchanged; there is no death mechanic, so Survived is always true.
func resolveRogueSegment(r *dice.Roller, upp UPP, maxTerms int, _ segmentContext) careerSegment {
	career := resolveRogueCareerWithBudget(r, upp, maxTerms)
	career.MusteringOut = ResolveRogueMusterOut(r, career)

	boostedUPP, bonuses := ApplyMusteringOut(career.MusteringOut, upp)

	ok := len(career.Terms) > 0

	cash := bonuses.Cash
	fame := bonuses.Fame

	if ok {
		for _, t := range career.Terms {
			cash += t.SchemePayoff

			if t.Imprisoned {
				fame += 3
			} else {
				fame += 2
			}
		}
	}

	return careerSegment{
		Career: career, UPP: boostedUPP, Survived: ok,
		Fame: fame, Cash: cash, Skills: allSkillsFromTerms(career.Terms),
	}
}

// resolveScholarSegment mirrors buildScholarCharacter's own body
// (scholar_character_generate.go), stopping short of finalizeAging.
func resolveScholarSegment(r *dice.Roller, upp UPP, maxTerms int, _ segmentContext) careerSegment {
	career, careerUPP := resolveScholarCareerWithBudget(r, upp, maxTerms)
	career.MusteringOut = ResolveScholarMusterOut(r, career, careerUPP)

	boostedUPP, bonuses := ApplyMusteringOut(career.MusteringOut, careerUPP)

	ok := len(career.Terms) > 0 && career.Terms[len(career.Terms)-1].RiskResult != Dead

	fame := bonuses.Fame

	if ok {
		startTier := scholarStartTier(int(careerUPP.Characteristics[C5]))
		fame += scholarCareerFame(career.Terms, startTier)
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
// Talent are separate scalars), so upp passes through unchanged; its
// "Dead" means Talent exhausted, not physical death, so Survived is
// always true and WoundBadges is always 0 (a Talent setback, not a
// physical wound).
func resolveEntertainerSegment(r *dice.Roller, upp UPP, maxTerms int, _ segmentContext) careerSegment {
	career, fame := resolveEntertainerCareerWithBudget(r, upp, maxTerms)
	career.MusteringOut = ResolveEntertainerMusterOut(r, career, fame)

	boostedUPP, bonuses := ApplyMusteringOut(career.MusteringOut, upp)

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
func resolveMerchantSegment(r *dice.Roller, upp UPP, maxTerms int, _ segmentContext) careerSegment {
	career, careerUPP, isOfficer, tier := resolveMerchantCareerWithBudget(r, upp, maxTerms)
	career.MusteringOut = ResolveMerchantMusterOut(r, career, isOfficer, tier)

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
func resolveCitizenSegment(r *dice.Roller, upp UPP, maxTerms int, _ segmentContext) careerSegment {
	career := resolveCitizenCareerWithBudget(r, upp, maxTerms)
	career.MusteringOut = ResolveCitizenMusterOut(r, career)

	boostedUPP, bonuses := ApplyMusteringOut(career.MusteringOut, upp)

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
	career, careerUPP, finalTier := resolveFunctionaryCareerWithBudget(r, upp, maxTerms, ctx)
	career.MusteringOut = ResolveFunctionaryMusterOut(r, career, finalTier)

	boostedUPP, bonuses := ApplyMusteringOut(career.MusteringOut, careerUPP)

	return careerSegment{
		Career: career, UPP: boostedUPP, Survived: true,
		Fame: bonuses.Fame, Cash: bonuses.Cash,
		Skills: allSkillsFromTerms(career.Terms),
	}
}

// validateCareerChain checks careerNames (already lowercased/trimmed by
// the caller) against the rules this slice's own plan-file Context
// derives from Book 1 pp.63-66: every name must be a known, chainable
// career; "citizen" may only be the first entry (p.64: "may not
// transfer to Citizen"); "functionary" may never be the first entry
// (p.63: "Functionary ... unavailable as initial careers" — stated
// explicitly rather than relying solely on BeginFunctionary's own
// emergent zero-target failure at 0 prior terms); "noble" is recognized
// but rejected with a clearer message than "unknown career" (Noble's
// Begin doesn't fit this shape); and no two adjacent entries may repeat
// the same career ("selecting a different career" — re-entering the
// very career just left is meaningless when Continue already offers
// that for free).
func validateCareerChain(careerNames []string) error {
	if len(careerNames) == 0 {
		return errors.New("career chain must not be empty")
	}

	for i, name := range careerNames {
		if name == "noble" {
			return fmt.Errorf("%q may only be used alone (-career noble), not in a multi-career chain", name)
		}

		if _, ok := careerChainRegistry[name]; !ok {
			return fmt.Errorf("unknown career %q; valid careers are %v", name, sortedCareerChainNames())
		}

		if name == "citizen" && i != 0 {
			return fmt.Errorf(
				"%q may only be the first career in a chain (a character may not transfer to Citizen)",
				name,
			)
		}

		if name == "functionary" && i == 0 {
			return fmt.Errorf("%q may not be the first career in a chain (Functionary is never a first career)", name)
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
func segmentBudget(ageTarget, maxAllowedTotalTerms, termsServed int) (int, bool) {
	if ageTarget == 0 {
		return maxCareerTerms, true
	}

	remaining := maxAllowedTotalTerms - termsServed
	if remaining <= 0 {
		return 0, false
	}

	return min(maxCareerTerms, remaining), true
}

// GenerateCareerChainCharacter generates a full Human Character across
// an ordered sequence of careerNames (already lowercased/trimmed),
// implementing Book 1's own "Changing Careers" (pp.65-66): each career
// runs to its own natural end (Continue failure, Disabled, Dead, or the
// existing maxCareerTerms safety cap) before the next listed career is
// attempted, in order. A career whose Begin fails contributes a
// zero-term Career entry (matching every existing single-career
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

	maxAllowedTotalTerms := 0
	if ageTarget != 0 {
		maxAllowedTotalTerms = max(0, (ageTarget-18)/4)
	}

	upp := GenerateUPP(r)
	homeworld, homeworldSkills := GenerateHomeworldSkills(r)

	acc := careerChainAccumulator{skills: slices.Clone(homeworldSkills)}
	everSucceeded, survived := false, true
	precedingCareer := ""

	for _, name := range careerNames {
		maxTerms, attemptAllowed := segmentBudget(ageTarget, maxAllowedTotalTerms, acc.termsServed)
		if !attemptAllowed {
			break
		}

		ctx := segmentContext{PrecedingCareer: precedingCareer, TermsServedSoFar: acc.termsServed}

		seg := careerChainRegistry[name](r, upp, maxTerms, ctx)
		upp = seg.UPP
		acc.addSegment(seg)

		if len(seg.Career.Terms) == 0 {
			continue
		}

		everSucceeded = true
		precedingCareer = name

		if !seg.Survived {
			survived = false

			break
		}
	}

	if !everSucceeded {
		if maxTerms, attemptAllowed := segmentBudget(ageTarget, maxAllowedTotalTerms, acc.termsServed); attemptAllowed {
			seg := careerChainRegistry["citizen"](r, upp, maxTerms, segmentContext{})
			upp = seg.UPP
			acc.addSegment(seg)
		}
	}

	finalUPP, age, lifeStage, notes := finalizeAging(r, upp, acc.termsServed, survived)
	birthdate := GenerateBirthdate(r, age)

	return Character{
		Species:        "Human",
		GeneticProfile: humanGeneticProfile,
		UPP:            finalUPP,
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
		Skills:         acc.skills,
		Medals:         acc.medals,
	}, survived, nil
}
