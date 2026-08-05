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
	Career   Career
	UPP      UPP // boosted by this segment's own Mustering Out
	Survived bool
	Fame     int
	// FameAwards is this segment's own Fame as individual awards, which
	// Book 1 p.91's Fame Stacks rule needs — the chain resolves it over
	// every award from every career, not over per-career subtotals.
	FameAwards  []int
	Cash        int
	WoundBadges int
	Skills      []SkillLevel
	Medals      []string
	Equipment   []string // Craftsman only: Masterpieces created this segment
	// Masterpieces are the same items as Equipment, structured — QREBS,
	// Master Points and creation age (masterpiece.go).
	Masterpieces []Masterpiece
	// LandGrants are Book 1 p.88's own awards of territory earned in this
	// segment — Noble fiefs and Scout Discovery grants. Carried through
	// the chain because p.68 retains them at Mustering Out, so a later
	// career never takes them away.
	LandGrants []LandGrant
	// Education is this segment's own ctx.Education, returned back out
	// so GenerateCareerChainCharacter can carry a Later Education
	// mutation (#113 item 5) both into the final Character and forward
	// into the next segment's own ctx.Education — segmentContext is
	// otherwise built fresh, by value, per segment. Every adapter
	// returns it, even the ones that never mutate it (a plain pass-
	// through of what they were given), so the chain loop's own
	// education = seg.Education never zeroes it out for a career that
	// doesn't yet offer Later Education.
	Education Education
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
	// Education is CharGen step C's own result, which Scholar alone reads:
	// Book 1 p.76 gives every Scholar a Major and a Minor, taken from his
	// degree if he has one. Every other career's skill tables resolve
	// their Major/Minor cells at final assembly instead, where the
	// character is in view.
	Education Education
	// PriorCareers are every career already served, in order. Rogue
	// needs the whole list rather than PrecedingCareer alone, for Book 1
	// p.84's "may select for his Scheme (rather than roll) any previous
	// career" — any, not merely the last.
	PriorCareers []string
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

