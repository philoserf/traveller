package character

import (
	"slices"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// HumanSpecies and HumanGeneticProfile are the values every character
// this codebase generates carries — nothing produces a non-Human
// character yet. Exported so render can tell a default apart from a
// deviation worth showing (render/character.go's own writeMetadata)
// rather than hardcoding the same two strings a second time.
//
// HumanGeneticProfile is Str Dex End Int Edu Soc, matching Position's
// own C1-C6 doc comment and the exact example ("e.g. SDEIES")
// Character.GeneticProfile's own field comment already gives.
const (
	HumanSpecies        = "Human"
	HumanGeneticProfile = "SDEIES"
)

// HumanCharacteristicMax is the highest value a Human characteristic can
// reach through in-generation awards (Book 1 p.70). It is far below
// ehex.Max, which is only the largest value a single extended-hex digit
// can encode (33) — a storage limit, not a species one.
//
// Awards in this codebase are all +1, so "clamped at 15" and p.70's own
// "an award that would exceed 15 is lost" describe the same outcome; the
// distinction would only matter for a +2 award, which no table grants.
//
// Reductions are not bounded by this: Aging and Risk both drive
// characteristics down, and their floor is 0, not 15.
const HumanCharacteristicMax = 15

// awardCharacteristic applies an award of amount to current, per Book 1
// p.70: an award that would take the result past HumanCharacteristicMax
// is lost, leaving the characteristic untouched.
//
// Lost, not clamped down to the cap. A test fixture starting above 15 —
// several use 20 to make Risk unfailable — must not be dragged down to
// 15 by receiving an award, which is what a plain min() would do.
func awardCharacteristic(current ehex.Value, amount int) ehex.Value {
	boosted := int(current) + amount
	if boosted > HumanCharacteristicMax {
		return current
	}

	//nolint:gosec // bounded above by HumanCharacteristicMax
	return ehex.Value(boosted)
}

// humanGeneticProfile is HumanGeneticProfile's own unexported alias,
// kept so this package's many struct literals read unchanged.
const humanGeneticProfile = HumanGeneticProfile

// scoutWoundBadges counts one Wound Badge per term whose Risk roll left
// the Controlling Characteristic reduced — Wounded or Disabled, both
// "reduced" outcomes per Book 1 p.65: "Reduced. If the Controlling
// Characteristic is reduced, the Character has been wounded and receives
// a Wound Badge." Disabled is a wound too, not a wound-badge-exempt
// separate category — p.69's own Disability Muster Out sidebar describes
// a Disabling result as caused by "Risk Failure produces an Injury or
// Wound".
func scoutWoundBadges(career Career) int {
	n := 0

	for _, t := range career.Terms {
		if t.RiskResult == Wounded || t.RiskResult == Disabled {
			n++
		}
	}

	return n
}

// allSkillsFromTerms flattens genuine skills from every term, in term
// order. Personal awards are characteristic improvements, not skills.
func allSkillsFromTerms(terms []Term) []SkillLevel {
	var skills []SkillLevel

	for _, t := range terms {
		for _, skill := range t.SkillsAwarded {
			if skill.Kind != Personal {
				skills = append(skills, skill)
			}
		}
	}

	return skills
}

// allMedalsFromTerms flattens every term's own Medals, in term order —
// mirrors allSkillsFromTerms exactly. Scout never populates Term.Medals
// (Armed Forces only), so this is a no-op for Scout callers; Marine is
// the first real source.
func allMedalsFromTerms(terms []Term) []string {
	var medals []string
	for _, t := range terms {
		medals = append(medals, t.Medals...)
	}

	return medals
}

// lastTermRank returns the rank held after the final term — "" if there
// are no terms, or if the career has no rank mechanic (Term.Rank is
// Armed Forces only; always "" for Scout, since a Scout term never sets
// it).
func lastTermRank(terms []Term) string {
	if len(terms) == 0 {
		return ""
	}

	return terms[len(terms)-1].Rank
}

// GenerateScoutCharacter generates a full Human Scout Character end to
// end: a UPP, a homeworld and its background skills, a full multi-term
// Scout career, and Scout's own Mustering Out benefits — the first
// function in this package to construct a Character value.
//
// Returns false specifically when Career Resolution itself didn't produce
// a usable character: either the career never qualified (Begin and Retry
// both failed — Book 1's own "this career may not be used," not an error;
// no other career exists in this codebase yet to fall back to) or the
// last term ended in Death (Book 1 p.69's "Dying During Character
// Generation": "If the Controlling Characteristic is reduced to zero or
// less, the Character is dead (and all efforts in this particular
// character creation process are lost)" — stronger than "no benefits,"
// RAW voids the whole attempt). Either way the partial Character is still
// returned, not zero-valued, so a caller has something to inspect or
// retry from — matching the existing return-both-value-and-signal pattern
// already used throughout this package (BeginScout, ResolveScoutTerm).
//
// An Aging-caused death (character/aging.go's own ResolveAging, applied
// via finalizeAging below) also returns ok=false, though it reaches that
// answer by a different route than p.69's career-death rule above. This
// package previously returned ok=true for it, reasoning that p.89's
// "ultimately... bringing on inevitable death" is a normal outcome
// rather than a voided attempt — someone dying "of old age decades after
// a successful career," a story for Notes to tell rather than a signal.
// That reasoning didn't survive contact with what finalizeAging actually
// simulates: Aging runs only through the career-end age, never past it,
// so such a death lands inside the lifetime generation just built rather
// than decades beyond it. See finalizeAging's own doc comment.
//
// Age, LifeStage, Notes are computed via finalizeAging (character/aging.go):
// Age is only an approximation (18 plus 4 years per term served — whether
// Begin succeeded on the first roll or needed the 1-year Retry isn't
// surfaced by BeginScout's bool return, so this can undercount by up to a
// year, per AgeFromTermsServed's own doc comment), and Notes only carries
// text when Aging actually produced an illness or death event. Birthdate
// is computed via GenerateBirthdate (character/birthdate.go) from that
// same approximate Age, so it inherits the same imprecision.
//
// Left at zero-value: Name — nothing in this codebase generates it yet.
// Rank, Medals, Commendations — correctly zero for Scout, not a gap: p.65
// states outright that "the Citizen, Entertainer, Craftsman, Scout, Agent,
// and Rogue careers have no rank," and Scout's own p.79 box grants neither
// Medals nor Commendations (both Armed-Forces/Agent-specific per the
// generic Reward-Mods table, p.64). Equipment stays deferred — no
// structured field to apply the remaining narrative Mustering Out
// benefits to yet — but Fame and Cash are now genuinely computed via
// ApplyMusteringOut (character/muster_out_apply.go), which parses
// MusteringOut's own Money/Benefits strings ("Cr30,000," "Fame +2") into
// the numeric totals below.
//
// Skills concatenates homeworld skills and every term's SkillsAwarded with
// no deduplication or level-merging — e.g. two separate "Jack of all
// Trades" grants stay two separate SkillLevel{Level: 1} entries rather
// than merging into one Level: 2 entry, even though Book 1's general
// skill-acquisition text (p.65-66) implies repeated grants of the same
// skill should increase its level. This is a deliberate, documented
// simplification — proper merge semantics are separate, nontrivial logic
// better scoped once more than one career/skill source exists to validate
// the merge rule against.
func GenerateScoutCharacter(r *dice.Roller) (Character, bool) {
	upp := GenerateUPP(r)
	homeworld, homeworldSkills := GenerateHomeworldSkills(r)

	return buildScoutCharacter(r, upp, homeworld, homeworldSkills)
}

// buildScoutCharacter assembles a Character from an already-rolled upp,
// homeworld, and homeworldSkills, resolving a Scout career and Mustering
// Out against upp via r. Split out from GenerateScoutCharacter so tests
// can pin upp to a fixture (never-qualified, near-certain-death, immortal
// — the same fixtures career_loop_test.go/career_muster_out_test.go
// already establish) instead of seed-hunting for rare GenerateUPP
// outcomes, mirroring ResolveScoutCareer's own upp-as-parameter
// precedent. Delegates to buildRiskCareerCharacter, below — see its own
// doc comment for why.
func buildScoutCharacter(r *dice.Roller, upp UPP, homeworld string, homeworldSkills []SkillLevel) (Character, bool) {
	return buildRiskCareerCharacter(
		r, upp, homeworld, homeworldSkills, func(r *dice.Roller, upp UPP, aging *agingSimulation) (Career, UPP) {
			return resolveScoutCareerWithBudget(r, upp, maxCareerTerms, aging)
		}, ResolveScoutMusterOut, scoutDiscoveryFame)
}

// scoutDiscoveryFame is Book 1 p.91's own "Scout: Discoveries x4" —
// Scout's own intrinsic Fame source, distinct from ApplyMusteringOut's
// separate Fame accumulation (Mustering Out's own "Fame +N" rolls).
func scoutDiscoveryFame(career Career) int {
	discoveries := 0

	for _, t := range career.Terms {
		if t.RewardResult == "Discovery" {
			discoveries++
		}
	}

	return discoveries * 4
}

// buildRiskCareerCharacter assembles a Character for any career sharing
// Scout's own shape: resolveCareer returns (Career, UPP) threading a
// Risk-driven characteristic reduction forward, ok is len(Terms) > 0 &&
// the last term's RiskResult != Dead, and WoundBadges counts Wounded/
// Disabled terms via the already-generic scoutWoundBadges. Extracted
// once Marine became a second real, byte-identical caller of this exact
// shape (flagged by golangci-lint's own dupl check) — the same
// "generalize on 2nd instance" discipline already applied repeatedly
// this session (musterOutRow, resolveRisk/resolveReward/riskOutcome,
// resolveCareerLoop). Citizen and Noble are NOT unified in here: each
// has a genuinely different ok/WoundBadges shape of its own (no death
// mechanic), not just a different Resolve* pair — forcing them into
// this same shape would paper over a real difference, not remove
// duplication.
//
// careerFame is caller-supplied (not hardcoded) because Scout's and
// Marine's own intrinsic Fame formulas (Book 1 p.91) genuinely differ —
// Discoveries x4 vs. correctly-always-0 for an Enlisted-only Marine —
// unlike resolveCareer/resolveMusterOut, which Scout and Marine happen
// to share the same shape of but not the same values for either.
func buildRiskCareerCharacter(
	r *dice.Roller,
	upp UPP,
	homeworld string,
	homeworldSkills []SkillLevel,
	resolveCareer func(r *dice.Roller, upp UPP, aging *agingSimulation) (Career, UPP),
	resolveMusterOut func(r *dice.Roller, career Career) MusteringOut,
	careerFame func(career Career) int,
) (Character, bool) {
	var aging agingSimulation

	career, updatedUPP := resolveCareer(r, upp, &aging)
	if aging.alive() {
		career.MusteringOut = resolveMusterOut(r, career)
	}

	boostedUPP, bonuses := ApplyMusteringOut(career, updatedUPP)

	// Clone before appending: appending onto the caller's own
	// homeworldSkills slice in place could silently corrupt an earlier
	// Character's Skills if that slice is ever reused across multiple
	// calls (append reuses spare backing-array capacity when there is
	// any).
	skills := append(slices.Clone(homeworldSkills), allSkillsFromTerms(career.Terms)...)
	skills = append(skills, bonuses.Skills...)

	survivedCareer := len(career.Terms) > 0 && career.Terms[len(career.Terms)-1].RiskResult != Dead

	age, lifeStage, notes, ok := finalizeAging(&aging, survivedCareer)
	birthdate := GenerateBirthdate(r, age)

	// careerFame is gated on ok, matching buildNobleCharacter's own
	// analogous Fame terms: a character who died mid-career (Book 1
	// p.69's "Dying During Character Generation") shouldn't retain Fame
	// from an earlier term's Discovery — bonuses.Fame doesn't need the
	// same gate, since resolveMusterOut's own roll-count already zeroes
	// out on a Dead last term.
	fame := bonuses.Fame
	if survivedCareer {
		fame += careerFame(career)
	}

	return Character{
		Species:        "Human",
		GeneticProfile: humanGeneticProfile,
		UPP:            boostedUPP,
		Homeworld:      homeworld,
		Birthworld:     homeworld, // same world; see GenerateHomeworldSkills' own doc comment
		Birthdate:      birthdate,
		Age:            age,
		LifeStage:      lifeStage,
		Notes:          notes,
		Rank:           lastTermRank(career.Terms),
		Fame:           fame,
		Cash:           bonuses.Cash,
		Careers:        []Career{career},
		Skills:         aggregateSkills(skills),
		Medals:         allMedalsFromTerms(career.Terms),
		WoundBadges:    scoutWoundBadges(career),
	}, ok
}
