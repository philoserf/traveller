package character

import "github.com/philoserf/traveller/dice"

// continueRogue mirrors continueMarineOutcome's own natural-2 override,
// but Rogue's own Continue check additionally uses the same natural-12-
// always-fails exception as rogueSucceeds instead of plain
// succeedsAgainst — Book 1's "Continue CC*" shares its own "*Mod +Terms.
// But, 12 is always automatic failure" footnote with Begin/Risk/Reward.
func continueRogue(r *dice.Roller, cc, mod int) bool {
	roll := r.TwoD6()

	switch roll {
	case 2:
		return true
	case 12:
		return false
	default:
		return succeedsAgainst(roll, cc, mod)
	}
}

// ResolveRogueCareer resolves a full multi-term Rogue career (Book 1
// p.84). The fixed CC is selected once via rollRogueCC and passed to
// resolveCareerLoop as a single-element positions slice — nextCC
// (career_loop.go) already returns the same element every call when
// len(positions)==1 (usedThisCycle fills to length 1 and clears every
// call), so "fixed for the whole career, not rotating" falls out of the
// existing shared loop for free, no new generalization needed.
//
// resolveCareerLoop's own RiskResult.Survived()/Disabled stop checks key
// off Term.RiskResult, which Rogue's own terms never set (stays at its
// zero value, Unharmed) — both checks pass through as "keep going" for
// every Rogue term. Confirmed by inspection, not just assumed: Rogue has
// no death/disability concept at all, so this is the correct behavior,
// not a coincidental side effect.
func ResolveRogueCareer(r *dice.Roller, upp UPP) Career {
	career, _ := resolveRogueCareerAndUPPWithBudget(r, upp, maxCareerTerms, &agingSimulation{}, nil, nil)

	return career
}

// resolveRogueCareerWithBudget is ResolveRogueCareer's own body, with
// the resolveCareerLoop term cap threaded as a parameter — see
// resolveCareerLoop's own doc comment for why.
// priorCareers are the careers this character already served, for Book
// 1 p.84's own "A Rogue may select for his Scheme (rather than roll) any
// previous career." Empty for a standalone Rogue, who has none.
// education is #113 item 5's own Later Education pilot (p.59): nil for
// callers with no Education context (ResolveRogueCareer, test-only);
// buildRogueCharacter and resolveRogueSegment thread the character's
// real, mutable Education through so the beforeTerm hook below can
// both read it (is a better institution now reachable?) and update it
// (attendInstitution) in place.
func resolveRogueCareerAndUPPWithBudget(
	r *dice.Roller, upp UPP, maxTerms int, aging *agingSimulation, priorCareers []string, education *Education,
) (Career, UPP) {
	career := Career{Name: RogueCareerName}

	ccPos := rollRogueCC(r)
	cc := int(upp.Characteristics[ccPos])

	if !BeginRogue(r, cc) {
		// Book 1 p.65: a failed Begin roll costs a year. Careers whose
		// Begin is a prerequisite rather than a roll (Noble's Soc B+,
		// Craftsman's held skills, Citizen's automatic entry) charge
		// nothing here — there was no attempt to fail.
		upp = aging.chargeFailedAttempt(r, upp)

		return career, upp
	}

	// termCount tracks "+Terms" via local closure state, not
	// career.Terms — resolveCareerLoop builds its own local slice and
	// only assigns it to career.Terms after the whole loop returns, so
	// reading career.Terms from inside either closure would always see
	// zero terms. The same reason Marine/Soldier/Spacer accumulate their
	// own priorTerms locally instead of reading back through Career.
	termCount := 0

	// Years of prison owed at the start of the next term.
	pendingPrison := 0

	terms, finalUPP := resolveCareerLoop(r, upp, []Position{ccPos},
		func(r *dice.Roller, upp UPP, ccPos Position) (Term, UPP) {
			// The sentence earned last term is served at the start of this
			// one (Book 1 p.84), so it is threaded forward rather than
			// applied where it was incurred.
			term := ResolveRogueTerm(
				r,
				int(upp.Characteristics[ccPos]),
				termCount,
				pendingPrison,
				priorCareers,
			) // "+Terms"
			term.ControllingCharacteristic = ccPos
			pendingPrison = term.PrisonYears
			termCount++

			return term, upp // Rogue never modifies its own CC
		},
		func(r *dice.Roller, currentUPP UPP) bool {
			return continueRogue(r, int(currentUPP.Characteristics[ccPos]), termCount)
		},
		maxTerms,
		aging,
		nil,
		laterEducationHook(education),
	)
	career.Terms = terms

	return career, finalUPP
}

// laterEducationHook builds resolveCareerLoop's own beforeTerm hook
// (#113 item 5) from a caller's Education pointer, or reports nil if
// there is none — ResolveRogueCareer has no Education context at all,
// so Later Education is simply unavailable there rather than treated
// as "never attended anything."
//
// shouldAttemptLaterEducation reads *education as it stands at the
// start of this term (any Personal-row Edu growth from an earlier term
// has already landed via applyPersonalAwards, career_loop.go, since
// this hook runs after that on every iteration but the first);
// attendInstitution then mutates *education in place, so the next call
// sees this term's own attempt. Rogue is the pilot career for this
// mechanism — see PLAN.md.
func laterEducationHook(education *Education) beforeTerm {
	if education == nil {
		return nil
	}

	return func(r *dice.Roller, upp UPP) (Term, UPP, bool) {
		school, ok := shouldAttemptLaterEducation(upp, *education)
		if !ok {
			return Term{}, upp, false
		}

		updatedUPP, skills, admitted := attendInstitution(r, upp, school, education)

		// p.59: "if accepted substitutes that process for the entire
		// term." A rejected application does not consume the term —
		// resolveCareerLoop falls through to the ordinary resolveTerm
		// path for this iteration instead, the same as if Later
		// Education had never been offered.
		if !admitted {
			return Term{}, updatedUPP, false
		}

		return Term{
			Length:               4,
			LaterEducation:       true,
			LaterEducationSchool: school.Name,
			SkillsAwarded:        skills,
		}, updatedUPP, true
	}
}