// commissionAppliesTo is #113's "consumed once" rule for a Service
// Academy/OTC/NOTC Commission: careerName is eligible only if it's one of
// commissionCareers (Education.CommissionCareers) and none of the
// character's priorCareers already consumed one — a Commission covering
// both Spacer and Marine (NOTC) must not auto-commission a second chain
// entry once the first has used it. No new mutable state is needed:
// segmentContext.PriorCareers already records every career served so far.
func commissionAppliesTo(commissionCareers, priorCareers []string, careerName string) bool {
	if !slices.Contains(commissionCareers, careerName) {
		return false
	}

	for _, prior := range priorCareers {
		if slices.Contains(commissionCareers, prior) {
			return false
		}
	}

	return true
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
// Character-assembly functions: resolve the career, apply Mustering Out,
// then tally Fame/Land Grants/Wound Badges/Skills/Medals the same gates
// and order as the single-career builder, so a chain segment and a
// standalone career produce the identical character from the identical
// seed.
//
// commissioned/flightSchool thread through to resolveCareer for the
// three Commission-eligible careers; Scout and Agent always pass false,
// false and adapt their own narrower resolveCareer signature to match
// (resolveScoutSegment/resolveAgentSegment). flightSchool (#113) adds
// Honors on top of commissioned — p.61's "Service Academy Honors
// Graduates"/"College or University Honors Graduates who participated in
// OTC or NOTC may attend Flight School" both reduce to "commissioned AND
// Honors", so callers compute it as a one-line derivation rather than
// tracking separate state.
func resolveRiskCareerSegment(
	r *dice.Roller,
	upp UPP,
	maxTerms int,
	ctx segmentContext,
	resolveCareer func(
		r *dice.Roller, upp UPP, maxTerms int, aging *agingSimulation, commissioned, flightSchool bool,
	) (Career, UPP),
	resolveMusterOut func(r *dice.Roller, career Career) MusteringOut,
	careerFameAwards func(career Career) []int,
	commissioned, flightSchool bool,
) careerSegment {
	aging := ctx.aging()

	career, careerUPP := resolveCareer(r, upp, maxTerms, aging, commissioned, flightSchool)

	if aging.alive() {
		career.MusteringOut = resolveMusterOut(r, career)
	}

	boostedUPP, bonuses := ApplyMusteringOut(r, career, careerUPP)

	ok := len(career.Terms) > 0 && career.Terms[len(career.Terms)-1].RiskResult != Dead

	fameAwards := bonuses.FameAwards
	if ok {
		fameAwards = append(fameAwards, careerFameAwards(career)...)
	}

	fame := sumInts(fameAwards)

	// Same gate and same nil-for-non-Scout behavior as the single-career
	// builder (character_generate.go), so a chain segment and a standalone
	// career produce the identical character from the identical seed.
	var landGrants []LandGrant
	if ok {
		landGrants = scoutDiscoveryLandGrants(r, career)
	}

	return careerSegment{
		Career: career, UPP: boostedUPP, Survived: ok, LandGrants: append(landGrants, bonuses.LandGrants...),
		Fame: fame, FameAwards: fameAwards, Cash: bonuses.Cash, WoundBadges: careerWoundBadges(career),
		Skills: append(allSkillsFromTerms(career.Terms), bonuses.Skills...), Medals: allMedalsFromTerms(career.Terms),
		Education: ctx.Education,
	}
}

// commissionStatus is commissioned/flightSchool for careerName, shared by
// resolveMarineSegment/resolveSoldierSegment/resolveSpacerSegment —
// commissioned via commissionAppliesTo, flightSchool (#113) adding
// Honors on top per p.61's "Service Academy Honors Graduates"/"College
// or University Honors Graduates who participated in OTC or NOTC may
// attend Flight School", both reducing to "commissioned AND Honors".
func commissionStatus(ctx segmentContext, careerName string) (bool, bool) {
	commissioned := commissionAppliesTo(ctx.Education.CommissionCareers, ctx.PriorCareers, careerName)
	flightSchool := commissioned && ctx.Education.Honors

	return commissioned, flightSchool
}

// dropCommissionParams adapts a career whose own resolveCareer signature
// has no room for commissioned/flightSchool (Scout, Agent — neither is
// Commission-eligible) to resolveRiskCareerSegment's shared, wider
// signature.
func dropCommissionParams(
	resolveCareer func(r *dice.Roller, upp UPP, maxTerms int, aging *agingSimulation) (Career, UPP),
) func(r *dice.Roller, upp UPP, maxTerms int, aging *agingSimulation, commissioned, flightSchool bool) (Career, UPP) {
	return func(r *dice.Roller, upp UPP, maxTerms int, aging *agingSimulation, _, _ bool) (Career, UPP) {
		return resolveCareer(r, upp, maxTerms, aging)
	}
}

func resolveScoutSegment(r *dice.Roller, upp UPP, maxTerms int, ctx segmentContext) careerSegment {
	return resolveRiskCareerSegment(
		r, upp, maxTerms, ctx,
		dropCommissionParams(resolveScoutCareerWithBudget),
		ResolveScoutMusterOut,
		scoutDiscoveryFameAwards,
		false, false,
	)
}

func resolveMarineSegment(r *dice.Roller, upp UPP, maxTerms int, ctx segmentContext) careerSegment {
	commissioned, flightSchool := commissionStatus(ctx, MarineCareerName)

	return resolveRiskCareerSegment(
		r, upp, maxTerms, ctx,
		resolveMarineCareerWithBudget, ResolveMarineMusterOut, marineCareerFameAwards,
		commissioned, flightSchool,
	)
}

func resolveSoldierSegment(r *dice.Roller, upp UPP, maxTerms int, ctx segmentContext) careerSegment {
	commissioned, flightSchool := commissionStatus(ctx, SoldierCareerName)

	return resolveRiskCareerSegment(
		r, upp, maxTerms, ctx,
		resolveSoldierCareerWithBudget, ResolveSoldierMusterOut, soldierCareerFameAwards,
		commissioned, flightSchool,
	)
}

func resolveSpacerSegment(r *dice.Roller, upp UPP, maxTerms int, ctx segmentContext) careerSegment {
	commissioned, flightSchool := commissionStatus(ctx, SpacerCareerName)

	return resolveRiskCareerSegment(
		r, upp, maxTerms, ctx,
		resolveSpacerCareerWithBudget, ResolveSpacerMusterOut, spacerCareerFameAwards,
		commissioned, flightSchool,
	)
}

func resolveAgentSegment(r *dice.Roller, upp UPP, maxTerms int, ctx segmentContext) careerSegment {
	return resolveRiskCareerSegment(
		r, upp, maxTerms, ctx,
		dropCommissionParams(resolveAgentCareerWithBudget),
		ResolveAgentMusterOut,
		agentCareerFameAwards,
		false, false,
	)
}

// resolveRogueSegment mirrors buildRogueCharacter's own body
// (rogue_character_generate.go), stopping short of finalizeAging. Rogue
// has no Risk-driven characteristic reduction; Personal awards can
// still improve UPP. There is no death mechanic, so Survived is always true.
// Fame/Cash share rogueTermsCash with buildRogueCharacter.
func resolveRogueSegment(r *dice.Roller, upp UPP, maxTerms int, ctx segmentContext) careerSegment {
	aging := ctx.aging()

	// &ctx.Education, not a chain-wide pointer: any Later Education
	// attendance during this segment mutates this segment's own copy.
	// That mutation does reach both the next chain segment and the
	// final character — careerSegment.Education carries it out, and
	// GenerateCareerChainCharacter's own loop threads it forward as
	// education = seg.Education after every segment (career_chain.go's
	// own doc comment on that assignment has the details).
	career, careerUPP := resolveRogueCareerAndUPPWithBudget(r, upp, maxTerms, aging, ctx.PriorCareers, &ctx.Education)
	if aging.alive() {
		career.MusteringOut = ResolveRogueMusterOut(r, career)
	}

	boostedUPP, bonuses := ApplyMusteringOut(r, career, careerUPP)

	ok := len(career.Terms) > 0

	cash := bonuses.Cash
	fameAwards := bonuses.FameAwards

	if ok {
		termCash := rogueTermsCash(career.Terms)
		fameAwards = append(fameAwards, rogueTermFameAwards(career.Terms)...)
		cash += termCash
	}

	fame := sumInts(fameAwards)

	return careerSegment{
		Career:     career,
		UPP:        boostedUPP,
		Survived:   ok,
		Fame:       fame,
		FameAwards: fameAwards,
		Cash:       cash,
		Skills:     append(allSkillsFromTerms(career.Terms), bonuses.Skills...),
		LandGrants: bonuses.LandGrants,
		Education:  ctx.Education,
	}
}

// resolveScholarSegment mirrors buildScholarCharacter's own body
// (scholar_character_generate.go), stopping short of finalizeAging.
func resolveScholarSegment(r *dice.Roller, upp UPP, maxTerms int, ctx segmentContext) careerSegment {
	aging := ctx.aging()

	career, careerUPP := resolveScholarCareerWithBudget(r, upp, maxTerms, aging, ctx.Education)
	if aging.alive() {
		// upp, not careerUPP — ResolveScholarMusterOut needs entry-time
		// Edu to know Scholar's starting rank tier (its own doc comment),
		// not the post-career value. Mirrors buildScholarCharacter.
		career.MusteringOut = ResolveScholarMusterOut(r, career, upp)
	}

	boostedUPP, bonuses := ApplyMusteringOut(r, career, careerUPP)

	ok := len(career.Terms) > 0 && career.Terms[len(career.Terms)-1].RiskResult != Dead

	fameAwards := bonuses.FameAwards

	if ok {
		// upp, not careerUPP — same entry-time-Edu requirement as
		// ResolveScholarMusterOut above.
		fameAwards = append(fameAwards, scholarSegmentFameAwards(upp, career.Terms)...)
	}

	fame := sumInts(fameAwards)

	// Same gate and same nil-for-non-Scout behavior as the single-career
	// builder (character_generate.go), so a chain segment and a standalone
	// career produce the identical character from the identical seed.
	var landGrants []LandGrant
	if ok {
		landGrants = scoutDiscoveryLandGrants(r, career)
	}

	return careerSegment{
		Career: career, UPP: boostedUPP, Survived: ok, LandGrants: append(landGrants, bonuses.LandGrants...),
		Fame: fame, FameAwards: fameAwards, Cash: bonuses.Cash, WoundBadges: careerWoundBadges(career),
		Skills: append(allSkillsFromTerms(career.Terms), bonuses.Skills...), Medals: allMedalsFromTerms(career.Terms),
		Education: ctx.Education,
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

	boostedUPP, bonuses := ApplyMusteringOut(r, career, careerUPP)

	ok := len(career.Terms) > 0

	return careerSegment{
		Career: career, UPP: boostedUPP, Survived: ok,
		Fame: fame + bonuses.Fame, FameAwards: append(bonuses.FameAwards, fame), Cash: bonuses.Cash,
		Skills: append(allSkillsFromTerms(career.Terms), bonuses.Skills...), LandGrants: bonuses.LandGrants,
		Education: ctx.Education,
	}
}

// resolveMerchantSegment mirrors buildMerchantCharacter's own body
// (merchant_character_generate.go), stopping short of finalizeAging.
// BeginMerchant never fails, so len(career.Terms) > 0 would be
// unconditional if maxTerms were always positive — but here maxTerms is
// chain-supplied and can be 0 (an exhausted -age budget), so this checks
// explicitly rather than indexing career.Terms unconditionally.
// buildMerchantCharacter guards identically even though its own
// hardcoded maxCareerTerms keeps this path unreachable there today.
func resolveMerchantSegment(r *dice.Roller, upp UPP, maxTerms int, ctx segmentContext) careerSegment {
	aging := ctx.aging()

	career, careerUPP, isOfficer, tier := resolveMerchantCareerWithBudget(r, upp, maxTerms, aging)
	if aging.alive() {
		career.MusteringOut = ResolveMerchantMusterOut(r, career, isOfficer, tier)
	}

	boostedUPP, bonuses := ApplyMusteringOut(r, career, careerUPP)

	ok := len(career.Terms) > 0 && career.Terms[len(career.Terms)-1].RiskResult != Dead

	fameAwards := bonuses.FameAwards
	if ok {
		// Two separate p.91 awards, appended individually rather than
		// pre-summed, so the Fame Stacks cap sees them as the book counts
		// them (fame.go's own resolveFameStacks) — mirrors
		// buildMerchantCharacter (merchant_character_generate.go) and
		// resolveNobleSegment below.
		fameAwards = append(fameAwards, merchantCareerFame(tier))
		if owner := merchantShipOwnerFame(r, career); owner > 0 {
			fameAwards = append(fameAwards, owner)
		}
	}

	fame := sumInts(fameAwards)

	// Same gate and same nil-for-non-Scout behavior as the single-career
	// builder (character_generate.go), so a chain segment and a standalone
	// career produce the identical character from the identical seed.
	var landGrants []LandGrant
	if ok {
		landGrants = scoutDiscoveryLandGrants(r, career)
	}

	return careerSegment{
		Career: career, UPP: boostedUPP, Survived: ok, LandGrants: append(landGrants, bonuses.LandGrants...),
		Fame: fame, FameAwards: fameAwards, Cash: bonuses.Cash, WoundBadges: careerWoundBadges(career),
		Skills:    append(allSkillsFromTerms(career.Terms), bonuses.Skills...),
		Education: ctx.Education,
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

	boostedUPP, bonuses := ApplyMusteringOut(r, career, careerUPP)

	return careerSegment{
		Career: career, UPP: boostedUPP, Survived: true,
		Fame: bonuses.Fame, FameAwards: bonuses.FameAwards, Cash: bonuses.Cash,
		Skills: append(allSkillsFromTerms(career.Terms), bonuses.Skills...), LandGrants: bonuses.LandGrants,
		Education: ctx.Education,
	}
}

// resolveFunctionarySegment is the only adapter that actually consumes
// ctx (character/functionary_generate.go's own BeginFunctionary needs
// ctx.TermsServedSoFar, and functionaryRankName needs
// ctx.PrecedingCareer for the F6 title). Office Politics has no death
// mechanic at all — Survived is always true, the same as Citizen/Rogue/
// Entertainer — and WoundBadges is never set (Disabled here means
// "career ends," not a physical wound; see ResolveFunctionaryTerm's own
// doc comment for why careerWoundBadges must not be called on this
// segment).
func resolveFunctionarySegment(r *dice.Roller, upp UPP, maxTerms int, ctx segmentContext) careerSegment {
	aging := ctx.aging()

	career, careerUPP, finalTier := resolveFunctionaryCareerWithBudget(r, upp, maxTerms, ctx, aging)
	if aging.alive() {
		career.MusteringOut = ResolveFunctionaryMusterOut(r, career, finalTier)
	}

	boostedUPP, bonuses := ApplyMusteringOut(r, career, careerUPP)

	return careerSegment{
		Career: career, UPP: boostedUPP, Survived: true,
		Fame: bonuses.Fame, FameAwards: bonuses.FameAwards, Cash: bonuses.Cash,
		Skills: append(allSkillsFromTerms(career.Terms), bonuses.Skills...), LandGrants: bonuses.LandGrants,
		Education: ctx.Education,
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

	boostedUPP, bonuses := ApplyMusteringOut(r, career, careerUPP)

	var equipment []string

	for _, t := range career.Terms {
		if t.RewardResult != "" && t.RewardResult != "None" {
			equipment = append(equipment, t.RewardResult)
		}
	}

	// Individual per-term awards, appended rather than pre-summed, so
	// the Fame Stacks cap — applied once, across every segment, in
	// assembleChainCharacter — sees them as the book counts them (#126;
	// see resolveNobleSegment's own comment below for the same shape,
	// and craftsmanTermFameAwards' own doc comment for the award
	// formula). Fame here is this segment's own running sum, same as
	// every other segment resolver (e.g. resolveNobleSegment below) —
	// resolveFameStacks itself only runs once, on the full chain.
	fameAwards := bonuses.FameAwards
	fameAwards = append(fameAwards, craftsmanTermFameAwards(career.Terms)...)

	return careerSegment{
		Career: career, UPP: boostedUPP, Survived: true,
		Fame:       sumInts(fameAwards),
		FameAwards: fameAwards, Cash: bonuses.Cash,
		Skills: append(allSkillsFromTerms(career.Terms), bonuses.Skills...), Equipment: equipment,
		Masterpieces: masterpiecesFromTerms(career.Terms), LandGrants: bonuses.LandGrants,
		Education: ctx.Education,
	}
}

func resolveNobleSegment(r *dice.Roller, upp UPP, maxTerms int, ctx segmentContext) careerSegment {
	aging := ctx.aging()

	career, careerUPP, landGrants := resolveNobleCareerAndUPPWithBudget(r, upp, maxTerms, aging)
	if aging.alive() {
		career.MusteringOut = ResolveNobleMusterOut(r, career)
	}

	boostedUPP, bonuses := ApplyMusteringOut(r, career, careerUPP)
	ok := len(career.Terms) > 0

	fameAwards := bonuses.FameAwards
	if ok {
		// Two separate p.91 awards, appended individually rather than
		// pre-summed, so the Fame Stacks cap sees them as the book counts
		// them (fame.go's own resolveFameStacks).
		fameAwards = append(fameAwards, nobleBaseFame(upp.Characteristics[C6]))
		if exile := nobleExileFame(career.Terms); exile > 0 {
			fameAwards = append(fameAwards, exile)
		}
	}

	fame := sumInts(fameAwards)

	return careerSegment{
		Career:     career,
		UPP:        boostedUPP,
		Survived:   ok,
		Fame:       fame,
		FameAwards: fameAwards,
		Cash:       bonuses.Cash,
		Skills:     append(allSkillsFromTerms(career.Terms), bonuses.Skills...),
		LandGrants: append(landGrants, bonuses.LandGrants...),
		Education:  ctx.Education,
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

// resolveCitizenPensionReplacement applies Book 1 p.70's precedence
// between the two overlapping pensions: "a Functionary receives
// Cr15,000 per year (which replaces a Citizen's pension, if any)."
//
// Only these two interact. p.70 is explicit that Entitlements otherwise
// stack — "A character may receive duplicate Entitlements (for example,
// a Reserve and a Functionary pension, or both Military and Professor's
// retirement pay)" — so a Professor's pension and Retirement Pay sit
// happily beside either of these.
//
// Resolved here rather than in the Functionary resolver because it is a
// statement about a whole character: a muster-out resolver sees one
// career and cannot know another produced a Citizen's pension.
func resolveCitizenPensionReplacement(careers []Career) {
	functionaryPensioned := false

	for _, career := range careers {
		if career.Name == FunctionaryCareerName && career.MusteringOut.Pension > 0 {
			functionaryPensioned = true

			break
		}
	}

	if !functionaryPensioned {
		return
	}

	for i := range careers {
		if careers[i].Name == CitizenCareerName {
			careers[i].MusteringOut.Pension = 0
		}
	}
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
	careers                        []Career
	skills                         []SkillLevel
	medals                         []string
	equipment                      []string
	masterpieces                   []Masterpiece
	cash, woundBadges, termsServed int
	// fameAwards are every Fame award from every career, kept separate
	// for Book 1 p.91's Fame Stacks rule (resolveFameStacks).
	fameAwards []int
	// landGrants accumulate across careers rather than being replaced,
	// per p.68 ("retains it at Mustering Out") and p.88 ("Land Grants Are
	// Cumulative").
	landGrants []LandGrant
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

	acc.termsServed += termsServed(seg.Career.Terms)
	acc.skills = append(acc.skills, seg.Skills...)
	acc.medals = append(acc.medals, seg.Medals...)
	acc.equipment = append(acc.equipment, seg.Equipment...)
	acc.masterpieces = append(acc.masterpieces, seg.Masterpieces...)
	acc.fameAwards = append(acc.fameAwards, seg.FameAwards...)
	acc.landGrants = append(acc.landGrants, seg.LandGrants...)
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
	// An unbounded request is not unbounded in fact: maxCareerTerms is
	// where Book 1 p.89's own account of a life runs out (see its doc
	// comment), so "no -age target" means "the end of the table", and
	// that is one budget for the whole life rather than a fresh one per
	// career.
	//
	// Handing each career its own maxCareerTerms was a real defect,
	// masked until #110 by the chain's one-term cap on non-final careers.
	// Without the cap it let a chain of two careers run twenty-eight
	// terms and reach age 130 — past the end of every table this codebase
	// has. The budget the cap replaced (a017108 dropped a terms-served
	// parameter when it added the cap) is restored here as the age it was
	// always really measuring.
	if ageTarget == 0 {
		ageTarget = AgeFromTermsServed(maxCareerTerms)
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

func generateCareerChainStart(r *dice.Roller) (UPP, string, []SkillLevel, Education) {
	return generateStart(r)
}

// chainSegmentContext builds the per-segment context from the chain's
// running state — the shared Aging simulation, terms served so far, the
// skills accumulated to date, and the immediately-preceding career name.
func chainSegmentContext(
	aging *agingSimulation,
	acc *careerChainAccumulator,
	precedingCareer string,
	education Education,
) segmentContext {
	priorCareers := make([]string, 0, len(acc.careers))
	for _, career := range acc.careers {
		priorCareers = append(priorCareers, career.Name)
	}

	return segmentContext{
		Aging:            aging,
		Education:        education,
		PrecedingCareer:  precedingCareer,
		TermsServedSoFar: acc.termsServed,
		SkillsSoFar:      aggregateSkills(acc.skills),
		PriorCareers:     priorCareers,
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
// Each listed career runs to its own natural end — until its Continue
// roll fails, or Aging, a Disabled result or the term budget stops it —
// before the next is attempted. The next career is therefore reached
// only if the previous one ended early enough to leave room, which is
// often not the case: a five-career chain reaches its last entry 13% of
// the time. That is the rule working rather than a failure. See #110 for
// why the previous behaviour — one term per non-final career — had no
// basis; Book 1's own one-term obligation (p.61) belongs to commissioned
// academy graduates and will arrive with them.
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

	upp, homeworld, homeworldSkills, education := generateCareerChainStart(r)

	acc := careerChainAccumulator{skills: slices.Clone(homeworldSkills)}

	// One simulation for the whole chain — see agingSimulation's own doc
	// comment for why it cannot be per-segment.
	var aging agingSimulation

	everSucceeded, survived := false, true
	precedingCareer := ""

	for _, name := range careerNames {
		maxTerms, attemptAllowed := segmentBudget(ageTarget, aging.age())
		if !attemptAllowed {
			break
		}

		seg := careerChainRegistry[name](r, upp, maxTerms,
			chainSegmentContext(&aging, &acc, precedingCareer, education))
		upp = seg.UPP
		// A Later Education attendance (#113 item 5) mutates the
		// segment's own copy of Education, not the chain-wide one
		// chainSegmentContext built by value — carried forward here so
		// the next segment sees it too, and so the final Character
		// below reflects it even for a single-segment chain.
		education = seg.Education
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

	resolveCitizenPensionReplacement(acc.careers)

	age, lifeStage, notes, survived := finalizeAging(&aging, survived)

	c := assembleChainCharacter(r, chainAssembly{
		upp: upp, homeworld: homeworld, education: education, acc: &acc,
		age: age, lifeStage: lifeStage, notes: notes,
	})

	return applyEducation(c, education), survived, nil
}

// chainAssembly is everything the final Character needs that is not
// already in the accumulator — split out only so
// GenerateCareerChainCharacter's own body stays inside the length the
// linter allows, once step C added a line to it.
type chainAssembly struct {
	upp            UPP
	homeworld      string
	education      Education
	acc            *careerChainAccumulator
	age, lifeStage int
	notes          string
}

func assembleChainCharacter(r *dice.Roller, a chainAssembly) Character {
	acc := a.acc
	birthdate := GenerateBirthdate(r, a.age)

	return Character{
		Species:        "Human",
		GeneticProfile: humanGeneticProfile,
		UPP:            a.upp,
		Homeworld:      a.homeworld,
		Birthworld:     a.homeworld,
		Birthdate:      birthdate,
		Age:            a.age,
		LifeStage:      a.lifeStage,
		Notes:          a.notes,
		Education:      a.education,
		// Skills below are resolved through applyEducation by the caller
		// rather than here, so both paths use one implementation.
		Rank:         chainRank(acc.careers),
		Fame:         resolveFameStacks(acc.fameAwards),
		Cash:         acc.cash,
		WoundBadges:  acc.woundBadges,
		Careers:      acc.careers,
		Skills:       aggregateSkills(acc.skills),
		Medals:       acc.medals,
		Equipment:    acc.equipment,
		Masterpieces: acc.masterpieces,
		LandGrants:   acc.landGrants,
	}
}
