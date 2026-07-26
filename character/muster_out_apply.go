package character

import (
	"strconv"
	"strings"

	"github.com/philoserf/traveller/ehex"
)

// musterOutCharacteristicNames maps every characteristic-boost token this
// codebase's Mustering Out tables use onto its Position — both the human
// name shorthand ("Str +1") and the raw C1-C6 box shorthand ("C5 +1"),
// since the same table literal mixes both forms (Book 1's own p.67-68 box
// text, not a codebase inconsistency).
var musterOutCharacteristicNames = map[string]Position{
	"Str": C1, "Dex": C2, "End": C3, "Int": C4, "Edu": C5, "Soc": C6,
	"C1": C1, "C2": C2, "C3": C3, "C4": C4, "C5": C5, "C6": C6,
}

// splitMusterOutBoost is the shared "Name +N" parse behind
// musterOutFameBonus and musterOutCharacteristicBoost.
func splitMusterOutBoost(entry string) (string, int, bool) {
	name, amountText, found := strings.Cut(entry, " +")
	if !found {
		return "", 0, false
	}

	amount, err := strconv.Atoi(amountText)
	if err != nil {
		return "", 0, false
	}

	return name, amount, true
}

// musterOutCashAmount parses a Money-column entry for its Cr value, per
// Book 1's own "CrN,NNN" literal notation. Passage/StarPass entries
// aren't cash and report false — see ApplyMusteringOut's own doc comment
// for why they're left unapplied.
func musterOutCashAmount(entry string) (int, bool) {
	if !strings.HasPrefix(entry, "Cr") {
		return 0, false
	}

	amount, err := strconv.Atoi(strings.ReplaceAll(entry[len("Cr"):], ",", ""))
	if err != nil {
		return 0, false
	}

	return amount, true
}

// musterOutFameBonus parses a Benefits-column entry for a Fame bonus, per
// Book 1's own "Fame +N" notation.
func musterOutFameBonus(entry string) (int, bool) {
	name, amount, ok := splitMusterOutBoost(entry)
	if !ok || name != "Fame" {
		return 0, false
	}

	return amount, true
}

// musterOutCharacteristicBoost parses a Benefits-column entry for a
// characteristic boost, per Book 1's own "Str +1"/"C2 +1" notation.
func musterOutCharacteristicBoost(entry string) (Position, int, bool) {
	name, amount, ok := splitMusterOutBoost(entry)
	if !ok {
		return 0, 0, false
	}

	p, ok := musterOutCharacteristicNames[name]
	if !ok {
		return 0, 0, false
	}

	return p, amount, true
}

// MusterOutBonuses is the numeric total of every Fame/Cash entry
// ApplyMusteringOut found in a MusteringOut's Money/Benefits — a named
// struct, not two positional same-typed ints, so a call site can't
// silently transpose Fame and Cash.
type MusterOutBonuses struct {
	Fame int
	Cash int
	// Skills are benefits that grant a skill outright rather than money,
	// Fame or a characteristic — Forbidden Knowledge today. Returned
	// rather than applied, since this function has no Character to apply
	// them to; callers fold them in beside the career's own term skills.
	Skills []SkillLevel
}

