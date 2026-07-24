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
// Returns false when this attempt didn't produce a usable, living
// character: either the career never qualified (Begin and Retry both
// failed — Book 1's own "this career may not be used," not an error; no
// other career exists in this codebase yet to fall back to) or the last
// term ended in Death (Book 1 p.69's "Dying During Character Generation":
// "If the Controlling Characteristic is reduced to zero or less, the
// Character is dead (and all efforts in this particular character
// creation process are lost)" — stronger than "no benefits," RAW voids
// the whole attempt). Either way the partial Character is still returned,
// not zero-valued, so a caller has something to inspect or retry from —
// matching the existing return-both-value-and-signal pattern already used
// throughout this package (BeginScout, ResolveScoutTerm).
//
// Left at zero-value, each for a distinct reason: Name, Birthdate, Age,
// LifeStage, Notes — nothing in this codebase generates any of these yet
// (Age specifically can't even be cheaply approximated: whether Begin
// succeeded on the first roll or needed the 1-year Retry, Book 1's own
// "each failed attempt... takes one year," isn't surfaced by BeginScout's
// bool return, so any guess would already be off by up to a year).
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

	return Character{
		Species:        "Human",
		GeneticProfile: humanGeneticProfile,
		UPP:            updatedUPP,
		Homeworld:      homeworld,
		Birthworld:     homeworld, // same world; see GenerateHomeworldSkills' own doc comment
		Careers:        []Career{career},
		Skills:         skills,
		WoundBadges:    scoutWoundBadges(career),
	}, ok
}
