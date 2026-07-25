package character

import (
	"slices"

	"github.com/philoserf/traveller/dice"
)

// humanGeneticProfile is the GeneticProfile for a Human character: Str
// Dex End Int Edu Soc, matching Position's own C1-C6 doc comment and the
// exact example ("e.g. SDEIES") Character.GeneticProfile's own field
// comment already gives.
const humanGeneticProfile = "SDEIES"

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

// allSkillsFromTerms flattens every term's SkillsAwarded, in term order.
func allSkillsFromTerms(terms []Term) []SkillLevel {
	var skills []SkillLevel
	for _, t := range terms {
		skills = append(skills, t.SkillsAwarded...)
	}

	return skills
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
// This does NOT cover an Aging-caused death (character/aging.go's own
// ResolveAging, applied via finalizeAging below): p.69's rule above is
// scoped to a Risk-roll reduction during an active Career Resolution
// attempt, a categorically different, narrower rule than Aging's own
// general "ultimately... bringing on inevitable death" (p.89), which the
// book frames as a normal outcome, not a voided attempt. A character who
// died of old age decades after a successful career still returns
// ok=true — the story of that death is recorded in Notes, not signaled
// by this return value. A caller that needs to detect it should check
// Notes rather than ok.
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
// generic Reward-Mods table, p.64). Fame, Cash, Equipment — deferred, not
// this slice's job: MusteringOut's own Money/Benefits fields are
// human-readable strings ("Cr30,000," "Fame +2"), and turning those into
// structured numeric/typed values is exactly the mechanical-application
// work career_muster_out.go's own doc comments already declined to do.
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
// precedent.
func buildScoutCharacter(r *dice.Roller, upp UPP, homeworld string, homeworldSkills []SkillLevel) (Character, bool) {
	career, updatedUPP := ResolveScoutCareer(r, upp)
	career.MusteringOut = ResolveScoutMusterOut(r, career)

	// Clone before appending: appending onto the caller's own
	// homeworldSkills slice in place could silently corrupt an earlier
	// Character's Skills if that slice is ever reused across multiple
	// buildScoutCharacter calls (append reuses spare backing-array
	// capacity when there is any).
	skills := append(slices.Clone(homeworldSkills), allSkillsFromTerms(career.Terms)...)

	ok := len(career.Terms) > 0 && career.Terms[len(career.Terms)-1].RiskResult != Dead

	finalUPP, age, lifeStage, notes := finalizeAging(r, updatedUPP, len(career.Terms), ok)
	birthdate := GenerateBirthdate(r, age)

	return Character{
		Species:        "Human",
		GeneticProfile: humanGeneticProfile,
		UPP:            finalUPP,
		Homeworld:      homeworld,
		Birthworld:     homeworld, // same world; see GenerateHomeworldSkills' own doc comment
		Birthdate:      birthdate,
		Age:            age,
		LifeStage:      lifeStage,
		Notes:          notes,
		Careers:        []Career{career},
		Skills:         skills,
		WoundBadges:    scoutWoundBadges(career),
	}, ok
}