// ApplyMusteringOut applies m's mechanical effects onto upp: Fame and
// Cash accumulate from every Money/Benefits entry that carries one, and
// characteristic boosts add directly to the relevant Position, clamped
// at HumanCharacteristicMax (Book 1 p.70: an award that would take a
// Human characteristic past 15 is lost). This used to clamp at ehex.Max,
// which is a digit-encoding limit rather than a species one and let
// repeated awards carry a characteristic to 17 in practice. Everything else in m — Passages, StarPass, Ship Share,
// Forbidden Knowledge, Wafer Jack, Life Insurance, TAS Fellow
// Membership, Knighthood — has no structured Character field to apply to
// yet; it stays recorded only in MusteringOut's own []string fields, the
// exact gap ResolveScoutMusterOut's own doc comment already flagged as
// deferred, not silently dropped.
func ApplyMusteringOut(career Career, upp UPP) (UPP, MusterOutBonuses) {
	var bonuses MusterOutBonuses

	for _, entry := range career.MusteringOut.Money {
		if amount, ok := musterOutCashAmount(entry); ok {
			bonuses.Cash += amount
		}
	}

	for _, entry := range career.MusteringOut.Benefits {
		if amount, ok := musterOutFameBonus(entry); ok {
			bonuses.Fame += amount

			continue
		}

		if p, amount, ok := musterOutCharacteristicBoost(entry); ok {
			upp.Characteristics[p] = awardCharacteristic(upp.Characteristics[p], amount)

			continue
		}

		switch entry {
		case knighthoodBenefit:
			upp.Characteristics[C6] = applyKnighthood(upp.Characteristics[C6], career)
		case forbiddenKnowledgeBenefit:
			bonuses.Skills = append(bonuses.Skills, skillLevel1(forbiddenKnowledgeSkill, Skill))
		}
	}

	return upp, bonuses
}

// Benefit strings this function resolves mechanically rather than
// leaving as narrative record. They must match the muster-out tables'
// own spellings exactly (career_muster_out.go and friends).
const (
	knighthoodBenefit         = "Knighthood"
	forbiddenKnowledgeBenefit = "Forbidden Knowledge"
)

// forbiddenKnowledgeSkill is what a Forbidden Knowledge award grants.
//
// Book 1 p.69 describes the benefit — "a skill or knowledge that is not,
// should not, or cannot, be mentioned in polite society… Each receipt
// provides skill-1" — but gives no table to roll on, only a single
// worked example: "a familiarity with the use of machineguns (Fighter)
// is not mentioned in polite conversation."
//
// So this grants that example rather than a list of this codebase's own
// invention. Widening it to a random pick would mean deciding which
// skills count as forbidden, which the book does not do.
const forbiddenKnowledgeSkill = "Fighter"

// knighthoodSocFloor is Book 1 p.68's "A Knighthood raises any value of
// Soc to B" — B being 11 in extended hex.
const knighthoodSocFloor = 11

// applyKnighthood resolves Book 1 p.68's Knighthood benefit against the
// character's Social Standing:
//
//	"A Knighthood raises any value of Soc to B; if the character is
//	already Soc 11+, he receives Soc +1 instead."
//	"In the Spacer, Soldier, and Marine careers, Knighthood is only
//	available to Officers. A non-officer receives Soc +1 (even if it
//	advances Soc to 11 or beyond)."
//
// Only Humans are generated here, so C6 is always Soc — p.68's separate
// C6=Caste and C6=Charisma cases can't arise (a Caste character keeps
// the title's social standing but no mechanical change; noted, not
// implemented, matching this package's treatment of other Human-only
// branches).
func applyKnighthood(soc ehex.Value, career Career) ehex.Value {
	if knighthoodIsOfficersOnly(career.Name) && !mustersOutAsOfficer(career) {
		return awardCharacteristic(soc, 1)
	}

	if soc >= knighthoodSocFloor {
		return awardCharacteristic(soc, 1)
	}

	// Not awardCharacteristic: this is a raise to a floor, not an
	// increment, and the floor is below HumanCharacteristicMax anyway.
	return knighthoodSocFloor
}

// knighthoodIsOfficersOnly reports the three careers p.68 restricts.
func knighthoodIsOfficersOnly(careerName string) bool {
	return careerName == SpacerCareerName ||
		careerName == SoldierCareerName ||
		careerName == MarineCareerName
}

// mustersOutAsOfficer reads Officer status off the rank held at the end
// of the career. Spacer, Soldier and Marine all format an Officer rank
// with an "O" prefix and their Enlisted ranks with something else (R, S
// and M respectively), so the prefix is the whole test — and it is only
// consulted for those three careers, where that holds.
func mustersOutAsOfficer(career Career) bool {
	return strings.HasPrefix(lastTermRank(career.Terms), "O")
}
