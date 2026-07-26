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

// MusterOutCashAmount parses a Money-column entry for its Cr value, per
// Book 1's own "CrN,NNN" literal notation. Passage/StarPass entries
// aren't cash and report false — see ApplyMusteringOut's own doc comment
// for why they're left unapplied. Exported: render's own writeMusteringOut
// (render/character.go) reuses this same parse to identify which Money
// entries duplicate the character-wide accumulated Cash total, rather
// than reimplementing "Cr" + comma-stripped parsing a second time.
func MusterOutCashAmount(entry string) (int, bool) {
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
}

// ApplyMusteringOut applies m's mechanical effects onto upp: Fame and
// Cash accumulate from every Money/Benefits entry that carries one, and
// characteristic boosts add directly to the relevant Position (clamped
// at ehex.Max, defensively — no realistic roll count gets remotely
// close). Everything else in m — Passages, StarPass, Ship Share,
// Forbidden Knowledge, Wafer Jack, Life Insurance, TAS Fellow
// Membership, Knighthood — has no structured Character field to apply to
// yet; it stays recorded only in MusteringOut's own []string fields, the
// exact gap ResolveScoutMusterOut's own doc comment already flagged as
// deferred, not silently dropped.
func ApplyMusteringOut(m MusteringOut, upp UPP) (UPP, MusterOutBonuses) {
	var bonuses MusterOutBonuses

	for _, entry := range m.Money {
		if amount, ok := MusterOutCashAmount(entry); ok {
			bonuses.Cash += amount
		}
	}

	for _, entry := range m.Benefits {
		if amount, ok := musterOutFameBonus(entry); ok {
			bonuses.Fame += amount

			continue
		}

		if p, amount, ok := musterOutCharacteristicBoost(entry); ok {
			boosted := min(int(upp.Characteristics[p])+amount, int(ehex.Max))

			//nolint:gosec // bounded by the min(...) clamp above, gosec can't see that
			upp.Characteristics[p] = ehex.Value(boosted)
		}
	}

	return upp, bonuses
}
