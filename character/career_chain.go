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

type careerSegmentResolver func(r *dice.Roller, upp UPP) careerSegment

// careerChainRegistry is every career resolvable as a link in a
// multi-career chain. Noble is deliberately absent (its Begin is
// Soc-gated with no Continue/retry-a-different-career shape); Functionary
// and Craftsman are absent because they aren't implemented at all yet —
// see this slice's own plan-file Context for why Functionary needs a
// later phase's real terms-served count first.
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
	resolveCareer func(r *dice.Roller, upp UPP) (Career, UPP),
	resolveMusterOut func(r *dice.Roller, career Career) MusteringOut,
	careerFame func(career Career) int,
) careerSegment {
	career, careerUPP := resolveCareer(r, upp)
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

func resolveScoutSegment(r *dice.Roller, upp UPP) careerSegment {
	return resolveRiskCareerSegment(r, upp, ResolveScoutCareer, ResolveScoutMusterOut, scoutDiscoveryFame)
}

func resolveMarineSegment(r *dice.Roller, upp UPP) careerSegment {
	return resolveRiskCareerSegment(r, upp, ResolveMarineCareer, ResolveMarineMusterOut, marineCareerFame)
}

func resolveSoldierSegment(r *dice.Roller, upp UPP) careerSegment {
	return resolveRiskCareerSegment(r, upp, ResolveSoldierCareer, ResolveSoldierMusterOut, soldierCareerFame)
}

func resolveSpacerSegment(r *dice.Roller, upp UPP) careerSegment {
	return resolveRiskCareerSegment(r, upp, ResolveSpacerCareer, ResolveSpacerMusterOut, spacerCareerFame)
}

func resolveAgentSegment(r *dice.Roller, upp UPP) careerSegment {
	return resolveRiskCareerSegment(r, upp, ResolveAgentCareer, ResolveAgentMusterOut, agentCareerFame)
}

// resolveRogueSegment mirrors buildRogueCharacter's own body
// (rogue_character_generate.go), stopping short of finalizeAging. Rogue
// never modifies a characteristic at all (rogue_loop.go's own doc
// comment: "Rogue never modifies its own CC"), so upp passes through
// unchanged; there is no death mechanic, so Survived is always true.
func resolveRogueSegment(r *dice.Roller, upp UPP) careerSegment {
	career := ResolveRogueCareer(r, upp)
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
func resolveScholarSegment(r *dice.Roller, upp UPP) careerSegment {
	career, careerUPP := ResolveScholarCareer(r, upp)
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
func resolveEntertainerSegment(r *dice.Roller, upp UPP) careerSegment {
	career, fame := ResolveEntertainerCareer(r, upp)
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
// BeginMerchant never fails, so ok collapses to didn't-die-on-the-last-term.
func resolveMerchantSegment(r *dice.Roller, upp UPP) careerSegment {
	career, careerUPP, isOfficer, tier := ResolveMerchantCareer(r, upp)
	career.MusteringOut = ResolveMerchantMusterOut(r, career, isOfficer, tier)

	boostedUPP, bonuses := ApplyMusteringOut(career.MusteringOut, careerUPP)

	ok := career.Terms[len(career.Terms)-1].RiskResult != Dead

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
func resolveCitizenSegment(r *dice.Roller, upp UPP) careerSegment {
	career := ResolveCitizenCareer(r, upp)
	career.MusteringOut = ResolveCitizenMusterOut(r, career)

	boostedUPP, bonuses := ApplyMusteringOut(career.MusteringOut, upp)

	return careerSegment{
		Career: career, UPP: boostedUPP, Survived: true,
		Fame: bonuses.Fame, Cash: bonuses.Cash,
		Skills: allSkillsFromTerms(career.Terms),
	}
}

// validateCareerChain checks careerNames (already lowercased/trimmed by
// the caller) against the rules this slice's own plan-file Context
// derives from Book 1 pp.65-66: every name must be a known, chainable
// career; "citizen" may only be the first entry (p.64: "may not
// transfer to Citizen"); "noble" is recognized but rejected with a
// clearer message than "unknown career" (Noble's Begin doesn't fit this
// shape); and no two adjacent entries may repeat the same career
// ("selecting a different career" — re-entering the very career just
// left is meaningless when Continue already offers that for free).
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
// Returns ok=false only when a segment ends in Dead (Book 1 p.69:
// "Dying During Character Generation" voids the whole attempt) — same
// as every existing single-career entry point, everything accumulated
// through the fatal segment is still returned, not discarded, matching
// cmd/chargen's own "the sheet above still prints in full" precedent.
func GenerateCareerChainCharacter(r *dice.Roller, careerNames []string) (Character, bool, error) {
	if err := validateCareerChain(careerNames); err != nil {
		return Character{}, false, err
	}

	upp := GenerateUPP(r)
	homeworld, homeworldSkills := GenerateHomeworldSkills(r)

	var careers []Career

	skills := slices.Clone(homeworldSkills)

	var medals []string

	var fame, cash, woundBadges, termsServed int

	everSucceeded, survived := false, true

	for _, name := range careerNames {
		seg := careerChainRegistry[name](r, upp)
		careers = append(careers, seg.Career)
		upp = seg.UPP

		if len(seg.Career.Terms) == 0 {
			continue
		}

		everSucceeded = true
		termsServed += len(seg.Career.Terms)
		skills = append(skills, seg.Skills...)
		medals = append(medals, seg.Medals...)
		fame += seg.Fame
		cash += seg.Cash
		woundBadges += seg.WoundBadges

		if !seg.Survived {
			survived = false

			break
		}
	}

	if !everSucceeded {
		seg := careerChainRegistry["citizen"](r, upp)
		careers = append(careers, seg.Career)
		upp = seg.UPP
		termsServed = len(seg.Career.Terms)
		skills = append(skills, seg.Skills...)
		fame, cash = seg.Fame, seg.Cash
	}

	finalUPP, age, lifeStage, notes := finalizeAging(r, upp, termsServed, survived)
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
		Rank:           chainRank(careers),
		Fame:           fame,
		Cash:           cash,
		WoundBadges:    woundBadges,
		Careers:        careers,
		Skills:         skills,
		Medals:         medals,
	}, survived, nil
}
